package admin

import (
	"north-post/service/internal/transport/http/v1/admin/handlers"
	"north-post/service/internal/transport/http/v1/middleware"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Address        *handlers.AddressHandler
	AddressRequest *handlers.AddressRequestHandler
	Prompt         *handlers.PromptHandler
	User           *handlers.UserHandler
	Music          *handlers.MusicHandler
	Typesense      *handlers.TypesenseHandler
}

func SetupAdminRouter(router *gin.RouterGroup, h *Handlers, middlewares *middleware.Middlewares) {
	admin := router.Group("/admin", middlewares.Auth)
	{
		address := admin.Group("/address")
		{
			address.GET("/tags", h.Address.GetAllTags)
			address.POST("", h.Address.GetAddresses)
			address.POST("/generate", h.Address.GenerateNewAddress)
			address.POST("/update", h.Address.UpdateAddress)
			address.POST("/sync", h.Address.SyncToTypesense)
			address.PUT("", h.Address.CreateNewAddress)
			address.DELETE("/:id", h.Address.DeleteAddress)
		}
		addressRequest := admin.Group("/address-request")
		{
			addressRequest.GET("", middlewares.LanguageFromQuery, h.AddressRequest.GetRequestsByStatus)
			addressRequest.POST("/update", middlewares.LanguageFromBody, h.AddressRequest.UpdateRequest)
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
