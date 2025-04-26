package libvirt_virtualization

import (
	"errors"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"sync"

	"github.com/digitalocean/go-libvirt"
)

var cloudInitMut sync.Mutex
var networkMut sync.Mutex

func init() {
	slog.Info("checking libvirt requirements")

	slog.Info("checking if pool directory exists")

	_, err := os.Stat(POOL_DIR)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		slog.Info("pool directory doesn't exist, creating")

		err = os.Mkdir(POOL_DIR, 711)
		if err != nil {
			slog.Error("could not create pool directory",
				"error", err.Error())
			os.Exit(1)
		}
	}

	slog.Info("checking libvirt requirements completed")
}

func InitLibvirtConnection() *libvirt.Libvirt {
	uri, _ := url.Parse(string(libvirt.QEMUSystem))
	c, err := libvirt.ConnectToURI(uri)
	if err != nil {
		slog.Error("error connecting to QEMU system",
			"err", err.Error(),
		)
		os.Exit(1)
	}

	return c
}
