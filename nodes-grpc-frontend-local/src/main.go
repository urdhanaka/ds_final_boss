package src

import (
	"log"
	"nodes-grpc-frontend-local/src/config"
	"nodes-grpc-frontend-local/src/routers"
)

func StartApp() {
	app := config.NewFiber()
	// db := config.NewPsql()
	// defer db.Close(context.Background())

	// routers.SetRouters(app, db)
	routers.SetRouters(app)

	log.Fatal(app.Listen(":3000"))
}
