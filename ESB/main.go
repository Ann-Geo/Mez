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

	//es broker ip address
	esb := broker.NewEdgeServerBroker("EdgeServer", "127.0.0.1:20000", *broker.ActController, *broker.StorePath)
	fmt.Println(*broker.ActController)
	fmt.Println(*broker.StorePath)
	go esb.StartEdgeServerBroker()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
