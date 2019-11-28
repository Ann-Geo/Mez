// Tests Edge server broker. Producer publishes images to edge node broker. Consumers subscribes images from edge server brokers.
// Test requires running the edge server, and edge node brokers (start_broker.go)

package broker

import (
	"fmt"
	"testing"
	"time"

	"github.com/arun-ravindran/Raven/api/edgenode"
	"github.com/arun-ravindran/Raven/api/edgeserver"
	"github.com/arun-ravindran/Raven/client"
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
		{"../../test_images/2.1M/", 800, 33, 10, 100, "NO", "S"},
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
