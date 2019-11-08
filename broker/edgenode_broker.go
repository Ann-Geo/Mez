package broker

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/arun-ravindran/Raven/api/edgenode"
	"github.com/arun-ravindran/Raven/api/edgeserver"
	"github.com/arun-ravindran/Raven/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type EdgeNodeBroker struct {
	serverName      string
	ipaddr          string
	mutex           sync.Mutex
	store           map[string]storage.Store
	stopSubcription bool
}

func NewEdgeNodeBroker(sname, ipaddr string) *EdgeNodeBroker {
	return &EdgeNodeBroker{
		serverName:      sname,
		ipaddr:          ipaddr,
		mutex:           sync.Mutex{},
		store:           make(map[string]storage.Store),
		stopSubcription: false,
	}
}

func (s *EdgeNodeBroker) StartEdgeNodeBroker(edgeServerIpaddr, login, password string) error {
	log.Println("Starting edge node broker", s.serverName)
	lis, err := net.Listen("tcp", s.ipaddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile("cert/server.crt", "cert/server.key")
	if err != nil {
		return fmt.Errorf("could not load TLS keys: %s", err)
	}

	// Create default log storage
	s.store[s.serverName] = storage.NewMemLog(storage.SEGSIZE, storage.LOGSIZE)

	// Register with edge server
	err = s.regsterWithEdgeServer(edgeServerIpaddr, login, password)
	if err != nil {
		return fmt.Errorf("EdgeNodeBroker %s\n", err)
	}

	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{grpc.Creds(creds), grpc.UnaryInterceptor(s.UnaryInterceptor)}

	grpcServer := grpc.NewServer(opts...)
	// Attach client API to broker
	edgenode.RegisterPubSubServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %s", err)
	}
	return nil
}

/************* Begin RPCs **************/

// Publish only supported by Edge node brokers
func (s *EdgeNodeBroker) Publish(stream edgenode.PubSub_PublishServer) error {
	// GRPC - Client side streaming
	numImagesRecvd := 0
	for {
		im, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&edgenode.Status{
				Status: true,
			})
		}
		if err != nil {
			return fmt.Errorf("EdgeNodeBroker %s\n", err)
		}

		// Store image and timestamp
		ts, _ := time.Parse(time.RFC850, im.GetTimestamp())

		s.store[s.serverName].Append(im.GetImage(), ts)

		numImagesRecvd++
	}
}

// Subscribe supported by both Edge server and Edge node brokers. However their implementation is different
// Edge node subscription interacts with the controller to produce a stream that satisfies the requested latency accuracy
func (s *EdgeNodeBroker) Subscribe(imPars *edgenode.ImageStreamParameters, stream edgenode.PubSub_SubscribeServer) error {
	// GRPC - Server side streaming

	if s.serverName != imPars.Camid {
		return ErrWrongNode
	}
	if s.stopSubcription {
		s.stopSubcription = false
	}

	tstart, _ := time.Parse(time.RFC850, imPars.Start)
	tstop, _ := time.Parse(time.RFC850, imPars.Stop)

	/* Latency Controller hook
		lat := imPars.latency
		acc := imPars.accuracy

		Measure latency
		If latency is not desired latency, have controller read images from s.store[s.serverName]
		Create a new log, and write the altered images into s.store["modified"]
	    Set imSource to s.store["modified"]
	    This needs to be done in a loop,
	*/

	imts := make(chan storage.ImageTimestamp, 200*1024)
	errch := make(chan error)
	defer close(imts)
	defer close(errch)

	// Concurrent read of images from store got from producer
	go s.store[s.serverName].Read(imts, tstart, tstop, errch)

	errc := <-errch
	if errc != nil {
		log.Printf("EdgeNodeBroker %s\n", errc)
	}

	var lastTs storage.Timestamp

	ok := true
	for ok {
		select {
		case image := <-imts:
			{
				lastTs = image.Ts
				if err := stream.Send(&edgenode.Image{
					Image:     image.Im,
					Timestamp: (image.Ts).Format(time.RFC850),
				}); err != nil {
					return fmt.Errorf("EdgeNodeBroker %s\n", err)
				}
				ok = true
			}
		default:
			ok = false
		}
	}

	numIter := 0
	for lastTs.Before(tstop) { // More reading to be done; Poll for maxPollTime
		if s.stopSubcription { // From Unsubscribe API
			break
		}
		numIter++
		if numIter > maxPollTime {
			break
		}

		tstart = lastTs.Add(1 * time.Second) // TODO: Hacky - Read from 1 second later timestamp
		go s.store[s.serverName].Read(imts, tstart, tstop, errch)

		errc := <-errch
		if errc == storage.ErrTimestampMissing {
			time.Sleep(1 * time.Second) // TODO: Hacky - Sleep for a second
			continue
		}

		ok = true
		for ok {
			select {
			case image := <-imts:
				{
					lastTs = image.Ts
					if err := stream.Send(&edgenode.Image{
						Image:     image.Im,
						Timestamp: (image.Ts).Format(time.RFC850),
					}); err != nil {
						return fmt.Errorf("EdgeNodeBroker %s\n", err)
					}
					ok = true
				}
			default:
				ok = false
			}
		}
		time.Sleep(1 * time.Second) // TODO: Hacky - Sleep for a second
	}

	return nil
}

func (s *EdgeNodeBroker) Unsubscribe(ctx context.Context, caminfo *edgenode.CameraInfo) (*edgenode.Status, error) {

	s.stopSubcription = true

	return &edgenode.Status{
		Status: true,
	}, nil
}

/************* End RPCs **************/

/**************Helpers **************/

// authenticateAgent check the client credentials
func (s *EdgeNodeBroker) authenticateClient(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		clientLogin := strings.Join(md["login"], "")
		clientPassword := strings.Join(md["password"], "")
		if clientLogin != "client" {
			return "", fmt.Errorf("EdgeNodeBroker: unknown user %s\n", clientLogin)
		}
		if clientPassword != "edge" {
			return "", fmt.Errorf("EdgeNodeBroker: bad password %s\n", clientPassword)
		}
		return "42", nil
	}
	return "", fmt.Errorf("EdgeNodeBroker: missing credentials")
}

// unaryInterceptor calls authenticateClient with current context
func (s *EdgeNodeBroker) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	clientID, err := s.authenticateClient(ctx)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, clientIDKey, clientID)
	return handler(ctx, req)
}

func (s *EdgeNodeBroker) regsterWithEdgeServer(edgeServerIpaddr, login, password string) error {
	en := NewEdgeNodeClient(login, password) //username and password

	creds, err := credentials.NewClientTLSFromFile("cert/server.crt", "")
	if err != nil {
		return fmt.Errorf("EdgeNodeBroker: could not load tls cert: %s\n", err)
	}

	// Connect to edge server
	conn, err := grpc.Dial(edgeServerIpaddr, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&en.Auth))
	if err != nil {
		return fmt.Errorf("EdgeNodeBroker did not connect: %s\n", err)
	}
	defer conn.Close()
	cl := edgeserver.NewPubSubClient(conn)

	err = en.Register(s.ipaddr, s.serverName, cl)
	if err != nil {
		return fmt.Errorf("EdgeNodeBroker %s\n", err)
	}
	return nil

}
