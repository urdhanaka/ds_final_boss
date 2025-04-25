package incus_virtualization

import (
	"log/slog"
	"os"

	incus "github.com/lxc/incus/client"
)

func InitIncusConnection() incus.InstanceServer {
	slog.Info("connecting to incus server using /run/incus/unix.socket")

	c, err := incus.ConnectIncusUnix("/run/incus/unix.socket", nil)
	if err != nil {
		slog.Error("error connecting to /run/incus/unix.socket",
			"err", err.Error(),
		)
		os.Exit(1)
	}

	return c
}
