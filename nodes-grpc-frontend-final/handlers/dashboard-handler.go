package handlers

import (
	"encoding/json"
	"net/http"
	"nodes-grpc-fe/models"

	"github.com/gin-gonic/gin"
)

func DashboardPageHandlers(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// me user data
		res, err := apiClient.Me(c)
		if err != nil || !res.Success {
			ErrorPageHandlers(err)(c)
		}
		meUser := &models.MeUserReturn{}
		jsonBytes, err := json.Marshal(res.Data)
		if err != nil {
			ErrorPageHandlers(err)(c)
		}
		err = json.Unmarshal(jsonBytes, meUser)
		if err != nil {
			ErrorPageHandlers(err)(c)
		}

		// group nodes data
		res, err = apiClient.GetGroupNodes(c)
		groupNodes := &[]*models.Nodes{}
		jsonBytes, err = json.Marshal(res.Data)
		if err != nil {
			ErrorPageHandlers(err)(c)
		}
		err = json.Unmarshal(jsonBytes, groupNodes)
		if err != nil {
			ErrorPageHandlers(err)(c)
		}

		// user clusters
		res, err = apiClient.GetUserClusters(c)
		userClusters := &[]*models.Clusters{}
		jsonBytes, err = json.Marshal(res.Data)
		if err != nil {
			ErrorPageHandlers(err)(c)
		}
		err = json.Unmarshal(jsonBytes, userClusters)
		if err != nil {
			ErrorPageHandlers(err)(c)
		}

		c.HTML(http.StatusOK, "dashboard-views.html", gin.H{
			"Name":           meUser.Name,
			"GroupName":      meUser.Group,
			"Vcpu":           meUser.Vcpu,
			"Ram":            meUser.Ram,
			"Storage":        meUser.Storage,
			"NodeSize":       meUser.NodeSize,
			"CurrentCluster": meUser.CurrentCluster,
			"MaxCluster":     meUser.MaxCluster,
			"GroupNodes":     *groupNodes,
			"UserClusters":   *userClusters,
		})
	}
}
