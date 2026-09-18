package ui

import (
	"strings"
)

// The two tabs inside the RESULTS view.
//
// The last query's output and the trace of the requests that fetched it are
// the same thing looked at two ways, and they used to take two of the five
// places on the tab line - a line that has the FILE and HELP buttons beside it
// and shortens the names to single letters when it runs out of room. They are
// one item on that line now, and the choice between them is a second, shorter
// line inside the view.
//
// The keys did not move. F4 opens the view on the results, F5 opens it on the
// trace, and both are what they reached before.

// resultTabRows is how many rows the inner tabs take, and so how much lower
// the content of these two views starts than every other view's.
const resultTabRows = 1

// resultTabs are the tabs inside the RESULTS view, in display order.
//
// The modes are the ones the rest of the shell already knows: what to scroll,
// what to select, what the wheel moves and what Esc closes are all decided by
// asking which of these the view is in, and none of that changed when the two
// views became one item on the line.
var resultTabs = []modeTab{
	// "QUERY RESULTS" rather than "RESULTS", which is the name of the view
	// these two are inside: a tab with the same name as the thing containing
	// it says nothing about which half you are looking at.
	{modes: []string{"table"}, label: "QUERY RESULTS", short: "R", key: "F4"},
	{modes: []string{"trace"}, label: "TRACE", short: "T", key: "F5"},
}

// resultModes is every view the RESULTS tab covers.
func resultModes() []string {
	modes := make([]string, 0, len(resultTabs))
	for _, t := range resultTabs {
		modes = append(modes, t.modes...)
	}
	return modes
}

// insideResults reports whether a view is one of the RESULTS tabs.
func insideResults(mode string) bool {
	for _, m := range resultModes() {
		if m == mode {
			return true
		}
	}
	return false
}

// viewTop is the first screen row of the view's content.
//
// Every view starts under the tab line. These two start one row further down,
// under their own tabs. Drawing, the trace rule, the analysis pane, the table
// selection and the sticky header all ask here rather than each carrying the
// figure, because a number written in several places is only ever right in the
// ones somebody remembered.
func (m *MainModel) viewTop() int {
	if insideResults(m.viewMode) {
		return tabBarHeight + resultTabRows
	}
	return tabBarHeight
}

// viewHeight is how many rows the view's content has, which is the window
// without the tabs, the bars and the prompt - less the inner tabs where there
// are some.
func (m *MainModel) viewHeight() int {
	if insideResults(m.viewMode) {
		return resultHeight(m.historyViewport.Height())
	}
	return m.historyViewport.Height()
}

// resultHeight is the room a RESULTS viewport has: the room every other view
// gets, less the row its own tabs take. WindowSizeMsg sizes both viewports
// from here, and viewHeight reports the same figure to whatever draws them.
func resultHeight(height int) int {
	return max(height-resultTabRows, 1)
}

// resultTabSpans is where each inner tab sits on its line.
//
// The same function draws them and decides what a click landed on, so a press
// always lands on the tab that was drawn there.
func (m *MainModel) resultTabSpans(width int) []tabSpan {
	if width <= 0 {
		return nil
	}

	labels, _ := fitTabs(resultTabs, width)

	spans := make([]tabSpan, 0, len(resultTabs))
	col := 0
	for i, t := range resultTabs {
		start := col
		if i > 0 {
			start += len(tabSeparator)
		}
		end := start + len(labels[i]) + 2
		if end > width {
			break
		}

		spans = append(spans, tabSpan{
			mode:      t.modes[0],
			label:     labels[i],
			start:     start,
			end:       end,
			available: m.tabAvailable(t.modes[0]),
			active:    m.viewMode == t.modes[0],
		})
		col = end
	}
	return spans
}

// viewResultTabs draws the line of inner tabs.
func (m *MainModel) viewResultTabs(width int) string {
	spans := m.resultTabSpans(width)
	if len(spans) == 0 {
		return strings.Repeat(" ", max(width, 0))
	}
	return renderTabs(spans, width)
}

// resultTabAt reports which inner tab a press landed on.
func (m *MainModel) resultTabAt(col, row int) (string, bool) {
	if !insideResults(m.viewMode) || row != tabBarHeight {
		return "", false
	}

	for _, span := range m.resultTabSpans(m.windowWidth) {
		if col >= span.start && col < span.end {
			return span.mode, span.available
		}
	}
	return "", false
}

// viewKeyLines is which key reaches which view, as the welcome message and the
// help both say it.
//
// Read off the tabs rather than written beside them: the RESULTS view has two
// tabs of its own, and both belong on a list of what the keys do.
func (m *MainModel) viewKeyLines() []string {
	var lines []string
	for _, tab := range modeTabs {
		if tab.covers("ai") && !m.aiAvailable() {
			continue
		}
		if tab.covers("table") {
			for _, inner := range resultTabs {
				lines = append(lines, inner.key+" - "+inner.label)
			}
			continue
		}
		lines = append(lines, tab.key+" - "+tab.label)
	}
	return lines
}
