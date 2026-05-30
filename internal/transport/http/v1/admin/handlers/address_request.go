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
	"strings"

	"github.com/gin-gonic/gin"
)

type addressRequestRepository interface {
	GetRequestsByStatus(
		ctx context.Context,
		opts *repository.GetRequestsByStatusOptions) (*repository.GetRequestsByStatusResponse, error)
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
