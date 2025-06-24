package services

import (
	"nodes-grpc-be/repositories"
)

type ClusterService struct {
	clusterNodeRepository *repositories.ClusterNodeRepository
	clusterRepository     *repositories.ClusterRepository
	nodeRepository        *repositories.NodeRepository
	groupRepository       *repositories.GroupRepository
}

func NewClusterService(
	clusterNodeRepository *repositories.ClusterNodeRepository,
	clusterRepository *repositories.ClusterRepository,
	nodeRepository *repositories.NodeRepository,
	groupRepository *repositories.GroupRepository,
) *ClusterService {
	return &ClusterService{
		clusterNodeRepository: clusterNodeRepository,
		clusterRepository:     clusterRepository,
		nodeRepository:        nodeRepository,
		groupRepository:       groupRepository,
	}
}

// func (s *ClusterService) CreateCluster(
// 	ctx context.Context,
// 	job *entities.Job,
// ) error {
// 	createClusterContext, cancel := context.WithCancel(ctx)
// 	defer cancel()
//
// 	thisGroup, err := s.groupRepository.GetGroupById(ctx, cluster.GroupId)
// 	if err != nil {
// 		return err
// 	}
// }
