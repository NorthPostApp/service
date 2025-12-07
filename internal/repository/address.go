package repository

import (
	"context"
	"fmt"
	"log/slog"
	"north-post/service/internal/domain/v1/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const addressTablePrefix = "addresses"

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
	Language models.Language
	Tags     []string
	Limit    int
}

type GetAddressByIdOption struct {
	Language models.Language
	ID       string
}

// Get All addresses from the repository
// TODO: Pagination when the content size is getting larger
func (r *AddressRepository) GetAllAddresses(ctx context.Context, opts GetAllAddressesOptions) ([]models.AddressItem, error) {
	collectionName := getCollectionName(addressTablePrefix, opts.Language)
	query := r.client.Collection(collectionName).Query
	// Apply filters
	if len(opts.Tags) > 0 {
		query = query.Where("tags", "array-contains-any", opts.Tags)
	}
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	// execute query
	iter := query.Documents(ctx)
	defer iter.Stop()
	var addresses []models.AddressItem
	failedDocs := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			r.logger.Error("failed to iterate documents", "error", err)
			return nil, fmt.Errorf("failed to fetch addresses: %w", err)
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
	return addresses, nil
}

// Get a address by ID
func (r *AddressRepository) GetAddressById(ctx context.Context, opts GetAddressByIdOption) (*models.AddressItem, error) {
	collectionName := getCollectionName(addressTablePrefix, opts.Language)
	docRef := r.client.Collection(collectionName).Doc(opts.ID)
	// get document
	doc, err := docRef.Get(ctx)
	if err != nil {
		r.logger.Error("failed to get address document", "addressID", opts.ID, "error", err)
		return nil, fmt.Errorf("failed to get address with ID %s: %w", opts.ID, err)
	}
	// parse data
	var address models.AddressItem
	if err := doc.DataTo(&address); err != nil {
		r.logger.Error("failed to parse address document", "addressID", opts.ID, "error", err)
		return nil, fmt.Errorf("failed to parse address data: %w", err)
	}
	return &address, nil
}

// =========== Helper functions ==========
func getCollectionName(tablePrefix string, language models.Language) string {
	return fmt.Sprintf("%s_%s", tablePrefix, language.Get())
}
