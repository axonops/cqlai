package ui

import (
	"strings"
	"time"

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
	return m.pasteText(msg.Content)
}

// pasteText puts text where the keys are going, whether it arrived from the
// terminal's own paste or from a right click asking for the clipboard.
func (m *MainModel) pasteText(text string) (*MainModel, tea.Cmd) {
	if text == "" {
		return m, nil
	}
	msg := tea.PasteMsg{Content: text}

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

// Right-click paste.
//
// The terminal's own right-click menu is gone as soon as cqlai takes the mouse,
// and with it the paste everyone reaches for. Asking the terminal for the
// clipboard is possible - OSC 52 has a read as well as a write - but most
// terminals refuse it, because a program that can read your clipboard can read
// what you copied out of your password manager.
//
// So a right click does both: it asks, and if nothing comes back it pastes what
// cqlai itself last copied. Text selected here always pastes; text copied from
// another application pastes wherever the terminal allows the read.

// pasteAnswerWait is how long to wait before giving up on both the terminal and
// the machine and pasting what cqlai itself copied.
//
// Longer than a local clipboard command takes to answer - reading it through
// WSL is a few hundred milliseconds - because the machine's clipboard is what
// was asked for, and beating it to the punch with an older copy of our own
// would paste the wrong thing.
const pasteAnswerWait = 700 * time.Millisecond

// pasteFallbackMsg says that a right click went unanswered.
type pasteFallbackMsg struct{ request int }

// requestPaste asks the terminal for the clipboard, and arranges to fall back
// to cqlai's own last copy if it says nothing.
func (m *MainModel) requestPaste() (*MainModel, tea.Cmd) {
	m.pasteRequest++
	request := m.pasteRequest

	return m, tea.Batch(
		// The terminal, which is the only route that works through ssh and
		// tmux. ReadClipboard is the message rather than a command returning
		// one, so it is already the shape Batch wants.
		tea.ReadClipboard,

		// And the machine, for the terminals that refuse - which is most of
		// them. Whichever answers first is the one pasted.
		readSystemClipboard(request),

		tea.Tick(pasteAnswerWait, func(time.Time) tea.Msg {
			return pasteFallbackMsg{request: request}
		}),
	)
}

// handleClipboard pastes what the terminal sent back.
//
// Only when something asked for it: a clipboard message arriving on its own is
// not a paste anyone requested, and pasting into the prompt on its own is how a
// stray sequence ends up as a query.
func (m *MainModel) handleClipboard(msg tea.ClipboardMsg) (*MainModel, tea.Cmd) {
	if m.pasteRequest == 0 {
		return m, nil
	}

	m.pasteRequest = 0
	return m.pasteText(msg.Content)
}

// handleSystemClipboard pastes what the machine says its clipboard holds.
//
// An empty answer leaves the request open rather than closing it: it means this
// machine has nothing that can read a clipboard, and the terminal may yet
// answer - and failing that, what cqlai copied is still better than nothing.
func (m *MainModel) handleSystemClipboard(msg systemClipboardMsg) (*MainModel, tea.Cmd) {
	if msg.request != m.pasteRequest || msg.text == "" {
		return m, nil
	}

	m.pasteRequest = 0
	return m.pasteText(msg.text)
}

// handlePasteFallback pastes what cqlai last copied, when neither the terminal
// nor the machine answered the right click that asked for it.
func (m *MainModel) handlePasteFallback(msg pasteFallbackMsg) (*MainModel, tea.Cmd) {
	// A later right click has been and gone, or the terminal answered this one:
	// either way this is not the request still waiting.
	if msg.request != m.pasteRequest {
		return m, nil
	}

	m.pasteRequest = 0
	return m.pasteText(m.lastCopied)
}
