package ui

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Drawing the TRACE view: the trace, and under it what the model made of it.
//
// One description of where the rows are, used to draw them and to work out
// what a press landed on. Two copies of that is how a drag grabs the row above
// the line being pointed at.

// analyseButton is what the button says, and what the rule says once the pane
// is open.
const (
	analyseButton  = "[ Analyse with AI ]"
	analysingLabel = "[ Analysing... ]"
	analysisLabel  = " Analysis "
	dragHint       = " drag to resize "
)

// traceGeometry is where each part of the view is, in rows from the top of it.
type traceGeometry struct {
	width  int
	height int

	traceTop  int // the first row of the trace
	traceRows int

	ruleRow      int // -1 while the pane is closed
	analysisTop  int
	analysisRows int

	buttonFrom, buttonTo int // the columns the button covers on the header row
}

// traceLayout is where everything in the TRACE view sits.
func (m *MainModel) traceLayout() traceGeometry {
	width, height := m.windowWidth, m.traceHeight()

	button := analyseButton
	if m.trace.running {
		button = analysingLabel
	}

	g := traceGeometry{
		width:      width,
		height:     height,
		traceTop:   traceHeaderRows,
		traceRows:  max(height-traceHeaderRows, 1),
		ruleRow:    -1,
		buttonFrom: max(width-lipgloss.Width(button)-1, 0),
		buttonTo:   max(width-1, 0),
	}

	if !m.trace.open {
		return g
	}

	g.traceRows = m.traceViewport.Height()
	g.ruleRow = g.traceTop + g.traceRows
	g.analysisTop = g.ruleRow + traceRuleRows
	g.analysisRows = m.trace.viewport.Height()
	return g
}

// viewTrace draws the view.
func (m *MainModel) viewTrace(width, height int) string {
	_ = height // the panes were given their rows by fitTracePanes
	g := m.traceLayout()

	rows := []string{m.traceHeaderRow(g, width)}

	if m.hasTrace {
		rows = append(rows, m.paneWithScrollbar(&m.traceViewport)...)
	} else {
		rows = append(rows, m.styles.MutedText.Render(
			"  No trace data available. Enable tracing with 'TRACING ON' to capture query traces."))
		for range max(g.traceRows-1, 0) {
			rows = append(rows, "")
		}
	}

	if m.trace.open {
		rows = append(rows, m.traceRule(width))
		rows = append(rows, m.paneWithScrollbar(&m.trace.viewport)...)
	}

	// Kept for the selection, which is over what is on screen: this view is
	// two panes and a rule rather than the one viewport it used to be, and a
	// selection over the trace alone cannot reach the analysis under it.
	m.trace.drawn = rows
	return strings.Join(rows, "\n")
}

// paneWithScrollbar is a pane's rows with a scrollbar down the right of them.
//
// Both panes hold more than they show - a trace is forty rows and an answer is
// a dozen - and a pane that scrolls without saying so looks like one that ends
// where the screen does.
func (m *MainModel) paneWithScrollbar(pane *viewport.Model) []string {
	lines := strings.Split(pane.View(), "\n")

	scrolls := pane.TotalLineCount() > pane.Height()
	thumb := scrollbarColumn(pane.Height(), pane.YOffset(), pane.TotalLineCount())

	thumbStyle := lipgloss.NewStyle().Foreground(m.styles.Accent)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	for i, line := range lines {
		lines[i] = line + scrollbarCell(thumb, i, scrolls, thumbStyle, trackStyle)
	}
	return lines
}

// traceHeaderRow is the row above the trace, holding the button.
func (m *MainModel) traceHeaderRow(g traceGeometry, width int) string {
	button := analyseButton
	style := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)

	switch {
	case m.trace.running:
		button = analysingLabel
		style = lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())
	case !m.hasTrace:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(formPlaceholderColour))
	}

	hint := ""
	if m.hasTrace && !m.trace.running {
		hint = m.styles.MutedText.Render("  Alt+A reads this trace with the AI")
	}

	gap := max(g.buttonFrom-lipgloss.Width(stripAnsi(hint)), 0)
	return hint + strings.Repeat(" ", gap) + style.Render(button)
}

// traceRule is the line between the panes, which is what is dragged to move
// them.
func (m *MainModel) traceRule(width int) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))
	if m.trace.dragging {
		style = lipgloss.NewStyle().Foreground(m.styles.Accent)
	}

	label := m.styles.MutedText.Render(analysisLabel) + style.Render(dragHint)
	drawn := lipgloss.Width(stripAnsi(label))

	line := strings.Repeat("─", max(width-drawn-2, 0))
	return style.Render("─") + label + style.Render(line) + style.Render("─")
}

// traceRuleAt reports whether a press landed on the line between the panes.
func (m *MainModel) traceRuleAt(row int) bool {
	if m.viewMode != "trace" || !m.trace.open {
		return false
	}
	return row-tabBarHeight == m.traceLayout().ruleRow
}

// traceButtonAt reports whether a press landed on the button.
func (m *MainModel) traceButtonAt(col, row int) bool {
	if m.viewMode != "trace" {
		return false
	}

	g := m.traceLayout()
	return row-tabBarHeight == 0 && col >= g.buttonFrom && col <= g.buttonTo
}

// dragTraceRule moves the line between the panes to where the pointer is.
//
// The pane keeps whatever it is given between its own smallest size and the
// one that leaves the trace its smallest: a line dragged off the top or the
// bottom of the view stops at the last row that leaves both panes usable.
func (m *MainModel) dragTraceRule(row int) {
	if !m.trace.dragging {
		return
	}

	height := m.traceHeight()
	rule := min(max(row-tabBarHeight, traceHeaderRows+traceMin), height-traceRuleRows-analysisMin)

	m.trace.rows = height - rule - traceRuleRows
	m.fitTracePanes()
}

// scrollTrace moves whichever pane the pointer is over.
func (m *MainModel) scrollTrace(row, by int) (*MainModel, tea.Cmd) {
	pane := &m.traceViewport
	if m.overTraceAnalysis(row) {
		pane = &m.trace.viewport
	}

	furthest := max(pane.TotalLineCount()-pane.Height(), 0)
	pane.SetYOffset(min(max(pane.YOffset()+by, 0), furthest))
	return m, nil
}

// overTraceAnalysis reports whether a screen row is in the analysis pane, the
// rule between the two counting as the trace above it.
func (m *MainModel) overTraceAnalysis(row int) bool {
	if !m.trace.open {
		return false
	}
	return row-tabBarHeight > m.traceLayout().ruleRow
}
