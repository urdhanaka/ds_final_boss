package handlers

import (
	// "nodes-grpc-frontend-local/src/components/base"
	"nodes-grpc-frontend-local/src/components/login"
	"nodes-grpc-frontend-local/src/services"

	"github.com/gofiber/fiber/v2"
)

type NodeHandler struct {
	NodeService *services.NodeService
}

func NewNodeHandler(nodeService *services.NodeService) *NodeHandler {
	return &NodeHandler{
		NodeService: nodeService,
	}
}

func (h *NodeHandler) LoginPage(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html")

    return login.LoginPage().Render(c.Context(), c.Response().BodyWriter())
}

// func (h *NodeHandler) HomePage(c *fiber.Ctx) error {
// 	c.Set("Content-Type", "text/html")
//
// 	// res, err := h.NodeService.GetAllNodes(c.Context())
// 	// if err != nil {
// 	// 	return err
// 	// }
//
// 	return base.Page(res).Render(c.Context(), c.Response().BodyWriter())
// }
