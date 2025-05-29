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
	"time"

	"google.golang.org/grpc"
)

type NodeServer struct {
	// proto rpc
	proto_model.UnimplementedNodeServiceServer

	// queue
	queue *queue.Queue
}

func NewNodeServer(
	queue *queue.Queue,
) *NodeServer {
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

	res := new(proto_model.CreateMasterResponse)

	instanceName := generateRandom(8)

	virtSpecs := virtualization_model.CreateInstanceRequest{
		Name:     instanceName,
		IsMaster: true,
		Token:    createMasterRequest.ClusterToken,
		Cpu:      createMasterRequest.Requirements.Cpu,
		Memory:   createMasterRequest.Requirements.Memory,
		Storage:  createMasterRequest.Requirements.Storage,
	}

	err := s.queue.AddToQueue(provisionCtx, virtSpecs)
	if err != nil {
		return res, err
	}

	sub := s.queue.Subscribe(ctx, instanceName)
	defer sub.Close()

	msgCh := sub.Channel()
	select {
	case msg := <-msgCh:
		res.DashboardToken = msg.Payload
	case <-time.After(PROVISIONING_TIMEOUT * time.Second):
		fmt.Println("timeout exceeded")
	}

	return res, nil
}

func (s *NodeServer) CreateWorker(
	ctx context.Context,
	createWorkerRequest *proto_model.CreateWorkerRequest,
) (*proto_model.CreateWorkerResponse, error) {
	provisionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	res := new(proto_model.CreateWorkerResponse)

	instanceName := generateRandom(8)

	virtSpecs := virtualization_model.CreateInstanceRequest{
		Name:     instanceName,
		IsMaster: false,
		Token:    createWorkerRequest.ClusterToken,
		Cpu:      createWorkerRequest.Requirements.Cpu,
		Memory:   createWorkerRequest.Requirements.Memory,
		Storage:  createWorkerRequest.Requirements.Storage,
	}

	err := s.queue.AddToQueue(provisionCtx, virtSpecs)
	if err != nil {
		return res, err
	}

	select {
	case <-time.After(PROVISIONING_TIMEOUT * time.Second):
		fmt.Println("timeout exceeded")
	}

	return res, nil
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

func StartWorker(connection *InitStruct) {
	worker := queue.NewWorker(
		connection.QueueService,
		connection.VirtualizationService,
	)
	worker.DoWork()
}

func StartWebsocket(connection *InitStruct) {
	connection.WebsocketService.Start()
}
