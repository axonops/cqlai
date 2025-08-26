package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// AIInfoRequestModal represents a modal for AI requesting more information from user
type AIInfoRequestModal struct {
	Active      bool
	Message     string           // The AI's message/question
	Input       textinput.Model  // Text input for user response
	Width       int
	Height      int
}

// NewAIInfoRequestModal creates a new information request modal
func NewAIInfoRequestModal(message string) *AIInfoRequestModal {
	input := textinput.New()
	input.Placeholder = "Type your response..."
	input.Focus()
	input.CharLimit = 500
	input.Width = 50
	
	return &AIInfoRequestModal{
		Active:  true,
		Message: message,
		Input:   input,
		Width:   60,
		Height:  15,
	}
}

// Update handles input for the modal
func (m *AIInfoRequestModal) Update(msg tea.Msg) (*AIInfoRequestModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			// Cancel the modal
			m.Active = false
			return m, nil
		case tea.KeyEnter:
			// Submit the response
			if strings.TrimSpace(m.Input.Value()) != "" {
				m.Active = false
				return m, nil
			}
		default:
			// Handle text input
			var cmd tea.Cmd
			m.Input, cmd = m.Input.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

// GetResponse returns the user's response
func (m *AIInfoRequestModal) GetResponse() string {
	return strings.TrimSpace(m.Input.Value())
}

// Render renders the information request modal
func (m *AIInfoRequestModal) Render(screenWidth, screenHeight int, styles *Styles) string {
	if !m.Active {
		return ""
	}

	// Create modal box style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		BorderBackground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#000000")).
		Padding(1, 2).
		Width(m.Width).
		MaxWidth(screenWidth - 4)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.Width - 4)

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.AccentText.GetForeground()).
		Width(m.Width - 4).
		Align(lipgloss.Left).
		MarginTop(1)

	// AI icon and message
	aiMessageStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a1a")).
		Padding(1).
		Width(m.Width - 6).
		MarginTop(1).
		MarginBottom(1)

	// Input field style
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Accent).
		Padding(0, 1).
		Width(m.Width - 8).
		MarginTop(1)

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Italic(true).
		Align(lipgloss.Center).
		Width(m.Width - 4).
		MarginTop(1)

	// Build the modal content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("🤖 AI Needs More Information"),
		messageStyle.Render("The AI needs clarification to proceed:"),
		aiMessageStyle.Render(m.Message),
		messageStyle.Render("Your response:"),
		inputStyle.Render(m.Input.View()),
		"",
		instructionStyle.Render("Enter: Submit  •  Esc: Cancel"),
	)

	modalBox := modalStyle.Render(content)

	// Center the modal on screen
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#000000")),
	)
}