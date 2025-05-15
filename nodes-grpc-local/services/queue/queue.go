package queue

import "context"

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
