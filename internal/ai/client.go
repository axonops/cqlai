package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
)

// AIClient defines the interface for an AI client.
type AIClient interface {
	GenerateCQLWithTools(ctx context.Context, prompt string, schema string) (*QueryPlan, error)
	SetAPIKey(key string)
}

// ConvertDBConfigToAIConfig converts config.AIConfig to local AIConfig for the AI client
func ConvertDBConfigToAIConfig(dbConfig *config.AIConfig) *AIConfig {
	logger.DebugfToFile("AI", "ConvertDBConfigToAIConfig called")

	if dbConfig == nil {
		logger.DebugfToFile("AI", "dbConfig is nil, returning mock config")
		// Return default mock config if no AI config provided
		return &AIConfig{
			Provider: "mock",
			APIKey:   "",
			Model:    "",
		}
	}

	logger.DebugfToFile("AI", "dbConfig.Provider: %s", dbConfig.Provider)

	config := &AIConfig{
		Provider: dbConfig.Provider,
		APIKey:   dbConfig.APIKey,
		Model:    dbConfig.Model,
	}

	if config.Provider == "" {
		logger.DebugfToFile("AI", "Provider is empty, defaulting to mock")
		config.Provider = "mock" // Default to mock for safety
	}

	// Use provider-specific config if available
	switch Provider(config.Provider) {
	case ProviderOpenAI:
		if dbConfig.OpenAI != nil {
			if dbConfig.OpenAI.APIKey != "" {
				config.APIKey = dbConfig.OpenAI.APIKey
			}
			if dbConfig.OpenAI.Model != "" {
				config.Model = dbConfig.OpenAI.Model
			}
		}
		if config.Model == "" {
			config.Model = DefaultOpenAIModel
		}
	case ProviderAnthropic:
		logger.DebugfToFile("AI", "Processing Anthropic config")
		if dbConfig.Anthropic != nil {
			logger.DebugfToFile("AI", "Anthropic config exists: APIKey=%v, Model=%s",
				dbConfig.Anthropic.APIKey != "", dbConfig.Anthropic.Model)
			if dbConfig.Anthropic.APIKey != "" {
				config.APIKey = dbConfig.Anthropic.APIKey
				logger.DebugfToFile("AI", "Set Anthropic API key")
			}
			if dbConfig.Anthropic.Model != "" {
				config.Model = dbConfig.Anthropic.Model
				logger.DebugfToFile("AI", "Set Anthropic model: %s", config.Model)
			}
		}
		if config.Model == "" {
			config.Model = DefaultAnthropicModel
			logger.DebugfToFile("AI", "Using default Anthropic model: %s", config.Model)
		}
	case ProviderGemini:
		if dbConfig.Gemini != nil {
			if dbConfig.Gemini.APIKey != "" {
				config.APIKey = dbConfig.Gemini.APIKey
			}
			if dbConfig.Gemini.Model != "" {
				config.Model = dbConfig.Gemini.Model
			}
		}
		if config.Model == "" {
			config.Model = DefaultGeminiModel
		}
	case ProviderOllama:
		if dbConfig.Ollama != nil {
			if dbConfig.Ollama.APIKey != "" {
				config.APIKey = dbConfig.Ollama.APIKey
			}
			if dbConfig.Ollama.Model != "" {
				config.Model = dbConfig.Ollama.Model
			}
		}
		if config.Model == "" {
			config.Model = DefaultOllamaModel
		}
	}

	logger.DebugfToFile("AI", "Final AI config: Provider=%s, HasAPIKey=%v, Model=%s",
		config.Provider, config.APIKey != "", config.Model)

	return config
}

// AIConfig holds configuration for AI providers
type AIConfig struct {
	Provider string // "gemini", "openai", "anthropic", "mock"
	APIKey   string
	Model    string // Optional model override
}

// BaseAIClient provides common functionality
type BaseAIClient struct {
	APIKey string
	Model  string
}

func (c *BaseAIClient) SetAPIKey(key string) {
	c.APIKey = key
}

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// extractJSON attempts to extract JSON from a text response
func extractJSON(text string) string {
	// Look for JSON between ```json and ``` markers
	startMarker := JSONStartMarker
	endMarker := JSONEndMarker
	startIdx := strings.Index(text, startMarker)
	if startIdx != -1 {
		startIdx += len(startMarker)
		endIdx := strings.Index(text[startIdx:], endMarker)
		if endIdx != -1 {
			return strings.TrimSpace(text[startIdx : startIdx+endIdx])
		}
	}

	// Look for JSON between { and }
	startIdx = strings.Index(text, "{")
	if startIdx != -1 {
		endIdx := strings.LastIndex(text, "}")
		if endIdx != -1 && endIdx > startIdx {
			return text[startIdx : endIdx+1]
		}
	}

	return ""
}

// createAIClient creates the appropriate AI client based on configuration
func createAIClient(aiConfig *AIConfig, providerConfig *config.AIConfig) (AIClient, error) {
	if aiConfig == nil {
		return nil, fmt.Errorf("AI configuration is required")
	}

	switch Provider(aiConfig.Provider) {
	case ProviderOpenAI:
		return NewOpenAIClient(aiConfig.APIKey, aiConfig.Model), nil
	case ProviderAnthropic:
		return NewAnthropicClient(aiConfig.APIKey, aiConfig.Model), nil
	case ProviderOllama:
		if providerConfig == nil || providerConfig.Ollama == nil {
			return nil, fmt.Errorf("Ollama configuration is missing from cqlai.json")
		}
		return NewOllamaClient(providerConfig.Ollama), nil
	default:
		// Add a mock client for safety, so the app doesn't crash if config is missing
		if aiConfig.Provider == "mock" {
			return &MockAIClient{}, nil
		}
		return nil, fmt.Errorf(ErrInvalidProvider, aiConfig.Provider)
	}
}

// MockAIClient is a mock client for testing and safe fallback.
type MockAIClient struct{}

func (m *MockAIClient) GenerateCQLWithTools(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	return nil, fmt.Errorf(ErrUnsupportedMethod, "mock", "tool calling")
}

func (m *MockAIClient) SetAPIKey(key string) {
	// No-op for mock client
}

// GenerateCQLFromRequest is a high-level function that combines schema fetching and CQL generation
func GenerateCQLFromRequest(ctx context.Context, session *db.Session, aiConfig *config.AIConfig, userRequest string) (*QueryPlan, string, error) {
	// Initialize local AI for fuzzy search if needed
	if err := InitializeLocalAI(session); err != nil {
		// Continue without local AI
	}

	// Get minimal schema context (just list of keyspaces for initial context)
	schemaContext := "Available keyspaces: "
	if globalAI != nil && globalAI.cache != nil {
		globalAI.cache.Mu.RLock()
		if len(globalAI.cache.Keyspaces) > 0 {
			schemaContext += strings.Join(globalAI.cache.Keyspaces[:min(10, len(globalAI.cache.Keyspaces))], ", ")
			if len(globalAI.cache.Keyspaces) > 10 {
				schemaContext += fmt.Sprintf(" (and %d more)", len(globalAI.cache.Keyspaces)-10)
			}
		}
		globalAI.cache.Mu.RUnlock()
	} else {
		// Fallback to getting schema from session
		sc, err := session.GetSchemaContext(20)
		if err == nil {
			schemaContext = sc
		}
	}

	// Convert config to local AI config
	localConfig := ConvertDBConfigToAIConfig(aiConfig)
	logger.DebugfToFile("AI", "AI Config: provider=%s, has_api_key=%v", localConfig.Provider, localConfig.APIKey != "")

	// Create AI client
	client, err := createAIClient(localConfig, aiConfig)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create AI client: %v", err)
	}
	logger.DebugfToFile("AI", "Created AI client of type: %T", client)

	// Generate with tools (all clients support tools now)
	logger.DebugfToFile("AI", "Attempting to generate CQL with tools for: %s", userRequest)
	plan, err := client.GenerateCQLWithTools(ctx, userRequest, schemaContext)
	if err != nil {
		// Check if this is an interaction request
		if _, ok := err.(*InteractionRequest); ok {
			logger.DebugfToFile("AI", "User interaction needed, returning to UI")
			return nil, "", err // Return the request as-is to preserve type
		}

		logger.DebugfToFile("AI", "GenerateCQLWithTools failed: %v", err)
		return nil, "", fmt.Errorf("failed to generate CQL plan: %v", err)
	}
	logger.DebugfToFile("AI", "GenerateCQLWithTools succeeded")

	// Validate plan
	validator := &PlanValidator{Schema: nil} // TODO: Pass actual schema
	if err := validator.ValidatePlan(plan); err != nil {
		return nil, "", fmt.Errorf("invalid plan: %v", err)
	}

	// Render CQL
	cql, err := RenderCQL(plan)
	if err != nil {
		return nil, "", fmt.Errorf("failed to render CQL: %v", err)
	}

	return plan, cql, nil
}

// FormatPlanAsJSON returns a pretty-printed JSON representation of the plan
func FormatPlanAsJSON(plan *QueryPlan) string {
	b, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting plan: %v", err)
	}
	return string(b)
}
