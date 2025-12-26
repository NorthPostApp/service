package handlers

import (
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"

	"github.com/gin-gonic/gin"
)

type PromptHandler struct {
	prompt *services.PromptService
	logger *slog.Logger
}

func NewPromptHandler(prompt *services.PromptService, logger *slog.Logger) *PromptHandler {
	return &PromptHandler{
		prompt: prompt,
		logger: logger,
	}
}

func (h *PromptHandler) GetSystemAddressGenerationPrompt(c *gin.Context) {
	languageStr := c.Query("language")
	// we can skip language validation here because we've set fallback
	// language as en in the prompt repository
	opts := services.GetSystemAddressGenerationPromptInput{
		Language: models.Language(languageStr),
	}
	prompt, err := h.prompt.GetSystemAddressGenerationPrompt(c.Request.Context(), opts)
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
