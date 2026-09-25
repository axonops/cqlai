package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The quit shortcut.
//
// Ctrl+Q on all three platforms, macOS included: Command+Q is what a Mac user
// reaches for and it never arrives, because the terminal takes it and quits
// itself. A shortcut cqlai cannot receive is not one it can offer.

// TestCtrlQAsksRatherThanLeaving.
//
// It is written beside FILE > QUIT, so it has to do what that entry does. A
// key that quit where the menu asked would make the menu a liar.
func TestCtrlQAsksRatherThanLeaving(t *testing.T) {
	m := helpModel()
	require.Equal(t, ModalNone, m.modal.Type)

	m, cmd := m.handleKeyboardInput(keyPress("ctrl+q"))

	assert.NotEqual(t, ModalNone, m.modal.Type, "Ctrl+Q should put the question up")
	assert.Nil(t, cmd, "and not leave on its own")
}

// TestCtrlQAndTheMenuEntryDoTheSameThing.
func TestCtrlQAndTheMenuEntryDoTheSameThing(t *testing.T) {
	byKey := helpModel()
	byKey, _ = byKey.handleKeyboardInput(keyPress("ctrl+q"))

	byMenu := helpModel()
	byMenu, _ = byMenu.openFileMenu(0)
	byMenu.fileMenu.selected = indexOfQuit(t)
	byMenu, _ = byMenu.chooseFileMenuItem()

	assert.Equal(t, byMenu.modal.Type, byKey.modal.Type)
	assert.False(t, byMenu.fileMenu.active, "picking it closes the menu")
}

// TestTheMenuShowsTheKeyThatDoesTheSameThing: nobody opens the README to find
// out how to leave.
func TestTheMenuShowsTheKeyThatDoesTheSameThing(t *testing.T) {
	m := helpModel()
	m.windowWidth, m.windowHeight = 120, 40
	m, _ = m.openFileMenu(0)

	layer, ok := m.viewFileMenu(120, 40)
	require.True(t, ok)

	drawn := stripAnsi(layer.Content)
	assert.Contains(t, drawn, quitKeyLabel)

	// On the QUIT row, not floating on one of its own.
	for _, line := range strings.Split(drawn, "\n") {
		if strings.Contains(line, "QUIT") {
			assert.Contains(t, line, quitKeyLabel, "the key belongs on the entry it works for")
			return
		}
	}
	t.Fatal("no QUIT row")
}

// TestTheKeyTheMenuShowsIsTheKeyThatIsBound: the label is only worth drawing
// if pressing what it says does something.
func TestTheKeyTheMenuShowsIsTheKeyThatIsBound(t *testing.T) {
	key := strings.ToLower(quitKeyLabel)

	assert.True(t, isShellKey(key), "%s should reach the shell from every view", key)

	m := helpModel()
	m, _ = m.handleKeyboardInput(keyPress(key))
	assert.NotEqual(t, ModalNone, m.modal.Type, "pressing %s should do what the menu says", quitKeyLabel)
}

// TestCtrlQLeavesFromInsideAConversation, which is a view that used to swallow
// every key it did not recognise.
func TestCtrlQLeavesFromInsideAConversation(t *testing.T) {
	m := chatModel(t)

	m, _ = m.handleKeyboardInput(keyPress("ctrl+q"))

	assert.NotEqual(t, ModalNone, m.modal.Type, "Ctrl+Q should work in CHAT too")
	assert.Empty(t, m.aiConversationInput.Value(), "and not be typed into the message")
}

// TestTheQuitKeyIsNotClaimedAsACommandKey: Command+Q never reaches a terminal
// application, so the docs must not promise it.
func TestTheQuitKeyIsNotClaimedAsACommandKey(t *testing.T) {
	assert.Equal(t, "Ctrl+Q", quitKeyLabel)
	assert.NotContains(t, quitKeyLabel, "⌘")
}

// indexOfQuit is where QUIT sits in the menu.
func indexOfQuit(t *testing.T) int {
	t.Helper()

	for i, item := range fileMenuItems() {
		if item.quits {
			return i
		}
	}
	t.Fatal("the menu has no QUIT")
	return -1
}
