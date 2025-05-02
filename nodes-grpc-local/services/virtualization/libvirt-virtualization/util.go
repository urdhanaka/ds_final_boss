package libvirt_virtualization

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

func generateRandom(stringLength int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, stringLength)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}

	return string(b)
}

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
