package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	proto_model "nodes-grpc-local/services/model/proto-model"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"nodes-grpc-local/services/queue"
	"os"

	"google.golang.org/grpc"
)

const (
	// port to listen
	GRPC_PORT = ":50051"
)

type NodeServer struct {
	// proto rpc
	proto_model.UnimplementedNodeServiceServer

	// virtualization service
	dispatcher *queue.Dispatcher
}

func NewNodeServer(
	dispatcher *queue.Dispatcher,
) NodeServerInterface {
	return &NodeServer{
		dispatcher: dispatcher,
	}
}

func (s *NodeServer) CreateMaster(
	ctx context.Context,
	createMasterRequest *proto_model.CreateMasterRequest,
) (*proto_model.CreateMasterResponse, error) {
	instanceRequest := virtualization_model.CreateInstanceRequest{
		IsMaster:        true,
		Token:           createMasterRequest.Token,
		MasterIpAddress: "",
		Cpu:             createMasterRequest.Requirements.Cpu,
		Memory:          createMasterRequest.Requirements.Memory,
		Storage:         createMasterRequest.Requirements.Storage,
	}

	err := s.dispatcher.AddJob(ctx, instanceRequest)
	if err != nil {
		return new(proto_model.CreateMasterResponse), err
	}

	return new(proto_model.CreateMasterResponse), nil
}

func (s *NodeServer) CreateWorker(
	ctx context.Context,
	createWorkerRequest *proto_model.CreateWorkerRequest,
) (*proto_model.CreateWorkerResponse, error) {
	instanceRequest := virtualization_model.CreateInstanceRequest{
		IsMaster:        false,
		Token:           createWorkerRequest.Token,
		MasterIpAddress: createWorkerRequest.IpAddress,
		Cpu:             createWorkerRequest.Requirements.Cpu,
		Memory:          createWorkerRequest.Requirements.Memory,
		Storage:         createWorkerRequest.Requirements.Storage,
	}

	err := s.dispatcher.AddJob(ctx, instanceRequest)
	if err != nil {
		return new(proto_model.CreateWorkerResponse), err
	}

	return new(proto_model.CreateWorkerResponse), nil
}

func StartGrpcServer(connection *InitStruct) {
	lis, err := net.Listen("tcp", GRPC_PORT)
	if err != nil {
		slog.Error("could not start grpc server",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	s := grpc.NewServer()
	proto_model.RegisterNodeServiceServer(s, &NodeServer{
		dispatcher: connection.DispatcherService,
	})

	slog.Info(fmt.Sprintf("starting grpc server at %s", GRPC_PORT))

	if err := s.Serve(lis); err != nil {
		slog.Error("StartGrpcServer(): failed to serve",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
