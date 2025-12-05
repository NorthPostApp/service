package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"north-post/service/internal/firebase"
	"north-post/service/internal/repository"
	"north-post/service/internal/transport/http/v1/admin"
	"north-post/service/internal/transport/http/v1/admin/handlers"
	"north-post/service/internal/transport/http/v1/admin/services"
	"north-post/service/internal/transport/http/v1/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const PORT_NUMBER = 8080

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := godotenv.Load(); err != nil {
		logger.Info("no .env file found, using system environment variables")
	}

	// Initialize Firebase client
	firebaseClient, err := firebase.NewFirebaseClient(logger)
	if err != nil {
		log.Fatalf("failed to initialize firebase: %v", err)
	}
	defer func() {
		if err := firebaseClient.Close(); err != nil {
			logger.Error("failed to close firebase client", "error", err)
		}
	}()

	addressRepo := repository.NewAddressRepository(firebaseClient.Firestore, logger)
	addressService := services.NewAddressService(addressRepo)
	addressHandler := handlers.NewAddressHandler(addressService, logger)

	router := gin.Default()
	router_v1 := router.Group("/v1")

	adminMiddleware := middleware.AdminAuthMiddleware(logger)
	admin.SetupAdminRouter(router_v1,
		&admin.Handlers{
			Address: addressHandler,
		},
		adminMiddleware)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	logger.Info("starting server", "port", PORT_NUMBER)
	if err := router.Run(fmt.Sprintf(`:%d`, PORT_NUMBER)); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
