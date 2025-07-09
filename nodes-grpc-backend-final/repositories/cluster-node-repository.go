package repositories

import (
	"context"
	"nodes-grpc-be/entities"

	"github.com/jackc/pgx/v5"
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
		cluster.ClusterId, node.NodeID, instanceName,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterNodeRepository) GetNodesByClusterId(
	ctx context.Context,
	cluster *entities.ClusterNode,
) ([]entities.ClusterNode, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT * FROM cluster_nodes WHERE cluster_id=$1", cluster.ClusterId)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[entities.ClusterNode])
}

func (r *ClusterNodeRepository) DeleteEntriesByClusterId(
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
		"DELETE FROM cluster_nodes where cluster_id=$1", // delete all the instances of a cluster
		cluster.ClusterId,
	)
	if err != nil {
		return err
	}

	return nil
}
