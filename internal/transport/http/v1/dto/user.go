package dto

import "north-post/service/internal/domain/v1/models"

type SignInAdminUserByIdRequest struct {
	Uid string `json:"uid" binding:"required"`
}

type SignInAdminUserByIdResponse struct {
	Data AdminUserDTO `json:"data"`
}

type AdminUserDTO struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	LastLogin   int64  `json:"lastLogin"`
	ImageUrl    string `json:"imageUrl,omitempty"`
}

func ToAdminUserDTO(adminUser models.AdminUser) AdminUserDTO {
	return AdminUserDTO{
		Email:       adminUser.Email,
		DisplayName: adminUser.DisplayName,
		LastLogin:   adminUser.LastLogin,
		ImageUrl:    adminUser.ImageUrl,
	}
}
