package db

import (
	"errors"
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sqlx.DB {
	file_not_exist := false

	// check if db file exist
	_, err := os.Stat("node.db")
	if errors.Is(err, os.ErrNotExist) {
		file_not_exist = true

		_, err := os.Create("node.db")
		if err != nil {
			slog.Error("DB: cannot create sqlite db file",
				"error", err.Error(),
			)
			os.Exit(1)
		}
	}

	db, err := sqlx.Open("sqlite3", "node.db")
	if err != nil {
		slog.Error("DB: cannot open sqlite db file",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	if file_not_exist {
		if _, err := db.Exec(`
        CREATE TABLE Nodes (
        	UUID        BLOB PRIMARY KEY,
            NodeIP      TEXT,
            VirtType    TEXT
        );`); err != nil {
			slog.Error("DB: cannot create table inside db file",
				"error", err.Error(),
			)
			os.Exit(1)
		}
	}

	return db
}
