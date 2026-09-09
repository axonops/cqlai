package ui

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
)

// TopBarModel is the Bubble Tea model for the top status bar.
type TopBarModel struct {
	LastCommand  string
	QueryTime    time.Duration
	RowCount     int
	HasQueryData bool
	// Not displayed here any more - the status bar shows it with the other
	// session settings - but still needed to decide whether the row count is
	// shown as "100+" with more to fetch.
	AutoFetch   bool
	HasMoreData bool // Indicates if there's more data to fetch
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

	// The mode used to be named here. The tab line above shows it now, and
	// highlights it, so repeating it would only take up room.
	content := ""

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
				labelStyle.Render("Rows: ")

			// Show row count with "+" if more data is available
			if m.HasMoreData && !m.AutoFetch {
				content += rowCountStyle.Render(fmt.Sprintf("%d+", m.RowCount))
			} else {
				content += rowCountStyle.Render(fmt.Sprintf("%d", m.RowCount))
			}
		}
	}

	// Apply style to the entire bar without forced background
	barStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(width)

	return barStyle.Render(content)
}
