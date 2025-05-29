package queue

const (
    // redis address
    REDIS_ADDRESS = "localhost:6379"
    // main redis queue key
    REDIS_MAIN_QUEUE = "spawn-queue"

	// valkey hostname
	VALKEY_HOST = "localhost"
	// valkey port
	VALKEY_PORT = 6379
	// valkey main queue
	VALKEY_MAIN_QUEUE = "spawn-queue"
	// valkey processing queue
	VALKEY_PROCESSING_QUEUE = "spawn-queue-backup"
)
