package broker

import (
	"fmt"
	"testing"
	"time"

	"github.com/arun-ravindran/Raven/api/edgenode"
	"github.com/arun-ravindran/Raven/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

/***************Helper functions start*********************************/

func runProducerClientTest(imageFilesPath string, numImagesInsert, frameRate uint64, imSizeParam string) (string, []string, []int, uint64) {
	var tsPublished []string
	var imSizePublished []int
	var numPublished uint64
	publishErrMsg := "publish success"
	producer := client.NewProducerClient("prodClient", "edge")

	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	if err != nil {
		return "could not load tls cert - producer", tsPublished, imSizePublished, numPublished
	}

	// Connect test client with edge node broker
	connProdClient, err := grpc.Dial("127.0.0.1:10000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&producer.Auth))
	if err != nil {
		return "producer client could not connect with EN broker", tsPublished, imSizePublished, numPublished
	}
	defer connProdClient.Close()
	cl := edgenode.NewPubSubClient(connProdClient)
	publishErrMsg, tsPublished, imSizePublished, numPublished = producer.PublishImageTest(cl, imageFilesPath, numImagesInsert, frameRate, imSizeParam)

	return publishErrMsg, tsPublished, imSizePublished, numPublished

}

func runConsumerClientTest(camid, latency, accuracy string, tStart, tStop string) (string, []string, []int, uint64) {
	subErrMsg := "subscribe success"
	tsSubscribed := make([]string, 0)
	imSizeSubscribed := make([]int, 0)
	var numImagesRecvd uint64
	consumer := client.NewConsumerClient("consClient", "edge") //user name password

	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	if err != nil {
		return "could not load tls cert - consumer", tsSubscribed, imSizeSubscribed, numImagesRecvd
	}

	// Connect consumer application with edge server broker
	connConsClient, err := grpc.Dial("127.0.0.1:10000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&consumer.Auth))
	if err != nil {
		return "consumer client could not connect with EN broker", tsSubscribed, imSizeSubscribed, numImagesRecvd
	}
	defer connConsClient.Close()
	cl := edgenode.NewPubSubClient(connConsClient)
	subErrMsg, tsSubscribed, imSizeSubscribed, numImagesRecvd = consumer.SubscribeImageTest(cl, camid, latency, accuracy, tStart, tStop) // Test Subscribe API
	return subErrMsg, tsSubscribed, imSizeSubscribed, numImagesRecvd
}

/*To compare equality of timestamp slices (Appended and Read)
and to compare equality of image size slices (Appended and Read)
Input - Appended and Read slices
Output- true or false boolean
*/
func sliceEquality(tsAppended, tsRead []string, imSizeAppended, imSizeRead []int) (string, bool) {

	var errMsg string
	var status bool = true
	if len(tsAppended) != len(tsRead) {
		errMsg = "timestamp logs length mismatch"
		status = false

	}

	if len(imSizeAppended) != len(imSizeRead) {
		errMsg = "image logs length mismatch"
		status = false

	}

	for i := 0; i < len(tsAppended); i++ {
		if tsAppended[i] != tsRead[i] {
			errMsg = "timestamp logs contents mismatch"
			status = false

		}
		if imSizeAppended[i] != imSizeRead[i] {
			errMsg = "image logs contents mismatch"
			status = false

		}
	}

	if status == false {
		return errMsg, status
	} else {
		return "slices are equal", true
	}

}

/*****************************Helper functions end********************************************/
/******************************EN broker tests start*****************************************/

func TestSubscribeAfterPublish(t *testing.T) {

	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		frameRate       uint64
		logSize         uint64
		segSize         uint64
		fillType        string
		imSizeParam     string
	}{
		{"../../test_images/2.1M/", 800, 33, 10, 100, "NO", "S"},
		{"../../test_images/2.1M/", 1500, 33, 10, 100, "O", "S"},
		{"../../test_images/1000_images/", 1000, 33, 10, 100, "NO", "V"},
	}

	for _, test := range tests {
		tStart := (time.Now()).Format(customTimeformat)
		publishErrMsg, tsPublished, imSizePublished, numPublished := runProducerClientTest(test.imageFilesPath, test.numImagesInsert,
			test.frameRate, test.imSizeParam)
		if publishErrMsg != "publish success" {
			t.Fatalf("Publish failed - %v\n", publishErrMsg)
		}
		tStop := (time.Now()).Format(customTimeformat)
		latency := "100"
		accuracy := "1"
		camid := "cam1"

		subErrMsg, tsSubscribed, imSizeSubscribed, numImagesRecvd := runConsumerClientTest(camid, latency, accuracy, tStart, tStop)
		if subErrMsg != "subscribe success" {
			t.Fatalf("Subscribe failed - %v\n", subErrMsg)
		}

		if test.fillType == "NO" {
			if !(numPublished == numImagesRecvd) {
				t.Errorf("Num images published not equal to num images subscribed")
			} else {
				testErrMsg, testStatus := sliceEquality(tsPublished, tsSubscribed, imSizePublished, imSizeSubscribed)
				if !testStatus {
					t.Errorf("Subscribe after Publish fail - %v\n", testErrMsg)
				}
			}
		} else {
			if !(numImagesRecvd == test.logSize*test.segSize) {
				t.Errorf("Num images published not equal to num images subscribed")
			}
		}

	}

}

func TestSubscribeUnsubscribe(t *testing.T) {

	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		frameRate       uint64
		logSize         uint64
		segSize         uint64
		fillType        string
		imSizeParam     string
	}{
		{"../../test_images/2.1M/", 800, 33, 10, 100, "NO", "S"},
	}

	for _, test := range tests {
		tStart := (time.Now()).Format(customTimeformat)
		runProducerClientTest(test.imageFilesPath, test.numImagesInsert,
			test.frameRate, test.imSizeParam)

		tStop := (time.Now()).Format(customTimeformat)
		latency := "100"
		accuracy := "1"
		camid := "cam1"

		runConsumerClientTestConcurrent(camid, latency, accuracy, tStart, tStop)

		if client.NumImRcvdUnsubTest == 500 {
			t.Errorf("Unsubscribe failed")
		}

	}

}

func TestPublishSubscribeConcurrent(t *testing.T) {
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		frameRate       uint64
		logSize         uint64
		segSize         uint64
		fillType        string
		imSizeParam     string
	}{
		{"../../test_images/2.1M/", 800, 33, 10, 100, "NO", "S"},
		{"../../test_images/2.1M/", 1500, 33, 10, 100, "O", "S"},
		{"../../test_images/1000_images/", 1000, 33, 10, 100, "NO", "V"},
	}

	for _, test := range tests {
		//publishErrMsg := "publish success"
		//subErrMsg := "subscribe success"
		tStart := time.Now()
		tStartFmtd := tStart.Format(customTimeformat)

		go runProducerClientTestConcurrent(test.imageFilesPath, test.numImagesInsert, test.frameRate, test.imSizeParam)
		/*
			if publishErrMsg != "publish success" {
				t.Fatalf("Publish failed - %v\n", publishErrMsg)
			}
		*/
		time.Sleep(100 * time.Millisecond)
		tStop := tStart.Add(time.Millisecond * 17000)
		tStopFmtd := tStop.Format(customTimeformat)
		latency := "100"
		accuracy := "1"
		camid := "cam1"
		fmt.Println("tStart --", tStartFmtd)
		fmt.Println("tStop --", tStopFmtd)

		go runConsumerClientTestConcurrent(camid, latency, accuracy, tStartFmtd, tStopFmtd)
		/*
			if subErrMsg != "subscribe success" {
				t.Fatalf("Subscribe failed - %v\n", subErrMsg)
			}
		*/
		time.Sleep(20 * time.Second)
	}

}

func runProducerClientTestConcurrent(imageFilesPath string, numImagesInsert, frameRate uint64, imSizeParam string) {
	//var tsPublished []string
	//var imSizePublished []int
	//var numPublished uint64
	publishErrMsg := "publish success"
	producer := client.NewProducerClient("prodClient", "edge")

	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	if err != nil {
		//return "could not load tls cert - producer", tsPublished, imSizePublished, numPublished
		publishErrMsg = "could not load tls cert - producer"
		//return publishErrMsg
	}

	// Connect test client with edge node broker
	connProdClient, err := grpc.Dial("127.0.0.1:10000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&producer.Auth))
	if err != nil {
		//return "producer client could not connect with EN broker", tsPublished, imSizePublished, numPublished
		publishErrMsg = "producer client could not connect with EN broker"
	}
	defer connProdClient.Close()
	cl := edgenode.NewPubSubClient(connProdClient)
	//publishErrMsg, tsPublished, imSizePublished, numPublished =
	producer.PublishImageTestConcurrent(cl, imageFilesPath, numImagesInsert, frameRate, imSizeParam)

	fmt.Println(publishErrMsg)

	//return publishErrMsg, tsPublished, imSizePublished, numPublished

}

func runConsumerClientTestConcurrent(camid, latency, accuracy string, tStart, tStop string) {
	subErrMsg := "subscribe success"
	//tsSubscribed := make([]string, 0)
	//imSizeSubscribed := make([]int, 0)
	//var numImagesRecvd uint64
	consumer := client.NewConsumerClient("consClient", "edge") //user name password

	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	if err != nil {
		//return "could not load tls cert - consumer", tsSubscribed, imSizeSubscribed, numImagesRecvd
		subErrMsg = "could not load tls cert - consumer"
	}

	// Connect consumer application with edge server broker
	connConsClient, err := grpc.Dial("127.0.0.1:10000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&consumer.Auth))
	if err != nil {
		//return "consumer client could not connect with EN broker", tsSubscribed, imSizeSubscribed, numImagesRecvd
		subErrMsg = "consumer client could not connect with EN broker"
	}
	defer connConsClient.Close()
	cl := edgenode.NewPubSubClient(connConsClient)
	//subErrMsg, tsSubscribed, imSizeSubscribed, numImagesRecvd =
	consumer.SubscribeImageTestConcurrent(cl, camid, latency, accuracy, tStart, tStop) // Test Subscribe API
	//return subErrMsg, tsSubscribed, imSizeSubscribed, numImagesRecvd
	fmt.Println(subErrMsg)

}
