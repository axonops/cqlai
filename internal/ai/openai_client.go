package ai

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/axonops/cqlai/internal/logger"
)

// OpenAIClient implements the AIClient interface for OpenAI
type OpenAIClient struct {
	BaseAIClient
	client *openai.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey string, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIClient{
		BaseAIClient: BaseAIClient{
			APIKey: apiKey,
			Model:  model,
		},
		client: openai.NewClient(apiKey),
	}
}

// GenerateCQLWithTools implements tool calling for OpenAI
func (c *OpenAIClient) GenerateCQLWithTools(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	// Build prompt that instructs OpenAI to use special commands
	systemPrompt := SystemPrompt
	userPrompt := UserPrompt(prompt, schema)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	// Allow up to 3 rounds of interaction
	for attempts := 0; attempts < 3; attempts++ {
		logger.DebugfToFile("OpenAI", "Round %d: Sending request", attempts+1)

		resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:     c.Model,
			Messages:  messages,
			MaxTokens: 1024,
		})
		if err != nil {
			return nil, fmt.Errorf("OpenAI API error: %v", err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("no response from OpenAI")
		}

		// Extract response text
		responseText := resp.Choices[0].Message.Content
		logger.DebugfToFile("OpenAI", "Response: %s", responseText)

		// Check if the response contains a command
		if cmd, arg, found := ParseCommand(responseText); found {
			logger.DebugfToFile("OpenAI", "Executing command: %s with arg: %s", cmd, arg)
			
			result := ExecuteCommand(cmd, arg)
			
			// Check if user interaction is needed
			if result.NeedsUserSelection {
				logger.DebugfToFile("OpenAI", "User selection needed, returning to UI")
				return nil, &InteractionRequest{
					Type:             "selection",
					SelectionType:    result.SelectionType,
					SelectionOptions: result.SelectionOptions,
				}
			}
			
			if result.NeedsMoreInfo {
				logger.DebugfToFile("OpenAI", "More info needed, returning to UI")
				return nil, &InteractionRequest{
					Type:        "info",
					InfoMessage: result.InfoMessage,
				}
			}
			
			// Check for error
			if result.Error != nil {
				logger.DebugfToFile("OpenAI", "Command execution failed: %v", result.Error)
				// Add error to conversation and continue
				messages = append(messages,
					openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: responseText,
					},
					openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("Error: %v\nNow generate the QueryPlan JSON for the original request.", result.Error),
					},
				)
				continue
			}
			
			// Success - add result to conversation and continue
			messages = append(messages,
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: responseText,
				},
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: result.Data + "\nNow generate the QueryPlan JSON for the original request.",
				},
			)
			continue
		}

		// Not a command, try to parse as JSON
		jsonStr := extractJSON(responseText)
		if jsonStr == "" {
			jsonStr = responseText
		}

		var plan QueryPlan
		if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
			logger.DebugfToFile("OpenAI", "Failed to parse JSON: %v", err)
			// Try one more time with clearer instruction
			messages = append(messages,
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: responseText,
				},
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: "Please respond with ONLY the QueryPlan JSON object, no other text.",
				},
			)
			continue
		}

		return &plan, nil
	}

	return nil, fmt.Errorf("failed to generate CQL after 3 attempts")
}

// GenerateCQL implements basic CQL generation without tools
func (c *OpenAIClient) GenerateCQL(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	// Since we now use a unified approach, just call GenerateCQLWithTools
	return c.GenerateCQLWithTools(ctx, prompt, schema)
}

