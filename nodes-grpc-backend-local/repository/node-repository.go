package repository

import (
	"context"
	"nodes-grpc-backend-local/entity"

	"github.com/jackc/pgx/v5"
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

func (r *NodeRepository) GetAll(ctx context.Context) ([]entity.Node, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(ctx, "SELECT * FROM nodes")
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.Node])
}

func (r *NodeRepository) AddNode(ctx context.Context, node *entity.Node) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(
		ctx,
		"INSERT INTO nodes (node_id, hostname, ip_address, group_id, cpu, ram, storage) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		node.NodeID, node.Hostname, node.IpAddress, node.GroupId, node.Cpu, node.Ram, node.Storage,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *NodeRepository) GetNodesFromGroup(
	ctx context.Context,
	groupId int,
) ([]entity.Node, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(ctx, "SELECT * FROM nodes WHERE group_id=$1", groupId)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.Node])
}

func (r *NodeRepository) DeleteNode(ctx context.Context, node *entity.Node) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(
		ctx,
		"DELETE FROM nodes WHERE ip_address=$1",
		node.IpAddress,
	)
	if err != nil {
		return err
	}

	return nil
}
