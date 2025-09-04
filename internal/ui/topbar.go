package ui

import (
	"fmt"
	"time"
	
	"github.com/charmbracelet/lipgloss"
)

// TopBarModel is the Bubble Tea model for the top status bar.
type TopBarModel struct {
	LastCommand string
	QueryTime   time.Duration
	RowCount    int
	HasQueryData bool
}

// NewTopBarModel creates a new TopBarModel.
func NewTopBarModel() TopBarModel {
	return TopBarModel{}
}

// View renders the top status bar.
func (m TopBarModel) View(width int, styles *Styles, viewMode string) string {
	// Define component styles
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))
	
	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D7FF")).
		Bold(true)
	
	queryTimeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87FF00"))
	
	rowCountStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700"))
	
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))
	
	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF87FF")).
		Bold(true)

	// Start with the mode
	var modeText string
	switch viewMode {
	case "ai":
		modeText = "AI"
	case "table":
		modeText = "TABLE"
	case "trace":
		modeText = "TRACE"
	default:
		modeText = "HISTORY"
	}
	content := labelStyle.Render("Mode: ") + modeStyle.Render(modeText)

	// Add command information if available
	if m.LastCommand != "" {
		content += separatorStyle.Render(" │ ") +
			labelStyle.Render("Last: ") + commandStyle.Render(m.LastCommand)
		
		if m.HasQueryData {
			content += separatorStyle.Render(" │ ") +
				labelStyle.Render("Query: ") + queryTimeStyle.Render(fmt.Sprintf("%v", m.QueryTime.Round(time.Millisecond))) +
				separatorStyle.Render(" │ ") +
				labelStyle.Render("Rows: ") + rowCountStyle.Render(fmt.Sprintf("%d", m.RowCount))
		}
	}

	// Apply style to the entire bar without forced background
	barStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(width)

	return barStyle.Render(content)
}
