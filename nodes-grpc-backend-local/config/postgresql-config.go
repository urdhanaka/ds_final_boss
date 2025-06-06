package config

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// this can be handled better
	PSQL_URL = "postgres://root:root@localhost:5432"
)

func NewPsql() *pgxpool.Pool {
	conn, err := pgxpool.New(context.Background(), PSQL_URL)
	if err != nil {
		slog.Error("Could not connect to psql",
			"error", err,
		)

		os.Exit(1)
	}

	err = createTableIfNotExist(conn)

	return conn
}

func createTableIfNotExist(conn *pgxpool.Pool) error {
	return nil
}
