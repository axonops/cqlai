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

		plan, cql, err := ai.GenerateCQLFromRequest(ctx, session, userRequest)
		if err != nil {
			// Check if this is an interaction request that needs to be bubbled up
			if interactionReq, ok := err.(*ai.InteractionRequest); ok {
				return AICQLResultMsg{
					Error: interactionReq, // Pass as error to maintain compatibility
				}
			}
			return AICQLResultMsg{Error: err}
		}

		return AICQLResultMsg{
			Plan: plan,
			CQL:  cql,
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
