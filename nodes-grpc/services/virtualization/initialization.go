package virtualization

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/digitalocean/go-libvirt"
	"github.com/docker/docker/client"
	incus "github.com/lxc/incus/client"
)

// initialize docker connection using docker socket
func InitDockerConnection() *client.Client {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to docker daemon: %s", err.Error())
		os.Exit(1)
	}

	return apiClient
}

// initialize libvirt connection using qemu
func InitLibvirtConnection() *libvirt.Libvirt {
	uri, _ := url.Parse(string(libvirt.QEMUSystem))
	connection, err := libvirt.ConnectToURI(uri)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to qemu: %s", err.Error())
		os.Exit(1)
	}

	return connection
}

func InitIncusConnection() incus.InstanceServer {
	slog.Info("checking incus socket...")

	slog.Info("using /run/incus/unix.socket ...")
	c, err := incus.ConnectIncusUnix("/run/incus/unix.socket", nil)
	if err != nil {
		slog.Error("error connecting to /run/incus/unix.socket",
			"err", err.Error(),
		)
	}

	slog.Info("using /var/lib/incus/unix.socket ...")
	c, err = incus.ConnectIncusUnix("/var/lib/incus/unix.socket", nil)
	if err != nil {
		slog.Error("error connecting to /var/lib/incus/unix.socket",
			"err", err.Error(),
		)
		os.Exit(1)
	}

	return c
}
