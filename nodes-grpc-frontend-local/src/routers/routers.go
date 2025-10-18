package routers

import (
	"nodes-grpc-frontend-local/src/handlers"
	"nodes-grpc-frontend-local/src/services"

	"github.com/gofiber/fiber/v2"
)

func SetRouters(fiberApp *fiber.App) {
	nodeService := services.NewNodeService()
	nodeHandler := handlers.NewNodeHandler(nodeService)

	// set static assets
	fiberApp.Get("assets/styles.css", func(c *fiber.Ctx) error {
		return c.SendFile("src/assets/css/styles.css")
	})
	fiberApp.Get("assets/script.js", func(c *fiber.Ctx) error {
		return c.SendFile("src/assets/js/script.js")
	})

	// homepage
	fiberApp.Get("", nodeHandler.LoginPage)

	// // POST, register a node
	// fiberApp.Post("/register_node", nodeHandler.RegisterNode)

	// // POST, create the cluster
	// fiberApp.Post("/create_cluster", nodeHandler.CreateCluster)

	// websocket
	// fiberApp.Get("/ws/receive_logs/:cluster_name",
	// 	websocket.New(websocketHandler.ReceiveLogs))
	//
	// fiberApp.Get("/ws/stream_logs/:cluster_name",
	// 	websocket.New(websocketHandler.StreamLogs))
}
