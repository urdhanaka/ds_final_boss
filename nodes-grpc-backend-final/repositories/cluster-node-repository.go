package repositories

import (
	"context"
	"nodes-grpc-be/entities"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ClusterNodeRepository struct {
	dbPool *pgxpool.Pool
}

func NewClusterNodeRepository(
	dbPool *pgxpool.Pool,
) *ClusterNodeRepository {
	return &ClusterNodeRepository{
		dbPool,
	}
}

func (r *ClusterNodeRepository) AddEntry(
	ctx context.Context,
	cluster *entities.Cluster,
	node *entities.Node,
	instanceName string,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
    defer conn.Release()

	_, err = conn.Exec(
		ctx,
		"INSERT INTO cluster_nodes (cluster_id, node_id, instance_name) VALUES ($1, $2, $3)",
		cluster.ClusterID, node.NodeID, instanceName,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterNodeRepository) DeleteEntries(
	ctx context.Context,
	cluster *entities.Cluster,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
    defer conn.Release()

	_, err = conn.Exec(
		ctx,
		"DELETE FROM cluster_nodes where cluster_id=$1",
		cluster.ClusterID,
	)
	if err != nil {
		return err
	}

	return nil
}
