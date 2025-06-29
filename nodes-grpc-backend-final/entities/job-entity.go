package entities

import (
	"time"
)

type (
	JobStatus string
	JobType   string
)

const (
	JOB_STATUS_QUEUED   JobStatus = "queued"
	JOB_STATUS_WORKING  JobStatus = "working"
	JOB_STATUS_DONE     JobStatus = "done"
	JOB_STATUS_RETRYING JobStatus = "retrying"
	JOB_STATUS_FAILED   JobStatus = "failed"

	JOB_TYPE_PROVISIONING JobType = "provision_job"
	JOB_TYPE_CLEANUP      JobType = "cleanup_job"
)

type Job struct {
	ID          string
	Type        JobType
	Status      JobStatus
	Payload     any
	Result      any
	Error       error
	Retries     int
	MaxRetries  int
	CreatedAt   time.Time
	CompletedAt *time.Time
}
