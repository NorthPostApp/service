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
	Data  []AddressDTO `json:"data"`
	Count int          `json:"count"`
}

type AddressDTO struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	BriefIntro string   `json:"briefIntro"`
	Tags       []string `json:"tags" firestore:"tags"`
}

func ToAddressDTO(address models.AddressItem) AddressDTO {
	return AddressDTO{
		ID:         address.ID,
		Name:       address.Name,
		BriefIntro: address.BriefIntro,
		Tags:       address.Tags,
	}
}

func ToAddressDTOs(addresses []models.AddressItem) []AddressDTO {
	output := make([]AddressDTO, len(addresses))
	for i, address := range addresses {
		output[i] = ToAddressDTO(address)
	}
	return output
}
