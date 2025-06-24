package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorPageHandlers(err error) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login-views.html", gin.H{
			"Error": err,
		})
	}
}
