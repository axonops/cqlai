package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
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
		m.closeHistorySearch()
	}
	return m, nil
}

// typeIntoHistorySearch adds a keypress to the search query and re-filters.
//
// Both the ordinary typing path and the space key come through here: a space is
// part of what you are searching for, and it used to be swallowed by the pager
// before it got this far.
func (m *MainModel) typeIntoHistorySearch(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	switch msg.String() {
	case "backspace", "delete":
		if len(m.historySearchQuery) > 0 {
			m.historySearchQuery = m.historySearchQuery[:len(m.historySearchQuery)-1]
		}
	default:
		if len(msg.Text) > 0 && len(m.historySearchQuery) < 100 {
			m.historySearchQuery += msg.Text
		}
	}

	if m.historyManager != nil {
		m.historySearchResults = m.historyManager.SearchHistory(m.historySearchQuery)
	} else {
		// Fallback to in-memory history search
		m.historySearchResults = []string{}
		queryLower := strings.ToLower(m.historySearchQuery)
		for i := len(m.commandHistory) - 1; i >= 0; i-- {
			if strings.Contains(strings.ToLower(m.commandHistory[i]), queryLower) {
				m.historySearchResults = append(m.historySearchResults, m.commandHistory[i])
			}
		}
	}

	// Start on the newest match, with the window showing the end of the list.
	m.historySearchIndex = max(len(m.historySearchResults)-1, 0)
	m.historySearchScrollOffset = max(len(m.historySearchResults)-historyModalRows, 0)
	return m, nil
}
