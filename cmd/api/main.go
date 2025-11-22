package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/north-post/service/internal/transport/http/v1/admin"
)

func main() {
	router := gin.Default()

	router_v1 := router.Group("/v1")
	admin.SetupAdminRouters(router_v1)

	// Define a simple GET endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	if err := router.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
