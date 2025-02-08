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

	// welcome page
	fiberApp.Get("/", func(c fiber.Ctx) error {
		nodes := nodeHandler.GetLocalStorage(c)

		return Render(c, page.Dashboard(nodes))
	})

    // nodes related
	fiberApp.Get("/nodes", nodeHandler.GetNodes)
	fiberApp.Post("/nodes", nodeHandler.Register)

    // create the cluster
	fiberApp.Post("/create-cluster", nodeHandler.CreateCluster)
}
