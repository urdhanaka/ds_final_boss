package handlers

import (
	"net/http"
	"nodes-grpc-backend-local/model"
	"nodes-grpc-backend-local/services"

	"github.com/gin-gonic/gin"
)

type ClusterHandler struct {
	clusterService *services.ClusterService
	queueService   *services.QueueService
}

func NewClusterHandler(
	clusterService *services.ClusterService,
	queueService *services.QueueService,
) *ClusterHandler {
	return &ClusterHandler{
		clusterService,
		queueService,
	}
}

func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	addCluster := new(model.AddCluster)

	err := c.BindJSON(addCluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, "request is invalid"))
	}

	err = h.clusterService.AddCluster(c, addCluster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not add cluster"))
	}

	c.JSON(http.StatusOK, NewSuccessResponse("add cluster success"))
}
