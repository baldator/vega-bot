package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto"
	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto/api"
	"golang.org/x/net/context"
)

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

func initializeSentry(sentryDsn string) {
	log.Println("Initialize sentry")
	if sentryDsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: sentryDsn,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}
	defer sentry.Flush(2 * time.Second)
}

func initializePrometheus(port int) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Printf("listen on %d\n", port)
		portString := ":" + strconv.Itoa(port)
		log.Fatal(http.ListenAndServe(portString, nil))
	}()
}

func logError(err error, sentryEnabled bool) {
	if sentryEnabled {
		sentry.CaptureException(err)
		sentry.Flush(time.Second * 5)
	}
	log.Fatal(err)
}
