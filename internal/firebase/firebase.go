package firebase

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type FirebaseClient struct {
	Firestore *firestore.Client
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
	client, err := firestore.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing firestore: %w", err)
	}
	logger.Info("firebase initialized successfully")

	return &FirebaseClient{
		Firestore: client,
	}, nil
}

func (f *FirebaseClient) Close() error {
	if f.Firestore != nil {
		return f.Firestore.Close()
	}
	return nil
}
