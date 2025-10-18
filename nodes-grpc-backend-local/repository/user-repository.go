package repository

import (
	"context"
	"database/sql"
	"errors"
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

func (r *UserRepository) GetUserByName(
	ctx context.Context,
	user *entity.User,
) (*entity.User, error) {
	getUser := new(entity.User)

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return user, err
	}

	err = conn.QueryRow(ctx, "SELECT id, name, group_id, is_admin FROM users WHERE name=$1",
		user.Name,
	).Scan(getUser)
	if err != nil {
		// skip if no row is found
		if !errors.Is(err, sql.ErrNoRows) {
			return getUser, err
		}
	}

	return getUser, nil
}

func (r *UserRepository) GetUserByEmail(
	ctx context.Context,
	user *entity.User,
) (*entity.User, error) {
	getUser := new(entity.User)

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return getUser, err
	}

	err = conn.QueryRow(ctx, "SELECT user_id, name, email, password, group_id, is_admin FROM users WHERE email=$1",
		user.Name,
	).Scan(getUser)
	if err != nil {
		// skip if no row is found
		if !errors.Is(err, sql.ErrNoRows) {
			return getUser, err
		}
	}

	return getUser, nil
}

func (r *UserRepository) GetAllUser(ctx context.Context) ([]entity.User, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(ctx, "SELECT * FROM users")
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.User])
}

func (r *UserRepository) AddUser(
	ctx context.Context, user *entity.User,
) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(
		ctx,
		"INSERT INTO user (user_id, name, email, password, group_id, role) VALUES ($1, $2, $3, $4, $5, $6)",
		user.UserId, user.Name, user.Email, user.Password, user.GroupID, user.Role,
	)
	if err != nil {
		return err
	}

	return nil
}
