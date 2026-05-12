package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/middleware"
	"north-post/service/internal/transport/http/v1/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type userRepository interface {
	AuthenticateAppUserById(
		ctx context.Context,
		opts repository.GetUserByIdOptions) (*models.AppUser, error)
	UpdateUserSavedAddresses(
		ctx context.Context,
		opts *repository.UpdateUserSavedAddressesOptions,
	) (string, error)
}

type UserHandler struct {
	repo   userRepository
	logger *slog.Logger
}

func NewUserHandler(repository userRepository, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		repo:   repository,
		logger: logger,
	}
}

// AuthenticateAppUser godoc
// @Summary Authenticate app user
// @Description Authenticate an app user with the idToken and update last login timestamp
// @Tags User
// @Param Authorization header string true "Bearer idToken"
// @Produce json
// @Success 200 {object} dto.AuthenticateAppUserResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/signin [post]
func (h *UserHandler) AuthenticateAppUser(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	if !h.validateUser(c, uid) {
		return
	}
	opts := repository.GetUserByIdOptions{Uid: uid}
	output, err := h.repo.AuthenticateAppUserById(c.Request.Context(), opts)
	if err != nil {
		h.logger.Error("failed to authenticate app user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate app user"})
		return
	}
	response := dto.AuthenticateAppUserResponse{
		Data: dto.ToAppUserDTO(output),
	}
	c.JSON(http.StatusOK, response)
}

// UpdateSavedAddresses godoc
// @Summary Update user saved addresses
// @Description Add or remove a saved address for the authenticated user
// @Tags User
// @Param Authorization header string true "Bearer idToken"
// @Param request body dto.UpdateUserSavedAddressRequest true "Address ID and action (add/remove)"
// @Produce json
// @Success 200 {object} dto.UpdateUserSavedAddressesResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /address/book [post]
func (h *UserHandler) UpdateSavedAddresses(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	if !h.validateUser(c, uid) {
		return
	}
	var req dto.UpdateUserSavedAddressRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}
	opts := &repository.UpdateUserSavedAddressesOptions{
		UserID:    uid,
		AddressID: req.AddressId,
		Action:    h.convertUpdateMethod(req.Action),
	}
	output, err := h.repo.UpdateUserSavedAddresses(c.Request.Context(), opts)
	if err != nil {
		h.logger.Error("failed to update user saved address", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user saved addresses"})
		return
	}
	response := dto.UpdateUserSavedAddressesResponse{Data: output}
	c.JSON(http.StatusOK, response)
}

// Helper functions
func (h *UserHandler) validateUser(c *gin.Context, uid string) bool {
	if uid == "" {
		h.logger.Error(
			"missing user id from the middleware context",
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized id token"})
		return false
	}
	return true
}

func (h *UserHandler) convertUpdateMethod(action string) repository.UpdateSavedAddressesAction {
	action = strings.ToLower(action)
	switch action {
	case "add":
		return repository.Add
	default:
		return repository.Delete
	}
}
