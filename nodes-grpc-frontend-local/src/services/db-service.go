package services

import (
	"context"
	"nodes-grpc-frontend-local/src/model/virtualization_model"

	"github.com/jackc/pgx/v5"
)

type DatabaseService struct {
	connection *pgx.Conn
}

func NewDatabaseService(conn *pgx.Conn) *DatabaseService {
	return &DatabaseService{
		connection: conn,
	}
}

func (s *DatabaseService) SaveCluster(
	ctx context.Context,
	clusterRequest virtualization_model.CreateClusterRequest,
) error {
	tx, err := s.connection.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	_, err = tx.Exec(
		ctx,
		"INSERT INTO clusters (name) VALUES ($1)",
		clusterRequest.Name,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) DeleteCluster(
	ctx context.Context,
	clusterRequest virtualization_model.CreateClusterRequest,
) error {
	tx, err := s.connection.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	_, err = tx.Exec(
		ctx,
		"DELETE FROM clusters WHERE name = '($1)'",
		clusterRequest.Name,
	)
	if err != nil {
		return err
	}

	return nil
}
