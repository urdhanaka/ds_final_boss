package services

import (
	"context"
	"nodes-grpc-backend-local/model"
	"nodes-grpc-backend-local/repository"
)

type GroupService struct {
	groupRepository *repository.GroupRepository
}

func NewGroupService(groupRepository *repository.GroupRepository) *GroupService {
	return &GroupService{
		groupRepository,
	}
}

func (s *GroupService) GetAllGroups(ctx context.Context) *model.DataPass {
	dataPass := &model.DataPass{
		Data:    nil,
		Message: "",
		Code:    0,
		IsError: false,
		Error:   nil,
	}

    groups, err := s.groupRepository.GetAllGroups(ctx)
    if err != nil {
		dataPass.IsError = true
		dataPass.Message = "could not get the group"
		dataPass.Error = err
    }

    dataPass.Data = groups
	dataPass.Message = "get all groups success"

    return dataPass
}

func (s *GroupService) GetGroupFromName(
	ctx context.Context,
	groupName string,
) *model.DataPass {
	dataPass := &model.DataPass{
		Data:    nil,
		Message: "",
		Code:    200,
		IsError: false,
		Error:   nil,
	}

    group, err := s.groupRepository.GetGroupByName(ctx, groupName)
    if err != nil {
		dataPass.IsError = true
		dataPass.Message = "could not get the group"
		dataPass.Error = err
    }

    dataPass.Data = group
	dataPass.Message = "get group success"

    return dataPass
}
