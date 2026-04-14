package infra

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"north-post/service/internal/domain/v1/models"
	"os"

	"github.com/typesense/typesense-go/v4/typesense"
	"github.com/typesense/typesense-go/v4/typesense/api"
)

type TypesenseClient struct {
	Client *typesense.Client
	logger *slog.Logger
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
	return &TypesenseClient{Client: client, logger: logger}, nil
}

// The typesense collection schema should be the same as this struct
type TypesenseAddressRecord struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	BriefIntro string   `json:"briefIntro"`
	Tags       []string `json:"tags"`
}

func (c *TypesenseClient) GetAddressCollectionSchema(name string) *api.CollectionSchema {
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

func (c *TypesenseClient) CreateAddressRecord(addressItem *models.AddressItem) TypesenseAddressRecord {
	return TypesenseAddressRecord{
		ID:         addressItem.ID,
		Name:       addressItem.Name,
		BriefIntro: addressItem.BriefIntro,
		Tags:       addressItem.Tags,
	}
}

type SyncDatabaseResult struct {
	Total   int
	Success int
	Failed  int
}

func (c *TypesenseClient) SyncAddressDatabase(
	ctx context.Context,
	collectionName string,
	documents []interface{}) (*SyncDatabaseResult, error) {
	// first drop the entire collection to avoid mismatch records
	_, err := c.Client.Collection(collectionName).Delete(ctx)
	if err != nil {
		var httpErr *typesense.HTTPError
		if !errors.As(err, &httpErr) || httpErr.Status != 404 {
			c.logger.Error("failed to delete typesense collection when syncing",
				"collectionName", collectionName,
				"error", err,
			)
			return nil, fmt.Errorf("failed to delete typesense collection when syncing: %w", err)
		}
	}
	// create a new collection
	schema := c.GetAddressCollectionSchema(collectionName)
	_, err = c.Client.Collections().Create(ctx, schema)
	if err != nil {
		c.logger.Error(
			"failed to create typesense collection",
			"collectionName", collectionName,
			"error", err,
		)
		return nil, fmt.Errorf("failed to create typesense collection: %w", err)
	}
	// directly return 0 records if no documents were found
	if len(documents) == 0 {
		return &SyncDatabaseResult{}, nil
	}
	// sync documents
	action := api.Create
	params := &api.ImportDocumentsParams{Action: &action}
	results, err := c.Client.Collection(collectionName).
		Documents().
		Import(ctx, documents, params)
	if err != nil {
		c.logger.Error(
			"failed to import documents to typesense",
			"collectionName", collectionName,
			"error", err,
		)
		return nil, fmt.Errorf("failed to import documents to typesense: %w", err)
	}
	// Tally results
	success := 0
	for _, result := range results {
		if result.Success {
			success++
		} else {
			c.logger.Warn(
				"failed to sync document to typesense",
				"id", result.Id,
				"error", result.Error,
			)
		}
	}
	return &SyncDatabaseResult{
		Total:   len(results),
		Success: success,
		Failed:  len(results) - success,
	}, nil
}

func (c *TypesenseClient) UpsertAddressData(
	ctx context.Context, collectionName string, addressItem *models.AddressItem) {
	record := c.CreateAddressRecord(addressItem)
	_, err := c.Client.
		Collection(collectionName).
		Documents().
		Upsert(ctx, record, &api.DocumentIndexParameters{})
	if err != nil {
		c.logger.Warn("failed to upsert document",
			"collectionName", collectionName,
			"id", addressItem.ID,
			"error", err,
		)
	}
}

func (c *TypesenseClient) DeleteAddressData(
	ctx context.Context, collectionName string, addressID string) {
	_, err := c.Client.Collection(collectionName).Document(addressID).Delete(ctx)
	if err != nil {
		c.logger.Warn("failed to delete document",
			"collectionName", collectionName,
			"id", addressID,
			"error", err,
		)
	}

}
