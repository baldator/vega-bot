package main

import (
	"fmt"
	"io"
	"log"

	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto"
	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {

	conn, err := grpc.Dial("n06.testnet.vega.xyz:3002", grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(256<<20)))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	dataClient := api.NewTradingDataServiceClient(conn)
	eventType := []proto.BusEventType{
		proto.BusEventType_BUS_EVENT_TYPE_NETWORK_PARAMETER,
		proto.BusEventType_BUS_EVENT_TYPE_LOSS_SOCIALIZATION,
		proto.BusEventType_BUS_EVENT_TYPE_AUCTION,
		proto.BusEventType_BUS_EVENT_TYPE_PROPOSAL,
		//proto.BusEventType_BUS_EVENT_TYPE_TRADE,
		//proto.BusEventType_BUS_EVENT_TYPE_ORDER,
	}
	events, err := dataClient.ObserveEventBus(context.Background())

	done := make(chan bool)
	go func() {
		for {
			resp, err := events.Recv()

			if err == io.EOF {
				// read done.
				log.Println("gRPC EOF error")
				//close(done)
				return
			}

			if err != nil {
				log.Fatal(err)
			}

			for _, event := range resp.Events {
				log.Println(event)
			}
		}
	}()

	// When the batchSize is too small -> "rpc error: code = Unknown desc = EOF"
	observerEvent := api.ObserveEventBusRequest{Type: eventType, BatchSize: 10000}
	events.Send(&observerEvent)
	events.CloseSend()

	fmt.Println("OK")

	<-done //we will wait until all response is received

	log.Println("finished")
}
