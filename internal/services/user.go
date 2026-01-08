package services

import (
	"context"
	"fmt"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
)

type userRepository interface {
	SignInAdminUserById(ctx context.Context, opts repository.GetUserByIdOptions) (*models.AdminUser, error)
}

type UserService struct {
	repo userRepository
}

type SignInAdminUserByIdInput struct {
	Uid string
}

type SignInAdminUserByIdOutput struct {
	UserData models.AdminUser
}

func NewUserService(repo userRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) SignInAdminUserById(
	ctx context.Context,
	input SignInAdminUserByIdInput) (*SignInAdminUserByIdOutput, error) {
	opts := repository.GetUserByIdOptions{Uid: input.Uid}
	adminUserData, err := s.repo.SignInAdminUserById(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &SignInAdminUserByIdOutput{UserData: *adminUserData}, nil
}
