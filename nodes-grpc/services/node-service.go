package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	proto_model "nodes-grpc/services/model/proto-model"
	virtualization_model "nodes-grpc/services/model/virtualization-model"
	"nodes-grpc/services/virtualization"
	"os"
	"sync"

	"google.golang.org/grpc"
)

type NodeServer struct {
	proto_model.UnimplementedNodeServiceServer

	virtualizationService virtualization.VirtualizationInterface
}

func NewNodeServer(
	virtualizationService virtualization.VirtualizationInterface,
) *NodeServer {
	return &NodeServer{
		virtualizationService: virtualizationService,
	}
}

// func connectToMainServer(mainServerIPAddress string) error {
// 	bodyBytes, _ := json.Marshal(thisNodeIdentification)
// 	bodyReader := bytes.NewReader(bodyBytes)
//
// 	req, err := http.NewRequest(http.MethodPost, mainServerIPAddress, bodyReader)
// 	if err != nil {
// 		slog.Error("Error connecting main server",
// 			"error", err.Error(),
// 		)
//
// 		return err
// 	}
//
// 	slog.Info("Connecting to main server...")
//
// 	req.Header.Set("Content-Type", "application/json")
// 	_, err = http.DefaultClient.Do(req)
// 	if err != nil {
// 		slog.Error("Error getting main server response",
// 			"error", err.Error(),
// 		)
//
// 		return err
// 	}
//
// 	return nil
// }

// func (n *NodeServer) ThisNodeStatus(context.Context, *model.Empty) (*model.Temp, error) {
// 	return new(model.Temp), nil
// }

// func (n *NodeServer) CreateNode(
// 	_ context.Context,
// 	nodeRequirements *model.NodeRequirements,
// ) (*model.Empty, error) {
// 	nodeCreateRequest := virtualization_model.NodeCreateRequest{
// 		IsMaster: true,
// 		Cpu:      nodeRequirements.Cpu,
// 		Memory:   nodeRequirements.Memory,
// 	}
//
// 	c := context.Background()
// 	contextId := virtualization_utils.CreateToken()
// 	c = context.WithValue(c, "contextId", contextId)
//
// 	slog.Info(fmt.Sprintf("starting process for %s context...", c.Value("contextId")))
//
// 	err := n.virtualizationService.Spawn(c, nodeCreateRequest)
// 	if err != nil {
// 		return new(model.Empty), err
// 	}
//
// 	return new(model.Empty), nil
// }

func (n *NodeServer) CreateMaster(
	c context.Context,
	masterRequirement *proto_model.CreateMasterRequest,
) (*proto_model.CreateMasterResponse, error) {
	// // check node availability first
	// isThisNodeAvailable := utils.IsNodeAvailable()
	// if !isThisNodeAvailable {
	// 	slog.Info("this node is not available")
	//
	// 	return new(model.Empty), errors.New("node not available")
	// }
	masterIpAddress, err := n.virtualizationService.CreateMaster(
		context.Background(),
		virtualization_model.NodeCreateRequest{
			IsMaster: true,
			Token:    masterRequirement.Token,
			Cpu:      masterRequirement.Requirements.GetCpu(),
			Memory:   masterRequirement.Requirements.GetMemory(),
			Storage:  masterRequirement.Requirements.GetStorage(),
		})
	if err != nil {
		return new(proto_model.CreateMasterResponse), err
	}

	return &proto_model.CreateMasterResponse{
		IpAddress: masterIpAddress,
	}, nil
}

func (n *NodeServer) CreateWorker(
	c context.Context,
	workerRequest *proto_model.CreateWorkerRequest,
) (*proto_model.CreateWorkerResponse, error) {
	// check node availability first
	// isThisNodeAvailable := utils.IsNodeAvailable()
	// if !isThisNodeAvailable {
	// 	slog.Info("this node is not available")
	//
	// 	return new(model.Empty), errors.New("node not available")
	// }

	err := n.virtualizationService.CreateWorker(context.Background(), virtualization_model.NodeCreateRequest{
		IsMaster:        false,
		MasterIpAddress: workerRequest.IpAddress,
		Token:           workerRequest.Token,
		Cpu:             workerRequest.Requirements.GetCpu(),
		Memory:          workerRequest.Requirements.GetMemory(),
		Storage:         workerRequest.Requirements.GetStorage(),
	})
	if err != nil {
		return new(proto_model.CreateWorkerResponse), err
	}

	return new(proto_model.CreateWorkerResponse), nil
}

func (n *NodeServer) StopNode(context.Context, *proto_model.Empty) (*proto_model.Empty, error) {
	return new(proto_model.Empty), nil
}

func StartGrpcServer(
	serviceStruct *Connection,
	waitGroup *sync.WaitGroup,
) {
	defer waitGroup.Done()

	lis, err := net.Listen("tcp", GRPC_PORT)
	if err != nil {
		slog.Error("grpc_server: could not start grpc server",
			"error", err.Error(),
		)
	}

	s := grpc.NewServer()

	proto_model.RegisterNodeServiceServer(s, &NodeServer{
		virtualizationService: serviceStruct.VirtualizationService,
	})

	slog.Info(fmt.Sprintf("StartGrpcServer(): starting server at %s", GRPC_PORT))

	if err := s.Serve(lis); err != nil {
		slog.Error("StartGrpcServer(): failed to serve",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}

func ProdStartGrpcServer(
	servicesStruct *Connection,
) {
	lis, err := net.Listen("tcp", GRPC_PORT)
	if err != nil {
		slog.Error("grpc_server: could not start grpc server",
			"error", err.Error(),
		)
	}

	s := grpc.NewServer()

	proto_model.RegisterNodeServiceServer(s, &NodeServer{
		virtualizationService: servicesStruct.VirtualizationService,
	})

	slog.Info(fmt.Sprintf("StartGrpcServer(): starting server at %s", GRPC_PORT))

	if err := s.Serve(lis); err != nil {
		slog.Error("StartGrpcServer(): failed to serve",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
