package virtualization

import (
	"fmt"
	"log/slog"
)

func SlogFunction(instanceName string, message string, err error) {
	fullMessage := fmt.Sprintf("%s | %s", instanceName, message)

	if err != nil {
		slog.Error(fullMessage,
			"err", err.Error(),
		)
	} else {
		slog.Info(fullMessage)
	}
}
