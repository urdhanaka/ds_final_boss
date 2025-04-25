package services

import (
	"context"
	proto_model "nodes-grpc-local/services/model/proto-model"
)

type NodeServerInterface interface {
	CreateMaster(ctx context.Context, createMasterRequest *proto_model.CreateMasterRequest) (*proto_model.CreateMasterResponse, error)
	CreateWorker(c context.Context, workerRequest *proto_model.CreateWorkerRequest) (*proto_model.CreateWorkerResponse, error)
}
