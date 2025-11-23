package main

import (
	"fmt"
	"log"
	"net/http"

	"north-post/service/internal/transport/http/v1/admin"

	"github.com/gin-gonic/gin"
)

const PORT_NUMBER = 8080

func main() {
	router := gin.Default()
	router_v1 := router.Group("/v1")

	admin.SetupAdminRouters(router_v1)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	if err := router.Run(fmt.Sprintf(`:%d`, PORT_NUMBER)); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
