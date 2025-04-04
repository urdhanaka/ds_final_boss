package db_service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"nodes-grpc-frontend/common/config"

	"github.com/jmoiron/sqlx"
)

type Node struct {
	UUID     string `json:"uuid" db:"uuid"`
	NodeIP   string `json:"node_ip" db:"node_ip"`
	GrpcPort string `json:"grpc_port" db:"grpc_port"`
}

type DatabaseInterface interface {
	Store(ctx context.Context, nodeModel Node) error
	Delete(ctx context.Context, nodeModel Node) error
	Get(ctx context.Context, nodeModel Node) (Node, error)
	GetAll(ctx context.Context) ([]Node, error)
}

type DbUsecaseImpl struct {
	dbConnection *sqlx.DB
}

func NewDbUsecaseImpl(dbConnection *sqlx.DB) DatabaseInterface {
	return &DbUsecaseImpl{
		dbConnection: dbConnection,
	}
}

func (d *DbUsecaseImpl) Store(ctx context.Context, nodeModel Node) error {
	tx, err := d.dbConnection.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		fmt.Sprintf("INSERT INTO %s (uuid, node_ip, grpc_port) VALUES ($2, $3, $4);", config.TABLE_NAME),
		nodeModel.UUID, nodeModel.NodeIP, nodeModel.GrpcPort,
	)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *DbUsecaseImpl) Delete(ctx context.Context, nodeModel Node) error {
	tx, err := d.dbConnection.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM $1 WHERE node_ip = $2;",
		config.TABLE_NAME, nodeModel.NodeIP,
	)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *DbUsecaseImpl) Get(ctx context.Context, nodeModel Node) (Node, error) {
	node := Node{}

	row := d.dbConnection.QueryRowx(fmt.Sprintf(`
        SELECT node_ip FROM %s WHERE node_ip = '%s';
    `, config.TABLE_NAME, nodeModel.NodeIP))
	err := row.StructScan(&node)
	if err != nil {
		// handle empty result
		if errors.Is(err, sql.ErrNoRows) {
			return node, nil
		}

		return node, err
	}

	return node, nil
}

func (d *DbUsecaseImpl) GetAll(ctx context.Context) ([]Node, error) {
	nodes := make([]Node, 0)

	rows, err := d.dbConnection.Queryx(fmt.Sprintf(
		`SELECT node_ip, grpc_port FROM %s;`, config.TABLE_NAME),
	)
	if err != nil {
		return nodes, err
	}

	for rows.Next() {
		var node Node

		err = rows.StructScan(&node)
		if err != nil {
			slog.Error("GetAll(): error getting value from db",
				"error", err.Error(),
			)

			continue
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
