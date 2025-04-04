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
	TABLE_NAME = "nodes"
	FILE_NAME  = "nodes.db"
)

func InitDB() *sqlx.DB {
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

	execString := []string{
		// create table
		fmt.Sprintf(`
            CREATE TABLE IF NOT EXISTS %s (
                uuid        TEXT PRIMARY KEY,
                node_ip     TEXT,
                grpc_port   TEXT,
                created_at  TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime')),
                deleted_at  TEXT
            );`,
			TABLE_NAME,
		),
	}
	if file_not_exist {
		for _, exec := range execString {
			if _, err := db.Exec(exec); err != nil {
				slog.Error("DB: cannot create table inside db file",
					"error", err.Error(),
				)
				os.Exit(1)
			}
		}
	}

	return db
}
