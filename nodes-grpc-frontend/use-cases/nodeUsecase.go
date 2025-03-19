package usecases

import (
	"context"
	"errors"
	"log/slog"
	"nodes-grpc-frontend/common/config"
	proto_model "nodes-grpc-frontend/common/model/proto-model"
	web_model "nodes-grpc-frontend/common/model/web-model"
	"nodes-grpc-frontend/utils"
	"time"
)

type NodeUsecase interface {
	GetAllNodes(context.Context) *proto_model.NodeList
	RegisterNode(_ context.Context, nodeModel *web_model.RegisterNodeRequest) error
	CreateCluster(ctx context.Context, tokenModel *proto_model.ServerToken) error
	CheckNodesHealth(context.Context) ([]*proto_model.NodeStatus, error)
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

func (u *nodeUsecaseImpl) CheckNodesHealth(ctx context.Context) ([]*proto_model.NodeStatus, error) {
	return localStorage.Nodes, nil

	// masterNode := localStorage.Nodes[0]
	// workerNode := localStorage.Nodes[1]
	//
	// nodeStatus := make([]*proto_model.NodeStatus, 0)
	//
	// masterNodeGrpcClient, err := config.NewNodeClient(masterNode.IpAddress, masterNode.GrpcPort)
	// if err != nil {
	// 	slog.Error("Could not create grpc client for master node",
	// 		"error", err.Error(),
	// 	)
	//
	// 	return nodeStatus, err
	// }
	//
	// masterNodeStatus, err := masterNodeGrpcClient.Heartbeat(ctx, new(proto_model.Empty))
	// if err != nil {
	// 	slog.Error("Could not check master node",
	// 		"error", err.Error(),
	// 	)
	//
	// 	return nodeStatus, err
	// }
	//
	// nodeStatus = append(nodeStatus, masterNodeStatus)
	//
	// workerNodeGrpcClient, err := config.NewNodeClient(workerNode.IpAddress, workerNode.GrpcPort)
	// if err != nil {
	// 	slog.Error("Could not create grpc client for worker node",
	// 		"error", err.Error(),
	// 	)
	//
	// 	return nodeStatus, err
	// }
	//
	// workerNodeStatus, err := workerNodeGrpcClient.Heartbeat(ctx, new(proto_model.Empty))
	// if err != nil {
	// 	slog.Error("Could not check worker node",
	// 		"error", err.Error(),
	// 	)
	//
	// 	return nodeStatus, err
	// }
	//
	// nodeStatus = append(nodeStatus, workerNodeStatus)
	//
	// return nodeStatus, nil
}

func (u *nodeUsecaseImpl) RegisterNode(_ context.Context, nodeModel *web_model.RegisterNodeRequest) error {
	// if node exists in localStorage, return
	for _, node := range localStorage.Nodes {
		if node.IpAddress == nodeModel.IpAddress {
			return nil
		}
	}

	nodeGrpcModel := new(proto_model.NodeStatus)

	nodeGrpcModel.Hostname = nodeModel.Hostname
	nodeGrpcModel.IpAddress = nodeModel.IpAddress
	nodeGrpcModel.GrpcPort = nodeModel.GrpcPort

	localStorage.Nodes = append(localStorage.Nodes, nodeGrpcModel)

	return nil
}

func (u *nodeUsecaseImpl) CreateCluster(ctx context.Context, tokenModel *proto_model.ServerToken) error {
	masterNode, err := u.startMaster(ctx, tokenModel)
	if err != nil {
		slog.Error("Could not start master node",
			"error", err)

		return err
	}

	time.Sleep(time.Second * 5)

	workerNode, err := u.startWorker(ctx, masterNode, tokenModel)
	if err != nil {
		slog.Error("Could not start worker node",
			"error", err)

		return err
	}

	// TODO: Send kubeconfig

	// TODO: remove the nodes from pool
	utils.DeleteFromPool(localStorage, masterNode)
	utils.DeleteFromPool(localStorage, workerNode)

	return nil
}

func (u *nodeUsecaseImpl) startMaster(ctx context.Context, tokenModel *proto_model.ServerToken) (*proto_model.NodeStatus, error) {
	masterNode, err := utils.GetRandomNode(localStorage)
	if err != nil {
		return new(proto_model.NodeStatus), err
	}

	masterNodeGrpcClient, err := config.NewNodeClient(masterNode.IpAddress, masterNode.GrpcPort)
	if err != nil {
		slog.Error("Could not create grpc client for master node",
			"error", err.Error(),
		)

		return new(proto_model.NodeStatus), err
	}

	// check node status
	nodeStatus, err := masterNodeGrpcClient.Heartbeat(ctx, new(proto_model.Empty))
	if err != nil {
		slog.Error("Could not check master node status",
			"error", err.Error(),
		)

		return new(proto_model.NodeStatus), err
	}

	if nodeStatus.GetStatus() == proto_model.Status_STATUS_UNAVAILABLE {
		slog.Info("Master node is not available")

		return new(proto_model.NodeStatus), errors.New("Master node not available")
	}

	_, err = masterNodeGrpcClient.StartMaster(ctx, tokenModel)
	if err != nil {
		slog.Error("Could not start master node",
			"error", err.Error(),
		)

		return new(proto_model.NodeStatus), err
	}

	return masterNode, nil
}

func (u *nodeUsecaseImpl) startWorker(ctx context.Context,
	masterNode *proto_model.NodeStatus,
	tokenModel *proto_model.ServerToken,
) (*proto_model.NodeStatus, error) {
	workerNode, err := utils.GetRandomNode(localStorage)
	if err != nil {
		return new(proto_model.NodeStatus), err
	}

	workerNodeGrpcClient, err := config.NewNodeClient(workerNode.IpAddress, workerNode.GrpcPort)
	if err != nil {
		slog.Error("Could not create grpc client for worker node",
			"error", err.Error(),
		)

		return new(proto_model.NodeStatus), err
	}

	masterNodeModel := new(proto_model.MasterNode)
	masterNodeModel.IpAddress = masterNode.IpAddress
	masterNodeModel.GrpcPort = masterNode.GrpcPort
	masterNodeModel.NodeToken = tokenModel.GetToken()

	nodeStatus, err := workerNodeGrpcClient.Heartbeat(ctx, new(proto_model.Empty))
	if err != nil {
		slog.Error("Could not check worker node status",
			"error", err.Error(),
		)

		return new(proto_model.NodeStatus), err
	}

	if nodeStatus.GetStatus() == proto_model.Status_STATUS_UNAVAILABLE {
		slog.Info("Worker node is not available")

		return new(proto_model.NodeStatus), errors.New("Worker node not available")
	}

	_, err = workerNodeGrpcClient.StartWorker(ctx, masterNodeModel)
	if err != nil {
		slog.Error("Could not start worker node",
			"error", err,
		)

		return new(proto_model.NodeStatus), err
	}

	return workerNode, nil
}
