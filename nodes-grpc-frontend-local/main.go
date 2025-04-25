package main

import (
	"context"
	"fmt"
	"log/slog"
	"nodes-grpc-frontend-local/model/proto_model"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newNodeClient() (proto_model.NodeServiceClient, error) {
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

func main() {
	nodeClient, err := newNodeClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = nodeClient.CreateMaster(context.Background(), &proto_model.CreateMasterRequest{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
