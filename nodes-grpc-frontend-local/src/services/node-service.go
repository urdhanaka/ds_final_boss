package services

import (
	"context"
	"nodes-grpc-frontend-local/src/config"
	"nodes-grpc-frontend-local/src/model/proto_model"
	"nodes-grpc-frontend-local/src/model/virtualization_model"
	"strconv"
)

const (
	MASTER_IP_ADDRESS = "192.168.122.49"
	WORKER_IP_ADDRESS = "192.168.122.50"
)

type NodeService struct{}

func NewNodeService() *NodeService {
	return &NodeService{}
}

func (n *NodeService) CreateCluster(
	ctx context.Context,
	clusterRequest *virtualization_model.CreateClusterRequest,
) error {
	cpu, _ := strconv.ParseInt(clusterRequest.VCPU, 10, 64)
	memory, _ := strconv.ParseInt(clusterRequest.Memory, 10, 64)
	storage, _ := strconv.ParseInt(clusterRequest.Storage, 10, 64)

	err := n.createMaster(ctx, &proto_model.CreateMasterRequest{
		Requirements: &proto_model.CreateNodeRequirements{
			Cpu:     cpu,
			Memory:  memory,
			Storage: storage,
		},
	})
	if err != nil {
		return err
	}

	err = n.createWorker(ctx, &proto_model.CreateWorkerRequest{
		Requirements: &proto_model.CreateNodeRequirements{
			Cpu:     cpu,
			Memory:  memory,
			Storage: storage,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeService) createMaster(
	ctx context.Context,
	masterRequest *proto_model.CreateMasterRequest,
) error {
	grpcClient, err := config.NewNodeClient()
	if err != nil {
		return err
	}

	_, err = grpcClient.CreateMaster(ctx, masterRequest)
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeService) createWorker(
	ctx context.Context,
	workerRequest *proto_model.CreateWorkerRequest,
) error {
	grpcClient, err := config.NewNodeClient()
	if err != nil {
		return err
	}

	_, err = grpcClient.CreateWorker(ctx, workerRequest)
	if err != nil {
		return err
	}

	return nil
}
