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
