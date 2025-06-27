package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	consts "nodes-grpc-be/const"
	"nodes-grpc-be/entities"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	MAX_JOB_RETRIES = 3
)

type RedisJobQueue struct {
	client         *redis.Client
	queueName      string
	ctx            context.Context
	workers        int
	cancel         context.CancelFunc
	retryDelay     time.Duration
	clusterService *ClusterService
}

func NewRedisJobQueue(
	redisClient *redis.Client,
	clusterService *ClusterService,
) *RedisJobQueue {
	ctx, cancel := context.WithCancel(context.Background())

	jq := &RedisJobQueue{
		client:         redisClient,
		queueName:      consts.REDIS_NAME,
		ctx:            ctx,
		cancel:         cancel,
		retryDelay:     30 * time.Second,
		clusterService: clusterService,
	}

	return jq
}

func (jq *RedisJobQueue) AddJob(job *entities.Job) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	job.Status = entities.JOB_QUEUED
	job.CreatedAt = time.Now()
	job.MaxRetries = MAX_JOB_RETRIES

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	jobKey := fmt.Sprintf("job:%s", job.ID)
	if err := jq.client.Set(jq.ctx, jobKey, jobData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store job: %w", err)
	}

	if err := jq.client.LPush(jq.ctx, jq.queueName, job.ID).Err(); err != nil {
		return fmt.Errorf("failed to queue job: %w", err)
	}

	slog.Info(fmt.Sprintf("Job %s added to queue", job.ID))
	return nil
}

func (jq *RedisJobQueue) GetJob(id string) (*entities.Job, error) {
	jobKey := fmt.Sprintf("job:%s", id)
	jobData, err := jq.client.Get(jq.ctx, jobKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("job not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	var job entities.Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

func (jq *RedisJobQueue) updateJob(job *entities.Job) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	jobKey := fmt.Sprintf("job:%s", job.ID)
	return jq.client.Set(jq.ctx, jobKey, jobData, 24*time.Hour).Err()
}

func (jq *RedisJobQueue) processJob(job *entities.Job) {
	job.Status = entities.JOB_WORKING
	jq.updateJob(job)

	var result any
	var err error

	switch job.Type {
	case entities.JOB_PROVISIONING:
		// err = jq.clusterSvc.CreateCluster(jq.ctx, job)
		err = fmt.Errorf("unknown job type: %s", job.Type)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}

	now := time.Now()
	job.CompletedAt = &now

	if err != nil {
		job.Retries++
		if job.Retries < job.MaxRetries {
			job.Status = entities.JOB_RETRYING
			job.Error = err

			retryKey := fmt.Sprintf("retry:%d", time.Now().Add(jq.retryDelay).Unix())
			jq.client.LPush(jq.ctx, retryKey, job.ID)
			jq.client.Expire(jq.ctx, retryKey, jq.retryDelay+time.Minute)
		} else {
			job.Status = entities.JOB_FAILED
			job.Error = err
		}
	} else {
		job.Status = entities.JOB_DONE
		job.Result = result
	}

	jq.updateJob(job)
	slog.Info(
		fmt.Sprintf("Job %s completed with status %s", job.ID, job.Status),
	)
}

// worker function
func (jq *RedisJobQueue) worker(id int) {
	slog.Info(fmt.Sprintf("Worker %d started", id))

	for {
		select {
		case <-jq.ctx.Done():
			slog.Info("Worker stopping...")
			return
		default:
			// block for queue pop with 1 second timeout
			result, err := jq.client.BRPop(jq.ctx, 1*time.Second, jq.queueName).Result()
			if err == redis.Nil {
				continue // timeout, try again
			} else if err != nil {
				slog.Error("Worker error",
					"error", err,
				)
				continue
			}

			jobID := result[1]
			job, err := jq.GetJob(jobID)
			if err != nil {
				slog.Error(
					fmt.Sprintf("Worker %d failed to get job %s", id, jobID),
					"error", err,
				)
				continue
			}

			slog.Info(fmt.Sprintf("Worker %d processing job %s", id, jobID))
			jq.processJob(job)
		}
	}
}
