package handlers

import (
	"net/http"
	"nodes-grpc-backend-local/services"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupService *services.GroupService
}

func NewGroupHandler(groupService *services.GroupService) *GroupHandler {
	return &GroupHandler{
		groupService,
	}
}

func (h *GroupHandler) GetAllGroups(c *gin.Context) {
	data := h.groupService.GetAllGroups(c)
	if data.IsError == true {
		c.JSON(http.StatusBadRequest, NewErrorResponse(data.Error, data.Message))
	}

	c.JSON(http.StatusOK, NewSuccessResponseWithData(data.Data, data.Message))
}
