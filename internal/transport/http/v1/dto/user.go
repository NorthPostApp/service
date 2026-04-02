package dto

import "north-post/service/internal/domain/v1/models"

type SignInAdminUserResponse struct {
	Data AdminUserDTO `json:"data"`
}

type AuthenticateAppUserResponse struct {
	Data AppUserDTO `json:"data"`
}

type AdminUserDTO struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	LastLogin   int64  `json:"lastLogin"`
	ImageUrl    string `json:"imageUrl,omitempty"`
}

type AppUserDTO struct {
	Email       string   `json:"email"`
	DisplayName string   `json:"displayName"`
	CreatedAt   int64    `json:"createdAt"`
	LastLogin   int64    `json:"lastLogin"`
	ImageUrl    string   `json:"imageUrl"`
	LikedMusics []string `json:"likedMusics"`
	Drafts      []string `json:"drafts"`
}

func ToAdminUserDTO(adminUser models.AdminUser) AdminUserDTO {
	return AdminUserDTO{
		Email:       adminUser.Email,
		DisplayName: adminUser.DisplayName,
		LastLogin:   adminUser.LastLogin,
		ImageUrl:    adminUser.ImageUrl,
	}
}

func ToAppUserDTO(appUser models.AppUser) AppUserDTO {
	return AppUserDTO{
		Email:       appUser.Email,
		DisplayName: appUser.DisplayName,
		CreatedAt:   appUser.CreatedAt,
		LastLogin:   appUser.LastLogin,
		ImageUrl:    appUser.ImageUrl,
		LikedMusics: appUser.LikedMusics,
		Drafts:      appUser.Drafts,
	}
}
