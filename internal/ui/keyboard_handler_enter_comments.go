package ui

import (
	"strings"

	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// handleCommentLine handles single-line comments
func (m *MainModel) handleCommentLine(command string) (*MainModel, tea.Cmd) {
	// This is a line comment, process it as complete
	// The router will strip it and return empty result
	m.input.Reset()
	// Add to history to show it was entered
	m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()
	// Process the comment (router will strip it)
	_ = router.ProcessCommand(command, m.session, m.sessionManager)
	return m, nil
}

// handleBlockComment handles block comments
func (m *MainModel) handleBlockComment(command string) (*MainModel, tea.Cmd) {
	// Starting a block comment
	if strings.Contains(command, "*/") {
		// Single-line block comment - complete
		m.input.Reset()
		m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		_ = router.ProcessCommand(command, m.session, m.sessionManager)
		return m, nil
	} else if !m.multiLineMode {
		// Multi-line block comment - enter special mode
		m.multiLineMode = true
		m.multiLineBuffer = []string{command}
		m.input.Placeholder = "... (in block comment, end with */)"
		m.input.Reset()
		return m, nil
	}
	return m, nil
}

// handleMultiLineBlockComment handles ending a multi-line block comment
// Returns nil, nil if not handling a block comment
func (m *MainModel) handleMultiLineBlockComment(command string) (*MainModel, tea.Cmd) {
	if strings.HasPrefix(m.multiLineBuffer[0], "/*") && strings.Contains(command, "*/") {
		// End of multi-line block comment
		m.multiLineBuffer = append(m.multiLineBuffer, command)
		fullComment := strings.Join(m.multiLineBuffer, "\n")
		m.multiLineMode = false
		m.multiLineBuffer = nil
		m.input.Placeholder = "Enter CQL command..."
		m.input.Reset()
		
		// Add to history
		for _, line := range strings.Split(fullComment, "\n") {
			m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+line)
		}
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		
		// Process (will be stripped as comment)
		_ = router.ProcessCommand(fullComment, m.session, m.sessionManager)
		return m, nil
	}
	// Not a block comment - return nil to indicate no handling
	return nil, nil
}