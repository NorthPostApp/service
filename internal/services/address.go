package services

import (
	"context"
	"fmt"

	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
)

const defaultPageSize = 100

type addressRepository interface {
	GetAllAddresses(context.Context, repository.GetAllAddressesOptions) ([]models.AddressItem, error)
	GetAddressById(context.Context, repository.GetAddressByIdOption) (*models.AddressItem, error)
}

type AddressService struct {
	repo addressRepository
}

func NewAddressService(repo addressRepository) *AddressService {
	return &AddressService{
		repo: repo,
	}
}

type GetAddressesInput struct {
	Language models.Language
	Tags     []string
	Limit    int
}

type GetAddressesOutput struct {
	Addresses []models.AddressItem
	Count     int
}

type GetAddressByIdInput struct {
	Language models.Language
	ID       string
}

type GetAddressByIdOutput struct {
	Address models.AddressItem
}

func (s *AddressService) GetAddresses(ctx context.Context, input GetAddressesInput) (*GetAddressesOutput, error) {
	limit := input.Limit
	if limit <= 0 || limit > defaultPageSize {
		limit = defaultPageSize
	}
	opts := repository.GetAllAddressesOptions{
		Language: input.Language,
		Tags:     input.Tags,
		Limit:    limit,
	}
	addresses, err := s.repo.GetAllAddresses(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}
	return &GetAddressesOutput{Addresses: addresses, Count: len(addresses)}, nil
}

func (s *AddressService) GetAddressById(ctx context.Context, input GetAddressByIdInput) (*GetAddressByIdOutput, error) {
	opts := repository.GetAddressByIdOption{
		Language: input.Language,
		ID:       input.ID,
	}
	address, err := s.repo.GetAddressById(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}
	return &GetAddressByIdOutput{Address: *address}, nil
}
