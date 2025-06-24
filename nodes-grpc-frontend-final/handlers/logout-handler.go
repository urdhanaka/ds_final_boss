package handlers

import (
	"net/http"
	"nodes-grpc-fe/consts"

	"github.com/gin-gonic/gin"
)

func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.SetCookie(
			consts.COOKIE_NAME,
			"",
			-1,
			"/",
			"localhost",
            true,
			true,
		)
        c.Header("Authorization", "")

		c.Redirect(http.StatusSeeOther, "/login")
	}
}
