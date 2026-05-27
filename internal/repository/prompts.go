package repository

import (
	"context"
	"fmt"
	"log/slog"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/infra"

	"cloud.google.com/go/firestore"
)

const (
	promptTable          = "prompts"
	addressGenerationKey = "address_generation"
)

type PromptRepository struct {
	client *firestore.Client
	logger *slog.Logger
}

func NewPromptRepository(client *infra.FirebaseClient, logger *slog.Logger) *PromptRepository {
	return &PromptRepository{
		client: client.Firestore,
		logger: logger,
	}
}

type GetSystemPromptOptions struct {
	Language models.Language
	Key      string
}

type GetSystemAddressGenerationPromptOptions struct {
	Language models.Language
}

// get prompt by language and key
func (r *PromptRepository) GetSystemPrompt(
	ctx context.Context, opts GetSystemPromptOptions) (string, error) {
	logger := r.logger.With(
		"path", "repository.prompts.GetSystemPrompt",
		"key", opts.Key,
		"language", opts.Language,
	)
	language := opts.Language
	key := opts.Key
	docRef := r.client.Collection(promptTable).Doc(getPromptLanguage(language))
	doc, err := docRef.Get(ctx)
	if err != nil {
		logger.Error("failed to get prompt", "error", err)
		return "", fmt.Errorf("failed to get prompt")
	}
	data := doc.Data()
	prompt, ok := data[key].(string)
	if !ok {
		logger.Warn("prompt key missing or not a string")
		return "", fmt.Errorf("prompt key missing or not a string")
	}
	return prompt, nil
}

// get address generation system prompt
func (r *PromptRepository) GetSystemAddressGenerationPrompt(
	ctx context.Context,
	opts GetSystemAddressGenerationPromptOptions) (string, error) {
	getPromptOpts := GetSystemPromptOptions{
		Language: opts.Language,
		Key:      addressGenerationKey,
	}
	return r.GetSystemPrompt(ctx, getPromptOpts)
}

// ============ Helper functions ===========
func getPromptLanguage(language models.Language) string {
	if err := language.Validate(); err != nil {
		return models.LanguageEN.Get() // set fallback language as english
	}
	return language.Get()
}
