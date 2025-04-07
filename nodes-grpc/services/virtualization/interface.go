package virtualization

import (
	"context"
	virtualization_model "nodes-grpc/common/model/virtualization"
)

type VirtualizationInterface interface {
	// Spawn(ctx context.Context, virtRequest virtualization_model.InstanceCreateRequest) error
	SpawnMaster(ctx context.Context, virtRequest virtualization_model.InstanceCreateRequest) (string, error)
	SpawnWorker(ctx context.Context, virtRequest virtualization_model.InstanceCreateRequest) error

	Stop(ctx context.Context, instanceIdentification virtualization_model.InstanceIdentification) error
	List() error
}
