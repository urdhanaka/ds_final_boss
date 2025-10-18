package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	consts "nodes-grpc-be/const"
	"nodes-grpc-be/entities"
	"nodes-grpc-be/models"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	MAX_JOB_RETRIES  = 3
	PROVISION_WORKER = 2
	// CLEANUP_WORKER   = 1
)

type RedisJobQueue struct {
	client                *redis.Client
	provisioningQueueName string
	cleanupQueueName      string
	retryDelay            time.Duration
	clusterService        *ClusterService
}

func NewRedisJobQueue(
	redisClient *redis.Client,
	clusterService *ClusterService,
) *RedisJobQueue {
	jq := &RedisJobQueue{
		client:                redisClient,
		provisioningQueueName: consts.REDIS_PROVISION_NAME,
		cleanupQueueName:      consts.REDIS_CLEANUP_NAME,
		retryDelay:            30 * time.Second,
		clusterService:        clusterService,
	}

	return jq
}

func (jq *RedisJobQueue) AddJob(
	ctx context.Context,
	job *entities.Job,
	jobType entities.JobType,
) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	job.Type = jobType
	job.Status = entities.JOB_STATUS_QUEUED
	job.Retries = 0
	job.MaxRetries = MAX_JOB_RETRIES

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	jobKey := fmt.Sprintf("job:%s", job.ID)
	if err := jq.client.Set(ctx, jobKey, jobData, 24*time.Hour).Err(); err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to store job: %w", err)
	}

	// insert according to the job type
	// only 2 types of job.Type
	switch job.Type {
	case entities.JOB_TYPE_PROVISIONING:
		addClusterModel := job.Payload.(*models.AddCluster)
		err = jq.clusterService.AddClusterToDatabase(ctx, &entities.Cluster{
			ClusterId:     addClusterModel.ClusterId,
			ClusterName:   addClusterModel.ClusterName,
			UserId:        addClusterModel.UserId,
			GroupId:       addClusterModel.GroupId,
			ClusterStatus: string(job.Status),
			CreatedAt:     time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to queue job: %w", err)
		}

		if err := jq.client.LPush(ctx, jq.provisioningQueueName, job.ID).Err(); err != nil {
			return fmt.Errorf("failed to queue job: %w", err)
		}
	case entities.JOB_TYPE_CLEANUP:
		if err := jq.client.LPush(ctx, jq.cleanupQueueName, job.ID).Err(); err != nil {
			return fmt.Errorf("failed to queue job: %w", err)
		}
	}

	slog.Info(fmt.Sprintf("Job %s added to queue", job.ID))
	return nil
}

func (jq *RedisJobQueue) GetJob(ctx context.Context, id string) (*entities.Job, error) {
	jobKey := fmt.Sprintf("job:%s", id)
	jobData, err := jq.client.Get(ctx, jobKey).Result()
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

func (jq *RedisJobQueue) StartWorker(ctx context.Context) {
	for i := range PROVISION_WORKER {
		go jq.provisionWorker(ctx, i)
	}

	go jq.cleanupWorker(ctx)
}

func (jq *RedisJobQueue) updateJobStatus(ctx context.Context, job *entities.Job) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}
	status := job.Status
	jobPop := new(models.AddCluster)

	jobPayloadBytes, _ := json.Marshal(job.Payload)
	err = json.Unmarshal(jobPayloadBytes, jobPop)

	updatedClusterEntity := &entities.Cluster{
		ClusterId:     jobPop.ClusterId,
		ClusterStatus: string(status),
	}

	// update cluster status on database
	err = jq.clusterService.UpdateClusterStatusByClusterId(ctx, updatedClusterEntity)
	if err != nil {
		return err
	}

	jobKey := fmt.Sprintf("job:%s", job.ID)
	return jq.client.Set(ctx, jobKey, jobData, 24*time.Hour).Err()
}

func (jq *RedisJobQueue) processJob(
	ctx context.Context,
	job *entities.Job,
	workerId int,
) {
	startTime := time.Now()
	job.Status = entities.JOB_STATUS_WORKING
	jq.updateJobStatus(ctx, job)

	slog.Info(fmt.Sprintf("Job %s started processing", job.ID))

	var result any
	var err error

	switch job.Type {
	// case entities.JOB_TEST_TYPE_PROVISIONING:
	// 	result, err = jq.clusterService.CreateClusterWithoutPickTest(ctx, job)
	case entities.JOB_TEST_TYPE_PROVISIONING:
		result, err = jq.clusterService.CreateClusterFinalTest(ctx, job, workerId)
	case entities.JOB_TYPE_PROVISIONING:
		result, err = jq.clusterService.CreateClusterWithPick(ctx, job)
	// case entities.JOB_TYPE_PROVISIONING: // NOTE: WITHOUT PICK
	// 	result, err = jq.clusterService.CreateClusterWithoutPick(ctx, job)
	case entities.JOB_TYPE_CLEANUP:
		err = jq.clusterService.CleanCluster(ctx, job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}

	duration := time.Since(startTime)

	if err != nil {
		slog.Error(fmt.Sprintf("Job %s failed after %v", job.ID, duration), "error", err)
		job.Retries++

		if job.Retries < job.MaxRetries {
			// update job status
			jq.updateJobStatus(ctx, job)

			// push the
			jq.client.LPush(ctx, jq.provisioningQueueName, job.ID)
		} else {
			job.Status = entities.JOB_STATUS_FAILED
			job.Error = err
		}
	} else {
		job.Status = entities.JOB_STATUS_DONE
		job.Result = result
		slog.Info(fmt.Sprintf("Job %s completed successfully in %v", job.ID, duration))
	}

	jq.updateJobStatus(ctx, job)
}

// worker function
func (jq *RedisJobQueue) provisionWorker(ctx context.Context, workerId int) {
	slog.Info("provision worker started",
		"id", workerId,
	)

	for {
		select {
		case <-ctx.Done():
			slog.Info("provision worker stopping...")
			return
		default:
			// block for queue pop with 1 second timeout
			result, err := jq.client.BRPop(ctx, 1*time.Second, jq.provisioningQueueName).Result()
			if err == redis.Nil {
				continue // timeout, try again
			} else if err != nil {
				slog.Error("provision worker error",
					"error", err,
				)
				continue
			}

			jobID := result[1]
			job, err := jq.GetJob(ctx, jobID)
			if err != nil {
				slog.Error(
					fmt.Sprintf("provision worker failed to get job %s", jobID),
					"error", err,
				)
				continue
			}

			slog.Info(fmt.Sprintf("provision worker processing job %s", jobID))
			jq.processJob(ctx, job, workerId)
		}
	}
}

func (jq *RedisJobQueue) cleanupWorker(ctx context.Context) {
	slog.Info("cleanup worker started")

	for {
		select {
		case <-ctx.Done():
			slog.Info("provision worker stopping...")
			return
		default:
			// block for queue pop with 1 second timeout
			result, err := jq.client.BRPop(ctx, 1*time.Second, jq.cleanupQueueName).Result()
			if err == redis.Nil {
				continue // timeout, try again
			} else if err != nil {
				slog.Error("cleanup worker error",
					"error", err,
				)
				continue
			}

			jobID := result[1]
			job, err := jq.GetJob(ctx, jobID)
			if err != nil {
				slog.Error(
					fmt.Sprintf("cleanup worker failed to get job %s", jobID),
					"error", err,
				)
				continue
			}

			slog.Info(fmt.Sprintf("cleanup worker processing job %s", jobID))
			jq.processJob(ctx, job, 1)
		}
	}
}
