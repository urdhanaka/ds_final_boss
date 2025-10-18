package virtualization

import (
	"context"
	virtualization_model "nodes-grpc/services/model/virtualization-model"
)

type VirtualizationInterface interface {
	// Spawn(ctx context.Context, virtRequest virtualization_model.InstanceCreateRequest) error

	// SpawnMaster function returns the master vm ip address to be used on the worker and error
	CreateMaster(ctx context.Context, virtRequest virtualization_model.NodeCreateRequest) (string, error)

	// SpawnWorker function returns error
	CreateWorker(ctx context.Context, virtRequest virtualization_model.NodeCreateRequest) error
	StopNode(ctx context.Context, instance virtualization_model.InstanceIdentification) error

	// Stop(ctx context.Context, instanceIdentification virtualization_model.InstanceIdentification) error
	// List() error
}
