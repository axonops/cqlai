package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// widthModel is a Console at the given terminal width, sized the way the app
// sizes it: through a WindowSizeMsg, so the widths under test are the ones the
// program computes rather than ones the test picked.
func widthModel(t *testing.T, width int) *MainModel {
	t.Helper()

	input := textinput.New()

	m := &MainModel{
		input:     input,
		styles:    DefaultStyles(),
		viewMode:  "history",
		statusBar: NewStatusBarModel(),
		topBar:    NewTopBarModel(),
	}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: 24})
	m, ok := updated.(*MainModel)
	require.True(t, ok, "Update returns the model")
	require.True(t, m.ready, "a window size makes it ready to draw")

	m.fullHistoryContent = strings.Repeat("line\n", 40)
	m.historyViewport.SetContent(m.wrapHistoryContent(consoleWidth(width)))

	return m
}

// widest returns the widest row of a rendered view, and how wide it is.
func widest(view string) (string, int) {
	var row string
	var width int
	for _, line := range strings.Split(view, "\n") {
		if w := lipgloss.Width(line); w > width {
			row, width = line, w
		}
	}
	return row, width
}

// TestNoRowIsWiderThanTheTerminal is the defect behind #135.
//
// View built a string called scrollInfo every frame - the current view, the
// scroll position, the output format, the mouse state - and appended it to the
// status line. StatusBarModel.View ends with Width(width), so the bar always
// comes back exactly the terminal's width whatever it holds, and the append
// therefore started one column past the right-hand edge.
//
// None of it was ever on screen. Both halves are in the initial commit, so it
// had never worked in any version of the program: it read as live code, nothing
// errored, no test failed, and there was no moment where something visible
// stopped being visible. What it left behind was rows wider than the terminal -
// 109 columns on an 80-column terminal - which is why the Console rows came out
// over-wide.
//
// Measuring the rendered view is the check that would have caught it on the day
// it was written, and it holds for anything else appended past the edge later.
func TestNoRowIsWiderThanTheTerminal(t *testing.T) {
	for _, width := range []int{40, 80, 120, 200} {
		m := widthModel(t, width)

		row, got := widest(m.View().Content)

		assert.LessOrEqual(t, got, width,
			"at %d columns the widest row is %d: %q", width, got, stripAnsiForTest(row))
	}
}

// TestTheStatusLineIsExactlyTheTerminalWidth. The bar pads itself to the full
// width, which is what left no room for anything concatenated onto it.
func TestTheStatusLineIsExactlyTheTerminalWidth(t *testing.T) {
	m := widthModel(t, 80)

	rows := strings.Split(m.View().Content, "\n")
	require.NotEmpty(t, rows)

	last := rows[len(rows)-1]
	assert.Equal(t, 80, lipgloss.Width(last),
		"the status line fills the width: %q", stripAnsiForTest(last))
}

// TestTheInputRowFitsTheTerminal.
//
// The textinput draws its prompt and a trailing cursor cell outside the width
// it is given, so it renders three columns wider than it is told. It was being
// given newWidth-2, so the input row was a column wider than the terminal and
// wrapped onto the next one.
func TestTheInputRowFitsTheTerminal(t *testing.T) {
	for _, width := range []int{40, 80, 120} {
		m := widthModel(t, width)

		assert.LessOrEqual(t, lipgloss.Width(m.input.View()), width,
			"the input row fits in %d columns", width)
	}
}

// TestATinyTerminalDoesNotPanic. v2's textinput sizes its placeholder buffer
// from the width and panics on a negative one, and a terminal that reports no
// size at all is enough to reach that.
func TestATinyTerminalDoesNotPanic(t *testing.T) {
	for _, width := range []int{0, 1, 2, 3, 4} {
		assert.NotPanics(t, func() { widthModel(t, width).View() },
			"a %d-column terminal must still draw", width)
	}
}
