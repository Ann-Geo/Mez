package main

import (
	"fmt"
	"io"
	"log"
	"time"
	"vsc_workspace/Mez_upload/api/edgeserver"
	"vsc_workspace/Mez_upload/client"
)

var customTimeformat string = "Monday, 02-Jan-06 15:04:05.00000 MST"

func main() {

	tStart := (time.Now()).Format(customTimeformat)
	fmt.Println(tStart)
	time.Sleep(30 * time.Second)

	//create a consumer client with login and password
	consumer := client.NewConsumerClient("client", "edge") //user name password

	//Connect API, specify EN broker url and user address
	err := consumer.Connect("127.0.0.1:20000", "127.0.0.1:9051")
	if err != nil {
		log.Fatalf("error while calling Connect")
	}

	defer consumer.ConnConsClient.Close()
	defer consumer.Cancel()

	/**************************sub parameters***************************************/
	tStop := (time.Now()).Format(customTimeformat)
	fmt.Println(tStop)
	latency := "1"
	accuracy := "0.33 duke medium"
	camid := "cam1"
	/******************************************************************************/

	imPars := &edgeserver.ImageStreamParameters{Camid: camid, Latency: latency, Accuracy: accuracy,
		Start: tStart, Stop: tStop}

	fmt.Println("for meas", time.Now())

	//subscibe API call
	stream, err := consumer.Cl.Subscribe(consumer.Ctx, imPars)

	if err != nil {
		log.Fatalln("error while subscribing")
	}

	//start receiving files
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

	fmt.Println("done", time.Now())

}
