package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
)

// The TRACE view, and the analysis under it.
//
// A trace is forty rows of microsecond timings and node names, and what anyone
// wants from it is which step was slow and what to do about it.

// traceModel is the TRACE view with a trace in it.
func traceModel(t *testing.T) *MainModel {
	t.Helper()

	m := helpModel()
	m.windowWidth, m.windowHeight = 100, 30
	m.viewMode = "trace"
	m.hasTrace = true
	m.historyViewport = viewport.New(viewport.WithWidth(100), viewport.WithHeight(20))
	m.traceViewport = viewport.New(viewport.WithWidth(100), viewport.WithHeight(20))
	m.trace.viewport = viewport.New(viewport.WithWidth(100), viewport.WithHeight(5))
	m.traceData = [][]string{
		{"activity", "source", "source_elapsed"},
		{"Parsing SELECT * FROM users", "10.0.0.1", "120"},
		{"Read 3 sstables", "10.0.0.2", "9100"},
	}
	m.traceInfo = &db.TraceInfo{Coordinator: "10.0.0.1", Duration: 9600}
	m.traceViewport.SetContent("activity                    source     source_elapsed\n" +
		"Parsing SELECT * FROM users 10.0.0.1   120\n" +
		"Read 3 sstables             10.0.0.2   9100")
	m.aiConfig = &config.AIConfig{Provider: "anthropic", APIKey: "sk-ant-test"}

	m.fitTracePanes()
	return m
}

// TestTheTraceIsSentAsItWasRead.
//
// The rows from system_traces rather than the drawn table: that one's columns
// are cut to the width of the screen, and what is cut off is the end of the
// activity - which is the part that says what happened.
func TestTheTraceIsSentAsItWasRead(t *testing.T) {
	m := traceModel(t)

	said := m.traceAsText()

	assert.Contains(t, said, "Coordinator: 10.0.0.1")
	assert.Contains(t, said, "Total duration: 9600 microseconds")
	assert.Contains(t, said, "Parsing SELECT * FROM users | 10.0.0.1 | 120")
	assert.Contains(t, said, "Read 3 sstables | 10.0.0.2 | 9100")
}

// TestTheViewHasAButtonToReadTheTrace.
func TestTheViewHasAButtonToReadTheTrace(t *testing.T) {
	m := traceModel(t)

	drawn := stripAnsi(m.viewTrace(m.windowWidth, m.traceHeight()))

	assert.Contains(t, drawn, analyseButton)
	assert.Contains(t, drawn, "Alt+A")
}

// TestAskingForTheAnalysisOpensThePaneAndSaysItIsWorking.
func TestAskingForTheAnalysisOpensThePaneAndSaysItIsWorking(t *testing.T) {
	m := traceModel(t)

	m, cmd := m.startTraceAnalysis()
	require.NotNil(t, cmd, "the trace should have been sent")

	assert.True(t, m.trace.open)
	assert.True(t, m.trace.running)

	drawn := stripAnsi(m.viewTrace(m.windowWidth, m.traceHeight()))
	assert.Contains(t, drawn, analysingLabel)
	assert.Contains(t, drawn, "Reading the trace")
}

// TestTheAnswerAppearsUnderTheTrace.
func TestTheAnswerAppearsUnderTheTrace(t *testing.T) {
	m := traceModel(t)
	m, _ = m.startTraceAnalysis()

	m, _ = m.traceAnalysed(traceAnalysedMsg{text: "Most of the time went on reading three sstables."})

	assert.False(t, m.trace.running)
	drawn := stripAnsi(m.viewTrace(m.windowWidth, m.traceHeight()))

	assert.Contains(t, drawn, "Read 3 sstables", "the trace is still there")
	assert.Contains(t, drawn, analysisLabel)
	assert.Contains(t, drawn, "reading three sstables")

	// The answer is under the trace, not over it.
	lines := strings.Split(drawn, "\n")
	assert.Less(t, lineHolding(lines, "Read 3 sstables"), lineHolding(lines, "reading three sstables"))
}

// TestWhatWentWrongIsSaidInThePane, rather than in a message that goes away.
func TestWhatWentWrongIsSaidInThePane(t *testing.T) {
	m := traceModel(t)
	m, _ = m.startTraceAnalysis()

	m, _ = m.traceAnalysed(traceAnalysedMsg{text: "anthropic API error: 401 Unauthorized", failed: true})

	assert.True(t, m.trace.failed)
	assert.Contains(t, stripAnsi(m.viewTrace(m.windowWidth, m.traceHeight())), "401 Unauthorized")
}

// TestThereIsNothingToAnalyseWithoutATrace, and nothing to analyse it with
// without a provider.
func TestThereIsNothingToAnalyseWithoutATrace(t *testing.T) {
	m := traceModel(t)
	m.hasTrace = false

	m, cmd := m.startTraceAnalysis()
	assert.Nil(t, cmd)
	assert.False(t, m.trace.open)
	assert.Contains(t, m.fullHistoryContent, "no trace to analyse")

	m = traceModel(t)
	m.aiConfig = nil
	m, cmd = m.startTraceAnalysis()
	assert.Nil(t, cmd)
	assert.Contains(t, m.fullHistoryContent, "No AI provider is configured")
}

// TestTheLineBetweenThePanesIsDragged.
//
// One description of where the rows are, used to draw them and to work out
// what a press landed on: two copies of that is how a drag grabs the row above
// the line being pointed at.
func TestTheLineBetweenThePanesIsDragged(t *testing.T) {
	m := traceModel(t)
	m, _ = m.startTraceAnalysis()

	rule := m.traceLayout().ruleRow
	require.True(t, m.traceRuleAt(rule+tabBarHeight), "the line is where the view says it is")
	assert.False(t, m.traceRuleAt(rule+tabBarHeight+1), "and the row under it is the analysis")

	// Pressing it starts the drag, moving takes the line with it, and letting
	// go stops.
	m, _ = m.handleMousePress(tea.Mouse{X: 10, Y: rule + tabBarHeight, Button: tea.MouseLeft})
	require.True(t, m.trace.dragging)

	m.dragTraceRule(rule + tabBarHeight - 4)
	assert.Equal(t, rule-4, m.traceLayout().ruleRow, "the line followed the pointer")
	assert.Greater(t, m.trace.viewport.Height(), 3, "and the pane grew")

	m.trace.dragging = false
	m.dragTraceRule(rule + tabBarHeight + 2)
	assert.Equal(t, rule-4, m.traceLayout().ruleRow, "let go, and it stays where it was left")
}

// TestNeitherPaneIsDraggedAway: a line dragged off the top or the bottom stops
// at the last row that leaves both panes something to show.
func TestNeitherPaneIsDraggedAway(t *testing.T) {
	m := traceModel(t)
	m, _ = m.startTraceAnalysis()
	m.trace.dragging = true

	m.dragTraceRule(-50)
	assert.GreaterOrEqual(t, m.traceViewport.Height(), traceMin)
	assert.GreaterOrEqual(t, m.trace.viewport.Height(), analysisMin)

	m.dragTraceRule(500)
	assert.GreaterOrEqual(t, m.traceViewport.Height(), traceMin)
	assert.GreaterOrEqual(t, m.trace.viewport.Height(), analysisMin)
}

// TestThePaneCanBePutAway, and the trace has the view back.
func TestThePaneCanBePutAway(t *testing.T) {
	m := traceModel(t)
	m, _ = m.startTraceAnalysis()
	m, _ = m.traceAnalysed(traceAnalysedMsg{text: "Nothing remarkable."})

	m, _ = m.handleEscapeKey()

	assert.False(t, m.trace.open)
	assert.Equal(t, -1, m.traceLayout().ruleRow)
	assert.NotContains(t, stripAnsi(m.viewTrace(m.windowWidth, m.traceHeight())), "Nothing remarkable")
}

// TestTheButtonIsWhereTheViewDrawsIt.
func TestTheButtonIsWhereTheViewDrawsIt(t *testing.T) {
	m := traceModel(t)
	g := m.traceLayout()

	assert.True(t, m.traceButtonAt(g.buttonFrom, tabBarHeight))
	assert.True(t, m.traceButtonAt(g.buttonTo, tabBarHeight))
	assert.False(t, m.traceButtonAt(g.buttonFrom-1, tabBarHeight))
	assert.False(t, m.traceButtonAt(g.buttonFrom, tabBarHeight+1), "the row under it is the trace")
}

// lineHolding is the line this text is on, or -1.
func lineHolding(lines []string, wanted string) int {
	for i, line := range lines {
		if strings.Contains(line, wanted) {
			return i
		}
	}
	return -1
}

// TestTheSectionsOfTheAnswerAreDrawnAsSections.
//
// The answer comes back as TIME, FINDINGS and WHAT TO DO, and the headings are
// drawn as headings so the shape of it can be seen without reading it.
func TestTheSectionsOfTheAnswerAreDrawnAsSections(t *testing.T) {
	assert.True(t, isHeadingLine("TIME"))
	assert.True(t, isHeadingLine("WHAT TO DO"))
	assert.True(t, isHeadingLine("  FINDINGS  "))

	assert.False(t, isHeadingLine(""))
	assert.False(t, isHeadingLine("1453us  submitting range requests"))
	assert.False(t, isHeadingLine("- Nothing remarkable."))
	assert.False(t, isHeadingLine("Read 100 live rows and 0 tombstone cells"))

	m := traceModel(t)
	m, _ = m.startTraceAnalysis()
	m, _ = m.traceAnalysed(traceAnalysedMsg{text: "TIME\n  3.9ms  seq scan across 3 sstables  ReadStage-2\n\nFINDINGS\n- A range scan over the whole ring.\n\nWHAT TO DO\n- Query by partition key.\n"})

	drawn := m.viewTrace(m.windowWidth, m.traceHeight())
	for _, line := range strings.Split(drawn, "\n") {
		if strings.Contains(stripAnsi(line), "FINDINGS") {
			assert.NotEqual(t, line, stripAnsi(line), "the heading is drawn as one")
			return
		}
	}
	t.Fatal("the answer is not in the view")
}

// TestTheAnalysisCanBeCopied.
//
// The view is two panes and a rule rather than the one viewport it used to be,
// and a selection over the trace alone cannot reach the analysis under it -
// which is the part worth copying out into a ticket.
func TestTheAnalysisCanBeCopied(t *testing.T) {
	m := traceModel(t)
	m, _ = m.startTraceAnalysis()
	m, _ = m.traceAnalysed(traceAnalysedMsg{
		text: "TIME\n  3.9ms  seq scan across 3 sstables\n\nFINDINGS\n- A range scan across the whole ring.",
	})
	_ = m.viewTrace(m.windowWidth, m.traceHeight())

	finding := lineHolding(m.trace.drawn, "- A range scan")
	require.GreaterOrEqual(t, finding, 0, "the analysis is drawn")

	m, _ = m.handleMousePress(tea.Mouse{X: 0, Y: finding + tabBarHeight, Button: tea.MouseLeft})
	m, _ = m.extendSelection(40, finding+tabBarHeight)
	m, cmd := m.endSelection()

	assert.Equal(t, "- A range scan across the whole ring.", m.lastCopied)
	assert.NotNil(t, cmd, "and it goes to the clipboard")
}

// TestTheTraceCanStillBeCopied, which it could before the pane was there.
func TestTheTraceCanStillBeCopied(t *testing.T) {
	m := traceModel(t)
	_ = m.viewTrace(m.windowWidth, m.traceHeight())

	row := lineHolding(m.trace.drawn, "Read 3 sstables")
	require.GreaterOrEqual(t, row, 0)

	m, _ = m.handleMousePress(tea.Mouse{X: 0, Y: row + tabBarHeight, Button: tea.MouseLeft})
	m, _ = m.extendSelection(60, row+tabBarHeight)
	m, _ = m.endSelection()

	assert.Contains(t, m.lastCopied, "Read 3 sstables")
}

// TestAPressOnTheRuleDoesNotStartASelection: it starts a drag of the line.
func TestAPressOnTheRuleDoesNotStartASelection(t *testing.T) {
	m := traceModel(t)
	m, _ = m.startTraceAnalysis()
	_ = m.viewTrace(m.windowWidth, m.traceHeight())

	m, _ = m.handleMousePress(tea.Mouse{X: 20, Y: m.traceLayout().ruleRow + tabBarHeight, Button: tea.MouseLeft})

	assert.True(t, m.trace.dragging)
	assert.False(t, m.selection.dragging)
}

// TestResizingTheTerminalLeavesThePaneWhereItWas.
//
// The trace was given the whole view back on every resize, which pushed the
// line between the panes off the bottom: the analysis looked squashed into the
// last rows and there was no line left on screen to drag it back with.
func TestResizingTheTerminalLeavesThePaneWhereItWas(t *testing.T) {
	m := traceModel(t)
	m, _ = m.startTraceAnalysis()
	m, _ = m.traceAnalysed(traceAnalysedMsg{text: "TIME\n  3.9ms  seq scan across 3 sstables"})
	m.ready = true

	asked := m.trace.viewport.Height()
	require.Greater(t, asked, analysisMin, "the pane opens at half the view")

	// Small enough that the panes have to be squeezed.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 14})
	m = updated.(*MainModel)

	assert.GreaterOrEqual(t, m.trace.viewport.Height(), analysisMin, "the pane is still there")
	assert.GreaterOrEqual(t, m.traceViewport.Height(), traceMin)
	assert.Less(t, m.traceLayout().ruleRow, m.traceHeight(), "and the line is still on screen to drag")

	// And back: the pane has the size it was asked for, not the size it was
	// squeezed to.
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = updated.(*MainModel)

	assert.Equal(t, asked, m.trace.viewport.Height())
}

// TestAPaneDraggedToASizeKeepsIt through a resize.
func TestAPaneDraggedToASizeKeepsIt(t *testing.T) {
	m := traceModel(t)
	m, _ = m.startTraceAnalysis()
	m.ready = true

	m.trace.dragging = true
	m.dragTraceRule(m.traceLayout().ruleRow + tabBarHeight - 3)
	m.trace.dragging = false
	dragged := m.trace.viewport.Height()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 12})
	m = updated.(*MainModel)
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = updated.(*MainModel)

	assert.Equal(t, dragged, m.trace.viewport.Height())
}

// TestBothPanesSayHowMuchThereIsToRead.
//
// A trace is forty rows and an answer is a dozen; a pane that scrolls without
// saying so looks like one that ends where the screen does.
func TestBothPanesSayHowMuchThereIsToRead(t *testing.T) {
	m := traceModel(t)
	m.traceViewport.SetContent(strings.Repeat("a row of the trace\n", 40))

	m, _ = m.startTraceAnalysis()
	m, _ = m.traceAnalysed(traceAnalysedMsg{text: strings.Repeat("a line of the answer\n", 30)})

	rows := strings.Split(stripAnsi(m.viewTrace(m.windowWidth, m.traceHeight())), "\n")
	g := m.traceLayout()

	bars := func(from, count int) int {
		found := 0
		for i := from; i < from+count && i < len(rows); i++ {
			if strings.HasSuffix(rows[i], "█") || strings.HasSuffix(rows[i], "░") {
				found++
			}
		}
		return found
	}

	assert.Equal(t, g.traceRows, bars(g.traceTop, g.traceRows), "every row of the trace has a bar")
	assert.Equal(t, g.analysisRows, bars(g.analysisTop, g.analysisRows), "and so does every row of the analysis")

	// And the rows are all the same width, bar and all.
	for i, row := range rows[g.traceTop:] {
		assert.Equal(t, m.windowWidth, len([]rune(row)), "row %d is a different width", i+g.traceTop)
	}
}

// TestTheWheelMovesTheTracePaneItIsOver.
func TestTheWheelMovesTheTracePaneItIsOver(t *testing.T) {
	m := traceModel(t)
	m.traceViewport.SetContent(strings.Repeat("a row of the trace\n", 40))
	m, _ = m.startTraceAnalysis()
	m, _ = m.traceAnalysed(traceAnalysedMsg{text: strings.Repeat("a line of the answer\n", 30)})

	g := m.traceLayout()

	m, _ = m.scrollTrace(g.traceTop+1+tabBarHeight, 3)
	assert.Equal(t, 3, m.traceViewport.YOffset())
	assert.Equal(t, 0, m.trace.viewport.YOffset(), "the pane under it did not move")

	m, _ = m.scrollTrace(g.analysisTop+1+tabBarHeight, 5)
	assert.Equal(t, 5, m.trace.viewport.YOffset())
	assert.Equal(t, 3, m.traceViewport.YOffset(), "and the trace stayed where it was")

	// Neither scrolls past what it holds.
	m, _ = m.scrollTrace(g.analysisTop+1+tabBarHeight, 500)
	assert.Equal(t, m.trace.viewport.TotalLineCount()-m.trace.viewport.Height(), m.trace.viewport.YOffset())
}

// TestTheSummarySaysHowMuchOfTheQueryTheTraceCovers.
//
// A paged query is traced per page and the total is the sum of them: without
// this the view says a scan took nine milliseconds without saying that was six
// requests, or that there were forty more it is not showing.
func TestTheSummarySaysHowMuchOfTheQueryTheTraceCovers(t *testing.T) {
	assert.Empty(t, tracePagesSaid(nil))
	assert.Empty(t, tracePagesSaid(&db.TraceInfo{Pages: 1, Shown: 1}), "a query that fitted in a page")

	assert.Equal(t, " | 6 pages", tracePagesSaid(&db.TraceInfo{Pages: 6, Shown: 6}))
	assert.Equal(t, " | 57 pages, the last 20 shown", tracePagesSaid(&db.TraceInfo{Pages: 57, Shown: 20}))
}
