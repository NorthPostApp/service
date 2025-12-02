package firebase

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

var (
	FirestoreClient *firestore.Client
)

func InitializeFirebase() error {
	ctx := context.Background()
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if credentialsPath == "" {
		log.Println("[Firebase] Warning: Using Application Default Credentials")
	} else {
		log.Println("[Firebase] Using local Credentials file")
	}
	opt := option.WithCredentialsFile(credentialsPath)
	client, err := firestore.NewClient(ctx, projectID, opt)
	if err != nil {
		return fmt.Errorf("error initializing app: %v", err)
	}
	FirestoreClient = client
	log.Println(("[Firebase] Successfully initialized"))
	return nil
}
