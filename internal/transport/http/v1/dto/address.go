package dto

import (
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/services"
)

type AddressID struct {
	ID string `json:"id"`
}

type GetAllAddressesRequest struct {
	Language  models.Language `json:"language" binding:"required"`
	Tags      []string        `json:"tags,omitempty"`
	PageSize  int             `json:"pageSize,omitempty"`
	LastDocID string          `json:"lastDocId,omitempty"`
}

type GetAllAddressResponse struct {
	Data GetAllAddressesResponseDTO `json:"data"`
}

type GetAddressByIdResponse struct {
	Data AddressItemDTO `json:"data"`
}

type CreateAddressRequest struct {
	Language   models.Language `json:"language" binding:"required"`
	Name       string          `json:"name" binding:"required"`
	BriefIntro string          `json:"briefIntro" binding:"required"`
	Tags       []string        `json:"tags" binding:"required"`
	Address    AddressDTO      `json:"address" binding:"required"`
}

type CreateAddressResponse struct {
	ID string `json:"id"`
}

type UpdateAddressRequest struct {
	Language models.Language `json:"language" binding:"required"`
	ID       string          `json:"id" binding:"required"`
	Address  AddressItemDTO  `json:"address" binding:"required"`
}

type UpdateAddressResponse struct {
	Data AddressItemDTO `json:"data"`
}

type GetTagsResponse struct {
	Data models.TagsRecord `json:"data"`
}

type DeleteAddressResponse struct {
	Data AddressID `json:"data"`
}

type GenerateNewAddressRequest struct {
	Language        models.Language `json:"language" binding:"required"`
	Prompt          string          `json:"prompt" binding:"required"`
	SystemPrompt    string          `json:"systemPrompt,omitempty"`
	Model           string          `json:"model,omitempty"`
	ReasoningEffort string          `json:"reasoningEffort,omitempty"`
}

type GenerateNewAddressResponse struct {
	Data []AddressItemDTO `json:"data"`
}

type AddressItemDTO struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	BriefIntro string     `json:"briefIntro"`
	Tags       []string   `json:"tags"`
	CreatedAt  int64      `json:"createdAt"`
	UpdatedAt  int64      `json:"updatedAt"`
	Address    AddressDTO `json:"address"`
}

type AddressDTO struct {
	City         string `json:"city" binding:"required"`
	Country      string `json:"country" binding:"required"`
	Line1        string `json:"line1" binding:"required"`
	Line2        string `json:"line2,omitempty"`
	BuildingName string `json:"buildingName,omitempty"`
	PostalCode   string `json:"postalCode,omitempty"`
	Region       string `json:"region" binding:"required"`
}

type GetAllAddressesResponseDTO struct {
	Addresses  []AddressItemDTO `json:"addresses"`
	TotalCount int64            `json:"totalCount"`
	LastDocID  string           `json:"lastDocId"`
	HasMore    bool             `json:"hasMore"`
	Language   models.Language  `json:"language"`
}

func FromAddressDTO(address AddressDTO) models.Address {
	return models.Address{
		Country:      address.Country,
		City:         address.City,
		Line1:        address.Line1,
		Line2:        address.Line2,
		BuildingName: address.BuildingName,
		PostalCode:   address.PostalCode,
		Region:       address.Region,
	}
}

func ToAddressDTO(addressItem models.AddressItem) AddressItemDTO {
	address := addressItem.Address
	addressDto := AddressDTO{
		City:         address.City,
		Country:      address.Country,
		Line1:        address.Line1,
		Line2:        address.Line2,
		BuildingName: address.BuildingName,
		PostalCode:   address.PostalCode,
		Region:       address.Region,
	}
	return AddressItemDTO{
		ID:         addressItem.ID,
		Name:       addressItem.Name,
		BriefIntro: addressItem.BriefIntro,
		Tags:       addressItem.Tags,
		CreatedAt:  addressItem.CreatedAt,
		UpdatedAt:  addressItem.UpdatedAt,
		Address:    addressDto,
	}
}

func ToAddressDTOs(addresses []models.AddressItem) []AddressItemDTO {
	output := make([]AddressItemDTO, len(addresses))
	for i, addressItem := range addresses {
		output[i] = ToAddressDTO(addressItem)
	}
	return output
}

func ToGetAllAddressesResponseDTO(output *services.GetAllAddressesOutput, language models.Language) GetAllAddressesResponseDTO {
	return GetAllAddressesResponseDTO{
		Addresses:  ToAddressDTOs(output.Addresses),
		TotalCount: output.TotalCount,
		LastDocID:  output.LastDocID,
		HasMore:    output.HasMore,
		Language:   language,
	}
}

func FromCreateAddressDTO(req CreateAddressRequest) models.AddressItem {
	return models.AddressItem{
		Name:       req.Name,
		BriefIntro: req.BriefIntro,
		Tags:       req.Tags,
		Address:    FromAddressDTO(req.Address),
	}
}

func FromUpdateAddressDTO(req UpdateAddressRequest) models.AddressItem {
	addressItem := req.Address
	return models.AddressItem{
		ID:         addressItem.ID,
		Name:       addressItem.Name,
		BriefIntro: addressItem.BriefIntro,
		Tags:       addressItem.Tags,
		CreatedAt:  addressItem.CreatedAt,
		UpdatedAt:  addressItem.UpdatedAt,
		Address:    FromAddressDTO(addressItem.Address),
	}
}
