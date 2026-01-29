package services

import (
	"context"
	"testing"

	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/infra"
	"north-post/service/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository implementations
type mockAddressRepository struct {
	mock.Mock
}

func (m *mockAddressRepository) GetAllAddresses(ctx context.Context, opts repository.GetAllAddressesOptions) (*repository.GetAllAddressesResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.GetAllAddressesResponse), args.Error(1)
}

func (m *mockAddressRepository) GetAddressById(ctx context.Context, opts repository.GetAddressByIdOptions) (*models.AddressItem, error) {
	args := m.Called(ctx, opts)
	var address *models.AddressItem
	if value := args.Get(0); value != nil {
		address, _ = value.(*models.AddressItem)
	}
	return address, args.Error(1)
}

func (m *mockAddressRepository) CreateNewAddress(ctx context.Context, opts repository.CreateNewAddressOption) (string, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func (m *mockAddressRepository) UpdateAddress(ctx context.Context, opts repository.UpdateAddressOption) (*models.AddressItem, error) {
	args := m.Called(ctx, opts)
	var address *models.AddressItem
	if value := args.Get(0); value != nil {
		address, _ = value.(*models.AddressItem)
	}
	return address, args.Error(1)
}

func (m *mockAddressRepository) DeleteAddress(ctx context.Context, opts repository.DeleteAddressOption) (string, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

type mockLLMClient struct {
	mock.Mock
}

func (m *mockLLMClient) StructuredCompletion(ctx context.Context, opts infra.StructuredCompletionOptions, schemaInstance interface{}, result interface{}) error {
	args := m.Called(ctx, opts, schemaInstance, result)
	if args.Get(0) != nil && result != nil {
		if out, ok := result.(*models.BatchAddressGenerationSchema); ok {
			*out = args.Get(0).(models.BatchAddressGenerationSchema)
		}
	}
	return args.Error(1)
}

func setupAddressService() (*AddressService, *mockAddressRepository, *mockLLMClient) {
	repo := new(mockAddressRepository)
	llm := new(mockLLMClient)
	service := NewAddressService(repo, llm)
	return service, repo, llm
}

// Tests
func TestAddressService_GetAddresses(t *testing.T) {
	t.Parallel()
	service, repo, _ := setupAddressService()
	input := GetAllAddressesInput{
		Language: models.LanguageEN,
		Tags:     []string{"featured", "outdoor"},
		PageSize: 0,
	}
	expectedOptions := repository.GetAllAddressesOptions{
		Language: input.Language,
		Tags:     input.Tags,
		PageSize: defaultPageSize,
	}
	expectedAddresses := []models.AddressItem{
		{ID: "1", Name: "Address One"},
		{ID: "2", Name: "Address Two"},
	}
	expectedResponse := &repository.GetAllAddressesResponse{
		Addresses:  expectedAddresses,
		HasMore:    false,
		TotalCount: 2,
	}
	repo.On(
		"GetAllAddresses",
		mock.Anything,
		mock.MatchedBy(func(opts repository.GetAllAddressesOptions) bool {
			if opts.Language != expectedOptions.Language || opts.PageSize != expectedOptions.PageSize {
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
	).Return(expectedResponse, nil).Once()
	output, err := service.GetAllAddresses(context.Background(), input)
	repo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, expectedResponse.Addresses, output.Addresses)
	assert.Equal(t, expectedResponse.TotalCount, output.TotalCount)
	assert.Equal(t, expectedResponse.HasMore, output.HasMore)
}

func TestAddressService_GetAddresses_Error(t *testing.T) {
	t.Parallel()
	service, repo, _ := setupAddressService()
	input := GetAllAddressesInput{
		Language: models.LanguageZH,
		Tags:     []string{"cafe"},
		PageSize: defaultPageSize,
	}
	repo.On(
		"GetAllAddresses",
		mock.Anything,
		mock.Anything,
	).Return(nil, assert.AnError).Once()
	output, err := service.GetAllAddresses(context.Background(), input)
	repo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, output)
}

func TestAddressService_GetAddresses_PageLimit(t *testing.T) {
	t.Parallel()
	service, repo, _ := setupAddressService()
	input := GetAllAddressesInput{
		Language: models.LanguageZH,
		Tags:     []string{""},
		PageSize: defaultPageSize + 5,
	}
	expectedOptions := repository.GetAllAddressesOptions{
		Language: input.Language,
		Tags:     input.Tags,
		PageSize: defaultPageSize,
	}
	repo.On(
		"GetAllAddresses",
		mock.Anything,
		mock.MatchedBy(func(opts repository.GetAllAddressesOptions) bool {
			if opts.Language != expectedOptions.Language {
				return false
			}
			if opts.PageSize != expectedOptions.PageSize {
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
	service.GetAllAddresses(context.Background(), input)
}

func TestAddressService_GetAddressById(t *testing.T) {
	t.Parallel()
	service, repo, _ := setupAddressService()
	expectedAddress := models.AddressItem{ID: "2", Name: "Address Two"}
	input := GetAddressByIdInput{
		Language: models.LanguageZH,
		ID:       "",
	}
	expectedOptions := repository.GetAddressByIdOptions{
		Language: input.Language,
		ID:       input.ID,
	}
	repo.On("GetAddressById",
		mock.Anything,
		mock.MatchedBy(func(opts repository.GetAddressByIdOptions) bool {
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
	service, repo, _ := setupAddressService()
	input := GetAddressByIdInput{
		Language: models.LanguageZH,
		ID:       "",
	}
	expectedOptions := repository.GetAddressByIdOptions{
		Language: input.Language,
		ID:       input.ID,
	}
	repo.On("GetAddressById",
		mock.Anything,
		mock.MatchedBy(func(opts repository.GetAddressByIdOptions) bool {
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
}

func TestAddressService_CreateNewAddress(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	service, repo, _ := setupAddressService()
	expectedID := "expected_id"
	input := CreateNewAddressInput{
		Language: models.LanguageZH,
		Address:  models.AddressItem{Name: "test", ID: "test"},
	}
	repo.On("CreateNewAddress",
		mock.Anything,
		mock.Anything,
	).Return(expectedID, nil).Once()
	output, err := service.CreateNewAddress(context.Background(), input)
	repo.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, output.ID, expectedID)
}

func TestAddressService_CreateNewAddress_Error(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	service, repo, _ := setupAddressService()
	input := CreateNewAddressInput{
		Language: models.LanguageZH,
		Address:  models.AddressItem{Name: "test", ID: "test"},
	}
	repo.On("CreateNewAddress",
		mock.Anything,
		mock.Anything,
	).Return(nil, assert.AnError).Once()
	output, err := service.CreateNewAddress(context.Background(), input)
	repo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, output)
}

func TestAddressService_GenerateNewAddress_EmptyPrompt(t *testing.T) {
	t.Parallel()
	service, _, _ := setupAddressService()
	input := GenerateAddressInput{
		SystemPrompt:    "sys",
		Prompt:          "",
		Model:           "gpt-5-mini",
		ReasoningEffort: "minimum",
	}
	output, err := service.GenerateNewAddress(context.Background(), input)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "prompt cannot be empty")
}

func TestAddressService_GenerateNewAddress(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	llm := new(mockLLMClient)
	service := NewAddressService(repo, llm)
	input := GenerateAddressInput{
		SystemPrompt:    "sys",
		Prompt:          "generate an address",
		Model:           "gpt-5-mini",
		ReasoningEffort: "minimum",
	}
	expectedBatch := models.BatchAddressGenerationSchema{
		Addresses: []models.AddressGenerationSchema{
			{Name: "test"},
		},
	}
	llm.On(
		"StructuredCompletion",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(expectedBatch, nil).Once()
	output, err := service.GenerateNewAddress(context.Background(), input)
	llm.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, 1, len(output.Addresses))
	assert.Equal(t, "test", output.Addresses[0].Name)
}

func TestAddressService_GenerateNewAddress_Error(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	llm := new(mockLLMClient)
	service := NewAddressService(repo, llm)
	input := GenerateAddressInput{
		SystemPrompt:    "sys",
		Prompt:          "generate an address",
		Model:           "gpt-5-mini",
		ReasoningEffort: "minimum",
	}
	llm.On(
		"StructuredCompletion",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, assert.AnError).Once()
	output, err := service.GenerateNewAddress(context.Background(), input)
	llm.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "failed to generate address")
}

func TestAddressService_UpdateAddress(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	llm := new(mockLLMClient)
	addressItem := models.AddressItem{ID: "123", Name: "Test", BriefIntro: "Brief introduction"}
	service := NewAddressService(repo, llm)
	input := UpdateAddressInput{
		Language: "EN",
		ID:       "123",
		Address:  models.AddressItem{ID: "123", Name: "Test", BriefIntro: "Brief introduction"},
	}
	repo.On("UpdateAddress", mock.Anything, mock.Anything).Return(&addressItem, nil).Once()
	output, err := service.UpdateAddress(context.Background(), input)
	assert.Nil(t, err)
	assert.NotNil(t, output)
}

func TestAddressService_UpdateAddress_Error(t *testing.T) {
	t.Parallel()
	repo := new(mockAddressRepository)
	llm := new(mockLLMClient)
	service := NewAddressService(repo, llm)
	input := UpdateAddressInput{
		Language: "EN",
		ID:       "123",
		Address:  models.AddressItem{ID: "123", Name: "Test", BriefIntro: "Brief introduction"},
	}
	repo.On("UpdateAddress", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()
	output, err := service.UpdateAddress(context.Background(), input)
	assert.Nil(t, output)
	assert.Error(t, err)
}
