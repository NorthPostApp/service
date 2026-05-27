package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
	"north-post/service/internal/transport/http/v1/dto"

	"github.com/gin-gonic/gin"
)

type promptRepository interface {
	GetSystemAddressGenerationPrompt(
		ctx context.Context,
		opts *repository.GetSystemAddressGenerationPromptOptions) (string, error)
}

type PromptHandler struct {
	repo   promptRepository
	logger *slog.Logger
}

func NewPromptHandler(repo promptRepository, logger *slog.Logger) *PromptHandler {
	return &PromptHandler{
		repo:   repo,
		logger: logger,
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
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/prompt/system-address-generation [get]
func (h *PromptHandler) GetSystemAddressGenerationPrompt(c *gin.Context) {
	languageStr := c.Query("language")
	// we can skip language validation here because we've set fallback
	// language as en in the prompt repository
	opts := &repository.GetSystemAddressGenerationPromptOptions{
		Language: models.Language(languageStr),
	}
	prompt, err := h.repo.GetSystemAddressGenerationPrompt(c.Request.Context(), opts)
	if err != nil {
		h.logger.Error("failed to get system address generation prompt", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	response := dto.GetSystemAddressGenerationPromptResponse{
		Data: prompt,
	}
	c.JSON(http.StatusOK, response)
}
