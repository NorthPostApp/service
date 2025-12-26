package services

import (
	"context"
	"fmt"

	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
)

type promptRepository interface {
	GetSystemPrompt(ctx context.Context, opts repository.GetSystemPromptOptions) (string, error)
	GetSystemAddressGenerationPrompt(ctx context.Context, opts repository.GetSystemAddressGenerationPromptOptions) (string, error)
}

type PromptService struct {
	repo promptRepository
}

func NewPromptService(repo promptRepository) *PromptService {
	return &PromptService{repo: repo}
}

type GetSystemAddressGenerationPromptInput struct {
	Language models.Language
}

type GetSystemAddressGenerationPromptOutput struct {
	Prompt string
}

func (p *PromptService) GetSystemAddressGenerationPrompt(
	ctx context.Context,
	input GetSystemAddressGenerationPromptInput) (*GetSystemAddressGenerationPromptOutput, error) {
	opts := repository.GetSystemAddressGenerationPromptOptions{
		Language: input.Language,
	}
	prompt, err := p.repo.GetSystemAddressGenerationPrompt(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &GetSystemAddressGenerationPromptOutput{Prompt: prompt}, nil
}
