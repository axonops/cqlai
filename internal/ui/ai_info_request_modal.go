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
	input.Width = 40
	
	return &AIInfoRequestModal{
		Active:  true,
		Message: message,
		Input:   input,
		Width:   50,
		Height:  10,
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

// Render renders the information request modal - following the exact pattern of Modal.Render
func (m *AIInfoRequestModal) Render(screenWidth, screenHeight int, styles *Styles) string {
	if !m.Active {
		return ""
	}

	// Create modal box style - exact same as Modal
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		BorderBackground(lipgloss.Color("#1A1A1A")).
		Background(lipgloss.Color("#1A1A1A")).
		Padding(1, 2).
		Width(m.Width).
		MaxWidth(screenWidth - 4)

	// Title style - same as Modal
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.Width - 4)

	// Message style - same as Modal  
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Align(lipgloss.Center).
		Width(m.Width - 4).
		MarginTop(1)

	// Input container style - similar to command style in Modal
	inputStyle := lipgloss.NewStyle().
		Foreground(styles.AccentText.GetForeground()).
		Background(lipgloss.Color("#2D2D2D")).
		Padding(0, 1).
		Align(lipgloss.Left).
		Width(m.Width - 6).
		MarginTop(1).
		MarginBottom(1)

	// Instructions - same as Modal
	instructionStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Italic(true).
		Align(lipgloss.Center).
		Width(m.Width - 4)

	instructions := instructionStyle.Render("Enter: Submit  •  Esc: Cancel")

	// Show only first 2 lines of message to keep it compact
	messageLines := strings.Split(m.Message, "\n")
	shortMessage := messageLines[0]
	if len(messageLines) > 1 && strings.TrimSpace(messageLines[1]) != "" {
		shortMessage += "\n" + messageLines[1]
	}
	if len(messageLines) > 2 {
		shortMessage += "..."
	}

	// Combine all elements - same structure as Modal
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("🤖 AI Needs More Info"),
		messageStyle.Render(shortMessage),
		inputStyle.Render(m.Input.View()),
		"", // Empty line for spacing
		instructions,
	)

	modalBox := modalStyle.Render(content)
	
	// Place the modal - exact same as Modal
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1A1A1A")),
	)
}