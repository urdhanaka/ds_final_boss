package monitor

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"libvirt.org/go/libvirt"
)

const (
	// qemu system mode
	LIBVIRT_SYSTEM_CONNECTION = "qemu:///system"

	// psql uri
	PSQL_URI = "postgresql://postgre:postgre@localhost:5432/ta"
)

func InitPostgreConnection() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), PSQL_URI)
	if err != nil {
		slog.Error("error connecting to psql database",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	return conn
}

func InitLibvirtConnection() *libvirt.Connect {
	conn, err := libvirt.NewConnect(LIBVIRT_SYSTEM_CONNECTION)
	if err != nil {
		slog.Error("error connecting to QEMU system",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	return conn
}
