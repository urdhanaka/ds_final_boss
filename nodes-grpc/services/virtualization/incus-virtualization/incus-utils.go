package incus_virtualization

import (
	"fmt"
	"log/slog"

	incus "github.com/lxc/incus/client"
)

func incusSlogFunction(instanceName string, message string, err error) {
	fullMessage := fmt.Sprintf("%s | %s", instanceName, message)

	if err != nil {
		slog.Error(fullMessage,
			"err", err.Error(),
		)
	} else {
		slog.Info(fullMessage)
	}
}

func incusCleanup(incusConnection incus.InstanceServer, instanceName string) error {
	return nil
}
