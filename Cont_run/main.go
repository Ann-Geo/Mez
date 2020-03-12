package main

import (
	"os"
	"os/signal"

	"github.com/Ann-Geo/Mez/latcontroller"
)

func main() {

	cont := latcontroller.NewController("127.0.0.1:9002", "/home/research/goworkspace/src/vsc_workspace/")
	go cont.StartLatencyController()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
