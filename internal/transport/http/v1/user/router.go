package user

import (
	"north-post/service/internal/transport/http/v1/middleware"
	"north-post/service/internal/transport/http/v1/user/handlers"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Music   *handlers.MusicHandler
	User    *handlers.UserHandler
	Address *handlers.AddressHandler
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
			address.GET("/tags", h.Address.GetAllTags)
		}
	}
}
