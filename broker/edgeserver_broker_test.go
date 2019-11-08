// Tests Edge server broker. Producer publishes images to edge node broker. Consumers subscribes images from edge server brokers.
// Test requires running the edge server, and edge node brokers (start_broker.go)

package broker

import (
	"testing"
	"time"

	"github.com/arun-ravindran/Raven/api/edgenode"
	"github.com/arun-ravindran/Raven/api/edgeserver"
	"github.com/arun-ravindran/Raven/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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
