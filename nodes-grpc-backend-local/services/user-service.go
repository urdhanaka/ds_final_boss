package services

import (
	"context"
	"nodes-grpc-backend-local/entity"
	"nodes-grpc-backend-local/repository"
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(userRepository *repository.UserRepository) *UserService {
	return &UserService{
		userRepository,
	}
}

func (s *UserService) GetAllUser(ctx context.Context) ([]entity.User, error) {
	users, err := s.userRepository.GetAllUser(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}
