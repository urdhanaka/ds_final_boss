package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"nodes-grpc/common/config"
	"nodes-grpc/common/model"
	"os"
	"os/exec"

	"google.golang.org/grpc"
)

type NodeServer struct {
	model.UnimplementedNodeServer
}

type RegisterThisNode struct {
	Hostname  string `json:"hostname"`
	IpAddress string `json:"ip_address"`
	GrpcPort  string `json:"grpc_port"`
}

var (
	localStorage *model.NodeList
	thisNode     config.NodeIdentification
)

func init() {
	localStorage = new(model.NodeList)
	localStorage.Nodes = make([]*model.NodeStatus, 0)

	thisNode = config.ThisNodeIdentification()
}

func connectToMainServer(mainServerIPAddress string) error {
	bodyBytes, _ := json.Marshal(thisNode)
	bodyReader := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest(http.MethodPost, mainServerIPAddress, bodyReader)
	if err != nil {
		slog.Error("Error connecting main server",
			"error", err.Error())

		return err
	}

	req.Header.Set("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Error getting main server response",
			"error", err.Error())

		return err
	}

	return nil
}

func (NodeServer) Heartbeat(context.Context, *model.Empty) (*model.NodeStatus, error) {
	nodeStatus := model.NodeStatus{
		Hostname:  thisNode.Hostname,
		IpAddress: thisNode.IpAddress,
		GrpcPort:  thisNode.GrpcPort,
		Status:    model.Status(model.Status_STATUS_AVAILABLE),
	}

	return &nodeStatus, nil
}

func (NodeServer) StartMaster(_ context.Context, serverToken *model.ServerToken) (*model.Empty, error) {
	command := config.StartCommand(
		true,                   // is this node the master node?
		serverToken.GetToken(), // string of token
		"",                     // IP of master node (used if this node is the worker node)
	)

	cmd := exec.Command("bash", "-c",
		command,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return new(model.Empty), err
	}

	return new(model.Empty), nil
}

func (NodeServer) StartWorker(_ context.Context, masterNode *model.MasterNode) (*model.Empty, error) {
	command := config.StartCommand(
		false,                // is this node the master node?
		masterNode.NodeToken, // string of token
		masterNode.IpAddress, // IP of master node (used if this node is the worker node)
	)

	cmd := exec.Command("bash", "-c",
		command,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return new(model.Empty), err
	}

	return new(model.Empty), nil
}

func (NodeServer) GetAllNode(context.Context, *model.Empty) (*model.NodeList, error) {
	return new(model.NodeList), nil
}

func main() {
	mainServerAddress := os.Args[1]
	if mainServerAddress == "" {
		fmt.Println("Provide main server IP address")
		os.Exit(1)
	}

	var nodeSrv NodeServer
	srv := grpc.NewServer()
	model.RegisterNodeServer(srv, nodeSrv)

	slog.Info("Starting node server...",
		"port", config.NodeServiceGRPCPort,
	)

	l, err := net.Listen("tcp", ":"+config.NodeServiceGRPCPort)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = connectToMainServer(mainServerAddress)
	if err != nil {
		slog.Error("Could not start node service",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	log.Fatal(srv.Serve(l))
}
