package handlers

import (
	"context"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
)

type userRepository interface {
	AuthenticateAppUserById(
		ctx context.Context,
		opts *repository.GetUserByIdOptions) (*models.AppUser, error)
	UpdateUserSavedAddresses(
		ctx context.Context,
		opts *repository.UpdateUserSavedAddressesOptions,
	) (string, error)
	GetUserSavedAddresses(
		ctx context.Context,
		opts *repository.GetUserSavedAddressesOptions,
	) ([]string, error)
	UpdateUserAddressRequests(
		ctx context.Context,
		opts *repository.UpdateUserAddressRequestsOptions,
	) (string, error)
	GetUserRequests(
		ctx context.Context,
		opts *repository.GetUserRequestsOptions,
	) (*models.UserAddressRequests, error)
	UpdateUserActiveRequestCount(
		ctx context.Context, opts *repository.UpdateUserActiveRequestCountOptions,
	) error
}

type addressRequestRepository interface {
	CreateNewRequestWithLimit(
		ctx context.Context,
		opts *repository.CreateRequestOptions) (string, error)
	DeleteRequests(
		ctx context.Context,
		opts *repository.DeleteRequestsOptions) error
	GetRequestsByIDs(
		ctx context.Context,
		opts *repository.GetRequestsByIDsOptions) (
		*repository.GetRequestsByIDsResponse, error,
	)
}

type addressRepository interface {
	GetAddressesByIDs(
		ctx context.Context,
		opts *repository.GetAddressesByIDsOptions,
	) (*repository.GetAddressesByIDsResponse, error)
}
