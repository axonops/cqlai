package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/axonops/cqlai/internal/logger"
	openai "github.com/sashabaranov/go-openai"
)

// ConversationManager manages ongoing AI conversations
type ConversationManager struct {
	mu            sync.RWMutex
	conversations map[string]*AIConversation
}

// AIConversation represents a single AI conversation session
type AIConversation struct {
	ID              string
	Provider        string
	Model           string
	APIKey          string
	OriginalRequest string
	SchemaContext   string
	CreatedAt       time.Time
	LastActivity    time.Time
	
	// Conversation state
	Messages        []ConversationMessage
	CurrentRound    int
	MaxRounds       int
	
	// Provider-specific clients
	anthropicClient *anthropic.Client
	openaiClient    *openai.Client
}

// ConversationMessage represents a message in the conversation
type ConversationMessage struct {
	Role    string // "user", "assistant", "system"
	Content string
}

var conversationManager = &ConversationManager{
	conversations: make(map[string]*AIConversation),
}

// GetConversationManager returns the singleton conversation manager
func GetConversationManager() *ConversationManager {
	return conversationManager
}

// StartConversation starts a new AI conversation
func (cm *ConversationManager) StartConversation(provider, model, apiKey, request, schemaContext string) (*AIConversation, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	conv := &AIConversation{
		ID:              fmt.Sprintf("conv-%d", time.Now().UnixNano()),
		Provider:        provider,
		Model:           model,
		APIKey:          apiKey,
		OriginalRequest: request,
		SchemaContext:   schemaContext,
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
		Messages:        []ConversationMessage{},
		CurrentRound:    0,
		MaxRounds:       10,
	}
	
	// Initialize provider-specific client
	switch provider {
	case "anthropic":
		client := anthropic.NewClient(option.WithAPIKey(apiKey))
		conv.anthropicClient = &client
	case "openai":
		conv.openaiClient = openai.NewClient(apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
	
	// Store the conversation
	cm.conversations[conv.ID] = conv
	
	logger.DebugfToFile("ConversationManager", "Started new conversation %s with provider %s", conv.ID, provider)
	
	return conv, nil
}

// GetConversation retrieves an existing conversation
func (cm *ConversationManager) GetConversation(id string) (*AIConversation, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	conv, exists := cm.conversations[id]
	if !exists {
		return nil, fmt.Errorf("conversation %s not found", id)
	}
	
	return conv, nil
}

// CleanupOldConversations removes conversations older than the specified duration
func (cm *ConversationManager) CleanupOldConversations(maxAge time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	for id, conv := range cm.conversations {
		if conv.LastActivity.Before(cutoff) {
			delete(cm.conversations, id)
			logger.DebugfToFile("ConversationManager", "Cleaned up old conversation %s", id)
		}
	}
}

// Continue continues the conversation with user input (or empty string for continuation)
func (conv *AIConversation) Continue(ctx context.Context, userInput string) (*QueryPlan, *InteractionRequest, error) {
	conv.LastActivity = time.Now()
	conv.CurrentRound++
	
	if conv.CurrentRound > conv.MaxRounds {
		return nil, nil, fmt.Errorf("conversation exceeded maximum rounds (%d)", conv.MaxRounds)
	}
	
	logger.DebugfToFile("AIConversation", "[%s] Round %d: Continuing with input: %s", conv.ID, conv.CurrentRound, userInput)
	
	switch conv.Provider {
	case "anthropic":
		return conv.continueAnthropic(ctx, userInput)
	case "openai":
		return conv.continueOpenAI(ctx, userInput)
	default:
		return nil, nil, fmt.Errorf("unsupported provider: %s", conv.Provider)
	}
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
		return nil, nil, fmt.Errorf("Anthropic API error: %v", err)
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

// continueOpenAI continues an OpenAI conversation
func (conv *AIConversation) continueOpenAI(ctx context.Context, userInput string) (*QueryPlan, *InteractionRequest, error) {
	// Build messages array from conversation history
	var messages []openai.ChatCompletionMessage
	
	// Always start with system prompt
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: SystemPrompt,
	})
	
	// On first call, add the original request
	if len(conv.Messages) == 0 {
		userPrompt := UserPrompt(conv.OriginalRequest, conv.SchemaContext)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		})
		conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userPrompt})
	} else {
		// Build message history
		for _, msg := range conv.Messages {
			switch msg.Role {
			case "user":
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: msg.Content,
				})
			case "assistant":
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: msg.Content,
				})
			}
		}
		
		// Add new user input if provided
		if userInput != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: userInput,
			})
			conv.Messages = append(conv.Messages, ConversationMessage{Role: "user", Content: userInput})
		}
	}
	
	// Make API call
	logger.DebugfToFile("AIConversation", "[%s] Calling OpenAI API with %d messages", conv.ID, len(messages))
	
	response, err := conv.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     conv.Model,
		Messages:  messages,
		MaxTokens: 1024,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("OpenAI API error: %v", err)
	}
	
	if len(response.Choices) == 0 {
		return nil, nil, fmt.Errorf("no response from OpenAI")
	}
	
	responseText := response.Choices[0].Message.Content
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