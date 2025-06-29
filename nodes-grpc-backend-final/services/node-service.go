package services

import (
	"context"
	"nodes-grpc-be/entities"
	"nodes-grpc-be/models"
	"nodes-grpc-be/repositories"

	"github.com/google/uuid"
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
