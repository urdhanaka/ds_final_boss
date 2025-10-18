package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Final(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := apiClient.Me(c)
		if err != nil || !res.Success {
			if err == nil {
				ErrorPageHandlers(err, res)(c)
				return
			}

			ErrorPageHandlers(err, res)(c)
			return
		}

        c.HTML(http.StatusOK, "final-final-final.html", gin.H{})
	}
}
