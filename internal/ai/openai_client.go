package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
	"github.com/openai/openai-go/shared/constant"
)

// OpenAIClient implements the AIClient interface for OpenAI using the official SDK
type OpenAIClient struct {
	BaseAIClient
	client *openai.Client
}

// NewOpenAIClient creates a new OpenAI client using the official SDK
func NewOpenAIClient(apiKey string, model string) *OpenAIClient {
	if model == "" {
		model = string(openai.ChatModelGPT4oMini)
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIClient{
		BaseAIClient: BaseAIClient{
			APIKey: apiKey,
			Model:  model,
		},
		client: &client,
	}
}

// getOpenAITools returns all tool definitions for OpenAI's function calling API
func getOpenAITools() []openai.ChatCompletionToolParam {
	// Get common tool definitions
	commonTools := GetCommonToolDefinitions()

	// Convert to OpenAI format
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

// GenerateCQLWithTools implements tool calling using OpenAI's native function calling API
func (c *OpenAIClient) GenerateCQLWithTools(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("API key is required for %s", "OpenAI")
	}

	// Build the initial messages
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(SystemPrompt),
		openai.UserMessage(fmt.Sprintf("Context: %s\n\nUser Request: %s", schema, prompt)),
	}

	// Get all tool definitions
	tools := getOpenAITools()

	// Make the API call with tools
	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    openai.ChatModel(c.Model),
		Tools:    tools,
	}

	// Allow up to 5 rounds of tool calls
	for attempts := 0; attempts < 5; attempts++ {
		logger.DebugfToFile("OpenAI", "Round %d: Sending request with tools", attempts+1)

		completion, err := c.client.Chat.Completions.New(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("OpenAI API error: %v", err)
		}

		if len(completion.Choices) == 0 {
			return nil, fmt.Errorf("no response from OpenAI")
		}

		choice := completion.Choices[0]

		// Check if the model wants to call functions
		if len(choice.Message.ToolCalls) > 0 {
			logger.DebugfToFile("OpenAI", "Model requested %d tool calls", len(choice.Message.ToolCalls))

			// Add the assistant's message with tool calls
			assistantMsg := openai.AssistantMessage("")
			// We need to manually add the tool calls to the message
			// This is a simplified version - the actual SDK may have different requirements
			messages = append(messages, assistantMsg)

			// Process each tool call
			for _, toolCall := range choice.Message.ToolCalls {
				logger.DebugfToFile("OpenAI", "Processing tool call: %s", toolCall.Function.Name)

				// Parse the arguments
				var args map[string]any
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					logger.DebugfToFile("OpenAI", "Failed to parse tool arguments: %v", err)
					messages = append(messages, openai.ToolMessage(toolCall.ID, fmt.Sprintf("Error parsing arguments: %v", err)))
					continue
				}

				// Execute the tool and get the result
				result := ExecuteToolCall(toolCall.Function.Name, args)

				// Check if this is a submit_query_plan tool and it succeeded
				if toolCall.Function.Name == ToolSubmitQueryPlan.String() && result.Success && result.QueryPlan != nil {
					logger.DebugfToFile("OpenAI", "Query plan submitted via tool, returning immediately")
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

				// Add the tool response
				responseContent := result.Data
				if result.Error != nil {
					responseContent = fmt.Sprintf("Error: %v", result.Error)
				}
				messages = append(messages, openai.ToolMessage(toolCall.ID, responseContent))
			}

			// Update params with new messages
			params.Messages = messages
			continue
		}

		// No tool calls, try to parse the response as a QueryPlan
		responseText := choice.Message.Content
		logger.DebugfToFile("OpenAI", "Response: %s", responseText)

		// Extract JSON from the response
		jsonStr := extractJSON(responseText)
		if jsonStr == "" {
			jsonStr = responseText
		}

		var plan QueryPlan
		if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
			logger.DebugfToFile("OpenAI", "Failed to parse JSON: %v", err)
			// Add a message asking for proper JSON format
			messages = append(messages, openai.AssistantMessage(responseText))
			messages = append(messages, openai.UserMessage("Please respond with ONLY the QueryPlan JSON object, no other text."))
			params.Messages = messages
			continue
		}

		logger.DebugfToFile("OpenAI", "Successfully parsed QueryPlan")
		return &plan, nil
	}

	return nil, fmt.Errorf("failed to generate query plan after 5 attempts")
}

// continueOpenAI continues an OpenAI conversation
func (conv *AIConversation) continueOpenAI(ctx context.Context, userInput string) (*QueryPlan, *InteractionRequest, error) {
	// Build messages array from conversation history using the official SDK types
	var messages []openai.ChatCompletionMessageParamUnion

	// Always start with system prompt
	messages = append(messages, openai.SystemMessage(SystemPrompt))

	// On first call, add the original request
	if len(conv.Messages) == 0 {
		userPrompt := UserPrompt(conv.OriginalRequest, conv.SchemaContext)
		messages = append(messages, openai.UserMessage(userPrompt))
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userPrompt})
	} else {
		// Build message history
		for _, msg := range conv.Messages {
			switch msg.Role {
			case "user":
				messages = append(messages, openai.UserMessage(msg.Content))
			case "assistant":
				messages = append(messages, openai.AssistantMessage(msg.Content))
			}
		}

		// Add new user input if provided
		if userInput != "" {
			messages = append(messages, openai.UserMessage(userInput))
			conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userInput})
		}
	}

	// Make API call using the official SDK client
	logger.DebugfToFile("AIConversation", "[%s] Calling OpenAI API with %d messages", conv.ID, len(messages))

	// Create a new OpenAI client using the official SDK
	client := openai.NewClient(option.WithAPIKey(conv.APIKey))

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:  messages,
		Model:     openai.ChatModel(conv.Model),
		MaxTokens: param.NewOpt(int64(1024)),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("openAI API error: %v", err)
	}

	if len(completion.Choices) == 0 {
		return nil, nil, fmt.Errorf("no response from OpenAI")
	}

	responseText := completion.Choices[0].Message.Content
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
