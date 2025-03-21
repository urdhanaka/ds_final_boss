package virtualization

import (
	"fmt"
	"net/url"
	"os"

	"github.com/digitalocean/go-libvirt"
	"github.com/docker/docker/client"
)

// initialize docker connection using docker socket
func initDockerConnection() *client.Client {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to docker daemon: %s", err.Error())
		os.Exit(1)
	}

	return apiClient
}

// initialize libvirt connection using qemu
func initKvmConnection() *libvirt.Libvirt {
	uri, _ := url.Parse(string(libvirt.QEMUSystem))
	connection, err := libvirt.ConnectToURI(uri)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to qemu: %s", err.Error())
	}

	return connection
}
