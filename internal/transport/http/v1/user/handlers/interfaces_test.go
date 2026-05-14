package handlers

import (
	"context"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
	"north-post/service/internal/transport/http/v1/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// --------- Mock User Repo ----------
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) AuthenticateAppUserById(
	ctx context.Context,
	opts repository.GetUserByIdOptions) (*models.AppUser, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AppUser), args.Error(1)
}

func (m *mockUserRepo) UpdateUserSavedAddresses(
	ctx context.Context,
	opts *repository.UpdateUserSavedAddressesOptions,
) (string, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *mockUserRepo) GetUserSavedAddresses(
	ctx context.Context,
	opts *repository.GetUserSavedAddressesOptions,
) ([]string, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// --------- Mock Address Repo ----------
type mockAddressRepo struct {
	mock.Mock
}

func (m *mockAddressRepo) GetAddressesByIDs(
	ctx context.Context,
	opts *repository.GetAddressesByIDsOptions,
) (*repository.GetAddressesByIDsResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.GetAddressesByIDsResponse), args.Error(1)
}

// --------- Mock Utils ----------
func mockAuthMiddleware(uid string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if uid != "" {
			c.Set(middleware.UidKey, uid)
		}
		c.Next()
	}
}

func mockLanguageMiddleware(language string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if language != "" {
			c.Set(middleware.LanguageKey, language)
		}
		c.Next()
	}
}
