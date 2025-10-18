package virtualization_utils

import (
	virtualization_model "nodes-grpc/services/model/virtualization-model"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

func NewContainerConfig(
	resourceConfig *virtualization_model.VirtualizationCreateRequest,
) *container.Config {
	return &container.Config{
		Image:           "docker-virt",
		NetworkDisabled: false,
	}
}

func NewContainerHostConfig() *container.HostConfig {
	return &container.HostConfig{
		Privileged: true,
	}
}

// for now, it needs only memory and cpu shares.
//
// probably need more than that
func NewContainerResourceConfig(
	resourceConfig *virtualization_model.VirtualizationCreateRequest,
) *container.Resources {
	return &container.Resources{
		CPUShares: resourceConfig.Cpu,
		Memory:    resourceConfig.Memory,
	}
}

func NewContainerNetworkConfig() *network.NetworkingConfig {
	return &network.NetworkingConfig{}
}
