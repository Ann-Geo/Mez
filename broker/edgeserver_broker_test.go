// Tests Edge server broker. Producer publishes images to edge node broker. Consumers subscribes images from edge server brokers.
// Test requires running the edge server, and edge node brokers (start_broker.go)

package broker

import (
	"fmt"
	"testing"
	"time"

	"vsc_workspace/Mez_upload/api/edgenode"
	"vsc_workspace/Mez_upload/api/edgeserver"
	"vsc_workspace/Mez_upload/client"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

/*****************************Helper functions**********************/
func runProducerClient(t *testing.T) {
	producer := client.NewProducerClient("client", "edge") // Username password

	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	if err != nil {
		t.Errorf("could not load tls cert: %s", err)
	}

	// Connect test client with edge node broker
	conn, err := grpc.Dial("127.0.0.1:10000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&producer.Auth))
	if err != nil {
		t.Errorf("did not connect: %s", err)
	}
	defer conn.Close()
	cl := edgenode.NewPubSubClient(conn)
	err = producer.PublishImage(cl) // Test Publish API
	if err != nil {
		t.Errorf("Publish error %s", err)
	}

}

func runProducerClientTestESB(imageFilesPath string, numImagesInsert, frameRate uint64, imSizeParam string) error {

	producer := client.NewProducerClient("prodClient", "edge")

	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	if err != nil {
		return err
	}

	// Connect test client with edge node broker
	connProdClient, err := grpc.Dial("127.0.0.1:10000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&producer.Auth))
	if err != nil {
		return err
	}
	defer connProdClient.Close()
	cl := edgenode.NewPubSubClient(connProdClient)
	err = producer.PublishImageTestESB(cl, imageFilesPath, numImagesInsert, frameRate, imSizeParam)

	return err

}

func runConsumerClientTestFromESB(camid, latency, accuracy string, tStart, tStop string) (string, []string, []int, uint64) {
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
	connConsClient, err := grpc.Dial("127.0.0.1:20000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&consumer.Auth))
	if err != nil {
		return "consumer client could not connect with EN broker", tsSubscribed, imSizeSubscribed, numImagesRecvd
	}
	defer connConsClient.Close()
	cl := edgeserver.NewPubSubClient(connConsClient)
	subErrMsg, tsSubscribed, imSizeSubscribed, numImagesRecvd = consumer.SubscribeImageTestESB(cl, camid, latency, accuracy, tStart, tStop) // Test Subscribe API
	fmt.Println(numImagesRecvd)
	return subErrMsg, tsSubscribed, imSizeSubscribed, numImagesRecvd
}

/***********************************************************************/

func TestEdgeServerBroker(t *testing.T) {
	tbegin := time.Now()
	runProducerClient(t) // Publish to edgenode broker

	consumer := client.NewConsumerClient("client", "edge") //user name password

	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	if err != nil {
		t.Errorf("could not load tls cert: %s", err)
	}

	// Connect consumer application with edge server broker
	connEdgeserver, err := grpc.Dial("127.0.0.1:20000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&consumer.Auth))
	if err != nil {
		t.Errorf("did not connect: %s", err)
	}
	cl := edgeserver.NewPubSubClient(connEdgeserver)

	err = consumer.SubscribeImage(cl, tbegin) // Test Subscribe API
	if err != nil {
		t.Error(err)
	}
	consumer.Unsubscribe(cl)
}

func TestESBrokerSubscribe(t *testing.T) {
	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		frameRate       uint64
		logSize         uint64
		segSize         uint64
		fillType        string
		imSizeParam     string
	}{
		{"../../test_images/2.1M/", 1000, 33, 10, 100, "NO", "S"},
		{"../../test_images/2.1M/", 1500, 33, 10, 100, "O", "S"},
		{"../../test_images/1000_images/", 1000, 33, 10, 100, "NO", "V"},
	}

	for _, test := range tests {
		tStart := (time.Now()).Format(customTimeformat)
		err := runProducerClientTestESB(test.imageFilesPath, test.numImagesInsert, test.frameRate, test.imSizeParam)
		if err != nil {
			t.Fatalf("Publish failed - %v\n", err)
		}
		tStop := (time.Now()).Format(customTimeformat)
		latency := "100"
		accuracy := "1"
		camid := "cam1"

		subErrMsg, tsSubscribed, imSizeSubscribed, numImagesRecvd := runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
		if subErrMsg != "subscribe success" {
			t.Fatalf("Subscribe failed - %v\n", subErrMsg)
		}

		if test.fillType == "NO" {
			if !(numImagesRecvd == 1000) {
				t.Errorf("Num images published not equal to num images subscribed")
			}
		} else {
			if !(numImagesRecvd == test.logSize*test.segSize) {
				t.Errorf("Num images published not equal to num images subscribed")
			}
		}

		fmt.Println(len(tsSubscribed))
		fmt.Println(len(imSizeSubscribed))
		fmt.Println(numImagesRecvd)
	}
}

func TestESBSubscribeUnsubscribe(t *testing.T) {
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
		err := runProducerClientTestESB(test.imageFilesPath, test.numImagesInsert, test.frameRate, test.imSizeParam)
		if err != nil {
			t.Fatalf("Publish failed - %v\n", err)
		}
		tStop := (time.Now()).Format(customTimeformat)
		latency := "100"
		accuracy := "1"
		camid := "cam1"

		subErrMsg, tsSubscribed, imSizeSubscribed, numImagesRecvd := runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
		if subErrMsg != "subscribe success" {
			t.Fatalf("Subscribe failed - %v\n", subErrMsg)
		}

		fmt.Println(len(tsSubscribed))
		fmt.Println(len(imSizeSubscribed))
		fmt.Println(numImagesRecvd)

		if client.NumImRcvdUnsubTest == 500 {
			t.Errorf("Unsubscribe failed")
		}
	}
}

func TestESBMultiSubOnePub(t *testing.T) {
	type subReturn struct {
		subErrMsg        string
		tsSubscribed     []string
		imSizeSubscribed []int
		numImagesRecvd   uint64
	}

	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		frameRate       uint64
		logSize         uint64
		segSize         uint64
		fillType        string
		imSizeParam     string
	}{
		{"../../test_images/2.1M/", 100, 33, 10, 100, "NO", "S"},
		//{"../../test_images/2.1M/", 1500, 33, 10, 100, "O", "S"},
		//{"../../test_images/1000_images/", 1000, 33, 10, 100, "NO", "V"},
	}

	var sub1, sub2, sub3, sub4, sub5, sub6, sub7, sub8, sub9, sub10 subReturn

	for _, test := range tests {

		tStart := (time.Now()).Format(customTimeformat)
		err := runProducerClientTestESB(test.imageFilesPath, test.numImagesInsert, test.frameRate, test.imSizeParam)
		if err != nil {
			t.Fatalf("Publish failed - %v\n", err)
		}
		tStop := (time.Now()).Format(customTimeformat)
		latency := "100"
		accuracy := "1"
		camid := "cam1"

		go func() {

			sub1.subErrMsg, sub1.tsSubscribed, sub1.imSizeSubscribed, sub1.numImagesRecvd = runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub1.subErrMsg, test.fillType, sub1.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(5000 * time.Millisecond)
			sub2.subErrMsg, sub2.tsSubscribed, sub2.imSizeSubscribed, sub2.numImagesRecvd = runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub2.subErrMsg, test.fillType, sub2.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {
			time.Sleep(5000 * time.Millisecond)
			sub3.subErrMsg, sub3.tsSubscribed, sub3.imSizeSubscribed, sub3.numImagesRecvd = runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub3.subErrMsg, test.fillType, sub3.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()
		go func() {
			time.Sleep(1000 * time.Millisecond)
			sub4.subErrMsg, sub4.tsSubscribed, sub4.imSizeSubscribed, sub4.numImagesRecvd = runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub4.subErrMsg, test.fillType, sub4.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()
		go func() {
			time.Sleep(5000 * time.Millisecond)
			sub5.subErrMsg, sub5.tsSubscribed, sub5.imSizeSubscribed, sub5.numImagesRecvd = runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub5.subErrMsg, test.fillType, sub5.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()
		go func() {
			time.Sleep(5000 * time.Millisecond)
			sub6.subErrMsg, sub6.tsSubscribed, sub6.imSizeSubscribed, sub6.numImagesRecvd = runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub6.subErrMsg, test.fillType, sub6.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()
		go func() {
			time.Sleep(5000 * time.Millisecond)
			sub7.subErrMsg, sub7.tsSubscribed, sub7.imSizeSubscribed, sub7.numImagesRecvd = runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub7.subErrMsg, test.fillType, sub7.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()
		go func() {
			time.Sleep(5000 * time.Millisecond)
			sub8.subErrMsg, sub8.tsSubscribed, sub8.imSizeSubscribed, sub8.numImagesRecvd = runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub8.subErrMsg, test.fillType, sub8.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()
		go func() {
			time.Sleep(5000 * time.Millisecond)
			sub9.subErrMsg, sub9.tsSubscribed, sub9.imSizeSubscribed, sub9.numImagesRecvd = runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub9.subErrMsg, test.fillType, sub9.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()
		go func() {
			time.Sleep(5000 * time.Millisecond)
			sub10.subErrMsg, sub10.tsSubscribed, sub10.imSizeSubscribed, sub10.numImagesRecvd = runConsumerClientTestFromESB(camid, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub10.subErrMsg, test.fillType, sub10.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		time.Sleep(30000 * time.Millisecond)

		//time.Sleep(38000 * time.Millisecond)

	}
}

func multiSubErrorCheck(errMsg, fillType string, numImagesRecvd, logSize, segSize uint64) string {

	if errMsg != "subscribe success" {
		return errMsg
	}

	if fillType == "NO" {
		if !(numImagesRecvd == 10) {
			return "Num images published not equal to num images subscribed"
		}
	} else {
		if !(numImagesRecvd == logSize*segSize) {
			return "Num images published not equal to num images subscribed"
		}
	}

	return "subscribe success"

}

func TestESBOneSubMultiPub(t *testing.T) {
	type subReturn struct {
		subErrMsg        string
		tsSubscribed     []string
		imSizeSubscribed []int
		numImagesRecvd   uint64
	}

	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		frameRate       uint64
		logSize         uint64
		segSize         uint64
		fillType        string
		imSizeParam     string
		numProd         uint64
	}{
		{"../../test_images/2.1M/", 100, 33, 10, 100, "NO", "S", 2},
		//{"../../test_images/2.1M/", 1500, 33, 10, 100, "O", "S"},
		//{"../../test_images/1000_images/", 1000, 33, 10, 100, "NO", "V"},
	}

	var subCam1, subCam2, subCam3, subCam4 subReturn

	for _, test := range tests {

		tStart := (time.Now()).Format(customTimeformat)
		err := runMultiProducerClientTestESB(test.imageFilesPath, test.numImagesInsert, test.frameRate, test.imSizeParam, test.numProd)
		if err != nil {
			t.Fatalf("Publish failed - %v\n", err)
		}
		tStop := (time.Now()).Format(customTimeformat)
		latency := "100"
		accuracy := "1"
		camid1 := "cam1"
		camid2 := "cam2"
		camid3 := "cam3"
		camid4 := "cam4"

		go func() {

			subCam1.subErrMsg, subCam1.tsSubscribed, subCam1.imSizeSubscribed, subCam1.numImagesRecvd = runConsumerClientTestFromESB(camid1, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(subCam1.subErrMsg, test.fillType, subCam1.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(1000 * time.Millisecond)
			subCam2.subErrMsg, subCam2.tsSubscribed, subCam2.imSizeSubscribed, subCam2.numImagesRecvd = runConsumerClientTestFromESB(camid2, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(subCam2.subErrMsg, test.fillType, subCam2.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(1000 * time.Millisecond)
			subCam3.subErrMsg, subCam3.tsSubscribed, subCam3.imSizeSubscribed, subCam3.numImagesRecvd = runConsumerClientTestFromESB(camid3, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(subCam3.subErrMsg, test.fillType, subCam3.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(1000 * time.Millisecond)
			subCam4.subErrMsg, subCam4.tsSubscribed, subCam4.imSizeSubscribed, subCam4.numImagesRecvd = runConsumerClientTestFromESB(camid4, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(subCam4.subErrMsg, test.fillType, subCam4.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		time.Sleep(30000 * time.Millisecond)

		//time.Sleep(38000 * time.Millisecond)

	}
}

func TestESBMultiSubMultiPub(t *testing.T) {
	type subReturn struct {
		subErrMsg        string
		tsSubscribed     []string
		imSizeSubscribed []int
		numImagesRecvd   uint64
	}

	var tests = []struct {
		imageFilesPath  string
		numImagesInsert uint64
		frameRate       uint64
		logSize         uint64
		segSize         uint64
		fillType        string
		imSizeParam     string
		numProd         uint64
	}{
		{"../../test_images/2.1M/", 10, 33, 10, 100, "NO", "S", 2},
		//{"../../test_images/2.1M/", 1500, 33, 10, 100, "O", "S"},
		//{"../../test_images/1000_images/", 1000, 33, 10, 100, "NO", "V"},
	}

	var subCam1, subCam2, subCam3, subCam4, sub1Cam1, sub1Cam2, sub1Cam3, sub1Cam4 subReturn

	for _, test := range tests {

		tStart := (time.Now()).Format(customTimeformat)
		err := runMultiProducerClientTestESB(test.imageFilesPath, test.numImagesInsert, test.frameRate, test.imSizeParam, test.numProd)
		if err != nil {
			t.Fatalf("Publish failed - %v\n", err)
		}
		tStop := (time.Now()).Format(customTimeformat)
		latency := "100"
		accuracy := "1"
		camid1 := "cam1"
		camid2 := "cam2"
		camid3 := "cam3"
		camid4 := "cam4"

		go func() {

			subCam1.subErrMsg, subCam1.tsSubscribed, subCam1.imSizeSubscribed, subCam1.numImagesRecvd = runConsumerClientTestFromESB(camid1, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(subCam1.subErrMsg, test.fillType, subCam1.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(1000 * time.Millisecond)
			sub1Cam1.subErrMsg, sub1Cam1.tsSubscribed, sub1Cam1.imSizeSubscribed, sub1Cam1.numImagesRecvd = runConsumerClientTestFromESB(camid1, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub1Cam1.subErrMsg, test.fillType, sub1Cam1.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(2000 * time.Millisecond)
			subCam2.subErrMsg, subCam2.tsSubscribed, subCam2.imSizeSubscribed, subCam2.numImagesRecvd = runConsumerClientTestFromESB(camid2, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(subCam2.subErrMsg, test.fillType, subCam2.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(1000 * time.Millisecond)
			sub1Cam2.subErrMsg, sub1Cam2.tsSubscribed, sub1Cam2.imSizeSubscribed, sub1Cam2.numImagesRecvd = runConsumerClientTestFromESB(camid2, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub1Cam2.subErrMsg, test.fillType, sub1Cam2.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(3000 * time.Millisecond)
			subCam3.subErrMsg, subCam3.tsSubscribed, subCam3.imSizeSubscribed, subCam3.numImagesRecvd = runConsumerClientTestFromESB(camid3, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(subCam3.subErrMsg, test.fillType, subCam3.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(1000 * time.Millisecond)
			sub1Cam3.subErrMsg, sub1Cam3.tsSubscribed, sub1Cam3.imSizeSubscribed, sub1Cam3.numImagesRecvd = runConsumerClientTestFromESB(camid3, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub1Cam3.subErrMsg, test.fillType, sub1Cam3.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(4000 * time.Millisecond)
			subCam4.subErrMsg, subCam4.tsSubscribed, subCam4.imSizeSubscribed, subCam4.numImagesRecvd = runConsumerClientTestFromESB(camid4, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(subCam4.subErrMsg, test.fillType, subCam4.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		go func() {

			time.Sleep(1000 * time.Millisecond)
			sub1Cam4.subErrMsg, sub1Cam4.tsSubscribed, sub1Cam4.imSizeSubscribed, sub1Cam4.numImagesRecvd = runConsumerClientTestFromESB(camid4, latency, accuracy, tStart, tStop)
			errMsg := multiSubErrorCheck(sub1Cam4.subErrMsg, test.fillType, sub1Cam4.numImagesRecvd, test.logSize, test.segSize)
			if errMsg != "subscribe success" {
				t.Fatalf(errMsg)
			}
		}()

		time.Sleep(40000 * time.Millisecond)

		//time.Sleep(38000 * time.Millisecond)

	}
}

func runMultiProducerClientTestESB(imageFilesPath string, numImagesInsert, frameRate uint64, imSizeParam string, numProd uint64) error {

	producer1 := client.NewProducerClient("prodClient1", "edge1")
	producer2 := client.NewProducerClient("prodClient2", "edge2")
	producer3 := client.NewProducerClient("prodClient3", "edge3")
	producer4 := client.NewProducerClient("prodClient4", "edge4")

	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	if err != nil {
		return err
	}

	// Connect test client with edge node broker
	connProdClient1, err := grpc.Dial("127.0.0.1:10000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&producer1.Auth))
	if err != nil {
		return err
	}
	defer connProdClient1.Close()
	cl1 := edgenode.NewPubSubClient(connProdClient1)

	err = producer1.PublishImageTestESB(cl1, imageFilesPath, numImagesInsert, frameRate, imSizeParam)

	// Connect test client with edge node broker
	connProdClient2, err := grpc.Dial("127.0.0.1:30000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&producer2.Auth))
	if err != nil {
		return err
	}
	defer connProdClient2.Close()
	cl2 := edgenode.NewPubSubClient(connProdClient2)
	err = producer2.PublishImageTestESB(cl2, imageFilesPath, numImagesInsert, frameRate, imSizeParam)

	// Connect test client with edge node broker
	connProdClient3, err := grpc.Dial("127.0.0.1:40000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&producer3.Auth))
	if err != nil {
		return err
	}
	defer connProdClient3.Close()
	cl3 := edgenode.NewPubSubClient(connProdClient3)
	err = producer3.PublishImageTestESB(cl3, imageFilesPath, numImagesInsert, frameRate, imSizeParam)

	// Connect test client with edge node broker
	connProdClient4, err := grpc.Dial("127.0.0.1:50000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&producer4.Auth))
	if err != nil {
		return err
	}
	defer connProdClient4.Close()
	cl4 := edgenode.NewPubSubClient(connProdClient4)
	err = producer4.PublishImageTestESB(cl4, imageFilesPath, numImagesInsert, frameRate, imSizeParam)

	return err

}
