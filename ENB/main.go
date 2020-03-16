package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Ann-Geo/Mez/broker"
)

func main() {

	/*var actController string = "0"
	arguments := os.Args
	for i := 0; i < len(arguments); i++ {
		if strings.Compare(arguments[i], "-c") == 0 {
			if strings.Compare(arguments[i+1], "1") == 0 {
				actController = "1"
				break
			}
		}
	}*/

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
