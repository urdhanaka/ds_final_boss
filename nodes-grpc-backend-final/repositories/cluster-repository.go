package repositories

import (
	"context"
	"nodes-grpc-be/entities"

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
	cluster *entities.Cluster,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
    defer conn.Release()

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
