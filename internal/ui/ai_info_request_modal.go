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

	// Keep modal compact - max width of 50
	modalWidth := 50
	if screenWidth < modalWidth+4 {
		modalWidth = screenWidth - 4
	}

	// Very simple approach: show only first non-empty line of message
	messageLines := strings.Split(m.Message, "\n")
	displayMessage := ""
	if len(messageLines) > 0 {
		// Take first non-empty line as main message
		for _, line := range messageLines {
			if strings.TrimSpace(line) != "" {
				displayMessage = line
				if len(displayMessage) > 45 {
					displayMessage = displayMessage[:42] + "..."
				}
				break
			}
		}
	}
	if displayMessage == "" {
		displayMessage = "Please provide more information..."
	}

	// Create modal box style - compact
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		BorderBackground(lipgloss.Color("#1A1A1A")).
		Background(lipgloss.Color("#1A1A1A")).
		Padding(1, 2).
		Width(modalWidth)

	// Title style - compact
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(modalWidth - 4)

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Width(modalWidth - 4).
		Align(lipgloss.Center)

	// Input field style
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Accent).
		Padding(0, 1).
		Width(modalWidth - 8)

	// Instructions style - compact
	instructionStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Italic(true).
		Align(lipgloss.Center).
		Width(modalWidth - 4)

	// Build the modal content - very simple
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("Need More Info"),
		"",
		messageStyle.Render(displayMessage),
		"",
		inputStyle.Render(m.Input.View()),
		"",
		instructionStyle.Render("Enter: Submit • Esc: Cancel"),
	)

	modalBox := modalStyle.Render(content)
	
	// Use center positioning like other modals
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1A1A1A")),
	)
}