package virtualization

import (
	"context"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
)

type VirtualizationInterface interface {
	CreateMaster(ctx context.Context, virtRequest virtualization_model.CreateInstanceRequest) error
	CreateWorker(ctx context.Context, virtRequest virtualization_model.CreateInstanceRequest) error
}
