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

	//es broker ip address
	esb := broker.NewEdgeServerBroker("EdgeServer", "127.0.0.1:20000", *broker.ActController, *broker.StorePath, *broker.BrokerRestart)
	fmt.Println(*broker.ActController)
	fmt.Println(*broker.StorePath)
	go esb.StartEdgeServerBroker()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
