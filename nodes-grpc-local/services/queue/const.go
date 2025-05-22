package queue

const (
	// valkey hostname
	VALKEY_HOST = "localhost"
	// valkey port
	VALKEY_PORT = 6379
	// valkey main queue
	VALKEY_MAIN_QUEUE = "spawn-queue"
	// valkey processing queue
	VALKEY_PROCESSING_QUEUE = "spawn-queue-backup"

	// maximum worker that can handle the queue
	MAX_WORKER_SIZE int = 3
)
