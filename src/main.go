package main

import (
	"io"
	"log"
	"time"

	"github.com/baldator/vega-alerts/social"
	"github.com/baldator/vega-alerts/socialevents"

	"github.com/getsentry/sentry-go"
	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto"
	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	// Read application config
	conf, err := ReadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}

	initializeSentry(conf)

	func() {
		defer sentry.Recover()

		log.Println("Starting server")
		log.Println("Initialize social webservice connection")

		socialPost, err := social.NewSocialChannel(conf.SocialServiceURL, conf.SocialServiceKey, conf.SocialServiceSecret, conf.SocialTwitterEnabled, conf.SocialDiscordEnabled, conf.SocialSlackEnabled, conf.SocialTelegramEnabled)
		if err != nil {
			sentry.CaptureException(err)
			sentry.Flush(time.Second * 5)
			log.Fatal(err)
		}

		conn, err := grpc.Dial(conf.GrpcNodeURL, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(256<<20)))
		if err != nil {
			logError(err)
		}
		defer conn.Close()

		dataClient := api.NewTradingDataServiceClient(conn)
		eventType := []proto.BusEventType{
			proto.BusEventType_BUS_EVENT_TYPE_NETWORK_PARAMETER,
			proto.BusEventType_BUS_EVENT_TYPE_LOSS_SOCIALIZATION,
			proto.BusEventType_BUS_EVENT_TYPE_AUCTION,
			proto.BusEventType_BUS_EVENT_TYPE_PROPOSAL,
			proto.BusEventType_BUS_EVENT_TYPE_TRADE,
			proto.BusEventType_BUS_EVENT_TYPE_ORDER,
		}
		events, err := dataClient.ObserveEventBus(context.Background())

		currentEthereumConfig, err := readEthereumConfig(dataClient)
		if err != nil {
			logError(err)
		}

		done := make(chan bool)
		go func() {
			for {
				resp, err := events.Recv()

				if err == io.EOF {
					close(done)
					return
				}

				if err != nil {
					sentry.CaptureException(err)
					sentry.Flush(time.Second * 5)
					log.Println(err)
				}

				for _, event := range resp.Events {
					switch eventTypeLoop := event.Type; eventTypeLoop {
					case proto.BusEventType_BUS_EVENT_TYPE_NETWORK_PARAMETER: // Network has been reset (network ID has changed/block height reset)
						networkParameter := event.GetNetworkParameter()
						if networkParameter.Key == "blockchains.ethereumConfig" {
							message := socialevents.NetworkParametesNotification(dataClient, networkParameter, currentEthereumConfig)
							if message != "" {
								log.Println(message)
								socialPost.SendMessage(message)
							}
						}
					case proto.BusEventType_BUS_EVENT_TYPE_LOSS_SOCIALIZATION: // Loss socialisation alerts (distribution of funds generated by defaulting traders)
						lossSocialization := event.GetLossSocialization()
						message, err := socialevents.LossSocializationNotification(dataClient, lossSocialization)
						if err != nil {
							logError(err)
						}
						log.Println(message)
						socialPost.SendMessage(message)
					case proto.BusEventType_BUS_EVENT_TYPE_AUCTION: // Market price monitoring auction started/ended
						auction := event.GetAuction()
						message, err := socialevents.AuctionNotification(dataClient, auction)
						log.Println(message)
						if err != nil {
							logError(err)
						}
						socialPost.SendMessage(message)
					case proto.BusEventType_BUS_EVENT_TYPE_PROPOSAL: //New Market Proposal created, updated, enacted
						log.Println("BUS_EVENT_TYPE_PROPOSAL: ", event)
						proposal := event.GetProposal()
						message, err := socialevents.MarketProposalNotification(dataClient, proposal.Id, proposal.State)
						if err != nil {
							logError(err)
						}
						log.Println(message)
						socialPost.SendMessage(message)
					case proto.BusEventType_BUS_EVENT_TYPE_TRADE: // Rekt alert
						trade := event.GetTrade()
						if trade.Type == proto.Trade_TYPE_NETWORK_CLOSE_OUT_BAD {
							message, err := socialevents.RektNotification(dataClient, trade)
							if err != nil {
								logError(err)
							}
							log.Println(message)
							socialPost.SendMessage(message)
						}
					case proto.BusEventType_BUS_EVENT_TYPE_ORDER: // Whale alert
						order := event.GetOrder()
						if order.Status == proto.Order_STATUS_ACTIVE {
							value := order.Size * order.Price
							marketVal, marketFlag, _ := getMarketValue(dataClient, order.MarketId, order.Side, conf.WhaleOrdersThreshold)
							if float64(value) > (float64(marketVal)*conf.WhaleThreshold) && marketFlag {
								message, err := socialevents.WhaleNotification(dataClient, order)
								if err != nil {
									logError(err)
								}
								log.Println(message)
								socialPost.SendMessage(message)
							}
						}
					case proto.BusEventType_BUS_EVENT_TYPE_MARKET_CREATED:
						log.Println("BusEventType_BUS_EVENT_TYPE_MARKET_CREATED: ", event)

					}
				}
			}
		}()

		// When the batchSize is too small -> "rpc error: code = Unknown desc = EOF"
		observerEvent := api.ObserveEventBusRequest{Type: eventType, BatchSize: 10000}
		events.Send(&observerEvent)
		events.CloseSend()

		<-done //we will wait until all response is received

		log.Println("finished")
	}()
}

func readEthereumConfig(dataClient api.TradingDataServiceClient) (*proto.NetworkParameter, error) {
	log.Println("Initialize network parameters")
	request := api.NetworkParametersRequest{}
	network, err := dataClient.NetworkParameters(context.Background(), &request)
	if err != nil {
		return nil, err
	}

	var currentEthereumConfig *proto.NetworkParameter
	for _, param := range network.GetNetworkParameters() {
		if param.Key == "blockchains.ethereumConfig" {
			currentEthereumConfig = param
		}
	}

	return currentEthereumConfig, nil
}

func getMarketValue(dataClient api.TradingDataServiceClient, marketID string, side proto.Side, whaleOrdersThreshold int) (uint64, bool, error) {
	requestMarketDepth := api.MarketDepthRequest{MarketId: marketID}
	marketDepthObject, err := dataClient.MarketDepth(context.Background(), &requestMarketDepth)
	if err != nil {
		return 0, false, err
	}

	var marketValue uint64
	marketValue = 0
	marketOrdersBuy := len(marketDepthObject.Buy)
	marketOrderSell := len(marketDepthObject.Sell)
	marketOrdersFlag := false
	if marketOrdersBuy > whaleOrdersThreshold && marketOrderSell > whaleOrdersThreshold {
		marketOrdersFlag = true
	}

	if side == proto.Side_SIDE_BUY {
		for _, val := range marketDepthObject.Buy {
			marketValue = marketValue + val.Volume*val.Price
		}
	}

	if side == proto.Side_SIDE_SELL {
		for _, val := range marketDepthObject.Sell {
			marketValue = marketValue + val.Volume*val.Price
		}
	}

	return marketValue, marketOrdersFlag, nil
}

func initializeSentry(conf ConfigVars) {
	log.Println("Initialize sentry")
	if conf.SentryDsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: conf.SentryDsn,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}
	defer sentry.Flush(2 * time.Second)
}

func logError(err error) {
	sentry.CaptureException(err)
	sentry.Flush(time.Second * 5)
	log.Fatal(err)
}
