package services

import (
	"context"
	"errors"
	"log/slog"
	"nodes-grpc-backend-local/config"
	"nodes-grpc-backend-local/entity"
	"nodes-grpc-backend-local/model"
	"nodes-grpc-backend-local/model/proto_model"
	"nodes-grpc-backend-local/repository"
	"time"

	"github.com/google/uuid"
)

type ClusterService struct {
	clusterRepository     *repository.ClusterRepository
	nodeRepository        *repository.NodeRepository
	clusterNodeRepository *repository.ClusterNodeRepository
}

func NewClusterService(
	clusterRepository *repository.ClusterRepository,
	nodeRepository *repository.NodeRepository,
	clusterNodeRepository *repository.ClusterNodeRepository,
) *ClusterService {
	return &ClusterService{
		clusterRepository,
		nodeRepository,
		clusterNodeRepository,
	}
}

func (s *ClusterService) AddCluster(
	ctx context.Context,
	cluster *model.AddCluster,
) error {
	createClusterContext, cancel := context.WithCancel(ctx)
	defer cancel()

	masterResponse := new(proto_model.CreateMasterResponse)

	// to track the picked nodes
	// for atomic process
	pickedNodes := make(map[entity.Node][]string, cluster.NodeSize)

	thisGroupNodes, err := s.nodeRepository.GetNodesFromGroup(createClusterContext, cluster.GroupId)
	if err != nil {
		return err
	}

	if len(thisGroupNodes) == 0 {
		return errors.New("no node for this group")
	}

	clusterToken := generateRandom(8)

	for i := 1; i <= cluster.NodeSize; i++ {
		// selected node from all the nodes
		nodeIndex := getRandomIndex(cluster.NodeSize)
		pickedNode := thisGroupNodes[nodeIndex]

		// create the grpc client
		grpcClient, err := config.NewNodeClient(pickedNode.IpAddress)
		if err != nil {
			return err
		}

		// generate instance name
		instanceName := generateRandom(8)

		// assume that the first index is always the control plane
		if i == 1 {
			res, err := grpcClient.CreateMaster(
				createClusterContext,
				&proto_model.CreateMasterRequest{
					ClusterName:  cluster.ClusterName,
					ClusterToken: clusterToken,
					Requirements: &proto_model.CreateNodeRequirements{
						NodeName: instanceName,
						Cpu:      int32(cluster.Vcpu),
						Memory:   int32(cluster.Memory),
						Storage:  int32(cluster.Storage),
					},
				},
			)
			if err != nil {
				deleteInstance(grpcClient, instanceName)
				clusterCleanup(&pickedNodes)
				return err
			}
			// status error
			if res.CreationStatus.Success == false {
				deleteInstance(grpcClient, instanceName)
				clusterCleanup(&pickedNodes)
				return errors.New(res.CreationStatus.Message)
			}

			// append to the picked nodes
			// for deleting the vm
			pickedNodes[pickedNode] = append(pickedNodes[pickedNode], instanceName)

			masterResponse.MasterIpAddress = res.GetMasterIpAddress()
		} else {
			res, err := grpcClient.CreateWorker(
				createClusterContext,
				&proto_model.CreateWorkerRequest{
					ClusterName:     cluster.ClusterName,
					ClusterToken:    clusterToken,
					MasterIpAddress: masterResponse.GetMasterIpAddress(),
					Requirements: &proto_model.CreateNodeRequirements{
						NodeName: instanceName,
						Cpu:      int32(cluster.Vcpu),
						Memory:   int32(cluster.Memory),
						Storage:  int32(cluster.Storage),
					},
				},
			)
			// error
			if err != nil {
				deleteInstance(grpcClient, instanceName)
				clusterCleanup(&pickedNodes)
				return err
			}
			// status error
			if res.CreationStatus.Success == false {
				deleteInstance(grpcClient, instanceName)
				clusterCleanup(&pickedNodes)
				return errors.New(res.CreationStatus.Message)
			}

			// append to the picked nodes
			// for deleting the vm
			pickedNodes[pickedNode] = append(pickedNodes[pickedNode], instanceName)
		}
	}

	clusterId := uuid.New()
	thisCluster := &entity.Cluster{
		ClusterID: clusterId,
		Name:      cluster.ClusterName,
		UserID:    cluster.UserId,
		GroupID:   cluster.GroupId,
		CreatedAt: time.Now(),
	}

	s.clusterRepository.AddCluster(ctx, thisCluster)

	for node, instanceNames := range pickedNodes {
		for _, instanceName := range instanceNames {
			s.clusterNodeRepository.AddEntry(
				createClusterContext,
				thisCluster,
				&node,
				instanceName,
			)
		}
	}

	return nil
}

func (s *ClusterService) DeleteCluster(
	ctx context.Context,
	cluster *entity.Cluster,
) error {
	return s.clusterRepository.DeleteClusterById(ctx, cluster)
}

func clusterCleanup(
	nodes *map[entity.Node][]string,
) {
	if len(*nodes) == 0 {
		return
	}

	for node, instanceNames := range *nodes {
		if len(instanceNames) == 0 {
			continue
		}

		grpcClient, err := config.NewNodeClient(node.IpAddress)
		if err != nil {
			slog.Error("could not create grpc client",
				"error", err,
			)
			return
		}

		for _, instanceName := range instanceNames {
			deleteInstance(grpcClient, instanceName)
		}
	}
}

func deleteInstance(
	client proto_model.NodeServiceClient,
	instanceName string,
) error {
	_, err := client.DeleteInstance(
		context.Background(),
		&proto_model.DeleteInstanceRequest{
			InstanceName: instanceName,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
