package config

import (
	"log/slog"
	"nodes-grpc-frontend-local/src/model/proto_model"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewNodeClient() (proto_model.NodeServiceClient, error) {
	fullUrl := "localhost:50051"

	conn, err := grpc.NewClient(fullUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Could not connect to grpc server",
			"url", fullUrl,
			"error", err,
		)

		return nil, err
	}

	return proto_model.NewNodeServiceClient(conn), nil
}
