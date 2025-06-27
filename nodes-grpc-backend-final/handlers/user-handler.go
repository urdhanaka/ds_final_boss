package handlers

import (
	"net/http"
	"nodes-grpc-be/models"
	"nodes-grpc-be/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(
	userService *services.UserService,
) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	loginUser := new(models.LoginUser)

	err := c.BindJSON(loginUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, "request is invalid"))
        return
	}

	token, err := h.userService.Login(c, loginUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not complete the request"))
        return
	}

	c.JSON(http.StatusOK, NewSuccessResponseWithData(token, "success"))
}

func (h *UserHandler) MeUser(c *gin.Context) {
	token := c.MustGet("token").(string)

	user, err := h.userService.Me(c, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not complete the request"))
        return
	}

	c.JSON(http.StatusOK, NewSuccessResponseWithData(user, "success"))
}
