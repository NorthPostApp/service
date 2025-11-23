package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const AUTH_LOG_PREFIX = "[ADMIN-LOGIN]"

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		clientIP := c.ClientIP()
		// check if the auth header exists
		if authHeader == "" {
			err := "Authorization header required."
			log.Printf("%s Invalid request | IP: %s | Error: %v", AUTH_LOG_PREFIX, clientIP, err)
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
			log.Printf("%s Invalid header | IP: %s | Error: %v", AUTH_LOG_PREFIX, clientIP, err)
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
