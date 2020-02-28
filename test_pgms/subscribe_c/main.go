package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"vsc_workspace/Mez_upload_woa/api/edgeserver"
	"vsc_workspace/Mez_upload_woa/client"
)

var customTimeformat string = "Monday, 02-Jan-06 15:04:05.00000 MST"

func main() {

	tStartUnformat := time.Now()
	tStart := (tStartUnformat).Format(customTimeformat)
	fmt.Println(tStart)
	time.Sleep(5 * time.Second)

	//create a consumer client with login and password
	consumer := client.NewConsumerClient("client", "edge") //user name password

	//Connect API, specify ES broker url and user address
	err := consumer.Connect("127.0.0.1:20000", "127.0.0.1:9051")
	if err != nil {
		log.Fatalf("error while calling Connect")
	}

	defer consumer.ConnConsClient.Close()
	defer consumer.Cancel()

	/**************************sub parameters***************************************/
	tStopUnformat := tStartUnformat.Add(time.Second * 30)
	tStop := (tStopUnformat).Format(customTimeformat)
	fmt.Println(tStop)
	appid := "sub1"
	latency := "1"
	accuracy := "0.33 duke medium"
	camid := "cam1"
	/******************************************************************************/

	imPars := &edgeserver.ImageStreamParameters{Appid: appid, Camid: camid, Latency: latency, Accuracy: accuracy,
		Start: tStart, Stop: tStop}

	fmt.Println("for meas", time.Now())

	//subscibe API call
	stream, err := consumer.Cl.Subscribe(consumer.Ctx, imPars)

	if err != nil {
		log.Fatalln("error while subscribing")
	}

	resultFile, err := os.Create("total_pubsub_lat_raven.txt")
	if err != nil {
		log.Fatalf("Cannot create result file %v\n", err)
	}

	defer resultFile.Close()

	//start receiving files
	for {
		im, err := stream.Recv()
		trecvd := time.Now()
		fmt.Println(trecvd)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("error while receiving stream from Subscribe")
		}

		log.Printf("Image of size %d received with timestamp %s", len(im.GetImage()), im.GetTimestamp())
		ts, _ := time.Parse(customTimeformat, im.GetTimestamp())
		fmt.Fprintf(resultFile, "raven latency: %s\n", trecvd.Sub(ts))
	}

	fmt.Println("done", time.Now())

	appInfo := &edgeserver.AppInfo{Appid: appid, Camid: camid}

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	status, _ := consumer.Cl.Unsubscribe(ctx, appInfo)
	if status.GetStatus() == false {
		log.Fatalf("Could not Unsubscibe")
	}

	fmt.Println("unsubscribed from ")
}
