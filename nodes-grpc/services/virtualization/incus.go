package virtualization

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	incus "github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
)

type IncusVirtualization struct {
	Connection incus.InstanceServer
}

func NewIncusVirtualization(isLocal bool) *IncusVirtualization {
	return &IncusVirtualization{
		Connection: initIncusConnection(isLocal),
	}
}

func initIncusConnection(isLocal bool) incus.InstanceServer {
	var incusConnection incus.InstanceServer
	var err error

	if isLocal {
		incusConnection, err = incus.ConnectIncusUnix("/run/incus/unix.socket", nil)
	} else {
		incusConnection, err = incus.ConnectIncusUnix("", nil)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to incus socket: %s\n", err.Error())
		os.Exit(1)
	}

	return incusConnection
}

func (d IncusVirtualization) CreateNewVM() error {
	newUuid := uuid.New()

	op, err := d.Connection.CreateInstance(api.InstancesPost{
		Name: newUuid.String(),
		Source: api.InstanceSource{
			Type:     "image",
			Alias:    "ubuntu/jammy/cloud",
			Server:   "https://images.linuxcontainers.org",
			Protocol: "simplestreams",
		},
		Type:  api.InstanceTypeVM,
		Start: true,
	})
	if err != nil {
		return err
	}

	err = op.Wait()
	if err != nil {
		return err
	}

	return nil
}
