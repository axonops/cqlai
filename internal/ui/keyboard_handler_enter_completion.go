package ui

import (
	tea "charm.land/bubbletea/v2"
)

// handleCompletionSelection handles when Enter is pressed with completions showing
func (m *MainModel) handleCompletionSelection() (*MainModel, tea.Cmd) {
	// Get the selected completion (just the next word)
	selectedCompletion := m.completions[m.completionIndex]

	// A filename is not a word: the directory in front of it has to survive.
	if m.completingPath {
		m.applyPathCompletion(selectedCompletion)
		return m, nil
	}

	newValue := applyCompletion(m.input.Value(), selectedCompletion)

	m.input.SetValue(newValue)
	m.input.SetCursor(len(newValue))

	// Hide completions
	m.showCompletions = false
	m.completions = []string{}
	m.completionIndex = -1
	m.completionScrollOffset = 0
	return m, nil
}
