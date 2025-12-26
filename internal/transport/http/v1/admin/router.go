package admin

import (
	"north-post/service/internal/transport/http/v1/admin/handlers"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Address *handlers.AddressHandler
	Prompt  *handlers.PromptHandler
}

func SetupAdminRouter(router *gin.RouterGroup, h *Handlers, adminMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin", adminMiddleware)
	{
		address := admin.Group("/address")
		{
			address.GET("/:id", h.Address.GetAddressById)

			address.POST("", h.Address.GetAddresses)
			address.POST("/generate", h.Address.GenerateNewAddress)

			address.PUT("", h.Address.CreateNewAddress)
		}
		prompt := admin.Group("/prompt")
		{
			prompt.GET("/newaddress", h.Prompt.GetSystemAddressGenerationPrompt)
		}
	}
}
