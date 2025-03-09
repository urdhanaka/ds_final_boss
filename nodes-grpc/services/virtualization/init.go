package virtualization

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/digitalocean/go-libvirt"
)

const (
	LIBVIRT_NOT_INITIALIZED = "libvirt is not initialized"
)

func initKvmConnection() *libvirt.Libvirt {
	uri, _ := url.Parse(string(libvirt.QEMUSystem))
	connection, err := libvirt.ConnectToURI(uri)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to qemu: %s", err.Error())
	}

	return connection
}

func initIncusConnection() net.Conn {
	sock, err := net.Dial("unix", "/var/lib/incus/unix.socket")
	if err != nil {
        fmt.Fprintf(os.Stderr, "could not connect to incus socket: %s", err.Error())
        os.Exit(1)
	}

    return sock
}
