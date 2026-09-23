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

// TestClipboardWritersAnswerTheReaders: every platform that can be read from
// can be written to.
//
// Not one route for one route - Windows has two - but every tool the reader
// table names has something in the writer table that reaches the same
// clipboard.
func TestClipboardWritersAnswerTheReaders(t *testing.T) {
	writers := clipboardWriters()
	require.NotEmpty(t, writers)
	assert.GreaterOrEqual(t, len(writers), len(clipboardReaders()),
		"a platform that can read the clipboard should be able to write it")

	switch runtime.GOOS {
	case "darwin":
		assert.Equal(t, [][]string{{"pbcopy"}}, writers)
	case "windows":
		assert.Equal(t, "powershell.exe", writers[0][0])
		assert.Equal(t, "clip.exe", writers[1][0])
	default:
		assert.Equal(t, "wl-copy", writers[0][0])
		assert.Equal(t, "xclip", writers[1][0])
	}
}

// TestThePowerShellWriterIsGivenTheTextOnStdin.
//
// "powershell.exe -NoProfile -Command Set-Clipboard" exits 0 and writes
// nothing. It runs the cmdlet with no -Value, and the process's stdin never
// reaches the PowerShell pipeline, so it clears the clipboard instead of
// setting it. Exit 0 is what makes it dangerous: setSystemClipboard takes it
// for success and stops trying, so the copy is lost silently and takes
// whatever was on the clipboard with it.
//
// That is what Windows Terminal showed - copying and pasting inside cqlai
// worked, because pasting there falls back to what cqlai itself last copied,
// and nothing outside the terminal ever saw the text.
//
// Verified against the real Windows clipboard: with "$input |" the text
// arrives, without it the clipboard comes back empty.
func TestThePowerShellWriterIsGivenTheTextOnStdin(t *testing.T) {
	found := false
	for _, writer := range clipboardWriters() {
		if writer[0] != "powershell.exe" {
			continue
		}
		found = true

		command := writer[len(writer)-1]
		assert.Contains(t, command, "Set-Clipboard")
		assert.Contains(t, command, "$input |",
			"Set-Clipboard on its own clears the clipboard and exits 0: %q", command)
	}

	if runtime.GOOS != "darwin" {
		assert.True(t, found, "Windows and WSL reach the clipboard through PowerShell")
	}
}

// TestTheWritersTakeTheirTextOnStdin: every one of them is handed the text the
// same way, which is what setSystemClipboard does to all of them.
func TestTheWritersTakeTheirTextOnStdin(t *testing.T) {
	for _, writer := range clipboardWriters() {
		for _, arg := range writer[1:] {
			assert.NotContains(t, arg, "%s",
				"%q looks like it wants the text as an argument; it is given stdin", writer)
		}
	}
}
