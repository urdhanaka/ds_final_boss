package services

import (
	"context"
	"nodes-grpc-be/models"
	"nodes-grpc-be/repositories"
)

type NodeService struct {
	nodeRepository *repositories.NodeRepository
}

func NewNodeService(
	nodeRepository *repositories.NodeRepository,
) *NodeService {
	return &NodeService{
		nodeRepository,
	}
}

func (s *NodeService) GetGroupCluster(
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
			Ram:       node.Ram,
			Storage:   node.Storage,
			GroupId:   node.GroupId,
		}

		getGroupNodes = append(getGroupNodes, thisNode)
	}

	return getGroupNodes, nil
}
