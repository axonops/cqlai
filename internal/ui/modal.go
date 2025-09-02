package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// ModalType represents the type of modal
type ModalType int

const (
	ModalNone ModalType = iota
	ModalConfirmDangerous
)

// Modal represents a modal dialog
type Modal struct {
	Type        ModalType
	Title       string
	Message     string
	Command     string
	Choices     []string
	Selected    int
	Width       int
	Height      int
}

// NewConfirmationModal creates a new confirmation modal for dangerous commands
func NewConfirmationModal(command string) Modal {
	return Modal{
		Type:     ModalConfirmDangerous,
		Title:    "⚠️  Confirm Destructive Command",
		Message:  "This command may permanently modify or delete data:",
		Command:  command,
		Choices:  []string{"Cancel", "Execute"},
		Selected: 0, // Default to Cancel for safety
		Width:    60,
		Height:   10,
	}
}

// NextChoice moves to the next choice
func (m *Modal) NextChoice() {
	m.Selected = (m.Selected + 1) % len(m.Choices)
}

// PrevChoice moves to the previous choice
func (m *Modal) PrevChoice() {
	m.Selected--
	if m.Selected < 0 {
		m.Selected = len(m.Choices) - 1
	}
}

// Render renders the modal centered on screen
func (m Modal) Render(screenWidth, screenHeight int, styles *Styles, background string) string {
	if m.Type == ModalNone {
		return ""
	}

	// Create modal box style with a solid background to prevent bleed-through
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Warn).
		BorderBackground(lipgloss.Color("#1A1A1A")).
		Background(lipgloss.Color("#1A1A1A")).
		Padding(1, 2).
		Width(m.Width).
		MaxWidth(screenWidth - 4)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Warn).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.Width - 4)

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Align(lipgloss.Center).
		Width(m.Width - 4).
		MarginTop(1)

	// Command style - show the actual command in a box
	commandStyle := lipgloss.NewStyle().
		Foreground(styles.AccentText.GetForeground()).
		Background(lipgloss.Color("#2D2D2D")).
		Padding(0, 1).
		Align(lipgloss.Center).
		Width(m.Width - 6).
		MarginTop(1).
		MarginBottom(1)

	// Build button row - simpler approach
	cancelStyle := lipgloss.NewStyle().Padding(0, 2)
	executeStyle := lipgloss.NewStyle().Padding(0, 2)
	
	if m.Selected == 0 { // Cancel selected
		cancelStyle = cancelStyle.
			Foreground(lipgloss.Color("#1A1A1A")).
			Background(styles.Ok).
			Bold(true)
		executeStyle = executeStyle.
			Foreground(styles.MutedText.GetForeground()).
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.Border)
	} else { // Execute selected
		cancelStyle = cancelStyle.
			Foreground(styles.MutedText.GetForeground()).
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.Border)
		executeStyle = executeStyle.
			Foreground(lipgloss.Color("#1A1A1A")).
			Background(styles.Error).
			Bold(true)
	}
	
	cancelBtn := cancelStyle.Render("Cancel")
	executeBtn := executeStyle.Render("Execute")
	
	// Create the button row with proper spacing
	buttonRow := lipgloss.JoinHorizontal(
		lipgloss.Center,
		cancelBtn,
		"     ",
		executeBtn,
	)
	
	// Center the entire button row
	buttonRow = lipgloss.NewStyle().
		Width(m.Width - 4).
		Align(lipgloss.Center).
		Render(buttonRow)

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Italic(true).
		Align(lipgloss.Center).
		Width(m.Width - 4)

	instructions := instructionStyle.Render("← → / Tab: Navigate  •  Enter: Confirm  •  Esc: Cancel")

	// Combine all elements with proper spacing
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(m.Title),
		messageStyle.Render(m.Message),
		commandStyle.Render(m.Command),
		"", // Empty line for spacing
		buttonRow,
		"", // Empty line for spacing
		instructions,
	)

	modalBox := modalStyle.Render(content)
	
	// Place the modal in the center with a black background
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1A1A1A")),
	)
}