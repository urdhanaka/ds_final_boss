package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nodes-grpc-fe/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		if res.Data != nil {
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
		})
	}
}

func DeleteCluster(apiClient *ApiClient) gin.HandlerFunc {
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

		UuidParseId, _ := uuid.Parse(clusterId)
		deleteClusterModel := &models.DeleteCluster{
			ClusterId: UuidParseId,
			UserId:    meUser.UserId,
		}

		res, err = apiClient.DeleteCluster(c, deleteClusterModel)
		ErrorPageHandlers(err, res)
	}
}

func AccessClusterStatus(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterId := c.Params.ByName("cluster_id")

		// check if uuid is valid
		err := uuid.Validate(clusterId)
		if err != nil {
			c.HTML(http.StatusOK, "cluster-status-invalid-id.html", nil)
			return
		}

		c.HTML(http.StatusOK, "cluster-status.html", gin.H{
			"ClusterId": clusterId,
		})
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

func DownloadKubeconfigHandler(apiClient *ApiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterId := c.Params.ByName("cluster_id")

		err := apiClient.DownloadKubeconfig(c, clusterId)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
		}
	}
}
