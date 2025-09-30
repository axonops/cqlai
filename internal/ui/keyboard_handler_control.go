package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// handleCtrlC handles Ctrl+C - clear input or exit
func (m *MainModel) handleCtrlC() (*MainModel, tea.Cmd) {
	// If in history search mode, exit it
	if m.historySearchMode {
		m.historySearchMode = false
		m.historySearchQuery = ""
		m.historySearchResults = []string{}
		m.historySearchIndex = 0
		m.historySearchScrollOffset = 0
		return m, nil
	}
	// If modal is showing, close it
	if m.modal.Type != ModalNone {
		m.modal = Modal{Type: ModalNone}
		m.input.Placeholder = "Enter CQL command..."
		m.input.Reset()

		// Add cancellation message to history
		m.fullHistoryContent += "\n" + m.styles.MutedText.Render("Command cancelled.")
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		return m, nil
	}
	// If there's text in the input, clear it. Otherwise check for pagination.
	if m.input.Value() != "" {
		m.input.Reset()
		// Also clear any completions and confirmation
		m.showCompletions = false
		m.completions = []string{}
		m.completionIndex = -1
		m.historyIndex = -1
		m.confirmExit = false
		return m, nil
	}

	// If we're in the middle of paging, cancel it
	if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
		// Clear the "more data" state
		m.slidingWindow.hasMoreData = false
		m.slidingWindow.streamingResult = nil
		m.input.Placeholder = "Enter CQL command..."
		m.input.Focus()
		// Exit navigation mode if active
		if m.navigationMode {
			m.navigationMode = false
		}
		// Add a message to history to indicate paging was cancelled
		m.fullHistoryContent += "\n" + m.styles.MutedText.Render("Paging cancelled. Showing partial results.")
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		return m, nil
	}

	// If already confirming, exit. Otherwise show confirmation.
	if m.confirmExit {
		// Disable mouse tracking on exit
		fmt.Print("\x1b[?1000l") // Disable basic mouse tracking
		fmt.Print("\x1b[?1006l") // Disable SGR mouse mode
		return m, tea.Quit
	}
	m.confirmExit = true
	m.input.SetValue("")
	m.input.Placeholder = "Really exit? (Ctrl+C/Ctrl+D again to confirm, any other key to cancel)"
	return m, nil
}

// handleCtrlD handles Ctrl+D - exit
func (m *MainModel) handleCtrlD() (*MainModel, tea.Cmd) {
	// If confirming exit, quit. Otherwise show confirmation.
	if m.confirmExit {
		// Disable mouse tracking on exit
		fmt.Print("\x1b[?1000l") // Disable basic mouse tracking
		fmt.Print("\x1b[?1006l") // Disable SGR mouse mode
		return m, tea.Quit
	}
	m.confirmExit = true
	m.input.SetValue("")
	m.input.Placeholder = "Really exit? (Ctrl+C/Ctrl+D again to confirm, any other key to cancel)"
	return m, nil
}

// handleCtrlR handles Ctrl+R - toggle history search mode
func (m *MainModel) handleCtrlR() (*MainModel, tea.Cmd) {
	// Toggle history search mode
	if !m.historySearchMode {
		// Enter history search mode
		m.historySearchMode = true
		m.historySearchQuery = ""

		// Search with empty query shows all history
		if m.historyManager != nil {
			m.historySearchResults = m.historyManager.SearchHistory("")
		} else {
			// Fallback to in-memory history
			m.historySearchResults = make([]string, len(m.commandHistory))
			copy(m.historySearchResults, m.commandHistory)
		}

		// Start at the bottom (newest command)
		if len(m.historySearchResults) > 0 {
			m.historySearchIndex = len(m.historySearchResults) - 1
			// Set scroll offset to show the bottom
			if len(m.historySearchResults) > 10 {
				m.historySearchScrollOffset = len(m.historySearchResults) - 10
			} else {
				m.historySearchScrollOffset = 0
			}
		} else {
			m.historySearchIndex = 0
			m.historySearchScrollOffset = 0
		}
	} else {
		// Exit history search mode
		m.historySearchMode = false
		m.historySearchQuery = ""
		m.historySearchResults = []string{}
		m.historySearchIndex = 0
	}
	return m, nil
}

// handleCtrlP handles Ctrl+P - move to previous line in history
func (m *MainModel) handleCtrlP() (*MainModel, tea.Cmd) {
	// If in history search mode, navigate search results
	if m.historySearchMode {
		if m.historySearchIndex > 0 {
			m.historySearchIndex--
			// Adjust scroll offset if needed
			if m.historySearchIndex < m.historySearchScrollOffset {
				m.historySearchScrollOffset = m.historySearchIndex
			}
		}
		return m, nil
	}
	// In navigation mode, scroll the table instead of showing history
	if m.navigationMode {
		return m.handleSingleLineUp()
	}
	// Normal mode - handle command history
	return m.handleCommandHistoryUp()
}