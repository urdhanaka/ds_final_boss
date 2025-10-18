package main

import (
	"log"
	"nodes-grpc-frontend/common/config"
	"nodes-grpc-frontend/routers"
)

func main() {
	db := config.NewDB()
	app := config.NewFiber()

	routers.SetRouter(app, db)

	log.Fatal(app.Listen(":3000"))
}
