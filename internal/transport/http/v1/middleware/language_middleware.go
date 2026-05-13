package middleware

import (
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	LanguageKey = "language"
)

func LanguageFromQueryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		language := models.Language(c.Query("language"))
		if language == "" {
			logger.Error("language query is required")
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "language is required"})
			c.Abort()
			return
		}
		if !utils.ValidateLanguage(c, language, logger) {
			c.Abort()
			return
		}
		c.Set(LanguageKey, language.Get())
		c.Next()
	}
}

func LanguageFromBodyMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Language models.Language `json:"language"`
		}
		if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
			logger.Error("failed to bind language body", "error", err)
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
			c.Abort()
			return
		}
		if !utils.ValidateLanguage(c, req.Language, logger) {
			c.Abort()
			return
		}
		c.Set(LanguageKey, req.Language.Get())
		c.Next()
	}
}
