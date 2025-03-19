package routers

import (
	"nodes-grpc-frontend/components/page"
	"nodes-grpc-frontend/handlers"
	usecases "nodes-grpc-frontend/use-cases"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func SetRouter(fiberApp *fiber.App) {
	nodeUsecase := usecases.NewNodeUsecaseImpl()
	nodeHandler := handlers.NewNodeHandler(nodeUsecase)

	// static assets
	fiberApp.Get("assets/styles.css", static.New("assets/css/styles.css"))
	fiberApp.Get("assets/htmx.js", static.New("assets/js/htmx.js"))
	fiberApp.Get("assets/htmx-json-enc.js", static.New("assets/js/htmx-json-enc.js"))
	fiberApp.Get("assets/oval.svg", static.New("assets/svg/oval.svg"))

	// welcome page
	fiberApp.Get("/", func(c fiber.Ctx) error {
		nodes := nodeHandler.GetLocalStorage(c)

		return Render(c, page.Dashboard(nodes))
	})

	// nodes related
	fiberApp.Get("/nodes", nodeHandler.GetNodes)
	fiberApp.Post("/nodes", nodeHandler.Register)

	// get nodes status
	fiberApp.Get("/nodes-status", nodeHandler.GetNodesStatus)

	// create the cluster
	fiberApp.Post("/create-cluster", nodeHandler.CreateCluster)

	// temp
	fiberApp.Post("/tmp", nodeHandler.TestHtmx)
}
