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
	docRef := u.firestoreClient.Collection(tableName).Doc(opts.Uid)
	// get document
	doc, err := docRef.Get(ctx)
	if err != nil {
		u.logger.Error("failed to get admin user document", "uid", opts.Uid, "error", err)
		return nil, fmt.Errorf("failed to get user with UID: %w", err)
	}
	// parse data
	var adminUser models.AdminUser
	if err := doc.DataTo(&adminUser); err != nil {
		u.logger.Error("failed to parse admin user document", "uid", opts.Uid, "error", err)
		return nil, fmt.Errorf("failed to parse admin user data: %w", err)
	}
	now := time.Now().UnixMilli()
	adminUser.LastLogin = now
	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "lastLogin", Value: now},
	})
	if err != nil {
		u.logger.Error("failed to sign in admin user", "error", err)
		return nil, fmt.Errorf("failed to sign in admin user: %w", err)
	}
	return &adminUser, nil
}

/* ---- App user repository ---- */

func (u *UserRepository) AuthenticateAppUserById(
	ctx context.Context,
	opts *GetUserByIdOptions) (*models.AppUser, error) {
	tableName := appUserTable
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
			u.logger.Error("failed to create app user", "uid", opts.Uid, "error", err)
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		return appUser, nil
	}
	// if error other than not found, return an error
	if err != nil {
		u.logger.Error("failed to get app user document", "uid", opts.Uid)
		return nil, fmt.Errorf("failed to get user with UID: %w", err)
	}
	// parse data and update last login time
	var appUser models.AppUser
	if err := doc.DataTo(&appUser); err != nil {
		u.logger.Error("failed to parse app user document", "uid", opts.Uid, "error", err)
		return nil, fmt.Errorf("failed to parse app user data: %w", err)
	}
	now := time.Now().UnixMilli()
	appUser.LastLogin = now
	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "lastLogin", Value: now},
	})
	if err != nil {
		u.logger.Error("failed to sign in app user", "error", err)
		return nil, fmt.Errorf("failed to sign in app user: %w", err)
	}
	return &appUser, nil
}

func (u *UserRepository) CreateAppUser(
	ctx context.Context,
	uid string) (*models.AppUser, error) {
	userRecord, err := u.authClient.GetUser(ctx, uid)
	if err != nil {
		u.logger.Error("failed to retrieve user info from auth service", "uid", uid, "error", err)
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
	docRef := u.firestoreClient.
		Collection(tableName).Doc(opts.Uid).
		Collection(savedAddressesCollection).Doc(opts.Language.Get())
	doc, err := docRef.Get(ctx)
	// if not found, create a doc
	if status.Code(err) == codes.NotFound {
		err = createFirestorePath(ctx, docRef, models.SavedAddresses{IDs: []string{}})
	}
	if err != nil {
		u.logger.Error("failed to get app user document",
			"uid", opts.Uid,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get app user document: %w", err)
	}
	var savedAddresses models.SavedAddresses
	if err := doc.DataTo(&savedAddresses); err != nil {
		u.logger.Error("failed to parse app user document",
			"uid", opts.Uid,
			"error", err,
		)
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
	docRef := u.firestoreClient.
		Collection(tableName).Doc(opts.UserID).
		Collection(savedAddressesCollection).Doc(opts.Language.Get())

	_, err := docRef.Get(ctx)
	// if not found, create a doc
	if status.Code(err) == codes.NotFound {
		err = createFirestorePath(ctx, docRef, models.SavedAddresses{IDs: []string{}})
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
		u.logger.Error("unsupported update action",
			"path", "ids",
			"action", opts.Action,
		)
		return "", fmt.Errorf("unsupported update action")
	}
	result, err := docRef.Update(ctx, []firestore.Update{
		{Path: savedAddressesIDsPath, Value: updateValue},
	})
	if err != nil {
		u.logger.Error(
			"failed to update saved addresses",
			"uid", opts.UserID,
			"language", opts.Language,
			"error", err)
		return "", fmt.Errorf("failed to update saved addresses: %w", err)
	}
	return fmt.Sprintf("%d", result.UpdateTime.UnixMilli()), nil
}

// This function only updates the ids.
// because the updates can add/remove multiple items with different status
// the user's update request count will be handled by
// another function
func (u *UserRepository) UpdateUserAddressRequests(
	ctx context.Context,
	opts *UpdateUserAddressRequestsOptions,
) (string, error) {
	tableName := appUserTable
	docRef := u.firestoreClient.
		Collection(tableName).Doc(opts.UserID).
		Collection(addressRequestCollection).Doc(opts.Language.Get())

	_, err := docRef.Get(ctx)
	// if doc not existed, create one
	if status.Code(err) == codes.NotFound {
		err = createFirestorePath(
			ctx,
			docRef,
			models.AddressRequests{IDs: []string{}, ActiveRequestCount: 0})
	}
	if err != nil {
		u.logger.Error(
			"failed to update user address request",
			"path", "repository.user.UpdateUserAddressRequest",
			"action", opts.Action,
			"uid", opts.UserID,
			"language", opts.Language,
			"error", err,
		)
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
		u.logger.Error(
			"unsupported update action",
			"path", "repository.user.UpdateUserAddressRequest",
			"uid", opts.UserID,
			"language", opts.Language,
			"error", err,
		)
		return "", fmt.Errorf("unsupported update action")
	}
	result, err := docRef.Update(ctx, []firestore.Update{
		{Path: savedAddressesIDsPath, Value: updateValue},
	})
	if err != nil {
		u.logger.Error(
			"failed to update requests",
			"path", "repository.user.UpdateUserAddressRequest",
			"uid", opts.UserID,
			"language", opts.Language,
			"error", err,
		)
		return "", fmt.Errorf("failed to update requests: %w", err)
	}
	return fmt.Sprintf("%d", result.UpdateTime.UnixMilli()), nil
}

// ---------- helper functions ----------

func createFirestorePath(ctx context.Context, docRef *firestore.DocumentRef, defaultValue interface{}) error {
	if _, err := docRef.Set(ctx, defaultValue); err != nil {
		return fmt.Errorf("failed to create firestore document: %w", err)
	}
	return nil
}
