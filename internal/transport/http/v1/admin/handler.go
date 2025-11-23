package admin

import (
	"net/http"

	"north-post/service/internal/transport/http/v1/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAdminRouters(router *gin.RouterGroup) {
	// group admin routers and register middleware
	admin_router := router.Group("/admin")
	admin_router.Use(middleware.AdminAuthMiddleware())
	// register routes
	admin_router.GET("/check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "admin connected",
		})
	})
}
