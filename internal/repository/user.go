package repository

import (
	"context"
	"fmt"
	"log/slog"
	"north-post/service/internal/domain/v1/models"
	"time"

	"cloud.google.com/go/firestore"
)

const (
	adminUserTable = "admin_users"
)

type UserRepository struct {
	client *firestore.Client
	logger *slog.Logger
}

func NewUserRepository(client *firestore.Client, logger *slog.Logger) *UserRepository {
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
	docRef := u.client.Collection(tableName).Doc(opts.Uid)
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
	now := time.Now().Unix()
	adminUser.LastLogin = now
	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "lastLogin", Value: now},
	})
	if err != nil {
		u.logger.Error("failed to sign in user", "error", err)
		return nil, fmt.Errorf("failed to sign in user: %w", err)
	}
	return &adminUser, nil
}
