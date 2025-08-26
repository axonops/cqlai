package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/axonops/cqlai/internal/logger"
)

// AnthropicClient implements the AIClient interface for Anthropic's Claude
type AnthropicClient struct {
	BaseAIClient
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	if model == "" {
		model = "claude-3-sonnet-20240229"
	}
	return &AnthropicClient{
		BaseAIClient: BaseAIClient{
			APIKey: apiKey,
			Model:  model,
		},
	}
}

func (c *AnthropicClient) GenerateCQLWithTools(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	// Create Anthropic client
	client := anthropic.NewClient(option.WithAPIKey(c.APIKey))

	// Build prompt that instructs Claude to use special commands
	systemPrompt := SystemPrompt
	userPrompt := UserPrompt(prompt, schema)

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
	}

	// Allow up to 3 rounds of interaction
	for attempts := 0; attempts < 3; attempts++ {
		logger.DebugfToFile("Anthropic", "Round %d: Sending request", attempts+1)

		message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.Model(c.Model),
			MaxTokens: 1024,
			System: []anthropic.TextBlockParam{
				{Text: systemPrompt},
			},
			Messages: messages,
		})
		if err != nil {
			return nil, fmt.Errorf("Anthropic API error: %v", err)
		}

		// Extract response text
		var responseText string
		for _, content := range message.Content {
			if text := content.Text; text != "" {
				responseText += text
			}
		}

		logger.DebugfToFile("Anthropic", "Response: %s", responseText)

		// Check if the response contains a command
		if cmd, arg, found := ParseCommand(responseText); found {
			logger.DebugfToFile("Anthropic", "Executing command: %s with arg: %s", cmd, arg)
			
			result := ExecuteCommand(cmd, arg)
			
			// Check if user interaction is needed
			if result.NeedsUserSelection {
				logger.DebugfToFile("Anthropic", "User selection needed, returning to UI")
				return nil, &InteractionRequest{
					Type:             "selection",
					SelectionType:    result.SelectionType,
					SelectionOptions: result.SelectionOptions,
				}
			}
			
			if result.NeedsMoreInfo {
				logger.DebugfToFile("Anthropic", "More info needed, returning to UI")
				return nil, &InteractionRequest{
					Type:        "info",
					InfoMessage: result.InfoMessage,
				}
			}
			
			// Check for error
			if result.Error != nil {
				logger.DebugfToFile("Anthropic", "Command execution failed: %v", result.Error)
				// Add error to conversation and continue
				messages = append(messages,
					anthropic.NewAssistantMessage(anthropic.NewTextBlock(responseText)),
					anthropic.NewUserMessage(anthropic.NewTextBlock(fmt.Sprintf("Error: %v\nNow generate the QueryPlan JSON for the original request.", result.Error))),
				)
				continue
			}
			
			// Success - add result to conversation and continue
			messages = append(messages,
				anthropic.NewAssistantMessage(anthropic.NewTextBlock(responseText)),
				anthropic.NewUserMessage(anthropic.NewTextBlock(result.Data+"\nNow generate the QueryPlan JSON for the original request.")),
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
			logger.DebugfToFile("Anthropic", "Failed to parse JSON: %v", err)
			// Try one more time with clearer instruction
			messages = append(messages,
				anthropic.NewAssistantMessage(anthropic.NewTextBlock(responseText)),
				anthropic.NewUserMessage(anthropic.NewTextBlock("Please respond with ONLY the QueryPlan JSON object, no other text.")),
			)
			continue
		}

		return &plan, nil
	}

	return nil, fmt.Errorf("failed to generate CQL after 3 attempts")
}

func (c *AnthropicClient) GenerateCQL(ctx context.Context, prompt string, schema string) (*QueryPlan, error) {
	// Since we now use a unified approach, just call GenerateCQLWithTools
	return c.GenerateCQLWithTools(ctx, prompt, schema)
}
