package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleCompletionSelection handles when Enter is pressed with completions showing
func (m *MainModel) handleCompletionSelection() (*MainModel, tea.Cmd) {
	// Get the selected completion (just the next word)
	selectedCompletion := m.completions[m.completionIndex]

	// Apply the completion by appending to current input
	currentInput := m.input.Value()
	newValue := ""

	// Special case: if input ends with a dot (keyspace.), just append the table name
	if strings.HasSuffix(currentInput, ".") { //nolint:gocritic // more readable as if
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
			if strings.Contains(lastWord, ".") { //nolint:gocritic // more readable as if
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

	m.input.SetValue(newValue)
	m.input.SetCursor(len(newValue))

	// Hide completions
	m.showCompletions = false
	m.completions = []string{}
	m.completionIndex = -1
	m.completionScrollOffset = 0
	return m, nil
}