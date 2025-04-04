package web

import "github.com/google/uuid"

type Node struct {
	UUID      uuid.UUID `json:"uuid"`
	IpAddress string    `json:"ip_address"`
	GrpcPort  string    `json:"grpc_port"`
}
