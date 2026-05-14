package user

import (
	"north-post/service/internal/transport/http/v1/middleware"
	"north-post/service/internal/transport/http/v1/user/handlers"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Music       *handlers.MusicHandler
	User        *handlers.UserHandler
	Address     *handlers.AddressHandler
	AddressBook *handlers.AddressBookHandler
}

func SetupUserRouter(router *gin.RouterGroup, h *Handlers, middlewares *middleware.Middlewares) {
	user := router.Group("/user", middlewares.Auth)
	{
		music := user.Group("/music")
		{
			music.GET("/list", h.Music.GetMusicList)
			music.GET("/:genre/:track", h.Music.GetPresignedMusicURL)
		}
		signIn := user.Group("/signin")
		{
			signIn.POST("", h.User.AuthenticateAppUser)
		}
		address := user.Group("/address")
		{
			address.POST("", middlewares.LanguageFromBody, h.Address.GetAddresses)
			address.GET("/tags", middlewares.LanguageFromQuery, h.Address.GetAllTags)
		}
		addressBook := user.Group("/address-book")
		{
			addressBook.PATCH("", middlewares.LanguageFromBody, h.AddressBook.UpdateSavedAddresses)
			addressBook.GET("", middlewares.LanguageFromQuery, h.AddressBook.GetSavedAddresses)
		}
	}
}
