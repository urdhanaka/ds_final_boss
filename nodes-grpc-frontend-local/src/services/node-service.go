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
) error {
	cpu, _ := strconv.ParseInt(clusterRequest.VCPU, 10, 64)
	memory, _ := strconv.ParseInt(clusterRequest.Memory, 10, 64)
	storage, _ := strconv.ParseInt(clusterRequest.Storage, 10, 64)

	err := n.createMaster(ctx, &proto_model.CreateInstanceRequest{
		IsMaster:  true,
		Token:     "12345",
		IpAddress: "",
		Cpu:       cpu,
		Memory:    memory,
		Storage:   storage,
	})
	if err != nil {
		return err
	}

	err = n.createWorker(ctx, &proto_model.CreateInstanceRequest{
		IsMaster:  false,
		Token:     "12345",
		IpAddress: "",
		Cpu:       cpu,
		Memory:    memory,
		Storage:   storage,
	})
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeService) createMaster(
	ctx context.Context,
	instanceRequest *proto_model.CreateInstanceRequest,
) error {
	grpcClient, err := config.NewNodeClient()
	if err != nil {
		return err
	}

	_, err = grpcClient.CreateInstance(ctx, instanceRequest)
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeService) createWorker(
	ctx context.Context,
	instanceRequest *proto_model.CreateInstanceRequest,
) error {
	grpcClient, err := config.NewNodeClient()
	if err != nil {
		return err
	}

	_, err = grpcClient.CreateInstance(ctx, instanceRequest)
	if err != nil {
		return err
	}

	return nil
}
