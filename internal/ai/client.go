package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/db"
)

// AIClient defines the interface for an AI client.
type AIClient interface {
	GenerateCQL(ctx context.Context, prompt string, schema string) (*QueryPlan, error)
	SetAPIKey(key string)
}

// ConvertDBConfigToAIConfig converts db.AIConfig to local AIConfig for the AI client
func ConvertDBConfigToAIConfig(dbConfig *db.AIConfig) *AIConfig {
	if dbConfig == nil {
		// Return default mock config if no AI config provided
		return &AIConfig{
			Provider: "mock",
			APIKey:   "",
			Model:    "",
		}
	}
	
	config := &AIConfig{
		Provider: dbConfig.Provider,
		APIKey:   dbConfig.APIKey,
		Model:    dbConfig.Model,
	}
	
	if config.Provider == "" {
		config.Provider = "mock" // Default to mock for safety
	}
	
	// Use provider-specific config if available
	switch config.Provider {
	case "openai":
		if dbConfig.OpenAI != nil {
			if dbConfig.OpenAI.APIKey != "" {
				config.APIKey = dbConfig.OpenAI.APIKey
			}
			if dbConfig.OpenAI.Model != "" {
				config.Model = dbConfig.OpenAI.Model
			}
		}
		if config.Model == "" {
			config.Model = "gpt-4-turbo-preview"
		}
	case "anthropic":
		if dbConfig.Anthropic != nil {
			if dbConfig.Anthropic.APIKey != "" {
				config.APIKey = dbConfig.Anthropic.APIKey
			}
			if dbConfig.Anthropic.Model != "" {
				config.Model = dbConfig.Anthropic.Model
			}
		}
		if config.Model == "" {
			config.Model = "claude-3-sonnet-20240229"
		}
	case "gemini":
		if dbConfig.Gemini != nil {
			if dbConfig.Gemini.APIKey != "" {
				config.APIKey = dbConfig.Gemini.APIKey
			}
			if dbConfig.Gemini.Model != "" {
				config.Model = dbConfig.Gemini.Model
			}
		}
		if config.Model == "" {
			config.Model = "gemini-pro"
		}
	}
	
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

// buildPrompt creates the system and user prompts for the LLM
func buildPrompt(userRequest string, schemaContext string) (string, string) {
	systemPrompt := `You are a CQL (Cassandra Query Language) expert assistant. 
Your task is to convert natural language requests into structured query plans.

Rules:
1. Return ONLY valid JSON representing a QueryPlan
2. Be conservative - prefer read-only operations unless explicitly asked to modify
3. Use appropriate CQL data types and syntax
4. Include warnings for destructive operations
5. Set confidence level (0.0-1.0) based on clarity of request
6. If the request is ambiguous, choose the safest interpretation

QueryPlan JSON structure:
{
  "operation": "SELECT|INSERT|UPDATE|DELETE|CREATE|ALTER|DROP",
  "keyspace": "keyspace_name",
  "table": "table_name",
  "columns": ["col1", "col2"],
  "values": {"col1": "value1"},
  "where": [{"column": "id", "operator": "=", "value": 123}],
  "order_by": [{"column": "timestamp", "order": "DESC"}],
  "limit": 10,
  "allow_filtering": false,
  "confidence": 0.95,
  "warning": "optional warning message",
  "read_only": true
}`

	userPrompt := fmt.Sprintf(`Schema Context:
%s

User Request: %s

Generate a QueryPlan JSON for this request.`, schemaContext, userRequest)
	
	return systemPrompt, userPrompt
}

// --- Mock Client (for testing) ---
type MockClient struct {
	BaseAIClient
}

func (c *MockClient) GenerateCQL(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	// Generate a simple mock plan based on keywords in the prompt
	plan := &QueryPlan{
		Operation:  "SELECT",
		Confidence: 1.0,
		ReadOnly:   true,
	}
	
	lowerPrompt := strings.ToLower(prompt)
	
	// Try to extract table name from prompt
	if strings.Contains(lowerPrompt, "users") {
		plan.Table = "users"
		plan.Columns = []string{"*"}
	} else if strings.Contains(lowerPrompt, "products") {
		plan.Table = "products"
		plan.Columns = []string{"*"}
	} else {
		// Generic response
		plan.Table = "table_name"
		plan.Columns = []string{"*"}
		plan.Warning = "Mock response - please configure a real AI provider"
	}
	
	// Check for operations
	if strings.Contains(lowerPrompt, "insert") {
		plan.Operation = "INSERT"
		plan.ReadOnly = false
		plan.Values = map[string]interface{}{"id": 1, "name": "example"}
	} else if strings.Contains(lowerPrompt, "delete") {
		plan.Operation = "DELETE"
		plan.ReadOnly = false
		plan.Warning = "This will delete data"
		plan.Where = []WhereClause{{Column: "id", Operator: "=", Value: 1}}
	} else if strings.Contains(lowerPrompt, "create") {
		plan.Operation = "CREATE"
		plan.ReadOnly = false
		plan.Schema = map[string]string{"id": "int PRIMARY KEY", "name": "text"}
	}
	
	if strings.Contains(lowerPrompt, "limit") {
		plan.Limit = 10
	}
	
	return plan, nil
}

// --- Gemini Client ---
type GeminiClient struct {
	BaseAIClient
}

func (c *GeminiClient) GenerateCQL(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	// TODO: Implement actual Gemini API call
	// For now, return mock response
	return (&MockClient{}).GenerateCQL(ctx, prompt, schema)
}

// --- OpenAI Client ---
type OpenAIClient struct {
	BaseAIClient
}

func (c *OpenAIClient) GenerateCQL(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	// TODO: Implement actual OpenAI API call
	// For now, return mock response
	return (&MockClient{}).GenerateCQL(ctx, prompt, schema)
}

// --- Anthropic Client ---
type AnthropicClient struct {
	BaseAIClient
}

func (c *AnthropicClient) GenerateCQL(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	// TODO: Implement actual Anthropic API call
	// For now, return mock response
	return (&MockClient{}).GenerateCQL(ctx, prompt, schema)
}

// NewAIClient is a factory function that returns the appropriate AI client
func NewAIClient(config *AIConfig) (AIClient, error) {
	if config == nil {
		return nil, fmt.Errorf("AI configuration is required")
	}
	
	var client AIClient
	
	switch config.Provider {
	case "gemini":
		client = &GeminiClient{BaseAIClient{APIKey: config.APIKey, Model: config.Model}}
	case "openai":
		client = &OpenAIClient{BaseAIClient{APIKey: config.APIKey, Model: config.Model}}
	case "anthropic":
		client = &AnthropicClient{BaseAIClient{APIKey: config.APIKey, Model: config.Model}}
	case "mock":
		client = &MockClient{}
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", config.Provider)
	}
	
	return client, nil
}

// GenerateCQLFromRequest is a high-level function that combines schema fetching and CQL generation
func GenerateCQLFromRequest(ctx context.Context, session *db.Session, userRequest string) (*QueryPlan, string, error) {
	// Get schema context
	schemaContext, err := session.GetSchemaContext(50) // Limit to 50 tables
	if err != nil {
		// Continue without schema if it fails
		schemaContext = "No schema information available"
	}
	
	// Get AI config from session
	dbConfig := session.GetAIConfig()
	config := ConvertDBConfigToAIConfig(dbConfig)
	
	// Create AI client
	client, err := NewAIClient(config)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create AI client: %v", err)
	}
	
	// Generate plan
	plan, err := client.GenerateCQL(ctx, userRequest, schemaContext)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate CQL plan: %v", err)
	}
	
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
