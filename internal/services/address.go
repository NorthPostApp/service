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
	GetAddressById(context.Context, repository.GetAddressByIdOptions) (*models.AddressItem, error)
	CreateNewAddress(context.Context, repository.CreateNewAddressOption) (string, error)
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

type CreateNewAddressInput struct {
	Language models.Language
	Address  models.AddressItem
}

type CreateNewAddressOutput struct {
	ID string
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
		return nil, fmt.Errorf("%w", err)
	}
	return &GetAddressesOutput{Addresses: addresses, Count: len(addresses)}, nil
}

func (s *AddressService) GetAddressById(ctx context.Context, input GetAddressByIdInput) (*GetAddressByIdOutput, error) {
	opts := repository.GetAddressByIdOptions{
		Language: input.Language,
		ID:       input.ID,
	}
	address, err := s.repo.GetAddressById(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &GetAddressByIdOutput{Address: *address}, nil
}

func (s *AddressService) CreateNewAddress(ctx context.Context, input CreateNewAddressInput) (*CreateNewAddressOutput, error) {
	opts := repository.CreateNewAddressOption{
		Language:    input.Language,
		AddressItem: input.Address,
	}
	id, err := s.repo.CreateNewAddress(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &CreateNewAddressOutput{ID: id}, nil
}
