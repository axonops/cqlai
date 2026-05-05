package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	anthropicoption "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/openai/openai-go"
	openaioption "github.com/openai/openai-go/option"
)

const (
	openAiBaseURL        = "https://api.openai.com/v1"
	openRouterBaseURL    = "https://openrouter.ai/api/v1"
	ollamaBaseURL        = "http://localhost:11434/v1"
	maxConversations     = 100              // Maximum number of conversations to keep
	conversationMaxAge   = 24 * time.Hour   // Maximum age before cleanup
)

// ConversationManager manages ongoing AI conversations
type ConversationManager struct {
	mu            sync.RWMutex
	conversations map[string]*AIConversation
}

var conversationManager = &ConversationManager{
	conversations: make(map[string]*AIConversation),
}

// GetConversationManager returns the singleton conversation manager
func GetConversationManager() *ConversationManager {
	return conversationManager
}

// StartConversation starts a new AI conversation
func (cm *ConversationManager) StartConversation(provider, model, apiKey, baseURL, request, schemaContext string) (*AIConversation, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Cleanup old conversations if we have too many
	if len(cm.conversations) >= maxConversations {
		cm.cleanupLocked()
	}

	conv := &AIConversation{
		ID:              fmt.Sprintf("conv-%d", time.Now().UnixNano()),
		Provider:        provider,
		Model:           model,
		APIKey:          apiKey,
		BaseURL:         baseURL,
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
		client := anthropic.NewClient(anthropicoption.WithAPIKey(apiKey))
		conv.anthropicClient = &client
	case "openai":
		url := baseURL
		if url == "" {
			url = openAiBaseURL
		}
		client := openai.NewClient(
			openaioption.WithAPIKey(apiKey),
			openaioption.WithBaseURL(url),
		)
		conv.openaiClient = &client
	case "openrouter":
		url := baseURL
		if url == "" {
			url = openRouterBaseURL
		}
		client := openai.NewClient(
			openaioption.WithAPIKey(apiKey),
			openaioption.WithBaseURL(url),
		)
		conv.openrouterClient = &client
	case "ollama":
		url := baseURL
		if url == "" {
			url = ollamaBaseURL
		}
		client := openai.NewClient(
			openaioption.WithAPIKey(apiKey),
			openaioption.WithBaseURL(url),
		)
		conv.ollamaClient = &client
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

// GetConversationMessages returns the messages from a conversation
func (cm *ConversationManager) GetConversationMessages(conversationID string) []ConversationMessage {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if conv, exists := cm.conversations[conversationID]; exists {
		// Return a copy to avoid external modifications
		messages := make([]ConversationMessage, len(conv.Messages))
		copy(messages, conv.Messages)
		return messages
	}
	return nil
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

// cleanupLocked removes old conversations (must be called with lock held)
func (cm *ConversationManager) cleanupLocked() {
	cutoff := time.Now().Add(-conversationMaxAge)
	cleaned := 0

	// First pass: remove conversations older than maxAge
	for id, conv := range cm.conversations {
		if conv.LastActivity.Before(cutoff) {
			delete(cm.conversations, id)
			cleaned++
			logger.DebugfToFile("ConversationManager", "Cleaned up old conversation %s", id)
		}
	}

	// Second pass: if still over limit, remove oldest conversations
	for len(cm.conversations) >= maxConversations {
		var oldestID string
		var oldestTime time.Time
		first := true

		for id, conv := range cm.conversations {
			if first || conv.LastActivity.Before(oldestTime) {
				oldestID = id
				oldestTime = conv.LastActivity
				first = false
			}
		}

		if oldestID != "" {
			delete(cm.conversations, oldestID)
			cleaned++
			logger.DebugfToFile("ConversationManager", "Evicted oldest conversation %s to stay under limit", oldestID)
		} else {
			break
		}
	}

	if cleaned > 0 {
		logger.DebugfToFile("ConversationManager", "Cleanup complete: removed %d conversations, %d remaining", cleaned, len(cm.conversations))
	}
}

// Continue continues the conversation with user input (or empty string for continuation)
func (conv *AIConversation) Continue(ctx context.Context, userInput string) (*AIResult, *InteractionRequest, error) {
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
	case "openrouter":
		return conv.continueOpenRouter(ctx, userInput)
	case "ollama":
		return conv.continueOllama(ctx, userInput)
	default:
		return nil, nil, fmt.Errorf("unsupported provider: %s", conv.Provider)
	}
}
