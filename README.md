Raven_v3 - A PubSub messaging system for Distributed Vision applications at the Edge

To compile gRPC proto (from Raven_v3)
protoc -I api/edgeserver api/edgeserver/edgeserver_api.proto --go_out=plugins=grpc:api/edgeserver
protoc -I api/edgenode api/edgenode/edgenode_api.proto --go_out=plugins=grpc:api/edgenode

Builds off Raven_v2

v3 To dos:
Cleanup errors
Review concurrency 
Review timeouts
Breakup subscribe into separate gothreads for read and send communicating by channels

v4 To dos:
Test each method thoroughly for all components
Test on 2 machines - edgeserver, and edgenode
Secure certs and passwords
Release to Github
Mez will use this codebase and add controller

v5 To dos:
Add persistence
Add REST api for telemetry