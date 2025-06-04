package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	proto_model "nodes-grpc-local/services/model/proto-model"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"nodes-grpc-local/services/queue"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"
	"os"
	"time"

	"github.com/gorilla/websocket"
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

	instanceName := createMasterRequest.Requirements.NodeName

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

	go sendLogs(instanceName, createMasterRequest.ClusterName)

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

	instanceName := createWorkerRequest.Requirements.NodeName

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

	go sendLogs(instanceName, createWorkerRequest.ClusterName)

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
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/register_node", MAIN_SERVER_URL_RPL), bytes.NewBuffer(body))
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

func sendLogs(
	instanceName string,
	clusterName string,
) {
	var sock net.Conn
	var err error

	logSocketFile := libvirt_virtualization.INSTANCE_LOGS_DIR + "/" + instanceName + ".sock"

	for {
		sock, err = net.Dial("unix", logSocketFile)
		if err != nil {
			slog.Error(fmt.Sprintf("%s | error accessing socket file, retrying...", instanceName),
				"error", err,
			)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	defer sock.Close()

	u := url.URL{
		Scheme: "ws",
		Host:   MAIN_SERVER_URL_RPL,
		Path:   fmt.Sprintf("/ws/receive_logs/%s", clusterName),
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		slog.Error(fmt.Sprintf("%s | error dialing websocket", instanceName),
			"error", err,
		)
		return
	}
	defer c.Close()

	scanner := bufio.NewScanner(sock)
	for scanner.Scan() {
		line := scanner.Text()
		if err := c.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			slog.Error("Send error:", "error", err)
			break
		}
	}
}
