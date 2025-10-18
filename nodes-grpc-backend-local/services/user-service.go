package services

import (
	"context"
	"fmt"
	"nodes-grpc-backend-local/entity"
	"nodes-grpc-backend-local/model"
	"nodes-grpc-backend-local/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(userRepository *repository.UserRepository) *UserService {
	return &UserService{
		userRepository,
	}
}

func (s *UserService) GetAllUser(ctx context.Context) *model.DataPass {
	dataPass := &model.DataPass{
		Data:    nil,
		Message: "",
		Code:    0,
		IsError: false,
		Error:   nil,
	}

	users, err := s.userRepository.GetAllUser(ctx)
	if err != nil {
		dataPass.IsError = true
		dataPass.Message = "could not get all user"
		dataPass.Error = err
	}

	dataPass.Data = users
	dataPass.Message = "get all user success"

	return dataPass
}

func (s *UserService) AddUser(ctx context.Context, user *model.AddUser) error {
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return err
	}

	userEntity := &entity.User{
		UserId:   uuid.New(),
		Name:     user.Name,
		Email:    user.Email,
		Password: hashedPassword,
		GroupID:  user.GroupId,
		Role:     "user",
	}

	err = s.userRepository.AddUser(ctx, userEntity)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) Login(ctx context.Context, user *model.LoginUser) (*entity.User, error) {
	userEntity := &entity.User{
		Email:    user.Email,
		Password: user.Password,
	}

	userGet, err := s.userRepository.GetUserByEmail(ctx, userEntity)
	if err != nil {
		return userGet, err
	}

	// no user is found or password is wrong
	if userGet == nil || !checkPassword(userEntity.Password, userGet.Password) {
		return userGet, fmt.Errorf("email or password is wrong")
	}

	return userGet, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)

	return string(bytes), err
}

func checkPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return false
	}

	return true
}
