package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/v3"
)

type addressService interface {
	CreateNewAddress(ctx context.Context, input services.CreateNewAddressInput) (*services.CreateNewAddressOutput, error)
	GenerateNewAddress(ctx context.Context, input services.GenerateAddressInput) (*services.GenerateAddressOutput, error)
	GetAddressById(ctx context.Context, input services.GetAddressByIdInput) (*services.GetAddressByIdOutput, error)
	GetAllAddresses(ctx context.Context, input services.GetAllAddressesInput) (*services.GetAllAddressesOutput, error)
	UpdateAddress(ctx context.Context, input services.UpdateAddressInput) (*services.UpdateAddressOutput, error)
	DeleteAddress(ctx context.Context, input services.DeleteAddressInput) (*services.DeleteAddressOutput, error)
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
func (h *AddressHandler) GetAllAddresses(c *gin.Context) {
	var req dto.GetAllAddressesRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}
	if !utils.ValidateLanguage(c, req.Language, h.logger) {
		return
	}
	input := services.GetAllAddressesInput{
		Language:      req.Language,
		Tags:          req.Tags,
		PageSize:      req.PageSize,
		StartAfterDoc: req.LastDocID,
	}
	output, err := h.service.GetAllAddresses(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to get addresses", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := dto.GetAllAddressResponse{
		Data: dto.ToGetAllAddressesResponseDTO(output, req.Language),
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
	if strings.TrimSpace(id) == "" {
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

// UpdateAddress updates an existing address entry with language-specific information.
//
// @Summary Update an existing address
// @Description Update an existing address entry with language-specific information
// @Tags Admin Address
// @Accept json
// @Produce json
// @Param request body dto.UpdateAddressRequest true "Request body"
// @Success 200 {object} dto.UpdateAddressResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/address [post]
func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	var req dto.UpdateAddressRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}
	if !utils.ValidateLanguage(c, req.Language, h.logger) {
		return
	}
	input := services.UpdateAddressInput{
		Language: req.Language,
		ID:       req.ID,
		Address:  dto.FromUpdateAddressDTO(req),
	}
	output, err := h.service.UpdateAddress(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to update address", "address", req, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := dto.UpdateAddressResponse{
		Data: dto.ToAddressDTO(output.Address),
	}
	c.JSON(http.StatusOK, response)
}

// DeleteAddress godoc
// @Summary Delete an address
// @Description Delete an existing address entry by ID and language
// @Tags Admin Address
// @Accept json
// @Produce json
// @Param id path string true "Address ID"
// @Param language query string true "Language code (e.g., en, zh)"
// @Success 200 {object} dto.DeleteAddressResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/address/{id} [delete]
func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
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
	input := services.DeleteAddressInput{
		Language: language,
		ID:       id,
	}
	output, err := h.service.DeleteAddress(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to delete address", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := dto.DeleteAddressResponse{Data: dto.AddressID{ID: output.ID}}
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
