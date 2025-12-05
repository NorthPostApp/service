package services

import (
	"context"
	"fmt"

	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
)

const defaultPageSize = 100

type AddressService struct {
	repo *repository.AddressRepository
}

func NewAddressService(repo *repository.AddressRepository) *AddressService {
	return &AddressService{
		repo: repo,
	}
}

type GetAddressesInput struct {
	Language string
	Tags     []string
	Limit    int
}

type GetAddressesOutput struct {
	Addresses []models.AddressItem
	Count     int
}

// GetAddresses godoc
// @Summary Get all addresses
// @Description Get all addresses by language and filtered optional tags
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body dto.GetAddressesRequest true "Request body"
// @Success 200 {object} dto.GetAddressesResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/addresses [post]
func (s *AddressService) GetAddresses(ctx context.Context, input GetAddressesInput) (*GetAddressesOutput, error) {
	limit := input.Limit
	if limit <= 0 || limit > defaultPageSize {
		limit = defaultPageSize
	}

	opts := repository.GetAddressesOptions{
		Language: input.Language,
		Tags:     input.Tags,
		Limit:    limit,
	}

	persons, err := s.repo.GetAll(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}
	return &GetAddressesOutput{Addresses: persons, Count: len(persons)}, nil
}
