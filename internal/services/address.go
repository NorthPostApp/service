package services

import (
	"context"
	"fmt"

	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/infra"
	"north-post/service/internal/repository"

	"github.com/openai/openai-go/v3"
)

const defaultPageSize = 100

type addressRepository interface {
	GetAllAddresses(context.Context, repository.GetAllAddressesOptions) ([]models.AddressItem, error)
	GetAddressById(context.Context, repository.GetAddressByIdOptions) (*models.AddressItem, error)
	CreateNewAddress(context.Context, repository.CreateNewAddressOption) (string, error)
}

type AddressService struct {
	repo addressRepository
	llm  *infra.LLMClient
}

func NewAddressService(repo addressRepository, llm *infra.LLMClient) *AddressService {
	return &AddressService{
		repo: repo,
		llm:  llm,
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

type GenerateAddressInput struct {
	SystemPrompt    string
	Prompt          string
	Language        models.Language
	Model           openai.ChatModel
	ReasoningEffort openai.ReasoningEffort
	// Temperature *float64 // might be useful in the future
}

type GenerateAddressOutput struct {
	Addresses []models.AddressItem
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

func (s *AddressService) GenerateNewAddress(ctx context.Context, input GenerateAddressInput) (*GenerateAddressOutput, error) {
	// validate input
	if input.Prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}
	// configure structured completion options
	opts := infra.StructuredCompletionOptions{
		SystemPrompt:    input.SystemPrompt,
		Prompt:          input.Prompt,
		SchemaName:      "address_generation",
		Description:     "Generate a structured address with metadata",
		Model:           input.Model,
		ReasoningEffort: input.ReasoningEffort,
		// may insert temperature field here in the future
	}
	schema := models.BatchAddressGenerationSchema{}
	result, err := infra.StructuredCompletion(ctx, s.llm, opts, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate address: %w", err)
	}
	addresses := []models.AddressItem{}
	for _, address := range result.Addresses {
		addressItem := models.AddressItem{
			Name:       address.Name,
			BriefIntro: address.BriefIntro,
			Tags:       address.Tags,
			Address:    address.Address,
		}
		addresses = append(addresses, addressItem)
	}
	return &GenerateAddressOutput{Addresses: addresses}, nil
}
