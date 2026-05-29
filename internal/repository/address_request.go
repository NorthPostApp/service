package repository

import (
	"context"
	"fmt"
	"log/slog"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/infra"
	"time"

	"cloud.google.com/go/firestore"
)

const (
	requestTablePrefix = "address_requests"
	batchLimit         = 500
)

type AddressRequestRepository struct {
	client *firestore.Client
	logger *slog.Logger
}

func NewAddressRequestRepository(client *infra.FirebaseClient, logger *slog.Logger) *AddressRequestRepository {
	return &AddressRequestRepository{
		client: client.Firestore,
		logger: logger,
	}
}

// Input and output types

type CreateRequestOptions struct {
	Language models.Language
	UID      string
	Content  string
}

type DeleteRequestsOptions struct {
	Language   models.Language
	UID        string
	RequestIDs []string
}

type GetRequestsByIDsOptions struct {
	Language   models.Language
	UID        string
	RequestIDs []string
}

type GetRequestsByIDsResponse struct {
	InvalidIDs          []string
	Requests            []models.AddressRequest
	ActiveRequestsCount int64
}

// Repo data processing functions

// IMPORTANT should implement limits to this function
func (r *AddressRequestRepository) CreateNewRequest(
	ctx context.Context, opts *CreateRequestOptions) (string, error) {
	collectionName := getRequestCollectionName(opts.Language)
	logger := r.logger.With(
		"path", "repository.address_request.CreateNewRequest",
		"collection", collectionName,
	)
	docRef := r.client.Collection(collectionName).NewDoc()
	now := time.Now().UnixMilli()
	newAddressRequest := models.AddressRequest{
		ID:        docRef.ID,
		Content:   opts.Content,
		RequestBy: opts.UID,
		Status:    models.RequestStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := docRef.Set(ctx, newAddressRequest)
	if err != nil {
		logger.Error("failed to create address request", "error", err)
		return "", fmt.Errorf("failed to create address request: %w", err)
	}
	return newAddressRequest.ID, nil
}

func (r *AddressRequestRepository) DeleteRequests(
	ctx context.Context,
	opts *DeleteRequestsOptions) error {
	collectionName := getRequestCollectionName(opts.Language)
	logger := r.logger.With(
		"path", "repository.address_request.DeleteRequests",
		"collection", collectionName,
		"uid", opts.UID,
	)
	collectionRef := r.client.Collection(collectionName)

	type bulkWriteJob struct {
		id  string
		job *firestore.BulkWriterJob
	}

	for start := 0; start < len(opts.RequestIDs); start += batchLimit {
		end := start + batchLimit
		if end > len(opts.RequestIDs) {
			end = len(opts.RequestIDs)
		}
		bulkWriter := r.client.BulkWriter(ctx)
		// use jobs to record the deletion result
		jobs := make([]bulkWriteJob, end-start)

		for i, id := range opts.RequestIDs[start:end] {
			docRef := collectionRef.Doc(id)
			job, err := bulkWriter.Delete(docRef)
			if err != nil {
				logger.Error("failed to enqueue address request delete", "requestID", id, "error", err)
				bulkWriter.End()
				return fmt.Errorf("Failed to enqueue address request delete: %w", err)
			}
			jobs[i] = bulkWriteJob{id, job}
		}
		bulkWriter.Flush() // this action blocks execution
		for _, j := range jobs {
			if _, err := j.job.Results(); err != nil {
				logger.Warn("failed to delete address request", "requestID", j.id, "error", err)
				bulkWriter.End()
				return fmt.Errorf("failed to delete address request: %w", err)
			}
		}
		bulkWriter.End()
	}
	return nil
}

// Next step:
// 1. GetRequestByUser -> by ids -> update active request count
// This function is for AppUser
func (r *AddressRequestRepository) GetRequestsByIDs(
	ctx context.Context, opts *GetRequestsByIDsOptions) (*GetRequestsByIDsResponse, error) {
	collectionName := getRequestCollectionName(opts.Language)
	logger := r.logger.With(
		"path", "repository.address_request.GetRequestByIDs",
		"collectionName", collectionName,
		"uid", opts.UID,
	)
	collectionRef := r.client.Collection(collectionName)
	docRefs := make([]*firestore.DocumentRef, len(opts.RequestIDs))
	for i, id := range opts.RequestIDs {
		docRefs[i] = collectionRef.Doc(id)
	}
	docs, err := r.client.GetAll(ctx, docRefs)
	if err != nil {
		logger.Error("failed to batch fetch requests", "error", err)
		return nil, fmt.Errorf("failed to batch fetch requests: %w", err)
	}
	requests := []models.AddressRequest{}
	invalidIDs := []string{}
	var activeRequestCount int64 = 0
	for _, doc := range docs {
		if !doc.Exists() {
			invalidIDs = append(invalidIDs, doc.Ref.ID)
			logger.Warn("request not found", "docID", doc.Ref.ID)
			continue
		}
		var request models.AddressRequest
		if err := doc.DataTo(&request); err != nil {
			invalidIDs = append(invalidIDs, doc.Ref.ID)
			logger.Warn("failed to parse document", "docID", doc.Ref.ID)
			continue
		}
		requests = append(requests, request)
		if request.Status == models.RequestStatusPending || request.Status == models.RequestStatusProcessing {
			activeRequestCount += 1
		}
	}
	return &GetRequestsByIDsResponse{
		InvalidIDs:          invalidIDs,
		Requests:            requests,
		ActiveRequestsCount: activeRequestCount,
	}, nil
}

// 2. GetRequestByAdmin -> by status

// ---------- Helper functions ----------
func getRequestCollectionName(language models.Language) string {
	return fmt.Sprintf("%s_%s", requestTablePrefix, language.Get())
}
