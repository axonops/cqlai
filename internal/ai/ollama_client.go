package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
)

// OllamaClient represents a client for the Ollama API.
type OllamaClient struct {
	config *config.AIProviderConfig
}

// NewOllamaClient creates a new Ollama client.
func NewOllamaClient(config *config.AIProviderConfig) *OllamaClient {
	return &OllamaClient{
		config: config,
	}
}

// ollamaMessage represents a message in Ollama format
type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content,omitempty"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

// ollamaToolCall represents a tool call in Ollama format
type ollamaToolCall struct {
	Function ollamaFunctionCall `json:"function"`
}

// ollamaFunctionCall represents function details in a tool call
type ollamaFunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ollamaTool represents a tool definition for Ollama
type ollamaTool struct {
	Type     string         `json:"type"`
	Function ollamaFunction `json:"function"`
}

// ollamaFunction represents a function definition
type ollamaFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ollamaRequest represents the request payload for the Ollama API.
type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Tools    []ollamaTool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
}

// ollamaResponse represents a single response from the Ollama API.
type ollamaResponse struct {
	Model     string        `json:"model"`
	CreatedAt time.Time     `json:"created_at"`
	Message   ollamaMessage `json:"message"`
	Done      bool          `json:"done"`
}

// getOllamaTools returns all tool definitions for Ollama's function calling
func getOllamaTools() []ollamaTool {
	// Get common tool definitions
	commonTools := GetCommonToolDefinitions()

	// Convert to Ollama format
	tools := make([]ollamaTool, len(commonTools))
	for i, tool := range commonTools {
		tools[i] = ollamaTool{
			Type: "function",
			Function: ollamaFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: map[string]any{
					"type":       "object",
					"properties": tool.Parameters,
					"required":   tool.Required,
				},
			},
		}
	}
	return tools
}

// ProcessRequestWithTools implements tool calling for Ollama (if supported by the model)
func (c *OllamaClient) ProcessRequestWithTools(ctx context.Context, prompt string, schema string) (*AIResult, error) {
	if c.config.URL == "" {
		return nil, fmt.Errorf("Ollama URL is not configured")
	}

	// Build the initial messages
	messages := []ollamaMessage{
		{
			Role:    "system",
			Content: SystemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Context: %s\n\nUser Request: %s", schema, prompt),
		},
	}

	// Get all tool definitions
	tools := getOllamaTools()

	apiURL := c.config.URL + "/api/chat"

	// Allow up to 5 rounds of tool calls
	for attempts := 0; attempts < 5; attempts++ {
		logger.DebugfToFile("Ollama", "Round %d: Sending request with tools", attempts+1)

		reqPayload := ollamaRequest{
			Model:    c.config.Model,
			Messages: messages,
			Tools:    tools,
			Stream:   false,
		}

		payloadBytes, err := json.Marshal(reqPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 300 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("Ollama API error %d: %s", resp.StatusCode, string(bodyBytes))
		}

		var ollamaResp ollamaResponse
		if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		// Check if the model wants to call functions
		if len(ollamaResp.Message.ToolCalls) > 0 {
			logger.DebugfToFile("Ollama", "Model requested %d tool calls", len(ollamaResp.Message.ToolCalls))

			// Add the assistant's message
			messages = append(messages, ollamaResp.Message)

			// Process each tool call
			for _, toolCall := range ollamaResp.Message.ToolCalls {
				logger.DebugfToFile("Ollama", "Processing tool call: %s", toolCall.Function.Name)

				// Execute the tool
				result := ExecuteToolCall(toolCall.Function.Name, toolCall.Function.Arguments)

				// Check if user interaction is needed
				if result.NeedsUserSelection || result.NeedsMoreInfo || result.NotRelevant {
					if result.NeedsUserSelection {
						return nil, &InteractionRequest{
							Type:             "selection",
							SelectionType:    result.SelectionType,
							SelectionOptions: result.SelectionOptions,
						}
					}
					if result.NeedsMoreInfo {
						return nil, &InteractionRequest{
							Type:        "info",
							InfoMessage: result.InfoMessage,
						}
					}
					if result.NotRelevant {
						return nil, &InteractionRequest{
							Type:        "not_relevant",
							InfoMessage: result.InfoMessage,
						}
					}
				}

				// Add the tool response
				responseContent := result.Data
				if result.Error != nil {
					responseContent = fmt.Sprintf("Error: %v", result.Error)
				}
				messages = append(messages, ollamaMessage{
					Role:    "tool",
					Content: responseContent,
				})
			}
			continue
		}

		// No tool calls, try to parse the response as a QueryPlan
		responseText := ollamaResp.Message.Content
		logger.DebugfToFile("Ollama", "Response: %s", responseText)

		// Extract JSON from the response
		jsonStr := extractJSON(responseText)
		if jsonStr == "" {
			jsonStr = responseText
		}

		var plan AIResult
		if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
			logger.DebugfToFile("Ollama", "Failed to parse JSON: %v", err)
			// Add a message asking for proper JSON format
			messages = append(messages, ollamaResp.Message)
			messages = append(messages, ollamaMessage{
				Role:    "user",
				Content: "Please respond with ONLY the QueryPlan JSON object, no other text.",
			})
			continue
		}

		logger.DebugfToFile("Ollama", "Successfully parsed QueryPlan")
		return &plan, nil
	}

	return nil, fmt.Errorf("failed to generate query plan after 5 attempts")
}

// SetAPIKey is a placeholder to satisfy the AIClient interface.
func (c *OllamaClient) SetAPIKey(key string) {
	// Ollama client doesn't typically use an API key in the same way as cloud providers.
}
