package handlers

import (
	"fmt"
	"net/http"
	"nodes-grpc-fe/consts"

	"github.com/gin-gonic/gin"
)

func AddTokenHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// always check the cookies
		tokenCookies, err := c.Cookie(consts.COOKIE_NAME)
		if err != nil || tokenCookies == "" {
			Logout()
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		headerValue := fmt.Sprintf("Bearer %s", tokenCookies)
		c.Header("Authorization", headerValue)

		c.Set("token", tokenCookies)

		c.Next()
	}
}
