package ui

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// init keeps the tests off the machine's real clipboard. Copying writes to it
// on purpose, and a test run should not take over what the person at the
// keyboard had copied.
func init() {
	writeToMachineClipboard = func(string) {}
}

// TestCopyAlsoGoesToTheMachine is issue #210: on macOS the copy only went out
// as OSC 52, Terminal.app ignores that, and nothing reached the clipboard. The
// copy has to take both routes.
func TestCopyAlsoGoesToTheMachine(t *testing.T) {
	original := writeToMachineClipboard
	t.Cleanup(func() { writeToMachineClipboard = original })

	var got string
	var calls int
	writeToMachineClipboard = func(text string) {
		got = text
		calls++
	}

	m := selectionModel(5, "first line", "second line", "third line")
	cmd := drag(m, 0, 1, 5, 2)
	require.NotNil(t, cmd)

	// Runs both halves of the batch: the OSC 52 write and this one.
	msgs := runClipboardCmd(cmd)
	require.NotEmpty(t, msgs)

	assert.Equal(t, 1, calls, "the machine's clipboard should have been written once")
	assert.Equal(t, "first line\nsecon", got)
	assert.Equal(t, got, clipboardText(t, cmd),
		"both routes should carry the same text")
}

// TestClickWithoutDraggingLeavesTheMachineAlone: a stray click must not wipe
// the clipboard through the new route either.
func TestClickWithoutDraggingLeavesTheMachineAlone(t *testing.T) {
	original := writeToMachineClipboard
	t.Cleanup(func() { writeToMachineClipboard = original })

	calls := 0
	writeToMachineClipboard = func(string) { calls++ }

	m := selectionModel(5, "first line")
	m.beginSelection(4, 1)
	_, cmd := m.endSelection()

	require.Nil(t, cmd)
	assert.Zero(t, calls)
}

// TestClipboardWritersMirrorTheReaders: every platform that can be read from
// can be written to, and the two tables stay in step.
func TestClipboardWritersMirrorTheReaders(t *testing.T) {
	writers := clipboardWriters()
	require.NotEmpty(t, writers)
	assert.Len(t, writers, len(clipboardReaders()),
		"a platform that can read the clipboard should be able to write it")

	switch runtime.GOOS {
	case "darwin":
		assert.Equal(t, [][]string{{"pbcopy"}}, writers)
	case "windows":
		assert.Equal(t, "powershell.exe", writers[0][0])
	default:
		assert.Equal(t, "wl-copy", writers[0][0])
		assert.Equal(t, "xclip", writers[1][0])
	}
}
