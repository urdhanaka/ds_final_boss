package repositories

import (
	"context"
	"fmt"
	"nodes-grpc-be/entities"

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

func (r *GroupRepository) GetGroupByName(
	ctx context.Context,
	groupName string,
) (*entities.Group, error) {
	group := new(entities.Group)

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return group, err
	}
    defer conn.Release()

	err = conn.QueryRow(
		ctx,
		"SELECT * FROM groups WHERE name=$1",
		groupName,
	).Scan(
		&group.GroupId,
		&group.Name,
		&group.Vcpu,
		&group.Memory,
		&group.Storage,
		&group.NodeSize,
        &group.CurrentCluster,
		&group.MaxCluster,
	)
	if err != nil {
        fmt.Println("here", err)
		return group, err
	}

	return group, nil
}

func (r *GroupRepository) GetGroupById(
	ctx context.Context,
	groupId int,
) (*entities.Group, error) {
	group := new(entities.Group)

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return group, err
	}
    defer conn.Release()

	err = conn.QueryRow(
		ctx,
		"SELECT * FROM groups WHERE group_id=$1",
		groupId,
	).Scan(
		&group.GroupId,
		&group.Name,
		&group.Vcpu,
		&group.Memory,
		&group.Storage,
		&group.NodeSize,
		&group.CurrentCluster,
		&group.MaxCluster,
	)
	if err != nil {
		return group, err
	}

	return group, nil
}
