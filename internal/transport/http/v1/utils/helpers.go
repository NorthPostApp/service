package utils

import (
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"

	"github.com/gin-gonic/gin"
)

func BindJSON(c *gin.Context, req interface{}, logger *slog.Logger) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		logger.Error("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func ValidateLanguage(c *gin.Context, language models.Language, logger *slog.Logger) bool {
	if err := language.Validate(); err != nil {
		logger.Warn("invalid language", "language", language, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}
