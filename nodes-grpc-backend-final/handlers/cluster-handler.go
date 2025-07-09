package handlers

import (
	"database/sql"
	"errors"
	"fmt"
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
	userService       *services.UserService
}

func NewClusterHandler(
	redisQueueService *services.RedisJobQueue,
	jwtService *services.JwtService,
	clusterService *services.ClusterService,
	userService *services.UserService,
) *ClusterHandler {
	return &ClusterHandler{
		redisQueueService,
		jwtService,
		clusterService,
		userService,
	}
}

func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	token, tokenExists := c.Get("token")
	if !tokenExists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(errors.New("unauthorized"), "unauthorized"))
		return
	}

	thisUser, err := h.userService.Me(c, token.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(errors.New("unauthorized"), "unauthorized"))
		return
	}

	addCluster := &models.AddCluster{
		ClusterId: uuid.New(),
		UserId:    thisUser.UserId,
		GroupId:   thisUser.GroupId,
		Vcpu:      thisUser.Vcpu,
		Memory:    thisUser.Memory,
		Storage:   thisUser.Storage,
		NodeSize:  thisUser.NodeSize,
	}

	err = c.BindJSON(addCluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, "request is invalid"))
		return
	}

	createClusterJob := &entities.Job{
		Payload: addCluster,
	}

	err = h.redisQueueService.AddJob(c, createClusterJob, entities.JOB_TYPE_PROVISIONING)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "could not complete the request"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("job added"))
}

func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	token, tokenExists := c.Get("token")
	if !tokenExists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(errors.New("request is invalid"), "request is invalid"))
		return
	}

	clusterId := c.Params.ByName("cluster_id")
	clusterIdUuid, err := uuid.Parse(clusterId)

	user, err := h.userService.Me(c, token.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(err, "request is invalid"))
		return
	}

	clusterEntities := &entities.Cluster{
		ClusterId: clusterIdUuid,
	}
	deleteCluster := &models.DeleteCluster{
		ClusterId: clusterIdUuid,
	}

	thisCluster, err := h.clusterService.GetClusterById(c, clusterEntities)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, NewErrorResponse(err, "not found"))
			return
		}

		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "request is invalid"))
		return
	}

	if thisCluster.UserId != user.UserId {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(errors.New("unauthorized"), "unauthorized"))
		return
	}

	deleteClusterJob := &entities.Job{
		Payload: deleteCluster,
	}

	err = h.redisQueueService.AddJob(c, deleteClusterJob, entities.JOB_TYPE_CLEANUP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "request is invalid"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("success"))
}

func (h *ClusterHandler) GetUserCluster(c *gin.Context) {
	token, tokenExists := c.Get("token")
	if !tokenExists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(errors.New("request is invalid"), "request is invalid"))
		return
	}

	userId, err := h.jwtService.GetUserIDByToken(token.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(err, "request is invalid"))
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
		c.JSON(http.StatusUnauthorized, NewErrorResponse(errors.New("request is invalid"), "request is invalid"))
		return
	}

	clusterId := c.Params.ByName("cluster_id")
	clusterIdUuid, _ := uuid.Parse(clusterId)
	userId, err := h.jwtService.GetUserIDByToken(token.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(err, "request is invalid"))
		return
	}

	clusterEntities := &entities.Cluster{
		ClusterId: clusterIdUuid,
	}

	thisCluster, err := h.clusterService.GetClusterById(c, clusterEntities)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, NewErrorResponse(err, "not found"))
			return
		}

		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "request is invalid"))
		return
	}

	if thisCluster.UserId != userId {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(errors.New("unauthorized"), "unauthorized"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponseWithData(thisCluster, "success"))
}

func (h *ClusterHandler) GetClusterKubeconfig(c *gin.Context) {
	_, tokenExists := c.Get("token")
	if !tokenExists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(errors.New("request is invalid"), "request is invalid"))
		return
	}

	clusterId := c.Params.ByName("cluster_id")
	clusterIdUuid, err := uuid.Parse(clusterId)

	clusterEntities := &entities.Cluster{
		ClusterId: clusterIdUuid,
	}

	thisCluster, err := h.clusterService.GetClusterKubeconfig(c, clusterEntities)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, NewErrorResponse(err, "not found"))
			return
		}

		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, "request is invalid"))
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=%s-kubeconfig.yaml", thisCluster.ClusterId.String()),
	)
	c.Data(200, "application/octet-stream", thisCluster.KubeconfigContents)
}
