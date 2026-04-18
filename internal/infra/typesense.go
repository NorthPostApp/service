package infra

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/transport/http/v1/utils"
	"os"
	"strings"
	"time"

	"github.com/typesense/typesense-go/v4/typesense"
	"github.com/typesense/typesense-go/v4/typesense/api"
	"github.com/typesense/typesense-go/v4/typesense/api/pointer"
)

const (
	defaultPageSize = 20
	maxPagesize     = 100
)

type TypesenseClient struct {
	Client *typesense.Client
	logger *slog.Logger
}

type TypesenseSystemInfo struct {
	Health                     bool
	SystemCPUActivePercentage  float32
	SystemDiskTotalBytes       int64
	SystemDiskUsedBytes        int64
	SystemMemoryTotalBytes     int64
	SystemMemoryUsedBytes      int64
	SystemNetworkSentBytes     int64
	SystemNetworkReceivedBytes int64
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
	UpdatedAt  int64    `json:"updatedAt"`
}

func (c *TypesenseClient) GetAddressCollectionSchema(
	name string,
	language models.Language,
) *api.CollectionSchema {
	return &api.CollectionSchema{
		Name: name,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string", Locale: pointer.String(language.Get())},
			{Name: "briefIntro", Type: "string", Locale: pointer.String(language.Get())},
			{Name: "tags", Type: "string[]"},
			{Name: "updatedAt", Type: "int64"},
		},
	}
}

func (c *TypesenseClient) CreateAddressRecord(addressItem *models.AddressItem) TypesenseAddressRecord {
	return TypesenseAddressRecord{
		ID:         addressItem.ID,
		Name:       addressItem.Name,
		BriefIntro: addressItem.BriefIntro,
		Tags:       addressItem.Tags,
		UpdatedAt:  addressItem.UpdatedAt,
	}
}

type SyncDatabaseResult struct {
	Total   int
	Success int
	Failed  int
}

type SearchAddressesParams struct {
	CollectionName string
	Keywords       string
	Tags           []string
	PageSize       int
	Page           int
}

type SearchAddressesResult struct {
	Hits       []string // file IDs
	Page       int
	PageSize   int
	TotalCount int64
}

func (c *TypesenseClient) SyncAddressDatabase(
	ctx context.Context,
	language models.Language,
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
	schema := c.GetAddressCollectionSchema(collectionName, language)
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

func (c *TypesenseClient) SearchAddresses(
	ctx context.Context, params *SearchAddressesParams) (*SearchAddressesResult, error) {
	q := "*"
	if params.Keywords != "" {
		q = params.Keywords
	}
	page := params.Page
	if page <= 0 {
		page = 1
	}
	perPage := params.PageSize
	if perPage <= 0 {
		perPage = defaultPageSize
	} else if perPage > maxPagesize {
		perPage = maxPagesize
	}
	searchParams := &api.SearchCollectionParams{
		Q:       pointer.String(q),
		QueryBy: pointer.String("name,briefIntro"),
		SortBy:  pointer.String("updatedAt:desc"),
		Page:    &page,
		PerPage: &perPage,
		// Turn on these two parameters when project go online
		// UseCache: pointer.True(),
		// CacheTtl: pointer.Int(60), // 60 seconds cache
	}
	if len(params.Tags) > 0 {
		filterStr := fmt.Sprintf("tags:=[%s]", strings.Join(params.Tags, ","))
		searchParams.FilterBy = &filterStr
	}
	result, err := c.Client.Collection(params.CollectionName).Documents().Search(ctx, searchParams)
	if err != nil {
		c.logger.Error("typesense search failed",
			"collectionName", params.CollectionName,
			"keywords", params.Keywords,
			"tags", params.Tags,
			"page", params.Page,
			"pageSize", params.PageSize,
		)
		return nil, fmt.Errorf("typesense search failed: %w", err)
	}
	var records []string
	for _, hit := range *result.Hits {
		doc := *hit.Document
		records = append(records, doc["id"].(string))
	}
	return &SearchAddressesResult{
		Hits:       records,
		TotalCount: int64(*result.Found),
		Page:       int(*result.Page),
		PageSize:   perPage,
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

// Typesense cluster operations
func (c *TypesenseClient) GetSystemInfo(ctx context.Context) (*TypesenseSystemInfo, error) {
	health, err := c.Client.Health(ctx, 3*time.Second)
	if err != nil {
		c.logger.Error(
			"failed to get Typesense system health",
			"error", err,
			"url", os.Getenv("TYPESENSE_URL"),
		)
		return nil, fmt.Errorf("failed to get Typesense system health: %w", err)
	}
	metricsCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	metrics, err := c.Client.Metrics().Retrieve(metricsCtx)
	if err != nil {
		c.logger.Error(
			"failed to get Typesense system metrics",
			"error", err,
			"url", os.Getenv("TYPESENSE_URL"),
		)
		return nil, fmt.Errorf("failed to get Typesense system metrics: %w", err)
	}
	typesenseSystemInfo := &TypesenseSystemInfo{Health: health}
	if v, ok := metrics["system_cpu_active_percentage"].(string); ok {
		typesenseSystemInfo.SystemCPUActivePercentage = utils.StringToFloat32(v)
	}
	if v, ok := metrics["system_disk_total_bytes"].(string); ok {
		typesenseSystemInfo.SystemDiskTotalBytes = utils.StringToInt64(v)
	}
	if v, ok := metrics["system_disk_used_bytes"].(string); ok {
		typesenseSystemInfo.SystemDiskUsedBytes = utils.StringToInt64(v)
	}
	if v, ok := metrics["system_memory_total_bytes"].(string); ok {
		typesenseSystemInfo.SystemMemoryTotalBytes = utils.StringToInt64(v)
	}
	if v, ok := metrics["system_memory_used_bytes"].(string); ok {
		typesenseSystemInfo.SystemMemoryUsedBytes = utils.StringToInt64(v)
	}
	if v, ok := metrics["system_network_sent_bytes"].(string); ok {
		typesenseSystemInfo.SystemNetworkSentBytes = utils.StringToInt64(v)
	}
	if v, ok := metrics["system_network_received_bytes"].(string); ok {
		typesenseSystemInfo.SystemNetworkReceivedBytes = utils.StringToInt64(v)
	}
	return typesenseSystemInfo, nil
}
