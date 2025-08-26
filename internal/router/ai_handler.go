package router

import (
	"fmt"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/db"
)

// AIHandler manages AI command processing
type AIHandler struct {
	ai *ai.AI
}

// aiHandlerInstance is the singleton instance
var aiHandlerInstance *AIHandler

// InitAIHandler initializes the AI handler
func InitAIHandler(session *db.Session) error {
	config := &ai.Config{
		Provider:    "local", // Start with local processing
		PrivacyMode: "schema_only",
		Temperature: 0.3,
		MaxTokens:   2000,
	}
	
	aiInstance, err := ai.NewAI(session, config)
	if err != nil {
		return fmt.Errorf("failed to initialize AI: %w", err)
	}
	
	aiHandlerInstance = &AIHandler{
		ai: aiInstance,
	}
	
	return nil
}

// GetAIHandler returns the singleton AI handler
func GetAIHandler() *AIHandler {
	return aiHandlerInstance
}

// HandleAICommand processes AI commands
func (h *AIHandler) HandleAICommand(command string) interface{} {
	if h == nil {
		return "AI system not initialized. Please restart the application."
	}
	
	// Remove .ai prefix
	query := strings.TrimSpace(strings.TrimPrefix(command, ".ai"))
	query = strings.TrimSpace(strings.TrimPrefix(query, ".AI"))
	
	if query == "" {
		return "Usage: .ai <natural language query>\n" +
			"Examples:\n" +
			"  .ai show me users table\n" +
			"  .ai find all tables with email\n" +
			"  .ai refresh schema\n" +
			"  .ai status"
	}
	
	// Handle special commands
	if strings.ToLower(query) == "status" {
		return h.handleStatus()
	}
	
	if strings.ToLower(query) == "refresh" || strings.ToLower(query) == "refresh schema" {
		if err := h.ai.RefreshSchema(); err != nil {
			return fmt.Sprintf("Error refreshing schema: %v", err)
		}
		return "Schema cache refreshed successfully"
	}
	
	// For natural language queries, just return a message
	// The actual AI processing happens through the existing client.go
	return "AI query processing will use the configured LLM provider."
}

// handleStatus returns AI system status
func (h *AIHandler) handleStatus() string {
	stats := h.ai.GetSchemaStats()
	
	keyspaces := stats["keyspaces"].([]string)
	tableCount := stats["table_count"].(int)
	lastRefresh := stats["last_refresh"].(time.Time)
	
	return fmt.Sprintf("AI System Status:\n"+
		"  Schema Cache:\n"+
		"    Keyspaces: %d\n"+
		"    Tables: %d\n"+
		"    Last refresh: %s\n"+
		"  Fuzzy Search: Enabled\n"+
		"  LLM Provider: Configured in session",
		len(keyspaces), tableCount, lastRefresh.Format("2006-01-02 15:04:05"))
}

