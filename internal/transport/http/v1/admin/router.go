package admin

import (
	"north-post/service/internal/transport/http/v1/admin/handlers"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Address   *handlers.AddressHandler
	Prompt    *handlers.PromptHandler
	User      *handlers.UserHandler
	Music     *handlers.MusicHandler
	Typesense *handlers.TypesenseHandler
}

func SetupAdminRouter(router *gin.RouterGroup, h *Handlers, adminMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin", adminMiddleware)
	{
		address := admin.Group("/address")
		{
			// GET
			address.GET("/:id", h.Address.GetAddressById)
			address.GET("/tags", h.Address.GetAllTags)
			// POST
			address.POST("", h.Address.GetAddresses)
			address.POST("/generate", h.Address.GenerateNewAddress)
			address.POST("/update", h.Address.UpdateAddress)
			address.POST("/sync", h.Address.SyncToTypesense)
			// PUT
			address.PUT("", h.Address.CreateNewAddress)
			// DELETE
			address.DELETE("/:id", h.Address.DeleteAddress)
		}
		prompt := admin.Group("/prompt")
		{
			prompt.GET("/system/address", h.Prompt.GetSystemAddressGenerationPrompt)
		}
		music := admin.Group("/music")
		{
			music.GET("", h.Music.GetMusicList)
			music.GET("/:genre/:track", h.Music.GetPresignedMusicURL)
		}
		signIn := admin.Group("/signin")
		{
			signIn.POST("", h.User.SignInAdminUser)
		}
		typesense := admin.Group("/typesense")
		{
			typesense.GET("/info", h.Typesense.GetSystemInfo)
		}
	}
}
