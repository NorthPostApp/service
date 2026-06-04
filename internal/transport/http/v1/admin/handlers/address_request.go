package handlers

import (
	"context"
	"fmt"
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

type addressRequestRepository interface {
	GetRequestsByStatus(
		ctx context.Context,
		opts *repository.GetRequestsByStatusOptions) (*repository.GetRequestsByStatusResponse, error)
	UpdateRequestData(
		ctx context.Context, opts *repository.UpdateRequestDataOptions) error
}

type AddressRequestHandler struct {
	repo   addressRequestRepository
	logger *slog.Logger
}

func NewAddressRequestHandler(
	addressRequestRepo addressRequestRepository,
	logger *slog.Logger) *AddressRequestHandler {
	return &AddressRequestHandler{
		repo:   addressRequestRepo,
		logger: logger,
	}
}

// GetRequestsByStatus godoc
// @Summary Get address requests by status
// @Description Get address requests for the current language by request status
// @Tags Admin User
// @Param Authorization header string true "Bearer idToken"
// @Param language query string true "Language code (e.g., en, zh)"
// @Param status query string true "Address request status (pending, processing, completed, failed)"
// @Produce json
// @Success 200 {object} dto.GetRequestsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/address-request [get]
func (h *AddressRequestHandler) GetRequestsByStatus(c *gin.Context) {
	logger := h.logger.With("path", "admin.handlers.address_request.GetRequestsByStatus")
	language := models.Language(c.GetString(middleware.LanguageKey))
	status := models.AddressRequestStatus(strings.TrimSpace(c.Query("status")))
	if !status.IsValid() {
		logger.Error("invalid request status", "status", string(status))
		c.JSON(
			http.StatusBadRequest,
			dto.ErrorResponse{Error: fmt.Sprintf("invalid request status: %s", string(status))})
		return
	}
	opts := &repository.GetRequestsByStatusOptions{
		Language: models.Language(language),
		Status:   models.AddressRequestStatus(status),
	}
	output, err := h.repo.GetRequestsByStatus(c.Request.Context(), opts)
	if err != nil {
		logger.Error("failed to get requests by status", "error", err)
		c.JSON(
			http.StatusInternalServerError,
			dto.ErrorResponse{Error: "failed to get requests by status"})
		return
	}
	c.JSON(http.StatusOK, dto.GetRequestsResponse{Data: output.Requests})
}

// UpdateRequest godoc
// @Summary Update an address request
// @Description Update an address request for the current language
// @Tags Admin User
// @Param Authorization header string true "Bearer idToken"
// @Accept json
// @Produce json
// @Param request body dto.UpdateRequest true "Request body"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/address-request/update [post]
func (h *AddressRequestHandler) UpdateRequest(c *gin.Context) {
	logger := h.logger.With("path", "admin.handlers.address_request.UpdateRequest")
	language := models.Language(c.GetString(middleware.LanguageKey))
	var req dto.UpdateRequest
	if !utils.BindJSON(c, &req, logger) {
		return
	}
	if !req.UpdatedRequest.Status.IsValid() {
		logger.Error("invalid request status",
			"status", req.UpdatedRequest.Status,
		)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request status"})
		return
	}
	opts := &repository.UpdateRequestDataOptions{
		Language:       language,
		ID:             req.ID,
		UpdatedRequest: req.UpdatedRequest,
	}
	if err := h.repo.UpdateRequestData(c.Request.Context(), opts); err != nil {
		logger.Error(
			"failed to update request data",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "succeeded"})
}
