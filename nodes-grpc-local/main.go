package main

import (
	"nodes-grpc-local/services"
)

func main() {
	serviceStruct := services.NewInitStruct()

	// main grpc service
	services.StartGrpcServer(serviceStruct)
}
