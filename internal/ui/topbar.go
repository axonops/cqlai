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
func (m TopBarModel) View(width int, styles *Styles) string {
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

	// Build the content with colors
	var content string
	if m.LastCommand != "" {
		content = labelStyle.Render("Last command: ") + commandStyle.Render(m.LastCommand)
		
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
