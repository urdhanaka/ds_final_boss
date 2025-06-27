package handlers

import (
	"net/http"
	"nodes-grpc-be/entities"
	"nodes-grpc-be/models"
	"nodes-grpc-be/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClusterHandler struct {
	redisQueueService *services.RedisJobQueue
	jwtService        *services.JwtService
	clusterService    *services.ClusterService
}

func NewClusterHandler(
	redisQueueService *services.RedisJobQueue,
	jwtService *services.JwtService,
	clusterService *services.ClusterService,
) *ClusterHandler {
	return &ClusterHandler{
		redisQueueService,
		jwtService,
		clusterService,
	}
}

func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	token, tokenExists := c.Get("token")
	if !tokenExists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(nil, "request is invalid"))
	}

	_, err := h.jwtService.GetUserIDByToken(token.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(nil, "request is invalid"))
	}

	addCluster := new(models.AddCluster)

	err = c.BindJSON(addCluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, "request is invalid"))
	}

	createClusterJob := &entities.Job{
		Type:    entities.JOB_PROVISIONING,
		Payload: addCluster,
	}

	err = h.redisQueueService.AddJob(createClusterJob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not complete the request"))
	}
}

func (h *ClusterHandler) GetUserCluster(c *gin.Context) {
	token, tokenExists := c.Get("token")
	if !tokenExists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(nil, "request is invalid"))
		return
	}

	userId, err := h.jwtService.GetUserIDByToken(token.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(nil, "request is invalid"))
		return
	}

	thisUserEntity := &entities.User{
		UserId: userId,
	}

	thisUserClusters, err := h.clusterService.GetUserCluster(c, thisUserEntity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not complete the request"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponseWithData(thisUserClusters, "success"))
}

func (h *ClusterHandler) GetClusterDetails(c *gin.Context) {
	token, tokenExists := c.Get("token")
	if !tokenExists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(nil, "request is invalid"))
		return
	}

	clusterId := c.Params.ByName("cluster_id")
	clusterIdUuid, err := uuid.Parse(clusterId)
	userId, err := h.jwtService.GetUserIDByToken(token.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(nil, "request is invalid"))
		return
	}

	clusterEntities := &entities.Cluster{
		ClusterID: clusterIdUuid,
	}

	thisCluster, err := h.clusterService.GetClusterById(c, clusterEntities)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(nil, "request is invalid"))
		return
	}
	if thisCluster.UserId != userId {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(nil, "unauthorized"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponseWithData(thisCluster, "success"))
}
