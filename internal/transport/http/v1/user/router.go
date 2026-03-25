package user

import (
	"north-post/service/internal/transport/http/v1/user/handlers"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Music *handlers.MusicHandler
}

func SetupUserRouter(router *gin.RouterGroup, h *Handlers) {
	user := router.Group("/user")
	{
		music := user.Group("/music")
		{
			music.GET("/list", h.Music.GetMusicList)
			music.GET("/:genre/:track", h.Music.GetPresignedMusicURL)
		}
	}
}
