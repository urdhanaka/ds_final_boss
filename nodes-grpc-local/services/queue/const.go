package queue

const (
	// maximum size of the queue line
	//
	// if the queue currently have MAX_QUEUE_SIZE and
	// tried to be inserted, return error
	MAX_QUEUE_SIZE int = 20

	// maximum worker that can handle the queue
	MAX_WORKER_SIZE int = 3
)
