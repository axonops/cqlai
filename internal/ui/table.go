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
	s.Header = s.Header.Blink(false).Bold(true).Foreground(lipgloss.Color("212"))
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	t.SetStyles(s)

	return t
}
