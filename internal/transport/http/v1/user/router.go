package user

import (
	"north-post/service/internal/transport/http/v1/user/handlers"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Music *handlers.MusicHandler
	User  *handlers.UserHandler
}

func SetupUserRouter(router *gin.RouterGroup, h *Handlers, userMiddleware gin.HandlerFunc) {
	user := router.Group("/user", userMiddleware)
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
	}
}
