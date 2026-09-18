package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The two tabs inside the RESULTS view.
//
// The last query's output and the trace of the requests that fetched it took
// two of the five places on a line that shortens the names to single letters
// when it runs out of room. They are one item on that line now, with the
// choice between them inside the view - and the keys did not move.

func resultModel() *MainModel {
	m := &MainModel{
		viewMode:     "table",
		hasTable:     true,
		hasTrace:     true,
		windowWidth:  120,
		windowHeight: 30,
		styles:       DefaultStyles(),
	}
	m.historyViewport.SetWidth(120)
	m.historyViewport.SetHeight(20)
	return m
}

// TestTheTraceIsNotOnTheTabLine.
func TestTheTraceIsNotOnTheTabLine(t *testing.T) {
	m := resultModel()

	var onTheLine []string
	for _, tab := range modeTabs {
		onTheLine = append(onTheLine, tab.label)
	}

	assert.Equal(t, []string{"CONSOLE", "SCHEMA", "RESULTS", "CHAT"}, onTheLine)
	assert.NotContains(t, stripAnsiForTest(m.ViewTabBar(120)), "TRACE")
}

// TestTheResultsTabIsActiveForEitherOfItsTabs: it is the view you are in
// whichever half of it you are reading.
func TestTheResultsTabIsActiveForEitherOfItsTabs(t *testing.T) {
	for _, mode := range []string{"table", "trace"} {
		m := resultModel()
		m.viewMode = mode

		var active []string
		for _, span := range m.layoutTabs(120) {
			if span.active {
				active = append(active, span.label)
			}
		}

		require.Len(t, active, 1, "%s", mode)
		assert.True(t, strings.HasPrefix(active[0], "RESULTS"), "%s marked %q active", mode, active[0])
	}
}

// TestTheResultsTabIsDimmedOnlyWhenNeitherHasAnything.
func TestTheResultsTabIsDimmedOnlyWhenNeitherHasAnything(t *testing.T) {
	results, found := tabNamed("RESULTS")
	require.True(t, found)

	empty := resultModel()
	empty.hasTable, empty.hasTrace = false, false
	assert.False(t, empty.tabShows(results), "nothing to show either way")

	traced := resultModel()
	traced.hasTable = false
	assert.True(t, traced.tabShows(results), "a trace is something to show")

	listed := resultModel()
	listed.hasTrace = false
	assert.True(t, listed.tabShows(results), "a result is something to show")
}

// tabNamed finds a tab on the line by its label.
func tabNamed(label string) (modeTab, bool) {
	for _, tab := range modeTabs {
		if tab.label == label {
			return tab, true
		}
	}
	return modeTab{}, false
}

// TestEachInnerTabIsDimmedOnItsOwn: the trace can have something to show when
// the result does not, and the other way round.
func TestEachInnerTabIsDimmedOnItsOwn(t *testing.T) {
	m := resultModel()
	m.hasTrace = false

	for _, span := range m.resultTabSpans(120) {
		switch span.mode {
		case "table":
			assert.True(t, span.available, "there is a result")
		case "trace":
			assert.False(t, span.available, "there is no trace")
		}
	}
}

// TestTheInnerTabsSayWhichKeyReachesThem.
func TestTheInnerTabsSayWhichKeyReachesThem(t *testing.T) {
	m := resultModel()

	line := stripAnsiForTest(m.viewResultTabs(120))
	assert.Contains(t, line, "QUERY RESULTS (F4)")
	assert.Contains(t, line, "TRACE (F5)")
}

// TestTheKeysStillReachWhatTheyReached: F4 the results, F5 the trace. The line
// is shorter; nothing anyone types changed.
func TestTheKeysStillReachWhatTheyReached(t *testing.T) {
	for _, key := range []struct {
		code rune
		want string
	}{
		{tea.KeyF4, "table"},
		{tea.KeyF5, "trace"},
	} {
		m := tableModel(t)
		m.hasTrace = true
		m.session = connectedSession()
		m.viewMode = "history"

		m.handleKeyboardInput(tea.KeyPressMsg{Code: key.code})

		assert.Equal(t, key.want, m.viewMode)
		assert.Equal(t, key.want, m.resultTab, "and the view remembers which tab it was left on")
	}
}

// TestClickingAnInnerTabSwitchesToIt.
func TestClickingAnInnerTabSwitchesToIt(t *testing.T) {
	m := resultModel()

	var col int
	for _, span := range m.resultTabSpans(120) {
		if span.mode == "trace" {
			col = (span.start + span.end) / 2
		}
	}

	mode, available := m.resultTabAt(col, tabBarHeight)
	require.True(t, available)
	assert.Equal(t, "trace", mode)

	// And the row above it is the tab line, not these tabs.
	mode, _ = m.resultTabAt(col, 0)
	assert.Empty(t, mode, "row 0 is the line of views")
}

// TestTheInnerTabsAreOnlyInTheResultsView.
func TestTheInnerTabsAreOnlyInTheResultsView(t *testing.T) {
	m := resultModel()
	m.viewMode = "schema"

	mode, _ := m.resultTabAt(2, tabBarHeight)
	assert.Empty(t, mode, "the schema browser has its own first row")
	assert.False(t, insideResults("schema"))
	assert.False(t, insideResults("history"))
	assert.False(t, insideResults("ai"))
}

// TestTheResultsViewStartsARowLower, which is what the inner tabs cost it.
//
// The figure is in one place: the drawing, the trace rule, the analysis pane,
// the table selection and the sticky header all ask viewTop rather than each
// carrying it, because a number written in several places is only right in the
// ones somebody remembered.
func TestTheResultsViewStartsARowLower(t *testing.T) {
	m := resultModel()

	for _, mode := range []string{"table", "trace"} {
		m.viewMode = mode
		assert.Equal(t, tabBarHeight+resultTabRows, m.viewTop(), "%s", mode)
		assert.Equal(t, m.historyViewport.Height()-resultTabRows, m.viewHeight(), "%s", mode)
	}

	for _, mode := range []string{"history", "schema", "ai"} {
		m.viewMode = mode
		assert.Equal(t, tabBarHeight, m.viewTop(), "%s", mode)
		assert.Equal(t, m.historyViewport.Height(), m.viewHeight(), "%s", mode)
	}
}

// TestTheViewIsDrawnUnderItsTabs: the strip is a row of the screen, so what is
// drawn and what is hit-tested have to agree about how many.
func TestTheViewIsDrawnUnderItsTabs(t *testing.T) {
	m := resultModel()

	lines := strings.Split(stripAnsiForTest(m.viewResultTabs(120)), "\n")
	assert.Len(t, lines, resultTabRows, "the tabs take exactly the row the geometry reserves")
}

// TestTheWelcomeSaysWhichKeyReachesWhichView.
//
// It was a list written by hand beside the tabs, and it said F3 reached the
// table, F4 the trace and F5 the schema - none of which had been true since
// the tabs were reordered.
func TestTheWelcomeSaysWhichKeyReachesWhichView(t *testing.T) {
	m := resultModel()
	m.aiConfig = configuredAI()

	lines := m.viewKeyLines()

	assert.Equal(t, []string{
		"F2 - CONSOLE",
		"F3 - SCHEMA",
		"F4 - QUERY RESULTS",
		"F5 - TRACE",
		"F6 - CHAT",
	}, lines)

	// Without a provider there is no chat to reach, and the list says so by
	// leaving it out rather than naming a key that explains itself.
	m.aiConfig = nil
	assert.NotContains(t, strings.Join(m.viewKeyLines(), "\n"), "CHAT")
}

// TestTheButtonIsOnTheRowTheScreenDrawsItOn.
//
// The tests either side of this one ask viewTop where the view starts, which
// checks that drawing and hit-testing agree but not that either is right. This
// one counts the rows of the screen as drawn - the tab line, then the view's
// own tabs, then the trace - and asks the mouse handler about the row the
// button is actually on.
func TestTheButtonIsOnTheRowTheScreenDrawsItOn(t *testing.T) {
	m := traceModel(t)

	screen := []string{stripAnsi(m.ViewTabBar(m.windowWidth))}
	screen = append(screen, strings.Split(stripAnsi(m.viewResultTabs(m.windowWidth)), "\n")...)
	screen = append(screen, strings.Split(stripAnsi(m.viewTrace(m.windowWidth, m.viewHeight())), "\n")...)

	row := lineHolding(screen, analyseButton)
	require.GreaterOrEqual(t, row, 0, "the button should be on the screen somewhere")

	assert.True(t, m.traceButtonAt(m.traceLayout().buttonFrom, row),
		"the button is drawn on row %d and the mouse looks for it elsewhere", row)

	// And the tabs are the row between the line and the trace.
	assert.Contains(t, screen[tabBarHeight], "TRACE")
	assert.Equal(t, tabBarHeight+resultTabRows, row)
}
