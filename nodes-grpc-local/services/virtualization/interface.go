package virtualization

import (
	"context"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
)

type VirtualizationInterface interface {
	CreateInstance(ctx context.Context, virtRequest virtualization_model.CreateInstanceRequest) error
	StopInstance(ctx context.Context, instance virtualization_model.Instance) error
}
