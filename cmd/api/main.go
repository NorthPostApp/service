package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"north-post/service/internal/infra"
	"north-post/service/internal/repository"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/admin"
	"north-post/service/internal/transport/http/v1/admin/handlers"
	"north-post/service/internal/transport/http/v1/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const PORT_NUMBER = 8080

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Load environments
	env := os.Getenv("GO_ENV")
	if env == "" || env != "production" {
		env = "development"
	}
	logger.Info("loading environment", "env", env)
	envFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(envFile); err != nil {
		logger.Info("no .env file found, using system environment variables")
	} else {
		logger.Info("loaded environment configuration", "file", envFile)
	}

	// Initialize Firebase client
	firebaseClient, err := infra.NewFirebaseClient(logger)
	if err != nil {
		log.Fatalf("failed to initialize firebase: %v", err)
	}
	defer func() {
		if err := firebaseClient.Close(); err != nil {
			logger.Error("failed to close firebase client", "error", err)
		}
	}()

	// Initialize LLM client
	llmClient, err := infra.NewLLMClient(logger)

	addressRepo := repository.NewAddressRepository(firebaseClient.Firestore, logger)
	addressService := services.NewAddressService(addressRepo, llmClient)
	adminAddressHandler := handlers.NewAddressHandler(addressService, logger)

	promptRepo := repository.NewPromptRepository(firebaseClient.Firestore, logger)
	promptService := services.NewPromptService(promptRepo)
	promptHandler := handlers.NewPromptHandler(promptService, logger)

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	router_v1 := router.Group("/v1")

	adminMiddleware := middleware.AdminAuthMiddleware(logger)
	admin.SetupAdminRouter(router_v1,
		&admin.Handlers{
			Address: adminAddressHandler,
			Prompt:  promptHandler,
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
