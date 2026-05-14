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
	"time"

	"github.com/gin-gonic/gin"
)

type AddressBookHandler struct {
	userRepo    userRepository
	addressRepo addressRepository
	logger      *slog.Logger
}

func NewAddressBookHandler(
	userRepo userRepository,
	addressRepo addressRepository,
	logger *slog.Logger) *AddressBookHandler {
	return &AddressBookHandler{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		logger:      logger,
	}
}

// UpdateSavedAddresses godoc
// @Summary Update user saved addresses
// @Description Add or remove a saved address for the authenticated user
// @Tags App User
// @Param Authorization header string true "Bearer idToken"
// @Param request body dto.UpdateUserSavedAddressesRequest true "Address ID and action (add/remove)"
// @Produce json
// @Success 200 {object} dto.UpdateUserSavedAddressesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/address-book [patch]
func (h *AddressBookHandler) UpdateSavedAddresses(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	language := models.Language(c.GetString(middleware.LanguageKey))
	if !validateUser(c, uid, h.logger) {
		return
	}
	var req dto.UpdateUserSavedAddressesRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}
	opts := &repository.UpdateUserSavedAddressesOptions{
		UserID:     uid,
		Language:   language,
		AddressIDs: req.AddressIDs,
		Action:     h.convertUpdateMethod(req.Action),
	}
	output, err := h.userRepo.UpdateUserSavedAddresses(c.Request.Context(), opts)
	if err != nil {
		h.logger.Error("failed to update user saved address", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update user saved addresses"})
		return
	}
	response := dto.UpdateUserSavedAddressesResponse{Data: output}
	c.JSON(http.StatusOK, response)
}

// GetSavedAddresses godoc
// @Summary Get user saved addresses
// @Description Retrieve all saved addresses for the authenticated user
// @Tags App User
// @Param Authorization header string true "Bearer idToken"
// @Param language query string true "Language code (e.g., en, zh)"
// @Produce json
// @Success 200 {object} dto.GetSavedAddressesResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/address-book [get]
func (h *AddressBookHandler) GetSavedAddresses(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	language := models.Language(c.GetString(middleware.LanguageKey))
	if !validateUser(c, uid, h.logger) {
		return
	}
	getSavedAddressesOpts := &repository.GetUserSavedAddressesOptions{Uid: uid, Language: language}
	addressIDs, err := h.userRepo.GetUserSavedAddresses(c.Request.Context(), getSavedAddressesOpts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	getAddressesOpts := &repository.GetAddressesByIDsOptions{
		Language: language,
		IDs:      addressIDs,
	}
	results, err := h.addressRepo.GetAddressesByIDs(
		c.Request.Context(),
		getAddressesOpts,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	// if invalid ids is not empty, remove the invalid items in the background process
	if len(results.InvalidIDs) > 0 {
		h.removeInvalidIDsInBackground(uid, language, results.InvalidIDs)
	}
	response := dto.GetSavedAddressesResponse{Data: dto.ToAddressDTOs(results.Addresses)}
	c.JSON(http.StatusOK, response)
}

// Helper functions
func (h *AddressBookHandler) convertUpdateMethod(action string) repository.UpdateSavedAddressesAction {
	action = strings.ToLower(action)
	switch action {
	case "add":
		return repository.Add
	case "delete", "remove":
		return repository.Delete
	default:
		return -1
	}
}

// ---------- Helper methods ----------
func (h *AddressBookHandler) removeInvalidIDsInBackground(uid string, language models.Language, invalidIDs []string) {
	go func(invalidIDs []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := h.userRepo.UpdateUserSavedAddresses(
			ctx,
			&repository.UpdateUserSavedAddressesOptions{
				Language:   language,
				UserID:     uid,
				AddressIDs: invalidIDs,
				Action:     repository.Delete,
			},
		)
		if err != nil {
			h.logger.Error("Failed to remove invalid saved addresses",
				"error", err,
				"invalidIDs", invalidIDs,
			)
		}
	}(invalidIDs)
}
