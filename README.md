Mez - A PubSub messaging system for Distributed Vision applications at the Edge
(Work in progress)

Overview

Mez is a publish subscribe messaging system intended to serve computer vision applications distributed at the Edge. Typical deployment platform for Mez consists of multiple IoT camera nodes (embedded boards) and a single GPU equipped Edge server. Video frames are published and recorded at the IoT camera nodes in real-time. Computer vision applications subscribe to one or more IoT camera nodes to receive video frames published at the nodes. The IoT camera nodes communicate to Edge server through WiFi link. The role of Mez is to make available the video frames send from IoT camera nodes to the Edge server within the application requested latency and accuracy bounds in presence of WiFi channel interference.


Prerequisites
1. Golang, protocol buffers ans grpc 
2. PTP synchronization protocol
3. OpenCv library for Golang - gocv



Directory structure
broker: Message brokers handling the APIs log storage in Mez
storage: In-memory storage layer of Mez
api: APIs used by applications to communicate with Mez
client: Client library that provides helper functions for applications to communicate with Mez
latcontroller: Latency controller incorporated with Mez
test_pgms: Example publisher and subscriber applications





