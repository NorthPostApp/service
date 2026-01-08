package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

type authClient interface {
	VerifyIDToken(c context.Context, idToken string) (*auth.Token, error)
}

func AdminAuthMiddleware(auth authClient, logger *slog.Logger) gin.HandlerFunc {
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

		// verify the firebase id token
		idToken := headerParts[1]
		authToken, err := auth.VerifyIDToken(c, idToken)
		if err != nil {
			logger.Error("Failed to verify ID token", "error", err, "clientIP", clientIP)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		c.Set("user_id", authToken.UID)
		c.Next()
	}
}
