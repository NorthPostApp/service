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
		address := admin.Group("/address")
		{
			address.GET("/:id", h.Address.GetAddressById)

			address.POST("", h.Address.GetAddresses)

			address.PUT("", h.Address.CreateNewAddress)
		}
		// health check
		admin.GET("/healthcheck", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "admin connected",
			})
		})
	}
}
