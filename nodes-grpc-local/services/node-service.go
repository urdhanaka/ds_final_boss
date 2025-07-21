package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	proto_model "nodes-grpc-local/services/model/proto-model"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

var mut sync.Mutex

type NodeServer struct {
	proto_model.UnimplementedNodeServiceServer

	// queue                 *queue.Queue
	libvirtVirtualization *libvirt_virtualization.LibvirtVirtualization

	// hacky way
	mainServerIpAddress string
}

func (s *NodeServer) CreateMaster(
	ctx context.Context,
	createMasterRequest *proto_model.CreateMasterRequest,
) (*proto_model.CreateMasterResponse, error) {
	mut.Lock()
	defer mut.Unlock()

	provisionCtx, cancel := context.WithTimeout(
		ctx,
		time.Second*PROVISIONING_TIME,
	)
	defer cancel()

	res := new(proto_model.CreateMasterResponse)
	res.CreationStatus = new(proto_model.CreationStatus)

	instanceName := createMasterRequest.Requirements.NodeName

	virtSpecs := virtualization_model.CreateInstanceRequest{
		ClusterName:         createMasterRequest.GetClusterName(),
		Name:                instanceName,
		IsMaster:            true,
		Token:               createMasterRequest.GetClusterToken(),
		Cpu:                 int(createMasterRequest.Requirements.GetVcpu()),
		Memory:              int(createMasterRequest.Requirements.GetMemory()),
		Storage:             int(createMasterRequest.Requirements.GetStorage()),
		MainServerIpAddress: s.mainServerIpAddress,
		ClusterId:           createMasterRequest.GetClusterId(),
	}

	virtResult, err := s.libvirtVirtualization.CreateInstance(provisionCtx, &virtSpecs)
	if err != nil {
		return res, err
	}

	res.KubeconfigContents = virtResult.KubeconfigContents
	res.DashboardToken = virtResult.DashboardToken
	res.MasterIpAddress = virtResult.MasterIpAddress
	res.CreationStatus.Success = virtResult.CreationStatus
	res.CreationStatus.Message = virtResult.Message

	// err := s.queue.AddToSpawnQueue(provisionCtx, virtSpecs)
	// if err != nil {
	// 	return res, err
	// }

	// sub := s.queue.Subscribe(ctx, instanceName)
	// defer sub.Close()
	//
	// msgCh := sub.Channel()
	// select {
	// case msg := <-msgCh:
	// 	instanceRes := new(virtualization_model.VirtCreateInstanceResponse)
	//
	// 	_ = json.Unmarshal([]byte(msg.Payload), instanceRes)
	//
	// 	res.DashboardToken = instanceRes.DashboardToken
	// 	res.MasterIpAddress = instanceRes.MasterIpAddress
	//
	// 	// provisioning status
	// 	res.CreationStatus.Success = instanceRes.CreationStatus
	// 	res.CreationStatus.Message = instanceRes.Message
	//
	// case <-provisionCtx.Done():
	// 	slog.Info("timeout exceeded")
	//
	// 	res.CreationStatus.Success = false
	// 	res.CreationStatus.Message = "timeout exceeded"
	// }

	return res, nil
}

func (s *NodeServer) CreateWorker(
	ctx context.Context,
	createWorkerRequest *proto_model.CreateWorkerRequest,
) (*proto_model.CreateWorkerResponse, error) {
	mut.Lock()
	defer mut.Unlock()

	provisionCtx, cancel := context.WithTimeout(
		ctx,
		time.Second*PROVISIONING_TIME,
	)
	defer cancel()

	res := new(proto_model.CreateWorkerResponse)
	res.CreationStatus = new(proto_model.CreationStatus)

	instanceName := createWorkerRequest.Requirements.NodeName

	virtSpecs := virtualization_model.CreateInstanceRequest{
		ClusterName:     createWorkerRequest.ClusterName,
		Name:            instanceName,
		IsMaster:        false,
		MasterIpAddress: createWorkerRequest.MasterIpAddress,
		Token:           createWorkerRequest.ClusterToken,
		Cpu:             int(createWorkerRequest.Requirements.GetVcpu()),
		Memory:          int(createWorkerRequest.Requirements.GetMemory()),
		Storage:         int(createWorkerRequest.Requirements.GetStorage()),
	}

	virtResult, err := s.libvirtVirtualization.CreateInstance(provisionCtx, &virtSpecs)
	if err != nil {
		return res, err
	}

	res.CreationStatus.Success = virtResult.CreationStatus
	res.CreationStatus.Message = virtResult.Message

	// err := s.queue.AddToSpawnQueue(provisionCtx, virtSpecs)
	// if err != nil {
	// 	return res, err
	// }

	// sub := s.queue.Subscribe(ctx, instanceName)
	// defer sub.Close()
	//
	// msgCh := sub.Channel()
	// select {
	// case msg := <-msgCh:
	// 	instanceRes := new(virtualization_model.VirtCreateInstanceResponse)
	//
	// 	_ = json.Unmarshal([]byte(msg.Payload), instanceRes)
	//
	// 	// provisioning status
	// 	res.CreationStatus.Success = instanceRes.CreationStatus
	// 	res.CreationStatus.Message = instanceRes.Message
	//
	// case <-provisionCtx.Done():
	// 	slog.Info("timeout exceeded")
	//
	// 	res.CreationStatus.Success = false
	// 	res.CreationStatus.Message = "timeout exceeded"
	// }

	return res, nil
}

func (s *NodeServer) NodeStatus(
	ctx context.Context,
	nodeStatusRequest *proto_model.NodeStatusRequest,
) (*proto_model.NodeStatusResponse, error) {
	CpuUsage, err := getCpuStatus()
	if err != nil {
		return &proto_model.NodeStatusResponse{}, err
	}

	memoryStat, err := getMemoryStatus()
	if err != nil {
		return &proto_model.NodeStatusResponse{}, err
	}

	storageStatus, err := getStorageStatus()
	if err != nil {
		return &proto_model.NodeStatusResponse{}, err
	}

	return &proto_model.NodeStatusResponse{
		FreeVcpu:         uint32(CpuUsage.LogicalCounts) - uint32(CpuUsage.FreeLogical),
		MemoryAvailable:  uint32(memoryStat.Memory),
		StorageAvailable: uint32(storageStatus.Storage),
	}, nil
}

func (s *NodeServer) DeleteInstance(
	ctx context.Context,
	deleteInstanceRequest *proto_model.DeleteInstanceRequest,
) (*proto_model.DeleteInstanceResponse, error) {
	err := s.libvirtVirtualization.DeleteInstance(deleteInstanceRequest.InstanceName)
	if err != nil {
		slog.Error("error deleting instance",
			"error", err,
		)
		return &proto_model.DeleteInstanceResponse{}, err
	}

	return &proto_model.DeleteInstanceResponse{}, nil
}

func StartGrpcServer(connection *InitStruct) {
	lis, err := net.Listen("tcp", GRPC_PORT)
	if err != nil {
		slog.Error("could not start grpc server",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	s := grpc.NewServer()
	proto_model.RegisterNodeServiceServer(s, &NodeServer{
		libvirtVirtualization: connection.VirtualizationService,
		mainServerIpAddress:   connection.MainServerIp,
		// queue: connection.QueueService,
	})

	// background job
	// startWorker(connection)

	// connect to main server
	slog.Info("node is ready, connecting to main server")
	go func() {
		var res string
		var err error

		for range NODES_CONNECT_RETRIES {
			res, err = connectToServer(connection.MainServerIp, connection.Lab)
			if err != nil {
				slog.Error("StartGrpcServer(): failed to connect to main server",
					"error", err.Error(),
				)

				time.Sleep(10 * time.Second)

				continue
			} else {
				slog.Info("main server is responding",
					"response", res,
				)

				break
			}
		}

		if err != nil {
			slog.Error("could not connect to main server",
				"error", err,
			)
		}
	}()

	slog.Info(fmt.Sprintf("starting grpc server at %s", GRPC_PORT))

	if err = s.Serve(lis); err != nil {
		slog.Error("StartGrpcServer(): failed to serve",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}

func connectToServer(
	mainServerIp string,
	labName string,
) (string, error) {
	hostname := getHostname()
	ipAddress := getIpAddress()
	cpuStatus, _ := getCpuStatus()
	storageStatus, _ := getStorageStatus()
	memoryStatus, _ := getMemoryStatus()

	body := fmt.Appendf(
		nil,
		`{"hostname":"%s","ip_address":"%s","lab_name":"%s","vcpu":%d,"storage":%d,"memory":%d}`,
		hostname, ipAddress, labName, cpuStatus.LogicalCounts, storageStatus.Storage, memoryStatus.Memory,
	)
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/api/nodes", mainServerIp+":8000"), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	return string(respBody), nil
}

// NOTE: might not needed
// func startWorker(connection *InitStruct) {
// 	worker := queue.NewWorker(
// 		connection.QueueService,
// 		connection.VirtualizationService,
// 	)
//
// 	go worker.DoSpawnWork()
// 	go worker.DoDeleteWork()
// }

// func sendLogs(
// 	instanceName string,
// 	clusterName string,
// ) {
// 	var sock net.Conn
// 	var err error
//
// 	logSocketFile := libvirt_virtualization.INSTANCE_LOGS_DIR + "/" + instanceName + ".sock"
//
// 	for {
// 		sock, err = net.Dial("unix", logSocketFile)
// 		if err != nil {
// 			slog.Error(fmt.Sprintf("%s | error accessing socket file, retrying...", instanceName),
// 				"error", err,
// 			)
// 			time.Sleep(1 * time.Second)
//
// 			continue
// 		}
// 		break
// 	}
// 	defer sock.Close()
//
// 	u := url.URL{
// 		Scheme: "ws",
// 		Host:   MAIN_SERVER_URL_LOCAL, // directly send to the backend
// 		Path:   fmt.Sprintf("/api/logs/stream/%s", clusterName),
// 	}
// 	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
// 	if err != nil {
// 		slog.Error(fmt.Sprintf("%s | error dialing websocket", instanceName),
// 			"error", err,
// 		)
// 		return
// 	}
// 	defer c.Close()
//
// 	scanner := bufio.NewScanner(sock)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if err := c.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
// 			slog.Error("Send error:", "error", err)
// 			break
// 		}
// 	}
// }
