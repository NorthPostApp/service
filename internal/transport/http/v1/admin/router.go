package admin

import (
	"net/http"
	"north-post/service/internal/transport/http/v1/admin/handlers"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Address *handlers.AddressHandler
}

func SetupAdminRouter(router *gin.RouterGroup, h *Handlers, adminMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin", adminMiddleware)
	{
		addresses := admin.Group("/addresses")
		{
			addresses.POST("", h.Address.GetAddresses)
		}
		// health check
		admin.GET("/healthcheck", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "admin connected",
			})
		})
	}
}
