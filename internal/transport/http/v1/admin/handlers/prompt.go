package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"

	"github.com/gin-gonic/gin"
)

type promptService interface {
	GetSystemAddressGenerationPrompt(ctx context.Context, input services.GetSystemAddressGenerationPromptInput) (*services.GetSystemAddressGenerationPromptOutput, error)
}

type PromptHandler struct {
	service promptService
	logger  *slog.Logger
}

func NewPromptHandler(service promptService, logger *slog.Logger) *PromptHandler {
	return &PromptHandler{
		service: service,
		logger:  logger,
	}
}

// GetSystemAddressGenerationPrompt godoc
// @Summary Get system prompt for address generation
// @Description Retrieves the system prompt used for address generation, optionally based on language.
// @Tags Admin Prompt
// @Accept json
// @Produce json
// @Param language query string false "Language code"
// @Success 200 {object} dto.GetSystemAddressGenerationPromptResponse
// @Failure 500 {object} map[string]string
// @Router /admin/prompt/system-address-generation [get]
func (h *PromptHandler) GetSystemAddressGenerationPrompt(c *gin.Context) {
	languageStr := c.Query("language")
	// we can skip language validation here because we've set fallback
	// language as en in the prompt repository
	opts := services.GetSystemAddressGenerationPromptInput{
		Language: models.Language(languageStr),
	}
	prompt, err := h.service.GetSystemAddressGenerationPrompt(c.Request.Context(), opts)
	if err != nil {
		h.logger.Error("failed to get system address generation prompt", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := dto.GetSystemAddressGenerationPromptResponse{
		Data: prompt.Prompt,
	}
	c.JSON(http.StatusOK, response)
}
