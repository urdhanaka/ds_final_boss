package handlers

import (
	"nodes-grpc-backend-local/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService,
	}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userService.GetAllUser(c)
	if err != nil {
		c.JSON(400, NewErrorResponse(err))
	}

    c.JSON(200, NewSuccessResponseWithData(users))
}
