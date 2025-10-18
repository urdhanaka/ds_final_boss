package repository

import (
	"context"
	"nodes-grpc-backend-local/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepository struct {
	dbPool *pgxpool.Pool
}

func NewGroupRepository(dbPool *pgxpool.Pool) *GroupRepository {
	return &GroupRepository{
		dbPool,
	}
}

func (r *GroupRepository) GetAllGroups(ctx context.Context) ([]entity.Group, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(ctx, "SELECT * FROM groups")
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.Group])
}

func (r *GroupRepository) AddGroup(ctx context.Context, group *entity.Group) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(
		ctx,
		"INSERT INTO groups () VALUES ($1, $2, $3, $4, $5, $6)",
		group.GroupId, group.Name, group.Vcpu,
		group.Ram, group.Storage, group.MaxCluster,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *GroupRepository) GetGroupByName(ctx context.Context, groupName string) (*entity.Group, error) {
	group := new(entity.Group)

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return group, err
	}

	err = conn.QueryRow(
		ctx,
		"SELECT * FROM groups WHERE name=$1",
		groupName,
	).Scan(
		&group.GroupId,
		&group.Name,
		&group.Vcpu,
		&group.Ram,
		&group.Storage,
		&group.NodeSize,
		&group.MaxCluster,
	)
	if err != nil {
		return group, err
	}

	return group, nil
}
