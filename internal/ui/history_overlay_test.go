package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testInfoBar() TopBarModel {
	return TopBarModel{
		LastCommand:  "SELECT * FROM system.local",
		QueryTime:    12 * time.Millisecond,
		RowCount:     1,
		HasQueryData: true,
	}
}

// TestTheLineDoesNotStartWithASeparator: AutoFetch used to sit to the left of
// it. It moved to the connection bar and left its separator behind.
func TestTheLineDoesNotStartWithASeparator(t *testing.T) {
	rendered := stripAnsiForTest(testInfoBar().View(140, DefaultStyles(), "history"))

	assert.False(t, strings.HasPrefix(strings.TrimLeft(rendered, " "), "│"),
		"the line opens with a separator that separates nothing: %q", rendered)
	assert.True(t, strings.HasPrefix(strings.TrimLeft(rendered, " "), "History: "),
		"the line should start with the History field: %q", rendered)
}

// TestTheFieldIsCalledHistory, because that is what clicking it opens. The
// value is still the last command run.
func TestTheFieldIsCalledHistory(t *testing.T) {
	segs := testInfoBar().segments()
	require.NotEmpty(t, segs)

	assert.Equal(t, infoHistory, segs[0].field)
	assert.Equal(t, "History: ", segs[0].label)
	assert.Equal(t, "SELECT * FROM system.local", segs[0].value)
	assert.NotContains(t, strings.Join([]string{segs[0].label}, ""), "Last")
}

// TestInfoSegmentsMatchTheRenderedLine is what the click depends on: the
// columns used for hit testing have to describe the text actually drawn.
func TestInfoSegmentsMatchTheRenderedLine(t *testing.T) {
	m := testInfoBar()

	rendered := []rune(stripAnsiForTest(m.View(140, DefaultStyles(), "history")))

	for _, seg := range m.segments() {
		require.LessOrEqual(t, seg.end, len(rendered),
			"segment %q runs past the end of the line", seg.label)

		got := string(rendered[seg.start:seg.end])
		assert.Equal(t, seg.label+seg.value, got,
			"columns %d..%d should hold %q but the line has %q",
			seg.start, seg.end, seg.label+seg.value, got)
	}
}

// TestTheFieldsAreThereBeforeAnythingRuns: the bar used to appear from nowhere
// on the first query, with the fields sliding sideways as each one filled.
func TestTheFieldsAreThereBeforeAnythingRuns(t *testing.T) {
	m := TopBarModel{}

	segs := m.segments()
	require.Len(t, segs, 3)
	for _, seg := range segs {
		assert.Equal(t, infoPlaceholder, seg.value, "%q should stand in until there is a value", seg.label)
	}

	rendered := stripAnsiForTest(m.View(140, DefaultStyles(), "history"))
	for _, want := range []string{"History: ", "Query: ", "Rows: "} {
		assert.Contains(t, rendered, want)
	}

	// The field is still the clickable one; there is just nothing to show yet.
	field, ok := m.fieldAt((segs[0].start + segs[0].end) / 2)
	assert.True(t, ok)
	assert.Equal(t, infoHistory, field)
}

// TestTheFieldsDoNotMoveWhenAQueryRuns: the labels sit in the same columns
// before and after, so the line does not jump about as you work.
func TestTheFieldsDoNotMoveWhenAQueryRuns(t *testing.T) {
	empty := TopBarModel{}.segments()
	filled := testInfoBar().segments()

	require.Len(t, empty, len(filled))
	for i := range empty {
		assert.Equal(t, empty[i].label, filled[i].label)
	}
	assert.Equal(t, empty[0].start, filled[0].start, "History should not move")
}

// TestStaleTimingIsNotCarriedOver: a command that returns no rows clears
// HasQueryData, and showing the previous query's timing beside it would put
// numbers next to a command that never produced them.
func TestStaleTimingIsNotCarriedOver(t *testing.T) {
	m := testInfoBar()
	m.LastCommand = "USE system"
	m.HasQueryData = false

	segs := m.segments()
	assert.Equal(t, "USE system", segs[0].value)
	assert.Equal(t, infoPlaceholder, segs[1].value, "the timing belonged to an earlier query")
	assert.Equal(t, infoPlaceholder, segs[2].value)
}

// TestClickingHistoryWithNoHistoryOpensNothing: an empty list would be a box
// with nothing in it.
func TestClickingHistoryWithNoHistoryOpensNothing(t *testing.T) {
	m := historyModel(0)
	m.topBar = TopBarModel{}

	seg := m.topBar.segments()[0]
	pressAt(m, (seg.start+seg.end)/2, m.windowHeight-2)

	assert.False(t, m.historySearchMode)
}

// openList puts a model into the state clicking History or pressing Ctrl+R
// leaves it in: the command list showing, with a selection and a scroll offset.
func openList(m *MainModel, scroll, selected int) {
	m.historySearchMode = true
	m.historySearchResults = m.commandHistory
	m.historySearchIndex = selected
	m.historySearchScrollOffset = scroll
}

// historyModel builds a model with n commands in its history.
func historyModel(n int) *MainModel {
	commands := make([]string, n)
	for i := range commands {
		commands[i] = fmt.Sprintf("SELECT %d FROM system.local", i)
	}

	vp := viewport.New(viewport.WithWidth(100), viewport.WithHeight(10))
	return &MainModel{
		styles:          DefaultStyles(),
		mouseEnabled:    true,
		windowWidth:     100,
		windowHeight:    30,
		viewMode:        "history",
		historyViewport: vp,
		input:           newTestInput(),
		commandHistory:  commands,
		topBar:          testInfoBar(),
		statusBar:       testStatusBar(),
	}
}

// TestClickingHistoryOpensTheList. Worth noting what this fixes: nothing in the
// program ever set showHistoryModal to true, so the whole list - drawing,
// scrolling, Enter, Escape - had no way in at all.
func TestClickingHistoryOpensTheList(t *testing.T) {
	m := historyModel(20)

	seg := m.topBar.segments()[0]
	pressAt(m, (seg.start+seg.end)/2, m.windowHeight-2)

	require.True(t, m.historySearchMode, "clicking History should open the list")
	assert.Equal(t, len(m.commandHistory)-1, m.historySearchIndex,
		"the list should open on the command just run")
}

// TestClickingHistoryAgainClosesIt, the same as the lists on the status line.
func TestClickingHistoryAgainClosesIt(t *testing.T) {
	m := historyModel(20)
	seg := m.topBar.segments()[0]

	pressAt(m, (seg.start+seg.end)/2, m.windowHeight-2)
	require.True(t, m.historySearchMode)

	pressAt(m, (seg.start+seg.end)/2, m.windowHeight-2)
	assert.False(t, m.historySearchMode)
}

// TestClickingQueryFactsOpensNothing: the timing and the row count are facts,
// not controls.
func TestClickingQueryFactsOpensNothing(t *testing.T) {
	m := historyModel(20)

	for _, seg := range m.topBar.segments()[1:] {
		pressAt(m, (seg.start+seg.end)/2, m.windowHeight-2)
		assert.False(t, m.historySearchMode, "%q is not a control", seg.label)
	}
}

// TestHistoryRowsAreWhereTheBoxSaysTheyAre ties the arithmetic to the drawing.
// Counting the title, the arrows and the separator by hand in two places is the
// same mistake in a new shape, so this checks the row against the text.
func TestHistoryRowsAreWhereTheBoxSaysTheyAre(t *testing.T) {
	for _, scroll := range []int{0, 5} {
		m := historyModel(20)
		openList(m, scroll, 19)

		modal, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
		require.True(t, ok)

		rows := strings.Split(stripAnsiForTest(layer.Content), "\n")
		for i := range modal.visibleItems() {
			row := modal.firstItemRow() + i
			require.Less(t, row, len(rows))
			assert.Contains(t, rows[row], m.commandHistory[scroll+i],
				"scrolled %d: row %d should show command %d", scroll, row, scroll+i)
		}
	}
}

// TestClickingACommandPutsItInThePrompt, and runs nothing.
func TestClickingACommandPutsItInThePrompt(t *testing.T) {
	m := historyModel(20)
	openList(m, 8, 19)

	modal, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	// The third command showing.
	pressAt(m, layer.X+3, layer.Y+modal.firstItemRow()+2)

	assert.Equal(t, m.commandHistory[10], m.input.Value())
	assert.False(t, m.historySearchMode, "picking a command should close the list")
}

// TestEveryVisibleCommandIsClickable walks the box row by row.
func TestEveryVisibleCommandIsClickable(t *testing.T) {
	for i := range 10 {
		m := historyModel(20)
		openList(m, 10, 19)

		modal, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
		require.True(t, ok)
		require.Equal(t, 10, modal.visibleItems())

		pressAt(m, layer.X+3, layer.Y+modal.firstItemRow()+i)
		assert.Equal(t, m.commandHistory[10+i], m.input.Value(), "row %d", i)
	}
}

// TestClickingTheTitleOrInstructionsPicksNothing: the box has rows that are not
// commands, and a click on one must not apply the nearest command instead.
func TestClickingTheTitleOrInstructionsPicksNothing(t *testing.T) {
	m := historyModel(20)
	openList(m, 0, 19)

	modal, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	rows := []int{
		layer.Y,     // the top border
		layer.Y + 1, // the title
		layer.Y + modal.firstItemRow() + modal.visibleItems(), // past the last command
		layer.Y + layer.Height - 2,                            // the instructions
	}
	for _, row := range rows {
		_, _, hit := m.clickHistoryCommand(layer.X+3, row)
		assert.False(t, hit, "row %d is not a command", row-layer.Y)
	}
	assert.Empty(t, m.input.Value())
}

// TestSearchRowsAreWhereTheBoxSaysTheyAre covers the Ctrl+R list, whose header
// is a different height: it has a query line, and a count line once the matches
// outnumber the rows.
func TestSearchRowsAreWhereTheBoxSaysTheyAre(t *testing.T) {
	for _, n := range []int{4, 20} {
		m := historyModel(n)
		m.historySearchMode = true
		m.historySearchQuery = "SELECT"
		m.historySearchResults = m.commandHistory
		m.historySearchIndex = 0

		modal, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
		require.True(t, ok)

		rows := strings.Split(stripAnsiForTest(layer.Content), "\n")
		for i := range modal.visibleItems() {
			row := modal.firstItemRow() + i
			require.Less(t, row, len(rows))
			assert.Contains(t, rows[row], m.historySearchResults[i],
				"%d matches: row %d should show match %d", n, row, i)
		}
	}
}

// TestClickingASearchResultPutsItInThePrompt.
func TestClickingASearchResultPutsItInThePrompt(t *testing.T) {
	m := historyModel(20)
	m.historySearchMode = true
	m.historySearchQuery = "SELECT"
	m.historySearchResults = m.commandHistory
	m.historySearchIndex = 0

	modal, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	// Read before clicking: selecting a match clears the result list.
	want := m.historySearchResults[1]
	pressAt(m, layer.X+3, layer.Y+modal.firstItemRow()+1)

	assert.Equal(t, want, m.input.Value())
	assert.False(t, m.historySearchMode, "picking a match should leave search mode")
}

// TestClickingBesideTheListPicksNothing.
func TestClickingBesideTheListPicksNothing(t *testing.T) {
	m := historyModel(20)
	openList(m, 0, 19)

	modal, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	_, _, hit := m.clickHistoryCommand(layer.X+layer.Width+2, layer.Y+modal.firstItemRow())
	assert.False(t, hit)
	assert.Empty(t, m.input.Value())
}

// TestTheListIsNotOfferedWithoutHistory.
func TestTheListIsNotOfferedWithoutHistory(t *testing.T) {
	m := historyModel(0)

	m.openHistoryList()
	assert.False(t, m.historySearchMode)

	_, _, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	assert.False(t, ok)
}

// TestHistoryShowsTheNewestCommandAtStartup: the history file is read before
// the first render, so the field has something to say from the moment cqlai
// opens rather than after the first command of the session.
func TestHistoryShowsTheNewestCommandAtStartup(t *testing.T) {
	m := historyModel(20)
	m.lastCommand = ""

	assert.Equal(t, m.commandHistory[len(m.commandHistory)-1], m.latestCommand())
}

// TestThisSessionsCommandWins over the file, once there is one.
func TestThisSessionsCommandWins(t *testing.T) {
	m := historyModel(20)
	m.lastCommand = "SELECT now() FROM system.local"

	assert.Equal(t, "SELECT now() FROM system.local", m.latestCommand())
}

// TestNoHistoryAtAllLeavesThePlaceholder: a first run with no history file.
func TestNoHistoryAtAllLeavesThePlaceholder(t *testing.T) {
	m := historyModel(0)
	m.lastCommand = ""

	require.Empty(t, m.latestCommand())

	m.topBar.LastCommand = m.latestCommand()
	assert.Equal(t, infoPlaceholder, m.topBar.segments()[0].value)
}

// TestClickingHistoryAndCtrlRGiveTheSameBox. They did the same job through two
// separate pieces of state, drawn differently, and only one of them could be
// typed into to narrow the list - which was the whole point of the other.
func TestClickingHistoryAndCtrlRGiveTheSameBox(t *testing.T) {
	clicked := historyModel(20)
	seg := clicked.topBar.segments()[0]
	pressAt(clicked, (seg.start+seg.end)/2, clicked.windowHeight-2)
	_, clickedLayer, ok := clicked.historyOverlay(clicked.windowWidth, clicked.windowHeight)
	require.True(t, ok)

	typed := historyModel(20)
	typed.handleCtrlR()
	_, typedLayer, ok := typed.historyOverlay(typed.windowWidth, typed.windowHeight)
	require.True(t, ok)

	assert.Equal(t, typedLayer.Content, clickedLayer.Content)
}

// TestTypingNarrowsTheList is what the clicked list could not do.
func TestTypingNarrowsTheList(t *testing.T) {
	m := historyModel(20)
	seg := m.topBar.segments()[0]
	pressAt(m, (seg.start+seg.end)/2, m.windowHeight-2)
	require.Len(t, m.historySearchResults, 20)

	for _, r := range "SELECT 1" {
		m.handleKeyboardInput(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	require.NotEmpty(t, m.historySearchResults)
	assert.Less(t, len(m.historySearchResults), 20, "typing should have narrowed the list")
	for _, got := range m.historySearchResults {
		assert.Contains(t, got, "SELECT 1")
	}

	_, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)
	assert.Contains(t, stripAnsiForTest(layer.Content), "Search: SELECT 1")
}

// TestTheWheelMovesTheListNotTheViewBehindIt.
func TestTheWheelMovesTheListNotTheViewBehindIt(t *testing.T) {
	m := historyModel(40)
	m.historyViewport.SetContent(strings.Repeat("line\n", 200))
	openList(m, 0, 0)

	behind := m.historyViewport.YOffset()
	m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelDown, X: 5, Y: 5})

	assert.Equal(t, wheelLines, m.historySearchIndex, "the wheel should have moved the list")
	assert.Equal(t, behind, m.historyViewport.YOffset(), "the view behind must not have moved")
}

// TestPageKeysPageTheList rather than the view behind it.
func TestPageKeysPageTheList(t *testing.T) {
	m := historyModel(40)
	m.historyViewport.SetContent(strings.Repeat("line\n", 200))
	openList(m, 0, 0)

	behind := m.historyViewport.YOffset()
	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyPgDown})

	assert.Equal(t, historyModalRows, m.historySearchIndex)
	assert.Equal(t, behind, m.historyViewport.YOffset(), "the view behind must not have moved")

	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyPgUp})
	assert.Equal(t, 0, m.historySearchIndex)
}

// TestMovingKeepsTheSelectionInView, or you page down and see nothing change.
func TestMovingKeepsTheSelectionInView(t *testing.T) {
	m := historyModel(40)
	openList(m, 0, 0)

	for range 6 {
		m.moveHistorySelection(historyModalRows)

		list, _, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
		require.True(t, ok)
		assert.GreaterOrEqual(t, m.historySearchIndex, list.scroll)
		assert.Less(t, m.historySearchIndex, list.last())
	}
	assert.Equal(t, 39, m.historySearchIndex, "paging should stop at the last command")

	for range 20 {
		m.moveHistorySelection(-historyModalRows)
	}
	assert.Equal(t, 0, m.historySearchIndex, "and at the first")
	assert.Equal(t, 0, m.historySearchScrollOffset)
}

// TestBoxWidthIsMeasuredInColumns: the arrows and the bullet in these boxes are
// multi-byte, and measuring them by length is what left the two different
// widths for the same content.
func TestBoxWidthIsMeasuredInColumns(t *testing.T) {
	m := historyModel(20)
	openList(m, 0, 19)

	_, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	for _, row := range strings.Split(stripAnsiForTest(layer.Content), "\n") {
		assert.Equal(t, layer.Width, lipgloss.Width(row), "ragged row: %q", row)
	}
}

// TestAnEmptySearchSaysSo rather than drawing an empty box.
func TestAnEmptySearchSaysSo(t *testing.T) {
	m := historyModel(20)
	m.historySearchMode = true
	m.historySearchQuery = "nothing matches this"
	m.historySearchResults = nil

	list, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	assert.Contains(t, stripAnsiForTest(layer.Content), "No matching commands")
	assert.Zero(t, list.visibleItems())

	_, _, hit := m.clickHistoryCommand(layer.X+3, layer.Y+list.firstItemRow())
	assert.False(t, hit, "there is nothing to pick")
}

// TestSearchingForSomethingWithASpace: the space key used to be taken by the
// pager before it reached the query, so any search with a space in it quietly
// matched nothing.
func TestSearchingForSomethingWithASpace(t *testing.T) {
	m := historyModel(20)
	m.handleCtrlR()

	for _, r := range "SELECT 7" {
		m.handleKeyboardInput(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	assert.Equal(t, "SELECT 7", m.historySearchQuery)
	require.Len(t, m.historySearchResults, 1)
	assert.Equal(t, "SELECT 7 FROM system.local", m.historySearchResults[0])
}

// TestBackspaceWidensTheSearchAgain.
func TestBackspaceWidensTheSearchAgain(t *testing.T) {
	m := historyModel(20)
	m.handleCtrlR()

	for _, r := range "SELECT 1" {
		m.handleKeyboardInput(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	narrowed := len(m.historySearchResults)

	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyBackspace})

	assert.Equal(t, "SELECT ", m.historySearchQuery)
	assert.Greater(t, len(m.historySearchResults), narrowed)
}

// TestASearchWithNoMatchesPicksNothing: the selection must not point past the
// end of an empty result list.
func TestASearchWithNoMatchesPicksNothing(t *testing.T) {
	m := historyModel(20)
	m.handleCtrlR()

	for _, r := range "zzz" {
		m.handleKeyboardInput(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	require.Empty(t, m.historySearchResults)
	assert.Equal(t, 0, m.historySearchIndex)

	list, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)
	assert.Zero(t, list.visibleItems())
	assert.Contains(t, stripAnsiForTest(layer.Content), "No matching commands")

	m.moveHistorySelection(historyModalRows)
	assert.Equal(t, 0, m.historySearchIndex)
}

// TestTheBoxKeepsItsHeightWhileScrolling. The "older" and "newer" rows used to
// be left out when there was nothing that way, so the box grew a row as you
// wheeled off the top and shrank again at the bottom - moving under the pointer
// as you used it.
func TestTheBoxKeepsItsHeightWhileScrolling(t *testing.T) {
	m := historyModel(40)
	openList(m, 0, 0)

	_, first, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	for range 20 {
		m.moveHistorySelection(wheelLines)

		_, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
		require.True(t, ok)
		assert.Equal(t, first.Height, layer.Height,
			"the box changed height at selection %d", m.historySearchIndex)
		assert.Equal(t, first.Y, layer.Y, "and so moved up the screen")
	}
	require.Equal(t, 39, m.historySearchIndex, "this should have reached the end")
}

// TestTheArrowsStillSayWhichWayThereIsMore, even though both rows are always
// drawn.
func TestTheArrowsStillSayWhichWayThereIsMore(t *testing.T) {
	m := historyModel(40)
	openList(m, 0, 0)

	atTop := stripAnsiForTest(mustOverlay(t, m).Content)
	assert.NotContains(t, atTop, "▲", "nothing is above the first command")
	assert.Contains(t, atTop, "▼")

	m.moveHistorySelection(100)
	atBottom := stripAnsiForTest(mustOverlay(t, m).Content)
	assert.Contains(t, atBottom, "▲")
	assert.NotContains(t, atBottom, "▼", "nothing is below the last command")
}

func mustOverlay(t *testing.T, m *MainModel) Layer {
	t.Helper()
	_, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)
	return layer
}

// TestAShortListHasNoArrowRowsAtAll: reserving them for a list that fits would
// be two blank rows for nothing.
func TestAShortListHasNoArrowRowsAtAll(t *testing.T) {
	m := historyModel(4)
	openList(m, 0, 0)

	list, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	assert.False(t, list.scrolls())
	// Border, title, query, four commands, rule, hint, border.
	assert.Equal(t, 4+6, layer.Height)
}
