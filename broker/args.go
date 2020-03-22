package broker

import "flag"

//to enable latency controller
var ActController = flag.String("c", "0", "a string")

//to enable persistence
var StorePath = flag.String("p", "../../def_store/", "a string")

//to enable restart of brokers
var BrokerRestart = flag.String("r", "0", "a string")
