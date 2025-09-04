package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

// AICQLResultMsg is sent when AI CQL generation completes
type AICQLResultMsg struct {
	Plan           *ai.AIResult
	CQL            string
	Error          error
	ConversationID string // Track which conversation this result is for
}

// startAIConversation starts a new AI conversation
func startAIConversation(session *db.Session, aiConfig *config.AIConfig, userRequest string) tea.Cmd {
	return func() tea.Msg {
		logger.DebugfToFile("AI", "Starting new AI conversation for request: %s", userRequest)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Get schema context
		schemaContext := ""
		// Try to get from globalAI if available
		_ = ai.InitializeLocalAI(session) // Ignore error
		
		// Get minimal schema context
		schemaContext = "Available keyspaces: "
		sc, err := session.GetSchemaContext(20)
		if err == nil {
			schemaContext = sc
		}

		// Convert config to get proper provider-specific settings
		localConfig := ai.ConvertDBConfigToAIConfig(aiConfig)
		logger.DebugfToFile("AI", "Using provider: %s, model: %s, has_api_key: %v", 
			localConfig.Provider, localConfig.Model, localConfig.APIKey != "")

		// Start a new conversation
		cm := ai.GetConversationManager()
		conv, err := cm.StartConversation(
			string(localConfig.Provider),
			localConfig.Model,
			localConfig.APIKey,
			userRequest,
			schemaContext,
		)
		if err != nil {
			return AICQLResultMsg{Error: err}
		}
		
		logger.DebugfToFile("AI", "Started conversation with ID: %s", conv.ID)

		// Start the conversation with the user's request
		plan, interactionReq, err := conv.Continue(ctx, userRequest)

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
		if plan != nil && plan.Operation != "INFO" {
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
func continueAIConversation(aiConfig *config.AIConfig, conversationID string, userInput string) tea.Cmd {
	return func() tea.Msg {
		logger.DebugfToFile("AI", "Continuing conversation %s with input: %s", conversationID, userInput)

		// Validate that user input is not empty
		if strings.TrimSpace(userInput) == "" {
			logger.DebugfToFile("AI", "Empty user input detected, returning error")
			return AICQLResultMsg{
				Error:          fmt.Errorf("user input cannot be empty"),
				ConversationID: conversationID,
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
func generateAICQL(session *db.Session, aiConfig *config.AIConfig, userRequest string) tea.Cmd {
	return startAIConversation(session, aiConfig, userRequest)
}
