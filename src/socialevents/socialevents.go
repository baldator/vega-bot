package socialevents

import (
	"encoding/json"
	"strconv"

	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto"
	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto/api"
	"golang.org/x/net/context"
)

type EthereumConfig struct {
	NetworkID     string `json:"network_id"`
	ChainID       string `json:"chain_id"`
	BridgeAddress string `json:"bridge_address"`
	Confirmations int    `json:"confirmations"`
}

// MarketProposalNotification returns market proposal notification message
func MarketProposalNotification(dataClient api.TradingDataServiceClient, marketID string, state proto.Proposal_State) (string, error) {
	requestMarket := api.MarketByIDRequest{MarketId: marketID}
	MarketObject, err := dataClient.MarketByID(context.Background(), &requestMarket)
	if err != nil {
		return "", err
	}

	Market := MarketObject.GetMarket()
	stateString := getMarketProposalState(state)

	return "⚖️ Market proposal " + Market.TradableInstrument.Instrument.Name + " " + stateString, nil
}

func getMarketProposalState(state proto.Proposal_State) string {
	var stateString string
	switch state {
	case proto.Proposal_STATE_UNSPECIFIED:
		stateString = "undefined"
	case proto.Proposal_STATE_FAILED:
		stateString = "failed"
	case proto.Proposal_STATE_OPEN:
		stateString = "opened"
	case proto.Proposal_STATE_PASSED:
		stateString = "passed"
	case proto.Proposal_STATE_REJECTED:
		stateString = "rejected"
	case proto.Proposal_STATE_DECLINED:
		stateString = "declined"
	case proto.Proposal_STATE_ENACTED:
		stateString = "enacted"
	case proto.Proposal_STATE_WAITING_FOR_NODE_VOTE:
		stateString = "is waiting for node vote"
	}
	return stateString
}

// AuctionNotification returns auction notification message
func AuctionNotification(dataClient api.TradingDataServiceClient, auction *proto.AuctionEvent) (string, error) {
	market, err := getMarketByID(dataClient, auction.MarketId)
	if err != nil {
		return "", err
	}

	status := "started"
	if auction.Leave {
		status = "ended"
	}

	auctionType := getAuctionType(auction.Trigger)
	message := "🔨 " + auctionType + " on " + market.TradableInstrument.Instrument.Name + " has been " + status

	return message, nil
}

func getAuctionType(trigger proto.AuctionTrigger) string {
	var auctionType string
	switch trigger {
	case proto.AuctionTrigger_AUCTION_TRIGGER_UNSPECIFIED:
		auctionType = "undefined"
	case proto.AuctionTrigger_AUCTION_TRIGGER_BATCH:
		auctionType = "Batch auction"
	case proto.AuctionTrigger_AUCTION_TRIGGER_OPENING:
		auctionType = "Opening auction"
	case proto.AuctionTrigger_AUCTION_TRIGGER_PRICE:
		auctionType = "Price monitoring auction"
	case proto.AuctionTrigger_AUCTION_TRIGGER_LIQUIDITY:
		auctionType = "Liquidity monitoring auction"
	}

	return auctionType
}

// NetworkParametesNotification returns network notification message
func NetworkParametesNotification(dataClient api.TradingDataServiceClient, network *proto.NetworkParameter, current *proto.NetworkParameter) string {
	var currentConfig EthereumConfig
	var newConfig EthereumConfig
	message := ""

	json.Unmarshal([]byte(current.Value), &currentConfig)
	json.Unmarshal([]byte(network.Value), &newConfig)

	if currentConfig.NetworkID != newConfig.NetworkID {
		message = "🔄 Vega network restarted. New network id is: " + newConfig.NetworkID
	}

	return message
}

// MarketCreationNotification returns market creation notification message
func MarketCreationNotification(dataClient api.TradingDataServiceClient, market *proto.Market) (string, error) {
	return "⚖️ A new market created for " + market.TradableInstrument.Instrument.Name, nil
}

// LossSocializationNotification returns loss socialization notification message
func LossSocializationNotification(dataClient api.TradingDataServiceClient, lossSocialization *proto.LossSocialization) (string, error) {
	market, err := getMarketByID(dataClient, lossSocialization.MarketId)
	if err != nil {
		return "", err
	}
	message := "💰 Loss socialization on " + market.TradableInstrument.Instrument.Name + ". Amount distributed: " + strconv.FormatInt(lossSocialization.Amount, 10)
	return message, nil
}

// RektNotification returns rekt notification message
func RektNotification(dataClient api.TradingDataServiceClient, trade *proto.Trade) (string, error) {
	market, err := getMarketByID(dataClient, trade.MarketId)
	if err != nil {
		return "", err
	}
	message := " 💸 A position on " + market.TradableInstrument.Instrument.Name + "has been liquidated. Position size: " + strconv.FormatUint(trade.Size, 10) + ", position price: " + strconv.FormatUint(trade.Price, 10)
	return message, nil
}

func getMarketByID(dataClient api.TradingDataServiceClient, marketID string) (*proto.Market, error) {
	requestMarket := api.MarketByIDRequest{MarketId: marketID}
	MarketObject, err := dataClient.MarketByID(context.Background(), &requestMarket)
	if err != nil {
		return nil, err
	}

	return MarketObject.GetMarket(), nil
}

// WhaleNotification return whale notification message
func WhaleNotification(dataClient api.TradingDataServiceClient, order *proto.Order) (string, error) {
	market, err := getMarketByID(dataClient, order.MarketId)
	if err != nil {
		return "", err
	}
	message := "🐋 Whale alert on " + market.TradableInstrument.Instrument.Name + ". order value: " + strconv.FormatUint(order.Size*order.Price, 10)

	return message, nil
}
