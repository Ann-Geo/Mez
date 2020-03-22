package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Ann-Geo/Mez/broker"
)

func main() {

	flag.Parse()
	//en broker address
	enb1 := broker.NewEdgeNodeBroker("cam1", "127.0.0.1:10000", *broker.ActController, *broker.StorePath, *broker.BrokerRestart)
	fmt.Println(*broker.ActController)
	//es broker address
	go enb1.StartEdgeNodeBroker("127.0.0.1:20000", "client", "edge")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
