package config

import (
	"context"
	"embed"
	"errors"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PSQL_URL      = "postgres://root:root@localhost:5432"
	PSQL_ROOT_URL = "postgres://root:root@localhost:5432/root?sslmode=disable"
)

func NewPsqlConnection(migrationFS embed.FS) *pgxpool.Pool {
	conn, err := pgxpool.New(context.Background(), PSQL_URL)
	if err != nil {
		slog.Error("Could not connect to psql",
			"error", err,
		)
		os.Exit(1)
	}

	// check connection
	err = conn.Ping(context.Background())
	if err != nil {
		slog.Error("Could not connect to psql",
			"error", err,
		)
		os.Exit(1)
	}

	d, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		slog.Error("Could not create new migrate instances",
			"error", err,
		)
		os.Exit(1)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, PSQL_ROOT_URL)
	if err != nil {
		slog.Error("Could not create new migrate instances",
			"error", err,
		)
		os.Exit(1)
	}
	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			slog.Error("Could not create the tables",
				"error", err,
			)
			os.Exit(1)
		}
	}

	return conn
}

func DropTable(migrationFS embed.FS) {
	d, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		slog.Error("Could not create new migrate instances",
			"error", err,
		)
		os.Exit(1)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, PSQL_ROOT_URL)
	if err != nil {
		slog.Error("Could not create new migrate instances",
			"error", err,
		)
		os.Exit(1)
	}
	err = m.Down()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			slog.Error("Could not create the tables",
				"error", err,
			)
			os.Exit(1)
		}
	}
}
