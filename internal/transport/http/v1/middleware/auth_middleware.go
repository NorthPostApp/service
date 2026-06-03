package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"

	"firebase.google.com/go/v4/auth"
)

const (
	UidKey          = "user_id"
	adminCollection = "admin_users"
	userCollection  = "app_users"
)

type MiddlewareType int8

const (
	AdminMiddleware MiddlewareType = iota
	UserMiddleware
)

func AuthMiddleware(
	middlewareType MiddlewareType,
	auth *auth.Client,
	db *firestore.Client,
	logger *slog.Logger) gin.HandlerFunc {
	var collectionName string
	if middlewareType == AdminMiddleware {
		collectionName = adminCollection
	} else if middlewareType == UserMiddleware {
		collectionName = userCollection
	}
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
		if _, err := db.Collection(collectionName).Doc(authToken.UID).Get(c); err != nil {
			logger.Error("invalid user group",
				"uid", authToken.UID,
				"err", err,
				"clientIP", clientIP,
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user group",
			})
			c.Abort()
			return
		}
		c.Set(UidKey, authToken.UID)
		c.Next()
	}
}
