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

type CreateNewRequest struct {
	Language models.Language `json:"language"`
	Content  string          `json:"content"`
}

type CreateNewRequestResponse struct {
	Data string `json:"data"`
}

type GetRequestsResponse struct {
	Data []models.AddressRequest `json:"data"`
}

type UpdateRequest struct {
	Language       models.Language       `json:"language"`
	ID             string                `json:"id"`
	UpdatedRequest models.AddressRequest `json:"updatedRequest"`
}
