package ui

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheSaveWindowWritesTheFile is the reported bug: picking a format and a
// path in the Save window produced no file, no error and no message.
func TestTheSaveWindowWritesTheFile(t *testing.T) {
	for _, format := range []string{"CSV", "JSON", "PARQUET"} {
		t.Run(format, func(t *testing.T) {
			m := helpModel()
			m.windowHeight = 30
			m.lastTableData = [][]string{{"id", "name"}, {"1", "alice"}}
			m.columnTypes = []string{"int", "text"}

			path := filepath.Join(t.TempDir(), "out")

			m.openSavePanel()
			m.capture.format = slices.Index(m.capture.formats, format)
			require.GreaterOrEqual(t, m.capture.format, 0, "the window offers %s", format)
			m.chooseCaptureFormat()
			m.capture.input.SetValue(path)
			m.startCapture()

			assert.FileExists(t, path, "the window wrote nothing")
			assert.Contains(t, stripAnsiForTest(m.fullHistoryContent), "Successfully saved 1 rows",
				"and says so, rather than leaving you to go and look")
		})
	}
}

// TestBothRoutesToASaveDoTheSameThing: the window builds the command typing it
// would, and both carry it out through the same call, so neither can start
// writing a different file from the other.
func TestBothRoutesToASaveDoTheSameThing(t *testing.T) {
	data := [][]string{{"id", "name"}, {"1", "alice"}, {"2", "bob"}}
	types := []string{"int", "text"}
	dir := t.TempDir()

	typed := helpModel()
	typed.lastTableData, typed.columnTypes = data, types
	typedPath := filepath.Join(dir, "typed.csv")
	typed.runCommand("SAVE TO '" + typedPath + "' AS CSV")

	clicked := helpModel()
	clicked.windowHeight = 30
	clicked.lastTableData, clicked.columnTypes = data, types
	clickedPath := filepath.Join(dir, "clicked.csv")
	clicked.openSavePanel()
	clicked.capture.format = slices.Index(clicked.capture.formats, "CSV")
	clicked.chooseCaptureFormat()
	clicked.capture.input.SetValue(clickedPath)
	clicked.startCapture()

	fromTyping, err := os.ReadFile(typedPath) //nolint:gosec // a path this test just wrote
	require.NoError(t, err)
	fromClicking, err := os.ReadFile(clickedPath) //nolint:gosec // a path this test just wrote
	require.NoError(t, err)

	assert.Equal(t, string(fromTyping), string(fromClicking))
	assert.Contains(t, string(fromTyping), "alice")
}

// TestASaveWithNothingToSaveSaysSo rather than writing an empty file.
func TestASaveWithNothingToSaveSaysSo(t *testing.T) {
	m := helpModel()
	path := filepath.Join(t.TempDir(), "out.csv")

	m.runCommand("SAVE TO '" + path + "' AS CSV")

	assert.NoFileExists(t, path)
	assert.Contains(t, stripAnsiForTest(m.fullHistoryContent), "No query results available")
}

// TestTheSettingsListsStillPrintTheirResult.
//
// runCommand is theirs first - PAGING, CONSISTENCY, OUTPUT all come back as a
// string - and adding the SAVE case in front must not stop that being printed,
// nor throw the results away the way a query result would.
func TestTheSettingsListsStillPrintTheirResult(t *testing.T) {
	m := helpModel()
	m.lastTableData = [][]string{{"id"}, {"1"}}

	m.runCommand("EXPAND")

	printed := stripAnsiForTest(m.fullHistoryContent)
	assert.Contains(t, printed, "> EXPAND", "the command is echoed")
	assert.Contains(t, printed, "Expand mode is currently", "and its result printed")
	assert.True(t, m.hasTable, "a settings change does not throw the results away")
}
