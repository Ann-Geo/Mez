package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"
	"vsc_workspace/Mez_upload/api/edgeserver"
	"vsc_workspace/Mez_upload/client"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var customTimeformat string = "Monday, 02-Jan-06 15:04:05.00000 MST"

func main() {

	tStart := (time.Now()).Format(customTimeformat)
	fmt.Println(tStart)

	time.Sleep(30 * time.Second)

	consumer := client.NewConsumerClient("consClient", "edge") //user name password

	creds, err := credentials.NewClientTLSFromFile("../../cert/server.crt", "")
	if err != nil {
		log.Fatalln("tls error")
	}

	// Connect consumer application with edge server broker
	connConsClient, err := grpc.Dial("127.0.0.1:20000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&consumer.Auth))
	if err != nil {
		log.Fatalln("could not connect with ESB")
	}
	defer connConsClient.Close()
	cl := edgeserver.NewPubSubClient(connConsClient)

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	/**************************sub parameters***************************************/

	tStop := (time.Now()).Format(customTimeformat)
	fmt.Println(tStop)
	latency := "1"
	accuracy := "0.47 jaad simple"
	camid := "cam1"
	/******************************************************************************/

	imPars := &edgeserver.ImageStreamParameters{Camid: camid, Latency: latency, Accuracy: accuracy,
		Start: tStart, Stop: tStop}

	stream, err := cl.Subscribe(ctx, imPars)

	if err != nil {
		log.Fatalln("error while subscribing")
	}

	//appInfo := &edgeserver.AppInfo{Camid: camid, Appid: "1"}

	for {
		im, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("error while receiving stream from Subscribe")
		}

		log.Printf("Image of size %d received with timestamp %s", len(im.GetImage()), im.GetTimestamp())

	}

}
