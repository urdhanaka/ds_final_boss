package handlers

import (
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
		email := c.Request.FormValue("email")
		password := c.Request.FormValue("password")

		res, err := apiClient.Login(c, email, password)
		if err != nil {
			c.HTML(http.StatusOK, "login-views.html", nil)
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
			consts.COOKIE_NAME,
			res.Data.(string),
			3600,
			"/",
			"localhost",
			false,
			true,
		)
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}
