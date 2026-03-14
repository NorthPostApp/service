package infra

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type FirebaseClient struct {
	Firestore *firestore.Client
	Auth      *auth.Client
}

func NewFirebaseClient(logger *slog.Logger) (*FirebaseClient, error) {
	ctx := context.Background()
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("GOOGLE_PROJECT_ID environment variable is required")
	}

	var opts []option.ClientOption
	if credentialsPath != "" {
		logger.Info("initializing firebase client using credentials file")
		opts = append(opts, option.WithCredentialsFile(credentialsPath))
	} else {
		logger.Info("initializing firebase using application default credentials")
	}

	// Initialize firebase app
	config := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}

	// initialize firestore client
	databaseID := "(default)"
	if os.Getenv("MODE") == "dev" {
		databaseID = "development"
	}
	firestoreClient, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID, opts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing firestore: %w", err)
	}

	// initialize auth client
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing auth: %w", err)
	}
	logger.Info("firebase initialized successfully", "Database ID", databaseID)

	return &FirebaseClient{
		Firestore: firestoreClient,
		Auth:      authClient,
	}, nil
}

func (f *FirebaseClient) Close() error {
	if f.Firestore != nil {
		return f.Firestore.Close()
	}
	return nil
}
