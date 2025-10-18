package repositories

import (
	"context"
	"database/sql"
	"errors"
	"nodes-grpc-be/entities"

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

// wrapper for add or update
func (r *NodeRepository) AddOrUpdateNode(
	ctx context.Context,
	node *entities.Node,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// get exists node by the ip address
	existsNode := &entities.Node{
		IpAddress: node.IpAddress,
	}
	err = r.GetNodeByIpAddress(ctx, existsNode)
	if errors.Is(err, sql.ErrNoRows) {
		return r.AddNode(ctx, node)
	}

	return r.UpdateNode(ctx, node)
}

// update by node_id
func (r *NodeRepository) UpdateNode(
	ctx context.Context,
	node *entities.Node,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(
		ctx,
		`UPDATE nodes
        SET hostname=$1, ip_address=$2, group_id=$3, vcpu=$4, memory=$5, storage=$6
        `,
		node.Hostname, node.IpAddress, node.GroupId, node.VCpu, node.Memory, node.Storage,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *NodeRepository) AddNode(ctx context.Context, node *entities.Node) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(
		ctx,
		"INSERT INTO nodes (node_id, hostname, ip_address, group_id, vcpu, memory, storage) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		node.NodeID, node.Hostname, node.IpAddress, node.GroupId, node.VCpu, node.Memory, node.Storage,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *NodeRepository) GetNodesFromGroup(
	ctx context.Context,
	groupId int,
) ([]entities.Node, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT * FROM nodes WHERE group_id=$1", groupId)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[entities.Node])
}

func (r *NodeRepository) DeleteNode(ctx context.Context, node *entities.Node) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

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

func (r *NodeRepository) GetNodeByHostnameAndGroupId(
	ctx context.Context,
	node *entities.Node,
) (*entities.Node, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	getNode := new(entities.Node)

	err = conn.QueryRow(
		ctx,
		"SELECT * FROM nodes WHERE hostname=$1 AND group_id=$2",
		node.Hostname, node.GroupId,
	).Scan(
		&getNode.NodeID,
		&getNode.Hostname,
		&getNode.IpAddress,
		&getNode.GroupId,
		&getNode.VCpu,
		&getNode.Memory,
		&getNode.Storage,
	)
	if err != nil {
		return nil, err
	}

	return getNode, nil
}

func (r *NodeRepository) UpdateNodeResourcesByNodeId(
	ctx context.Context,
	node *entities.Node,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(
		ctx,
		`UPDATE nodes
        SET vcpu = $1, memory = $2, storage = $3
        WHERE node_id = $4`,
		node.VCpu, node.Memory, node.Storage, node.NodeID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *NodeRepository) GetNodeByIpAddress(
	ctx context.Context,
	node *entities.Node,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil
	}
	defer conn.Release()

	err = conn.QueryRow(
		ctx,
		`SELECT node_id, ip_address
        FROM nodes
        WHERE ip_address=$1`,
		node.IpAddress,
	).Scan(
		&node.NodeID,
        &node.IpAddress,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *NodeRepository) GetNodeIpAddressByNodeId(
	ctx context.Context,
	node *entities.Node,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil
	}
	defer conn.Release()

	err = conn.QueryRow(
		ctx,
		`SELECT ip_address
        FROM nodes
        WHERE node_id=$1`,
		node.NodeID,
	).Scan(
		&node.IpAddress,
	)
	if err != nil {
		return err
	}

	return nil
}
