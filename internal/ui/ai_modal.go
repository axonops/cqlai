package ui

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/charmbracelet/bubbles/viewport"
	bt "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AIModalState represents the state of the AI modal
type AIModalState int

const (
	AIModalStateGenerating AIModalState = iota
	AIModalStatePreview
	AIModalStateError
	AIModalStateFollowUp // New state for entering follow-up questions
)

// AIModal represents the AI-powered CQL generation modal
type AIModal struct {
	State          AIModalState
	UserRequest    string       // The natural language request
	Plan           *ai.AIResult // The generated plan
	CQL            string       // The rendered CQL
	Error          string       // Error message if any
	Selected       int          // 0: Cancel, 1: Execute, 2: Edit
	Width          int
	Height         int
	ShowPlan       bool           // Toggle between showing plan JSON and CQL
	FollowUpInput  string         // Input for follow-up questions
	CursorPosition int            // Cursor position in follow-up input
	viewport       viewport.Model // Viewport for scrollable content
	viewportReady  bool           // Whether viewport is initialized
	lastContent    string         // Track last content to avoid resetting viewport
	lastWidth      int            // Track last width to detect resize
}

// Update handles messages for the AI modal
func (m *AIModal) Update(msg bt.Msg) bt.Cmd {
	var cmd bt.Cmd

	// Only handle scrolling when in preview mode
	if m.State == AIModalStatePreview {
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return cmd
}

// NewAIModal creates a new AI modal for generating CQL
func NewAIModal(userRequest string) AIModal {
	vp := viewport.New(70, 10) // Width will be adjusted, height is for content area
	vp.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#1a1a1a"))

	return AIModal{
		State:         AIModalStateGenerating,
		UserRequest:   userRequest,
		Selected:      0,
		Width:         80,
		Height:        20,
		ShowPlan:      false,
		viewport:      vp,
		viewportReady: false,
		lastWidth:     0,
	}
}

// SetResult sets the generation result
func (m *AIModal) SetResult(plan *ai.AIResult, cql string) {
	m.Plan = plan
	m.CQL = cql
	m.State = AIModalStatePreview
	// Invalidate the cache to force re-wrapping in renderPreview
	m.lastContent = ""
	m.lastWidth = 0
}

// SetError sets an error state
func (m *AIModal) SetError(err error) {
	m.Error = err.Error()
	m.State = AIModalStateError
}

// NextChoice moves to the next choice
func (m *AIModal) NextChoice() {
	if m.State == AIModalStatePreview {
		// Check if this is an INFO operation
		if m.Plan != nil && m.Plan.Operation == "INFO" {
			// For INFO operations, no choice navigation - just input field
			return
		} else {
			// For regular operations: 0: Cancel, 1: Execute, 2: Edit
			m.Selected = (m.Selected + 1) % 3
		}
	} else if m.State == AIModalStateError {
		m.Selected = 0 // Only cancel available on error
	}
}

// PrevChoice moves to the previous choice
func (m *AIModal) PrevChoice() {
	if m.State == AIModalStatePreview {
		// Check if this is an INFO operation
		if m.Plan != nil && m.Plan.Operation == "INFO" {
			// For INFO operations, no choice navigation - just input field
			return
		} else {
			// For regular operations: 0: Cancel, 1: Execute, 2: Edit
			m.Selected--
			if m.Selected < 0 {
				m.Selected = 2
			}
		}
	}
}

// ToggleView toggles between showing plan JSON and CQL
func (m *AIModal) ToggleView() {
	if m.State == AIModalStatePreview {
		m.ShowPlan = !m.ShowPlan
	}
}

// Render renders the AI modal
func (m *AIModal) Render(screenWidth, screenHeight int, styles *Styles) string {
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

	var content string

	switch m.State {
	case AIModalStateGenerating:
		content = m.renderGenerating(titleStyle, styles)
	case AIModalStatePreview:
		content = m.renderPreview(titleStyle, styles)
	case AIModalStateError:
		content = m.renderError(titleStyle, styles)
	case AIModalStateFollowUp:
		content = m.renderFollowUp(titleStyle, styles)
	}

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

func (m *AIModal) renderGenerating(titleStyle lipgloss.Style, styles *Styles) string {
	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Align(lipgloss.Center).
		Width(m.Width - 4).
		MarginTop(2)

	// Request style
	requestStyle := lipgloss.NewStyle().
		Foreground(styles.AccentText.GetForeground()).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1).
		Align(lipgloss.Center).
		Width(m.Width - 6).
		MarginTop(1).
		MarginBottom(2)

	// Loading animation
	loadingStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Align(lipgloss.Center).
		Width(m.Width - 4).
		MarginTop(2)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("🤖 AI CQL Assistant"),
		messageStyle.Render("Generating an answer from your request..."),
		requestStyle.Render(m.UserRequest),
		loadingStyle.Render("⣾⣽⣻⢿⡿⣟⣯⣷ Processing..."),
		"",
		messageStyle.Render("Press Esc to cancel"),
	)
}

func (m *AIModal) renderPreview(titleStyle lipgloss.Style, styles *Styles) string {
	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Width(m.Width - 4).
		MarginTop(1)

	// Set viewport dimensions
	viewportHeight := 10
	m.viewport.Width = m.Width - 8
	m.viewport.Height = viewportHeight

	// Content box style for viewport wrapper
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Border).
		Width(m.Width-6).
		Height(viewportHeight+2). // +2 for border
		Padding(0, 1)

	// Warning style
	warningStyle := lipgloss.NewStyle().
		Foreground(styles.Warn).
		Bold(true).
		Width(m.Width - 4).
		Align(lipgloss.Center)

	// Build button row based on operation type
	var buttonRow string

	if m.Plan != nil && m.Plan.Operation == "INFO" {
		// For INFO operations, show a single Done button
		doneStyle := lipgloss.NewStyle().Padding(0, 2).
			Foreground(lipgloss.Color("#000000")).
			Background(styles.Ok).
			Bold(true)

		doneBtn := doneStyle.Render("Done")

		buttonRow = lipgloss.NewStyle().
			Width(m.Width - 4).
			Align(lipgloss.Center).
			Render(doneBtn)
	} else {
		// Regular CQL operations
		cancelStyle := lipgloss.NewStyle().Padding(0, 2)
		executeStyle := lipgloss.NewStyle().Padding(0, 2)
		editStyle := lipgloss.NewStyle().Padding(0, 2)

		// Style buttons based on selection
		switch m.Selected {
		case 0: // Cancel selected
			cancelStyle = cancelStyle.
				Foreground(lipgloss.Color("#000000")).
				Background(styles.MutedText.GetForeground()).
				Bold(true)
			executeStyle = executeStyle.
				Foreground(styles.MutedText.GetForeground()).
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.Border)
			editStyle = editStyle.
				Foreground(styles.MutedText.GetForeground()).
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.Border)
		case 1: // Execute selected
			cancelStyle = cancelStyle.
				Foreground(styles.MutedText.GetForeground()).
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.Border)
			if m.Plan != nil && !m.Plan.ReadOnly {
				// Dangerous operation - use warning color
				executeStyle = executeStyle.
					Foreground(lipgloss.Color("#000000")).
					Background(styles.Error).
					Bold(true)
			} else {
				// Safe operation
				executeStyle = executeStyle.
					Foreground(lipgloss.Color("#000000")).
					Background(styles.Ok).
					Bold(true)
			}
			editStyle = editStyle.
				Foreground(styles.MutedText.GetForeground()).
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.Border)
		case 2: // Edit selected
			cancelStyle = cancelStyle.
				Foreground(styles.MutedText.GetForeground()).
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.Border)
			executeStyle = executeStyle.
				Foreground(styles.MutedText.GetForeground()).
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.Border)
			editStyle = editStyle.
				Foreground(lipgloss.Color("#000000")).
				Background(styles.Accent).
				Bold(true)
		}

		cancelBtn := cancelStyle.Render("Cancel")
		executeBtn := executeStyle.Render("Execute")
		editBtn := editStyle.Render("Edit")

		buttonRow = lipgloss.JoinHorizontal(
			lipgloss.Center,
			cancelBtn,
			"   ",
			executeBtn,
			"   ",
			editBtn,
		)
	}

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

	scrollInfo := fmt.Sprintf("(%3.f%%)", m.viewport.ScrollPercent()*100)

	var instructions string
	if m.Plan != nil && m.Plan.Operation == "INFO" {
		instructions = instructionStyle.Render(fmt.Sprintf("Type follow-up • ↑↓/PgUp/PgDn: Scroll %s • Enter: Send • Esc: Done", scrollInfo))
	} else {
		instructions = instructionStyle.Render(fmt.Sprintf("← → / Tab: Navigate • ↑↓/PgUp/PgDn: Scroll %s • Enter: Confirm • P: Toggle Plan/CQL • Esc: Cancel", scrollInfo))
	}

	// Prepare content to display
	var contentTitle string
	var currentContent string

	if m.Plan != nil && m.Plan.Operation == "INFO" {
		contentTitle = m.Plan.InfoTitle
		if contentTitle == "" {
			contentTitle = "Information:"
		}
		currentContent = m.Plan.InfoContent
	} else if m.ShowPlan && m.Plan != nil {
		contentTitle = "Query Plan (JSON):"
		currentContent = ai.FormatPlanAsJSON(m.Plan)
	} else {
		contentTitle = "Generated CQL:"
		currentContent = m.CQL
	}

	// Update viewport content if it has changed or if width changed
	if currentContent != m.lastContent || m.viewport.Width != m.lastWidth {
		lines := strings.Split(currentContent, "\n")
		wrapStyle := lipgloss.NewStyle().Width(m.viewport.Width)
		var wrappedLines []string
		for _, line := range lines {
			wrappedLines = append(wrappedLines, wrapStyle.Render(line))
		}
		wrappedContent := strings.Join(wrappedLines, "\n")
		m.viewport.SetContent(wrappedContent)
		m.viewport.GotoTop()
		m.lastContent = currentContent
		m.lastWidth = m.viewport.Width
	}

	// Add confidence indicator
	confidenceStr := ""
	if m.Plan != nil {
		confidence := int(m.Plan.Confidence * 100)
		confidenceStr = fmt.Sprintf(" (Confidence: %d%%)", confidence)
	}

	parts := []string{
		titleStyle.Render("🤖 AI Assistant"),
		messageStyle.Render(contentTitle + confidenceStr),
		contentStyle.Render(m.viewport.View()),
	}

	// Add warning if present
	if m.Plan != nil && m.Plan.Warning != "" {
		parts = append(parts, warningStyle.Render("⚠️  "+m.Plan.Warning))
	}

	// Add operation type indicator (only for non-INFO operations)
	if m.Plan != nil && m.Plan.Operation != "INFO" {
		opStyle := lipgloss.NewStyle().
			Foreground(styles.AccentText.GetForeground()).
			Align(lipgloss.Center).
			Width(m.Width - 4)

		opType := fmt.Sprintf("Operation: %s | Read-Only: %v",
			strings.ToUpper(m.Plan.Operation),
			m.Plan.ReadOnly)
		parts = append(parts, opStyle.Render(opType))
	}

	// Add follow-up input field for INFO operations
	if m.Plan != nil && m.Plan.Operation == "INFO" {
		// Input field style
		inputStyle := lipgloss.NewStyle().
			Foreground(styles.AccentText.GetForeground()).
			Background(lipgloss.Color("#1a1a1a")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.Border).
			Padding(0, 1).
			Width(m.Width - 6).
			MarginTop(1)

		// Show input field with cursor
		inputContent := m.FollowUpInput
		if m.CursorPosition <= len(inputContent) {
			inputContent = inputContent[:m.CursorPosition] + "█" + inputContent[m.CursorPosition:]
		} else {
			inputContent = inputContent + "█"
		}

		if inputContent == "█" {
			inputContent = "Enter follow-up question...█"
		}

		parts = append(parts, "", inputStyle.Render(inputContent))
	}

	parts = append(parts, "", buttonRow, "", instructions)

	return lipgloss.JoinVertical(lipgloss.Center, parts...)
}

func (m *AIModal) renderFollowUp(titleStyle lipgloss.Style, styles *Styles) string {
	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Align(lipgloss.Center).
		Width(m.Width - 4).
		MarginTop(2)

	// Input field style
	inputStyle := lipgloss.NewStyle().
		Foreground(styles.AccentText.GetForeground()).
		Background(lipgloss.Color("#1a1a1a")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Accent).
		Padding(0, 1).
		Width(m.Width - 6).
		MarginTop(1).
		MarginBottom(2)

	// Show input field with cursor
	inputContent := m.FollowUpInput
	if m.CursorPosition <= len(inputContent) {
		inputContent = inputContent[:m.CursorPosition] + "█" + inputContent[m.CursorPosition:]
	} else {
		inputContent = inputContent + "█"
	}

	// Loading animation
	loadingStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Align(lipgloss.Center).
		Width(m.Width - 4).
		MarginTop(2)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("🤖 AI Assistant"),
		messageStyle.Render("Processing your follow-up question..."),
		inputStyle.Render(inputContent),
		loadingStyle.Render("⣾⣽⣻⢿⡿⣟⣯⣷ Processing..."),
		"",
		messageStyle.Render("Press Esc to cancel"),
	)
}

func (m *AIModal) renderError(titleStyle lipgloss.Style, styles *Styles) string {
	// Error message style
	errorStyle := lipgloss.NewStyle().
		Foreground(styles.Error).
		Align(lipgloss.Center).
		Width(m.Width - 4).
		MarginTop(2)

	// Request style
	requestStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1).
		Align(lipgloss.Center).
		Width(m.Width - 6).
		MarginTop(1).
		MarginBottom(2)

	// Button
	cancelStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("#000000")).
		Background(styles.MutedText.GetForeground()).
		Bold(true)

	cancelBtn := cancelStyle.Render("OK")

	buttonRow := lipgloss.NewStyle().
		Width(m.Width - 4).
		Align(lipgloss.Center).
		Render(cancelBtn)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("🤖 AI CQL Generator - Error"),
		errorStyle.Render("Failed to generate CQL:"),
		requestStyle.Render(m.Error),
		"",
		errorStyle.Render("Request:"),
		requestStyle.Render(m.UserRequest),
		"",
		buttonRow,
		"",
		errorStyle.Render("Press Enter or Esc to close"),
	)
}
