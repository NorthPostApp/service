package services

import (
	"context"
	"errors"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) SignInAdminUserById(
	ctx context.Context,
	opts repository.GetUserByIdOptions) (*models.AdminUser, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AdminUser), args.Error(1)
}

func setupUserService() (*UserService, *mockUserRepository) {
	repo := new(mockUserRepository)
	service := NewUserService(repo)
	return service, repo
}

// Tests
func TestUserService_SignInAdminUserById(t *testing.T) {
	t.Parallel()
	service, repo := setupUserService()
	ctx := context.Background()
	input := SignInAdminUserByIdInput{
		Uid: "test-uid-123",
	}
	expectedUser := &models.AdminUser{
		Email:       "admin@example.com",
		DisplayName: "Admin User",
		LastLogin:   1234567890,
		ImageUrl:    "https://example.com/image.jpg",
	}
	repo.On(
		"SignInAdminUserById",
		mock.Anything,
		mock.Anything,
	).Return(expectedUser, nil).Once()
	output, err := service.SignInAdminUserById(ctx, input)
	repo.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser.Email, output.UserData.Email)
	assert.Equal(t, expectedUser.DisplayName, output.UserData.DisplayName)
	assert.Equal(t, expectedUser.LastLogin, output.UserData.LastLogin)
	assert.Equal(t, expectedUser.ImageUrl, output.UserData.ImageUrl)
}

func TestUserService_SignInAdminUserById_Error(t *testing.T) {
	t.Parallel()
	service, repo := setupUserService()
	ctx := context.Background()
	input := SignInAdminUserByIdInput{
		Uid: "test-uid-123",
	}
	repo.On(
		"SignInAdminUserById",
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("user not found")).Once()
	output, err := service.SignInAdminUserById(ctx, input)
	repo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, output)
}
