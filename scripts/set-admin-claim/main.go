// scripts/set-admin-claim/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	firebase "firebase.google.com/go/v4"
)

func main() {
	ctx := context.Background()
	envFile := fmt.Sprintf(".env.%s", "development")
	if err := godotenv.Load(envFile); err != nil {
		log.Fatal("env file is required")
	}
	uid := flag.String("uid", "", "Firebase Auth user UID")
	admin := flag.Bool("admin", true, "Whether the user should have admin access")
	flag.Parse()
	if *uid == "" {
		log.Fatal("missing required -uid")
	}
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GOOGLE_PROJECT_ID environment variable is required")
	}
	// Initialize firebase app
	config := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, config)
	if err != nil {
		log.Fatalf("failed to initialize firebase: %v", err)
	}
	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("failed to get auth client: %v", err)
	}
	userRecord, err := authClient.GetUser(ctx, *uid)
	if err != nil {
		log.Fatalf("failed to get user %q: %v", *uid, err)
	}
	claims := userRecord.CustomClaims
	if claims == nil {
		claims = map[string]interface{}{}
	}
	claims["admin"] = *admin
	if err := authClient.SetCustomUserClaims(ctx, *uid, claims); err != nil {
		log.Fatalf("Failed to set custom claims for user")
	}
	fmt.Printf("successfully set admin=%t for uid=%q\n", *admin, *uid)
}
