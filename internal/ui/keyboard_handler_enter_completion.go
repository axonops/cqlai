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

	// Debug logging (temporary)
	// fmt.Fprintf(os.Stderr, "DEBUG: currentInput='%s', selectedCompletion='%s'\n", currentInput, selectedCompletion)
	// fmt.Fprintf(os.Stderr, "DEBUG: ends with quote? %v\n", strings.HasSuffix(currentInput, "'") || strings.HasSuffix(currentInput, "\""))

	// Special case: if input ends with a dot (keyspace.), just append the table name
	if strings.HasSuffix(currentInput, ".") { //nolint:gocritic // more readable as if
		newValue = currentInput + selectedCompletion
	} else if strings.HasSuffix(currentInput, "=") {
		// For option assignments (FORMAT=, COMPRESSION=), just append the value
		newValue = currentInput + "'" + selectedCompletion + "'"
	} else if strings.HasSuffix(currentInput, "'") || strings.HasSuffix(currentInput, "\"") {
		// Input ends with a quote (complete quoted string), append with space
		newValue = currentInput + " " + selectedCompletion
	} else if strings.HasSuffix(currentInput, " ") {
		// Just append the completion
		newValue = currentInput + selectedCompletion
	} else {
		// Check if we have a partial word to replace
		lastSpace := strings.LastIndex(currentInput, " ")
		if lastSpace >= 0 {
			// Check if the last word is a complete token that shouldn't be replaced
			lastWord := currentInput[lastSpace+1:]
			upperLastWord := strings.ToUpper(lastWord)

			// Check if this is a complete parameter=value assignment
			isCompleteAssignment := false
			if strings.Contains(lastWord, "=") {
				// Check various complete assignment patterns
				// Use case-insensitive check for boolean values
				lowerLastWord := strings.ToLower(lastWord)
				if strings.HasSuffix(lastWord, "'") ||  // String value: FORMAT='parquet'
				   strings.HasSuffix(upperLastWord, "TRUE") ||  // Boolean: HEADER=TRUE
				   strings.HasSuffix(upperLastWord, "FALSE") ||  // Boolean: HEADER=FALSE
				   strings.HasSuffix(lowerLastWord, "true") ||  // Boolean: header=true (lowercase)
				   strings.HasSuffix(lowerLastWord, "false") ||  // Boolean: header=false (lowercase)
				   (len(lastWord) > 0 && lastWord[len(lastWord)-1] >= '0' && lastWord[len(lastWord)-1] <= '9') {  // Number: PAGESIZE=1000
					isCompleteAssignment = true
				}
			}

			// Check if last word is a quoted string (file path or value)
			isQuotedString := (strings.HasPrefix(lastWord, "'") && strings.HasSuffix(lastWord, "'")) ||
				(strings.HasPrefix(lastWord, "\"") && strings.HasSuffix(lastWord, "\""))

			// Determine how to apply the completion based on the last word
			switch {
			case upperLastWord == "TO" || upperLastWord == "FROM":
				// Don't replace TO/FROM, just append the file path with space
				newValue = currentInput + " " + selectedCompletion
			case isQuotedString:
				// Last word is a complete quoted string, don't replace, append with space
				newValue = currentInput + " " + selectedCompletion
			case strings.HasSuffix(lastWord, "="):
				// Don't replace, just append the value with quotes
				newValue = currentInput + "'" + selectedCompletion + "'"
			case isCompleteAssignment:
				// This is a complete assignment, don't replace, just append with space
				newValue = currentInput + " " + selectedCompletion
			case strings.Contains(lastWord, "."):
				// Check if this is completing a table name or completing after table name
				// If the completion starts with "(" it's column list for INSERT, not a table name
				if strings.HasPrefix(selectedCompletion, "(") {
					// Append column list after table name
					newValue = currentInput + " " + selectedCompletion
				} else {
					// For keyspace.table patterns, replace the part after the dot
					// The completion engine returns just the table name
					dotIndex := strings.LastIndex(currentInput, ".")
					newValue = currentInput[:dotIndex+1] + selectedCompletion
				}
			case lastWord == "*" || strings.HasSuffix(lastWord, ")"):
				// Don't replace, just append
				newValue = currentInput + " " + selectedCompletion
			default:
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