package handlers

import (
	"net/http"
	"nodes-grpc-be/entities"
	"nodes-grpc-be/models"
	"nodes-grpc-be/services"

	"github.com/gin-gonic/gin"
)

type ClusterHandler struct {
	redisQueueService *services.RedisJobQueue
	jwtService        *services.JwtService
}

func NewClusterHandler(
	redisQueueService *services.RedisJobQueue,
	jwtService *services.JwtService,
) *ClusterHandler {
	return &ClusterHandler{
		redisQueueService: redisQueueService,
		jwtService:        jwtService,
	}
}

func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	token := c.MustGet("token").(string)
	_, err := h.jwtService.GetUserIDByToken(token)
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
