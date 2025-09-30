package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleCommandHistoryUp navigates up in command history
func (m *MainModel) handleCommandHistoryUp() (*MainModel, tea.Cmd) {
	if len(m.commandHistory) == 0 {
		return m, nil
	}

	// If we're at the bottom (not in history), save current input
	if m.historyIndex == -1 {
		m.currentInput = m.input.Value()
		m.historyIndex = len(m.commandHistory) - 1
	} else if m.historyIndex > 0 {
		// Move up in history
		m.historyIndex--
	}

	// Set the input to the history entry
	if m.historyIndex >= 0 && m.historyIndex < len(m.commandHistory) {
		m.input.SetValue(m.commandHistory[m.historyIndex])
		m.input.SetCursor(len(m.commandHistory[m.historyIndex]))
	}

	return m, nil
}

// handleCommandHistoryDown navigates down in command history
func (m *MainModel) handleCommandHistoryDown() (*MainModel, tea.Cmd) {
	if m.historyIndex == -1 {
		// Not in history, nothing to do
		return m, nil
	}

	if m.historyIndex < len(m.commandHistory)-1 {
		// Move down in history
		m.historyIndex++
		m.input.SetValue(m.commandHistory[m.historyIndex])
		m.input.SetCursor(len(m.commandHistory[m.historyIndex]))
	} else {
		// Return to current input
		m.historyIndex = -1
		m.input.SetValue(m.currentInput)
		m.input.SetCursor(len(m.currentInput))
	}

	return m, nil
}

// handleHistorySearchUp navigates up in history search results
func (m *MainModel) handleHistorySearchUp() (*MainModel, tea.Cmd) {
	if m.historySearchIndex > 0 {
		m.historySearchIndex--
		// Adjust scroll offset if needed
		if m.historySearchIndex < m.historySearchScrollOffset {
			m.historySearchScrollOffset = m.historySearchIndex
		}
	}
	return m, nil
}

// handleHistorySearchDown navigates down in history search results
func (m *MainModel) handleHistorySearchDown() (*MainModel, tea.Cmd) {
	if m.historySearchIndex < len(m.historySearchResults)-1 {
		m.historySearchIndex++
		// Adjust scroll offset if needed
		if m.historySearchIndex >= m.historySearchScrollOffset+10 {
			m.historySearchScrollOffset = m.historySearchIndex - 9
		}
	}
	return m, nil
}

// handleHistorySearchSelect selects the current history search entry
func (m *MainModel) handleHistorySearchSelect() (*MainModel, tea.Cmd) {
	if len(m.historySearchResults) > 0 && m.historySearchIndex < len(m.historySearchResults) {
		// Set the input value to the selected history entry
		m.input.SetValue(m.historySearchResults[m.historySearchIndex])
		// Exit history search mode
		m.historySearchMode = false
		m.historySearchQuery = ""
		m.historySearchResults = []string{}
		m.historySearchIndex = 0
	}
	return m, nil
}

// handleHistoryModalSelect selects the current entry from history modal
func (m *MainModel) handleHistoryModalSelect() (*MainModel, tea.Cmd) {
	if len(m.commandHistory) > 0 {
		// Use the index directly - the modal display handles the reversal
		if m.historyModalIndex >= 0 && m.historyModalIndex < len(m.commandHistory) {
			m.input.SetValue(m.commandHistory[m.historyModalIndex])
		}
		// Close the modal
		m.showHistoryModal = false
		m.historyModalIndex = 0
		m.historyModalScrollOffset = 0
	}
	return m, nil
}