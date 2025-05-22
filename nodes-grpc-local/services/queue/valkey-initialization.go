package queue

import (
	"log/slog"
	"os"

	"github.com/valkey-io/valkey-glide/go/api"
)

func InitValkeyConnection() api.GlideClientCommands {
	valkeyConf := api.NewGlideClientConfiguration().WithAddress(&api.NodeAddress{
		Host: VALKEY_HOST,
		Port: VALKEY_PORT,
	})

	valkeyClient, err := api.NewGlideClient(valkeyConf)
	if err != nil {
		slog.Error("error connecting to valkey address",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	// connection test
	_, err = valkeyClient.Ping()
	if err != nil {
		slog.Error("valkey error",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	return valkeyClient
}
