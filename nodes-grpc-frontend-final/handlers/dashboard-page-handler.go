package handlers

import (
	"encoding/json"
	"net/http"
	"nodes-grpc-fe/models"

	"github.com/gin-gonic/gin"
)

func DashboardPageHandlers(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := apiClient.Me(c)
		if err != nil || !res.Success {
			ErrorPageHandlers(err)
		}

		returnedUser := &models.MeUserReturn{}

		jsonBytes, err := json.Marshal(res.Data)
		if err != nil {
			ErrorPageHandlers(err)
		}

		err = json.Unmarshal(jsonBytes, returnedUser)
		if err != nil {
			ErrorPageHandlers(err)
		}

		c.HTML(http.StatusOK, "dashboard-views.html", gin.H{
			"Name":           returnedUser.Name,
			"GroupName":      returnedUser.Group,
			"Vcpu":           returnedUser.Vcpu,
			"Ram":            returnedUser.Ram,
			"Storage":        returnedUser.Storage,
			"NodeSize":       returnedUser.NodeSize,
			"CurrentCluster": returnedUser.CurrentCluster,
			"MaxCluster":     returnedUser.MaxCluster,
		})
	}
}
