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
	adminUserTable = "admin_users"
	appUserTable   = "app_users"
)

type UserRepository struct {
	client *infra.FirebaseClient
	logger *slog.Logger
}

func NewUserRepository(client *infra.FirebaseClient, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		client: client,
		logger: logger,
	}
}

type GetUserByIdOptions struct {
	Uid string
}

func (u *UserRepository) SignInAdminUserById(ctx context.Context, opts GetUserByIdOptions) (*models.AdminUser, error) {
	tableName := adminUserTable
	docRef := u.client.Firestore.Collection(tableName).Doc(opts.Uid)
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

// func (u *UserRepository) SignInAppUserById(
// 	ctx context.Context,
// 	opts GetUserByIdOptions) (*models.AppUser, error) {
// 	tableName := appUserTable
// 	docRef := u.client.Firestore.Collection(tableName).Doc(opts.Uid)
// 	// get user document
// 	doc, err := docRef.Get(ctx)
// 	if err != nil {
// 		u.logger.Error("failed to get app user document", "uid", opts.Uid)
// 		return nil, fmt.Errorf("failed to get user with UID: %w", err)
// 	}
// 	// parse data
// 	var appUser models.AppUser
// 	if err := doc.DataTo(&appUser); err != nil {
// 		u.logger.Error("failed to parse app user document", "uid", opts.Uid, "error", err)
// 		return nil, fmt.Errorf("failed to parse app user data: %w", err)
// 	}
// 	now := time.Now().UnixMilli()
// 	appUser.LastLogin = now
// 	_, err = docRef.Update(ctx, []firestore.Update{
// 		{Path: "lastLogin", Value: now},
// 	})
// 	if err != nil {
// 		u.logger.Error("failed to sign in app user", "error", err)
// 		return nil, fmt.Errorf("failed to sign in app user: %w", err)
// 	}
// 	return nil, nil
// }

// func (u *UserRepository) SignUpAppUser(
// 	ctx context.Context,
// 	opts GetUserByIdOptions) (*models.AppUser, error) {

// 	return nil, nil
// }
