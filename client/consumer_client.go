// Consumer client (Analytics application)
package client

import (
	"github.com/arun-ravindran/Raven/api/edgeserver"
	"context"
	"io"
	"log"
	"time"
)

type ConsumerClient struct {
	Auth Authentication
}

func NewConsumerClient(login, password string) *ConsumerClient {
	return &ConsumerClient{Auth: Authentication{login: login, password: password}}
}

func (cc *ConsumerClient) SubscribeImage(client edgeserver.PubSubClient, tbegin time.Time) error {
	// Server side streaming
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := tbegin.Format(time.RFC850)
	stop := tbegin.Add(5 * time.Second).Format(time.RFC850)

	imPars := &edgeserver.ImageStreamParameters{Camid: "cam1", Latency: "100", Accuracy: "100",
		Start: start, Stop: stop}
	// Open a stream to gRPC server
	stream, err := client.Subscribe(ctx, imPars)
	if err != nil {
		return err
	}

	numImagesRecvd := 0
	for {
		im, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		numImagesRecvd++
		log.Printf("Consumer client: Number of images received %d, of size %d, and timestamp %s", numImagesRecvd, len(im.GetImage()), im.GetTimestamp())

		// {Here would be application code to process image received}

	}
	return nil
}

func (cc *ConsumerClient) Unsubscribe(client edgeserver.PubSubClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	appInfo := &edgeserver.AppInfo{Appid: "app1", Camid: "cam1"}
	_, err := client.Unsubscribe(ctx, appInfo)
	if err != nil {
		log.Fatalf("Unsubscribe failed")
	}
}
