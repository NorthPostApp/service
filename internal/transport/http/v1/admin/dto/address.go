package dto

import (
	"north-post/service/internal/domain/v1/models"
)

type GetAllAddressesRequest struct {
	Language Language `json:"language" binding:"required"`
	Tags     []string `json:"tags,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

type GetAllAddressResponse struct {
	Data  []AddressItemDTO `json:"data"`
	Count int              `json:"count"`
}

type AddressItemDTO struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	BriefIntro string     `json:"briefIntro"`
	Tags       []string   `json:"tags"`
	Address    AddressDTO `json:"address"`
}

type AddressDTO struct {
	City         string `json:"city"`
	Country      string `json:"country"`
	Line1        string `json:"line1"`
	Line2        string `json:"line2,omitempty"`
	BuildingName string `json:"buildingName,omitempty"`
	PostalCode   string `json:"postalCode"`
	Region       string `json:"region"`
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
