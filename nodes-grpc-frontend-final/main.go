package main

import (
	"log"
	"nodes-grpc-fe/config"
)

func main() {
	ginInstance := config.NewGin()
	restyClient := config.NewResty()

	setRouters(ginInstance, restyClient)

	log.Fatal(ginInstance.Run())
}
