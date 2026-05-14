package handlers

import (
	"context"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
)

type userRepository interface {
	AuthenticateAppUserById(
		ctx context.Context,
		opts repository.GetUserByIdOptions) (*models.AppUser, error)
	UpdateUserSavedAddresses(
		ctx context.Context,
		opts *repository.UpdateUserSavedAddressesOptions,
	) (string, error)
	GetUserSavedAddresses(
		ctx context.Context,
		opts *repository.GetUserSavedAddressesOptions,
	) ([]string, error)
}

type addressRepository interface {
	GetAddressesByIDs(
		ctx context.Context,
		opts *repository.GetAddressesByIDsOptions,
	) (*repository.GetAddressesByIDsResponse, error)
}
