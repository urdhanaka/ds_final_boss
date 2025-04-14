package config

import (
	"fmt"
	"log/slog"
	"nodes-grpc-frontend/common/model/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewNodeClient(ip_address, port string) (proto.NodeServiceClient, error) {
	fullUrl := fmt.Sprintf("%s:%s", ip_address, port)

	conn, err := grpc.NewClient(fullUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Could not connect to grpc server",
			"url", fullUrl,
			"error", err)

		return nil, err
	}

	return proto.NewNodeServiceClient(conn), nil
}
