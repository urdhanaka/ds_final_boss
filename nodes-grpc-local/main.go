package main

import (
	"nodes-grpc-local/services"
)

func main() {
	serviceStruct := services.NewInitStruct()

	// background job
	go services.StartWorker(serviceStruct)
	go services.StartWebsocket(serviceStruct)

	// main grpc service
	services.StartGrpcServer(serviceStruct)
}
