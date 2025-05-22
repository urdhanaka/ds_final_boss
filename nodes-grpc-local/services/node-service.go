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

	// queue
	queue *queue.Queue
}

func NewNodeServer(
	queue *queue.Queue,
) NodeServerInterface {
	return &NodeServer{
		queue: queue,
	}
}

func (s *NodeServer) CreateMaster(
	ctx context.Context,
	createMasterRequest *proto_model.CreateMasterRequest,
) (*proto_model.CreateMasterResponse, error) {
	provisionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	instanceName := generateRandom(16)

	instanceRequest := virtualization_model.CreateInstanceRequest{
		Name:            instanceName,
		IsMaster:        true,
		Token:           createMasterRequest.Token,
		MasterIpAddress: "",
		Cpu:             createMasterRequest.Requirements.Cpu,
		Memory:          createMasterRequest.Requirements.Memory,
		Storage:         createMasterRequest.Requirements.Storage,
	}

	err := s.queue.AddToQueue(provisionCtx, instanceRequest)
	if err != nil {
		return new(proto_model.CreateMasterResponse), err
	}

	return new(proto_model.CreateMasterResponse), nil
}

func (s *NodeServer) CreateWorker(
	ctx context.Context,
	createWorkerRequest *proto_model.CreateWorkerRequest,
) (*proto_model.CreateWorkerResponse, error) {
	provisionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	instanceName := generateRandom(16)

	instanceRequest := virtualization_model.CreateInstanceRequest{
		Name:            instanceName,
		IsMaster:        false,
		Token:           createWorkerRequest.Token,
		MasterIpAddress: createWorkerRequest.IpAddress,
		Cpu:             createWorkerRequest.Requirements.Cpu,
		Memory:          createWorkerRequest.Requirements.Memory,
		Storage:         createWorkerRequest.Requirements.Storage,
	}

	err := s.queue.AddToQueue(provisionCtx, instanceRequest)
	if err != nil {
		return new(proto_model.CreateWorkerResponse), err
	}

	return new(proto_model.CreateWorkerResponse), nil
}

func (s *NodeServer) CreateInstance(
	ctx context.Context,
	createInstanceRequest *proto_model.CreateInstanceRequest,
) (*proto_model.CreateInstanceResponse, error) {
	provisionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	instanceName := generateRandom(16)

	virtSpecs := virtualization_model.CreateInstanceRequest{
		Name:            instanceName,
		IsMaster:        createInstanceRequest.IsMaster,
		Token:           createInstanceRequest.Token,
		MasterIpAddress: createInstanceRequest.IpAddress,
		Cpu:             createInstanceRequest.Cpu,
		Memory:          createInstanceRequest.Memory,
		Storage:         createInstanceRequest.Storage,
	}

	err := s.queue.AddToQueue(provisionCtx, virtSpecs)
	if err != nil {
		return new(proto_model.CreateInstanceResponse), err
	}

	return &proto_model.CreateInstanceResponse{}, nil
}

func StartGrpcServer(connection *InitStruct) {
	lis, err := net.Listen("tcp", GRPC_PORT)
	if err != nil {
		slog.Error("could not start grpc server",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	// start queue
	for i := 1; i <= queue.MAX_WORKER_SIZE; i++ {
		go func() {
			valkeyClient := queue.InitValkeyConnection()
			worker := queue.NewWorker(valkeyClient, connection.VirtualizationService)
			worker.DoWork(context.Background())
		}()
	}

    // start the websocket
    go connection.WebsocketService.Start()

	s := grpc.NewServer()
	proto_model.RegisterNodeServiceServer(s, &NodeServer{
		queue: connection.QueueService,
	})

	slog.Info(fmt.Sprintf("starting grpc server at %s", GRPC_PORT))

	if err := s.Serve(lis); err != nil {
		slog.Error("StartGrpcServer(): failed to serve",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
