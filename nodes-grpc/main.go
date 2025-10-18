package main

import (
	"context"
	"fmt"
	"nodes-grpc/services"
	proto_model "nodes-grpc/services/model/proto-model"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// func prodStartGrpcServer() {
//     servicesStruct := services.NewConnection()
//
//     services.ProdStartGrpcServer(servicesStruct)
// }

func startGrpcServer(wg *sync.WaitGroup) {
	serviceStruct := services.NewConnection()

	services.StartGrpcServer(serviceStruct, wg)
}

func startGrpcUser() {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost%s", services.GRPC_PORT),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	c := proto_model.NewNodeServiceClient(conn)

	createMasterResponse, err := c.CreateMaster(context.Background(), &proto_model.CreateMasterRequest{
		Token: "12345",
		Requirements: &proto_model.CreateNodeRequirements{
			Cpu:     int64(2),
			Memory:  int64(2),
			Storage: int64(10),
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

    fmt.Println(createMasterResponse)

	// _, err = c.CreateWorker(context.Background(), &proto_model.CreateWorkerRequest{
	// 	Token:     "12345",
	// 	IpAddress: createMasterResponse.IpAddress,
	// 	Requirements: &proto_model.CreateNodeRequirements{
	// 		Cpu:     int64(2),
	// 		Memory:  int64(2),
	// 		Storage: int64(10),
	// 	},
	// })
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	os.Exit(1)
	// }
}

func main() {
	runtime.GOMAXPROCS(2)

	var wg sync.WaitGroup

	fmt.Println("starting server")
	go startGrpcServer(&wg)

	time.Sleep(5 * time.Second)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	startGrpcUser()

	<-sigChan
	fmt.Println("receive shutdown signal...")

	wg.Wait()
}
