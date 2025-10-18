package services

import (
	"context"
	"encoding/json"
	"nodes-grpc-backend-local/model"

	"github.com/redis/go-redis/v9"
)

const (
	REDIS_SPAWN_QUEUE = "create-cluster-queue"
)

type QueueService struct {
	redisClient *redis.Client
}

func NewQueueService(redisClient *redis.Client) *QueueService {
	return &QueueService{
		redisClient,
	}
}

func (s *QueueService) AddClusterToQueue(
	ctx context.Context,
	cluster *model.AddCluster,
) error {
	requestString, err := json.Marshal(cluster)
	if err != nil {
		return err
	}

	_, err = s.redisClient.LPush(ctx, REDIS_SPAWN_QUEUE, string(requestString)).Result()
	if err != nil {
		return err
	}

	return nil
}

func (s *QueueService) PopClusterFromQueue(
	ctx context.Context,
) (*model.AddCluster, error) {
	cluster := new(model.AddCluster)

	job, err := s.redisClient.BRPop(ctx, 0, REDIS_SPAWN_QUEUE).Result()
	if err != nil {
		return cluster, err
	}

	err = json.Unmarshal([]byte(job[1]), cluster)
	if err != nil {
		return cluster, err
	}

	return cluster, nil
}
