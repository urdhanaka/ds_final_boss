package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
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
		instanceRes := new(virtualization_model.VirtCreateInstanceResponse)

		_ = json.Unmarshal([]byte(msg.Payload), instanceRes)

		res.DashboardToken = instanceRes.DashboardToken
		res.MasterIpAddress = instanceRes.MasterIpAddress

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

	// background job
	go startWorker(connection)
	go startWebsocket(connection)

	// connect to main server
	slog.Info("node is ready, connecting to main server")
	res, err := connectToServer()
	if err != nil {
		slog.Error("StartGrpcServer(): failed to connect to main server",
			"error", err.Error(),
		)
		os.Exit(1)
	}
	slog.Info("main server is responding",
		"response", res)

	slog.Info(fmt.Sprintf("starting grpc server at %s", GRPC_PORT))

	if err = s.Serve(lis); err != nil {
		slog.Error("StartGrpcServer(): failed to serve",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}

func connectToServer() (string, error) {
	hostname := getHostname()
	ipAddress := getIpAddress()

	body := []byte(fmt.Sprintf(`{"hostname":"%s","ip_address":"%s"}`, hostname, ipAddress))
	req, err := http.NewRequest("POST", "http://localhost:3000/register_node", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	return string(respBody), nil
}

func startWorker(connection *InitStruct) {
	worker := queue.NewWorker(
		connection.QueueService,
		connection.VirtualizationService,
	)
	worker.DoWork()
}

func startWebsocket(connection *InitStruct) {
	connection.WebsocketService.Start()
}
