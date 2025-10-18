package node_service

import (
	"context"
	"errors"
	"log/slog"
	"nodes-grpc-frontend/common/config"
	"nodes-grpc-frontend/common/model/proto"
	"nodes-grpc-frontend/common/model/web"
	"nodes-grpc-frontend/services/db_service"
	"strconv"

	"github.com/google/uuid"
)

type NodeUsecase interface {
	RegisterNode(ctx context.Context, nodeModel *web.RegisterNodeRequest) error
	GetAllNodes(ctx context.Context) ([]web.Node, error)
	CreateCluster(ctx context.Context, createClusterRequest *web.CreateClusterRequest) (string, error)
	// AccessCluster(ctx context.Context, token string) (web.Dashboard, error)
	// CheckNodesHealth(context.Context) ([]*proto_model.NodeStatus, error)
}

type NodeUsecaseImpl struct {
	dbInterface db_service.DatabaseInterface
}

func NewNodeUsecaseImpl(dbConnection db_service.DatabaseInterface) NodeUsecase {
	return &NodeUsecaseImpl{
		dbInterface: dbConnection,
	}
}

func (u *NodeUsecaseImpl) GetAllNodes(ctx context.Context) ([]web.Node, error) {
	nodes, err := u.dbInterface.GetAllNode(ctx)
	if err != nil {
		return []web.Node{}, err
	}

	return nodes, nil
}

// func (u *nodeUsecaseImpl) CheckNodesHealth(ctx context.Context) ([]*proto_model.NodeStatus, error) {
// 	return localStorage.Nodes, nil

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
// }

func (u *NodeUsecaseImpl) RegisterNode(ctx context.Context, nodeModel *web.RegisterNodeRequest) error {
	dbNode := web.Node{
		UUID:     uuid.New().String(),
		NodeIP:   nodeModel.NodeIP,
		GrpcPort: nodeModel.GrpcPort,
	}

	// check if record exists
	nodeExists, err := u.dbInterface.GetNode(ctx, dbNode)
	if err != nil {
		return err
	}

	if nodeExists.NodeIP != "" {
		return nil
	}

	err = u.dbInterface.StoreNode(ctx, dbNode)
	if err != nil {
		return err
	}

	return nil
}

func (u *NodeUsecaseImpl) CreateCluster(ctx context.Context, createClusterRequest *web.CreateClusterRequest) (string, error) {
	masterNode, err := u.startMaster(ctx, createClusterRequest)
	if err != nil {
		slog.Error("Could not start master node",
			"error", err)

		return "", err
	}

	err = u.startWorker(ctx, masterNode, createClusterRequest)
	if err != nil {
		slog.Error("Could not start worker node",
			"error", err)

		return "", err
	}

	return "", nil
}

func (u *NodeUsecaseImpl) startMaster(
	ctx context.Context,
	createClusterRequest *web.CreateClusterRequest,
) (string, error) {
	masterNode, err := u.pickNode(ctx)
	if err != nil {
		return "", err
	}

	masterNodeGrpcClient, err := config.NewNodeClient(masterNode.NodeIP, masterNode.GrpcPort)
	if err != nil {
		slog.Error("Could not create grpc client for master node",
			"error", err.Error(),
		)

		return "", err
	}

	// check node status
	// nodeStatus, err := masterNodeGrpcClient.Heartbeat(ctx, new(proto_model.Empty))
	// if err != nil {
	// 	slog.Error("Could not check master node status",
	// 		"error", err.Error(),
	// 	)
	//
	// 	return new(proto_model.NodeStatus), err
	// }
	//
	// if nodeStatus.GetStatus() == proto_model.Status_STATUS_UNAVAILABLE {
	// 	slog.Info("Master node is not available")
	//
	// 	return new(proto_model.NodeStatus), errors.New("Master node not available")
	// }

	cpuInt64, _ := strconv.ParseInt(createClusterRequest.Vcpu, 10, 64)
	memoryInt64, _ := strconv.ParseInt(createClusterRequest.Memory, 10, 64)
	storageInt64, _ := strconv.ParseInt(createClusterRequest.Storage, 10, 64)

	createMasterResponse, err := masterNodeGrpcClient.CreateMaster(ctx, &proto.CreateMasterRequest{
		Token: createClusterRequest.Token,
		Requirements: &proto.CreateNodeRequirements{
			Cpu:     cpuInt64,
			Memory:  memoryInt64,
			Storage: storageInt64,
		},
	})
	if err != nil {
		slog.Error("Could not start master node",
			"error", err.Error(),
		)

		return "", err
	}

	// save the master
	u.dbInterface.StoreNode(ctx, web.Node{})

	// ip address of the master
	masterIpAddress := createMasterResponse.GetIpAddress()

	return masterIpAddress, nil
}

func (u *NodeUsecaseImpl) startWorker(
	ctx context.Context,
	masterIpAddress string,
	createClusterRequest *web.CreateClusterRequest,
) error {
	workerNode, err := u.pickNode(ctx)
	if err != nil {
		return err
	}

	workerNodeGrpcClient, err := config.NewNodeClient(workerNode.NodeIP, workerNode.GrpcPort)
	if err != nil {
		slog.Error("Could not create grpc client for worker node",
			"error", err.Error(),
		)

		return err
	}

	cpuInt64, _ := strconv.ParseInt(createClusterRequest.Vcpu, 10, 64)
	memoryInt64, _ := strconv.ParseInt(createClusterRequest.Memory, 10, 64)
	storageInt64, _ := strconv.ParseInt(createClusterRequest.Storage, 10, 64)

	createWorkerRequest := new(proto.CreateWorkerRequest)
	createWorkerRequest.IpAddress = masterIpAddress
	createWorkerRequest.Token = createClusterRequest.Token
	createWorkerRequest.Requirements = &proto.CreateNodeRequirements{
		Cpu:     cpuInt64,
		Memory:  memoryInt64,
		Storage: storageInt64,
	}

	// nodeStatus, err := workerNodeGrpcClient.Heartbeat(ctx, new(proto_model.Empty))
	// if err != nil {
	// 	slog.Error("Could not check worker node status",
	// 		"error", err.Error(),
	// 	)
	//
	// 	return new(proto_model.NodeStatus), err
	// }

	// if nodeStatus.GetStatus() == proto_model.Status_STATUS_UNAVAILABLE {
	// 	slog.Info("Worker node is not available")
	//
	// 	return new(proto_model.NodeStatus), errors.New("Worker node not available")
	// }

	_, err = workerNodeGrpcClient.CreateWorker(ctx, createWorkerRequest)
	if err != nil {
		slog.Error("Could not start worker node",
			"error", err,
		)

		return err
	}

	return nil
}

// TODO: remove on prod, used only on development phase
// only use for testing purpose
func (u *NodeUsecaseImpl) pickNode(ctx context.Context) (web.Node, error) {
	allNodes, err := u.dbInterface.GetAllNode(ctx)
	if err != nil {
		return web.Node{}, err
	}

	if len(allNodes) == 0 {
		return web.Node{}, errors.New("No nodes available")
	}

	return allNodes[0], nil
}
