package main

import (
	"nodes-grpc-frontend-local/src"
)

func main() {
	// nodeClient, err := config.NewNodeClient()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	//
	// slog.Info("spawning master...")
	// _, err = nodeClient.CreateMaster(context.Background(), &proto_model.CreateMasterRequest{})
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	//
	// time.Sleep(time.Second * 20)
	//
	// slog.Info("spawning worker...")
	// _, err = nodeClient.CreateWorker(context.Background(), &proto_model.CreateWorkerRequest{})
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	src.StartApp()
}
