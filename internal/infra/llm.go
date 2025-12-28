package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/akane9506/gptschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type LLMClient struct {
	LLM    *openai.Client
	logger *slog.Logger // LLM needs a logger for process monitoring
}

func NewLLMClient(logger *slog.Logger) (*LLMClient, error) {
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	if openaiApiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}
	client := openai.NewClient(option.WithAPIKey(openaiApiKey))
	logger.Info("llm client initialized successfully")
	return &LLMClient{LLM: &client, logger: logger}, nil
}

type StructuredCompletionOptions struct {
	Prompt          string
	SystemPrompt    string
	SchemaName      string
	Description     string
	Model           openai.ChatModel
	ReasoningEffort openai.ReasoningEffort
	// we might need these two params in the future
	// Temperature *float64
	// MaxTokens   *int64
}

func (l *LLMClient) StructuredCompletion(
	ctx context.Context,
	opts StructuredCompletionOptions,
	schemaInstance interface{},
	result interface{},
) error {
	schema, err := gptschema.GenerateSchema(schemaInstance)
	if err != nil {
		return fmt.Errorf("failed to generate schema: %w", err)
	}
	// Set default model and effort if not provided
	model := opts.Model
	if model == "" {
		model = openai.ChatModelGPT5Mini
	}
	// Build the schema parameter
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        opts.SchemaName,
		Description: openai.String(opts.Description),
		Schema:      schema,
		Strict:      openai.Bool(true),
	}
	// Build completion parameters
	messages := []openai.ChatCompletionMessageParamUnion{}
	if strings.TrimSpace(opts.SystemPrompt) != "" {
		messages = append(messages, openai.SystemMessage(opts.SystemPrompt))
	}
	if strings.TrimSpace(opts.Prompt) != "" {
		messages = append(messages, openai.UserMessage(opts.Prompt))
	}
	if len(messages) == 0 {
		l.logger.Error("no system and user message provided")
		return fmt.Errorf("no user or system message provided")
	}
	completionParams := openai.ChatCompletionNewParams{
		Messages: messages,
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: schemaParam},
		},
		Model: model,
	}
	// Only gpt 5 models supports reasoning effort
	if strings.HasPrefix(model, "gpt-5") {
		effort := opts.ReasoningEffort
		if effort == "" {
			effort = "low"
		}
		completionParams.ReasoningEffort = effort
	}
	// query the chat completion API
	chat, err := l.LLM.Chat.Completions.New(ctx, completionParams)
	if err != nil {
		l.logger.Error("chat completion failed", "error", err)
		return fmt.Errorf("chat completion failed: %w", err)
	}
	// check if we got a valid response
	if len(chat.Choices) == 0 {
		l.logger.Error("no choices returned from completion")
		return fmt.Errorf("no choices returned from completion")
	}
	// parse JSON response into the target struct
	content := chat.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), result); err != nil {
		l.logger.Error("failed to unmarshal response", "error", err, "content", content)
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}
