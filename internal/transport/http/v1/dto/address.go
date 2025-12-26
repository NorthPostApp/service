package dto

import (
	"north-post/service/internal/domain/v1/models"
)

type GetAllAddressesRequest struct {
	Language models.Language `json:"language" binding:"required"`
	Tags     []string        `json:"tags,omitempty"`
	Limit    int             `json:"limit,omitempty"`
}

type GetAllAddressResponse struct {
	Data  []AddressItemDTO `json:"data"`
	Count int              `json:"count"`
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

func FromCreateAddressDTO(req CreateAddressRequest) models.AddressItem {
	return models.AddressItem{
		Name:       req.Name,
		BriefIntro: req.BriefIntro,
		Tags:       req.Tags,
		Address: models.Address{
			Country:      req.Address.Country,
			City:         req.Address.City,
			Line1:        req.Address.Line1,
			Line2:        req.Address.Line2,
			BuildingName: req.Address.BuildingName,
			PostalCode:   req.Address.PostalCode,
			Region:       req.Address.Region,
		},
	}
}
