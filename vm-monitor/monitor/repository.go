package monitor

import (
	"context"
	"database/sql"
	"errors"
	"ta-vm-monitor/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NodeRepository struct {
	pgxPool *pgxpool.Pool
}

func NewNodeRepository(pgxPool *pgxpool.Pool) *NodeRepository {
	return &NodeRepository{
		pgxPool,
	}
}

func (r *NodeRepository) IsNodeNameExists(ctx context.Context, name string) bool {
	nodeEntity := new(entity.Node)

	row := r.pgxPool.QueryRow(ctx, "SELECT id FROM nodes WHERE name=$1", name)
	err := row.Scan(nodeEntity)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return true
	}

	return false
}

func (r *NodeRepository) GetAllNodes(ctx context.Context) ([]entity.Node, error) {
	var nodeList []entity.Node

	queryString := "SELECT id FROM node_list"
	rows, err := r.pgxPool.Query(ctx, queryString)
	if err != nil {
		return nodeList, err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeEntity entity.Node

		err := rows.Scan(&nodeEntity)
		if err != nil {
			return nodeList, err
		}

		nodeList = append(nodeList, nodeEntity)
	}

	if err := rows.Err(); err != nil {
		return nodeList, err
	}

	return nodeList, nil
}
