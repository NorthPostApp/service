package handlers

import (
	"log/slog"
	"net/http"
	"north-post/service/internal/transport/http/v1/dto"

	"github.com/gin-gonic/gin"
)

// Helper functions
func validateUser(c *gin.Context, uid string, logger *slog.Logger) bool {
	if uid == "" {
		logger.Error(
			"missing user id from the middleware context",
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Unauthorized id token"})
		return false
	}
	return true
}
