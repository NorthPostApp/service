package repository

import (
	"context"
	"fmt"
	"log/slog"
	"north-post/service/internal/domain/v1/models"
	"os"
	"slices"
	"time"

	"cloud.google.com/go/firestore"
	pb "cloud.google.com/go/firestore/apiv1/firestorepb"
	"google.golang.org/api/iterator"
)

const (
	addressTablePrefix  = "addresses"
	getByNameLimit      = 10
	tagsSimilarityLimit = 0.6
)

type AddressRepository struct {
	client *firestore.Client
	logger *slog.Logger
}

func NewAddressRepository(client *firestore.Client, logger *slog.Logger) *AddressRepository {
	return &AddressRepository{
		client: client,
		logger: logger,
	}
}

type GetAllAddressesOptions struct {
	Language      models.Language
	Tags          []string
	PageSize      int
	StartAfterDoc string
}

type GetAllAddressesResponse struct {
	Addresses  []models.AddressItem
	TotalCount int64
	LastDocID  string
	HasMore    bool
}

type GetAddressByIdOptions struct {
	Language models.Language
	ID       string
}

type UpdateAddressOption struct {
	Language    models.Language
	ID          string
	AddressItem models.AddressItem
}

type DeleteAddressOption struct {
	Language models.Language
	ID       string
}

type CreateNewAddressOption struct {
	Language    models.Language
	AddressItem models.AddressItem
}

// Get All addresses from the repository
func (r *AddressRepository) GetAllAddresses(ctx context.Context, opts GetAllAddressesOptions) (*GetAllAddressesResponse, error) {
	collectionName := getAddressCollectionName(opts.Language)
	query := r.client.Collection(collectionName).Query
	// Apply filters
	if len(opts.Tags) > 0 {
		query = query.Where("tags", "array-contains-any", opts.Tags)
	}
	// Sort by updated_at descending
	firstPageQuery := opts.StartAfterDoc == ""
	query = query.OrderBy("updatedAt", firestore.Desc)
	// 1. get total count only for the first page
	var totalCount int64 // need a cache to store the total count when not first page query
	if firstPageQuery {
		aggregationQuery := query.NewAggregationQuery().WithCount("total")
		aggregationResult, err := aggregationQuery.Get(ctx)
		if err != nil {
			r.logger.Error("failed to get count", "error", err)
			return nil, fmt.Errorf("failed to get total count")
		}
		count, ok := aggregationResult["total"]
		if !ok {
			r.logger.Error("total not found in aggregation result")
			return nil, fmt.Errorf("total not found in aggregation result")
		}
		totalCount = count.(*pb.Value).GetIntegerValue()
	}
	// 2. Apply pagination
	if !firstPageQuery {
		startDoc, err := r.client.Collection(collectionName).Doc(opts.StartAfterDoc).Get(ctx)
		if err != nil {
			r.logger.Error("failed to get start document", "error", err, "docID", opts.StartAfterDoc)
			return nil, fmt.Errorf("failed to get start document %q", opts.StartAfterDoc)
		}
		query = query.StartAfter(startDoc)
	}
	// fetch one extra to check if there's more
	if opts.PageSize > 0 {
		query = query.Limit(opts.PageSize + 1)
	}
	// 3. execute query
	iter := query.Documents(ctx)
	defer iter.Stop()
	var addresses []models.AddressItem
	failedDocs := 0
	hasMore := false
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			r.logger.Error("failed to iterate documents", "error", err)
			return nil, fmt.Errorf("failed to fetch addresses: %w", err)
		}
		// check if we have arrived extra document
		if opts.PageSize > 0 && len(addresses) >= opts.PageSize {
			hasMore = true
			break
		}
		var address models.AddressItem
		if err := doc.DataTo(&address); err != nil {
			r.logger.Warn("failed to parse document", "docID", doc.Ref.ID, "error", err)
			failedDocs++
			continue
		}
		addresses = append(addresses, address)
	}
	if failedDocs > 0 {
		r.logger.Warn("some documents failed to parse", "count", failedDocs)
	}
	var lastDocId string
	if len(addresses) > 0 {
		lastDocId = addresses[len(addresses)-1].ID
	}
	return &GetAllAddressesResponse{
		Addresses:  addresses,
		TotalCount: totalCount,
		LastDocID:  lastDocId,
		HasMore:    hasMore,
	}, nil
}

// Get a address by ID
func (r *AddressRepository) GetAddressById(ctx context.Context, opts GetAddressByIdOptions) (*models.AddressItem, error) {
	collectionName := getAddressCollectionName(opts.Language)
	docRef := r.client.Collection(collectionName).Doc(opts.ID)
	// get document
	doc, err := docRef.Get(ctx)
	if err != nil {
		r.logger.Error("failed to get address document", "addressID", opts.ID, "error", err)
		return nil, fmt.Errorf("failed to get address with ID %s", opts.ID)
	}
	// parse data
	var address models.AddressItem
	if err := doc.DataTo(&address); err != nil {
		r.logger.Error("failed to parse address document", "addressID", opts.ID, "error", err)
		return nil, fmt.Errorf("failed to parse address data: %w", err)
	}
	return &address, nil
}

func (r *AddressRepository) UpdateAddress(ctx context.Context, opts UpdateAddressOption) (*models.AddressItem, error) {
	collectionName := getAddressCollectionName(opts.Language)
	addressItem := opts.AddressItem
	addressItem.UpdatedAt = time.Now().Unix() // update timestamp
	addressItem.ID = opts.ID                  // avoid this value been modified by admin user
	docRef := r.client.Collection(collectionName).Doc(opts.ID)
	_, err := docRef.Set(ctx, addressItem)
	if err != nil {
		r.logger.Error("failed to update address", "addressID", opts.ID, "error", err)
		return nil, fmt.Errorf("failed to update address with ID %s: %w", opts.ID, err)
	}
	return &addressItem, nil
}

func (r *AddressRepository) DeleteAddress(ctx context.Context, opts DeleteAddressOption) (string, error) {
	collectionName := getAddressCollectionName(opts.Language)
	docRef := r.client.Collection(collectionName).Doc(opts.ID)
	_, err := docRef.Delete(ctx)
	if err != nil {
		r.logger.Error("failed to delete address", "addressID", opts.ID, "error", err)
		return "", fmt.Errorf("failed to delete address with ID %s: %w", opts.ID, err)
	}
	return opts.ID, nil
}

// Create a new address
func (r *AddressRepository) CreateNewAddress(ctx context.Context, opts CreateNewAddressOption) (string, error) {
	collectionName := getAddressCollectionName(opts.Language)
	// first check if there exists data with the same name
	query := r.client.Collection(collectionName).Where("name", "==", opts.AddressItem.Name).Limit(getByNameLimit)
	iter := query.Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			r.logger.Error("failed to check for duplicate records", "error", err)
			return "", fmt.Errorf("failed to check for duplicate records: %w", err)
		}
		var existingAddress models.AddressItem
		if err := doc.DataTo(&existingAddress); err != nil {
			r.logger.Warn("failed to parse existing address", "docID", doc.Ref.ID, "error", err)
			continue
		}
		similarity := compareTags(opts.AddressItem.Tags, existingAddress.Tags)
		if similarity > tagsSimilarityLimit {
			return "", fmt.Errorf("address with name '%s' and similar tags (%.0f%% similarity) already exists", opts.AddressItem.Name, similarity*100)
		}
	}
	// Auto generate timestamp
	now := time.Now().Unix()
	opts.AddressItem.CreatedAt = now
	opts.AddressItem.UpdatedAt = now
	// Create document with auto-generated ID
	docRef := r.client.Collection(collectionName).NewDoc()
	opts.AddressItem.ID = docRef.ID
	_, err := docRef.Set(ctx, opts.AddressItem)
	if err != nil {
		r.logger.Error("failed to create address", "error", err)
		return "", fmt.Errorf("failed to create address: %w", err)
	}
	return docRef.ID, nil
}

// =========== Helper functions ==========
func getAddressCollectionName(language models.Language) string {
	return fmt.Sprintf("%s_%s_%s", os.Getenv("MODE"), addressTablePrefix, language.Get())
}

func compareTags(tagsNewItem []string, tagsExistingItem []string) float32 {
	sameTagCount := 0
	for _, tag := range tagsNewItem {
		if exists := slices.Contains(tagsExistingItem, tag); exists {
			sameTagCount += 1
		}
	}
	return float32(sameTagCount) / float32(len(tagsNewItem))
}
