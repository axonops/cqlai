package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleSpaceKey handles Space key press
func (m MainModel) handleSpaceKey(msg tea.KeyMsg) (MainModel, tea.Cmd) {
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

		// Apply the completion like in handleEnterKey
		currentInput := m.input.Value()
		newValue := ""

		// Special case: if input ends with a dot (keyspace.), just append the table name
		if strings.HasSuffix(currentInput, ".") {
			newValue = currentInput + selectedCompletion
		} else if strings.HasSuffix(currentInput, " ") {
			// Just append the completion
			newValue = currentInput + selectedCompletion
		} else {
			// Check if we have a partial word to replace
			lastSpace := strings.LastIndex(currentInput, " ")
			if lastSpace >= 0 {
				// Check if the last word is a complete token that shouldn't be replaced
				lastWord := currentInput[lastSpace+1:]

				// Check for keyspace.table pattern
				if strings.Contains(lastWord, ".") {
					// For keyspace.table patterns, always replace the part after the dot
					// The completion engine returns just the table name
					dotIndex := strings.LastIndex(currentInput, ".")
					newValue = currentInput[:dotIndex+1] + selectedCompletion
				} else if lastWord == "*" || strings.HasSuffix(lastWord, ")") {
					// Don't replace, just append
					newValue = currentInput + " " + selectedCompletion
				} else {
					// Replace the partial word
					newValue = currentInput[:lastSpace+1] + selectedCompletion
				}
			} else {
				// Replace the entire input (single partial word)
				newValue = selectedCompletion
			}
		}

		// Add a space after the completion
		completedText := newValue + " "
		m.input.SetValue(completedText)
		m.input.SetCursor(len(completedText)) // Move cursor to end after space

		// Hide completions
		m.showCompletions = false
		m.completions = []string{}
		m.completionIndex = -1
		m.completionScrollOffset = 0
		return m, nil
	}

	// If no completions, let the space key be handled normally by passing it to the input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}
