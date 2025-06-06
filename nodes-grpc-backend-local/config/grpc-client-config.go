package config

import (
	"log/slog"
	"nodes-grpc-backend-local/model/proto_model"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewNodeClient(nodeIp string) (proto_model.NodeServiceClient, error) {
	conn, err := grpc.NewClient(nodeIp, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Could not connect to grpc server",
			"url", nodeIp,
			"error", err,
		)

		return nil, err
	}

	return proto_model.NewNodeServiceClient(conn), nil
}
