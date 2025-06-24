package repositories

import (
	"context"
	"database/sql"
	"errors"
	"nodes-grpc-be/entities"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	dbPool *pgxpool.Pool
}

func NewUserRepository(dbPool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		dbPool: dbPool,
	}
}

func (r *UserRepository) GetUserByEmail(
	ctx context.Context,
	user *entities.User,
) (*entities.User, error) {
	getUser := new(entities.User)

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return getUser, err
	}
    defer conn.Release()

	err = conn.QueryRow(ctx, "SELECT user_id, name, email, password, group_id FROM users WHERE email=$1",
		user.Email,
	).Scan(
		&getUser.UserId,
		&getUser.Name,
		&getUser.Email,
		&getUser.Password,
		&getUser.GroupID,
	)
	if err != nil {
		// skip if no row is found
		if !errors.Is(err, sql.ErrNoRows) {
			return getUser, err
		}
	}

	return getUser, nil
}

func (r *UserRepository) GetUserById(
	ctx context.Context,
	user *entities.User,
) (*entities.User, error) {
	getUser := new(entities.User)

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return getUser, err
	}
    defer conn.Release()

	err = conn.QueryRow(ctx, "SELECT user_id, name, email, password, group_id FROM users WHERE user_id=$1",
		user.UserId,
	).Scan(
		&getUser.UserId,
		&getUser.Name,
		&getUser.Email,
		&getUser.Password,
		&getUser.GroupID,
	)
	if err != nil {
		// skip if no row is found
		if !errors.Is(err, sql.ErrNoRows) {
			return getUser, err
		}
	}

	return getUser, nil
}
