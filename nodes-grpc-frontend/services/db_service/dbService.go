package db_service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"nodes-grpc-frontend/common/config"
	"nodes-grpc-frontend/common/model/web"

	"github.com/jmoiron/sqlx"
)

type DatabaseInterface interface {
	// node table
	StoreNode(ctx context.Context, nodeModel web.Node) error
	DeleteNode(ctx context.Context, nodeModel web.Node) error
	GetNode(ctx context.Context, nodeModel web.Node) (web.Node, error)
	GetAllNode(ctx context.Context) ([]web.Node, error)
}

type DbUsecaseImpl struct {
	dbConnection *sqlx.DB
}

func NewDbUsecaseImpl(dbConnection *sqlx.DB) DatabaseInterface {
	return &DbUsecaseImpl{
		dbConnection: dbConnection,
	}
}

func (d *DbUsecaseImpl) StoreNode(ctx context.Context, nodeModel web.Node) error {
	tx, err := d.dbConnection.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		fmt.Sprintf("INSERT INTO %s (uuid, node_ip, grpc_port) VALUES ($2, $3, $4);", config.NODE_TABLE_NAME),
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

func (d *DbUsecaseImpl) DeleteNode(ctx context.Context, nodeModel web.Node) error {
	tx, err := d.dbConnection.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf(
		"DELETE FROM %s WHERE node_ip = $2 OR uuid = $3;", config.NODE_TABLE_NAME),
		nodeModel.NodeIP, nodeModel.UUID,
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

func (d *DbUsecaseImpl) GetNode(ctx context.Context, nodeModel web.Node) (web.Node, error) {
	node := web.Node{}

	row := d.dbConnection.QueryRowx(fmt.Sprintf(`
        SELECT node_ip FROM %s WHERE node_ip = '%s';
    `, config.NODE_TABLE_NAME, nodeModel.NodeIP))
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

func (d *DbUsecaseImpl) GetAllNode(ctx context.Context) ([]web.Node, error) {
	nodes := make([]web.Node, 0)

	rows, err := d.dbConnection.Queryx(fmt.Sprintf(
		`SELECT node_ip, grpc_port FROM %s;`, config.NODE_TABLE_NAME),
	)
	if err != nil {
		return nodes, err
	}

	for rows.Next() {
		var node web.Node

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
