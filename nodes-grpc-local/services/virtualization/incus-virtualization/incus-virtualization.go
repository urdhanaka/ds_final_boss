package incus_virtualization

import (
	"context"
	"fmt"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"nodes-grpc-local/services/virtualization"
	"nodes-grpc-local/services/virtualization/embedded"

	incus "github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
)

type IncusVirtualization struct {
	incusConnection incus.InstanceServer
}

func NewIncusVirtualization(
	incusConnection incus.InstanceServer,
) virtualization.VirtualizationInterface {
	return &IncusVirtualization{
		incusConnection: incusConnection,
	}
}

func (c *IncusVirtualization) CreateMaster() error {
	return nil
}

func (c *IncusVirtualization) CreateWorker() error {
	return nil
}

func (c *IncusVirtualization) createBase(
	ctx context.Context,
	instanceName string,
	instanceRequest virtualization_model.CreateInstanceRequest,
) error {
	embedded := embedded.ReturnEmbedded()
	profileFile, err := embedded.ReadFile("files/user-cloud-init.yaml")
	if err != nil {
		virtualization.SlogFunction(instanceName, "error reading cloud-init file", err)

		return err
	}

	req := api.InstancesPost{
		InstancePut: api.InstancePut{
			Architecture: "amd64",
			Config: map[string]string{
				"security.secureboot":  "false",
				"cloud-init.user-data": string(profileFile),
				"limits.cpu":           fmt.Sprintf("%d", instanceRequest.Cpu),
				"limits.memory":        fmt.Sprintf("%dGiB", instanceRequest.Memory),
			},
			Ephemeral: true, // delete the instance when poweroff
		},
		Name: instanceName,
		Type: api.InstanceTypeVM,
		Source: api.InstanceSource{
			Type:     "image",
			Alias:    "debian/bookworm/cloud",
			Server:   "https://images.linuxcontainers.org",
			Protocol: "simplestreams",
		},
		Start: true,
	}
	instanceOp, err := c.incusConnection.CreateInstance(req)
	if err != nil {
		virtualization.SlogFunction(instanceName, "error creating base instance", err)

		return err
	}
	err = instanceOp.Wait()
	if err != nil {
		virtualization.SlogFunction(instanceName, "error creating base instance", err)

		return err
	}
	return nil
}
