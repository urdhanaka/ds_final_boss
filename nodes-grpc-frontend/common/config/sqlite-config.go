package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	NODE_TABLE_NAME      = "nodes"
	DASHBOARD_TABLE_NAME = "dashboard"

	FILE_NAME = "nodes.db"
)

func NewDB() *sqlx.DB {
	file_not_exist := false

	// check if db file exist
	_, err := os.Stat(FILE_NAME)
	if errors.Is(err, os.ErrNotExist) {
		file_not_exist = true

		_, err := os.Create(FILE_NAME)
		if err != nil {
			slog.Error("DB: cannot create sqlite db file",
				"error", err.Error(),
			)
			os.Exit(1)
		}
	}

	db, err := sqlx.Open("sqlite3", FILE_NAME)
	if err != nil {
		slog.Error("DB: cannot open sqlite db file",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	// create table
	execString := fmt.Sprintf(`
            -- nodes table, contains record of nodes running
            -- and it's ip and grpc port
            CREATE TABLE IF NOT EXISTS %s (
                uuid        TEXT PRIMARY KEY,
                node_ip     TEXT,
                grpc_port   TEXT,
                created_at  TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime')),
                deleted_at  TEXT
            );

            -- ui table, contains record of master node
            -- and it's ip
            CREATE TABLE IF NOT EXISTS %s (
                uuid        TEXT PRIMARY KEY,
                node_ip     TEXT,
                created_at  TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime'))
            );
            `,
		NODE_TABLE_NAME, DASHBOARD_TABLE_NAME,
	)

	if file_not_exist {
		if _, err := db.Exec(execString); err != nil {
			slog.Error("DB: cannot create table inside db file",
				"error", err.Error(),
			)
			os.Exit(1)
		}
	}

	return db
}
