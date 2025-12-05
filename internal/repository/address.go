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

type GetAddressesOptions struct {
	Language string
	Tags     []string
	Limit    int
}

func (r *AddressRepository) GetAll(c context.Context, opts GetAddressesOptions) ([]models.AddressItem, error) {
	collectionName := fmt.Sprintf("%s_%s", addressTablePrefix, opts.Language)
	query := r.client.Collection(collectionName).Query

	// Apply filters
	if len(opts.Tags) > 0 {
		query = query.Where("tags", "array-contains-any", opts.Tags)
	}
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}

	// execute query
	iter := query.Documents(c)
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
