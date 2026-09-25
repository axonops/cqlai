package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The keys that belong to the shell rather than to whatever has focus.
//
// CHAT hands every key it does not recognise to its input field. It remembered
// F2 to F6 and nothing else, so from inside a conversation Alt+F opened no FILE
// menu, Alt+H and F1 opened no help, and the letters went into the message
// instead. The list was written twice - the main handler's switch, and a case
// in the CHAT handler - and only one was kept up.

// chatModel is the CHAT view with a conversation open.
func chatModel(t *testing.T) *MainModel {
	t.Helper()

	m := helpModel()
	m.windowWidth, m.windowHeight = 100, 30
	m.viewMode = "ai"
	m.aiConversationActive = true
	m.aiConfig = configuredAI()
	m.aiConversationInput = textinput.New()
	m.aiConversationInput.SetWidth(80)
	m.aiConversationInput.Focus()
	return m
}

// TestTheShellsKeysAreNotTypedIntoTheConversation.
func TestTheShellsKeysAreNotTypedIntoTheConversation(t *testing.T) {
	for _, key := range shellKeys() {
		m := chatModel(t)

		result, _ := m.handleAIConversationInput(keyPress(key))

		assert.Nil(t, result, "%s belongs to the shell and must reach the main handler", key)
		assert.Empty(t, m.aiConversationInput.Value(), "%s should not be typed into the message", key)
	}
}

// TestAltFOpensTheFileMenuFromInsideAConversation, which is the report: a view
// that eats the FILE menu is one you cannot get out of except by the keys it
// happens to remember.
func TestAltFOpensTheFileMenuFromInsideAConversation(t *testing.T) {
	m := chatModel(t)
	require.False(t, m.fileMenu.active)

	m, _ = m.handleKeyboardInput(keyPress("alt+f"))

	assert.True(t, m.fileMenu.active, "Alt+F should open the FILE menu in CHAT")
}

// TestF1AndAltHOpenTheHelpFromInsideAConversation.
func TestF1AndAltHOpenTheHelpFromInsideAConversation(t *testing.T) {
	for _, key := range []string{"f1", "alt+h"} {
		m := chatModel(t)
		require.False(t, m.help.active)

		m, _ = m.handleKeyboardInput(keyPress(key))

		assert.True(t, m.help.active, "%s should open the help in CHAT", key)
	}
}

// TestEveryViewIsReachableFromInsideAConversation: the tab line says which key
// reaches which view, and it has to be telling the truth from in here too.
func TestEveryViewIsReachableFromInsideAConversation(t *testing.T) {
	for _, tab := range append(append([]modeTab{}, modeTabs...), resultTabs...) {
		key := strings.ToLower(tab.key)

		m := chatModel(t)
		result, _ := m.handleAIConversationInput(keyPress(key))

		assert.Nil(t, result, "%s reaches %s and must not be swallowed", tab.key, tab.label)
	}
}

// TestWhatIsTypedStillReachesTheMessage: the point of the view is the
// conversation, and the fix must not cost it any keys.
func TestWhatIsTypedStillReachesTheMessage(t *testing.T) {
	m := chatModel(t)

	for _, key := range []string{"h", "f", "a"} {
		m, _ = m.handleAIConversationInput(keyPress(key))
	}

	assert.Equal(t, "hfa", m.aiConversationInput.Value(),
		"the letters behind the shell's Alt and F keys are still letters")
}

// TestNoShellKeyEndsUpAsText, all the way through the real handler rather than
// only the one that passes it up.
func TestNoShellKeyEndsUpAsText(t *testing.T) {
	for _, key := range shellKeys() {
		m := chatModel(t)

		m, _ = m.handleKeyboardInput(keyPress(key))

		assert.Empty(t, m.aiConversationInput.Value(),
			"%s should have been answered, not typed", key)
	}
}

// keyPress is a key as the handlers receive it.
func keyPress(key string) tea.KeyPressMsg {
	switch {
	case strings.HasPrefix(key, "alt+"):
		return tea.KeyPressMsg{Code: rune(key[len("alt+")]), Mod: tea.ModAlt}
	case key == "f1":
		return tea.KeyPressMsg{Code: tea.KeyF1}
	case key == "f2":
		return tea.KeyPressMsg{Code: tea.KeyF2}
	case key == "f3":
		return tea.KeyPressMsg{Code: tea.KeyF3}
	case key == "f4":
		return tea.KeyPressMsg{Code: tea.KeyF4}
	case key == "f5":
		return tea.KeyPressMsg{Code: tea.KeyF5}
	case key == "f6":
		return tea.KeyPressMsg{Code: tea.KeyF6}
	}
	// A printable key carries its text as well as its code; the input field
	// reads the text.
	return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
}
