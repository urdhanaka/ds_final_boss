package entities

type (
	JobStatus string
	JobType   string
	RpcStatus string
)

const (
	JOB_STATUS_QUEUED   JobStatus = "queued"
	JOB_STATUS_WORKING  JobStatus = "working"
	JOB_STATUS_DONE     JobStatus = "done"
	JOB_STATUS_RETRYING JobStatus = "retrying"
	JOB_STATUS_FAILED   JobStatus = "failed"

	NODE_NOT_ENOUGH_RESOURCES RpcStatus = "NODE_NOT_ENOUGH_RESOURCES"

	JOB_TYPE_PROVISIONING      JobType = "provision_job"
	JOB_TEST_TYPE_PROVISIONING JobType = "provision_job_test"
	JOB_TYPE_CLEANUP           JobType = "cleanup_job"
)

type Job struct {
	ID         string
	Type       JobType
	Status     JobStatus
	Payload    any
	Result     any
	Error      error
	Retries    int
	MaxRetries int
}
