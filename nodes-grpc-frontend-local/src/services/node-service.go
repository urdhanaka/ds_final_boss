package services

import (
	"context"
	"nodes-grpc-frontend-local/src/config"
	"nodes-grpc-frontend-local/src/model/proto_model"
	"nodes-grpc-frontend-local/src/model/virtualization_model"
	"nodes-grpc-frontend-local/src/utils"
	"strconv"
)

const (
	MASTER_IP_ADDRESS = "192.168.122.49"
	WORKER_IP_ADDRESS = "192.168.122.50"
)

type NodeService struct {
	// databaseService *DatabaseService
}

func NewNodeService() *NodeService {
	return &NodeService{
		// databaseService: databaseService,
	}
}

func (n *NodeService) CreateCluster(
	ctx context.Context,
	clusterRequest *virtualization_model.CreateClusterRequest,
) (*virtualization_model.VirtCreateInstanceResponse, error) {
	res := new(virtualization_model.VirtCreateInstanceResponse)

	cpu, _ := strconv.ParseInt(clusterRequest.VCPU, 10, 64)
	memory, _ := strconv.ParseInt(clusterRequest.Memory, 10, 64)
	storage, _ := strconv.ParseInt(clusterRequest.Storage, 10, 64)

	clusterToken := utils.GenerateRandom(8)

	createMasterRes, err := n.createMaster(ctx, &proto_model.CreateMasterRequest{
		ClusterToken: clusterToken,
		Requirements: &proto_model.CreateNodeRequirements{
			Cpu:     cpu,
			Memory:  memory,
			Storage: storage,
		},
	})
	if err != nil {
		return res, err
	}

	_, err = n.createWorker(ctx, &proto_model.CreateWorkerRequest{
		ClusterToken: clusterToken,
		IpAddress:    MASTER_IP_ADDRESS,
		Requirements: &proto_model.CreateNodeRequirements{
			Cpu:     cpu,
			Memory:  memory,
			Storage: storage,
		},
	})
	if err != nil {
		return res, err
	}

	res.DashboardToken = createMasterRes.DashboardToken

	return res, nil
}

func (n *NodeService) createMaster(
	ctx context.Context,
	masterRequest *proto_model.CreateMasterRequest,
) (*proto_model.CreateMasterResponse, error) {
	grpcClient, err := config.NewNodeClient()
	if err != nil {
		return nil, err
	}

	res, err := grpcClient.CreateMaster(ctx, masterRequest)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (n *NodeService) createWorker(
	ctx context.Context,
	workerRequest *proto_model.CreateWorkerRequest,
) (*proto_model.CreateWorkerResponse, error) {
	grpcClient, err := config.NewNodeClient()
	if err != nil {
		return nil, err
	}

	res, err := grpcClient.CreateWorker(ctx, workerRequest)
	if err != nil {
		return res, err
	}

	return res, nil
}
