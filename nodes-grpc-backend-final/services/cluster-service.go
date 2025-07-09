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
	"slices"
	"strings"
	"time"
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

func (s *ClusterService) AddClusterToDatabase(
	ctx context.Context,
	cluster *entities.Cluster,
) error {
	return s.clusterRepository.AddCluster(ctx, cluster)
}

func (s *ClusterService) CreateClusterWithoutPickTest(
	ctx context.Context,
	job *entities.Job,
) (*models.CreateClusterResponse, error) {
	createClusterContext, cancel := context.WithCancel(ctx)
	defer cancel()

	createClusterResponse := new(models.CreateClusterResponse)
	addClusterModel := new(models.AddCluster)

	jsonBytes, _ := json.Marshal(job.Payload)
	_ = json.Unmarshal(jsonBytes, addClusterModel)

	// start the cluster creation here
	// keep track of the nodes and the domain name
	// format: <ip_address>,<instance_name>
	nodeAndInstanceName := []string{}
	isCreateSuccess := false
	defer func() {
		if len(nodeAndInstanceName) == 0 {
			return
		}

		if !isCreateSuccess {
			for _, node := range nodeAndInstanceName {
				ipAddress := strings.Split(node, ",")[0]
				domainName := strings.Split(node, ",")[1]

				grpcClient, _ := config.NewNodeClient(ipAddress)
				grpcClient.DeleteInstance(context.Background(), &proto_model.DeleteInstanceRequest{
					InstanceName: domainName,
				})
			}
		}
	}()

	// keep track of master node creation
	isMasterNode := true
	currentDomainIndex := 0

	// cluster token to create the cluster using k3s
	clusterToken := createRandomString(8)

	// get group nodes
	thisGroupNodes, err := s.nodeRepository.GetNodesFromGroup(createClusterContext, addClusterModel.GroupId)
	if err != nil {
		return createClusterResponse, err
	}

	for currentDomainIndex < addClusterModel.NodeSize {
		randomIndex := rand.Intn(len(thisGroupNodes))
		pickedNode := thisGroupNodes[randomIndex]

		grpcClient, err := config.NewNodeClient(pickedNode.IpAddress)
		if err != nil {
			slog.Error("could not connect to nodes, removing the nodes",
				"error", err,
			)

			continue
		}

		// check if node can be used
		nodeStatus, _ := grpcClient.NodeStatus(
			createClusterContext,
			&proto_model.NodeStatusRequest{},
		)

		// check resources
		if addClusterModel.Storage > int(nodeStatus.StorageAvailable) ||
			addClusterModel.Memory > int(nodeStatus.MemoryAvailable) {
			slog.Info("worker node has less resources than needed, removing from current loop",
				"hostname", pickedNode.Hostname,
			)

			thisGroupNodes = slices.Delete(thisGroupNodes, randomIndex, randomIndex+1)

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
					ClusterId: addClusterModel.ClusterId.String(),
				},
			)
			if err != nil {
				slog.Error("error creating master",
					"error", err,
				)

				return createClusterResponse, err
			}
			if !res.CreationStatus.Success {
				slog.Error("error creating cluster",
					"error", res.CreationStatus.Message,
				)

				return createClusterResponse, err
			}

			// master node created
			isMasterNode = false
			createClusterResponse.DashboardToken = res.DashboardToken
			createClusterResponse.MasterIpAddress = res.MasterIpAddress
			createClusterResponse.KubeconfigContents = res.KubeconfigContents

			// add the masterNode to the array
			nodeAndInstanceName = append(nodeAndInstanceName, fmt.Sprintf("%s,%s", pickedNode.IpAddress, domainName))
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

				return createClusterResponse, err
			}

			nodeAndInstanceName = append(nodeAndInstanceName, fmt.Sprintf("%s,%s", pickedNode.IpAddress, domainName))
		}

		currentDomainIndex += 1
	}

	isCreateSuccess = true

	return createClusterResponse, nil
}

func (s *ClusterService) CreateClusterWithoutPick(
	ctx context.Context,
	job *entities.Job,
) (*models.CreateClusterResponse, error) {
	createClusterContext, cancel := context.WithCancel(ctx)
	defer cancel()

	createClusterResponse := new(models.CreateClusterResponse)
	addClusterModel := new(models.AddCluster)

	jsonBytes, _ := json.Marshal(job.Payload)
	_ = json.Unmarshal(jsonBytes, addClusterModel)

	// start the cluster creation here
	// keep track of the nodes and the domain name
	// format: <ip_address>,<instance_name>
	nodeAndInstanceName := []string{}
	isCreateSuccess := false
	defer func() {
		if len(nodeAndInstanceName) == 0 {
			return
		}

		if !isCreateSuccess {
			for _, node := range nodeAndInstanceName {
				ipAddress := strings.Split(node, ",")[0]
				domainName := strings.Split(node, ",")[1]

				grpcClient, _ := config.NewNodeClient(ipAddress)
				grpcClient.DeleteInstance(context.Background(), &proto_model.DeleteInstanceRequest{
					InstanceName: domainName,
				})
			}

			s.clusterNodeRepository.DeleteEntriesByClusterId(
				createClusterContext,
				&entities.Cluster{
					ClusterId: addClusterModel.ClusterId,
				},
			)
		}
	}()

	// keep track of master node creation
	isMasterNode := true
	currentDomainIndex := 0

	// cluster token to create the cluster using k3s
	clusterToken := createRandomString(8)

	// get group nodes
	thisGroupNodes, err := s.nodeRepository.GetNodesFromGroup(createClusterContext, addClusterModel.GroupId)
	if err != nil {
		return createClusterResponse, err
	}

	for currentDomainIndex < addClusterModel.NodeSize {
		randomIndex := rand.Intn(len(thisGroupNodes))
		pickedNode := thisGroupNodes[randomIndex]

		grpcClient, err := config.NewNodeClient(pickedNode.IpAddress)
		if err != nil {
			slog.Error("could not connect to nodes, removing the nodes",
				"error", err,
			)

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
			slog.Info("worker node has less resources than needed, removing from current loop",
				"hostname", pickedNode.Hostname,
			)

			thisGroupNodes = slices.Delete(thisGroupNodes, randomIndex, randomIndex+1)

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
					ClusterId: addClusterModel.ClusterId.String(),
				},
			)
			if err != nil {
				slog.Error("error creating master",
					"error", err,
				)

				return createClusterResponse, err
			}
			if !res.CreationStatus.Success {
				slog.Error("error creating cluster",
					"error", res.CreationStatus.Message,
				)

				return createClusterResponse, err
			}

			// master node created
			isMasterNode = false
			createClusterResponse.DashboardToken = res.DashboardToken
			createClusterResponse.MasterIpAddress = res.MasterIpAddress
			createClusterResponse.KubeconfigContents = res.KubeconfigContents

			// add the masterNode to the array
			nodeAndInstanceName = append(nodeAndInstanceName, fmt.Sprintf("%s,%s", pickedNode.IpAddress, domainName))

			// add to the clusterNode
			err = s.clusterNodeRepository.AddEntry(
				createClusterContext,
				&entities.Cluster{ClusterId: addClusterModel.ClusterId},
				&entities.Node{NodeID: pickedNode.NodeID},
				domainName,
			)
			if err != nil {
				return createClusterResponse, err
			}
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

				return createClusterResponse, err
			}

			nodeAndInstanceName = append(nodeAndInstanceName, fmt.Sprintf("%s,%s", pickedNode.IpAddress, domainName))

			// add to the clusterNode
			err = s.clusterNodeRepository.AddEntry(
				createClusterContext,
				&entities.Cluster{ClusterId: addClusterModel.ClusterId},
				&entities.Node{NodeID: pickedNode.NodeID},
				domainName,
			)
			if err != nil {
				return createClusterResponse, err
			}
		}

		currentDomainIndex += 1
	}

	err = s.clusterRepository.UpdateIpAddressAndTokenByClusterId(
		createClusterContext,
		&entities.Cluster{
			ClusterId:          addClusterModel.ClusterId,
			IpAddress:          &createClusterResponse.MasterIpAddress,
			AccessToken:        &createClusterResponse.DashboardToken,
			KubeconfigContents: createClusterResponse.KubeconfigContents,
		},
	)
	if err != nil {
		return createClusterResponse, err
	}

	isCreateSuccess = true

	return createClusterResponse, nil
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

	// start the cluster creation here
	// keep track of the nodes and the domain name
	// format: <ip_address>,<instance_name>
	nodeAndInstanceName := []string{}
	isCreateSuccess := false
	defer func() {
		if len(nodeAndInstanceName) == 0 {
			return
		}

		if !isCreateSuccess {
			for _, node := range nodeAndInstanceName {
				ipAddress := strings.Split(node, ",")[0]
				domainName := strings.Split(node, ",")[1]

				grpcClient, _ := config.NewNodeClient(ipAddress)
				grpcClient.DeleteInstance(context.Background(), &proto_model.DeleteInstanceRequest{
					InstanceName: domainName,
				})
			}
		}
	}()

	// keep track of master node creation
	isMasterNode := true
	currentDomainIndex := 0

	// cluster token to create the cluster using k3s
	clusterToken := createRandomString(8)

	for currentDomainIndex < addClusterModel.NodeSize {
		nodesCopy := addClusterModel.Nodes

		// pickedIndex := currentDomainIndex % addClusterModel.NodeSize % len(nodesCopy)
		pickedIndex := 0
		nodeIdentity := nodesCopy[pickedIndex]

		ipAddress := strings.Split(nodeIdentity, ",")[1]

		grpcClient, err := config.NewNodeClient(ipAddress)
		if err != nil {
			slog.Error("could not connect to nodes, removing the nodes",
				"error", err,
			)

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

			// need to pop the node out of available nodes
			nodesCopy = slices.Delete(nodesCopy, pickedIndex, pickedIndex+1)

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
			if !res.CreationStatus.Success {
				slog.Error("error creating cluster",
					"error", res.CreationStatus.Message,
				)

				return createClusterResponse, err
			}

			// master node created
			isMasterNode = false
			createClusterResponse.DashboardToken = res.DashboardToken
			createClusterResponse.MasterIpAddress = res.MasterIpAddress

			// add the masterNode to the array
			nodeAndInstanceName = append(nodeAndInstanceName, fmt.Sprintf("%s,%s", ipAddress, domainName))
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

			nodeAndInstanceName = append(nodeAndInstanceName, fmt.Sprintf("%s,%s", ipAddress, domainName))
		}

		currentDomainIndex += 1
	}

	err := s.clusterRepository.UpdateIpAddressAndTokenByClusterId(
		createClusterContext,
		&entities.Cluster{
			ClusterId:   addClusterModel.ClusterId,
			IpAddress:   &createClusterResponse.MasterIpAddress,
			AccessToken: &createClusterResponse.DashboardToken,
		},
	)
	if err != nil {
		return createClusterResponse, err
	}

	for _, node := range nodeAndInstanceName {
		ipAddress := strings.Split(node, ",")[0]
		domainName := strings.Split(node, ",")[1]

		nodeEntity := &entities.Node{IpAddress: ipAddress}

		err := s.nodeRepository.GetNodeByIpAddress(
			ctx,
			nodeEntity,
		)
		if err != nil {
			slog.Error("error saving cluster nodes",
				"error", err,
			)
		}

		err = s.clusterNodeRepository.AddEntry(
			ctx,
			&entities.Cluster{ClusterId: addClusterModel.ClusterId},
			nodeEntity,
			domainName,
		)
		if err != nil {
			slog.Error("error saving cluster nodes",
				"error", err,
			)
		}
	}

	isCreateSuccess = true

	return createClusterResponse, nil
}

func (s *ClusterService) CleanCluster(
	ctx context.Context,
	job *entities.Job,
) error {
	cleanClusterContext, cancel := context.WithCancel(ctx)
	defer cancel()

	deleteClusterModel := new(models.DeleteCluster)

	jsonBytes, _ := json.Marshal(job.Payload)
	_ = json.Unmarshal(jsonBytes, deleteClusterModel)

	nodeAndDomain, err := s.clusterNodeRepository.GetNodesByClusterId(
		cleanClusterContext,
		&entities.ClusterNode{ClusterId: deleteClusterModel.ClusterId},
	)
	if err != nil {
		return err
	}

	for _, currentNodeAndDomain := range nodeAndDomain {
		nodeEntity := &entities.Node{
			NodeID: currentNodeAndDomain.NodeId,
		}
		err := s.nodeRepository.GetNodeIpAddressByNodeId(cleanClusterContext, nodeEntity)
		if err != nil {
			slog.Error("could not get the node IP address",
				"error", err,
			)
		}

		grpcClient, err := config.NewNodeClient(nodeEntity.IpAddress)
		if err != nil {
			slog.Error("could not connect to nodes, could not delete the domain",
				"error", err,
			)

			continue
		}

		_, err = grpcClient.DeleteInstance(
			cleanClusterContext,
			&proto_model.DeleteInstanceRequest{
				InstanceName: currentNodeAndDomain.InstanceName,
			},
		)
		if err != nil {
			slog.Error("worker node could not delete the domain",
				"error", err,
			)
		}
	}

	err = s.clusterRepository.DeleteClusterByClusterId(
		ctx,
		&entities.Cluster{ClusterId: deleteClusterModel.ClusterId},
	)
	if err != nil {
		slog.Error("could not delete the data",
			"error", err,
		)
	}

	// err = s.clusterNodeRepository.DeleteEntriesByClusterId(
	// 	ctx,
	// 	&entities.Cluster{
	// 		ClusterId: deleteClusterModel.ClusterId,
	// 	},
	// )
	// if err != nil {
	// 	slog.Error("could not delete the data",
	// 		"error", err,
	// 	)
	// }

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
			ClusterId:     cluster.ClusterId,
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
		ClusterId:     getCluster.ClusterId,
		ClusterName:   getCluster.ClusterName,
		UserId:        getCluster.UserId,
		GroupId:       getCluster.GroupId,
		ClusterStatus: getCluster.ClusterStatus,
		IpAddress:     getCluster.IpAddress,
		AccessToken:   getCluster.AccessToken,
		CreatedAt:     getCluster.CreatedAt,
	}

	return clusterDetails, nil
}

func (s *ClusterService) GetClusterKubeconfig(
	ctx context.Context,
	cluster *entities.Cluster,
) (*models.GetClusterDetails, error) {
	getCluster, err := s.clusterRepository.GetKubeconfigFromClusterId(ctx, cluster)
	if err != nil {
		return nil, err
	}

	clusterDetails := &models.GetClusterDetails{
		KubeconfigContents: getCluster.KubeconfigContents,
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

func pickRandomIndexFromSlice(arrayOfString []string) int {
	return 0
}
