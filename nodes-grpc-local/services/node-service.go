package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	proto_model "nodes-grpc-local/services/model/proto-model"
	"nodes-grpc-local/services/virtualization"
	"os"

	"google.golang.org/grpc"
)

const (
	GRPC_PORT = ":50051"
)

type NodeServer struct {
	// proto rpc
	proto_model.UnimplementedNodeServiceServer

	// virtualization service
	virtualizationService virtualization.VirtualizationInterface
}

func NewNodeServer(
	virtualizationService virtualization.VirtualizationInterface,
) *NodeServer {
	return &NodeServer{
		virtualizationService: virtualizationService,
	}
}

func (s *NodeServer) CreateMaster(
	ctx context.Context,
	createMasterRequest *proto_model.CreateMasterRequest,
) (*proto_model.CreateMasterResponse, error) {
	return new(proto_model.CreateMasterResponse), nil
}

func (n *NodeServer) CreateWorker(
	c context.Context,
	workerRequest *proto_model.CreateWorkerRequest,
) (*proto_model.CreateWorkerResponse, error) {
	return new(proto_model.CreateWorkerResponse), nil
}

func StartGrpcServer(connection *Connection) {
	lis, err := net.Listen("tcp", GRPC_PORT)
	if err != nil {
		slog.Error("could not start grpc server",
			"error", err.Error(),
		)
	}

	s := grpc.NewServer()
	proto_model.RegisterNodeServiceServer(s, &NodeServer{
		// NOTE: CHANGE VIRTUALIZATION METHOD HERE
		virtualizationService: connection.VirtualizationService,
	})

	slog.Info(fmt.Sprintf("starting server at %s", GRPC_PORT))

	if err := s.Serve(lis); err != nil {
		slog.Error("StartGrpcServer(): failed to serve",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
