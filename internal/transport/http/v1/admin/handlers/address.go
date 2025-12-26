package handlers

import (
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/utils"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/v3"
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

// GetAddresses godoc
// @Summary Get all addresses
// @Description Get all addresses by language and optional tag filters
// @Tags Admin Address
// @Accept json
// @Produce json
// @Param request body dto.GetAllAddressesRequest true "Request body"
// @Success 200 {object} dto.GetAllAddressResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/address [post]
func (h *AddressHandler) GetAddresses(c *gin.Context) {
	var req dto.GetAllAddressesRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}
	if !utils.ValidateLanguage(c, req.Language, h.logger) {
		return
	}
	input := services.GetAddressesInput{
		Language: req.Language,
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

// GetAddressById godoc
// @Summary Get address by ID
// @Description Get a single address by ID with language query parameter
// @Tags Admin Address
// @Accept json
// @Produce json
// @Param id path string true "Address ID"
// @Param language query string true "Language code (e.g., en, zh)"
// @Success 200 {object} dto.GetAddressByIdResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/address/{id} [get]
func (h *AddressHandler) GetAddressById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.Warn("missing address id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Address ID is required"})
		return
	}
	languageStr := c.Query("language")
	if languageStr == "" {
		h.logger.Warn("missing language parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Language is required"})
		return
	}
	language := models.Language(languageStr)
	if !utils.ValidateLanguage(c, language, h.logger) {
		return
	}
	input := services.GetAddressByIdInput{
		Language: language,
		ID:       id,
	}
	output, err := h.service.GetAddressById(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to get address", "addressId", input.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := dto.GetAddressByIdResponse{
		Data: dto.ToAddressDTO(output.Address),
	}
	c.JSON(http.StatusOK, response)
}

// CreateNewAddress godoc
// @Summary Create a new address
// @Description Create a new address entry with language-specific information
// @Tags Admin Address
// @Accept json
// @Produce json
// @Param request body dto.CreateAddressRequest true "Request body"
// @Success 200 {object} dto.CreateAddressResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/address [put]
func (h *AddressHandler) CreateNewAddress(c *gin.Context) {
	var req dto.CreateAddressRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}
	if !utils.ValidateLanguage(c, req.Language, h.logger) {
		return
	}
	input := services.CreateNewAddressInput{
		Language: req.Language,
		Address:  dto.FromCreateAddressDTO(req),
	}
	output, err := h.service.CreateNewAddress(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to create new address", "address", req, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := dto.CreateAddressResponse{
		ID: output.ID,
	}
	c.JSON(http.StatusOK, response)
}

// GenerateNewAddress godoc
// @Summary Generate new address suggestions
// @Description Uses LLM to generate new address suggestions based on prompts and reasoning effort
// @Tags Admin Address
// @Accept json
// @Produce json
// @Param request body dto.GenerateNewAddressRequest true "Request body"
// @Success 200 {object} dto.GenerateNewAddressResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/address/generate [post]
func (h *AddressHandler) GenerateNewAddress(c *gin.Context) {
	var req dto.GenerateNewAddressRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}
	if !utils.ValidateLanguage(c, req.Language, h.logger) {
		return
	}
	input := services.GenerateAddressInput{
		Language:        req.Language,
		SystemPrompt:    req.SystemPrompt,
		Prompt:          req.Prompt,
		Model:           req.Model,
		ReasoningEffort: openai.ReasoningEffort(req.ReasoningEffort),
	}
	output, err := h.service.GenerateNewAddress(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to generate new address", "request", req, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := dto.GenerateNewAddressResponse{
		Data: dto.ToAddressDTOs(output.Addresses),
	}
	c.JSON(http.StatusOK, response)
}
