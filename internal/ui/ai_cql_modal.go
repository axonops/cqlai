package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// AICQLModal represents a modal for displaying AI-generated CQL with execution options
type AICQLModal struct {
	Active   bool
	CQL      string // The generated CQL query
	Selected int    // 0: Execute, 1: Cancel
}

// NewAICQLModal creates a new CQL execution modal
func NewAICQLModal(cql string) *AICQLModal {
	return &AICQLModal{
		Active:   true,
		CQL:      cql,
		Selected: 0,
	}
}

// NextChoice moves to the next button
func (m *AICQLModal) NextChoice() {
	if m.Selected < 1 {
		m.Selected++
	}
}

// PrevChoice moves to the previous button
func (m *AICQLModal) PrevChoice() {
	if m.Selected > 0 {
		m.Selected--
	}
}

// Render renders the CQL execution modal
func (m *AICQLModal) Render(screenWidth, screenHeight int, styles *Styles) string {
	if !m.Active {
		return ""
	}

	// Modal dimensions
	modalWidth := 80
	if modalWidth > screenWidth-4 {
		modalWidth = screenWidth - 4
	}

	// Create modal box style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		Padding(1, 2).
		Width(modalWidth)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true).
		Width(modalWidth - 4).
		Align(lipgloss.Center)

	// CQL display style
	cqlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Background(lipgloss.Color("#0D0D0D")).
		Padding(1, 2).
		Width(modalWidth - 8).
		Align(lipgloss.Left)

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Width(modalWidth - 4).
		Align(lipgloss.Center).
		Margin(1, 0)

	// Button styles
	activeButtonStyle := lipgloss.NewStyle().
		Padding(0, 3).
		Foreground(lipgloss.Color("#000000")).
		Background(styles.Accent).
		Bold(true)

	inactiveButtonStyle := lipgloss.NewStyle().
		Padding(0, 3).
		Foreground(styles.MutedText.GetForeground()).
		Background(lipgloss.Color("#2D2D2D"))

	// Create buttons
	executeBtn := "Execute"
	cancelBtn := "Cancel"

	if m.Selected == 0 {
		executeBtn = activeButtonStyle.Render(executeBtn)
		cancelBtn = inactiveButtonStyle.Render(cancelBtn)
	} else {
		executeBtn = inactiveButtonStyle.Render(executeBtn)
		cancelBtn = activeButtonStyle.Render(cancelBtn)
	}

	// Button row
	buttonRow := lipgloss.JoinHorizontal(
		lipgloss.Center,
		executeBtn,
		"  ",
		cancelBtn,
	)
	buttonRowStyle := lipgloss.NewStyle().
		Width(modalWidth - 4).
		Align(lipgloss.Center)

	// Navigation hint
	navHint := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Align(lipgloss.Center).
		Width(modalWidth - 4).
		Render("← → Navigate • Enter: Select • Esc: Cancel")

	// Build the modal content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("🤖 AI Generated CQL"),
		"",
		cqlStyle.Render(m.CQL),
		"",
		messageStyle.Render("Would you like to execute this query?"),
		"",
		buttonRowStyle.Render(buttonRow),
		"",
		navHint,
	)

	// Apply modal style and center on screen
	modal := modalStyle.Render(content)

	// Center the modal on screen
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}