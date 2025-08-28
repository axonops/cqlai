package ai

import (
	"context"
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
	Messages     []ConversationMessage
	CurrentRound int
	MaxRounds    int

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
