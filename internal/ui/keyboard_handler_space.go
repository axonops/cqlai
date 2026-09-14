package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/ui/completion"
)

// handleSpaceKey handles Space key press
func (m *MainModel) handleSpaceKey(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	// AI conversation view input is handled in handleKeyboardInput

	// If we have completions showing, accept the current one and add space
	if m.showCompletions && len(m.completions) > 0 {
		// Get the selected completion (just the next word)
		var selectedCompletion string
		if m.completionIndex >= 0 && m.completionIndex < len(m.completions) {
			selectedCompletion = m.completions[m.completionIndex]
		} else if len(m.completions) > 0 {
			// If no specific selection, just hide completions and pass through the space
			m.showCompletions = false
			m.completions = []string{}
			m.completionIndex = -1
			m.completionScrollOffset = 0
			// Let the space key be handled normally
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		// A hint is a note about what to type, so the space is the first
		// character of what is being typed rather than the end of a word.
		if completion.IsHint(selectedCompletion) {
			m.clearCompletions()
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		// Applied the same way Enter and Tab apply one, space included: which
		// key was pressed is not a reason to get a different answer.
		completedText := applyCompletion(m.input.Value(), selectedCompletion)
		m.input.SetValue(completedText)
		m.input.SetCursor(len(completedText))

		// Hide completions
		m.showCompletions = false
		m.completions = []string{}
		m.completionIndex = -1
		m.completionScrollOffset = 0
		return m, nil
	}

	// Check if we should page down when there's more data
	if m.viewMode == "table" && m.slidingWindow != nil && m.slidingWindow.hasMoreData {
		// If input is empty and we're in table view with more data, use Space for paging.
		//
		// Not while a statement is being typed over several lines: the line is
		// empty at the start of every one of them, and a space there is the
		// first character of the next line rather than a page down.
		if m.input.Value() == "" && !m.multiLineMode {
			// Page down (same as PgDn)
			return m.handlePageDown(msg)
		}
	}

	// Otherwise, let the space key be handled normally by passing it to the input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}
