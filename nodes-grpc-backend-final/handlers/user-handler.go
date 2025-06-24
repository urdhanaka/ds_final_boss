package handlers

import (
	"net/http"
	"nodes-grpc-be/models"
	"nodes-grpc-be/services"

	"github.com/gin-gonic/gin"
)

type UserHandlers struct {
	userService *services.UserService
}

func NewUserHandler(
	userService *services.UserService,
) *UserHandlers {
	return &UserHandlers{
		userService: userService,
	}
}

func (h *UserHandlers) Login(c *gin.Context) {
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

func (h *UserHandlers) MeUser(c *gin.Context) {
	token := c.MustGet("token").(string)

	user, err := h.userService.Me(c, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not complete the request"))
        return
	}

	c.JSON(http.StatusOK, NewSuccessResponseWithData(user, "success"))
}
