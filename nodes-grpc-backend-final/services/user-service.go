package services

import (
	"context"
	"fmt"
	"nodes-grpc-be/entities"
	"nodes-grpc-be/models"
	"nodes-grpc-be/repositories"
)

type UserService struct {
	userRepository *repositories.UserRepository
	groupResitory  *repositories.GroupRepository
	jwtService     *JwtService
}

func NewUserService(
	userRepository *repositories.UserRepository,
	groupResitory *repositories.GroupRepository,
	jwtService *JwtService,
) *UserService {
	return &UserService{
		userRepository,
		groupResitory,
		jwtService,
	}
}

func (s *UserService) Login(ctx context.Context, user *models.LoginUser) (string, error) {
	userEntity := &entities.User{
		Email:    user.Email,
		Password: user.Password,
	}

	userGet, err := s.userRepository.GetUserByEmail(ctx, userEntity)
	if err != nil {
		return "", err
	}

	// no user is found or password is wrong
	if userGet == nil || userEntity.Password != userGet.Password {
		fmt.Println("here")
		return "", fmt.Errorf("email or password is wrong")
	}

	res := s.jwtService.GenerateToken(userGet.UserId)

	return res, nil
}

func (s *UserService) Me(ctx context.Context, token string) (*models.MeUserReturn, error) {
	userId, err := s.jwtService.GetUserIDByToken(token)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepository.GetUserById(ctx, &entities.User{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	group, err := s.groupResitory.GetGroupById(ctx, user.GroupID)
	if err != nil {
		return nil, err
	}

	return &models.MeUserReturn{
		UserId:         userId,
		Name:           user.Name,
		GroupId:        user.GroupID,
		Group:          group.Name,
		Vcpu:           group.Vcpu,
		Memory:         group.Memory,
		Storage:        group.Storage,
		NodeSize:       group.NodeSize,
		CurrentCluster: group.CurrentCluster,
		MaxCluster:     group.MaxCluster,
	}, nil
}
