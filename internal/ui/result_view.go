package ui

import (
	"fmt"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/router"
)

// Drawing a result into the Results view.
//
// One function does it, for every format and from every path that puts rows on
// screen: the first display of a result, another page of rows arriving, and now
// a change of OUTPUT while the result is still up.
//
// It was four display functions and a refresh, each setting some of what the
// view needs and leaving the rest. That is why the format ended up written in
// three places - the session manager, m.resultFormat, and whichever display
// function last ran - and why the width the sideways scrolling clamps against
// was only ever set when a result first arrived.

// outputFormat is what OUTPUT is set to, or TABLE when there is nothing to ask.
func (m *MainModel) outputFormat() config.OutputFormat {
	if m.sessionManager == nil {
		return config.OutputFormatTable
	}
	return m.sessionManager.GetOutputFormat()
}

// renderResults draws allData - its first row the headers - into the Results
// view, and leaves behind everything that describes what it drew.
//
// Paging snaps scroll offsets to tableRowBoundaries. Those are line numbers, so
// they only mean anything for the layout that produced them and have to be
// replaced whenever the content is. Leaving stale ones behind caps scrolling:
// boundaries from the boxed table, where a record is one line, cannot reach past
// the row count in expand format, where a record is as tall as the table is wide.
func (m *MainModel) renderResults(allData [][]string) {
	format := m.outputFormat()
	m.resultFormat = format

	var (
		content    string
		boundaries []int
	)
	switch format {
	case config.OutputFormatASCII:
		content = FormatASCIITableWithTypes(allData, m.columnTypes)
	case config.OutputFormatJSON:
		content = formatRowsAsJSON(allData, m.rawRows())
	case config.OutputFormatExpand:
		content, boundaries = FormatExpandTableWithBoundaries(allData, m.styles)
	default:
		content = m.formatTableForViewport(allData)
		boundaries = m.tableRowBoundaries // filled in while rendering
	}

	content += m.moreRowsNotice(format)
	if content == "" {
		content = "No results"
	}

	m.tableRowBoundaries = boundaries
	m.tableViewport.SetContent(content)

	// What was drawn, so the cache check and anything else asking what is on
	// screen agree with the screen. Set after rendering: isSameTableData
	// compares against it to decide whether the boxed table needs rebuilding,
	// and setting it first would answer yes to every question.
	m.lastTableData = allData

	// A boxed table's width is the width of the whole table rather than of what
	// is on screen, since the content has already been cut to the window by the
	// horizontal offset; formatTableForViewport measures it before cutting.
	// Everything else is text, and as wide as its longest line.
	if format != config.OutputFormatTable {
		m.tableWidth = widestLine(content)
	}
}

// onePassFormat reports whether a format is drawn in one go from the rows in
// hand, rather than growing a row at a time as they arrive.
//
// ASCII art and JSON lines are: the whole thing is laid out at once, so what is
// on screen is exactly the rows fetched so far. The boxed table and expand
// format are read a screen at a time and page in as you scroll. It decides two
// things - whether to pull the rest of the rows in first, and whether to say
// how many are showing - and both follow from the same fact.
func onePassFormat(format config.OutputFormat) bool {
	return format == config.OutputFormatASCII || format == config.OutputFormatJSON
}

// moreRowsNotice says how much of a result is showing, at the end of it.
//
// At the end because that is where the rows run out, and where you are looking
// when you want the rest. It was written over the prompt's placeholder instead,
// which is one row shared with whatever else has something to say there: a
// statement typed over several lines lost its own hint to it.
//
// Every format says it. The boxed table and expand format page in as you
// scroll, so the line moves down the result as more of it arrives; ASCII art
// and JSON lines are drawn in one pass, so it sits under what there is.
func (m *MainModel) moreRowsNotice(_ config.OutputFormat) string {
	if m.slidingWindow == nil || !m.slidingWindow.hasMoreData {
		return ""
	}

	return "\n" + m.styles.MutedText.Render(fmt.Sprintf(
		"(Showing %d rows. More data available. Use PgDn/Space to load more, or AUTOFETCH ON to fetch all.)",
		len(m.slidingWindow.Rows)))
}

// redrawResults draws the result already on screen again, in whatever OUTPUT is
// now set to.
//
// The rows are in hand either way, so changing the format is a redraw rather
// than a reason to ask Cassandra the same question again. Nothing is fetched
// that would not have been fetched anyway: fetchAllPages does nothing unless
// AUTOFETCH is on, and a format switch turning into a full table scan is not
// what anyone pressing it is asking for.
func (m *MainModel) redrawResults() {
	if !m.hasTable || len(m.lastTableData) == 0 {
		return
	}

	// Everything that describes the layout is about to stop being true.
	m.cachedTableLines = nil
	m.initialColumnWidths = nil
	m.columnWidths = nil
	m.tableRowBoundaries = nil
	m.resetHorizontalScroll()

	format := m.outputFormat()
	m.fetchAllPages(string(format))

	m.renderResults(m.resultRows())

	// Back to the top: row 500 of a boxed table is not line 500 of JSON, and
	// there is no honest way to hold the place across the change.
	m.tableViewport.GotoTop()
}

// resultRows is the result on screen: the header row, then the rows in hand.
//
// A streaming result grows as pages arrive, so the sliding window is the newer
// answer whenever there is one for this result. A complete result has none -
// showQueryResult clears it - and what was drawn is what it was given.
func (m *MainModel) resultRows() [][]string {
	if m.slidingWindow != nil && len(m.slidingWindow.Headers) > 0 {
		return append([][]string{m.slidingWindow.Headers}, m.slidingWindow.Rows...)
	}
	return m.lastTableData
}

// processCommand runs a command, and draws the result on screen again if the
// command changed how results are drawn.
//
// Comparing the format either side of the command, rather than looking at the
// command itself: OUTPUT sets it, EXPAND ON and OFF set it, the control on the
// status line sets it by running OUTPUT, and none of them have to know that
// anything is on screen. Anything else that comes to set it is covered for
// free, which is the opposite of how the format used to be handled.
func (m *MainModel) processCommand(command string) interface{} {
	before := m.outputFormat()
	result := router.ProcessCommand(command, m.session, m.sessionManager)

	if m.outputFormat() != before {
		m.redrawResults()
	}
	return result
}

// rawRows is the values behind the rows on screen, where they are still in
// hand: a streaming result keeps them beside its window of rows, and a complete
// one keeps what it was given.
func (m *MainModel) rawRows() []map[string]interface{} {
	if m.slidingWindow != nil && len(m.slidingWindow.Headers) > 0 {
		return m.slidingWindow.RawRows
	}
	return m.lastRawData
}
