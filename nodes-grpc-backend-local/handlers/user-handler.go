package handlers

import (
	"net/http"
	"nodes-grpc-backend-local/model"
	"nodes-grpc-backend-local/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
	jwtService  *services.JwtService
}

func NewUserHandler(
	userService *services.UserService,
	jwtService *services.JwtService,
) *UserHandler {
	return &UserHandler{
		userService,
		jwtService,
	}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	data := h.userService.GetAllUser(c)
	if data.IsError == true {
		c.JSON(http.StatusBadRequest, NewErrorResponse(data.Error, data.Message))
	}

	c.JSON(http.StatusOK, NewSuccessResponseWithData(data.Data, data.Message))
}

func (h *UserHandler) AddUser(c *gin.Context) {
	addUser := new(model.AddUser)

	err := c.BindJSON(addUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, "request body is invalid"))
	}

	err = h.userService.AddUser(c, addUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not add user"))
	}

	c.JSON(http.StatusOK, NewSuccessResponse("add user success"))
}

func (h *UserHandler) Login(c *gin.Context) {
	loginUser := new(model.LoginUser)

	err := c.BindJSON(loginUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, "request body is invalid"))
	}

	// getUser, err := h.userService.Login(c, loginUser)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, NewErrorResponse(err, "could not log in"))
	// }

	// token := h.jwtService.GenerateToken(getUser.UserId, getUser.Role)
	// response := model.Auth{
	// 	Token: token,
	// 	Role:  getUser.Role,
	// }

	c.JSON(http.StatusOK, NewSuccessResponseWithData("", "login success"))
}
