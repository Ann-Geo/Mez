package main

import (
	"os"
	"os/signal"

	"github.com/arun-ravindran/Raven/broker"
)

func main() {

	esb := broker.NewEdgeServerBroker("EdgeServer", "127.0.0.1:20000")
	go esb.StartEdgeServerBroker()

	enb1 := broker.NewEdgeNodeBroker("cam1", "127.0.0.1:10000")
	go enb1.StartEdgeNodeBroker("127.0.0.1:20000", "client", "edge")

	enb2 := broker.NewEdgeNodeBroker("cam2", "127.0.0.1:30000")
	go enb2.StartEdgeNodeBroker("127.0.0.1:20000", "client", "edge")

	enb3 := broker.NewEdgeNodeBroker("cam3", "127.0.0.1:40000")
	go enb3.StartEdgeNodeBroker("127.0.0.1:20000", "client", "edge")

	enb4 := broker.NewEdgeNodeBroker("cam4", "127.0.0.1:50000")
	go enb4.StartEdgeNodeBroker("127.0.0.1:20000", "client", "edge")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
