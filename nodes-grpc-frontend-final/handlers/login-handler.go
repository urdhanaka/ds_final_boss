package handlers

import (
	"errors"
	"net/http"
	"nodes-grpc-fe/consts"

	"github.com/gin-gonic/gin"
)

func LoginPageHandlers() gin.HandlerFunc {
	return func(c *gin.Context) {
		// check if cookie is set
		_, err := c.Cookie(consts.COOKIE_NAME)
		if err != nil {
			c.HTML(http.StatusOK, "login-views.html", nil)
			return
		}

		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

func LoginHandlers(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		res, err := apiClient.Login(c, email, password)
		if err != nil {
            // server unavailable handler
			if errors.Is(err, ErrServerIsNotResponding) {
				c.HTML(http.StatusServiceUnavailable, "login-views.html", gin.H{
					"Error":    err.Error(),
					"Email":    email,
					"Password": password,
				})
				return
			}

			c.HTML(http.StatusBadRequest, "login-views.html", gin.H{
				"Error":    err.Error(),
				"Email":    email,
				"Password": password,
			})
			return
		}

		if !res.Success {
			c.HTML(http.StatusBadRequest, "login-views.html", gin.H{
				"Error":    res.Error,
				"Email":    email,
				"Password": password,
			})
			return
		}

		c.SetCookie(
			consts.COOKIE_NAME, // cookie name
			res.Data.(string),  // value of the cookie
			3600,               // max age
			"/",                // path
			"localhost",        // domain
			false,              // secure
			true,               // httponly
		)
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}
