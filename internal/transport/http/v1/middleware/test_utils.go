package middleware

import (
	"github.com/gin-gonic/gin"
)

// --------- Mock Utils ----------
func MockAuthMiddleware(uid string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if uid != "" {
			c.Set(UidKey, uid)
		}
		c.Next()
	}
}

func MockLanguageMiddleware(language string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if language != "" {
			c.Set(LanguageKey, language)
		}
		c.Next()
	}
}
