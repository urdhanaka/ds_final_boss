package main

import (
	"log"
	"nodes-grpc-frontend/common/config"
	proto_model "nodes-grpc-frontend/common/model/proto-model"
	"nodes-grpc-frontend/routers"
)

var nodeList *proto_model.NodeList

func main() {
	app := config.NewFiber()

	routers.SetRouter(app)

	log.Fatal(app.Listen(":3000"))
}
