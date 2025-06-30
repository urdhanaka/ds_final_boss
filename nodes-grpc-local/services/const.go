package services

const (
	GRPC_PORT string = ":50051"

	MAIN_SERVER_URL_LOCAL = "localhost:3000"
	MAIN_SERVER_URL_RPL   = "10.21.73.113:8000"

	PROVISIONING_TIMEOUT = 600
	PROVISIONING_TIME    = 300

	// cpu usage interval value in seconds
	CPU_USAGE_INTERVAL = 1

    // retries when connecting to main server
    NODES_CONNECT_RETRIES = 3
)
