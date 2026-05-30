package handlers

import (
	"context"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"

	"github.com/stretchr/testify/mock"
)

// --------- Mock User Repo ----------
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) AuthenticateAppUserById(
	ctx context.Context,
	opts *repository.GetUserByIdOptions) (*models.AppUser, error) {
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

func (m *mockUserRepo) UpdateUserAddressRequests(
	ctx context.Context,
	opts *repository.UpdateUserAddressRequestsOptions,
) (string, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *mockUserRepo) GetUserRequests(
	ctx context.Context,
	opts *repository.GetUserRequestsOptions,
) (*models.UserAddressRequests, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserAddressRequests), args.Error(1)
}
func (m *mockUserRepo) UpdateUserActiveRequestCount(
	ctx context.Context, opts *repository.UpdateUserActiveRequestCountOptions,
) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
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

// --------- Mock Address Request Repo ----------
type mockAddressRequestRepo struct {
	mock.Mock
}

func (m *mockAddressRequestRepo) CreateNewRequestWithLimit(
	ctx context.Context,
	opts *repository.CreateRequestOptions) (string, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *mockAddressRequestRepo) DeleteRequests(
	ctx context.Context,
	opts *repository.DeleteRequestsOptions) error {
	args := m.Called(ctx, opts)
	return args.Error(1)
}

func (m *mockAddressRequestRepo) GetRequestsByIDs(
	ctx context.Context,
	opts *repository.GetRequestsByIDsOptions) (
	*repository.GetRequestsByIDsResponse, error,
) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.GetRequestsByIDsResponse), args.Error(1)
}
