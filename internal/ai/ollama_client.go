package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/axonops/cqlai/internal/db"
)

// OllamaClient represents a client for the Ollama API.
type OllamaClient struct {
	config *db.AIProviderConfig
}

// NewOllamaClient creates a new Ollama client.
func NewOllamaClient(config *db.AIProviderConfig) *OllamaClient {
	return &OllamaClient{
		config: config,
	}
}

// ollamaRequest represents the request payload for the Ollama API.
type ollamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// ollamaResponse represents a single response from the Ollama API.
type ollamaResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Message   Message   `json:"message"`
	Done      bool      `json:"done"`
}

// Generate generates text using the Ollama API.
func (c *OllamaClient) Generate(ctx context.Context, messages []Message) (string, error) {
	if c.config.URL == "" {
		return "", fmt.Errorf("Ollama URL is not configured")
	}

	// Add a check to prevent sending empty messages
	if len(messages) == 0 {
		return "", fmt.Errorf("cannot generate from an empty set of messages")
	}

	apiURL := c.config.URL + "/api/chat"

	reqPayload := ollamaRequest{
		Model:    c.config.Model,
		Messages: messages,
		Stream:   false, // For simplicity, we'll use non-streaming responses
	}

	payloadBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned non-200 status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	return ollamaResp.Message.Content, nil
}

// GenerateCQL generates a CQL query from a natural language prompt.
func (c *OllamaClient) GenerateCQL(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	fullPrompt := fmt.Sprintf("Given the following Cassandra schema:\n%s\n\nTranslate this request into a valid CQL query: \"%s\". Respond with only the CQL query.", schema, prompt)

	messages := []Message{
		{Role: "user", Content: fullPrompt},
	}

	cqlQuery, err := c.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CQL from Ollama: %w", err)
	}

	// This is a simplified parser. A real implementation would need a proper
	// CQL parser to deconstruct the query into a full QueryPlan.
	// For now, we'll assume it's a SELECT and put the whole query in the Operation.
	// This is not ideal but will satisfy the interface for now.
	plan := &QueryPlan{
		Operation: "SELECT",  // Placeholder
		Table:     "unknown", // Placeholder
		// A more advanced implementation would parse the CQL to fill other fields.
	}

	// A hack to return the raw CQL: piggyback on the Warning field.
	// The UI can then decide to display this.
	plan.Warning = fmt.Sprintf("Raw CQL from Ollama: %s", cqlQuery)

	return plan, nil
}

// GenerateCQLWithTools is a placeholder to satisfy the AIClient interface.
func (c *OllamaClient) GenerateCQLWithTools(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	// For now, this will just call the non-tool version.
	return c.GenerateCQL(ctx, prompt, schema)
}

// SetAPIKey is a placeholder to satisfy the AIClient interface.
func (c *OllamaClient) SetAPIKey(key string) {
	// Ollama client doesn't typically use an API key in the same way as cloud providers.
}
