package ai

import (
	"context"
	"encoding/json"
	"fmt"

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
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicClient{
		BaseAIClient: BaseAIClient{
			APIKey: apiKey,
			Model:  model,
		},
		client: &client,
	}
}

// getAnthropicTools returns all tool definitions for Anthropic's tool use API
func getAnthropicTools() []anthropic.ToolUnionParam {
	// Get common tool definitions
	commonTools := GetCommonToolDefinitions()

	// Convert to Anthropic format
	tools := make([]anthropic.ToolParam, len(commonTools))
	for i, tool := range commonTools {
		tools[i] = anthropic.ToolParam{
			Name:        tool.Name,
			Description: param.NewOpt(tool.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Type:       constant.Object("object"),
				Properties: tool.Parameters,
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

// GenerateCQLWithTools implements tool calling using Anthropic's native tool use API
func (c *AnthropicClient) GenerateCQLWithTools(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
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

		response, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.Model(c.Model),
			MaxTokens: 1024,
			System: []anthropic.TextBlockParam{
				{Text: systemPrompt},
			},
			Messages: messages,
			Tools:    tools,
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

					var plan QueryPlan
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

				// Parse the input for executing the tool and preserving for assistant message
				var inputMap map[string]any
				if err := json.Unmarshal(content.Input, &inputMap); err != nil {
					logger.DebugfToFile("Anthropic", "Failed to parse tool input: %v", err)
					inputMap = make(map[string]any)
				}

				// We need to preserve the original input exactly as it was sent by the API
				// The API expects the input to be preserved as a JSON object, not a string
				assistantMessage.Content = append(assistantMessage.Content, anthropic.NewToolUseBlock(
					content.ID,
					inputMap, // Pass the parsed JSON object, not the string
					content.Name,
				))

				// Execute the tool
				result := ExecuteToolCall(content.Name, inputMap)

				// Check if this is a submit_query_plan tool and it succeeded
				if content.Name == ToolSubmitQueryPlan.String() && result.Success && result.QueryPlan != nil {
					logger.DebugfToFile("Anthropic", "Query plan submitted via tool, returning immediately")
					return result.QueryPlan, nil
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
				if tr.Result.Error != nil {
					responseContent = fmt.Sprintf("Error: %v", tr.Result.Error)
				}
				messages = append(messages, anthropic.NewUserMessage(anthropic.NewToolResultBlock(
					tr.ID,
					responseContent,
					false, // isError
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
func (conv *AIConversation) continueAnthropic(ctx context.Context, userInput string) (*QueryPlan, *InteractionRequest, error) {
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
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(userInput)))
			conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userInput})
		}
	}

	// Make API call
	logger.DebugfToFile("AIConversation", "[%s] Calling Anthropic API with %d messages", conv.ID, len(messages))

	response, err := conv.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(conv.Model),
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: SystemPrompt},
		},
		Messages: messages,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic API error: %v", err)
	}

	// Extract response text
	var responseText string
	for _, content := range response.Content {
		if text := content.Text; text != "" {
			responseText += text
		}
	}

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

	var plan QueryPlan
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
