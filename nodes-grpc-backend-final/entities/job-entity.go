package entities

import "time"

type (
	JobStatus string
	JobType   string
)

const (
	JOB_QUEUED   JobStatus = "queued"
	JOB_WORKING  JobStatus = "working"
	JOB_DONE     JobStatus = "done"
	JOB_RETRYING JobStatus = "retrying"
	JOB_FAILED   JobStatus = "failed"

	JOB_PROVISIONING JobType = "provisioning"
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
