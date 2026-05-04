package textmode

import (
	"strings"
)

// Completer is an interface that matches CompletionEngine.Complete so it can be
// mocked in tests.
type Completer interface {
	Complete(input string) []string
}

// ReadlineCompleter wraps a Completer and implements readline.AutoCompleter.
// It concatenates a multi-line buffer prefix with the current line up to the
// cursor position, calls Complete, then returns suffix offsets as readline expects.
type ReadlineCompleter struct {
	engine Completer
	// bufPrefix is the accumulated multi-line buffer joined with newlines.
	// It is updated by the REPL after each continuation line.
	bufPrefix string
}

// NewReadlineCompleter creates a new ReadlineCompleter.
func NewReadlineCompleter(engine Completer) *ReadlineCompleter {
	return &ReadlineCompleter{engine: engine}
}

// SetBufferPrefix updates the in-progress multi-line buffer context.
func (rc *ReadlineCompleter) SetBufferPrefix(prefix string) {
	rc.bufPrefix = prefix
}

// Do implements readline.AutoCompleter.
// line is the current line content (all runes up to len(line)), pos is the cursor position.
// readline passes the full line, not just up to cursor, but pos tells us the meaningful part.
func (rc *ReadlineCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	// Build the full input seen by the completion engine:
	// accumulated buffer + current line up to cursor.
	currentUpToCursor := string(line[:pos])
	var fullInput string
	if rc.bufPrefix != "" {
		fullInput = rc.bufPrefix + "\n" + currentUpToCursor
	} else {
		fullInput = currentUpToCursor
	}

	suggestions := rc.engine.Complete(fullInput)
	if len(suggestions) == 0 {
		return nil, 0
	}

	// Compute the word being completed (the suffix of currentUpToCursor after the last space).
	wordToComplete := ""
	if idx := strings.LastIndex(currentUpToCursor, " "); idx >= 0 {
		wordToComplete = currentUpToCursor[idx+1:]
	} else {
		wordToComplete = currentUpToCursor
	}

	// readline's AutoCompleter contract:
	//   length:  number of runes to delete from the cursor backwards before inserting.
	//   newLine: each entry is the SUFFIX to insert after those `length` runes are deleted.
	//
	// So if the user typed "SE" and the suggestion is "SELECT", we must return
	// the suffix "LECT" (not the full "SELECT"), and length = 2.  Returning the
	// full token would cause readline to produce "SESELECT".
	//
	// We operate on []rune throughout to handle multi-byte characters safely,
	// even though CQL keywords and typical keyspace/table names are ASCII.
	wordRunes := []rune(wordToComplete)
	length = len(wordRunes)
	newLine = make([][]rune, 0, len(suggestions))
	for _, s := range suggestions {
		sugRunes := []rune(s)
		if len(sugRunes) >= len(wordRunes) &&
			strings.EqualFold(string(sugRunes[:len(wordRunes)]), wordToComplete) {
			// Return only the suffix beyond what the user already typed.
			newLine = append(newLine, sugRunes[len(wordRunes):])
		}
		// Suggestions that don't start with the typed prefix are next-word
		// tokens (e.g. the completion engine returns "FROM" after "SELECT *").
		// We skip them here; the engine should only return relevant suggestions.
	}

	return newLine, length
}
