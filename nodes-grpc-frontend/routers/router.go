package routers

import (
	"fmt"
	"nodes-grpc-frontend/components/page"
	"nodes-grpc-frontend/handlers"
	"nodes-grpc-frontend/services/db_service"
	"nodes-grpc-frontend/services/node_service"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/proxy"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/jmoiron/sqlx"
)

func SetRouter(fiberApp *fiber.App, dbConnection *sqlx.DB) {
	dbUsecase := db_service.NewDbUsecaseImpl(dbConnection)

	nodeUsecase := node_service.NewNodeUsecaseImpl(dbUsecase)
	nodeHandler := handlers.NewNodeHandler(nodeUsecase)

	// static assets
	fiberApp.Get("assets/styles.css", static.New("assets/css/styles.css"))
	fiberApp.Get("assets/htmx.js", static.New("assets/js/htmx.js"))
	fiberApp.Get("assets/script.js", static.New("assets/js/script.js"))
	fiberApp.Get("assets/htmx-json-enc.js", static.New("assets/js/htmx-json-enc.js"))
	fiberApp.Get("assets/oval.svg", static.New("assets/svg/oval.svg"))

	// welcome page
	fiberApp.Get("/", func(c fiber.Ctx) error {
		nodes := nodeHandler.GetNodes(c)

		return Render(c, page.Dashboard(nodes))
	})

	// nodes related
	// fiberApp.Get("/nodes", nodeHandler.GetNodes)
	fiberApp.Post("/nodes", nodeHandler.Register)

	// get nodes status
	// fiberApp.Get("/nodes-status", nodeHandler.GetNodesStatus)

	// create the cluster
	fiberApp.Post("/create_cluster", nodeHandler.CreateCluster)

	// kubernetes dashboard
	fiberApp.Get("/dashboard/:token", func(c fiber.Ctx) error {
		masterIP, err := nodeHandler.GetDashboard(c)
		if err != nil {
			return err
		}

		return proxy.DomainForward(fmt.Sprintf("https://%s:8443", masterIP), "http://localhost:3000", nil)(c)
	})
}
