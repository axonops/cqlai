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
	if msg, ok := msg.(tea.KeyMsg); ok {
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

	// Debug: Log the screen dimensions
	// If screenHeight is too small, something is wrong with how it's calculated
	if screenHeight < 20 {
		// Force a reasonable minimum height
		screenHeight = 24
	}

	// Adjust width for smaller screens
	modalWidth := m.Width
	if screenWidth < modalWidth+4 {
		modalWidth = screenWidth - 4
	}

	// Simple approach: just truncate the message to a few lines
	maxMessageLines := 5
	messageLines := strings.Split(m.Message, "\n")
	if len(messageLines) > maxMessageLines {
		messageLines = messageLines[:maxMessageLines]
		messageLines = append(messageLines, "...")
	}
	truncatedMessage := strings.Join(messageLines, "\n")

	// Create modal box style with minimal padding
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		BorderBackground(lipgloss.Color("#1A1A1A")).
		Background(lipgloss.Color("#1A1A1A")).
		Padding(0, 1). // Reduced vertical padding
		Width(modalWidth)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(modalWidth - 4)

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.AccentText.GetForeground()).
		Width(modalWidth - 4).
		Align(lipgloss.Left)

	// Message box style (instead of viewport)
	messageBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Border).
		Padding(0, 1).
		Width(modalWidth - 6)

	// Input field style
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Accent).
		Padding(0, 1).
		Width(modalWidth - 8)

	// Instructions style
	instructionStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Italic(true).
		Align(lipgloss.Center).
		Width(modalWidth - 4)

	// Build the modal content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("🤖 AI Needs More Information"),
		messageStyle.Render("Please provide more details:"),
		messageBoxStyle.Render(truncatedMessage),
		messageStyle.Render("Your response:"),
		inputStyle.Render(m.Input.View()),
		instructionStyle.Render("Enter: Submit • Esc: Cancel"),
	)

	modalBox := modalStyle.Render(content)
	
	// Add top padding to ensure modal doesn't get cut off
	topPadding := 2
	paddedModal := lipgloss.NewStyle().
		MarginTop(topPadding).
		Render(modalBox)

	// Position the modal with top alignment to prevent cutoff
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Top,
		paddedModal,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1A1A1A")),
	)
}