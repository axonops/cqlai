package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureModel() *MainModel {
	m := chooserModel()
	m.windowWidth = 120
	return m
}

// captureColumn is the middle of the Capture field on the status line.
func captureColumn(m *MainModel) int {
	for _, seg := range placeSegments(m.statusBar.segments(), m.windowWidth) {
		if seg.setting == settingCapture {
			return (seg.start + seg.end) / 2
		}
	}
	return -1
}

// TestCaptureSitsWithTheOtherSettings, after Fetch.
//
// It was against the right-hand edge, apart from them, because it was put there
// as a control. It is both a control and a fact about the session, and the fact
// is the more important half.
func TestCaptureSitsWithTheOtherSettings(t *testing.T) {
	m := captureModel()

	segs := placeSegments(m.statusBar.segments(), m.windowWidth)
	last := segs[len(segs)-1]

	assert.Equal(t, settingCapture, last.setting, "last in the flow")
	require.GreaterOrEqual(t, len(segs), 2)
	assert.Equal(t, settingAutoFetch, segs[len(segs)-2].setting, "after Fetch")
	assert.Equal(t, last.start, segs[len(segs)-2].end+lipgloss.Width(statusSeparator),
		"a separator away, like any other field, rather than a gap to the edge")

	// And a click there resolves to it.
	setting, _, ok := m.statusBar.settingAt(m.windowWidth, captureColumn(m))
	require.True(t, ok)
	assert.Equal(t, settingCapture, setting)
}

// TestARunningCaptureIsNeverDropped.
//
// It is the one setting that keeps doing something after you have stopped
// thinking about it, and quietly writing to a file you have forgotten is
// exactly what wants to be on screen. A capture that is off takes its turn with
// the rest: "Capture: OFF" says nothing you had not already assumed.
func TestARunningCaptureIsNeverDropped(t *testing.T) {
	m := captureModel()

	shown := func(capturing bool, width int) []string {
		m.statusBar.Capturing = capturing
		var settings []string
		for _, seg := range placeSegments(m.statusBar.segments(), width) {
			settings = append(settings, seg.setting)
		}
		return settings
	}

	for _, width := range []int{30, 40, 60, 80} {
		assert.Contains(t, shown(true, width), settingCapture,
			"a running capture stays at %d columns", width)
	}
	assert.NotContains(t, shown(false, 40), settingCapture,
		"one that is off gives way like any other field")
}

// TestClickingCaptureOpensAndClosesIt.
func TestClickingCaptureOpensAndClosesIt(t *testing.T) {
	m := captureModel()

	m.clickStatusSetting(captureColumn(m))
	require.True(t, m.capture.active)
	assert.Equal(t, captureChooseFormat, m.capture.step)

	m.clickStatusSetting(captureColumn(m))
	assert.False(t, m.capture.active)
}

// TestTheFormatsComeFromTheCommand, so the window cannot offer one CAPTURE
// will refuse.
func TestTheFormatsComeFromTheCommand(t *testing.T) {
	m := captureModel()
	m.openCapturePanel(4)

	assert.Equal(t, router.CaptureFormats(), m.capture.formats,
		"the usage lists JSON, CSV and PARQUET; TEXT is not one of the choices")
}

// TestChoosingAFormatAsksForAPath, with something to edit rather than an empty
// box.
func TestChoosingAFormatAsksForAPath(t *testing.T) {
	m := captureModel()
	m.openCapturePanel(4)

	m.capture.format = 1 // JSON
	m.chooseCaptureFormat()

	assert.Equal(t, captureEnterPath, m.capture.step)
	assert.True(t, strings.HasSuffix(m.capture.input.Value(), string(filepath.Separator)),
		"AutoSave takes a directory, so the default is one: %q", m.capture.input.Value())
}

// TestTheCommandMatchesWhatTypingItWouldBe.
func TestTheCommandMatchesWhatTypingItWouldBe(t *testing.T) {
	capture := capturePanel{kind: capturing}
	assert.Equal(t, "AUTOSAVE JSON 'out.json'", capture.command("JSON", "out.json"))
	assert.Equal(t, "AUTOSAVE PARQUET 'data.parquet'", capture.command("PARQUET", "data.parquet"))

	// SAVE has a different form entirely: SAVE TO 'file' AS FORMAT.
	save := capturePanel{kind: saving}
	assert.Equal(t, "SAVE TO 'out.csv' AS CSV", save.command("CSV", "out.csv"))
	assert.Equal(t, "SAVE TO 'a.csv' AS CSV", save.command("CSV", "a.csv"))
}

// TestEscapeGoesBackAStep, so a mistyped path does not cost you the format.
func TestEscapeGoesBackAStep(t *testing.T) {
	m := captureModel()
	m.openCapturePanel(4)
	m.chooseCaptureFormat()
	require.Equal(t, captureEnterPath, m.capture.step)

	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Equal(t, captureChooseFormat, m.capture.step)
	assert.True(t, m.capture.active, "back a step, not closed")

	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.False(t, m.capture.active)
}

// TestTheArrowsMoveWithinTheFormats and stop at the ends.
func TestTheArrowsMoveWithinTheFormats(t *testing.T) {
	m := captureModel()
	m.openCapturePanel(4)

	for range 20 {
		m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	assert.Equal(t, len(m.capture.formats)-1, m.capture.format)

	for range 20 {
		m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyUp})
	}
	assert.Equal(t, 0, m.capture.format)
}

// TestTabCompletesThePath.
func TestTabCompletesThePath(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "results.csv"), nil, 0o600))

	m := captureModel()
	m.openCapturePanel(4)
	m.chooseCaptureFormat()
	m.capture.input.SetValue(filepath.Join(dir, "res"))

	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyTab})

	assert.Equal(t, filepath.Join(dir, "results.csv"), m.capture.input.Value())
}

// TestTabListsTheCandidatesWhenFillingInIsNotEnough.
func TestTabListsTheCandidatesWhenFillingInIsNotEnough(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"report.csv", "report.json"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
	}

	m := captureModel()
	m.openCapturePanel(4)
	m.chooseCaptureFormat()
	m.capture.input.SetValue(filepath.Join(dir, "rep"))

	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyTab})

	assert.Equal(t, filepath.Join(dir, "report."), m.capture.input.Value())
	assert.Equal(t, []string{"report.csv", "report.json"}, m.capture.matches)

	// A list, one per row, not a line running off the edge.
	layer, ok := m.viewCapturePanel(m.windowWidth, 30)
	require.True(t, ok)
	rows := strings.Split(stripAnsiForTest(layer.Content), "\n")
	assert.Contains(t, rows[1+captureMatchesStart], "report.csv")
	assert.Contains(t, rows[2+captureMatchesStart], "report.json")
}

// TestTypingClearsTheListedCandidates, since they no longer describe what is
// in the box.
func TestTypingClearsTheListedCandidates(t *testing.T) {
	m := captureModel()
	m.openCapturePanel(4)
	m.chooseCaptureFormat()
	m.capture.matches = []string{"a", "b"}

	m.handleCaptureKey(tea.KeyPressMsg{Code: 'x', Text: "x"})

	assert.Empty(t, m.capture.matches)
}

// TestAnEmptyPathStartsNothing.
func TestAnEmptyPathStartsNothing(t *testing.T) {
	m := captureModel()
	m.openCapturePanel(4)
	m.chooseCaptureFormat()
	m.capture.input.SetValue("   ")

	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.True(t, m.capture.active, "it should still be asking for a path")
}

// TestTheWindowSitsAboveTheField it was opened from.
func TestTheWindowSitsAboveTheField(t *testing.T) {
	const screenHeight = 30
	m := captureModel()
	m.openCapturePanel(captureColumn(m))

	layer, ok := m.viewCapturePanel(m.windowWidth, screenHeight)
	require.True(t, ok)

	assert.Equal(t, screenHeight-1, layer.Y+layer.Height, "it should sit on the status line")
	assert.LessOrEqual(t, layer.X+layer.Width, m.windowWidth, "and stay on the screen")
}

// TestEveryFormatIsClickable, and the row that answers a click is the row
// showing that format.
func TestEveryFormatIsClickable(t *testing.T) {
	const screenHeight = 30
	m := captureModel()
	m.openCapturePanel(4)

	layer, ok := m.viewCapturePanel(m.windowWidth, screenHeight)
	require.True(t, ok)
	rows := strings.Split(stripAnsiForTest(layer.Content), "\n")

	for i, format := range m.capture.formats {
		row := layer.Y + 1 + captureHeaderRows + i

		got, hit := m.formatAt(m.windowWidth, screenHeight, layer.X+3, row)
		require.True(t, hit, "row %d should be format %q", row, format)
		assert.Equal(t, i, got)

		// The row the click resolves to is the row that shows it.
		assert.Contains(t, rows[row-layer.Y], format,
			"row %d answers for %q but does not show it", row, format)
	}
}

// TestClickingAFormatMovesOnToThePath.
func TestClickingAFormatMovesOnToThePath(t *testing.T) {
	const screenHeight = 30
	m := captureModel()
	m.windowHeight = screenHeight
	m.openCapturePanel(4)

	layer, ok := m.viewCapturePanel(m.windowWidth, screenHeight)
	require.True(t, ok)

	// The JSON row.
	json := 1
	require.Equal(t, "JSON", m.capture.formats[json])
	pressAt(m, layer.X+3, layer.Y+1+captureHeaderRows+json)

	assert.Equal(t, captureEnterPath, m.capture.step)
	assert.True(t, strings.HasSuffix(m.capture.input.Value(), string(filepath.Separator)))
}

// TestClickingTheTitleOrHintsPicksNothing: the box has rows that are not
// formats, and a click on one must not apply the nearest.
func TestClickingTheTitleOrHintsPicksNothing(t *testing.T) {
	const screenHeight = 30
	m := captureModel()
	m.openCapturePanel(4)

	layer, ok := m.viewCapturePanel(m.windowWidth, screenHeight)
	require.True(t, ok)

	for _, row := range []int{
		layer.Y,     // the border
		layer.Y + 1, // the title
		layer.Y + 1 + captureHeaderRows + len(m.capture.formats), // the blank line after
		layer.Y + layer.Height - 2,                               // the hints
	} {
		_, hit := m.formatAt(m.windowWidth, screenHeight, layer.X+3, row)
		assert.False(t, hit, "row %d is not a format", row-layer.Y)
	}
}

// TestClickingOutsideClosesIt.
func TestClickingOutsideClosesIt(t *testing.T) {
	m := captureModel()
	m.windowHeight = 30
	m.openCapturePanel(4)
	require.True(t, m.capture.active)

	pressAt(m, 100, 5)

	assert.False(t, m.capture.active)
}

// TestClickingInThePathBoxLeavesTheTypingAlone.
func TestClickingInThePathBoxLeavesTheTypingAlone(t *testing.T) {
	const screenHeight = 30
	m := captureModel()
	m.windowHeight = screenHeight
	m.openCapturePanel(4)
	m.chooseCaptureFormat()
	m.capture.input.SetValue("half-typed-path")

	layer, ok := m.viewCapturePanel(m.windowWidth, screenHeight)
	require.True(t, ok)
	pressAt(m, layer.X+3, layer.Y+2)

	assert.True(t, m.capture.active, "a press in the box should not close it")
	assert.Equal(t, "half-typed-path", m.capture.input.Value())
}

// TestClickingCaptureAgainDoesNotReopenIt in the same press.
func TestClickingCaptureAgainDoesNotReopenIt(t *testing.T) {
	m := captureModel()
	m.windowHeight = 30
	m.openCapturePanel(captureColumn(m))
	require.True(t, m.capture.active)

	pressAt(m, captureColumn(m), m.windowHeight-1)

	assert.False(t, m.capture.active)
}

// manyFiles makes a directory with n candidates, more than the list shows.
func manyFiles(t *testing.T, n int) string {
	t.Helper()

	dir := t.TempDir()
	for i := range n {
		name := filepath.Join(dir, fmt.Sprintf("file%02d.csv", i))
		require.NoError(t, os.WriteFile(name, nil, 0o600))
	}
	return dir
}

// listingModel opens the capture window with a directory listed.
func listingModel(t *testing.T, files int) *MainModel {
	t.Helper()

	m := captureModel()
	m.windowHeight = 30
	m.openCapturePanel(4)
	m.chooseCaptureFormat()
	m.capture.input.SetValue(manyFiles(t, files) + string(filepath.Separator))
	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyTab})

	// The files, and the way up at the top of them.
	require.Len(t, m.capture.matches, files+1)
	require.Equal(t, parentEntry, m.capture.matches[0])
	return m
}

// TestTheCandidatesAreAListNotALine. Joined onto one line they ran off the
// right-hand edge with no way to see or reach the rest.
func TestTheCandidatesAreAListNotALine(t *testing.T) {
	m := listingModel(t, 4)

	layer, ok := m.viewCapturePanel(m.windowWidth, 30)
	require.True(t, ok)

	rows := strings.Split(stripAnsiForTest(layer.Content), "\n")
	for i, name := range m.capture.matches {
		assert.Contains(t, rows[1+captureMatchesStart+i], name,
			"candidate %d should be on its own row", i)
	}
	for _, row := range rows {
		assert.LessOrEqual(t, lipgloss.Width(row), layer.Width, "row runs past the box: %q", row)
	}
}

// TestALongListScrollsRatherThanFillingTheScreen.
func TestALongListScrollsRatherThanFillingTheScreen(t *testing.T) {
	m := listingModel(t, 40)

	first, last := m.capture.matchWindow()
	assert.Equal(t, captureMatchRows, last-first)

	layer, ok := m.viewCapturePanel(m.windowWidth, 30)
	require.True(t, ok)
	assert.Less(t, layer.Height, 20, "the window should not grow with the listing")
}

// TestTheArrowsWalkTheWholeList, with the window following.
func TestTheArrowsWalkTheWholeList(t *testing.T) {
	m := listingModel(t, 40)

	seen := map[int]bool{}
	for range 60 {
		first, last := m.capture.matchWindow()
		require.GreaterOrEqual(t, m.capture.match, first, "the highlight left the window")
		require.Less(t, m.capture.match, last, "the highlight left the window")
		seen[m.capture.match] = true
		m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyDown})
	}

	// Forty files and the way up above them.
	assert.Equal(t, 40, m.capture.match, "it should stop at the last candidate")
	assert.Len(t, seen, 41, "every candidate should be reachable")

	for range 60 {
		m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyUp})
	}
	assert.Equal(t, 0, m.capture.match)
	assert.Equal(t, 0, m.capture.matchScroll)
}

// TestEnterUsesTheHighlightedCandidate rather than starting the capture.
func TestEnterUsesTheHighlightedCandidate(t *testing.T) {
	m := listingModel(t, 4)

	// Past the way up at the top, then onto the second file.
	for range 2 {
		m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.True(t, strings.HasSuffix(m.capture.input.Value(), "file01.csv"),
		"got %q", m.capture.input.Value())
	assert.Empty(t, m.capture.matches, "the list should have gone")
	assert.True(t, m.capture.active, "and it should not have started capturing")
}

// TestClickingACandidateUsesIt.
func TestClickingACandidateUsesIt(t *testing.T) {
	m := listingModel(t, 4)

	layer, ok := m.viewCapturePanel(m.windowWidth, m.windowHeight)
	require.True(t, ok)
	// Row 0 is the way up, so the third row is the second file.
	pressAt(m, layer.X+3, layer.Y+1+captureMatchesStart+2)

	assert.True(t, strings.HasSuffix(m.capture.input.Value(), "file01.csv"),
		"got %q", m.capture.input.Value())
}

// TestChoosingADirectoryOpensIt, so a tree can be walked without retyping.
func TestChoosingADirectoryOpensIt(t *testing.T) {
	dir := t.TempDir()
	inner := filepath.Join(dir, "exports")
	require.NoError(t, os.Mkdir(inner, 0o750))
	for _, name := range []string{"one.csv", "two.csv"} {
		require.NoError(t, os.WriteFile(filepath.Join(inner, name), nil, 0o600))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.csv"), nil, 0o600))

	m := captureModel()
	m.windowHeight = 30
	m.openCapturePanel(4)
	m.chooseCaptureFormat()
	m.capture.input.SetValue(dir + string(filepath.Separator))
	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyTab})

	// The way up is first; the directory sorts above the file beside it.
	require.Equal(t, parentEntry, m.capture.matches[0])
	require.Equal(t, "exports"+string(filepath.Separator), m.capture.matches[1])

	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyDown})
	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, inner+string(filepath.Separator), m.capture.input.Value())
	assert.Equal(t, []string{parentEntry, "one.csv", "two.csv"}, m.capture.matches,
		"it should be listing what is inside, with the way back out")
}

// TestTheWayUpWalksBackOut, which is what makes browsing two-way.
func TestTheWayUpWalksBackOut(t *testing.T) {
	dir := t.TempDir()
	inner := filepath.Join(dir, "exports")
	require.NoError(t, os.Mkdir(inner, 0o750))

	m := captureModel()
	m.windowHeight = 30
	m.openCapturePanel(4)
	m.chooseCaptureFormat()
	m.capture.input.SetValue(inner + string(filepath.Separator))
	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyTab})

	require.Equal(t, parentEntry, m.capture.matches[0])
	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, dir+string(filepath.Separator), m.capture.input.Value())
	assert.Contains(t, m.capture.matches, "exports"+string(filepath.Separator),
		"and it lists where it has arrived")
}

// TestEscapePutsTheListAwayWithoutLosingThePath.
func TestEscapePutsTheListAwayWithoutLosingThePath(t *testing.T) {
	m := listingModel(t, 4)
	path := m.capture.input.Value()

	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.Empty(t, m.capture.matches)
	assert.Equal(t, path, m.capture.input.Value())
	assert.Equal(t, captureEnterPath, m.capture.step, "still asking for a path")
}

// TestTheWheelMovesTheList.
func TestTheWheelMovesTheList(t *testing.T) {
	m := listingModel(t, 40)

	m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelDown, X: 10, Y: 10})

	assert.Equal(t, wheelLines, m.capture.match)
}

// TestTypingCaptureCentresTheWindow.
//
// A window hugging the bottom right corner for a command typed at the prompt
// looks like it belongs to something you did not touch.
func TestTypingCaptureCentresTheWindow(t *testing.T) {
	const w, h = 120, 30
	m := captureModel()
	m.windowHeight = h
	m.statusBar = testStatusBar()

	_, _, handled := m.handleSpecialCommands("CAPTURE")
	require.True(t, handled)
	require.True(t, m.capture.active)

	layer, ok := m.viewCapturePanel(w, h)
	require.True(t, ok)

	assert.Equal(t, max((w-layer.Width)/2, 0), layer.X, "centred across")
	assert.Equal(t, max((h-layer.Height)/2, 0), layer.Y, "and down")
}

// TestClickingCaptureKeepsTheWindowOnTheField, since that is what it points at.
func TestClickingCaptureKeepsTheWindowOnTheField(t *testing.T) {
	const w, h = 120, 30
	m := captureModel()
	m.windowHeight = h
	m.clickStatusSetting(captureColumn(m))
	require.True(t, m.capture.active)

	layer, ok := m.viewCapturePanel(w, h)
	require.True(t, ok)

	assert.Equal(t, h-1, layer.Y+layer.Height, "sitting on the status line")
	assert.NotEqual(t, max((w-layer.Width)/2, 0), layer.X, "not centred")
}

// TestBothRoutesGiveTheSameWindow, only in a different place.
func TestBothRoutesGiveTheSameWindow(t *testing.T) {
	const w, h = 120, 30

	clicked := captureModel()
	clicked.windowHeight = h
	clicked.clickStatusSetting(captureColumn(clicked))
	clickedLayer, ok := clicked.viewCapturePanel(w, h)
	require.True(t, ok)

	typed := captureModel()
	typed.windowHeight = h
	typed.statusBar = testStatusBar()
	typed.handleSpecialCommands("CAPTURE")
	typedLayer, ok := typed.viewCapturePanel(w, h)
	require.True(t, ok)

	assert.Equal(t, clickedLayer.Content, typedLayer.Content, "the same window")
	assert.Equal(t, clickedLayer.Width, typedLayer.Width)
}

// TestTheSaveWindowBuildsTheSaveCommand, which has a different shape from
// CAPTURE's: SAVE TO 'file' AS FORMAT.
func TestTheSaveWindowBuildsTheSaveCommand(t *testing.T) {
	save := capturePanel{kind: saving}
	assert.Equal(t, "SAVE TO '/tmp/out.json' AS JSON", save.command("JSON", "/tmp/out.json"))

	capture := capturePanel{kind: capturing}
	assert.Equal(t, "AUTOSAVE JSON '/tmp/out.json'", capture.command("JSON", "/tmp/out.json"))
}

// TestSaveOffersFormatsSomethingElseCanOpen.
//
// ASCII was here and is not any more: it wrote the table with its box drawing,
// padded to whatever width the terminal happened to be, which nothing could
// read back.
func TestSaveOffersFormatsSomethingElseCanOpen(t *testing.T) {
	m := helpModel()
	m.lastTableData = [][]string{{"id"}, {"1"}}
	m.columnTypes = []string{"int"}
	m.openSavePanel()

	assert.Equal(t, []string{"CSV", "JSON", "PARQUET"}, m.capture.formats)
	assert.NotContains(t, m.capture.formats, "ASCII")
}

// TestParquetIsOfferedOnlyWithTheTypesToWriteIt.
//
// The writer builds an Arrow schema from the CQL types, and
// AppendValueToBuilder swallows a conversion it cannot do, so a result with no
// types would write a file that opens cleanly and holds nothing. DESCRIBE is
// the common way to arrive without them.
func TestParquetIsOfferedOnlyWithTheTypesToWriteIt(t *testing.T) {
	m := helpModel()
	m.lastTableData = [][]string{{"id", "name"}, {"1", "a"}}

	m.columnTypes = nil
	assert.NotContains(t, m.formatsFor(saving), "PARQUET", "no types, no Parquet")

	m.columnTypes = []string{"int"} // one type for two columns
	assert.NotContains(t, m.formatsFor(saving), "PARQUET", "a type per column, or none")

	m.columnTypes = []string{"int", "text"}
	assert.Contains(t, m.formatsFor(saving), "PARQUET")
}

// TestTheWindowAndTheCommandAgreeAboutParquet. The window asks the command
// whether it can write Parquet, so it cannot offer a format the command will
// then refuse.
func TestTheWindowAndTheCommandAgreeAboutParquet(t *testing.T) {
	m := helpModel()
	m.lastTableData = [][]string{{"id"}, {"1"}}

	for _, types := range [][]string{nil, {"int"}, {"int", "text"}} {
		m.columnTypes = types

		offered := slices.Contains(m.formatsFor(saving), "PARQUET")
		accepted := router.ParquetTypesUsable(m.lastTableData, m.columnTypes)

		assert.Equal(t, accepted, offered, "types %v", types)
	}
}

// TestTheDefaultNameSuitsTheKind.
func TestTheDefaultNameSuitsTheKind(t *testing.T) {
	m := helpModel()

	m.openSavePanel()
	m.capture.format = slices.Index(m.capture.formats, "JSON")
	require.GreaterOrEqual(t, m.capture.format, 0, "SAVE offers JSON")
	m.chooseCaptureFormat()
	assert.Contains(t, m.capture.input.Value(), "results_")
	assert.True(t, strings.HasSuffix(m.capture.input.Value(), ".json"))

	m.closeCapturePanel()
	m.openCapturePanel(0)
	m.chooseCaptureFormat()
	assert.Contains(t, m.capture.input.Value(), "autosave_")
	assert.True(t, strings.HasSuffix(m.capture.input.Value(), string(filepath.Separator)),
		"AutoSave is given a directory, not a filename")
}

// TestBothWindowsAreTheSameWindow, differing only in what they are for.
func TestBothWindowsAreTheSameWindow(t *testing.T) {
	const w, h = 120, 30

	save := helpModel()
	save.windowHeight = h
	save.openSavePanel()
	saveLayer, ok := save.viewCapturePanel(w, h)
	require.True(t, ok)

	capture := helpModel()
	capture.windowHeight = h
	capture.openCapturePanelCentred()
	captureLayer, ok := capture.viewCapturePanel(w, h)
	require.True(t, ok)

	// Each is centred for its own width; the two differ because the titles and
	// the format lists do.
	assert.Equal(t, max((w-saveLayer.Width)/2, 0), saveLayer.X, "save centred")
	assert.Equal(t, max((w-captureLayer.Width)/2, 0), captureLayer.X, "capture centred")
	assert.Contains(t, stripAnsiForTest(saveLayer.Content), "Save the last results")
	assert.Contains(t, stripAnsiForTest(captureLayer.Content), "Save each query")
}

// TestTabCompletesThePathInTheSaveWindowToo, which the old save modal could
// not do.
func TestTabCompletesThePathInTheSaveWindowToo(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "results.csv"), nil, 0o600))

	m := helpModel()
	m.openSavePanel()
	m.chooseCaptureFormat()
	m.capture.input.SetValue(filepath.Join(dir, "res"))

	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyTab})

	assert.Equal(t, filepath.Join(dir, "results.csv"), m.capture.input.Value())
}

// TestTheButtonAndTheCommandAgree about whether there is anything to save.
func TestTheButtonAndTheCommandAgree(t *testing.T) {
	m := helpModel()

	assert.False(t, m.hasResults(), "nothing run yet")

	m.lastTableData = [][]string{{"id"}, {"1"}}
	assert.True(t, m.hasResults())
}

// TestClickingATabWithTheSaveWindowOpenSwitchesView, rather than only closing
// the window: the tabs are still live under it.
func TestClickingATabWithTheSaveWindowOpenSwitchesView(t *testing.T) {
	m := helpModel()
	m.windowHeight = 30
	m.lastTableData = [][]string{{"id"}, {"1"}}
	m.openSavePanel()

	results := spans(m, "table")
	pressAt(m, (results.start+results.end)/2, 0)

	assert.False(t, m.capture.active, "the window should have gone")
	assert.Equal(t, "table", m.viewMode, "and the tab should have acted")
}
