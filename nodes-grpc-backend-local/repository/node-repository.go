package repository

import (
	"context"
	"nodes-grpc-backend-local/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NodeRepository struct {
	dbPool *pgxpool.Pool
}

func NewNodeRepository(dbPool *pgxpool.Pool) *NodeRepository {
	return &NodeRepository{
		dbPool,
	}
}

func (r *NodeRepository) AddNode(ctx context.Context, node *entity.Node) error {
	_, err := r.dbPool.Exec(
		ctx,
		"INSERT INTO node (node_id, hostname, ip_address) VALUES ($1, $2, $3)",
		node.ID, node.Hostname, node.IpAddress,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *NodeRepository) DeleteNode(ctx context.Context, node *entity.Node) error {
	_, err := r.dbPool.Exec(
		ctx,
		"DELETE FROM node WHERE ip_address=$1",
		node.IpAddress,
	)
	if err != nil {
		return err
	}

	return nil
}
