package handlers

import (
	"net/http"
	"nodes-grpc-fe/models"

	"github.com/gin-gonic/gin"
)

func ErrorPageHandlers(err error, apiResponse *models.ApiResponse) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiResponse == nil {
			c.HTML(http.StatusOK, "error-views.html", gin.H{
				"Error": err.Error(),
			})
		} else if err == nil {
			c.HTML(http.StatusOK, "error-views.html", gin.H{
				"Error": apiResponse.Error,
			})
		}
	}
}
