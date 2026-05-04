package textmode

import (
	"strings"

	"github.com/axonops/cqlai/internal/batch"
)

// InputBuffer accumulates lines typed by the user until a complete CQL
// statement (or meta-command) has been received.
type InputBuffer struct {
	lines []string
}

// Add appends a line to the buffer.
func (b *InputBuffer) Add(line string) {
	b.lines = append(b.lines, line)
}

// Text returns the accumulated lines joined by newlines.
func (b *InputBuffer) Text() string {
	return strings.Join(b.lines, "\n")
}

// Reset clears the buffer.
func (b *InputBuffer) Reset() {
	b.lines = nil
}

// IsEmpty returns true when no lines have been accumulated.
func (b *InputBuffer) IsEmpty() bool {
	return len(b.lines) == 0
}

// IsComplete returns true when the accumulated text represents at least one
// complete CQL statement. A statement is complete when:
//   - SplitStatements reports Incomplete == false (no unclosed batch or string), AND
//   - The last token of the last non-trivial statement is a semicolon (TokenEndtoken).
//
// This matches cqlsh behaviour where Enter only dispatches once the user types ";".
func (b *InputBuffer) IsComplete() bool {
	text := b.Text()
	if strings.TrimSpace(text) == "" {
		return false
	}
	result, err := batch.SplitStatements(text)
	if err != nil {
		// On lex/parse error, treat as complete so the executor can surface the error.
		return true
	}
	if result.Incomplete {
		return false
	}
	// Check that at least one statement ends with a semicolon token.
	// Skip junk tokens (whitespace, comments) when looking for the last real token.
	for _, stmt := range result.Statements {
		for i := len(stmt) - 1; i >= 0; i-- {
			tok := stmt[i]
			if tok.Type == batch.TokenJunk || tok.Type == batch.TokenEndline {
				continue
			}
			if tok.Type == batch.TokenEndtoken {
				return true
			}
			break
		}
	}
	return false
}

// IsMeta returns true when the first non-empty line looks like a meta-command
// (starts with a recognised keyword that does not require a semicolon).
// These are dispatched immediately without waiting for a semicolon.
func IsMeta(line string) bool {
	upper := strings.ToUpper(strings.TrimSpace(line))
	upper = strings.TrimSuffix(upper, ";")
	metaPrefixes := []string{
		"HELP", "QUIT", "EXIT",
		"DESCRIBE", "DESC",
		"CONSISTENCY", "OUTPUT",
		"PAGING", "AUTOFETCH",
		"TRACING", "SOURCE",
		"COPY", "SHOW", "EXPAND",
		"CAPTURE", "SAVE",
	}
	for _, prefix := range metaPrefixes {
		if upper == prefix || strings.HasPrefix(upper, prefix+" ") {
			return true
		}
	}
	return false
}
