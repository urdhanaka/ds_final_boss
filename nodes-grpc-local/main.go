package main

import (
	"nodes-grpc-local/services"
)

func main() {
	serviceStruct := services.NewConnection()

	services.StartGrpcServer(serviceStruct)
}
