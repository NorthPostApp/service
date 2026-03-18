package infra

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageBucketClient struct {
	R2Storage   *s3.Client
	R2Presigned *s3.PresignClient
}

func NewStorageBucketClient(logger *slog.Logger) (*StorageBucketClient, error) {
	accountId := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	if accountId == "" {
		return nil, fmt.Errorf("CLOUDFLARE_ACCOUNT_ID environment variable is required")
	}
	r2AccessKey := os.Getenv("CLOUDFLARE_R2_ACCESS_KEY")
	if r2AccessKey == "" {
		return nil, fmt.Errorf("CLOUDFLARE_R2_ACCESS_KEY environment variable is required")
	}
	r2SecretKey := os.Getenv("CLOUDFLARE_R2_SECRET_KEY")
	if r2SecretKey == "" {
		return nil, fmt.Errorf("CLOUDFLARE_R2_SECRET_KEY environment variable is required")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r2AccessKey, r2SecretKey, "")),
		config.WithRegion("auto"), // Required by SDK but not used by R2
	)
	if err != nil {
		logger.Error("failed to load R2 configuration", "error", err)
		return nil, fmt.Errorf("failed to initialize R2: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})
	presignedClient := s3.NewPresignClient(client)
	logger.Info("Storage bucket clients initialized successfully")
	return &StorageBucketClient{
		R2Storage:   client,
		R2Presigned: presignedClient,
	}, nil
}
