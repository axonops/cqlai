package ui

import (
	"context"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/db"
	tea "github.com/charmbracelet/bubbletea"
)

// AICQLResultMsg is sent when AI CQL generation completes
type AICQLResultMsg struct {
	Plan  *ai.QueryPlan
	CQL   string
	Error error
}

// generateAICQL generates CQL from a natural language request
func generateAICQL(session *db.Session, userRequest string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		plan, cql, err := ai.GenerateCQLFromRequest(ctx, session, userRequest)
		return AICQLResultMsg{
			Plan:  plan,
			CQL:   cql,
			Error: err,
		}
	}
}