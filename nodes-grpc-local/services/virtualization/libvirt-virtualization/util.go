package libvirt_virtualization

import (
	"fmt"
	"log/slog"
)

func slogFunction(clusterName string, instanceName string, message string, err error) {
	fullMessage := fmt.Sprintf("%s | %s | %s", clusterName, instanceName, message)

	if err != nil {
		slog.Error(fullMessage,
			"error", err,
		)
	} else {
		slog.Info(fullMessage)
	}
}
