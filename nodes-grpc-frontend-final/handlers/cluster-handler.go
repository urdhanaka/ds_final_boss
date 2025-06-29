package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nodes-grpc-fe/models"

	"github.com/gin-gonic/gin"
)

func CreateClusterPageHandler(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := apiClient.Me(c)
		ErrorPageHandlers(err, res)

		meUser := &models.MeUserReturn{}
		jsonBytes, err := json.Marshal(res.Data)
		if err != nil {
			ErrorPageHandlers(err, nil)
			return
		}
		err = json.Unmarshal(jsonBytes, meUser)
		if err != nil {
			ErrorPageHandlers(err, nil)
			return
		}

		res, err = apiClient.GetGroupNodes(c)
		if err != nil || !res.Success {
			if err == nil {
				ErrorPageHandlers(err, res)(c)
				return
			}

			ErrorPageHandlers(err, res)(c)
			return
		}
		groupNodes := &[]*models.Nodes{}
		jsonBytes, err = json.Marshal(res.Data)
		if err != nil {
			ErrorPageHandlers(err, nil)(c)
		}
		err = json.Unmarshal(jsonBytes, groupNodes)
		if err != nil {
			ErrorPageHandlers(err, nil)(c)
		}

		c.HTML(http.StatusOK, "create-cluster-views.html", gin.H{
			"Name":           meUser.Name,
			"GroupName":      meUser.Group,
			"Vcpu":           meUser.Vcpu,
			"Memory":         meUser.Memory,
			"Storage":        meUser.Storage,
			"NodeSize":       meUser.NodeSize,
			"CurrentCluster": meUser.CurrentCluster,
			"MaxCluster":     meUser.MaxCluster,
			"GroupNodes":     *groupNodes,
		})
	}
}

func AccessCluster(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterId := c.Params.ByName("cluster_id")

		res, err := apiClient.Me(c)
		ErrorPageHandlers(err, res)

		meUser := &models.MeUserReturn{}
		jsonBytes, err := json.Marshal(res.Data)
		if err != nil {
			ErrorPageHandlers(err, nil)
			return
		}
		err = json.Unmarshal(jsonBytes, meUser)
		if err != nil {
			ErrorPageHandlers(err, nil)
			return
		}

		res, err = apiClient.GetClusterDetails(c, clusterId)
		ErrorPageHandlers(err, res)

		thisCluster := &models.Clusters{}
		jsonBytes, err = json.Marshal(res.Data)
		if err != nil {
			ErrorPageHandlers(err, nil)
			return
		}
		err = json.Unmarshal(jsonBytes, thisCluster)
		if err != nil {
			ErrorPageHandlers(err, nil)
			return
		}

		c.HTML(http.StatusOK, "cluster-views.html", gin.H{
			"Name":           meUser.Name,
			"GroupName":      meUser.Group,
			"Vcpu":           meUser.Vcpu,
			"Memory":         meUser.Memory,
			"Storage":        meUser.Storage,
			"NodeSize":       meUser.NodeSize,
			"CurrentCluster": meUser.CurrentCluster,
			"MaxCluster":     meUser.MaxCluster,
			"Cluster":        *thisCluster,
		})
	}
}

func AccessClusterStatus(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// clusterId := c.Params.ByName("cluster_id")
		// res, err := apiClient.Me(c)
		// ErrorPageHandlers(err, res)
		c.HTML(http.StatusOK, "cluster-status.html", gin.H{})
	}
}

func CreateClusterHandler(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		createCluster := new(models.CreateCluster)

		err := c.BindJSON(createCluster)
		if err != nil {
			fmt.Println(err)
		}

		res, err := apiClient.CreateCluster(c, createCluster)
		ErrorPageHandlers(err, res)

		c.JSON(http.StatusOK, res)
	}
}
