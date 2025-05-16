package queue

import (
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"time"
)

const (
	DEFAULT_JOB_TIMEOUT = 10
	DEFAULT_JOB_RETRIES = 3
	DEFAULT_JOB_BACKOFF = 20
)

type Job struct {
	// create instance request
	Request virtualization_model.CreateInstanceRequest

	// how many retries if the process
	// could not be done
	Retries int

	// how many seconds before process
	// is considered failed
	Timeout time.Duration

	// time in seconds between retries
	Backoff time.Duration
}

func NewJob(request virtualization_model.CreateInstanceRequest) *Job {
	return &Job{
		Request: request,
		Retries: DEFAULT_JOB_RETRIES,
		Timeout: DEFAULT_JOB_TIMEOUT,
		Backoff: DEFAULT_JOB_BACKOFF,
	}
}
