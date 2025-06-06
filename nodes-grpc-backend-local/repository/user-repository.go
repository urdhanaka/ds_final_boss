package repository

import (
	"context"
	"nodes-grpc-backend-local/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	dbPool *pgxpool.Pool
}

func NewUserRepository(dbPool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		dbPool,
	}
}

func (r *UserRepository) GetUserByName(ctx context.Context, name string) (*entity.User, error) {
	user := new(entity.User)

	err := r.dbPool.QueryRow(ctx, "SELECT id, name, group_id, is_admin FROM user WHERE name=$1", name).Scan(user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (r *UserRepository) GetAllUser(ctx context.Context) ([]entity.User, error) {
	rows, err := r.dbPool.Query(ctx, "SELECT * FROM user")
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.User])
}
