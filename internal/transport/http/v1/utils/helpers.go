package utils

import (
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"
	"strconv"
	"strings"

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

func ValidateMusicFilename(c *gin.Context, genre string, track string, logger *slog.Logger) bool {
	if len(track) == 0 || len(genre) == 0 {
		logger.Error("invalid music filename", "track", track, "genre", genre)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid music genre or track"})
		return false
	}
	// reject path traversal and unexpected path separators in genre/track
	if strings.Contains(genre, "..") ||
		strings.Contains(track, "..") ||
		strings.ContainsAny(genre, `/\`) ||
		strings.ContainsAny(track, `/\`) {
		logger.Warn("possible path traversal attack", "ip", c.ClientIP())
		logger.Error("invalid characters in music filename", "genre", genre, "track", track)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid music genre or track"})
		return false
	}
	return true
}

func StringToFloat32(value string) float32 {
	if f, err := strconv.ParseFloat(value, 32); err == nil {
		return float32(f)
	}
	return float32(0)
}

func StringToInt64(value string) int64 {
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return int64(f)
	}
	return int64(0)
}
