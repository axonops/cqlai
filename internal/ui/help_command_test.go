package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHELPOpensTheHelpWindow, whatever OUTPUT is set to: printed as a result,
// the help came out as a page of JSON with OUTPUT JSON.
func TestHELPOpensTheHelpWindow(t *testing.T) {
	for _, typed := range []string{"HELP", "help", "HELP;", " Help ", "HELP CONSISTENCY"} {
		m := helpModel()
		m.hasTable = false
		m.input.SetValue(typed)

		_, _, handled := m.handleSpecialCommands(typed)

		assert.True(t, handled, "%q", typed)
		assert.True(t, m.help.active, "%q opens the window", typed)
		assert.Empty(t, m.input.Value(), "%q is taken from the prompt", typed)
		assert.False(t, m.hasTable, "%q is not a result", typed)
	}

	m := helpModel()
	_, _, handled := m.handleSpecialCommands("HELPER")
	assert.False(t, handled, "another word that starts the same")
}

// TestHELPAndACommandOpensItsTopic: HELP INSERT is how INSERT is written.
func TestHELPAndACommandOpensItsTopic(t *testing.T) {
	m := helpModel()
	m.handleSpecialCommands("HELP create table;")
	assert.Equal(t, "CREATE TABLE", m.help.topic)

	layer, ok := m.viewHelp(m.windowWidth, m.windowHeight)
	assert.True(t, ok)
	assert.Contains(t, stripAnsi(layer.Content), "HELP CREATE TABLE")
	assert.Contains(t, stripAnsi(layer.Content), "Syntax:")

	m = helpModel()
	m.handleSpecialCommands("HELP")
	assert.Empty(t, m.help.topic, "HELP on its own is everything")
}
