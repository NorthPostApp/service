package admin

import (
	"north-post/service/internal/transport/http/v1/admin/services"
	"north-post/service/internal/transport/http/v1/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAdminRouters(router *gin.RouterGroup) {
	// group admin routers and register middleware
	admin_router := router.Group("/admin")
	admin_router.Use(middleware.AdminAuthMiddleware())
	// register routes
	admin_router.GET("/check", services.HealthCheck)

	// get stored addresses
	admin_router.POST("/addresses", services.GetAddresses)
}
