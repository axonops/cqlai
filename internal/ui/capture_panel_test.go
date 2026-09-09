package ui

import (
	"fmt"
	"os"
	"path/filepath"
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

// TestCaptureSitsAtTheRightOfTheLine, out of the way of the settings.
func TestCaptureSitsAtTheRightOfTheLine(t *testing.T) {
	m := captureModel()

	segs := placeSegments(m.statusBar.segments(), m.windowWidth)
	last := segs[len(segs)-1]

	assert.Equal(t, settingCapture, last.setting)
	assert.Equal(t, m.windowWidth-statusBarPadding, last.end, "against the right edge")

	// And a click there resolves to it.
	setting, _, ok := m.statusBar.settingAt(m.windowWidth, captureColumn(m))
	require.True(t, ok)
	assert.Equal(t, settingCapture, setting)
}

// TestCaptureIsDroppedWhenTheLineIsFull rather than sitting on top of a field.
func TestCaptureIsDroppedWhenTheLineIsFull(t *testing.T) {
	m := captureModel()

	segs := placeSegments(m.statusBar.segments(), 60)
	for _, seg := range segs {
		assert.NotEqual(t, settingCapture, seg.setting, "there is no room for it here")
	}

	// The fields that remain do not overlap.
	for i := 1; i < len(segs); i++ {
		assert.GreaterOrEqual(t, segs[i].start, segs[i-1].end)
	}
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
	assert.True(t, strings.HasSuffix(m.capture.input.Value(), ".json"),
		"the default name should match the format: %q", m.capture.input.Value())
}

// TestTheCommandMatchesWhatTypingItWouldBe.
func TestTheCommandMatchesWhatTypingItWouldBe(t *testing.T) {
	assert.Equal(t, "CAPTURE JSON 'out.json'", captureCommand("JSON", "out.json"))
	assert.Equal(t, "CAPTURE PARQUET 'data.parquet'", captureCommand("PARQUET", "data.parquet"))
	assert.Equal(t, "CAPTURE CSV 'a/b.csv'", captureCommand("CSV", "a//b.csv"))
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
	assert.True(t, strings.HasSuffix(m.capture.input.Value(), ".json"))
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
	require.Len(t, m.capture.matches, files)
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

	assert.Equal(t, 39, m.capture.match, "it should stop at the last candidate")
	assert.Len(t, seen, 40, "every candidate should be reachable")

	for range 60 {
		m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyUp})
	}
	assert.Equal(t, 0, m.capture.match)
	assert.Equal(t, 0, m.capture.matchScroll)
}

// TestEnterUsesTheHighlightedCandidate rather than starting the capture.
func TestEnterUsesTheHighlightedCandidate(t *testing.T) {
	m := listingModel(t, 4)
	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyDown})

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
	pressAt(m, layer.X+3, layer.Y+1+captureMatchesStart+2)

	assert.True(t, strings.HasSuffix(m.capture.input.Value(), "file02.csv"),
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

	// The directory sorts first.
	require.Equal(t, "exports"+string(filepath.Separator), m.capture.matches[0])
	m.handleCaptureKey(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, inner+string(filepath.Separator), m.capture.input.Value())
	assert.Equal(t, []string{"one.csv", "two.csv"}, m.capture.matches,
		"it should be listing what is inside")
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
