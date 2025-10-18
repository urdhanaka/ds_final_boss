package db

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
)

const (
	TABLE_NAME = "Nodes"
)

type DatabaseConnection struct {
	dbConnection *sqlx.DB
}

func NewDBConnection(dbConnection *sqlx.DB) DatabaseInterface {
	return &DatabaseConnection{
		dbConnection: dbConnection,
	}
}

func (c *DatabaseConnection) Store(nodeModel NodesModel) error {
	tx, err := c.dbConnection.Beginx()
	if err != nil {
		slog.Error("db: Store(): could not start db transaction",
			"error", err.Error(),
		)
		return err
	}
	_, err = tx.Exec("INSERT INTO $1 (UUID, NodeIP) VALUES ($2, $3)",
		TABLE_NAME, nodeModel.UUID, nodeModel.NodeIP,
	)
	if err != nil {
		slog.Error("db: Store(): could not insert the node data on db",
			"error", err.Error(),
		)
		return err
	}
	err = tx.Commit()
	if err != nil {
		slog.Error("db: Store(): could not commit the transaction",
			"error", err.Error(),
		)
		return err
	}

	return nil
}

func (c *DatabaseConnection) Delete(nodeModel NodesModel) error {
	tx, err := c.dbConnection.Beginx()
	if err != nil {
		slog.Error("db: Delete(): could not start db transaction",
			"error", err.Error(),
		)
		return err
	}
	_, err = tx.Exec("DELETE FROM $1 WHERE UUID = $2", TABLE_NAME, nodeModel.UUID)
	if err != nil {
		slog.Error("db: Delete(): could not delete the node data on db",
			"error", err.Error(),
		)
		return err
	}
	err = tx.Commit()
	if err != nil {
		slog.Error("db: Delete(): could not commit the transaction",
			"error", err.Error(),
		)
		return err
	}

	return nil
}
