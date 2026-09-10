package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/ui/completion"
)

// pathHint is shown where a filename is expected but none has been typed.
//
// It is a note rather than a candidate: there is nothing to complete against
// yet, and listing the whole working directory is noise. The wording matches
// the usage the commands print - CAPTURE [JSON|CSV|PARQUET] 'filename' - so the
// two say the same thing.
const pathHint = "'filename'"

// Completing a file path at the prompt.
//
// The engine cannot do this itself: it would have to reach back into the UI for
// the filesystem, and the UI already imports it. So the Tab handler asks the
// engine only where a path is expected, and does the completing here.

// completePathAtPrompt completes a file path if the cursor is somewhere one is
// expected. It reports whether it handled the keypress.
func (m *MainModel) completePathAtPrompt(input string) (*MainModel, tea.Cmd, bool) {
	prefix, quote, ok := completion.PathPrefix(input)
	if !ok {
		return m, nil, false
	}

	// With nothing typed there is nothing to complete against, and listing the
	// whole working directory is noise - the binary and the log file are not
	// what you are looking for. Say what is expected instead.
	//
	// Once there is a quote you have committed to typing a path, so the
	// listing is what you want and the hint would be in the way.
	if strings.TrimSpace(prefix) == "" && quote == "" {
		m.showCompletions = true
		m.completions = []string{pathHint}
		m.completionIndex = 0
		m.completionScrollOffset = 0
		m.completingPath = true
		return m, nil, true
	}

	got := completePath(prefix)

	m.input.SetValue(completion.ReplacePath(input, prefix, quote, got.Completed))
	m.input.CursorEnd()

	// Show the candidates only when filling in was not enough to pick one,
	// which is what the shell does and what the capture window does.
	m.showCompletions = false
	m.completions = nil
	m.completionIndex = -1
	m.completionScrollOffset = 0

	m.completingPath = false
	if got.worthListing() {
		m.completions = got.Matches
		m.completionIndex = 0
		m.showCompletions = true
		m.completingPath = true
	}
	return m, nil, true
}

// applyPathCompletion puts a chosen filename into the path being typed.
//
// The candidates are bare names - "report.csv", "exports/" - so the directory
// in front of them has to be kept. Applying one the way a CQL word is applied
// replaces the last word, which for /tmp/rep means losing the /tmp.
func (m *MainModel) applyPathCompletion(name string) {
	// Choosing the hint starts the quoted filename off, so it is worth
	// selecting rather than being a dead row in the list. The next Tab
	// completes inside the quote.
	if name == pathHint {
		started := m.input.Value()
		if !strings.HasSuffix(started, " ") {
			started += " "
		}

		m.input.SetValue(started + "''")
		m.input.SetCursor(len(started) + 1) // between the quotes
		m.clearCompletions()
		return
	}

	input := m.input.Value()

	prefix, quote, ok := completion.PathPrefix(input)
	if !ok {
		return
	}

	// The candidates are names inside a directory, so the directory typed so
	// far stays in front of the one picked - except the way up, which replaces
	// it.
	dir, _ := splitPath(prefix)
	chosen := dir + name
	if name == parentEntry {
		chosen = parentDir(dir)
	}

	m.input.SetValue(completion.ReplacePath(input, prefix, quote, chosen))
	m.input.CursorEnd()
	m.clearCompletions()
}

// clearCompletions takes the list of candidates down.
func (m *MainModel) clearCompletions() {
	m.showCompletions = false
	m.completions = nil
	m.completionIndex = -1
	m.completionScrollOffset = 0
	m.completingPath = false
}
