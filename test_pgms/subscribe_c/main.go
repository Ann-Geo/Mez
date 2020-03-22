package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Ann-Geo/Mez/api/edgeserver"
	"github.com/Ann-Geo/Mez/client"
)

var customTimeformat string = "Monday, 02-Jan-06 15:04:05.00000 MST"

func main() {

	tStartUnformat := time.Now()
	tStart := (tStartUnformat).Format(customTimeformat)
	fmt.Println(tStart)

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
	latency := "10"
	accuracy := "0.40 jaad simple"
	camid := "cam1"
	/******************************************************************************/

	imPars := &edgeserver.ImageStreamParameters{Appid: appid, Camid: camid, Latency: latency, Accuracy: accuracy,
		Start: tStart, Stop: tStop}

	fmt.Println("for meas", time.Now())

	//subscibe API call
	fmt.Println("subscribe started: ", time.Now())
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
			numIter := 5
			for n := 0; n < numIter; n++ {
				time.Sleep(150 * time.Millisecond)
				if err == nil {
					break
				}
			}
			if err != nil {
				log.Fatalln("connection timeout: could not reach Edge server broker", err)
			}
		}

		log.Printf("Image of size %d received with timestamp %s", len(im.GetImage()), im.GetTimestamp())

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
