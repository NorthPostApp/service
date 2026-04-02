package services

import (
	"context"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
)

type userRepository interface {
	SignInAdminUserById(ctx context.Context, opts repository.GetUserByIdOptions) (*models.AdminUser, error)
	AuthenticateAppUserById(ctx context.Context, opts repository.GetUserByIdOptions) (*models.AppUser, error)
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

type AuthenticateAppUserByIdInput struct {
	Uid string
}

type AuthenticateAppUserByIdOutput struct {
	UserData models.AppUser
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
		return nil, err
	}
	return &SignInAdminUserByIdOutput{UserData: *adminUserData}, nil
}

func (s *UserService) AuthenticateAppUserById(
	ctx context.Context,
	input AuthenticateAppUserByIdInput) (*AuthenticateAppUserByIdOutput, error) {
	opts := repository.GetUserByIdOptions{Uid: input.Uid}
	appUserData, err := s.repo.AuthenticateAppUserById(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &AuthenticateAppUserByIdOutput{UserData: *appUserData}, nil
}
