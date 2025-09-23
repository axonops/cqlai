package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TopBarModel is the Bubble Tea model for the top status bar.
type TopBarModel struct {
	LastCommand  string
	QueryTime    time.Duration
	RowCount     int
	HasQueryData bool
	AutoFetch    bool
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

	// AutoFetch styles
	autoFetchOnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87FFD7")).
		Bold(true)

	autoFetchOffStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5F5F5F"))

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
		modeText = "CQL"
	}
	content := labelStyle.Render("Mode: ") + modeStyle.Render(modeText)

	// Add AutoFetch status
	autoFetchState := "OFF"
	autoFetchStyle := autoFetchOffStyle
	if m.AutoFetch {
		autoFetchState = "ON"
		autoFetchStyle = autoFetchOnStyle
	}
	content += separatorStyle.Render(" │ ") +
		labelStyle.Render("AutoFetch: ") + autoFetchStyle.Render(autoFetchState)

	// Add command information if available
	if m.LastCommand != "" {
		// Truncate long commands to fit in the top bar
		displayCommand := m.LastCommand
		maxCommandLength := 50
		if len(displayCommand) > maxCommandLength {
			displayCommand = displayCommand[:maxCommandLength] + "..."
		}
		content += separatorStyle.Render(" │ ") +
			labelStyle.Render("Last: ") + commandStyle.Render(displayCommand)

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
