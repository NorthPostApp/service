package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AdminAuthMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		clientIP := c.ClientIP()
		// check if the auth header exists
		if authHeader == "" {
			err := "Authorization header required."
			logger.Error("Invalid admin request", "error", err, "ip", clientIP)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err,
			})
			c.Abort()
			return
		}

		// check if the auth header format correct
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			err := fmt.Sprintf("Authorization header format must be Bearer {token}, got=%s", headerParts[0])
			logger.Error("Invalid authorization header", "error", err, "clientIP", clientIP)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err,
			})
			c.Abort()
			return
		}

		// TODO: implement authorization with firebase or cache

		c.Next()
	}
}
