package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"north-post/service/internal/infra"
	"north-post/service/internal/repository"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/admin"
	adminHandlers "north-post/service/internal/transport/http/v1/admin/handlers"
	"north-post/service/internal/transport/http/v1/middleware"
	"north-post/service/internal/transport/http/v1/user"
	userHandlers "north-post/service/internal/transport/http/v1/user/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var PORT_NUMBER = 8080

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

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
		logger.Error("failed to initialize firebase", "error", err)
		log.Fatalf("failed to initialize firebase: %v", err)
	}
	defer func() {
		if err := firebaseClient.Close(); err != nil {
			logger.Error("failed to close firebase client", "error", err)
		}
	}()

	// Initialize storage bucket client
	storageBucketClient, err := infra.NewStorageBucketClient(logger)
	if err != nil {
		logger.Error("failed to initialize storage bucker", "error", err)
		log.Fatalf("failed to initialize storage bucket: %v", err)
	}

	// Initialize LLM client
	llmClient, err := infra.NewLLMClient(logger)
	if err != nil {
		logger.Error("failed to initialize llm client", "error", err)
		log.Fatalf("failed to initialize llm client %v", err)
	}

	// Address service
	addressRepo := repository.NewAddressRepository(firebaseClient.Firestore, logger)
	addressService := services.NewAddressService(addressRepo, llmClient)
	adminAddressHandler := adminHandlers.NewAddressHandler(addressService, logger)

	// Prompt service
	promptRepo := repository.NewPromptRepository(firebaseClient.Firestore, logger)
	promptService := services.NewPromptService(promptRepo)
	promptHandler := adminHandlers.NewPromptHandler(promptService, logger)

	// User data service
	userRepo := repository.NewUserRepository(firebaseClient, logger)
	userService := services.NewUserService(userRepo)
	adminUserDataHandler := adminHandlers.NewUserHandler(userService, logger)
	appUserDataHandler := userHandlers.NewUserHandler(userService, logger)

	// Music service
	musicRepo := repository.NewMusicRepository(
		storageBucketClient.R2Storage,
		storageBucketClient.R2Presigned,
		firebaseClient.Firestore,
		logger,
	)
	musicService := services.NewMusicService(musicRepo)
	adminMusicHandler := adminHandlers.NewMusicHandler(musicService, logger)
	userMusicHandler := userHandlers.NewMusicHandler(musicService, logger)

	// Setup routers
	router := gin.Default()
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	origins := []string{}
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	router_v1 := router.Group("/v1")

	authMiddleware := middleware.AuthMiddleware(firebaseClient.Auth, logger)
	admin.SetupAdminRouter(router_v1,
		&admin.Handlers{
			Address: adminAddressHandler,
			Prompt:  promptHandler,
			User:    adminUserDataHandler,
			Music:   adminMusicHandler,
		},
		authMiddleware)

	user.SetupUserRouter(router_v1, &user.Handlers{
		Music: userMusicHandler,
		User:  appUserDataHandler,
	},
		authMiddleware)

	port := getPort()
	logger.Info("starting server", "port", port)
	if err := router.Run(fmt.Sprintf(`:%s`, port)); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
