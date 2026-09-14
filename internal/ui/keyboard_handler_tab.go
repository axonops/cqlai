package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/ui/completion"
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

	// A file path completes against the filesystem rather than the schema.
	// Every command that takes a file used to offer nothing after the keyword,
	// so paths had to be typed out in full and correctly.
	if updated, cmd, handled := m.completePathAtPrompt(currentInput); handled {
		return updated, cmd
	}

	// If in multi-line mode, combine the buffer with current input for completion
	fullInput := currentInput
	if m.multiLineMode && len(m.multiLineBuffer) > 0 {
		fullInput = strings.Join(m.multiLineBuffer, " ") + " " + currentInput
	}

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
		// Check if input ends with a complete value
		upperInput := strings.ToUpper(currentInput)
		endsWithCompleteValue := false

		// Check for various complete value patterns
		switch {
		case strings.HasSuffix(currentInput, "'") || strings.HasSuffix(currentInput, "\""):
			// A quoted string, if the quote is the one that closes it. An open
			// quote is the start of a value, and the space put after it went
			// inside the value: `compaction = {'` became `{' ` and the key
			// picked from the list landed after the stray quote.
			endsWithCompleteValue = quotesClosed(currentInput)
		case strings.HasSuffix(upperInput, "TRUE") || strings.HasSuffix(upperInput, "FALSE"):
			// Check if this is after an equals sign (boolean assignment)
			if strings.Contains(currentInput, "=") {
				endsWithCompleteValue = true
			}
		case len(currentInput) > 0:
			// Check if ends with a number (after equals)
			lastChar := currentInput[len(currentInput)-1]
			if lastChar >= '0' && lastChar <= '9' && strings.Contains(currentInput, "=") {
				endsWithCompleteValue = true
			}
		}

		if endsWithCompleteValue {
			// This is a complete value, add space for next completion
			currentInput += " "
			m.input.SetValue(currentInput)
			m.input.SetCursor(len(currentInput))
		} else {
			// Check if the last word looks complete (is a valid CQL keyword)
			words := strings.Fields(strings.ToUpper(currentInput))
			if len(words) > 0 {
				lastWord := words[len(words)-1]
				shouldAddSpace := false

				// Check if last word is a complete keyword
				if completion.IsCompleteKeyword(lastWord) {
					shouldAddSpace = true
				} else if len(words) >= 2 {
					// Special case for COPY command: after table name, add space
					if words[0] == "COPY" {
						// Check if we're after a table name (could be keyspace.table)
						// COPY tablename -> add space
						// COPY keyspace.tablename -> add space
						if len(words) == 2 || (len(words) == 2 && strings.Contains(words[1], ".")) {
							shouldAddSpace = true
						}
					}
				}

				if shouldAddSpace {
					// Add a space and return - don't get completions yet
					// Let the user press tab again to see next completions
					currentInput += " "
					m.input.SetValue(currentInput)
					m.input.SetCursor(len(currentInput))
					return m, nil
				}
			}
		}
	}

	// Get completions for current input (use fullInput which includes multi-line buffer if applicable)
	logger.DebugfToFile("Completion", "Tab pressed: currentInput='%s', fullInput='%s'", currentInput, fullInput)
	m.completions = m.completionEngine.Complete(fullInput)
	logger.DebugfToFile("Completion", "Got %d completions: %v", len(m.completions), m.completions)

	if len(m.completions) == 0 { //nolint:gocritic // more readable as if
		// No completions available
		m.showCompletions = false
		m.completionIndex = -1
		m.completionScrollOffset = 0
	} else if len(m.completions) == 1 && completion.IsHint(m.completions[0]) {
		// The one thing to say is what to type. Show it: there is nothing to
		// apply, and applying it would take it straight back down again.
		m.showCompletions = true
		m.completingPath = false
		m.completionIndex = 0
		m.completionScrollOffset = 0
	} else if len(m.completions) == 1 {
		// Single completion - apply it immediately, the same way Enter and
		// Space apply one that was picked from a list.
		newValue := applyCompletion(currentInput, m.completions[0])

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
		m.completingPath = false     // these are CQL words, not filenames
		m.completionIndex = 0        // Start with first item selected
		m.completionScrollOffset = 0 // Reset scroll position
	}

	return m, nil
}

// quotesClosed reports whether every quote in the text has been closed.
func quotesClosed(input string) bool {
	return strings.Count(input, "'")%2 == 0 && strings.Count(input, "\"")%2 == 0
}
