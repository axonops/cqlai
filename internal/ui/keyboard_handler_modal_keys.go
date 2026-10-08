package ui

import (
	tea "charm.land/bubbletea/v2"
)

// A modal is drawn over everything else, so its keys go to it first.
//
// They did not. Each key went through the view's own handling before anything
// asked whether a modal was open, so the view behind it answered instead: in
// TRACE, or RESULTS with a table showing, Esc switched navigation mode and the
// "No AI provider is configured" message stayed where it was; in CHAT, keys
// went to the conversation. Every modal is handled here, before any view.

// handleModalKey is what a key does while a modal is open.
func (m *MainModel) handleModalKey(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		return m.dismissModal()
	case "enter":
		return m.answerModal(m.modal.Selected)
	case "tab", "right":
		m.modal.NextChoice()
	case "shift+tab", "left":
		m.modal.PrevChoice()
	}
	// Anything else is not for the view behind the modal.
	return m, nil
}

// dismissModal closes a modal without its answer. A command waiting to be
// confirmed is cancelled, and says so; a message, or the question about
// quitting, just goes, leaving what was being typed where it was.
func (m *MainModel) dismissModal() (*MainModel, tea.Cmd) {
	cancelled := m.modal.Type == ModalConfirmDangerous
	m.modal = Modal{Type: ModalNone}
	if !cancelled {
		return m, nil
	}

	m.input.Placeholder = "Enter CQL command..."
	m.input.Reset()
	m.fullHistoryContent += "\n" + m.styles.MutedText.Render("Command cancelled.")
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()
	return m, nil
}
