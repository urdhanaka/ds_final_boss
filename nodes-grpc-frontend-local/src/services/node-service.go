package services

import (
	"context"
	randv2 "math/rand/v2"
	"nodes-grpc-frontend-local/src/config"
	"nodes-grpc-frontend-local/src/model/proto_model"
)

const (
	LOCAL_URL = "localhost:50051"

	// MASTER_IP_ADDRESS = "192.168.122.49"
	// WORKER_IP_ADDRESS = "192.168.122.50"
)

type NodeService struct{}

func NewNodeService() *NodeService {
	return &NodeService{}
}

// func (n *NodeService) CreateCluster(
// 	ctx context.Context,
// 	clusterRequest *virtualization_model.CreateClusterRequest,
// ) (*virtualization_model.VirtCreateInstanceResponse, error) {
// 	res := new(virtualization_model.VirtCreateInstanceResponse)
//
// 	cpu, _ := strconv.ParseInt(clusterRequest.VCPU, 10, 64)
// 	memory, _ := strconv.ParseInt(clusterRequest.Memory, 10, 64)
// 	storage, _ := strconv.ParseInt(clusterRequest.Storage, 10, 64)
// 	nodeSize, _ := strconv.ParseInt(clusterRequest.NodeSize, 10, 64)
//
// 	clusterToken := utils.GenerateRandom(8)
//
// 	nodes, _ := n.nodeRepository.GetAllNodes(ctx)
// 	nodesLength := len(nodes)
// 	if nodesLength == 0 {
// 		return res, errors.New("no node is available")
// 	}
//
// 	// masterNode := nodes[getRandomIndex(nodesLength)]
// 	masterNode := nodes[0]
//
// 	fmt.Println(masterNode)
//
// 	createMasterRes, err := n.createMaster(ctx, masterNode.IpAddress+":50051", &proto_model.CreateMasterRequest{
// 		ClusterName:  clusterRequest.Name,
// 		ClusterToken: clusterToken,
// 		Requirements: &proto_model.CreateNodeRequirements{
// 			NodeName: utils.GenerateRandom(8),
// 			Cpu:      cpu,
// 			Memory:   memory,
// 			Storage:  storage,
// 		},
// 	})
// 	if err != nil {
// 		return res, err
// 	}
//
// 	for i := 1; i < int(nodeSize); i++ {
// 		// currentWorkerNode := nodes[getRandomIndex(nodesLength)]
// 		currentWorkerNode := nodes[1]
// 		_, err = n.createWorker(ctx, currentWorkerNode.IpAddress+":50051", &proto_model.CreateWorkerRequest{
// 			ClusterName:     clusterRequest.Name,
// 			ClusterToken:    clusterToken,
// 			MasterIpAddress: createMasterRes.MasterIpAddress,
// 			Requirements: &proto_model.CreateNodeRequirements{
// 				NodeName: utils.GenerateRandom(8),
// 				Cpu:      cpu,
// 				Memory:   memory,
// 				Storage:  storage,
// 			},
// 		})
// 		if err != nil {
// 			return res, err
// 		}
// 	}
//
// 	res.MasterIpAddress = createMasterRes.MasterIpAddress
// 	res.DashboardToken = createMasterRes.DashboardToken
//
// 	return res, nil
// }

func (n *NodeService) createMaster(
	ctx context.Context,
	nodeIp string,
	masterRequest *proto_model.CreateMasterRequest,
) (*proto_model.CreateMasterResponse, error) {
	// grpcClient, err := config.NewNodeClient(LOCAL_URL)
	grpcClient, err := config.NewNodeClient(nodeIp)
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
	nodeIp string,
	workerRequest *proto_model.CreateWorkerRequest,
) (*proto_model.CreateWorkerResponse, error) {
	// grpcClient, err := config.NewNodeClient(LOCAL_URL)
	grpcClient, err := config.NewNodeClient(nodeIp)
	if err != nil {
		return nil, err
	}

	res, err := grpcClient.CreateWorker(ctx, workerRequest)
	if err != nil {
		return res, err
	}

	return res, nil
}

// func (n *NodeService) GetAllNodes(ctx context.Context) ([]entity.Node, error) {
// 	return n.nodeRepository.GetAllNodes(ctx)
// }

// func (n *NodeService) RegisterNode(
// 	ctx context.Context,
// 	registerNodeRequest *model.Node,
// ) error {
// 	nodeEntity := new(entity.Node)
// 	nodeEntity.ID = uuid.New()
// 	nodeEntity.IpAddress = registerNodeRequest.IpAddress
// 	nodeEntity.Hostname = registerNodeRequest.Hostname
//
// 	err := n.nodeRepository.AddNode(ctx, nodeEntity)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }

func getRandomIndex(maxNum int) int {
	return randv2.IntN(maxNum)
}
