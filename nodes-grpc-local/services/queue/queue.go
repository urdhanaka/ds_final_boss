package queue

import (
	"context"
	"encoding/json"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"

	"github.com/redis/go-redis/v9"
)

const (
	// redis address
	REDIS_ADDRESS = "localhost:6379"

	// redis queue key
	REDIS_SPAWN_QUEUE  = "spawn-queue"
	REDIS_DELETE_QUEUE = "delete-queue"
)

type Queue struct {
	redisClient *redis.Client
}

func NewQueue(
	redisClient *redis.Client,
) *Queue {
	return &Queue{
		redisClient: redisClient,
	}
}

// add new job to queue
func (s *Queue) AddToSpawnQueue(
	ctx context.Context,
	instanceRequest virtualization_model.CreateInstanceRequest,
) error {
	requestString, err := json.Marshal(instanceRequest)
	if err != nil {
		return err
	}

	_, err = s.redisClient.LPush(ctx, REDIS_SPAWN_QUEUE, string(requestString)).Result()
	if err != nil {
		return err
	}

	return nil
}

func (s *Queue) PopSpawnQueue(
	ctx context.Context,
) (*virtualization_model.CreateInstanceRequest, error) {
	instanceRequest := new(virtualization_model.CreateInstanceRequest)

	job, err := s.redisClient.BRPop(ctx, 0, REDIS_SPAWN_QUEUE).Result()
	if err != nil {
		return instanceRequest, err
	}

	// job value is the second element at the array
	err = json.Unmarshal([]byte(job[1]), instanceRequest)
	if err != nil {
		return instanceRequest, err
	}

	return instanceRequest, nil
}

func (s *Queue) AddToDeleteQueue(
	ctx context.Context,
	deleteRequest virtualization_model.DeleteInstanceRequest,
) error {
	requestString, err := json.Marshal(deleteRequest)
	if err != nil {
		return err
	}

	_, err = s.redisClient.LPush(ctx, REDIS_DELETE_QUEUE, string(requestString)).Result()
	if err != nil {
		return err
	}

	return nil
}

func (s *Queue) PopDeleteQueue(
	ctx context.Context,
) (*virtualization_model.DeleteInstanceRequest, error) {
	deleteRequest := new(virtualization_model.DeleteInstanceRequest)

	job, err := s.redisClient.BRPop(ctx, 0, REDIS_DELETE_QUEUE).Result()
	if err != nil {
		return deleteRequest, err
	}

	// job value is the second element at the array
	err = json.Unmarshal([]byte(job[1]), deleteRequest)
	if err != nil {
		return deleteRequest, err
	}

	return deleteRequest, nil
}

func (s *Queue) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return s.redisClient.Subscribe(ctx, channel)
}

func (s *Queue) Publish(
	ctx context.Context,
	channel string,
	message any,
) error {
	_, err := s.redisClient.Publish(ctx, channel, message).Result()
	if err != nil {
		return err
	}

	return nil
}
