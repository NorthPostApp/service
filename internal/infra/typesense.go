package infra

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/typesense/typesense-go/v4/typesense"
)

const (
	typesensePort = "8108"
)

type TypesenseClient struct {
	TypeSense *typesense.Client
}

func NewTypesenseClient(logger *slog.Logger) (*TypesenseClient, error) {
	host := os.Getenv("TYPESENSE_HOST")
	port := typesensePort
	apiKey := os.Getenv("TYPESENSE_API_KEY")
	if host == "" || apiKey == "" {
		return nil, fmt.Errorf("TYPESENSE_HOST and TYPESENSE_API_KEY are required")
	}
	client := typesense.NewClient(
		typesense.WithServer(fmt.Sprintf("%s:%s", host, port)),
		typesense.WithAPIKey(apiKey),
	)
	logger.Info("Typesense client initialized successfully")
	return &TypesenseClient{TypeSense: client}, nil
}
