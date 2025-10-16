package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
	"github.com/openai/openai-go/shared/constant"
)

const (
	openRouterBaseURL = "https://openrouter.ai/api/v1"
)

// OpenRouterClient implements the AIClient interface for OpenRouter using the OpenAI SDK
type OpenRouterClient struct {
	BaseAIClient
	client *openai.Client
}

// NewOpenRouterClient creates a new OpenRouter client using the OpenAI SDK with custom base URL
func NewOpenRouterClient(apiKey string, model string) *OpenRouterClient {
	if model == "" {
		model = DefaultOpenRouterModel
	}

	// Create HTTP client with timeout for better reliability
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	// OpenRouter uses OpenAI-compatible API with custom base URL
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(openRouterBaseURL),
		option.WithHTTPClient(httpClient),
	)

	return &OpenRouterClient{
		BaseAIClient: BaseAIClient{
			APIKey: apiKey,
			Model:  model,
		},
		client: &client,
	}
}

// retryWithBackoffOpenRouter retries a function with exponential backoff on rate limit or server errors
func retryWithBackoffOpenRouter[T any](ctx context.Context, maxRetries int, fn func(context.Context) (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable (rate limit or server error)
		errStr := err.Error()
		isRateLimit := strings.Contains(errStr, "429") || strings.Contains(errStr, "rate_limit")
		isServerError := strings.Contains(errStr, "500") || strings.Contains(errStr, "502") ||
			strings.Contains(errStr, "503") || strings.Contains(errStr, "504")

		if !isRateLimit && !isServerError {
			// Not a retryable error
			return result, err
		}

		if attempt < maxRetries-1 {
			// Calculate backoff with jitter
			// Cap exponent at 10 to prevent overflow
			exp := attempt
			if exp > 10 {
				exp = 10
			}
			backoff := time.Duration(1<<uint(exp)) * time.Second // #nosec G115 - exp is capped at 10
			// Add 10% jitter to avoid thundering herd
			jitter := time.Duration(float64(backoff) * 0.1)
			sleepTime := backoff + jitter

			logger.DebugfToFile("OpenRouter", "Retrying after %v (attempt %d/%d): %v", sleepTime, attempt+1, maxRetries, err)

			select {
			case <-time.After(sleepTime):
				// Continue to next retry
			case <-ctx.Done():
				return result, ctx.Err()
			}
		}
	}

	return result, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// getOpenRouterTools returns all tool definitions for OpenRouter's function calling API
// OpenRouter uses the same format as OpenAI
func getOpenRouterTools() []openai.ChatCompletionToolParam {
	// Get common tool definitions
	commonTools := GetCommonToolDefinitions()

	// Convert to OpenAI/OpenRouter format
	tools := make([]openai.ChatCompletionToolParam, len(commonTools))
	for i, tool := range commonTools {
		tools[i] = openai.ChatCompletionToolParam{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: param.NewOpt(tool.Description),
				Parameters: shared.FunctionParameters{
					"type":       "object",
					"properties": tool.Parameters,
					"required":   tool.Required,
				},
			},
		}
	}
	return tools
}

// ProcessRequestWithTools implements tool calling using OpenRouter's OpenAI-compatible function calling API
func (c *OpenRouterClient) ProcessRequestWithTools(ctx context.Context, prompt string, schema string) (*AIResult, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("API key is required for %s", "OpenRouter")
	}

	// Build the initial messages
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(SystemPrompt),
		openai.UserMessage(fmt.Sprintf("Context: %s\n\nUser Request: %s", schema, prompt)),
	}

	// Get all tool definitions
	tools := getOpenRouterTools()

	// Make the API call with tools
	params := openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       openai.ChatModel(c.Model),
		Tools:       tools,
		Temperature: param.NewOpt(0.2), // Low temperature for more deterministic output
	}

	// Allow up to 5 rounds of tool calls
	for attempts := 0; attempts < 5; attempts++ {
		logger.DebugfToFile("OpenRouter", "Round %d: Sending request with tools", attempts+1)

		// Use retry logic for API calls
		completion, err := retryWithBackoffOpenRouter(ctx, 3, func(ctx context.Context) (*openai.ChatCompletion, error) {
			return c.client.Chat.Completions.New(ctx, params)
		})
		if err != nil {
			return nil, fmt.Errorf("OpenRouter API error: %v", err)
		}

		if len(completion.Choices) == 0 {
			return nil, fmt.Errorf("no response from OpenRouter")
		}

		choice := completion.Choices[0]

		// Check if the model wants to call functions
		if len(choice.Message.ToolCalls) > 0 {
			logger.DebugfToFile("OpenRouter", "Model requested %d tool calls", len(choice.Message.ToolCalls))

			// Convert tool calls to param format
			toolCallParams := make([]openai.ChatCompletionMessageToolCallParam, len(choice.Message.ToolCalls))
			for i, tc := range choice.Message.ToolCalls {
				toolCallParams[i] = openai.ChatCompletionMessageToolCallParam{
					ID:   tc.ID,
					Type: tc.Type,
					Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}

			// Add the assistant's message with tool calls to the conversation
			assistantMsgParam := openai.ChatCompletionAssistantMessageParam{
				ToolCalls: toolCallParams,
			}
			// If there's content, set it using the helper to handle the union type properly
			if choice.Message.Content != "" {
				// Create a basic assistant message with content, then merge tool calls
				baseMsg := openai.AssistantMessage(choice.Message.Content)
				if baseMsg.OfAssistant != nil {
					baseMsg.OfAssistant.ToolCalls = toolCallParams
				}
				messages = append(messages, baseMsg)
			} else {
				// No content, just tool calls - create the union manually
				messages = append(messages, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &assistantMsgParam,
				})
			}

			// Process each tool call
			for _, toolCall := range choice.Message.ToolCalls {
				logger.DebugfToFile("OpenRouter", "Processing tool call: %s with ID: %s", toolCall.Function.Name, toolCall.ID)

				// Validate tool call ID length
				if len(toolCall.ID) > 40 {
					logger.DebugfToFile("OpenRouter", "Warning: tool call ID too long (%d chars): %s", len(toolCall.ID), toolCall.ID)
				}

				// Parse the arguments
				var args map[string]any
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					logger.DebugfToFile("OpenRouter", "Failed to parse tool arguments: %v", err)
					messages = append(messages, openai.ToolMessage(fmt.Sprintf("Error parsing arguments: %v", err), toolCall.ID))
					continue
				}

				// Execute the tool and get the result
				result := ExecuteToolCall(toolCall.Function.Name, args)

				// Log tool execution result
				logger.DebugfToFile("OpenRouter", "Tool %s execution: Success=%v, Data=%s, Error=%v",
					toolCall.Function.Name, result.Success, result.Data, result.Error)

				// Check if this is a submit_query_plan tool and it succeeded
				if toolCall.Function.Name == ToolSubmitQueryPlan.String() && result.Success && result.QueryPlan != nil {
					logger.DebugfToFile("OpenRouter", "Query plan submitted via tool, returning immediately")
					return result.QueryPlan, nil
				}

				// Check if this is an info tool and it succeeded
				if toolCall.Function.Name == ToolInfo.String() && result.Success && result.InfoResponse != nil {
					logger.DebugfToFile("OpenRouter", "Info response submitted via tool, returning as informational response")
					// Return a QueryPlan that represents an informational response
					return &AIResult{
						Operation:   "INFO",
						Confidence:  result.InfoResponse.Confidence,
						ReadOnly:    true,
						InfoContent: result.InfoResponse.Content,
						InfoTitle:   result.InfoResponse.Title,
					}, nil
				}

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
				messages = append(messages, openai.ToolMessage(responseContent, toolCall.ID))
			}

			// Update params with new messages
			params.Messages = messages
			continue
		}

		// No tool calls, try to parse the response as a QueryPlan
		responseText := choice.Message.Content
		logger.DebugfToFile("OpenRouter", "Response: %s", responseText)

		// Extract JSON from the response
		jsonStr := extractJSON(responseText)
		if jsonStr == "" {
			jsonStr = responseText
		}

		var plan AIResult
		if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
			logger.DebugfToFile("OpenRouter", "Failed to parse JSON: %v", err)
			// Add a message asking for proper JSON format
			messages = append(messages, openai.AssistantMessage(responseText))
			messages = append(messages, openai.UserMessage("Please respond with ONLY the QueryPlan JSON object, no other text."))
			params.Messages = messages
			continue
		}

		logger.DebugfToFile("OpenRouter", "Successfully parsed QueryPlan")
		return &plan, nil
	}

	return nil, fmt.Errorf("failed to generate query plan after 5 attempts")
}

// continueOpenRouter continues an OpenRouter conversation
func (conv *AIConversation) continueOpenRouter(ctx context.Context, userInput string) (*AIResult, *InteractionRequest, error) {
	// Build messages array from conversation history using the official SDK types
	var messages []openai.ChatCompletionMessageParamUnion

	// On first call, add system prompt and the original request
	if len(conv.Messages) == 0 {
		// Add system prompt only on first call
		messages = append(messages, openai.SystemMessage(SystemPrompt))
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "system", Content: SystemPrompt})

		// Add the original request
		userPrompt := UserPrompt(conv.OriginalRequest, conv.SchemaContext)
		messages = append(messages, openai.UserMessage(userPrompt))
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userPrompt})
	} else {
		// Build message history (which already includes system prompt from first call)
		for _, msg := range conv.Messages {
			switch msg.Role {
			case "system":
				messages = append(messages, openai.SystemMessage(msg.Content))
			case "user":
				messages = append(messages, openai.UserMessage(msg.Content))
			case "assistant":
				messages = append(messages, openai.AssistantMessage(msg.Content))
			}
		}

		// Add new user input if provided
		if userInput != "" {
			// For follow-up questions, make it clear this is a follow-up
			followUpMessage := userInput
			if len(conv.Messages) > 0 {
				// This is a follow-up in an existing conversation
				followUpMessage = "Follow-up question (answer only this, don't repeat previous response): " + userInput
			}
			messages = append(messages, openai.UserMessage(followUpMessage))
			conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userInput}) // Store original for history
		}
	}

	// Get tool definitions for continued conversation
	tools := getOpenRouterTools()

	// Make API call using the stored client
	logger.DebugfToFile("AIConversation", "[%s] Calling OpenRouter API with %d messages and %d tools", conv.ID, len(messages), len(tools))

	// Use the existing client from the conversation
	if conv.openrouterClient == nil {
		return nil, nil, fmt.Errorf("OpenRouter client not initialized")
	}

	// Use retry logic for API calls
	completion, err := retryWithBackoffOpenRouter(ctx, 3, func(ctx context.Context) (*openai.ChatCompletion, error) {
		return conv.openrouterClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages:    messages,
			Model:       openai.ChatModel(conv.Model),
			Tools:       tools,
			MaxTokens:   param.NewOpt(int64(1024)),
			Temperature: param.NewOpt(0.2), // Low temperature for more deterministic output
		})
	})
	if err != nil {
		return nil, nil, fmt.Errorf("OpenRouter API error: %v", err)
	}

	if len(completion.Choices) == 0 {
		return nil, nil, fmt.Errorf("no response from OpenRouter")
	}

	choice := completion.Choices[0]

	// Check if the model wants to call functions
	if len(choice.Message.ToolCalls) > 0 {
		logger.DebugfToFile("AIConversation", "[%s] Model requested %d tool calls", conv.ID, len(choice.Message.ToolCalls))

		// Process tool calls similarly to ProcessRequestWithTools
		for _, toolCall := range choice.Message.ToolCalls {
			logger.DebugfToFile("AIConversation", "[%s] Processing tool call: %s", conv.ID, toolCall.Function.Name)

			// Parse the arguments
			var args map[string]any
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				logger.DebugfToFile("AIConversation", "[%s] Failed to parse tool arguments: %v", conv.ID, err)
				continue
			}

			// Execute the tool
			result := ExecuteToolCall(toolCall.Function.Name, args)

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

			// Continue the conversation
			return conv.Continue(ctx, "")
		}
	}

	// Handle text response
	responseText := choice.Message.Content
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
