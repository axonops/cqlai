package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

// A modal's keys are its own, whatever view it is over.

// modalViews are the views a modal can open over, set up the way that made
// each one take the keys first: a trace or a table to navigate, a
// conversation to type into.
func modalViews() map[string]func(*MainModel) {
	return map[string]func(*MainModel){
		"trace with a trace":   func(m *MainModel) { m.viewMode = "trace"; m.hasTrace = true },
		"results with a table": func(m *MainModel) { m.viewMode = "table"; m.hasTable = true },
		"chat":                 func(m *MainModel) { m.viewMode = "ai"; m.aiConversationActive = true },
		"console":              func(m *MainModel) { m.viewMode = "history" },
	}
}

func pressCode(m *MainModel, code rune) *MainModel {
	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: code})
	return m
}

// TestEscClosesAMessageOverAnyView: the "No AI provider is configured"
// message over a trace stayed open, and Esc switched navigation mode behind it.
func TestEscClosesAMessageOverAnyView(t *testing.T) {
	for name, set := range modalViews() {
		m := helpModel()
		set(m)
		m.input.SetValue("SELECT * FROM shop.orders")
		m.modal = NewMessageModal("No AI provider is configured", "Set one up in PREFERENCES.")

		m = pressCode(m, tea.KeyEscape)

		assert.Equal(t, ModalNone, m.modal.Type, name)
		assert.False(t, m.navigationMode, "%s: Esc closed the message and did nothing else", name)
		assert.Equal(t, "SELECT * FROM shop.orders", m.input.Value(), "%s: what was typed is still there", name)
		assert.NotContains(t, m.fullHistoryContent, "Command cancelled", name)
	}
}

// TestEscCancelsAWaitingCommandOverAnyView, and says so.
func TestEscCancelsAWaitingCommandOverAnyView(t *testing.T) {
	for name, set := range modalViews() {
		m := helpModel()
		set(m)
		m.modal = NewConfirmationModal("DROP TABLE shop.orders")

		m = pressCode(m, tea.KeyEscape)

		assert.Equal(t, ModalNone, m.modal.Type, name)
		assert.Contains(t, m.fullHistoryContent, "Command cancelled", name)
	}
}

// TestTheQuitQuestionIsAnsweredOverAnyView: Esc keeps cqlai open, the arrows
// move between the buttons, and Enter answers.
func TestTheQuitQuestionIsAnsweredOverAnyView(t *testing.T) {
	for name, set := range modalViews() {
		m := helpModel()
		set(m)
		m.modal = NewQuitModal()

		m = pressCode(m, tea.KeyEscape)
		assert.Equal(t, ModalNone, m.modal.Type, "%s: Esc closes it", name)

		m.modal = NewQuitModal()
		start := m.modal.Selected
		m = pressCode(m, tea.KeyRight)
		assert.NotEqual(t, start, m.modal.Selected, "%s: right moves", name)
		m = pressCode(m, tea.KeyLeft)
		assert.Equal(t, start, m.modal.Selected, "%s: left moves back", name)
		m = pressCode(m, tea.KeyTab)
		assert.NotEqual(t, start, m.modal.Selected, "%s: Tab moves", name)

		_, cmd := m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEnter})
		if assert.NotNil(t, cmd, "%s: Enter on Quit answers it", name) {
			assert.IsType(t, tea.QuitMsg{}, cmd(), name)
		}
	}
}

// TestOtherKeysDoNotReachTheViewBehind: typing while a modal is open goes
// nowhere, not into the prompt or the conversation behind it.
func TestOtherKeysDoNotReachTheViewBehind(t *testing.T) {
	for name, set := range modalViews() {
		m := helpModel()
		set(m)
		m.modal = NewMessageModal("Nothing to analyse", "Run a query with TRACING ON first.")

		m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: 'x', Text: "x"})

		assert.Equal(t, ModalMessage, m.modal.Type, name)
		assert.Empty(t, m.input.Value(), name)
	}
}

// TestEscInTheConsoleDoesNotAskToQuit: with nothing open, Esc leaves the
// prompt and what is typed in it alone.
func TestEscInTheConsoleDoesNotAskToQuit(t *testing.T) {
	m := editModel("SELECT * FROM shop.orders", 6)
	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.False(t, m.confirmExit)
	assert.Equal(t, "SELECT * FROM shop.orders", m.input.Value())
}

// TestEscTakesBackAnExit: Esc after the first Ctrl+C stays in the shell.
func TestEscTakesBackAnExit(t *testing.T) {
	m := editModel("", 0)
	m.confirmExit = true
	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.False(t, m.confirmExit)
}
