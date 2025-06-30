package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"nodes-grpc-be/config"
	"nodes-grpc-be/entities"
	"nodes-grpc-be/models"
	"nodes-grpc-be/models/proto_model"
	"nodes-grpc-be/repositories"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ClusterService struct {
	clusterNodeRepository *repositories.ClusterNodeRepository
	clusterRepository     *repositories.ClusterRepository
	userRepository        *repositories.UserRepository
	nodeRepository        *repositories.NodeRepository
	groupRepository       *repositories.GroupRepository
}

func NewClusterService(
	clusterNodeRepository *repositories.ClusterNodeRepository,
	clusterRepository *repositories.ClusterRepository,
	userRepository *repositories.UserRepository,
	nodeRepository *repositories.NodeRepository,
	groupRepository *repositories.GroupRepository,
) *ClusterService {
	return &ClusterService{
		clusterNodeRepository: clusterNodeRepository,
		clusterRepository:     clusterRepository,
		userRepository:        userRepository,
		nodeRepository:        nodeRepository,
		groupRepository:       groupRepository,
	}
}

func (s *ClusterService) CreateCluster(
	ctx context.Context,
	job *entities.Job,
) (*models.CreateClusterResponse, error) {
	createClusterContext, cancel := context.WithCancel(ctx)
	defer cancel()

	createClusterResponse := new(models.CreateClusterResponse)
	addClusterModel := new(models.AddCluster)

	jsonBytes, _ := json.Marshal(job.Payload)
	_ = json.Unmarshal(jsonBytes, addClusterModel)

	clusterEntity := &entities.Cluster{
		ClusterID:     uuid.New(),
		ClusterName:   addClusterModel.ClusterName,
		UserID:        addClusterModel.UserId,
		GroupID:       addClusterModel.GroupId,
		ClusterStatus: string(entities.JOB_STATUS_QUEUED),
		CreatedAt:     time.Now(),
	}

	err := s.clusterRepository.AddCluster(createClusterContext, clusterEntity)
	if err != nil {
		return createClusterResponse, err
	}

	// start the cluster creation here
	// keep track of the nodes and the domain name
	nodeAndDomainName := []string{}

	// keep track of master node creation
	isMasterNode := true
	currentDomainIndex := 0

	// cluster token to create the cluster
	clusterToken := createRandomString(8)

	for currentDomainIndex < addClusterModel.NodeSize {
		nodesCopy := addClusterModel.Nodes

		pickedIndex := currentDomainIndex % addClusterModel.NodeSize % len(nodesCopy)
		nodeIdentity := nodesCopy[pickedIndex]

		ipAddress := strings.Split(nodeIdentity, ",")[1]

		grpcClient, err := config.NewNodeClient(ipAddress)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// check if node can be used
		nodeStatus, _ := grpcClient.NodeStatus(
			createClusterContext,
			&proto_model.NodeStatusRequest{},
		)

		// check resources
		if addClusterModel.Vcpu > int(nodeStatus.FreeVcpu) ||
			addClusterModel.Storage > int(nodeStatus.StorageAvailable) ||
			addClusterModel.Memory > int(nodeStatus.MemoryAvailable) {
			slog.Info("worker node has less resources than needed",
				"resources", "vcpu",
				"node", nodeIdentity,
			)
			continue
		}

		if isMasterNode {
			domainName := createRandomString(8)

			res, err := grpcClient.CreateMaster(
				createClusterContext,
				&proto_model.CreateMasterRequest{
					ClusterName:  addClusterModel.ClusterName,
					ClusterToken: clusterToken,
					Requirements: &proto_model.CreateNodeRequirements{
						NodeName: domainName,
						Vcpu:     int32(addClusterModel.Vcpu),
						Memory:   int32(addClusterModel.Memory),
						Storage:  int32(addClusterModel.Storage),
					},
				},
			)
			if err != nil {
				slog.Error("error creating master",
					"error", err,
				)
			}

			// master node created
			isMasterNode = false
			createClusterResponse.DashboardToken = res.DashboardToken
			createClusterResponse.MasterIpAddress = res.MasterIpAddress

			// add the masterNode to the array
			nodeAndDomainName = append(nodeAndDomainName, fmt.Sprintf("%s,%s", ipAddress, domainName))
		} else {
			domainName := createRandomString(8)

			_, err := grpcClient.CreateWorker(
				createClusterContext,
				&proto_model.CreateWorkerRequest{
					ClusterName:     addClusterModel.ClusterName,
					ClusterToken:    clusterToken,
					MasterIpAddress: createClusterResponse.MasterIpAddress,
					Requirements: &proto_model.CreateNodeRequirements{
						NodeName: domainName,
						Vcpu:     int32(addClusterModel.Vcpu),
						Memory:   int32(addClusterModel.Memory),
						Storage:  int32(addClusterModel.Storage),
					},
				},
			)
			if err != nil {
				slog.Error("error creating worker",
					"error", err,
				)
			}

			nodeAndDomainName = append(nodeAndDomainName, fmt.Sprintf("%s,%s", ipAddress, domainName))
		}
	}

	err = s.clusterRepository.UpdateIpAddressAndTokenByClusterId(
		createClusterContext,
		&entities.Cluster{
			ClusterID:     clusterEntity.ClusterID,
			ClusterStatus: "",
			IpAddress:     &createClusterResponse.MasterIpAddress,
			AccessToken:   &createClusterResponse.DashboardToken,
		},
	)
	if err != nil {
		return createClusterResponse, err
	}

	return createClusterResponse, nil
}

func (s *ClusterService) CleanCluster(
	ctx context.Context,
	job *entities.Job,
) error {
	return nil
}

func (s *ClusterService) GetUserCluster(
	ctx context.Context,
	user *entities.User,
) ([]*models.GetUserClusters, error) {
	getUserClusterContext, cancel := context.WithCancel(ctx)
	defer cancel()

	clusters, err := s.clusterRepository.GetClusterFromUserId(getUserClusterContext, user)
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

func (s *ClusterService) DeleteCluster(
	ctx context.Context,
	cluster *entities.Cluster,
) error {
	return s.clusterRepository.DeleteClusterByClusterId(ctx, cluster)
}

func (s *ClusterService) UpdateClusterStatusByClusterId(
	ctx context.Context,
	cluster *entities.Cluster,
) error {
	return s.clusterRepository.UpdateClusterStatusByClusterId(ctx, cluster)
}

func createRandomString(length int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}

	return string(b)
}
