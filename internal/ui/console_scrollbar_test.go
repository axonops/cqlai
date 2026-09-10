package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// consoleModel builds a Console holding n lines in a viewport showing height.
func consoleModel(n, height int) *MainModel {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = "line"
	}

	vp := viewport.New(viewport.WithWidth(40), viewport.WithHeight(height))
	vp.SetContent(strings.Join(lines, "\n"))

	return &MainModel{
		styles:          DefaultStyles(),
		viewMode:        "history",
		windowWidth:     41,
		windowHeight:    height + 4,
		historyViewport: vp,
	}
}

// bar is the scrollbar column of a rendered Console.
func bar(t *testing.T, m *MainModel) string {
	t.Helper()

	drawn := m.consoleScrollbar(m.historyViewport.View())
	var col []rune
	for _, row := range strings.Split(stripAnsiForTest(drawn), "\n") {
		runes := []rune(row)
		require.NotEmpty(t, runes)
		col = append(col, runes[len(runes)-1])
	}
	return string(col)
}

// TestTheScrollbarSaysWhereYouAreInTheTranscript.
//
// The scroll position was already worked out and appended to the status line,
// where it is pushed past the right-hand edge and has never been visible.
func TestTheScrollbarSaysWhereYouAreInTheTranscript(t *testing.T) {
	m := consoleModel(60, 10)

	atTop := bar(t, m)
	assert.Equal(t, '█', []rune(atTop)[0], "the thumb starts at the top: %q", atTop)
	assert.Contains(t, atTop, "░", "with track below it")

	m.historyViewport.GotoBottom()
	atBottom := bar(t, m)
	runes := []rune(atBottom)
	assert.Equal(t, '█', runes[len(runes)-1], "and reaches the bottom: %q", atBottom)
	assert.Equal(t, '░', runes[0], "with track above it")
}

// TestAFullBarWhenItAllFits.
//
// The column is reserved whether or not there is anything to scroll, so
// leaving it blank wastes the space and leaves you unsure whether there is
// nothing above or no scrollbar at all.
func TestAFullBarWhenItAllFits(t *testing.T) {
	m := consoleModel(3, 10)

	drawn := stripAnsiForTest(m.consoleScrollbar(m.historyViewport.View()))

	assert.Contains(t, drawn, "█", "the thumb fills the bar")
	assert.NotContains(t, drawn, "░", "with no track, because there is nowhere to go")
}

// TestNoLineLosesACharacterToIt: the viewport is a column narrower than the
// window and its content is wrapped to that width, so the bar sits beside the
// text rather than on top of it.
func TestNoLineLosesACharacterToIt(t *testing.T) {
	const window = 41
	assert.Equal(t, window-1, consoleWidth(window))

	m := consoleModel(60, 10)

	// Long lines, and enough of them to scroll - with everything showing there
	// is no bar to make room for.
	long := make([]string, 60)
	for i := range long {
		long[i] = strings.Repeat("x", 40)
	}
	m.historyViewport.SetContent(strings.Join(long, "\n"))

	drawn := stripAnsiForTest(m.consoleScrollbar(m.historyViewport.View()))
	rows := strings.Split(drawn, "\n")

	assert.True(t, strings.HasPrefix(rows[0], strings.Repeat("x", 40)),
		"the line should survive whole: %q", rows[0])
	assert.Equal(t, window, lipgloss.Width(rows[0]), "text plus the bar fills the window")
}

// TestTheConsoleThumbGrowsWithTheProportionShowing.
func TestTheConsoleThumbGrowsWithTheProportionShowing(t *testing.T) {
	count := func(m *MainModel) int {
		return strings.Count(bar(t, m), "█")
	}

	assert.Greater(t, count(consoleModel(20, 10)), count(consoleModel(200, 10)),
		"showing half should give a bigger thumb than showing a twentieth")
}

// TestAVeryLongTranscriptStillHasAVisibleThumb.
func TestAVeryLongTranscriptStillHasAVisibleThumb(t *testing.T) {
	m := consoleModel(5000, 10)

	assert.GreaterOrEqual(t, strings.Count(bar(t, m), "█"), 1)
}

// TestTheBarIsNotSelectable: the selection reads the viewport's own content,
// which the bar is not part of.
func TestTheBarIsNotSelectable(t *testing.T) {
	m := consoleModel(60, 10)
	m.input = newTestInput()
	m.mouseEnabled = true

	// Drag across the whole of the first visible line.
	m.beginSelection(0, tabBarHeight)
	m.extendSelection(60, tabBarHeight)

	assert.Equal(t, "line", m.selectedText(), "the bar is not part of the text")
}
