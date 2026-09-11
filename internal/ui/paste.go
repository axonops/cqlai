package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Pasting into whatever is being typed into.
//
// A terminal with bracketed paste on - which is every one worth using, and what
// Bubble Tea asks for - does not send a paste as key presses. It arrives whole,
// as one tea.PasteMsg, and cqlai read none of them: pasting did nothing at the
// prompt, in the forms, or in the preferences window. An API key is forty
// characters of noise and typing one out by hand is how it gets a character
// wrong.
//
// Where it goes follows the same order the keyboard does, for the same reason:
// what is on top of the screen is what you are typing into.

// handlePaste puts pasted text where the keys are going.
func (m *MainModel) handlePaste(msg tea.PasteMsg) (*MainModel, tea.Cmd) {
	if msg.Content == "" {
		return m, nil
	}

	// A confirmation dialog is a yes or no question. Pasting into the prompt
	// behind it would put the text somewhere that cannot be seen.
	if m.modal.Type != ModalNone {
		return m, nil
	}

	var cmd tea.Cmd
	switch {
	case m.preferences.active:
		field := m.preferences.current()
		if field == nil || field.spec.kind == prefYesNo {
			return m, nil
		}
		m.preferences.fields[m.preferences.focus].input, cmd = field.input.Update(msg)
		m.preferences.clearMatches() // what was listed no longer describes what is there

	case m.form.active:
		field := m.form.current()
		if field == nil || field.kind == fieldYesNo {
			return m, nil
		}
		field.input, cmd = field.input.Update(msg)
		m.form.clearMatches()

	case m.capture.active:
		if m.capture.step != captureEnterPath {
			return m, nil // the format is a list to pick from, not a box
		}
		m.capture.input, cmd = m.capture.input.Update(msg)
		m.clearMatches()

	case m.historySearchMode:
		// The search is a plain string rather than an input, and it is one line
		// of a box: what is worth pasting into it is a word to search for.
		m.historySearchQuery += firstLine(msg.Content)
		return m.refreshHistorySearch()

	case m.aiConversationActive:
		m.aiConversationInput, cmd = m.aiConversationInput.Update(msg)

	default:
		m.input, cmd = m.input.Update(msg)

		// The completion list was worked out from what was typed before.
		if m.showCompletions {
			m.showCompletions = false
			m.completionIndex = -1
		}
	}
	return m, cmd
}

// firstLine is the pasted text up to its first line break.
//
// The inputs sanitize what they are given; the history query is a bare string
// and would take the newlines with it, drawing the box open at the seams.
func firstLine(text string) string {
	if i := strings.IndexAny(text, "\r\n"); i >= 0 {
		return text[:i]
	}
	return text
}
