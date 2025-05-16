package libvirt_virtualization

import (
	"fmt"
	"log/slog"
)

func slogFunction(instanceName string, message string, err error) {
	fullMessage := fmt.Sprintf("%s | %s", instanceName, message)

	if err != nil {
		slog.Error(fullMessage,
			"err", err.Error(),
		)
	} else {
		slog.Info(fullMessage)
	}
}
