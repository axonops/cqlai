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
		return nil, fmt.Errorf("ollama URL is not configured")
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

		// Add Authorization header if API key is provided
		if c.config.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		}

		client := &http.Client{Timeout: 300 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("ollama API error %d: %s", resp.StatusCode, string(bodyBytes))
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

// continueOllama continues an Ollama conversation
func (conv *AIConversation) continueOllama(ctx context.Context, userInput string) (*AIResult, *InteractionRequest, error) {
	// Build messages array from conversation history
	var messages []ollamaMessage

	// On first call, add system prompt and the original request
	if len(conv.Messages) == 0 {
		// Add system prompt only on first call
		messages = append(messages, ollamaMessage{
			Role:    "system",
			Content: SystemPrompt,
		})
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "system", Content: SystemPrompt})

		// Add the original request
		userPrompt := UserPrompt(conv.OriginalRequest, conv.SchemaContext)
		messages = append(messages, ollamaMessage{
			Role:    "user",
			Content: userPrompt,
		})
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userPrompt})
	} else {
		// Build message history (which already includes system prompt from first call)
		for _, msg := range conv.Messages {
			messages = append(messages, ollamaMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		// Add new user input if provided
		if userInput != "" {
			// For follow-up questions, make it clear this is a follow-up
			followUpMessage := userInput
			if len(conv.Messages) > 0 {
				// This is a follow-up in an existing conversation
				followUpMessage = "Follow-up question (answer only this, don't repeat previous response): " + userInput
			}
			messages = append(messages, ollamaMessage{
				Role:    "user",
				Content: followUpMessage,
			})
			conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userInput}) // Store original for history
		}
	}

	// Get tool definitions for continued conversation
	tools := getOllamaTools()

	// Make API call using Ollama's chat endpoint
	logger.DebugfToFile("AIConversation", "[%s] Calling Ollama API with %d messages and %d tools", conv.ID, len(messages), len(tools))

	// Use configured BaseURL or fall back to default
	baseURL := conv.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}
	apiURL := baseURL + "/chat/completions"

	reqPayload := ollamaRequest{
		Model:    conv.Model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}

	payloadBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.DebugfToFile("AIConversation", "[%s] Payload: %s", conv.ID, string(payloadBytes))

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add Authorization header if API key is provided
	if conv.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+conv.APIKey)
	}

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("ollama API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if the model wants to call functions
	if len(ollamaResp.Message.ToolCalls) > 0 {
		logger.DebugfToFile("AIConversation", "[%s] Model requested %d tool calls", conv.ID, len(ollamaResp.Message.ToolCalls))

		// Process the first tool call (subsequent calls handled via recursive Continue)
		toolCall := ollamaResp.Message.ToolCalls[0]
		logger.DebugfToFile("AIConversation", "[%s] Processing tool call: %s", conv.ID, toolCall.Function.Name)

		// Execute the tool
		result := ExecuteToolCall(toolCall.Function.Name, toolCall.Function.Arguments)

		// Check if this is a submit_query_plan tool and it succeeded
		if toolCall.Function.Name == ToolSubmitQueryPlan.String() && result.Success && result.QueryPlan != nil {
			logger.DebugfToFile("AIConversation", "[%s] Query plan submitted via tool", conv.ID)
			return result.QueryPlan, nil, nil
		}

		// Check if this is an info tool
		if toolCall.Function.Name == ToolInfo.String() && result.Success && result.InfoResponse != nil {
			logger.DebugfToFile("AIConversation", "[%s] Info response submitted via tool", conv.ID)
			// Return a QueryPlan that represents an informational response
			return &AIResult{
				Operation:   "INFO",
				Confidence:  result.InfoResponse.Confidence,
				ReadOnly:    true,
				InfoContent: result.InfoResponse.Content,
				InfoTitle:   result.InfoResponse.Title,
			}, nil, nil
		}

		// Check if user interaction is needed
		if result.NeedsUserSelection {
			return nil, &InteractionRequest{
				Type:             "selection",
				SelectionType:    result.SelectionType,
				SelectionOptions: result.SelectionOptions,
				ConversationID:   conv.ID,
			}, nil
		}

		if result.NeedsMoreInfo {
			return nil, &InteractionRequest{
				Type:           "info",
				InfoMessage:    result.InfoMessage,
				ConversationID: conv.ID,
			}, nil
		}

		if result.NotRelevant {
			return nil, &InteractionRequest{
				Type:           "not_relevant",
				InfoMessage:    result.InfoMessage,
				ConversationID: conv.ID,
			}, nil
		}

		// If we have a result, we need to continue the conversation with the tool result
		responseContent := result.Data
		if result.Error != nil {
			responseContent = fmt.Sprintf("Error: %v", result.Error)
		}

		// Add the tool result and continue
		conv.Messages = append(conv.Messages, ConversationMessage{
			Role:    "assistant",
			Content: fmt.Sprintf("Tool %s result: %s", toolCall.Function.Name, responseContent),
		})

		// Continue the conversation (which will handle any subsequent tool calls)
		return conv.Continue(ctx, "")
	}

	// Handle text response
	responseText := ollamaResp.Message.Content
	logger.DebugfToFile("AIConversation", "[%s] Response: %s", conv.ID, responseText)

	// Add assistant response to conversation history
	conv.Messages = append(conv.Messages, ConversationMessage{Role: "assistant", Content: responseText})

	// Check if the response contains a command
	if cmd, arg, found := ParseCommand(responseText); found {
		logger.DebugfToFile("AIConversation", "[%s] Executing command: %s with arg: %s", conv.ID, cmd, arg)

		result := ExecuteCommand(cmd, arg)

		// Check if user interaction is needed
		if result.NeedsUserSelection {
			logger.DebugfToFile("AIConversation", "[%s] User selection needed for %s", conv.ID, result.SelectionType)
			return nil, &InteractionRequest{
				Type:             "selection",
				SelectionType:    result.SelectionType,
				SelectionOptions: result.SelectionOptions,
				ConversationID:   conv.ID,
			}, nil
		}

		if result.NeedsMoreInfo {
			logger.DebugfToFile("AIConversation", "[%s] More info needed: %s", conv.ID, result.InfoMessage)
			return nil, &InteractionRequest{
				Type:           "info",
				InfoMessage:    result.InfoMessage,
				ConversationID: conv.ID,
			}, nil
		}

		// Handle command result
		var resultMessage string
		if result.Error != nil {
			resultMessage = fmt.Sprintf("Error: %v\nNow generate the QueryPlan JSON for the original request.", result.Error)
		} else {
			resultMessage = result.Data + "\nNow generate the QueryPlan JSON for the original request."
		}

		// Add result to conversation and continue
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: resultMessage})

		// Recursively continue
		return conv.Continue(ctx, "")
	}

	// Try to parse as QueryPlan JSON
	jsonStr := extractJSON(responseText)
	if jsonStr == "" {
		jsonStr = responseText
	}

	var plan AIResult
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		logger.DebugfToFile("AIConversation", "[%s] Failed to parse JSON: %v", conv.ID, err)

		// Ask for clarification
		clarification := "Please respond with ONLY the QueryPlan JSON object, no other text."
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: clarification, SystemGenerated: true})

		// Try again
		return conv.Continue(ctx, "")
	}

	logger.DebugfToFile("AIConversation", "[%s] Successfully parsed QueryPlan", conv.ID)
	return &plan, nil, nil
}
