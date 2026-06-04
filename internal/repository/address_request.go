package repository

import (
	"context"
	"fmt"
	"log/slog"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/infra"
	"strings"
	"time"
	"unicode/utf8"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	requestTablePrefix   = "address_requests"
	batchLimit           = 500
	openRequestLimit     = 10
	minimumContentLength = 10
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

type GetRequestsByStatusOptions struct {
	Language models.Language
	Status   models.AddressRequestStatus
}

type GetRequestsByStatusResponse struct {
	Requests   []models.AddressRequest
	InvalidIDs []string
}

type UpdateRequestDataOptions struct {
	Language       models.Language
	ID             string
	UpdatedRequest models.AddressRequest
}

// Repo data processing functions

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

// This get method is for AppUsers
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

// This get method is for Admin Dashboard
func (r *AddressRequestRepository) GetRequestsByStatus(
	ctx context.Context,
	opts *GetRequestsByStatusOptions) (*GetRequestsByStatusResponse, error) {
	collectionName := getRequestCollectionName(opts.Language)
	logger := r.logger.With(
		"path", "repository.address_request.GetRequestsByStatus",
		"collection", collectionName,
	)
	if opts == nil {
		logger.Error("invalid nil options")
		return nil, fmt.Errorf("options cannot be nil")
	}
	if !opts.Status.IsValid() {
		logger.Error("invalid status", "status", opts.Status)
		return nil, fmt.Errorf("invalid status: %s", opts.Status)
	}
	collectionRef := r.client.Collection(collectionName)
	iter := collectionRef.
		Where("status", "==", opts.Status).
		OrderBy("updatedAt", firestore.Desc).Documents(ctx)
	invalidIDs := []string{}
	requests := []models.AddressRequest{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Error("error getting requests", "error", err)
			return nil, fmt.Errorf("error getting requests: %w", err)
		}
		var request models.AddressRequest
		if err := doc.DataTo(&request); err != nil {
			logger.Warn("failed to parse request", "id", doc.Ref.ID, "error", err)
			invalidIDs = append(invalidIDs, doc.Ref.ID)
			continue
		}
		requests = append(requests, request)
	}
	return &GetRequestsByStatusResponse{
		Requests:   requests,
		InvalidIDs: invalidIDs,
	}, nil
}

// update address request content
func (r *AddressRequestRepository) UpdateRequestData(
	ctx context.Context, opts *UpdateRequestDataOptions) error {
	collectionName := getRequestCollectionName(opts.Language)
	logger := r.logger.With(
		"path", "repository.address_request.UpdateRequestData",
		"language", opts.Language,
		"id", opts.ID,
	)
	docRef := r.client.Collection(collectionName).Doc(opts.ID)
	doc, err := docRef.Get(ctx)
	if err != nil {
		logger.Error("failed to get doc", "id", opts.ID, "error", err)
		return fmt.Errorf("failed to get doc: %w", err)
	}
	var request models.AddressRequest
	if err := doc.DataTo(&request); err != nil {
		logger.Error("failed to parse request", "id", opts.ID, "error", err)
		return fmt.Errorf("failed to parse request: %w", err)
	}
	newRequest := opts.UpdatedRequest
	// avoid data pollution
	newRequest.ID = request.ID
	newRequest.RequestBy = request.RequestBy
	newRequest.CreatedAt = request.CreatedAt
	newRequest.UpdatedAt = time.Now().UnixMilli()
	if newRequest.Status == models.RequestStatusPending {
		newRequest.Status = models.RequestStatusProcessing
	}
	_, err = docRef.Set(ctx, newRequest)
	if err != nil {
		logger.Error("failed to update request", "error", err)
		return fmt.Errorf("failed to update request: %w", err)
	}
	return nil
}

// ---------- Special Use Cases: Cross-Repo Processing ---------
func (r *AddressRequestRepository) CreateNewRequestWithLimit(
	ctx context.Context, opts *CreateRequestOptions) (string, error) {
	collectionName := getRequestCollectionName(opts.Language)
	logger := r.logger.With(
		"path", "repository.address_request.CreateNewRequestWithLimit",
		"uid", opts.UID,
		"collection", collectionName,
	)
	// if the content is too short, skip the rest transactions
	if !validRequestContentLength(opts.Language, opts.Content) {
		logger.Error("insufficient count length", "content", opts.Content)
		return "", fmt.Errorf("insufficient content length. content: %s", opts.Content)
	}

	newRequestDocRef := r.client.Collection(collectionName).NewDoc()
	userRequestRef := r.client.
		Collection(appUserTable).Doc(opts.UID).
		Collection(addressRequestCollection).Doc(opts.Language.Get())
	// transactional update to avoid race condition or abusive requests
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var currentUserRequests models.UserAddressRequests
		doc, err := tx.Get(userRequestRef)
		exists := true
		if status.Code(err) == codes.NotFound {
			currentUserRequests = models.UserAddressRequests{
				IDs:                []string{},
				ActiveRequestCount: 0,
			}
			exists = false
		} else if err != nil {
			return fmt.Errorf("failed to get user's request data: %w", err)
		} else if err := doc.DataTo(&currentUserRequests); err != nil {
			return fmt.Errorf("failed to parse user's request data: %w", err)
		}
		if currentUserRequests.ActiveRequestCount >= openRequestLimit {
			return fmt.Errorf("too many active address requests")
		}
		now := time.Now().UnixMilli()
		newAddressRequest := models.AddressRequest{
			ID:        newRequestDocRef.ID,
			Content:   opts.Content,
			RequestBy: opts.UID,
			Status:    models.RequestStatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Set(newRequestDocRef, newAddressRequest); err != nil {
			return fmt.Errorf("failed to create address request: %w", err)
		}
		if !exists {
			currentUserRequests.IDs = []string{newAddressRequest.ID}
			currentUserRequests.ActiveRequestCount = 1
			err = tx.Set(userRequestRef, currentUserRequests)
		} else {
			err = tx.Update(userRequestRef, []firestore.Update{
				{Path: requestIDsPath, Value: firestore.ArrayUnion(newAddressRequest.ID)},
				{Path: activeRequestCountPath, Value: currentUserRequests.ActiveRequestCount + 1},
			})
		}
		if err != nil {
			return fmt.Errorf("failed to update user request data: %w", err)
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to create new address request", "error", err.Error())
		return "", err
	}
	return newRequestDocRef.ID, nil
}

// ---------- Helper functions ----------
func getRequestCollectionName(language models.Language) string {
	return fmt.Sprintf("%s_%s", requestTablePrefix, language.Get())
}

func validRequestContentLength(language models.Language, content string) bool {
	content = strings.TrimSpace(content)
	switch language {
	case models.LanguageZH:
		return utf8.RuneCountInString(content) >= minimumContentLength
	case models.LanguageEN:
		return len(strings.Fields(content)) >= minimumContentLength
	default:
		return false
	}
}
