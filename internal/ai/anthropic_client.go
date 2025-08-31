package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
	"github.com/axonops/cqlai/internal/logger"
)

// AnthropicClient implements the AIClient interface for Anthropic's Claude
type AnthropicClient struct {
	BaseAIClient
	client *anthropic.Client
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	if model == "" {
		model = DefaultAnthropicModel
	}

	// Create HTTP client with timeout for better reliability
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(httpClient),
	)

	return &AnthropicClient{
		BaseAIClient: BaseAIClient{
			APIKey: apiKey,
			Model:  model,
		},
		client: &client,
	}
}

// retryWithBackoff retries a function with exponential backoff on rate limit or server errors
func retryWithBackoff[T any](ctx context.Context, maxRetries int, fn func(context.Context) (T, error)) (T, error) {
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
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			// Add 10% jitter to avoid thundering herd
			jitter := time.Duration(float64(backoff) * 0.1)
			sleepTime := backoff + jitter

			logger.DebugfToFile("Anthropic", "Retrying after %v (attempt %d/%d): %v", sleepTime, attempt+1, maxRetries, err)

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

// getAnthropicTools returns all tool definitions for Anthropic's tool use API
func getAnthropicTools() []anthropic.ToolUnionParam {
	// Get common tool definitions
	commonTools := GetCommonToolDefinitions()

	// Convert to Anthropic format
	tools := make([]anthropic.ToolParam, len(commonTools))
	for i, tool := range commonTools {
		// Convert properties to the correct Anthropic type
		props := make(map[string]interface{})
		if tool.Parameters != nil {
			// The Parameters field already contains the properly structured JSON schema
			// We just need to ensure it's in the right format for Anthropic
			for k, v := range tool.Parameters {
				props[k] = v
			}
		}

		tools[i] = anthropic.ToolParam{
			Name:        tool.Name,
			Description: param.NewOpt(tool.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Type:       constant.Object("object"),
				Properties: props,
				Required:   tool.Required,
			},
		}
	}

	// Convert ToolParam to ToolUnionParam
	unionTools := make([]anthropic.ToolUnionParam, len(tools))
	for i, tool := range tools {
		toolCopy := tool
		unionTools[i] = anthropic.ToolUnionParam{
			OfTool: &toolCopy,
		}
	}
	return unionTools
}

// ProcessRequestWithTools implements tool calling using Anthropic's native tool use API
func (c *AnthropicClient) ProcessRequestWithTools(ctx context.Context, prompt string, schema string) (*AIResult, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("API key is required for %s", "Anthropic")
	}

	// Build the initial messages
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(fmt.Sprintf("Context: %s\n\nUser Request: %s", schema, prompt))),
	}

	// Get all tool definitions
	tools := getAnthropicTools()

	// System prompt for CQL generation - use the proper one with tool descriptions
	systemPrompt := SystemPrompt

	// Allow up to 5 rounds of tool calls
	for attempts := 0; attempts < 5; attempts++ {
		logger.DebugfToFile("Anthropic", "Round %d: Sending request with tools", attempts+1)

		// Use retry logic for API calls
		response, err := retryWithBackoff(ctx, 3, func(ctx context.Context) (*anthropic.Message, error) {
			return c.client.Messages.New(ctx, anthropic.MessageNewParams{
				Model:       anthropic.Model(c.Model),
				MaxTokens:   1024,
				Temperature: param.NewOpt(0.2), // Low temperature for more deterministic output
				System: []anthropic.TextBlockParam{
					{Text: systemPrompt},
				},
				Messages: messages,
				Tools:    tools,
			})
		})
		if err != nil {
			return nil, fmt.Errorf("Anthropic API error: %v", err)
		}

		// Add the assistant's response to messages
		assistantMessage := anthropic.NewAssistantMessage()
		var toolResults []struct {
			ID     string
			Result *CommandResult
		}
		hasToolUse := false

		// First pass: check if there are any tool uses
		for _, content := range response.Content {
			if content.Type == "tool_use" {
				hasToolUse = true
				break
			}
		}

		// Process each content block
		for _, content := range response.Content {
			switch content.Type {
			case "text":
				// Assistant responded with text
				assistantMessage.Content = append(assistantMessage.Content, anthropic.NewTextBlock(content.Text))

				// Only try to parse as QueryPlan if no tools were used at all
				// (Pure text response that should be a JSON query plan)
				if !hasToolUse {
					jsonStr := extractJSON(content.Text)
					if jsonStr == "" {
						jsonStr = content.Text
					}

					var plan AIResult
					if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
						logger.DebugfToFile("Anthropic", "Text response not a valid QueryPlan, will ask for clarification")
						// Don't return yet, continue processing
					} else {
						logger.DebugfToFile("Anthropic", "Successfully parsed QueryPlan from text response")
						return &plan, nil
					}
				}

			case "tool_use":
				// Assistant wants to use a tool
				logger.DebugfToFile("Anthropic", "Tool use requested: %s", content.Name)

				// Parse the input for executing the tool
				var inputMap map[string]any
				if err := json.Unmarshal(content.Input, &inputMap); err != nil {
					logger.DebugfToFile("Anthropic", "Failed to parse tool input: %v", err)
					inputMap = make(map[string]any)
				}

				// Preserve the exact JSON bytes from the API to avoid key re-ordering issues
				// The API expects the input to be preserved exactly as sent
				assistantMessage.Content = append(assistantMessage.Content, anthropic.NewToolUseBlock(
					content.ID,
					content.Input, // Pass the original JSON bytes, not the parsed object
					content.Name,
				))

				// Execute the tool
				result := ExecuteToolCall(content.Name, inputMap)

				// Log tool execution result
				logger.DebugfToFile("Anthropic", "Tool %s execution: Success=%v, Data=%s, Error=%v",
					content.Name, result.Success, result.Data, result.Error)

				// Check if this is a submit_query_plan tool and it succeeded
				if content.Name == ToolSubmitQueryPlan.String() && result.Success && result.QueryPlan != nil {
					logger.DebugfToFile("Anthropic", "Query plan submitted via tool, returning immediately")
					return result.QueryPlan, nil
				}

				// Check if this is an info tool and it succeeded
				if content.Name == ToolInfo.String() && result.Success && result.InfoResponse != nil {
					logger.DebugfToFile("Anthropic", "Info response submitted via tool, returning as informational response")
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

				// Store tool result for later
				toolResults = append(toolResults, struct {
					ID     string
					Result *CommandResult
				}{ID: content.ID, Result: result})
			}
		}

		// If we have tool uses, add the assistant message and tool results
		if hasToolUse && len(assistantMessage.Content) > 0 {
			messages = append(messages, assistantMessage)

			// Add all tool results as user messages
			for _, tr := range toolResults {
				responseContent := tr.Result.Data
				isError := false
				if tr.Result.Error != nil {
					responseContent = fmt.Sprintf("Error: %v", tr.Result.Error)
					isError = true
				}
				messages = append(messages, anthropic.NewUserMessage(anthropic.NewToolResultBlock(
					tr.ID,
					responseContent,
					isError,
				)))
			}
		} else if len(assistantMessage.Content) > 0 {
			// No tools were used, but we have text content
			messages = append(messages, assistantMessage)
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock("Please respond with ONLY the QueryPlan JSON object, no other text.")))
		}
	}

	return nil, fmt.Errorf("failed to generate query plan after 5 attempts")
}

// continueAnthropic continues an Anthropic conversation
func (conv *AIConversation) continueAnthropic(ctx context.Context, userInput string) (*AIResult, *InteractionRequest, error) {
	// Build messages array from conversation history
	var messages []anthropic.MessageParam

	// On first call, start with the original request
	if len(conv.Messages) == 0 {
		userPrompt := UserPrompt(conv.OriginalRequest, conv.SchemaContext)
		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)))
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userPrompt})
	} else {
		// Build message history
		for _, msg := range conv.Messages {
			switch msg.Role {
			case "user":
				messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
			case "assistant":
				messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
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
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(followUpMessage)))
			conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userInput}) // Store original for history
		}
	}

	// Get tool definitions for continued conversation
	tools := getAnthropicTools()

	// Make API call with retry logic
	logger.DebugfToFile("AIConversation", "[%s] Calling Anthropic API with %d messages and %d tools", conv.ID, len(messages), len(tools))

	response, err := retryWithBackoff(ctx, 3, func(ctx context.Context) (*anthropic.Message, error) {
		return conv.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
			Model:       anthropic.Model(conv.Model),
			MaxTokens:   1024,
			Temperature: param.NewOpt(0.2), // Low temperature for more deterministic output
			System: []anthropic.TextBlockParam{
				{Text: SystemPrompt},
			},
			Messages: messages,
			Tools:    tools,
		})
	})
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic API error: %v", err)
	}

	// Check if the response contains tool uses
	hasToolUse := false
	for _, content := range response.Content {
		if content.Type == "tool_use" {
			hasToolUse = true
			break
		}
	}

	// If tool use is present, handle it
	if hasToolUse {
		logger.DebugfToFile("AIConversation", "[%s] Response contains tool uses", conv.ID)

		// Process tool uses
		assistantMessage := anthropic.NewAssistantMessage()
		var toolResults []struct {
			ID     string
			Result *CommandResult
		}

		for _, content := range response.Content {
			switch content.Type {
			case "text":
				assistantMessage.Content = append(assistantMessage.Content, anthropic.NewTextBlock(content.Text))
			case "tool_use":
				// Parse the tool input
				var inputMap map[string]any
				if err := json.Unmarshal(content.Input, &inputMap); err != nil {
					logger.DebugfToFile("AIConversation", "[%s] Failed to parse tool input: %v", conv.ID, err)
					continue
				}

				// Add the tool use to the assistant message, preserving exact JSON bytes
				assistantMessage.Content = append(assistantMessage.Content, anthropic.NewToolUseBlock(
					content.ID,
					content.Input, // Already preserving the original JSON bytes
					content.Name,
				))

				// Execute the tool
				result := ExecuteToolCall(content.Name, inputMap)

				// Check for special cases
				if content.Name == ToolSubmitQueryPlan.String() && result.Success && result.QueryPlan != nil {
					logger.DebugfToFile("AIConversation", "[%s] Query plan submitted via tool", conv.ID)
					return result.QueryPlan, nil, nil
				}

				if content.Name == ToolInfo.String() && result.Success && result.InfoResponse != nil {
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

				// Store tool result for later
				toolResults = append(toolResults, struct {
					ID     string
					Result *CommandResult
				}{ID: content.ID, Result: result})
			}
		}

		// Add assistant message to conversation history
		// Extract text from the response for conversation history
		var assistantText string
		for _, content := range response.Content {
			if content.Type == "text" {
				assistantText += content.Text
			}
		}
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "assistant", Content: assistantText})

		// If we have tool results, we need to continue the conversation
		if len(toolResults) > 0 {
			// Add tool results as user messages
			var resultMessages []string
			for _, tr := range toolResults {
				if tr.Result.Error != nil {
					resultMessages = append(resultMessages, fmt.Sprintf("Tool %s error: %v", tr.ID, tr.Result.Error))
				} else {
					resultMessages = append(resultMessages, tr.Result.Data)
				}
			}

			combinedResult := strings.Join(resultMessages, "\n")
			conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: combinedResult})

			// Continue the conversation
			return conv.Continue(ctx, "")
		}
	}

	// Extract response text for non-tool responses
	var responseText string
	for _, content := range response.Content {
		if text := content.Text; text != "" {
			responseText += text
		}
	}

	logger.DebugfToFile("AIConversation", "[%s] Response: %s", conv.ID, responseText)

	// Add assistant response to conversation history if not already added
	if !hasToolUse {
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "assistant", Content: responseText})
	}

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
				ConversationID:   conv.ID, // Include conversation ID for resumption
			}, nil
		}

		if result.NeedsMoreInfo {
			logger.DebugfToFile("AIConversation", "[%s] More info needed: %s", conv.ID, result.InfoMessage)
			return nil, &InteractionRequest{
				Type:           "info",
				InfoMessage:    result.InfoMessage,
				ConversationID: conv.ID, // Include conversation ID for resumption
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
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: clarification})

		// Try again
		return conv.Continue(ctx, "")
	}

	logger.DebugfToFile("AIConversation", "[%s] Successfully parsed QueryPlan", conv.ID)
	return &plan, nil, nil
}
