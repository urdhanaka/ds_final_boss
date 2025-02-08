package usecases

import (
	"context"
	"log/slog"
	"nodes-grpc-frontend/common/config"
	proto_model "nodes-grpc-frontend/common/model/proto-model"
	web_model "nodes-grpc-frontend/common/model/web-model"
	"time"
)

type NodeUsecase interface {
	GetAllNodes(context.Context) *proto_model.NodeList
	RegisterNode(_ context.Context, nodeModel *web_model.RegisterNodeRequest) error
	CreateCluster(ctx context.Context, tokenModel *proto_model.ServerToken) error
}

type nodeUsecaseImpl struct{}

func NewNodeUsecaseImpl() NodeUsecase {
	return &nodeUsecaseImpl{}
}

var localStorage *proto_model.NodeList

func init() {
	localStorage = new(proto_model.NodeList)
	localStorage.Nodes = make([]*proto_model.NodeStatus, 0)
}

func (u *nodeUsecaseImpl) GetAllNodes(context.Context) *proto_model.NodeList {
	return localStorage
}

func (u *nodeUsecaseImpl) RegisterNode(_ context.Context, nodeModel *web_model.RegisterNodeRequest) error {
	nodeGrpcModel := new(proto_model.NodeStatus)

	nodeGrpcModel.Hostname = nodeModel.Hostname
	nodeGrpcModel.IpAddress = nodeModel.IpAddress
	nodeGrpcModel.GrpcPort = nodeModel.GrpcPort

	localStorage.Nodes = append(localStorage.Nodes, nodeGrpcModel)

	return nil
}

func (u *nodeUsecaseImpl) CreateCluster(ctx context.Context, tokenModel *proto_model.ServerToken) error {
	err := u.startMaster(ctx, tokenModel)
	if err != nil {
		slog.Error("Could not start master node",
			"error", err)

		return err
	}

    time.Sleep(time.Second * 5)

	err = u.startWorker(ctx, tokenModel)
	if err != nil {
		slog.Error("Could not start worker node",
			"error", err)

		return err
	}

	return nil
}

func (u *nodeUsecaseImpl) startMaster(ctx context.Context, tokenModel *proto_model.ServerToken) error {
	masterNode := localStorage.Nodes[0]

	masterNodeGrpcClient, err := config.NewNodeClient(masterNode.IpAddress, masterNode.GrpcPort)
	if err != nil {
		slog.Error("Could not start master node (1)",
			"error", err)

		return err
	}

	_, err = masterNodeGrpcClient.StartMaster(ctx, tokenModel)
	if err != nil {
		slog.Error("Could not start master node (2)",
			"error", err)

		return err
	}

	return nil
}

func (u *nodeUsecaseImpl) startWorker(ctx context.Context, tokenModel *proto_model.ServerToken) error {
    masterNode := localStorage.Nodes[0]
	workerNode := localStorage.Nodes[1]

	workerNodeGrpcClient, err := config.NewNodeClient(workerNode.IpAddress, workerNode.GrpcPort)
	if err != nil {
		slog.Error("Could not start worker node",
			"error", err)

		return err
	}

	masterNodeModel := new(proto_model.MasterNode)
	masterNodeModel.IpAddress = masterNode.IpAddress
	masterNodeModel.GrpcPort = masterNode.GrpcPort
	masterNodeModel.NodeToken = tokenModel.GetToken()

	_, err = workerNodeGrpcClient.StartWorker(ctx, masterNodeModel)
	if err != nil {
		slog.Error("Could not start worker node",
			"error", err)

		return err
	}

	return nil
}
