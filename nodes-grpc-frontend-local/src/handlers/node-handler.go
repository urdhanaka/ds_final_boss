package handlers

import (
	"nodes-grpc-frontend-local/src/components/base"
	"nodes-grpc-frontend-local/src/model"
	"nodes-grpc-frontend-local/src/model/virtualization_model"
	"nodes-grpc-frontend-local/src/services"

	"github.com/gofiber/fiber/v3"
)

type NodeHandler struct {
	NodeService *services.NodeService
}

func NewNodeHandler(nodeService *services.NodeService) *NodeHandler {
	return &NodeHandler{
		NodeService: nodeService,
	}
}

func (h *NodeHandler) Homepage(c fiber.Ctx) error {
	c.Set("Content-Type", "text/html")

	res, err := h.NodeService.GetAllNodes(c.Context())
	if err != nil {
		return err
	}

	return base.Page(res).Render(c.Context(), c.Response().BodyWriter())
}

func (h *NodeHandler) RegisterNode(c fiber.Ctx) error {
	c.Accepts("application/json")

	registerNodeRequest := new(model.Node)

	err := c.Bind().Body(registerNodeRequest)
	if err != nil {
		return c.JSON(NewErrorResponse(err))
	}

	err = h.NodeService.RegisterNode(c.Context(), registerNodeRequest)
	if err != nil {
		return c.JSON(NewErrorResponse(err))
	}

	return c.JSON(NewSuccessResponse())
}

func (h *NodeHandler) CreateCluster(c fiber.Ctx) error {
	c.Accepts("application/json")

	clusterRequest := new(virtualization_model.CreateClusterRequest)

	err := c.Bind().Body(clusterRequest)
	if err != nil {
		return c.JSON(NewErrorResponse(err))
	}

	res, err := h.NodeService.CreateCluster(c.Context(), clusterRequest)
	if err != nil {
		return c.JSON(NewErrorResponse(err))
	}

	return c.JSON(NewSuccessResponseWithData(res))
}

func (h *NodeHandler) AccessCluster(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.SendString("id invalid")
	}

	return nil
}
