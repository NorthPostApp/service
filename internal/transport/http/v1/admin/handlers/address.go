package handlers

import (
	"log/slog"
	"net/http"
	"north-post/service/internal/transport/http/v1/admin/dto"
	"north-post/service/internal/transport/http/v1/admin/services"

	"github.com/gin-gonic/gin"
)

type AddressHandler struct {
	service *services.AddressService
	logger  *slog.Logger
}

func NewAddressHandler(service *services.AddressService, logger *slog.Logger) *AddressHandler {
	return &AddressHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AddressHandler) GetAddresses(c *gin.Context) {
	var req dto.GetAllAddressesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	if err := req.Language.Validate(); err != nil {
		h.logger.Warn("invalid language", "language", req.Language, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.GetAddressesInput{
		Language: req.Language.Get(),
		Tags:     req.Tags,
		Limit:    req.Limit,
	}

	output, err := h.service.GetAddresses(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to get addresses", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch address"})
		return
	}

	response := dto.GetAllAddressResponse{
		Data:  dto.ToAddressDTOs(output.Addresses),
		Count: output.Count,
	}
	c.JSON(http.StatusOK, response)
}
