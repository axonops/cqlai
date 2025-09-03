package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// AIInfoRequestModal represents a modal for AI requesting more information from user
type AIInfoRequestModal struct {
	Active      bool
	Message     string           // The AI's message/question
	Input       textinput.Model  // Text input for user response
	Viewport    viewport.Model  // Viewport for scrollable message
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
	
	vp := viewport.New(56, 5) // Initial size, will be adjusted in Render
	vp.SetContent(message)
	
	return &AIInfoRequestModal{
		Active:   true,
		Message:  message,
		Input:    input,
		Viewport: vp,
		Width:    60,
		Height:   15,
	}
}

// Update handles input for the modal
func (m *AIInfoRequestModal) Update(msg tea.Msg) (*AIInfoRequestModal, tea.Cmd) {
	var cmds []tea.Cmd
	
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
		case tea.KeyUp, tea.KeyPgUp:
			// Scroll the viewport up
			var cmd tea.Cmd
			m.Viewport, cmd = m.Viewport.Update(msg)
			cmds = append(cmds, cmd)
		case tea.KeyDown, tea.KeyPgDown:
			// Scroll the viewport down
			var cmd tea.Cmd
			m.Viewport, cmd = m.Viewport.Update(msg)
			cmds = append(cmds, cmd)
		default:
			// Handle text input
			var cmd tea.Cmd
			m.Input, cmd = m.Input.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
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

	// Adjust width for smaller screens
	modalWidth := m.Width
	if screenWidth < modalWidth+4 {
		modalWidth = screenWidth - 4
	}

	// Calculate dynamic viewport height based on screen size
	// Be very conservative with height to prevent overflow
	// Reserve space for: modal border/padding (4), title (3), "Please provide" (2),
	// "Your response" (2), input box (3), instructions (2), margins (4), top padding (2)
	reservedHeight := 22
	
	// Calculate viewport height (minimum 3, maximum 10 lines)
	availableHeight := screenHeight - reservedHeight
	if availableHeight < 0 {
		availableHeight = 3
	}
	viewportHeight := min(availableHeight, 10) // Cap at 10 lines max
	if viewportHeight < 3 {
		viewportHeight = 3 // Minimum 3 lines
	}
	
	// Update viewport dimensions if they've changed
	if m.Viewport.Width != modalWidth-6 || m.Viewport.Height != viewportHeight {
		m.Viewport.Width = modalWidth - 6
		m.Viewport.Height = viewportHeight
		// Re-set content to ensure it's wrapped correctly
		m.Viewport.SetContent(m.Message)
	}

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

	// Viewport wrapper style
	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Border).
		Width(modalWidth - 6).
		Height(viewportHeight + 2) // +2 for border, no margins

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
	
	// Add scroll indicator if content is scrollable
	scrollHint := ""
	if m.Viewport.TotalLineCount() > m.Viewport.Height {
		scrollHint = " • ↑↓: Scroll"
	}

	// Build the modal content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("🤖 AI Needs More Information"),
		messageStyle.Render("Please provide more details:"),
		viewportStyle.Render(m.Viewport.View()),
		messageStyle.Render("Your response:"),
		inputStyle.Render(m.Input.View()),
		instructionStyle.Render("Enter: Submit • Esc: Cancel" + scrollHint),
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