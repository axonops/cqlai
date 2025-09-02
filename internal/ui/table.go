package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// NewTable creates a new table model with some default styling.
func NewTable() table.Model {
	t := table.New(
		// We will set the columns and rows later.
	)

	s := table.DefaultStyles()
	// Use hex colors instead of ANSI codes for consistency
	// Color 212 (pink) -> #FF87D7
	// Color 229 (light yellow) -> #FFFFD7
	// Color 57 (purple) -> #5F00FF
	s.Header = s.Header.Blink(false).Bold(true).Foreground(lipgloss.Color("#FF87D7"))
	s.Selected = s.Selected.Foreground(lipgloss.Color("#FFFFD7")).Background(lipgloss.Color("#5F00FF")).Bold(false)
	t.SetStyles(s)

	return t
}
