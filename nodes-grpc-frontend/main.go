package main

import (
	"log"
	"nodes-grpc-frontend/common/config"
	proto_model "nodes-grpc-frontend/common/model/proto-model"
	"nodes-grpc-frontend/routers"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

var nodeList *proto_model.NodeList

type NodeIdentity struct {
	Hostname  string `json:"hostname"`
	IpAddress string `json:"ip_address"`
	GrpcPort  string `json:"grpc_port"`
}

type MasterToken struct {
	Token string `json:"token"`
}

func init() {
	nodeList = new(proto_model.NodeList)
	nodeList.Nodes = make([]*proto_model.NodeStatus, 0)
}

func Render(c fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")

	return component.Render(c.Context(), c.Response().BodyWriter())
}

func main() {
	app := config.NewFiber()

	routers.SetRouter(app)

	log.Fatal(app.Listen(":3000"))
}
