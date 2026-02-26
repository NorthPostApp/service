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
	"google.golang.org/genai"
)

type LLMClient struct {
	OpenAI *openai.Client
	Gemini *genai.Client
	logger *slog.Logger // LLM needs a logger for process monitoring
}

func NewLLMClient(logger *slog.Logger) (*LLMClient, error) {
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	if openaiApiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}
	if geminiApiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}
	openAIClient := openai.NewClient(option.WithAPIKey(openaiApiKey))
	logger.Info("OpenAI client initialized successfully")
	geminiClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{APIKey: geminiApiKey})
	if err != nil {
		logger.Error("failed to initialize Gemini client", "error", err)
		return nil, fmt.Errorf("failed to initialize Gemini: %w", err)
	}
	logger.Info("Gemini client initialized successfully")
	return &LLMClient{OpenAI: &openAIClient, Gemini: geminiClient, logger: logger}, nil
}

type StructuredCompletionOptions struct {
	Prompt          string
	SystemPrompt    string
	Model           string
	ReasoningEffort string
	ThinkingLevel   string
}

type OpenAIStructuredCompletionOptions struct {
	Prompt          string
	SystemPrompt    string
	SchemaName      string
	Description     string
	Model           string
	ReasoningEffort openai.ReasoningEffort
}

type GeminiStructuredCompletionOptions struct {
	Prompt        string
	SystemPrompt  string
	Model         string
	ThinkingLevel genai.ThinkingLevel
}

// The structured completion wrapper function to help route the generation task to corresponding LLM provider
func (l *LLMClient) StructuredCompletion(
	ctx context.Context,
	opts StructuredCompletionOptions,
	schemaInstance interface{},
	result interface{}) error {
	model := opts.Model
	// Check model type
	if !strings.HasPrefix(model, "gpt") && !strings.HasPrefix(model, "gemini") {
		l.logger.Error("invalid model name", "error", "The model should start with either gpt or gemini", "model", model)
		return fmt.Errorf("invalid model name, model name should start with either 'gpt' or 'gemini'")
	}
	// Avoid empty prompt
	if strings.TrimSpace(opts.Prompt) == "" {
		l.logger.Error("invalid prompt", "error", "the prompt shouldn't be empty")
		return fmt.Errorf("invalid prompt, the prompt shouldn't be an empty string")
	}
	schema, err := gptschema.GenerateSchema(schemaInstance)
	if err != nil {
		l.logger.Error("failed to generate schema", "error", err)
		return fmt.Errorf("failed to generate schema: %w", err)
	}
	if strings.HasPrefix(model, "gemini") {
		l.logger.Info("Address generation request", "model", model, "thinking level", opts.ThinkingLevel)
		geminiOpts := GeminiStructuredCompletionOptions{
			Prompt:        opts.Prompt,
			SystemPrompt:  opts.SystemPrompt,
			Model:         model,
			ThinkingLevel: genai.ThinkingLevel(opts.ThinkingLevel),
		}
		err := l.StructuredCompletionGemini(ctx, geminiOpts, schema, result)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(model, "gpt") {
		l.logger.Info("Address generation request", "model", model, "reasoning effort", opts.ReasoningEffort)
		gptOpts := OpenAIStructuredCompletionOptions{
			SystemPrompt:    opts.SystemPrompt,
			Prompt:          opts.Prompt,
			SchemaName:      "address_generation",
			Description:     "Generate a structured address with metadata",
			Model:           opts.Model,
			ReasoningEffort: openai.ReasoningEffort(opts.ReasoningEffort),
		}
		err := l.StructuredCompletionOpenAI(ctx, gptOpts, schema, result)
		if err != nil {
			return err
		}
	}
	return nil
}

// Generate structured contents with Gemini models
func (l *LLMClient) StructuredCompletionGemini(ctx context.Context, opts GeminiStructuredCompletionOptions, schema interface{}, result interface{}) error {
	model := opts.Model
	config := &genai.GenerateContentConfig{
		ResponseMIMEType:   "application/json",
		ResponseJsonSchema: schema,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: opts.SystemPrompt},
			},
		},
	}
	// Only add thinking level config to the gemini 3 series models
	if strings.HasPrefix(model, "gemini-3") {
		config.ThinkingConfig = &genai.ThinkingConfig{
			IncludeThoughts: false,
			ThinkingLevel:   opts.ThinkingLevel,
		}
	}
	output, err := l.Gemini.Models.GenerateContent(ctx, model, genai.Text(opts.Prompt), config)
	if err != nil {
		l.logger.Error("failed to generate content", "model", model, "error", err)
		return fmt.Errorf("failed to generate content: %w", err)
	}
	if len(output.Candidates) == 0 {
		l.logger.Error("no candidates returned from gemini")
		return fmt.Errorf("no candidates returned")
	}
	jsonContent := output.Text()
	if jsonContent == "" {
		l.logger.Error("empty content returned from gemini")
		return fmt.Errorf("empty content returned from gemini")
	}
	if err := json.Unmarshal([]byte(jsonContent), result); err != nil {
		l.logger.Error("failed to unmarshal response", "error", err, "content", jsonContent)
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

// Generate structured contents with OpenAI models
func (l *LLMClient) StructuredCompletionOpenAI(
	ctx context.Context,
	opts OpenAIStructuredCompletionOptions,
	schema interface{},
	result interface{},
) error {
	// Set default model and effort if not provided
	model := opts.Model
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
	messages = append(messages, openai.UserMessage(opts.Prompt))
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
	chat, err := l.OpenAI.Chat.Completions.New(ctx, completionParams)
	if err != nil {
		l.logger.Error("chat completion failed", "model", model, "error", err)
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
