package services

import (
	"context"
	"nodes-grpc-backend-local/entity"
	"nodes-grpc-backend-local/model"
	"nodes-grpc-backend-local/repository"

	"github.com/google/uuid"
)

type NodeService struct {
	nodeRepository  *repository.NodeRepository
	groupRepository *repository.GroupRepository
}

func NewNodeService(
	nodeRepository *repository.NodeRepository,
	groupRepository *repository.GroupRepository,
) *NodeService {
	return &NodeService{
		nodeRepository,
		groupRepository,
	}
}

func (s *NodeService) GetAllNodes(ctx context.Context) *model.DataPass {
	dataPass := &model.DataPass{
		Data:    nil,
		Message: "",
		Code:    0,
		IsError: false,
		Error:   nil,
	}

	nodes, err := s.nodeRepository.GetAll(ctx)
	if err != nil {
		dataPass.IsError = true
		dataPass.Message = "could not get all nodes"
		dataPass.Error = err
	}

	dataPass.Data = nodes
	dataPass.Message = "get all nodes success"

	return dataPass
}

func (s *NodeService) AddNode(ctx context.Context, node *model.AddNode) error {
	group, err := s.groupRepository.GetGroupByName(ctx, node.LabName)
	if err != nil {
		return err
	}

	nodeEntity := &entity.Node{
		NodeID:    uuid.New(),
		Hostname:  node.Hostname,
		IpAddress: node.IpAddress,
		Cpu:       node.Cpu,
		Ram:       node.Ram,
		Storage:   node.Storage,
		GroupId:   group.GroupId,
	}

	err = s.nodeRepository.AddNode(ctx, nodeEntity)
	if err != nil {
		return err
	}

	return nil
}
