package ui

import (
	"context"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

// AICQLResultMsg is sent when AI CQL generation completes
type AICQLResultMsg struct {
	Plan           *ai.QueryPlan
	CQL            string
	Error          error
	ConversationID string // Track which conversation this result is for
}

// startAIConversation starts a new AI conversation
func startAIConversation(session *db.Session, userRequest string) tea.Cmd {
	return func() tea.Msg {
		logger.DebugfToFile("AI", "Starting new AI conversation for request: %s", userRequest)
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Initialize local AI for fuzzy search if needed
		if err := ai.InitializeLocalAI(session); err != nil {
			logger.DebugfToFile("AI", "Warning: could not initialize local AI: %v", err)
		}

		// Get minimal schema context
		schemaContext := ""
		if globalAI := ai.GetGlobalAI(); globalAI != nil && globalAI.GetCache() != nil {
			cache := globalAI.GetCache()
			cache.Mu.RLock()
			if len(cache.Keyspaces) > 0 {
				schemaContext = "Available keyspaces: "
				limit := 10
				if len(cache.Keyspaces) < limit {
					limit = len(cache.Keyspaces)
				}
				for i := 0; i < limit; i++ {
					if i > 0 {
						schemaContext += ", "
					}
					schemaContext += cache.Keyspaces[i]
				}
				if len(cache.Keyspaces) > 10 {
					schemaContext += "..."
				}
			}
			cache.Mu.RUnlock()
		} else {
			// Fallback to getting schema from session
			sc, err := session.GetSchemaContext(20)
			if err == nil {
				schemaContext = sc
			}
		}

		// Get AI config from session
		dbConfig := session.GetAIConfig()
		config := ai.ConvertDBConfigToAIConfig(dbConfig)
		
		// Start a new conversation
		cm := ai.GetConversationManager()
		conv, err := cm.StartConversation(config.Provider, config.Model, config.APIKey, userRequest, schemaContext)
		if err != nil {
			return AICQLResultMsg{
				Error: err,
			}
		}
		
		// Continue the conversation (first round)
		plan, interactionReq, err := conv.Continue(ctx, "")
		
		// Check if interaction is needed
		if interactionReq != nil {
			return AICQLResultMsg{
				Error:          interactionReq, // Pass as error to maintain compatibility
				ConversationID: conv.ID,
			}
		}
		
		if err != nil {
			return AICQLResultMsg{
				Error:          err,
				ConversationID: conv.ID,
			}
		}
		
		// Render CQL from plan
		cql := ""
		if plan != nil {
			renderedCQL, err := ai.RenderCQL(plan)
			if err == nil {
				cql = renderedCQL
			}
		}
		
		return AICQLResultMsg{
			Plan:           plan,
			CQL:            cql,
			ConversationID: conv.ID,
		}
	}
}

// continueAIConversation continues an existing AI conversation with user input
func continueAIConversation(conversationID string, userInput string) tea.Cmd {
	return func() tea.Msg {
		logger.DebugfToFile("AI", "Continuing conversation %s with input: %s", conversationID, userInput)
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		// Get the conversation
		cm := ai.GetConversationManager()
		conv, err := cm.GetConversation(conversationID)
		if err != nil {
			return AICQLResultMsg{
				Error:          err,
				ConversationID: conversationID,
			}
		}
		
		// Continue the conversation with user input
		plan, interactionReq, err := conv.Continue(ctx, userInput)
		
		// Check if interaction is needed
		if interactionReq != nil {
			return AICQLResultMsg{
				Error:          interactionReq, // Pass as error to maintain compatibility
				ConversationID: conversationID,
			}
		}
		
		if err != nil {
			return AICQLResultMsg{
				Error:          err,
				ConversationID: conversationID,
			}
		}
		
		// Render CQL from plan
		cql := ""
		if plan != nil {
			renderedCQL, err := ai.RenderCQL(plan)
			if err == nil {
				cql = renderedCQL
			}
		}
		
		return AICQLResultMsg{
			Plan:           plan,
			CQL:            cql,
			ConversationID: conversationID,
		}
	}
}

// generateAICQL is kept for backward compatibility but now starts a new conversation
func generateAICQL(session *db.Session, userRequest string) tea.Cmd {
	return startAIConversation(session, userRequest)
}