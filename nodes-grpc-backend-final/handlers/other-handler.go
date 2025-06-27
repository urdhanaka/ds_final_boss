package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type OtherHandler struct{}

func NewOtherHandler() *OtherHandler {
	return &OtherHandler{}
}

func (h *OtherHandler) Checkhealth(c *gin.Context) {
	c.JSON(http.StatusOK, NewSuccessResponse("server is up"))
}
