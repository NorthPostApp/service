package dto

import "north-post/service/internal/domain/v1/models"

type UpdateUserSavedAddressesRequest struct {
	Language   models.Language `json:"language"`
	AddressIDs []string        `json:"addressIDs"`
	Action     string          `json:"action"`
}

type UpdateUserSavedAddressesResponse struct {
	Data string `json:"data"`
}

type GetSavedAddressesResponse struct {
	Data []AddressItemDTO `json:"data"`
}
