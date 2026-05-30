package repository

import (
	"context"
	"fmt"
	"log/slog"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/infra"
	"time"

	"cloud.google.com/go/firestore"

	"firebase.google.com/go/v4/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	adminUserTable           = "admin_users"
	appUserTable             = "app_users"
	savedAddressesCollection = "saved_addresses"
	addressRequestCollection = "requests"
	savedAddressesIDsPath    = "ids"
	requestIDsPath           = "ids"
	activeRequestCountPath   = "activeRequestCount"
)

type UpdateAddressBookAction int

const (
	Add UpdateAddressBookAction = iota
	Delete
)

type UserRepository struct {
	firestoreClient *firestore.Client
	authClient      *auth.Client
	logger          *slog.Logger
}

func NewUserRepository(client *infra.FirebaseClient, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		firestoreClient: client.Firestore,
		authClient:      client.Auth,
		logger:          logger,
	}
}

type GetUserByIdOptions struct {
	Uid string
}

type GetUserSavedAddressesOptions struct {
	Language models.Language
	Uid      string
}

type GetUserRequestsOptions struct {
	Language models.Language
	Uid      string
}

type UpdateUserActiveRequestCountOptions struct {
	Language models.Language
	UID      string
	Count    int64
}

type UpdateUserSavedAddressesOptions struct {
	Language   models.Language
	UserID     string
	AddressIDs []string
	Action     UpdateAddressBookAction
}

type UpdateUserAddressRequestsOptions struct {
	Language   models.Language
	UserID     string
	RequestIDs []string
	Action     UpdateAddressBookAction
}

/* ---- Admin user repository ---- */

func (u *UserRepository) SignInAdminUserById(ctx context.Context, opts GetUserByIdOptions) (*models.AdminUser, error) {
	tableName := adminUserTable
	logger := u.logger.With(
		"path", "repository.user.SignInAdminUserById",
		"uid", opts.Uid,
	)

	docRef := u.firestoreClient.Collection(tableName).Doc(opts.Uid)
	// get document
	doc, err := docRef.Get(ctx)
	if err != nil {
		logger.Error("failed to get admin user document", "error", err)
		return nil, fmt.Errorf("failed to get user with UID: %w", err)
	}
	// parse data
	var adminUser models.AdminUser
	if err := doc.DataTo(&adminUser); err != nil {
		logger.Error("failed to parse admin user document", "error", err)
		return nil, fmt.Errorf("failed to parse admin user data: %w", err)
	}
	now := time.Now().UnixMilli()
	adminUser.LastLogin = now
	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "lastLogin", Value: now},
	})
	if err != nil {
		logger.Error("failed to sign in admin user", "error", err)
		return nil, fmt.Errorf("failed to sign in admin user: %w", err)
	}
	return &adminUser, nil
}

/* ---- App user repository ---- */

func (u *UserRepository) AuthenticateAppUserById(
	ctx context.Context,
	opts *GetUserByIdOptions) (*models.AppUser, error) {
	tableName := appUserTable
	logger := u.logger.With(
		"path", "repository.user.AuthenticateAppUserById",
		"uid", opts.Uid,
	)

	docRef := u.firestoreClient.Collection(tableName).Doc(opts.Uid)
	// get user document
	doc, err := docRef.Get(ctx)
	// if user not found, create a new user
	if err != nil && status.Code(err) == codes.NotFound {
		appUser, err := u.CreateAppUser(ctx, opts.Uid)
		if err != nil {
			return nil, err
		}
		_, err = docRef.Set(ctx, appUser)
		if err != nil {
			logger.Error("failed to create app user", "error", err)
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		return appUser, nil
	}
	// if error other than not found, return an error
	if err != nil {
		logger.Error("failed to get app user document", "error", err)
		return nil, fmt.Errorf("failed to get user with UID: %w", err)
	}
	// parse data and update last login time
	var appUser models.AppUser
	if err := doc.DataTo(&appUser); err != nil {
		logger.Error("failed to parse app user document", "error", err)
		return nil, fmt.Errorf("failed to parse app user data: %w", err)
	}
	now := time.Now().UnixMilli()
	appUser.LastLogin = now
	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "lastLogin", Value: now},
	})
	if err != nil {
		logger.Error("failed to sign in app user", "error", err)
		return nil, fmt.Errorf("failed to sign in app user: %w", err)
	}
	return &appUser, nil
}

func (u *UserRepository) CreateAppUser(
	ctx context.Context,
	uid string) (*models.AppUser, error) {
	logger := u.logger.With(
		"path", "repository.user.CreateAppUser",
	)
	userRecord, err := u.authClient.GetUser(ctx, uid)
	if err != nil {
		logger.Error("failed to retrieve user info from auth service", "uid", uid, "error", err)
		return nil, fmt.Errorf("failed to retrieve user info from auth service: %w", err)
	}
	now := time.Now().UnixMilli()
	newUser := &models.AppUser{
		Email:       userRecord.Email,
		DisplayName: userRecord.DisplayName,
		CreatedAt:   now,
		LastLogin:   now,
		ImageUrl:    userRecord.PhotoURL,
	}
	return newUser, nil
}

/* ---- User Address Book ---- */

func (u *UserRepository) GetUserSavedAddresses(
	ctx context.Context,
	opts *GetUserSavedAddressesOptions,
) ([]string, error) {
	tableName := appUserTable
	logger := u.logger.With(
		"path", "repository.user.GetUserSavedAddresses",
		"uid", opts.Uid,
		"language", opts.Language,
	)
	docRef := u.firestoreClient.
		Collection(tableName).Doc(opts.Uid).
		Collection(savedAddressesCollection).Doc(opts.Language.Get())
	doc, err := docRef.Get(ctx)
	// if not found, create a doc
	if status.Code(err) == codes.NotFound {
		err = createFirestorePath(ctx, docRef, models.SavedAddresses{IDs: []string{}})
		return []string{}, nil
	}
	if err != nil {
		logger.Error("failed to get app user document", "error", err)
		return nil, fmt.Errorf("failed to get app user document: %w", err)
	}
	var savedAddresses models.SavedAddresses
	if err := doc.DataTo(&savedAddresses); err != nil {
		logger.Error("failed to parse app user document", "error", err)
		return nil, fmt.Errorf("failed to parse app user document: %w", err)
	}
	if savedAddresses.IDs == nil {
		return []string{}, nil
	}
	return savedAddresses.IDs, nil
}

func (u *UserRepository) UpdateUserSavedAddresses(
	ctx context.Context,
	opts *UpdateUserSavedAddressesOptions,
) (string, error) {
	tableName := appUserTable
	logger := u.logger.With(
		"path", "repository.user.UpdateUserSavedAddresses",
		"uid", opts.UserID,
		"language", opts.Language,
		"action", opts.Action,
	)
	docRef := u.firestoreClient.
		Collection(tableName).Doc(opts.UserID).
		Collection(savedAddressesCollection).Doc(opts.Language.Get())

	_, err := docRef.Get(ctx)
	// if not found, create a doc
	if status.Code(err) == codes.NotFound {
		err = createFirestorePath(ctx, docRef, models.SavedAddresses{IDs: []string{}})
	}
	if err != nil {
		logger.Error("failed to get or create saved addresses document", "error", err)
		return "", fmt.Errorf("failed to get or create saved addresses document: %w", err)
	}
	var updateValue any
	ids := make([]interface{}, len(opts.AddressIDs))
	for i, v := range opts.AddressIDs {
		ids[i] = v
	}
	switch opts.Action {
	case Add:
		updateValue = firestore.ArrayUnion(ids...)
	case Delete:
		updateValue = firestore.ArrayRemove(ids...)
	default:
		logger.Error("unsupported update action")
		return "", fmt.Errorf("unsupported update action")
	}
	result, err := docRef.Update(ctx, []firestore.Update{
		{Path: savedAddressesIDsPath, Value: updateValue},
	})
	if err != nil {
		logger.Error("failed to update saved addresses", "error", err)
		return "", fmt.Errorf("failed to update saved addresses: %w", err)
	}
	return fmt.Sprintf("%d", result.UpdateTime.UnixMilli()), nil
}

// Get user's address requests ids with the given language
func (u *UserRepository) GetUserRequests(
	ctx context.Context,
	opts *GetUserRequestsOptions) (*models.UserAddressRequests, error) {
	tableName := appUserTable
	logger := u.logger.With(
		"path", "repository.user.GetUserRequestsIDs",
		"uid", opts.Uid,
		"language", opts.Language,
	)
	docRef := u.firestoreClient.
		Collection(tableName).Doc(opts.Uid).
		Collection(addressRequestCollection).Doc(opts.Language.Get())
	doc, err := docRef.Get(ctx)
	// if doc not existed, create one
	if status.Code(err) == codes.NotFound {
		newRequests := models.UserAddressRequests{IDs: []string{}, ActiveRequestCount: 0}
		err = createFirestorePath(
			ctx,
			docRef,
			newRequests)
		if err != nil {
			logger.Error("failed to  create user address requests doc", "error", err)
			return nil, fmt.Errorf("failed to create user address request doc: %w", err)
		}
		// early return an empty list if the doc just created
		return &newRequests, nil
	}
	if err != nil {
		logger.Error("failed to get user address requests doc", "error", err)
		return nil, fmt.Errorf("failed to get user address request doc: %w", err)
	}
	var addressRequests models.UserAddressRequests
	if err := doc.DataTo(&addressRequests); err != nil {
		logger.Error("failed to parse user address request doc", "error", err)
		return nil, fmt.Errorf("failed to parse user address request doc: %w", err)
	}
	return &addressRequests, nil
}

// This function only updates the ids.
// because the updates can add/remove multiple items with different status at the same time
// the user's update request count will be handled by
// create request and update request functions (not implemented so far)
func (u *UserRepository) UpdateUserAddressRequests(
	ctx context.Context,
	opts *UpdateUserAddressRequestsOptions,
) (string, error) {
	tableName := appUserTable
	logger := u.logger.With(
		"path", "repository.user.UpdateUserAddressRequests",
		"uid", opts.UserID,
		"language", opts.Language,
		"action", opts.Action,
	)

	docRef := u.firestoreClient.
		Collection(tableName).Doc(opts.UserID).
		Collection(addressRequestCollection).Doc(opts.Language.Get())

	_, err := docRef.Get(ctx)
	// if doc not existed, create one
	if status.Code(err) == codes.NotFound {
		err = createFirestorePath(
			ctx,
			docRef,
			models.UserAddressRequests{IDs: []string{}, ActiveRequestCount: 0})
	}
	if err != nil {
		logger.Error("failed to update user address request", "error", err)
		return "", fmt.Errorf("failed to update user address request: %w", err)
	}
	var updateValue any
	ids := make([]interface{}, len(opts.RequestIDs))
	for i, v := range opts.RequestIDs {
		ids[i] = v
	}
	switch opts.Action {
	case Add:
		updateValue = firestore.ArrayUnion(ids...)
	case Delete:
		updateValue = firestore.ArrayRemove(ids...)
	default:
		logger.Error("unsupported update action")
		return "", fmt.Errorf("unsupported update action")
	}
	result, err := docRef.Update(ctx, []firestore.Update{
		{Path: requestIDsPath, Value: updateValue},
	})
	if err != nil {
		logger.Error("failed to update requests", "error", err)
		return "", fmt.Errorf("failed to update requests: %w", err)
	}
	return fmt.Sprintf("%d", result.UpdateTime.UnixMilli()), nil
}

func (u *UserRepository) UpdateUserActiveRequestCount(
	ctx context.Context, opts *UpdateUserActiveRequestCountOptions) error {
	tableName := appUserTable
	logger := u.logger.With(
		"path", "repository.user.UpdateUserActiveRequestCount",
		"uid", opts.UID,
		"language", opts.Language,
		"count", opts.Count,
	)
	docRef := u.firestoreClient.
		Collection(tableName).Doc(opts.UID).
		Collection(addressRequestCollection).Doc(opts.Language.Get())
	_, err := docRef.Update(ctx, []firestore.Update{{Path: activeRequestCountPath, Value: opts.Count}})
	if err != nil {
		logger.Error("failed to update user's active request count", "error", err)
		return fmt.Errorf("failed to update user's active request count: %w", err)
	}
	return nil
}

// ---------- helper functions ----------

func createFirestorePath(ctx context.Context, docRef *firestore.DocumentRef, defaultValue interface{}) error {
	if _, err := docRef.Set(ctx, defaultValue); err != nil {
		return fmt.Errorf("failed to create firestore document: %w", err)
	}
	return nil
}
