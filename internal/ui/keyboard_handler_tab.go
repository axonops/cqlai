package ui

import (
	"strings"

	"github.com/axonops/cqlai/internal/ui/completion"
	tea "github.com/charmbracelet/bubbletea"
)

// handleTabKey handles Tab key press
func (m *MainModel) handleTabKey() (*MainModel, tea.Cmd) {
	// Cancel exit confirmation if active
	if m.confirmExit {
		m.confirmExit = false
		m.input.Placeholder = "Enter CQL command..."
		return m, nil
	}

	currentInput := m.input.Value()

	// If completions are already showing, cycle through them
	if m.showCompletions && len(m.completions) > 0 {
		// Just cycle the selection, don't apply yet
		m.completionIndex = (m.completionIndex + 1) % len(m.completions)

		// Reset scroll to top when wrapping around
		if m.completionIndex == 0 {
			m.completionScrollOffset = 0
		}

		// Adjust scroll offset if selection moves out of view
		if m.completionIndex >= m.completionScrollOffset+10 {
			m.completionScrollOffset = m.completionIndex - 9
		}
		return m, nil
	}

	// If input doesn't end with space and we have content, add a space
	// This allows tab completion to continue after accepting a completion
	if currentInput != "" && !strings.HasSuffix(currentInput, " ") {
		// Check if the last word looks complete (is a valid CQL keyword)
		words := strings.Fields(strings.ToUpper(currentInput))
		if len(words) > 0 {
			lastWord := words[len(words)-1]
			// Check if last word is a complete keyword
			if completion.IsCompleteKeyword(lastWord) {
				// Add a space and get next completions
				currentInput = currentInput + " "
				m.input.SetValue(currentInput)
				m.input.SetCursor(len(currentInput))
			} else {
				// Special case: If we have INSERT INTO keyspace.table, treat it as complete
				upperInput := strings.ToUpper(currentInput)
				if strings.HasPrefix(upperInput, "INSERT INTO ") && strings.Contains(lastWord, ".") {
					// This is a keyspace.table reference after INSERT INTO
					// Don't add space, let the completion engine handle it
				}
			}
		}
	}

	// Get completions for current input
	m.completions = m.completionEngine.Complete(currentInput)

	if len(m.completions) == 0 {
		// No completions available
		m.showCompletions = false
		m.completionIndex = -1
		m.completionScrollOffset = 0
	} else if len(m.completions) == 1 {
		// Single completion - apply it immediately
		selectedCompletion := m.completions[0]

		// Apply the completion by appending to current input
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

		m.input.SetValue(newValue)
		m.input.SetCursor(len(newValue))

		// Clear completions
		m.showCompletions = false
		m.completions = []string{}
		m.completionIndex = -1
		m.completionScrollOffset = 0
	} else {
		// Multiple completions - show modal
		m.showCompletions = true
		m.completionIndex = 0        // Start with first item selected
		m.completionScrollOffset = 0 // Reset scroll position
	}

	return m, nil
}
