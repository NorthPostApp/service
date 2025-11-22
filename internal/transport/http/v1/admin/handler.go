package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupAdminRouters(router *gin.RouterGroup) {
	router.GET("/admin-check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "admin connected",
		})
	})
}
