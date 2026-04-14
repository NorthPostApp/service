package infra

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/typesense/typesense-go/v4/typesense"
	"github.com/typesense/typesense-go/v4/typesense/api"
)

type TypesenseClient struct {
	Client *typesense.Client
}

func NewTypesenseClient(logger *slog.Logger) (*TypesenseClient, error) {
	typesenseURL := os.Getenv("TYPESENSE_URL")
	apiKey := os.Getenv("TYPESENSE_API_KEY")
	if typesenseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("TYPESENSE_URL and TYPESENSE_API_KEY are required")
	}
	client := typesense.NewClient(
		typesense.WithServer(typesenseURL),
		typesense.WithAPIKey(apiKey),
	)
	logger.Info("Typesense client initialized successfully", "server", typesenseURL)
	return &TypesenseClient{Client: client}, nil
}

// The typesense collection schema should be the same as this struct
type TypesenseAddressRecord struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	BriefIntro string   `json:"briefIntro"`
	Tags       []string `json:"tags"`
}

func GetTypesenseAddressCollectionSchema(name string) *api.CollectionSchema {
	return &api.CollectionSchema{
		Name: name,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "briefIntro", Type: "string"},
			{Name: "tags", Type: "string[]"},
		},
	}
}
