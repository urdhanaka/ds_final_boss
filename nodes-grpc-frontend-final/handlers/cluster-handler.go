package handlers

import (
	"encoding/json"
	"net/http"
	"nodes-grpc-fe/models"

	"github.com/gin-gonic/gin"
)

func CreateClusterHandler(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func AccessCluster(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterId := c.Params.ByName("cluster_id")

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

		res, err = apiClient.GetClusterDetails(c, clusterId)
		if err != nil || !res.Success {
			ErrorPageHandlers(err)(c)
		}
		thisCluster := &models.Clusters{}
		jsonBytes, err = json.Marshal(res.Data)
		if err != nil {
			ErrorPageHandlers(err)(c)
		}
		err = json.Unmarshal(jsonBytes, thisCluster)
		if err != nil {
			ErrorPageHandlers(err)(c)
		}

		c.HTML(http.StatusOK, "cluster-views.html", gin.H{
			"Name":           meUser.Name,
			"GroupName":      meUser.Group,
			"Vcpu":           meUser.Vcpu,
			"Ram":            meUser.Ram,
			"Storage":        meUser.Storage,
			"NodeSize":       meUser.NodeSize,
			"CurrentCluster": meUser.CurrentCluster,
			"MaxCluster":     meUser.MaxCluster,
			"Cluster":        *thisCluster,
		})
	}
}
