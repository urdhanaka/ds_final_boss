package services

import (
	"context"
	"nodes-grpc-be/entities"
	"nodes-grpc-be/models"
	"nodes-grpc-be/repositories"
)

type ClusterService struct {
	clusterNodeRepository *repositories.ClusterNodeRepository
	clusterRepository     *repositories.ClusterRepository
	// userRepository *repositories.UserRepository
	// nodeRepository        *repositories.NodeRepository
	groupRepository *repositories.GroupRepository
}

func NewClusterService(
	clusterNodeRepository *repositories.ClusterNodeRepository,
	clusterRepository *repositories.ClusterRepository,
	// userRepository *repositories.UserRepository,
	nodeRepository *repositories.NodeRepository,
	groupRepository *repositories.GroupRepository,
) *ClusterService {
	return &ClusterService{
		clusterNodeRepository: clusterNodeRepository,
		clusterRepository:     clusterRepository,
		// userRepository: userRepository,
		// nodeRepository:        nodeRepository,
		groupRepository: groupRepository,
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

func (s *ClusterService) GetUserCluster(
	ctx context.Context,
	user *entities.User,
) ([]*models.GetUserClusters, error) {
	clusters, err := s.clusterRepository.GetClusterFromUserId(ctx, user)
	if err != nil {
		return nil, err
	}

	getUserClusters := []*models.GetUserClusters{}
	for _, cluster := range clusters {
		thisCluster := &models.GetUserClusters{
			ClusterId:     cluster.ClusterID,
			ClusterName:   cluster.ClusterName,
			ClusterStatus: cluster.ClusterStatus,
		}

		getUserClusters = append(getUserClusters, thisCluster)
	}

	return getUserClusters, nil
}

func (s *ClusterService) GetClusterById(
	ctx context.Context,
	cluster *entities.Cluster,
) (*models.GetClusterDetails, error) {
	getCluster, err := s.clusterRepository.GetClusterFromClusterId(ctx, cluster)
	if err != nil {
		return nil, err
	}

	clusterDetails := &models.GetClusterDetails{
		ClusterId:     getCluster.ClusterID,
		ClusterName:   getCluster.ClusterName,
		UserId:        getCluster.UserID,
		ClusterStatus: getCluster.ClusterStatus,
		IpAddress:     getCluster.IpAddress,
		AccessToken:   getCluster.AccessToken,
		CreatedAt:     getCluster.CreatedAt,
	}

	return clusterDetails, nil
}
