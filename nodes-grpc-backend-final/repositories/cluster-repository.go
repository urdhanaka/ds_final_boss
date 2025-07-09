package repositories

import (
	"context"
	"database/sql"
	"errors"
	"nodes-grpc-be/entities"

	"github.com/jackc/pgx/v5"
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
		"INSERT INTO clusters (cluster_id, cluster_name, user_id, group_id, cluster_status, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		cluster.ClusterId, cluster.ClusterName, cluster.UserId, cluster.GroupId, cluster.ClusterStatus, cluster.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterRepository) GetClusterFromUserId(
	ctx context.Context,
	user *entities.User,
) ([]entities.Cluster, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT * FROM clusters WHERE user_id=$1", user.UserId)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[entities.Cluster])
}

func (r *ClusterRepository) GetClusterFromClusterId(
	ctx context.Context,
	cluster *entities.Cluster,
) (*entities.Cluster, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	err = conn.QueryRow(
		ctx,
		"SELECT cluster_id, cluster_name, user_id, group_id, cluster_status, ip_address, access_token, created_at FROM clusters WHERE cluster_id=$1",
		cluster.ClusterId,
	).Scan(
		&cluster.ClusterId,
		&cluster.ClusterName,
		&cluster.UserId,
		&cluster.GroupId,
		&cluster.ClusterStatus,
		&cluster.IpAddress,
		&cluster.AccessToken,
		&cluster.CreatedAt,
	)
	if err != nil {
		// skip if no row is found
		if !errors.Is(err, sql.ErrNoRows) {
			return cluster, err
		}

		return nil, err
	}

	return cluster, err
}

func (r *ClusterRepository) GetKubeconfigFromClusterId(
	ctx context.Context,
	cluster *entities.Cluster,
) (*entities.Cluster, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	err = conn.QueryRow(
		ctx,
		"SELECT kubeconfig_contents FROM clusters WHERE cluster_id=$1",
		cluster.ClusterId,
	).Scan(
		&cluster.KubeconfigContents,
	)
	if err != nil {
		// skip if no row is found
		if !errors.Is(err, sql.ErrNoRows) {
			return cluster, err
		}

		return nil, err
	}

	return cluster, err
}

func (r *ClusterRepository) UpdateClusterStatusByClusterId(
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
		`UPDATE clusters
        SET cluster_status=$1
        WHERE cluster_id=$2`,
		cluster.ClusterStatus, cluster.ClusterId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterRepository) UpdateIpAddressAndTokenByClusterId(
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
		`UPDATE clusters
        SET access_token=$1, ip_address=$2, kubeconfig_contents=$3
        WHERE cluster_id=$4`,
		cluster.AccessToken, cluster.IpAddress, cluster.KubeconfigContents, cluster.ClusterId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterRepository) DeleteClusterByClusterId(
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
		`DELETE FROM clusters
        WHERE cluster_id=$1`,
		cluster.ClusterId,
	)
	if err != nil {
		return err
	}

	return nil
}
