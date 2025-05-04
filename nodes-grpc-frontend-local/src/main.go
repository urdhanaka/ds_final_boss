package src

import (
	"log"
	"nodes-grpc-frontend-local/src/config"
	"nodes-grpc-frontend-local/src/routers"
)

func StartApp() {
	app := config.NewFiber()

	routers.SetRouters(app)

	log.Fatal(app.Listen(":3000"))
}
