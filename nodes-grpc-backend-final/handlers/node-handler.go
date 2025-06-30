package handlers

import (
	"net/http"
	"nodes-grpc-be/models"
	"nodes-grpc-be/services"

	"github.com/gin-gonic/gin"
)

type NodeHandler struct {
	nodeService *services.NodeService
	userService *services.UserService
}

func NewNodeHandler(
	nodeService *services.NodeService,
	userService *services.UserService,
) *NodeHandler {
	return &NodeHandler{
		nodeService,
		userService,
	}
}

func (h *NodeHandler) GetGroupCluster(c *gin.Context) {
	token := c.MustGet("token").(string)

	user, err := h.userService.Me(c, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not complete the request"))
		return
	}

	err = h.nodeService.UpdateNode(c, user.GroupId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not complete the request"))
		return
	}

	nodes, err := h.nodeService.GetGroupNodes(c, user.GroupId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not complete the request"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponseWithData(nodes, "success"))
}

func (h *NodeHandler) AddNode(c *gin.Context) {
	addNode := new(models.AddNode)

	err := c.BindJSON(addNode)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, "request body is invalid"))
		return
	}

	err = h.nodeService.AddNode(c, addNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not complete the request"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("success"))
}
