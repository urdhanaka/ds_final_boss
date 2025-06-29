package config

import (
	"log/slog"
	"nodes-grpc-be/models/proto_model"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	CONNECT_GRPC_COUNT = 3
)

func NewNodeClient(nodeIp string) (proto_model.NodeServiceClient, error) {
	var conn *grpc.ClientConn
	var err error

	for attempt := 1; attempt <= CONNECT_GRPC_COUNT; attempt++ {
		slog.Info("trying to connect to worker node",
			"ip address", nodeIp,
		)

		conn, err = grpc.NewClient(nodeIp, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			slog.Error("Could not connect to grpc server, retrying...",
				"url", nodeIp,
				"error", err,
			)

			continue
		}
	}

	if err != nil {
		slog.Error("Could not connect to grpc server, node might be down",
			"url", nodeIp,
			"error", err,
		)

		return nil, err
	}

	slog.Info("grpc client connected",
		"ip address", nodeIp,
	)

	return proto_model.NewNodeServiceClient(conn), nil
}
