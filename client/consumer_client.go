// Consumer client (Analytics application)
package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/arun-ravindran/Raven/api/edgenode"
	"github.com/arun-ravindran/Raven/api/edgeserver"
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

	start := tbegin.Format(customTimeformat)
	stop := tbegin.Add(5 * time.Second).Format(customTimeformat)

	imPars := &edgeserver.ImageStreamParameters{Camid: "cam1", Latency: "100", Accuracy: "100",
		Start: start, Stop: stop}
	// Open a stream to gRPC server
	stream, err := client.Subscribe(ctx, imPars)
	fmt.Println(stream)
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

func (cc *ConsumerClient) SubscribeImageTest(client edgenode.PubSubClient, camid, latency, accuracy, tStart, tStop string) (string, []string, []int, uint64) {

	tsSubscribed := make([]string, 0)
	imSizeSubscribed := make([]int, 0)
	var numImagesRecvd uint64

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	imPars := &edgenode.ImageStreamParameters{Camid: "cam1", Latency: "100", Accuracy: "100",
		Start: tStart, Stop: tStop}

	stream, err := client.Subscribe(ctx, imPars)
	if err != nil {
		return "error while invoking Subscribe", tsSubscribed, imSizeSubscribed, numImagesRecvd
	}

	for {
		im, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "error while receiving stream from Subscribe", tsSubscribed, imSizeSubscribed, numImagesRecvd
		}
		numImagesRecvd++
		ts := im.GetTimestamp()
		imSize := len(im.GetImage())
		tsSubscribed = append(tsSubscribed, ts)
		imSizeSubscribed = append(imSizeSubscribed, imSize)

	}

	return "subscribe success", tsSubscribed, imSizeSubscribed, numImagesRecvd
}

func (cc *ConsumerClient) SubscribeImageTestConcurrent(client edgenode.PubSubClient, camid, latency, accuracy, tStart, tStop string) {

	subErrMsg := "subscribe success"
	tsSubscribed := make([]string, 0)
	imSizeSubscribed := make([]int, 0)
	var numImagesRecvd uint64

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	imPars := &edgenode.ImageStreamParameters{Camid: "cam1", Latency: "100", Accuracy: "100",
		Start: tStart, Stop: tStop}

	stream, err := client.Subscribe(ctx, imPars)
	if err != nil {
		//return "error while invoking Subscribe", tsSubscribed, imSizeSubscribed, numImagesRecvd
		subErrMsg = "error while invoking Subscribe"
	}

	for {
		im, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			//return "error while receiving stream from Subscribe", tsSubscribed, imSizeSubscribed, numImagesRecvd
			subErrMsg = "error while receiving stream from Subscribe"
		}
		numImagesRecvd++
		ts := im.GetTimestamp()
		fmt.Println(ts)
		imSize := len(im.GetImage())
		tsSubscribed = append(tsSubscribed, ts)
		imSizeSubscribed = append(imSizeSubscribed, imSize)

	}

	fmt.Println(numImagesRecvd)
	//return "subscribe success", tsSubscribed, imSizeSubscribed, numImagesRecvd
	fmt.Println(subErrMsg)
}

func (cc *ConsumerClient) SubscribeImageTestESB(client edgeserver.PubSubClient, camid, latency, accuracy,
	tStart, tStop string) (string, []string, []int, uint64) {

	tsSubscribed := make([]string, 0)
	imSizeSubscribed := make([]int, 0)
	var numImagesRecvd uint64

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	imPars := &edgeserver.ImageStreamParameters{Camid: "cam1", Latency: "100", Accuracy: "100",
		Start: tStart, Stop: tStop}

	stream, err := client.Subscribe(ctx, imPars)

	if err != nil {
		return "error while invoking Subscribe", tsSubscribed, imSizeSubscribed, numImagesRecvd
	}

	for {
		im, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "error while receiving stream from Subscribe", tsSubscribed, imSizeSubscribed, numImagesRecvd
		}
		numImagesRecvd++
		log.Printf("Consumer client: Number of images received %d, of size %d, and timestamp %s", numImagesRecvd, len(im.GetImage()), im.GetTimestamp())
		ts := im.GetTimestamp()
		imSize := len(im.GetImage())
		tsSubscribed = append(tsSubscribed, ts)
		imSizeSubscribed = append(imSizeSubscribed, imSize)

	}

	return "subscribe success", tsSubscribed, imSizeSubscribed, numImagesRecvd
}
