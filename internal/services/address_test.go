// internal/services/address_test.go
package services

import (
	"context"
	"testing"

	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository implementations
type mockAddressRepository struct {
	mock.Mock
}

func (m *mockAddressRepository) GetAllAddresses(ctx context.Context, opts repository.GetAllAddressesOptions) ([]models.AddressItem, error) {
	args := m.Called(ctx, opts)
	var addresses []models.AddressItem
	if value := args.Get(0); value != nil {
		addresses, _ = value.([]models.AddressItem)
	}
	return addresses, args.Error(1)
}

func (m *mockAddressRepository) GetAddressById(ctx context.Context, opts repository.GetAddressByIdOption) (*models.AddressItem, error) {
	args := m.Called(ctx, opts)
	var address *models.AddressItem
	if value := args.Get(0); value != nil {
		address, _ = value.(*models.AddressItem)
	}
	return address, args.Error(1)
}

// Tests
func TestAddressService_GetAddresses(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	service := NewAddressService(repo)
	input := GetAddressesInput{
		Language: models.LanguageEN,
		Tags:     []string{"featured", "outdoor"},
		Limit:    0,
	}
	expectedOptions := repository.GetAllAddressesOptions{
		Language: input.Language,
		Tags:     input.Tags,
		Limit:    defaultPageSize,
	}
	expectedAddresses := []models.AddressItem{
		{ID: "1", Name: "Address One"},
		{ID: "2", Name: "Address Two"},
	}
	repo.On(
		"GetAllAddresses",
		mock.Anything,
		mock.MatchedBy(func(opts repository.GetAllAddressesOptions) bool {
			if opts.Language != expectedOptions.Language || opts.Limit != expectedOptions.Limit {
				return false
			}
			if len(opts.Tags) != len(expectedOptions.Tags) {
				return false
			}
			for i := range opts.Tags {
				if opts.Tags[i] != expectedOptions.Tags[i] {
					return false
				}
			}
			return true
		}),
	).Return(expectedAddresses, nil).Once()
	output, err := service.GetAddresses(context.Background(), input)
	repo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, expectedAddresses, output.Addresses)
	assert.Equal(t, len(expectedAddresses), output.Count)
}

func TestAddressService_GetAddresses_Error(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	service := NewAddressService(repo)
	input := GetAddressesInput{
		Language: models.LanguageZH,
		Tags:     []string{"cafe"},
		Limit:    defaultPageSize + 5,
	}
	expectedOptions := repository.GetAllAddressesOptions{
		Language: input.Language,
		Tags:     input.Tags,
		Limit:    defaultPageSize,
	}
	repo.On(
		"GetAllAddresses",
		mock.Anything,
		mock.MatchedBy(func(opts repository.GetAllAddressesOptions) bool {
			if opts.Language != expectedOptions.Language || opts.Limit != expectedOptions.Limit {
				return false
			}
			if len(opts.Tags) != len(expectedOptions.Tags) {
				return false
			}
			for i := range opts.Tags {
				if opts.Tags[i] != expectedOptions.Tags[i] {
					return false
				}
			}
			return true
		}),
	).Return(nil, assert.AnError).Once()
	output, err := service.GetAddresses(context.Background(), input)
	repo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.ErrorContains(t, err, "failed to get addresses")
}

func TestAddressService_GetAddresses_PageLimit(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	service := NewAddressService(repo)
	input := GetAddressesInput{
		Language: models.LanguageZH,
		Tags:     []string{""},
		Limit:    5,
	}
	expectedOptions := repository.GetAllAddressesOptions{
		Language: input.Language,
		Tags:     input.Tags,
		Limit:    defaultPageSize,
	}
	repo.On(
		"GetAllAddresses",
		mock.Anything,
		mock.MatchedBy(func(opts repository.GetAllAddressesOptions) bool {
			if opts.Language != expectedOptions.Language {
				return false
			}
			if opts.Limit == expectedOptions.Limit {
				return false
			}
			if len(opts.Tags) != len(expectedOptions.Tags) {
				return false
			}
			for i := range opts.Tags {
				if opts.Tags[i] != expectedOptions.Tags[i] {
					return false
				}
			}
			return true
		}),
	).Return(nil, assert.AnError).Once()
	service.GetAddresses(context.Background(), input)
}

func TestAddressService_GetAddressById(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	service := NewAddressService(repo)
	expectedAddress := models.AddressItem{ID: "2", Name: "Address Two"}
	input := GetAddressByIdInput{
		Language: models.LanguageZH,
		ID:       "",
	}
	expectedOptions := repository.GetAddressByIdOption{
		Language: input.Language,
		ID:       input.ID,
	}
	repo.On("GetAddressById",
		mock.Anything,
		mock.MatchedBy(func(opts repository.GetAddressByIdOption) bool {
			if opts.Language != expectedOptions.Language {
				return false
			}
			if opts.ID != expectedOptions.ID {
				return false
			}
			return true
		}),
	).Return(&expectedAddress, nil).Once()
	output, err := service.GetAddressById(context.Background(), input)
	repo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, expectedAddress, output.Address)
}

func TestAddressService_GetAddressById_Error(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	service := NewAddressService(repo)
	input := GetAddressByIdInput{
		Language: models.LanguageZH,
		ID:       "",
	}
	expectedOptions := repository.GetAddressByIdOption{
		Language: input.Language,
		ID:       input.ID,
	}
	repo.On("GetAddressById",
		mock.Anything,
		mock.MatchedBy(func(opts repository.GetAddressByIdOption) bool {
			if opts.Language != expectedOptions.Language {
				return false
			}
			if opts.ID != expectedOptions.ID {
				return false
			}
			return true
		}),
	).Return(nil, assert.AnError).Once()
	output, err := service.GetAddressById(context.Background(), input)
	repo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.ErrorContains(t, err, "failed to get address")
}
