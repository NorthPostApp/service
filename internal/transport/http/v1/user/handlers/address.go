package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/middleware"
	"north-post/service/internal/transport/http/v1/utils"

	"github.com/gin-gonic/gin"
)

type addressService interface {
	GetAllTags(ctx context.Context, input services.GetAllTagsInput) (
		*services.GetAllTagsOutput, error,
	)
	GetAddresses(ctx context.Context, input services.GetAddressesInput) (
		*services.GetAddressesOutput, error,
	)
}

type AddressHandler struct {
	service addressService
	logger  *slog.Logger
}

func NewAddressHandler(service addressService, logger *slog.Logger) *AddressHandler {
	return &AddressHandler{
		service: service,
		logger:  logger,
	}
}

// GetAllTags godoc
// @Summary Get all address tags
// @Description Read and return address tags data by given language
// @Tags User Address
// @Accept json
// @Produce json
// @Param language query string true "Language code (e.g., en, zh)"
// @Success 200 {object} dto.GetAllTagsResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/address/tags [get]
func (h *AddressHandler) GetAllTags(c *gin.Context) {
	language := models.Language(c.GetString(middleware.LanguageKey))
	input := services.GetAllTagsInput{Language: language}
	output, err := h.service.GetAllTags(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to get all tags", "language", language, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get all tags"})
		return
	}
	response := dto.GetAllTagsResponse{Data: dto.ToTagsRecordDTO(output.TagsRecord, language)}
	c.JSON(http.StatusOK, response)
}

// GetAddresses godoc
// @Summary Get addresses
// @Description Search and return addresses by language, keywords, tags, and pagination
// @Tags User Address
// @Accept json
// @Produce json
// @Param request body dto.GetAddressesRequest true "Get addresses request"
// @Success 200 {object} dto.GetAddressesResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/address [post]
func (h *AddressHandler) GetAddresses(c *gin.Context) {
	var req dto.GetAddressesRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}
	input := services.GetAddressesInput{
		Language: req.Language,
		Keywords: req.Keywords,
		Tags:     req.Tags,
		PageSize: req.PageSize,
		Page:     req.Page,
	}
	output, err := h.service.GetAddresses(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to get addresses", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get addresses"})
		return
	}
	response := dto.GetAddressesResponse{
		Data: dto.ToGetAddressesResponseDTO(output, req.Language),
	}
	c.JSON(http.StatusOK, response)
}
