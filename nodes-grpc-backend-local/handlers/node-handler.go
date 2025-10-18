package handlers

import (
	"net/http"
	"nodes-grpc-backend-local/model"
	"nodes-grpc-backend-local/services"

	"github.com/gin-gonic/gin"
)

type NodeHandler struct {
	nodeService *services.NodeService
}

func NewNodeHandler(nodeService *services.NodeService) *NodeHandler {
	return &NodeHandler{
		nodeService,
	}
}

func (h *NodeHandler) GetAllNodes(c *gin.Context) {
	data := h.nodeService.GetAllNodes(c)
	if data.IsError == true {
		c.JSON(http.StatusBadRequest, NewErrorResponse(data.Error, data.Message))
	}

	c.JSON(http.StatusOK, NewSuccessResponseWithData(data.Data, data.Message))
}

func (h *NodeHandler) AddNode(c *gin.Context) {
	addNode := new(model.AddNode)

	err := c.BindJSON(addNode)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, "request is invalid"))
	}

	err = h.nodeService.AddNode(c, addNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not add node"))
	}

	c.JSON(http.StatusOK, NewSuccessResponse("add node success"))
}

func (h *NodeHandler) GetAllNodeStatus(c *gin.Context) {
}
