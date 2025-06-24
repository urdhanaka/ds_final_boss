package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RedirectToLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/login")
	}
}
