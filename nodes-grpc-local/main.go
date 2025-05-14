package main

import (
	"nodes-grpc-local/services"
)

func main() {
	serviceStruct := services.NewInitStruct()

	services.StartGrpcServer(serviceStruct)
}
