package queue

import "context"

const (
	// maximum size of the queue line
	//
	// if the queue currently have MAX_QUEUE_SIZE and
	// tried to be inserted, return error
	MAX_QUEUE_SIZE int = 20

	// maximum worker that can handle the queue
	MAX_WORKER_SIZE int = 3
)

type Queue struct {
	jobs chan *Job
}

func NewQueue() *Queue {
	return &Queue{
		jobs: make(chan *Job, MAX_QUEUE_SIZE),
	}
}

// TryAdd will try to add the job to the queue
//
// return true if job is successfully added,
// return false otherwise
func (s *Queue) TryAdd(ctx context.Context, job *Job) bool {
	select {
	case s.jobs <- job:
		return true
	default:
		return false
	}
}
