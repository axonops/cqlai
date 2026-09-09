package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestInput() textinput.Model {
	in := textinput.New()
	in.SetWidth(40)
	return in
}

// selectionModel builds a history view holding the given lines, in a viewport
// tall enough to show height of them.
func selectionModel(height int, lines ...string) *MainModel {
	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(height))
	vp.SetContent(strings.Join(lines, "\n"))

	return &MainModel{
		styles:          DefaultStyles(),
		mouseEnabled:    true,
		windowWidth:     80,
		windowHeight:    height + 4, // tabs, input, info bar, status bar
		viewMode:        "history",
		historyViewport: vp,
	}
}

// drag runs a whole press-move-release, the way the terminal delivers one.
func drag(m *MainModel, fromCol, fromRow, toCol, toRow int) tea.Cmd {
	m.beginSelection(fromCol, fromRow)
	m.extendSelection(toCol, toRow)
	_, cmd := m.endSelection()
	return cmd
}

// clipboardText runs the command endSelection returned and reads the text back
// out, which is what would reach the terminal as OSC 52.
//
// Bubble Tea keeps the message type unexported, so this goes by its name and
// its contents rather than a type assertion.
func clipboardText(t *testing.T, cmd tea.Cmd) string {
	t.Helper()
	require.NotNil(t, cmd, "releasing a drag should have produced a clipboard command")

	msg := cmd()
	require.Contains(t, fmt.Sprintf("%T", msg), "lipboard",
		"expected a clipboard command, got %T", msg)
	return fmt.Sprintf("%s", msg)
}

// TestDragSelectsAndCopies is the point of the feature: no modifier key, and
// the text ends up on the clipboard.
func TestDragSelectsAndCopies(t *testing.T) {
	m := selectionModel(5, "first line", "second line", "third line")

	// Row 0 is the tab bar, so the first content line is row 1.
	cmd := drag(m, 0, 1, 5, 2)

	assert.Equal(t, "first line\nsecon", clipboardText(t, cmd))
	assert.True(t, m.selection.active, "the highlight should stay up after release")
	assert.False(t, m.selection.dragging)
}

// TestDragBackwardsSelectsTheSameText: a selection dragged right to left is
// the same span, and span() is what puts it in reading order.
func TestDragBackwardsSelectsTheSameText(t *testing.T) {
	forwards := selectionModel(5, "first line", "second line")
	backwards := selectionModel(5, "first line", "second line")

	assert.Equal(t,
		clipboardText(t, drag(forwards, 2, 1, 6, 2)),
		clipboardText(t, drag(backwards, 6, 2, 2, 1)))
}

// TestClickWithoutDraggingCopiesNothing guards against a stray click wiping
// whatever the user had on their clipboard.
func TestClickWithoutDraggingCopiesNothing(t *testing.T) {
	m := selectionModel(5, "first line", "second line")

	m.beginSelection(4, 1)
	_, cmd := m.endSelection()

	assert.Nil(t, cmd, "a click that selected nothing should not touch the clipboard")
	assert.False(t, m.selection.active)
}

// TestSelectionSurvivesScrolling is why the span is stored against the content
// rather than against screen rows: dragging past the bottom edge scrolls, and
// the selection has to keep meaning the same characters.
func TestSelectionSurvivesScrolling(t *testing.T) {
	m := selectionModel(3, "one", "two", "three", "four", "five", "six")

	// Start on the top visible line, then drag below the bottom edge twice.
	m.beginSelection(0, 1)
	m.extendSelection(0, 4)
	m.extendSelection(3, 4)

	assert.Equal(t, 2, m.historyViewport.YOffset(), "dragging past the edge should scroll")

	// Two lines scrolled past, so the head sits on "five" rather than on the
	// row it was dragged to.
	_, cmd := m.endSelection()
	assert.Equal(t, "one\ntwo\nthree\nfour\nfiv", clipboardText(t, cmd))
}

// TestScrollingAfterReleaseKeepsTheHighlightOnItsText checks the same thing
// from the other side: the highlight follows the text up the screen.
func TestScrollingAfterReleaseKeepsTheHighlightOnItsText(t *testing.T) {
	m := selectionModel(3, "alpha", "bravo", "charlie", "delta", "echo")

	drag(m, 0, 2, 5, 2) // all of "bravo", the second line

	before := m.highlightSelection(m.historyViewport.View())
	m.historyViewport.SetYOffset(1)
	after := m.highlightSelection(m.historyViewport.View())

	// Second row before the scroll, first row after it, and marked both times.
	marked := selectionStyle.Render("bravo")
	assert.Contains(t, strings.Split(before, "\n")[1], marked,
		"bravo should be highlighted before scrolling")
	assert.Contains(t, strings.Split(after, "\n")[0], marked,
		"bravo should still be highlighted after scrolling")
	assert.NotContains(t, strings.Split(after, "\n")[1], selectionStyle.Render("charlie"),
		"charlie was never selected")
}

func TestDoubleClickSelectsAWord(t *testing.T) {
	m := selectionModel(5, "SELECT id FROM system_schema.tables")

	m.beginSelection(12, 1) // inside "FROM"
	m.endSelection()
	m.beginSelection(12, 1)
	_, cmd := m.endSelection()

	assert.Equal(t, "FROM", clipboardText(t, cmd))
}

// TestDoubleClickTakesAWholeIdentifier: underscores and digits are part of a
// word here, because the words worth double-clicking are schema names.
func TestDoubleClickTakesAWholeIdentifier(t *testing.T) {
	m := selectionModel(5, "keyspace system_schema_2 ok")

	for range 2 {
		m.beginSelection(12, 1) // inside "system_schema_2"
		m.endSelection()
	}

	assert.Equal(t, "system_schema_2", m.selectedText())
}

func TestTripleClickSelectsTheLine(t *testing.T) {
	m := selectionModel(5, "first line", "second line here", "third")

	for range 3 {
		m.beginSelection(4, 2)
		m.endSelection()
	}

	assert.Equal(t, "second line here", m.selectedText())
}

// TestSlowClicksDoNotCountAsADoubleClick: the count is on time as well as
// position, or every second click on the same spot would grab a word.
func TestSlowClicksDoNotCountAsADoubleClick(t *testing.T) {
	m := selectionModel(5, "SELECT id FROM tables")

	m.beginSelection(12, 1)
	m.endSelection()
	m.selection.lastPress = time.Now().Add(-2 * time.Second)
	m.beginSelection(12, 1)

	assert.True(t, m.selection.empty(), "a click long after the last one starts a fresh selection")
}

// TestPressOnTheTabBarStartsNoSelection: the rows outside the viewport belong
// to the controls, and a press there must not leave a selection behind.
func TestPressOnTheTabBarStartsNoSelection(t *testing.T) {
	m := selectionModel(5, "first line", "second line")
	m.selection.active = true

	m.beginSelection(3, 0)

	assert.False(t, m.selection.active)
}

// TestPressBelowTheViewportStartsNoSelection covers the input line and the
// bars under it.
func TestPressBelowTheViewportStartsNoSelection(t *testing.T) {
	m := selectionModel(3, "first line", "second line")

	m.beginSelection(3, 1+3) // one row past the last content row

	assert.False(t, m.selection.active)
}

// TestAnyKeyDropsTheSelection stops a highlight sitting over text that has
// since moved or changed.
func TestAnyKeyDropsTheSelection(t *testing.T) {
	m := selectionModel(5, "first line", "second line")
	m.input = newTestInput()

	drag(m, 0, 1, 5, 1)
	require.True(t, m.selection.active)

	m.handleKeyboardInput(tea.KeyPressMsg{Code: 'a', Text: "a"})
	assert.False(t, m.selection.active)
}

// TestEscapeOnlyDropsTheSelection: with a selection up, Escape takes it down
// and does nothing else.
func TestEscapeOnlyDropsTheSelection(t *testing.T) {
	m := selectionModel(5, "first line", "second line")
	m.input = newTestInput()
	m.input.SetValue("SELECT 1")

	drag(m, 0, 1, 5, 1)
	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.False(t, m.selection.active)
	assert.Equal(t, "SELECT 1", m.input.Value(), "Escape should not have cleared the prompt as well")
}

// TestSwitchingViewDropsTheSelection: the span belongs to one viewport's
// content, so it must not be painted onto another's.
func TestSwitchingViewDropsTheSelection(t *testing.T) {
	m := selectionModel(5, "first line", "second line")
	drag(m, 0, 1, 5, 1)

	m.viewMode = "trace"
	m.hasTrace = true
	m.traceViewport = viewport.New(viewport.WithWidth(80), viewport.WithHeight(5))
	m.traceViewport.SetContent("trace line\nanother")

	assert.Equal(t, "", m.selectedText(), "the selection belongs to the console view")
	assert.Equal(t, m.traceViewport.View(), m.highlightSelection(m.traceViewport.View()))
}

// TestCopiedLinesLoseTrailingSpaces: a table pads every cell, and pasting a
// column of blanks into an editor is not what anyone wanted.
func TestCopiedLinesLoseTrailingSpaces(t *testing.T) {
	m := selectionModel(5, "id    ", "1     ", "2     ")

	cmd := drag(m, 0, 1, 6, 3)

	assert.Equal(t, "id\n1\n2", clipboardText(t, cmd))
}

// TestHighlightKeepsTheLineIntact guards the ANSI arithmetic: the visible
// characters must survive being cut into three and put back together.
func TestHighlightKeepsTheLineIntact(t *testing.T) {
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("hello") + " world"

	for from := range 11 {
		for to := from; to <= 11; to++ {
			got := stripAnsiForTest(highlightColumns(styled, from, to))
			assert.Equal(t, "hello world", got, "columns %d..%d lost characters", from, to)
		}
	}
}

// TestHighlightMarksOnlyTheSelectedColumns.
func TestHighlightMarksOnlyTheSelectedColumns(t *testing.T) {
	line := "hello world"

	marked := highlightColumns(line, 6, 11)

	assert.Contains(t, marked, selectionStyle.Render("world"))
	assert.False(t, strings.Contains(marked, selectionStyle.Render("hello")))
}

// TestColumnsAreCellsNotBytes: a column is a cell on screen. Counting bytes
// puts the highlight in the wrong place the moment a name is not ASCII.
func TestColumnsAreCellsNotBytes(t *testing.T) {
	m := selectionModel(5, "id | naïve café")

	cmd := drag(m, 5, 1, 15, 1)

	assert.Equal(t, "naïve café", clipboardText(t, cmd))
}

// TestWideCharactersFillTwoColumns.
func TestWideCharactersFillTwoColumns(t *testing.T) {
	cells := lineCells("a世b")

	require.Len(t, cells, 4, "a wide character takes two columns")
	assert.Equal(t, 'a', cells[0])
	assert.Equal(t, '世', cells[1])
	assert.Equal(t, rune(0), cells[2], "the second column of a wide character holds nothing of its own")
	assert.Equal(t, 'b', cells[3])
}

// TestStickyHeaderRowsAreNotSelectable: while a table is scrolled the top
// three rows show the top of the table, not the lines at that scroll position,
// so a selection anchored there would copy something else entirely.
func TestStickyHeaderRowsAreNotSelectable(t *testing.T) {
	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(6))
	vp.SetContent(strings.Join([]string{"top", "head", "sep", "r1", "r2", "r3", "r4", "r5"}, "\n"))
	vp.SetYOffset(2)

	m := &MainModel{
		styles:        DefaultStyles(),
		viewMode:      "table",
		hasTable:      true,
		resultFormat:  config.OutputFormatTable,
		tableViewport: vp,
		tableHeaders:  []string{"id"},
		lastTableData: [][]string{{"id"}},
		columnWidths:  []int{4},
		windowWidth:   80,
		windowHeight:  10,
	}

	require.Equal(t, stickyHeaderHeight, m.stickyHeaderRows())

	for row := tabBarHeight; row < tabBarHeight+stickyHeaderHeight; row++ {
		_, _, ok := m.docPosition(0, row)
		assert.False(t, ok, "row %d is the frozen header", row)
	}

	line, _, ok := m.docPosition(0, tabBarHeight+stickyHeaderHeight)
	require.True(t, ok)
	assert.Equal(t, 5, line, "the first selectable row is the content at the scroll offset")
}

// TestSelectionIsOffWhenThereIsNothingToSelect: the empty-view messages are
// drawn through a throwaway viewport, with no content to anchor to.
func TestSelectionIsOffWithoutContent(t *testing.T) {
	m := &MainModel{styles: DefaultStyles(), viewMode: "table", hasTable: false}

	vp, name := m.selectionTarget()
	assert.Nil(t, vp)
	assert.Equal(t, "", name)

	m.beginSelection(2, 2)
	assert.False(t, m.selection.active)
}

// TestMouseOffDropsTheSelection: with the mouse handed back there is nothing
// keeping the highlight up to date.
func TestMouseOffDropsTheSelection(t *testing.T) {
	m := selectionModel(5, "first line", "second line")
	m.input = newTestInput()
	drag(m, 0, 1, 5, 1)

	m.handleMouseCommand("MOUSE OFF")

	assert.False(t, m.mouseEnabled)
	assert.False(t, m.selection.active)
}
