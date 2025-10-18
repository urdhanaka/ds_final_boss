package repository

import (
	"context"
	"nodes-grpc-backend-local/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ClusterRepository struct {
	dbPool *pgxpool.Pool
}

func NewClusterRepository(
	dbPool *pgxpool.Pool,
) *ClusterRepository {
	return &ClusterRepository{
		dbPool,
	}
}

func (r *ClusterRepository) AddCluster(
	ctx context.Context,
	cluster *entity.Cluster,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(
		ctx,
		"INSERT INTO clusters (cluster_id, name, user_id, group_id, created_at) VALUES ($1, $2, $3, $4, $5)",
		cluster.ClusterID, cluster.Name, cluster.UserID, cluster.GroupID, cluster.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterRepository) DeleteClusterById(
	ctx context.Context,
	cluster *entity.Cluster,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(
		ctx,
		"DELETE FROM clusters WHERE cluster_id=$1",
		cluster.ClusterID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterRepository) DeleteClusterByName(
	ctx context.Context,
	cluster *entity.Cluster,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(
		ctx,
		"DELETE FROM clusters WHERE name=$1",
		cluster.Name,
	)
	if err != nil {
		return err
	}

	return nil
}
