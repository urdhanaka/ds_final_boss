package routers

import (
	"nodes-grpc-frontend-local/src/handlers"
	"nodes-grpc-frontend-local/src/repository"
	"nodes-grpc-frontend-local/src/services"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetRouters(fiberApp *fiber.App, dbPool *pgxpool.Pool) {
	nodeRepository := repository.NewNodeRepository(dbPool)
	nodeService := services.NewNodeService(nodeRepository)

	nodeHandler := handlers.NewNodeHandler(nodeService)

	// set static assets
	fiberApp.Get("assets/styles.css", static.New("src/assets/css/styles.css"))
	fiberApp.Get("assets/script.js", static.New("src/assets/js/script.js"))
	// fiberApp.Get("assets/oval.svg", static.New("src/assets/svg/oval.svg"))

	// homepage
	fiberApp.Get("/", nodeHandler.Homepage)

	// POST, register a node
	fiberApp.Post("/register_node", nodeHandler.RegisterNode)

	// POST, create the cluster
	fiberApp.Post("/create_cluster", nodeHandler.CreateCluster)

	// GET, access the cluster
	fiberApp.Get("/dashboard/:id", nodeHandler.AccessCluster)
}
