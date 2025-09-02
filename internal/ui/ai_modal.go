package ui

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AIModalState represents the state of the AI modal
type AIModalState int

const (
	AIModalStateGenerating AIModalState = iota
	AIModalStatePreview
	AIModalStateError
	AIModalStateFollowUp     // New state for entering follow-up questions
	AIModalStateInfoFollowUp // New state for the info follow-up input
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
	ScreenHeight   int            // Current screen height for dynamic sizing
	ShowPlan       bool           // Toggle between showing plan JSON and CQL
	FollowUpInput  string         // Input for follow-up questions
	CursorPosition int            // Cursor position in follow-up input
	viewport       viewport.Model // Viewport for scrollable content
	viewportReady  bool           // Whether viewport is initialized
	lastContent    string         // Track last content to avoid resetting viewport
	lastWidth      int            // Track last width to detect resize
}

// Update handles messages for the AI modal
func (m *AIModal) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Import logger for debugging
	logger.DebugfToFile("AI", "AIModal.Update called with msg type: %T", msg)

	// Always pass window size messages to viewport
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		logger.DebugfToFile("AI", "AIModal: WindowSizeMsg received, width=%d, height=%d", msg.Width, msg.Height)
		m.ScreenHeight = msg.Height
		// Don't set viewport dimensions here, let renderPreview handle it dynamically
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		logger.DebugfToFile("AI", "AIModal: KeyMsg received, key=%s, state=%v", msg.String(), m.State)
		// Only handle key messages if we're in preview state
		if m.State == AIModalStatePreview {
			oldYOffset := m.viewport.YOffset
			oldScrollPercent := m.viewport.ScrollPercent()

			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)

			logger.DebugfToFile("AI", "AIModal: After viewport.Update - YOffset: %d->%d, ScrollPercent: %.2f->%.2f, TotalLines: %d, Height: %d",
				oldYOffset, m.viewport.YOffset,
				oldScrollPercent, m.viewport.ScrollPercent(),
				m.viewport.TotalLineCount(), m.viewport.Height)
		} else {
			logger.DebugfToFile("AI", "AIModal: Ignoring key because state is %v", m.State)
		}
	case tea.MouseMsg:
		logger.DebugfToFile("AI", "AIModal: MouseMsg received")
		// Handle mouse scroll events if we're in preview state
		if m.State == AIModalStatePreview {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

// NewAIModal creates a new AI modal for generating CQL
func NewAIModal(userRequest string) AIModal {
	vp := viewport.New(70, 10) // Width will be adjusted, height is for content area
	vp.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#2D2D2D"))
	vp.KeyMap = viewport.DefaultKeyMap() // Enable keyboard navigation

	return AIModal{
		State:         AIModalStateGenerating,
		UserRequest:   userRequest,
		Selected:      0,
		Width:         80,
		Height:        20,
		ScreenHeight:  24, // Default screen height
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
	switch m.State {
	case AIModalStatePreview:
		if m.Plan != nil && m.Plan.Operation == "INFO" {
			m.Selected = (m.Selected + 1) % 2 // 0: Done, 1: Reply
		} else {
			m.Selected = (m.Selected + 1) % 3 // 0: Cancel, 1: Execute, 2: Edit
		}
	case AIModalStateError:
		m.Selected = 0 // Only cancel available on error
	}
}

// PrevChoice moves to the previous choice
func (m *AIModal) PrevChoice() {
	if m.State == AIModalStatePreview {
		if m.Plan != nil && m.Plan.Operation == "INFO" {
			m.Selected--
			if m.Selected < 0 {
				m.Selected = 1
			}
		} else {
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
	// Store the screen height for dynamic sizing
	if screenHeight > 0 {
		m.ScreenHeight = screenHeight
	}
	// Create modal box style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		BorderBackground(lipgloss.Color("#1A1A1A")).
		Background(lipgloss.Color("#1A1A1A")).
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
		content = m.renderPreview(titleStyle, styles, screenHeight)
	case AIModalStateError:
		content = m.renderError(titleStyle, styles)
	case AIModalStateFollowUp:
		content = m.renderFollowUp(titleStyle, styles)
	case AIModalStateInfoFollowUp:
		content = m.renderInfoFollowUp(titleStyle, styles)
	}

	modalBox := modalStyle.Render(content)

	// Center the modal on screen
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1A1A1A")),
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
		Background(lipgloss.Color("#2D2D2D")).
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

func (m *AIModal) renderPreview(titleStyle lipgloss.Style, styles *Styles, screenHeight int) string {

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Width(m.Width - 4).
		MarginTop(1)

	// Calculate dynamic viewport height based on screen size
	// Reserve space for: modal border/padding (4), title (2), message (2), buttons (3),
	// instructions (3), operation indicator (2), warning if present (2)
	reservedHeight := 16
	if m.Plan != nil && m.Plan.Warning != "" {
		reservedHeight += 2 // Extra space for warning
	}
	if m.Plan != nil && m.Plan.Operation != "INFO" {
		reservedHeight += 2 // Extra space for operation type indicator
	}

	// Calculate viewport height (minimum 5, maximum based on screen)
	viewportHeight := max(5, min(screenHeight-reservedHeight, 25))

	// Set viewport dimensions if they've changed
	if m.viewport.Width != m.Width-6 || m.viewport.Height != viewportHeight {
		m.viewport.Width = m.Width - 6 // Adjusted to match content box width
		m.viewport.Height = viewportHeight
		m.viewportReady = true
	}

	// Content box style for viewport wrapper
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Border).
		Width(m.Width - 6).
		Height(viewportHeight + 2) // +2 for border

	// Warning style
	warningStyle := lipgloss.NewStyle().
		Foreground(styles.Warn).
		Bold(true).
		Width(m.Width - 4).
		Align(lipgloss.Center)

	// Build button row and input based on operation type
	var buttonRow string
	var replyInput string

	if m.Plan != nil && m.Plan.Operation == "INFO" {
		// Input field for follow-up questions
		inputStyle := lipgloss.NewStyle().
			Foreground(styles.AccentText.GetForeground()).
			Background(lipgloss.Color("#2D2D2D")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.Border).
			Padding(0, 1).
			Width(m.Width - 8).
			MarginTop(1)

		// Show input field with cursor
		inputContent := m.FollowUpInput
		if m.CursorPosition <= len(inputContent) {
			inputContent = inputContent[:m.CursorPosition] + "│"
		}

		if inputContent == "" || inputContent == "│" {
			inputContent = "Type your follow-up question here..."
			if m.CursorPosition == 0 {
				inputContent = "│" + inputContent
			}
		}

		replyInput = inputStyle.Render(inputContent)

		// Simple Done button
		doneStyle := lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(styles.MutedText.GetForeground()).
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.Border)

		buttonRow = doneStyle.Render("Done (Esc)")
	} else {
		// Regular CQL operations
		cancelStyle := lipgloss.NewStyle().Padding(0, 2)
		executeStyle := lipgloss.NewStyle().Padding(0, 2)
		editStyle := lipgloss.NewStyle().Padding(0, 2)

		// Style buttons based on selection
		switch m.Selected {
		case 0: // Cancel selected
			cancelStyle = cancelStyle.
				Foreground(lipgloss.Color("#1A1A1A")).
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
					Foreground(lipgloss.Color("#1A1A1A")).
					Background(styles.Error).
					Bold(true)
			} else {
				// Safe operation
				executeStyle = executeStyle.
					Foreground(lipgloss.Color("#1A1A1A")).
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
				Foreground(lipgloss.Color("#1A1A1A")).
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
		instructions = instructionStyle.Render(fmt.Sprintf("↑↓/PgUp/PgDn: Scroll %s • Enter: Send follow-up • Esc: Done", scrollInfo))
	} else {
		instructions = instructionStyle.Render(fmt.Sprintf("← → / Tab: Navigate • ↑↓/PgUp/PgDn: Scroll %s • Enter: Confirm • P: Toggle Plan/CQL • Esc: Cancel", scrollInfo))
	}

	// Prepare content to display
	var contentTitle string
	var currentContent string

	if m.Plan != nil && m.Plan.Operation == "INFO" { //nolint:gocritic // more readable as if
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
	contentChanged := currentContent != m.lastContent
	widthChanged := m.viewport.Width != m.lastWidth

	// logger.DebugfToFile("AI", "renderPreview: Content check - changed=%v, len(current)=%d, len(last)=%d",
	// 	contentChanged, len(currentContent), len(m.lastContent))
	// logger.DebugfToFile("AI", "renderPreview: Width check - changed=%v, current=%d, last=%d",
	// 	widthChanged, m.viewport.Width, m.lastWidth)

	if contentChanged || widthChanged {
		logger.DebugfToFile("AI", "renderPreview: Setting viewport content because content or width changed")
		logger.DebugfToFile("AI", "renderPreview: Setting viewport content, contentLen=%d, lines=%d",
			len(currentContent), strings.Count(currentContent, "\n")+1)

		// Process content line by line for proper wrapping
		lines := strings.Split(currentContent, "\n")
		wrapStyle := lipgloss.NewStyle().Width(m.viewport.Width)
		var wrappedLines []string
		for _, line := range lines {
			// Wrap each line to fit viewport width
			wrapped := wrapStyle.Render(line)
			wrappedLines = append(wrappedLines, wrapped)
		}
		wrappedContent := strings.Join(wrappedLines, "\n")
		m.viewport.SetContent(wrappedContent)

		logger.DebugfToFile("AI", "renderPreview: After SetContent - TotalLines=%d, Height=%d, YOffset=%d",
			m.viewport.TotalLineCount(), m.viewport.Height, m.viewport.YOffset)

		// Only go to top if this is new content, not just a width change
		if currentContent != m.lastContent {
			m.viewport.GotoTop()
			logger.DebugfToFile("AI", "renderPreview: Called GotoTop because content changed")
			m.lastContent = currentContent
		}
		m.lastWidth = m.viewport.Width
		m.viewportReady = true // Mark viewport as ready after content is set
	}
	//  else {
	// 	logger.DebugfToFile("AI", "renderPreview: Content unchanged, not calling SetContent. YOffset=%d", m.viewport.YOffset)
	// }

	// Add confidence indicator
	confidenceStr := ""
	if m.Plan != nil {
		confidence := int(m.Plan.Confidence * 100)
		confidenceStr = fmt.Sprintf(" (Confidence: %d%%)", confidence)
	}

	viewportContent := m.viewport.View()
	// logger.DebugfToFile("AI", "renderPreview: viewport.View() returned %d chars, YOffset=%d, ScrollPercent=%.2f",
	// 	len(viewportContent), m.viewport.YOffset, m.viewport.ScrollPercent())

	// Simply use the viewport content for now
	// We'll show scroll position in the instructions instead of a visual scrollbar

	parts := []string{
		titleStyle.Render("🤖 AI Assistant"),
		messageStyle.Render(contentTitle + confidenceStr),
		contentStyle.Render(viewportContent),
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

	// Add reply input if INFO operation
	if m.Plan != nil && m.Plan.Operation == "INFO" {
		parts = append(parts, "", replyInput, "", buttonRow, "", instructions)
	} else {
		parts = append(parts, "", buttonRow, "", instructions)
	}

	return lipgloss.JoinVertical(lipgloss.Center, parts...)
}

func (m *AIModal) renderInfoFollowUp(titleStyle lipgloss.Style, styles *Styles) string {
	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Width(m.Width - 4)

	// Previous message box style
	previousMsgStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Border).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#2D2D2D")).
		Padding(1).
		Width(m.Width - 6).
		MaxHeight(10).
		MarginTop(1)

	// Get the previous message content
	previousContent := ""
	if m.Plan != nil && m.Plan.Operation == "INFO" {
		previousContent = m.Plan.InfoContent
		// Truncate if too long for display
		lines := strings.Split(previousContent, "\n")
		if len(lines) > 8 {
			previousContent = strings.Join(lines[:8], "\n") + "\n..."
		}
	}

	// Input field style
	inputStyle := lipgloss.NewStyle().
		Foreground(styles.AccentText.GetForeground()).
		Background(lipgloss.Color("#2D2D2D")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Accent).
		Padding(0, 1).
		Width(m.Width - 6).
		MarginTop(1).
		MarginBottom(1)

	// Show input field with cursor
	inputContent := m.FollowUpInput
	if m.CursorPosition <= len(inputContent) {
		inputContent = inputContent[:m.CursorPosition] + "█" + inputContent[m.CursorPosition:]
	} else {
		inputContent += "█"
	}

	if inputContent == "█" {
		inputContent = "Enter your follow-up question...█"
	}

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Italic(true).
		Align(lipgloss.Center).
		Width(m.Width - 4)

	instructions := instructionStyle.Render("Enter: Submit  •  Esc: Back to message")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("🤖 AI Assistant - Reply"),
		messageStyle.Render("Previous response:"),
		previousMsgStyle.Render(previousContent),
		"",
		messageStyle.Render("Your follow-up question:"),
		inputStyle.Render(inputContent),
		"",
		instructions,
	)
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
		Background(lipgloss.Color("#2D2D2D")).
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
		inputContent += "█"
	}

	if inputContent == "█" {
		inputContent = "Enter follow-up question...█"
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
		Background(lipgloss.Color("#2D2D2D")).
		Padding(0, 1).
		Align(lipgloss.Center).
		Width(m.Width - 6).
		MarginTop(1).
		MarginBottom(2)

	// Button
	cancelStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("#1A1A1A")).
		Background(styles.MutedText.GetForeground()).
		Bold(true)

	cancelBtn := cancelStyle.Render("OK")

	buttonRow := lipgloss.NewStyle().
		Width(m.Width - 4).
		Align(lipgloss.Center).
		Render(cancelBtn)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("🤖 AI CQL Helper - Error"),
		errorStyle.Render("I failed:"),
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
