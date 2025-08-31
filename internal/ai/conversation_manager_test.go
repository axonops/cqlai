package ai

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestConversationManager(t *testing.T) {
	cm := GetConversationManager()
	
	// Test starting a conversation with anthropic provider
	conv, err := cm.StartConversation("anthropic", "claude-3-sonnet-20240229", "test-key", "test request", "test schema")
	if err != nil {
		t.Fatalf("Failed to start conversation: %v", err)
	}
	
	if conv.ID == "" {
		t.Error("Conversation ID should not be empty")
	}
	
	if conv.Provider != "anthropic" {
		t.Errorf("Expected provider 'anthropic', got %s", conv.Provider)
	}
	
	if conv.OriginalRequest != "test request" {
		t.Errorf("Expected original request 'test request', got %s", conv.OriginalRequest)
	}
	
	// Test retrieving conversation
	retrieved, err := cm.GetConversation(conv.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve conversation: %v", err)
	}
	
	if retrieved.ID != conv.ID {
		t.Error("Retrieved conversation ID doesn't match")
	}
	
	// Test conversation not found
	_, err = cm.GetConversation("nonexistent-id")
	if err == nil {
		t.Error("Expected error for nonexistent conversation")
	}
}

func TestConversationCleanup(t *testing.T) {
	cm := GetConversationManager()
	
	// Create a conversation
	conv, err := cm.StartConversation("anthropic", "claude-3-sonnet-20240229", "test-key", "cleanup test", "test schema")
	if err != nil {
		t.Fatalf("Failed to start conversation: %v", err)
	}
	
	// Manually set LastActivity to old time
	conv.LastActivity = time.Now().Add(-2 * time.Hour)
	
	// Clean up old conversations (1 hour)
	cm.CleanupOldConversations(1 * time.Hour)
	
	// Should not be able to retrieve the conversation
	_, err = cm.GetConversation(conv.ID)
	if err == nil {
		t.Error("Expected error retrieving cleaned up conversation")
	}
}

func TestConversationMessageHistory(t *testing.T) {
	// Create a conversation
	conv := &AIConversation{
		ID:              "test-conv",
		Provider:        "anthropic",
		Model:           "claude-3",
		APIKey:          "test-key",
		OriginalRequest: "test query",
		SchemaContext:   "test schema",
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
		Messages:        []ConversationMessage{},
		CurrentRound:    0,
		MaxRounds:       10,
	}
	
	// Test adding messages
	conv.Messages = append(conv.Messages, ConversationMessage{
		Role:    "user",
		Content: "Hello",
	})
	
	conv.Messages = append(conv.Messages, ConversationMessage{
		Role:    "assistant",
		Content: "Hi there",
	})
	
	if len(conv.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(conv.Messages))
	}
	
	if conv.Messages[0].Role != "user" {
		t.Error("First message should be from user")
	}
	
	if conv.Messages[1].Role != "assistant" {
		t.Error("Second message should be from assistant")
	}
}

func TestParseCommandInConversation(t *testing.T) {
	// Test parsing commands that might appear in conversation
	testCases := []struct {
		response    string
		expectCmd   bool
		cmdType     ToolName
		cmdArg      string
	}{
		{
			response:  `Let me search for that. {"tool": "fuzzy_search", "params": {"query": "users"}} Searching...`,
			expectCmd: true,
			cmdType:   ToolFuzzySearch,
			cmdArg:    "users",
		},
		{
			response:  `{"tool": "user_selection", "params": {"type": "table", "options": ["users", "accounts", "sessions"]}}`,
			expectCmd: true,
			cmdType:   ToolUserSelection,
			cmdArg:    "table:users,accounts,sessions",
		},
		{
			response:  `{"tool": "not_enough_info", "params": {"message": "Please specify which keyspace"}}`,
			expectCmd: true,
			cmdType:   ToolNotEnoughInfo,
			cmdArg:    "Please specify which keyspace",
		},
		{
			response:  "Here is your query plan JSON",
			expectCmd: false,
		},
	}
	
	for _, tc := range testCases {
		cmd, arg, found := ParseCommand(tc.response)
		
		if found != tc.expectCmd {
			t.Errorf("For response '%s': expected found=%v, got %v", 
				tc.response, tc.expectCmd, found)
		}
		
		if tc.expectCmd {
			if cmd != tc.cmdType {
				t.Errorf("For response '%s': expected cmd=%v, got %v",
					tc.response, tc.cmdType, cmd)
			}
			if !strings.Contains(arg, strings.TrimSpace(tc.cmdArg)) {
				t.Errorf("For response '%s': expected arg to contain '%s', got '%s'",
					tc.response, tc.cmdArg, arg)
			}
		}
	}
}

func TestConversationRoundLimit(t *testing.T) {
	conv := &AIConversation{
		ID:           "test-conv",
		CurrentRound: 10,
		MaxRounds:    10,
		Messages:     []ConversationMessage{},
	}
	
	ctx := context.Background()
	
	// Should fail when exceeding max rounds
	_, _, err := conv.Continue(ctx, "test")
	if err == nil {
		t.Error("Expected error when exceeding max rounds")
	}
	
	if !strings.Contains(err.Error(), "exceeded maximum rounds") {
		t.Errorf("Expected max rounds error, got: %v", err)
	}
}