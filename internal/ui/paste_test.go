package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
)

const pastedKey = "sk-ant-api03-0123456789abcdefghijklmnopqrstuvwxyz"

func paste(m *MainModel, text string) *MainModel {
	updated, _ := m.handlePaste(tea.PasteMsg{Content: text})
	return updated
}

// TestPastingIntoASetting: an API key is forty characters of noise, which is
// what pasting is for.
func TestPastingIntoASetting(t *testing.T) {
	m := prefModel(t, &config.Config{})
	i := prefIndex(t, m, "AI.APIKey")
	m.preferences.focusField(i)

	m = paste(m, pastedKey)
	assert.Equal(t, pastedKey, m.preferences.fields[i].value())

	// It goes in at the cursor, like typing does.
	m.preferences.focusField(prefIndex(t, m, "Host"))
	m = press(m, "a")
	m = paste(m, "bc")
	m = press(m, "d")
	assert.Equal(t, "abcd", m.preferences.fields[prefIndex(t, m, "Host")].value())
}

// TestPastingClosesTheCandidateList, which was worked out from what was there
// before the paste.
func TestPastingClosesTheCandidateList(t *testing.T) {
	m := prefModel(t, &config.Config{})
	m.preferences.focusField(prefIndex(t, m, "Consistency"))

	m = press(m, "tab")
	require.NotEmpty(t, m.preferences.matches)

	m = paste(m, "QUORUM")
	assert.Empty(t, m.preferences.matches)
	assert.Equal(t, "QUORUM", m.preferences.fields[prefIndex(t, m, "Consistency")].value())
}

// TestPastingIntoAYesNoSettingDoesNothing: there is nothing to type into one,
// and the pasted text would go somewhere it could not be seen.
func TestPastingIntoAYesNoSettingDoesNothing(t *testing.T) {
	m := prefModel(t, &config.Config{})
	i := prefIndex(t, m, "SSL.Enabled")
	m.preferences.focusField(i)

	m = paste(m, "yes")
	assert.Equal(t, "false", m.preferences.fields[i].value())
}

// TestPastingIntoAForm, which had the same gap.
func TestPastingIntoAForm(t *testing.T) {
	m := formModel(t, copyingTo)
	focus(t, m, "File")
	set(t, m, "File", "")

	m = paste(m, "/tmp/a rather long file name.parquet")
	assert.Equal(t, "/tmp/a rather long file name.parquet", m.form.field("File"))
}

// TestPastingIntoThePrompt: the prompt read no pastes either.
func TestPastingIntoThePrompt(t *testing.T) {
	m := helpModel()
	m.input.Focus() // as the prompt is for as long as cqlai is running

	m = paste(m, "SELECT * FROM system.local")
	assert.Equal(t, "SELECT * FROM system.local", m.input.Value())
}

// TestPastingIntoTheAutoSavePath.
func TestPastingIntoTheAutoSavePath(t *testing.T) {
	m := helpModel()
	m.windowWidth = 120
	m.windowHeight = 40
	m, _ = m.openCapturePanelCentred()
	require.True(t, m.capture.active)

	// The first step is a list of formats to pick from, so there is nothing to
	// paste into yet.
	require.Equal(t, captureChooseFormat, m.capture.step)
	m = paste(m, "/exports/")
	assert.Equal(t, captureChooseFormat, m.capture.step)

	m, _ = m.chooseCaptureFormat()
	require.Equal(t, captureEnterPath, m.capture.step)

	m.capture.input.SetValue("")
	m = paste(m, "/exports/nightly/")
	assert.Equal(t, "/exports/nightly/", m.capture.input.Value())
}

// TestPastingIntoTheHistorySearch narrows the list, the same as typing it.
func TestPastingIntoTheHistorySearch(t *testing.T) {
	m := helpModel()
	m.commandHistory = []string{"SELECT * FROM users", "DROP TABLE events"}
	m, _ = m.handleCtrlR()
	require.True(t, m.historySearchMode)

	// A paste with a line break in it is one line: the query is a single line
	// of a box rather than an input that would sanitize it.
	m = paste(m, "users\nDROP TABLE events")
	assert.Equal(t, "users", m.historySearchQuery)
	assert.Equal(t, []string{"SELECT * FROM users"}, m.historySearchResults)
}

// TestPastingWithAQuestionOnScreenDoesNothing: a confirmation is a yes or no,
// and the prompt behind it cannot be seen.
func TestPastingWithAQuestionOnScreenDoesNothing(t *testing.T) {
	m := helpModel()
	m.modal = NewConfirmationModal("DROP TABLE events")

	m = paste(m, "yes")
	assert.Empty(t, m.input.Value())
}

// TestRightClickPastesWhatTheTerminalSaysItHolds.
func TestRightClickPastesWhatTheTerminalSaysItHolds(t *testing.T) {
	m := helpModel()
	m.input.Focus()

	m, cmd := m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseRight, X: 10, Y: 5})
	require.NotNil(t, cmd, "a right click should have asked for the clipboard")
	assert.Equal(t, 1, m.pasteRequest)

	m, _ = m.handleClipboard(tea.ClipboardMsg{Content: "SELECT * FROM users"})
	assert.Equal(t, "SELECT * FROM users", m.input.Value())
	assert.Zero(t, m.pasteRequest, "the request should be answered")
}

// TestRightClickFallsBackToWhatCqlaiCopied.
//
// Most terminals refuse to say what their clipboard holds - a program that can
// read it can read what you copied out of a password manager - so a right click
// that goes unanswered pastes what was last copied here.
func TestRightClickFallsBackToWhatCqlaiCopied(t *testing.T) {
	m := selectionModel(5, "CREATE TABLE users (", "    id uuid PRIMARY KEY")
	m.input = newTestInput()
	m.input.Focus()
	drag(m, 0, 1, 20, 1)
	require.Equal(t, "CREATE TABLE users (", m.lastCopied)

	m, _ = m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseRight, X: 10, Y: 5})
	m, _ = m.handlePasteFallback(pasteFallbackMsg{request: m.pasteRequest})

	assert.Equal(t, "CREATE TABLE users (", m.input.Value())
}

// TestAnAnsweredRightClickDoesNotPasteTwice: the fallback is for a click the
// terminal ignored, and a late answer to one already served is not a paste.
func TestAnAnsweredRightClickDoesNotPasteTwice(t *testing.T) {
	m := helpModel()
	m.input.Focus()
	m.lastCopied = "from cqlai"

	m, _ = m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseRight, X: 10, Y: 5})
	request := m.pasteRequest

	m, _ = m.handleClipboard(tea.ClipboardMsg{Content: "from the terminal"})
	m, _ = m.handlePasteFallback(pasteFallbackMsg{request: request})
	assert.Equal(t, "from the terminal", m.input.Value())

	// And the other way round: the fallback pastes, the answer turns up late.
	m.input.SetValue("")
	m, _ = m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseRight, X: 10, Y: 5})
	m, _ = m.handlePasteFallback(pasteFallbackMsg{request: m.pasteRequest})
	m, _ = m.handleClipboard(tea.ClipboardMsg{Content: "from the terminal"})
	assert.Equal(t, "from cqlai", m.input.Value())
}

// TestAClipboardMessageNobodyAskedForIsIgnored: pasting into the prompt on its
// own is how a stray sequence ends up as a query.
func TestAClipboardMessageNobodyAskedForIsIgnored(t *testing.T) {
	m := helpModel()
	m.input.Focus()

	m, _ = m.handleClipboard(tea.ClipboardMsg{Content: "DROP TABLE users"})
	assert.Empty(t, m.input.Value())
}

// TestARightClickGoesWhereTheKeysGo, the same as any other paste.
func TestARightClickGoesWhereTheKeysGo(t *testing.T) {
	m := prefModel(t, &config.Config{})
	i := prefIndex(t, m, "AI.APIKey")
	m.preferences.focusField(i)

	m, _ = m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseRight, X: 10, Y: 5})
	m, _ = m.handleClipboard(tea.ClipboardMsg{Content: pastedKey})

	assert.Equal(t, pastedKey, m.preferences.fields[i].value())
}

// TestRightClickPastesWhatTheMachineHolds.
//
// The terminal is asked first, because OSC 52 is the only route that works
// through ssh and tmux, but most terminals refuse to answer. The machine's own
// clipboard is what "paste" means to everyone else, so it is asked too.
func TestRightClickPastesWhatTheMachineHolds(t *testing.T) {
	m := helpModel()
	m.input.Focus()
	m.lastCopied = "an older copy of our own"

	m, _ = m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseRight, X: 10, Y: 5})
	m, _ = m.handleSystemClipboard(systemClipboardMsg{request: m.pasteRequest, text: "copied somewhere else"})

	assert.Equal(t, "copied somewhere else", m.input.Value())
	assert.Zero(t, m.pasteRequest)
}

// TestAnEmptyAnswerLeavesTheRequestOpen: nothing here can read a clipboard, so
// the terminal may yet answer - and failing that, what cqlai copied is still
// better than nothing.
func TestAnEmptyAnswerLeavesTheRequestOpen(t *testing.T) {
	m := helpModel()
	m.input.Focus()
	m.lastCopied = "from cqlai"

	m, _ = m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseRight, X: 10, Y: 5})
	request := m.pasteRequest

	m, _ = m.handleSystemClipboard(systemClipboardMsg{request: request, text: ""})
	assert.Equal(t, request, m.pasteRequest, "an empty answer is not an answer")
	assert.Empty(t, m.input.Value())

	m, _ = m.handlePasteFallback(pasteFallbackMsg{request: request})
	assert.Equal(t, "from cqlai", m.input.Value())
}

// TestALateAnswerFromTheMachineIsIgnored, the same as a late one from the
// terminal: the click it belongs to has already been served.
func TestALateAnswerFromTheMachineIsIgnored(t *testing.T) {
	m := helpModel()
	m.input.Focus()

	m, _ = m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseRight, X: 10, Y: 5})
	m, _ = m.handleClipboard(tea.ClipboardMsg{Content: "from the terminal"})
	m, _ = m.handleSystemClipboard(systemClipboardMsg{request: 1, text: "from the machine"})

	assert.Equal(t, "from the terminal", m.input.Value())
}

// TestEveryPlatformHasSomewhereToLook. The commands are a fixed table, and a
// platform with nothing in it would silently never paste from outside.
func TestEveryPlatformHasSomewhereToLook(t *testing.T) {
	readers := clipboardReaders()
	require.NotEmpty(t, readers)

	for _, reader := range readers {
		assert.NotEmpty(t, reader, "a command with no name")
	}
}
