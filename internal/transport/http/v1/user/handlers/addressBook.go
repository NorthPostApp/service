package handlers

import (
	"log/slog"
	"net/http"
	"north-post/service/internal/repository"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/middleware"
	"north-post/service/internal/transport/http/v1/utils"
	"strings"

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
// @Param request body dto.UpdateUserSavedAddressRequest true "Address ID and action (add/remove)"
// @Produce json
// @Success 200 {object} dto.UpdateUserSavedAddressesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/address-book [post]
func (h *AddressBookHandler) UpdateSavedAddresses(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	if !validateUser(c, uid, h.logger) {
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
	output, err := h.userRepo.UpdateUserSavedAddresses(c.Request.Context(), opts)
	if err != nil {
		h.logger.Error("failed to update user saved address", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update user saved addresses"})
		return
	}
	response := dto.UpdateUserSavedAddressesResponse{Data: output}
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
