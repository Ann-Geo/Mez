## Mez - A PubSub messaging system for Distributed Vision applications at the Edge
(Work in progress)

Overview

Mez is a publish subscribe messaging system intended to serve computer vision applications distributed at the Edge. Typical deployment platform for Mez consists of multiple IoT camera nodes (embedded boards) and a single GPU equipped Edge server. Video frames are published and recorded at the IoT camera nodes in real-time. Computer vision applications subscribe to one or more IoT camera nodes to receive video frames published at the nodes. The IoT camera nodes communicate to Edge server through WiFi link. The role of Mez is to make available the video frames send from IoT camera nodes to the Edge server within the application requested latency and accuracy bounds in presence of WiFi channel interference.


Prerequisites
1. Golang, protocol buffers ans grpc 
2. PTP synchronization protocol
3. OpenCv library for Golang - gocv



Directory structure
- broker: Message brokers handling the APIs log storage in Mez
- storage: In-memory storage layer of Mez
- api: APIs used by applications to communicate with Mez
- client: Client library that provides helper functions for applications to communicate with Mez
- latcontroller: Latency controller incorporated with Mez
- test_pgms: Example publisher and subscriber applications


Quick start
- Clone the Mez repository
- Run Edge server broker (ESB), go run ESB/main.go
- Run Edge node broker (ENB), go run ENB/main.go
- Run Subscriber, go run test_pgms/subscribe_c/main.go
- Run publisher, go run test_pgms/publish/main.go
- In case controller flag is enabled (details given below), run the controller, go run Controller/main.go

Flags
- -c => to enable control action (makes sense only if ESB and ENB are deployed on two separate physical nodes and communicate through WiFi)
    - 1 to enable controller, default value is 0
- -p => to enable persistence for the brokers
    - specify path to a directory, by default persistence is disabled
- -r => to restart the brokers, during restart brokers recover in memory logs if persistence was enabled
    - 1 to enable restart, default value is 0






