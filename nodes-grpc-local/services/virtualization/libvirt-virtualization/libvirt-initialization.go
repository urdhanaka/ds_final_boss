package libvirt_virtualization

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"sync"

	"libvirt.org/go/libvirt"
)

var (
	cloudInitMut sync.Mutex
	imageMut     sync.Mutex
	networkMut   sync.Mutex
	efiMut       sync.Mutex
	spawnMut     sync.Mutex
)

func init() {
	slog.Info("checking libvirt requirements")

	slog.Info("checking if pool directory exists")
	_, err := os.Stat(POOL_DIR)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		slog.Info("pool directory doesn't exist, creating...")

		err = os.Mkdir(POOL_DIR, 0711)
		if err != nil {
			slog.Error("could not create pool directory",
				"error", err.Error())
			os.Exit(1)
		}
	}

	slog.Info("checking libvirt requirements completed")
}

func InitLibvirtConnection() *libvirt.Connect {
	c, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		slog.Error("error connecting to QEMU system",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	return c
}
