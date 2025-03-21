package db

import (
	"database/sql"
)

type DatabaseConnection struct {
	dbConnection *sql.DB
}

func NewDBConnection(dbConnection *sql.DB) *DatabaseConnection {
	return &DatabaseConnection{
		dbConnection: dbConnection,
	}
}

func (c DatabaseConnection) Store() error {
	return nil
}
