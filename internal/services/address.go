package services

import (
	"context"
	"fmt"

	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/infra"
	"north-post/service/internal/repository"

	"github.com/google/uuid"
)

const defaultPageSize = 100

type addressRepository interface {
	GetAllAddresses(context.Context, repository.GetAllAddressesOptions) (
		*repository.GetAllAddressesResponse, error)
	GetAddressById(context.Context, repository.GetAddressByIdOptions) (*models.AddressItem, error)
	CreateNewAddress(context.Context, repository.CreateNewAddressOption) (string, error)
	UpdateAddress(context.Context, repository.UpdateAddressOption) (*models.AddressItem, error)
	DeleteAddress(context.Context, repository.DeleteAddressOption) (string, error)
	RefreshTags(context.Context, repository.RefreshTagsOption) (*models.TagsRecord, error)
	GetAllTags(context.Context, repository.GetAllTagsOption) (*models.TagsRecord, error)
	SyncToTypesense(context.Context, repository.SyncToTypesenseOption) (*repository.SyncToTypesenseResult, error)
}

type llmClient interface {
	StructuredCompletion(
		context.Context,
		infra.StructuredCompletionOptions,
		interface{},
		interface{}) error
}

type AddressService struct {
	repo addressRepository
	llm  llmClient
}

func NewAddressService(repo addressRepository, llm llmClient) *AddressService {
	return &AddressService{
		repo: repo,
		llm:  llm,
	}
}

type GetAllAddressesInput struct {
	Language      models.Language
	Tags          []string
	PageSize      int
	StartAfterDoc string
}

type GetAllAddressesOutput struct {
	Addresses  []models.AddressItem
	TotalCount int64
	LastDocID  string
	HasMore    bool
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

type UpdateAddressInput struct {
	Language models.Language
	ID       string
	Address  models.AddressItem
}

type UpdateAddressOutput struct {
	Address models.AddressItem
}

type DeleteAddressInput struct {
	Language models.Language
	ID       string
}

type DeleteAddressOutput struct {
	ID string
}

type GenerateAddressInput struct {
	SystemPrompt    string
	Prompt          string
	Language        models.Language
	Model           string
	ReasoningEffort string
	ThinkingLevel   string
}

type GenerateAddressOutput struct {
	Addresses []models.AddressItem
}

type RefreshTagsInput struct {
	Language models.Language
}

type RefreshTagsOutput struct {
	TagsRecord models.TagsRecord
}

type GetAllTagsInput struct {
	Language models.Language
}

type GetAllTagsOutput struct {
	TagsRecord models.TagsRecord
}

type SyncToTypesenseInput struct {
	Language models.Language
}

type SyncToTypesenseOutput struct {
	Total   int
	Success int
	Failed  int
}

func (s *AddressService) GetAllAddresses(
	ctx context.Context,
	input GetAllAddressesInput,
) (*GetAllAddressesOutput, error) {
	pageSize := input.PageSize
	if pageSize <= 0 || pageSize > defaultPageSize {
		pageSize = defaultPageSize
	}
	opts := repository.GetAllAddressesOptions{
		Language:      input.Language,
		Tags:          input.Tags,
		PageSize:      pageSize,
		StartAfterDoc: input.StartAfterDoc,
	}
	response, err := s.repo.GetAllAddresses(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &GetAllAddressesOutput{
			Addresses:  response.Addresses,
			TotalCount: response.TotalCount,
			LastDocID:  response.LastDocID,
			HasMore:    response.HasMore,
		},
		nil
}

func (s *AddressService) GetAddressById(ctx context.Context, input GetAddressByIdInput) (*GetAddressByIdOutput, error) {
	opts := repository.GetAddressByIdOptions{
		Language: input.Language,
		ID:       input.ID,
	}
	address, err := s.repo.GetAddressById(ctx, opts)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return &CreateNewAddressOutput{ID: id}, nil
}

func (s *AddressService) UpdateAddress(ctx context.Context, input UpdateAddressInput) (*UpdateAddressOutput, error) {
	opts := repository.UpdateAddressOption{
		Language:    input.Language,
		ID:          input.ID,
		AddressItem: input.Address,
	}
	addressItem, err := s.repo.UpdateAddress(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &UpdateAddressOutput{Address: *addressItem}, nil
}

func (s *AddressService) DeleteAddress(ctx context.Context, input DeleteAddressInput) (*DeleteAddressOutput, error) {
	opts := repository.DeleteAddressOption{
		Language: input.Language,
		ID:       input.ID,
	}
	deletedId, err := s.repo.DeleteAddress(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &DeleteAddressOutput{ID: deletedId}, nil
}

func (s *AddressService) GenerateNewAddress(ctx context.Context, input GenerateAddressInput) (*GenerateAddressOutput, error) {
	// validate input
	if input.Prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}
	// configure structured completion options
	opts := infra.StructuredCompletionOptions{
		Prompt:          input.Prompt,
		SystemPrompt:    input.SystemPrompt,
		Model:           input.Model,
		ReasoningEffort: input.ReasoningEffort,
		ThinkingLevel:   input.ThinkingLevel,
	}
	schema := models.BatchAddressGenerationSchema{}
	var result models.BatchAddressGenerationSchema
	err := s.llm.StructuredCompletion(ctx, opts, schema, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to generate address: %w", err)
	}
	addresses := []models.AddressItem{}
	for _, address := range result.Addresses {
		addressItem := models.AddressItem{
			ID:         uuid.NewString(),
			Name:       address.Name,
			BriefIntro: address.BriefIntro,
			Tags:       address.Tags,
			Address:    address.Address,
		}
		addresses = append(addresses, addressItem)
	}
	return &GenerateAddressOutput{Addresses: addresses}, nil
}

func (s *AddressService) RefreshTags(
	ctx context.Context,
	input RefreshTagsInput,
) (*RefreshTagsOutput, error) {
	opts := repository.RefreshTagsOption{Language: input.Language}
	record, err := s.repo.RefreshTags(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &RefreshTagsOutput{TagsRecord: *record}, nil
}

func (s *AddressService) GetAllTags(
	ctx context.Context,
	input GetAllTagsInput,
) (*GetAllTagsOutput, error) {
	opts := repository.GetAllTagsOption{Language: input.Language}
	record, err := s.repo.GetAllTags(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &GetAllTagsOutput{TagsRecord: *record}, nil
}

func (s *AddressService) SyncToTypesense(
	ctx context.Context,
	input SyncToTypesenseInput) (*SyncToTypesenseOutput, error) {
	opts := repository.SyncToTypesenseOption{Language: input.Language}
	result, err := s.repo.SyncToTypesense(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &SyncToTypesenseOutput{
		Total:   result.Total,
		Success: result.Success,
		Failed:  result.Failed,
	}, nil
}
