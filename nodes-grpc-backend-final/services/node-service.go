package services

import (
	"context"
	"log/slog"
	"nodes-grpc-be/config"
	"nodes-grpc-be/entities"
	"nodes-grpc-be/models"
	"nodes-grpc-be/models/proto_model"
	"nodes-grpc-be/repositories"

	"github.com/google/uuid"
)

const (
	GRPC_CLIENT_CONNECT_RETRIES = 3
)

type NodeService struct {
	nodeRepository  *repositories.NodeRepository
	groupRepository *repositories.GroupRepository
}

func NewNodeService(
	nodeRepository *repositories.NodeRepository,
	groupRepository *repositories.GroupRepository,
) *NodeService {
	return &NodeService{
		nodeRepository,
		groupRepository,
	}
}

func (s *NodeService) GetGroupNodes(
	ctx context.Context,
	groupId int,
) ([]*models.GetGroupNodes, error) {
	nodes, err := s.nodeRepository.GetNodesFromGroup(ctx, groupId)
	if err != nil {
		return nil, err
	}

	getGroupNodes := []*models.GetGroupNodes{}
	for _, node := range nodes {
		thisNode := &models.GetGroupNodes{
			NodeId:    node.NodeID,
			Hostname:  node.Hostname,
			IpAddress: node.IpAddress,
			VCpu:      node.VCpu,
			Memory:    node.Memory,
			Storage:   node.Storage,
			GroupId:   node.GroupId,
		}

		getGroupNodes = append(getGroupNodes, thisNode)
	}

	return getGroupNodes, nil
}

func (s *NodeService) AddNode(
	ctx context.Context,
	addNode *models.AddNode,
) error {
	group, err := s.groupRepository.GetGroupByName(ctx, addNode.LabName)
	if err != nil {
		return err
	}

	nodeEntity := &entities.Node{
		NodeID:    uuid.New(),
		Hostname:  addNode.Hostname,
		IpAddress: addNode.IpAddress,
		GroupId:   group.GroupId,
		VCpu:      addNode.VCpu,
		Memory:    addNode.Memory,
		Storage:   addNode.Storage,
	}

	err = s.nodeRepository.AddNode(ctx, nodeEntity)
	if err != nil {
		return err
	}

	return nil
}

func (s *NodeService) UpdateNode(
	ctx context.Context,
	groupId int,
) error {
	nodes, err := s.GetGroupNodes(ctx, groupId)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		grpcClient, err := config.NewNodeClient(node.IpAddress)
		if err != nil {
			slog.Error("could not connect to node",
				"ip address", node.IpAddress,
				"hostname", node.Hostname,
			)
			continue
		}

		nodeResources, err := grpcClient.NodeStatus(ctx, &proto_model.NodeStatusRequest{})
		if err != nil {
			slog.Error("could not execute NodeStatus procedure",
				"ip address", node.IpAddress,
				"hostname", node.Hostname,
			)
			continue
		}

		thisNodeEntity := &entities.Node{
			NodeID:  node.NodeId,
			Memory:  int(nodeResources.MemoryAvailable),
			VCpu:    int(nodeResources.FreeVcpu),
			Storage: int(nodeResources.StorageAvailable),
		}

		err = s.nodeRepository.UpdateNodeResourcesByNodeId(ctx, thisNodeEntity)
		if err != nil {
			slog.Error("could not update node resources info",
				"ip address", node.IpAddress,
				"hostname", node.Hostname,
			)
			continue
		}
	}

	return nil
}
