package ui

import (
	"context"
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/router"
	session2 "github.com/axonops/cqlai/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// outputModel builds a model holding one page of a result, with a second page
// still to come, with OUTPUT set to the given format.
func outputModel(t *testing.T, autoFetch bool, format config.OutputFormat) *MainModel {
	t.Helper()

	session := &db.Session{}
	session.SetAutoFetch(autoFetch)

	manager := session2.NewManager(&config.Config{})
	require.NoError(t, manager.SetOutputFormat(format))

	// OUTPUT is answered by the router, which keeps its own pointer to the
	// manager: without this a command in a test would set the format somewhere
	// the model is not looking.
	router.InitRouter(manager)
	t.Cleanup(func() { router.InitRouter(nil) })

	window := NewSlidingWindowTable(1000, 10)
	window.Headers = []string{"id", "name"}
	window.ColumnNames = window.Headers
	for _, row := range [][]string{{"1", "alice"}, {"2", "bob"}} {
		window.AddRow(row, rowValues(window.Headers, row))
	}

	// A second page, delivered once and then exhausted.
	remaining := [][]string{{"3", "carol"}, {"4", "dave"}}
	window.hasMoreData = true
	window.streamingResult = &db.StreamingResult{
		LoadMore: func(_ context.Context, _ int) (db.Page, error) {
			page := db.Page{Rows: remaining}
			for _, row := range remaining {
				page.Raw = append(page.Raw, rowValues(window.Headers, row))
			}
			remaining = nil
			return page, nil
		},
	}

	return &MainModel{
		styles:          DefaultStyles(),
		session:         session,
		sessionManager:  manager,
		windowWidth:     100,
		windowHeight:    30,
		viewMode:        "history",
		slidingWindow:   window,
		input:           newTestInput(),
		historyViewport: viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
		tableViewport:   viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}
}

// TestASCIIResultsGoToTheResultsView whichever way AutoFetch is set.
//
// With it on, they used to be dumped into the console as text. The console
// wraps to the window, so an ASCII table came out with its borders broken in
// half - and the Results view is where it can scroll sideways instead.
func TestASCIIResultsGoToTheResultsView(t *testing.T) {
	for _, autoFetch := range []bool{false, true} {
		m := outputModel(t, autoFetch, config.OutputFormatASCII)
		before := m.fullHistoryContent

		m.displayResult(m.slidingWindow.Headers, nil)

		assert.Equal(t, "table", m.viewMode, "AutoFetch %v", autoFetch)
		assert.True(t, m.hasTable, "AutoFetch %v", autoFetch)
		assert.Equal(t, before, m.fullHistoryContent,
			"AutoFetch %v: nothing should have been written to the console", autoFetch)
		assert.Contains(t, m.tableViewport.GetContent(), "alice")
	}
}

// TestJSONResultsGoToTheResultsView: the same bug, in the other format.
func TestJSONResultsGoToTheResultsView(t *testing.T) {
	for _, autoFetch := range []bool{false, true} {
		m := outputModel(t, autoFetch, config.OutputFormatJSON)
		before := m.fullHistoryContent

		m.displayResult(m.slidingWindow.Headers, nil)

		assert.Equal(t, "table", m.viewMode, "AutoFetch %v", autoFetch)
		assert.True(t, m.hasTable, "AutoFetch %v", autoFetch)
		assert.Equal(t, before, m.fullHistoryContent,
			"AutoFetch %v: nothing should have been written to the console", autoFetch)
		assert.Contains(t, m.tableViewport.GetContent(), "alice")
	}
}

// TestAutoFetchStillFetchesEverything. It decides how much is pulled in, which
// is the part that should not have changed.
func TestAutoFetchStillFetchesEverything(t *testing.T) {
	m := outputModel(t, true, config.OutputFormatASCII)

	m.displayResult(m.slidingWindow.Headers, nil)

	assert.False(t, m.slidingWindow.hasMoreData, "AutoFetch should have taken the rest")
	assert.Len(t, m.slidingWindow.Rows, 4)
	assert.Contains(t, m.tableViewport.GetContent(), "dave", "the second page should be on screen")
}

// TestWithoutAutoFetchTheRestIsLeft, and the notice says so.
func TestWithoutAutoFetchTheRestIsLeft(t *testing.T) {
	m := outputModel(t, false, config.OutputFormatASCII)

	m.displayResult(m.slidingWindow.Headers, nil)

	assert.True(t, m.slidingWindow.hasMoreData)
	assert.Len(t, m.slidingWindow.Rows, 2)
	assert.Contains(t, stripAnsiForTest(m.tableViewport.GetContent()), "More data available")
}

// TestTheNoticeIsGoneOnceEverythingIsFetched: it tells you to press PgDn for
// rows that are already on screen otherwise.
func TestTheNoticeIsGoneOnceEverythingIsFetched(t *testing.T) {
	m := outputModel(t, true, config.OutputFormatASCII)

	m.displayResult(m.slidingWindow.Headers, nil)

	assert.NotContains(t, stripAnsiForTest(m.tableViewport.GetContent()), "More data available")
}

// TestAllFourFormatsAgreeOnWhereResultsGo. ASCII and JSON were the odd ones
// out; TABLE and EXPAND always used the Results view.
func TestAllFourFormatsAgreeOnWhereResultsGo(t *testing.T) {
	for _, format := range config.OutputFormats() {
		parsed, err := config.ParseOutputFormat(format)
		require.NoError(t, err)

		for _, autoFetch := range []bool{false, true} {
			m := outputModel(t, autoFetch, parsed)
			m.displayResult(m.slidingWindow.Headers, nil)
			assert.Equal(t, "table", m.viewMode, "%s with AutoFetch %v", format, autoFetch)
			assert.Equal(t, parsed, m.resultFormat, "%s with AutoFetch %v", format, autoFetch)
		}
	}
}

// TestWidestLineMeasuresColumns, since it drives horizontal scrolling and the
// content can carry styling.
func TestWidestLineMeasuresColumns(t *testing.T) {
	assert.Equal(t, 5, widestLine("abc\nabcde\nab"))
	assert.Equal(t, 0, widestLine(""))

	styled := DefaultStyles().MutedText.Render("abcde")
	assert.Equal(t, 5, widestLine(styled), "styling is not width: %q", styled)
	assert.Equal(t, 4, widestLine(strings.Join([]string{"ab", "世界"}, "\n")),
		"a wide character takes two columns")
}

// queryResultModel builds a model for the non-streaming path, which is the one
// that carries all the rows at once in a QueryResult.
func queryResultModel(t *testing.T) *MainModel {
	t.Helper()

	return &MainModel{
		styles:          DefaultStyles(),
		sessionManager:  session2.NewManager(&config.Config{}),
		windowWidth:     100,
		windowHeight:    30,
		viewMode:        "history",
		input:           newTestInput(),
		historyViewport: viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
		tableViewport:   viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}
}

// TestEveryFormatUsesTheResultsViewOnTheDirectPath.
//
// There are two paths a result can take: streaming, and everything-at-once.
// Fixing only the streaming one left this second switch sending ASCII and JSON
// to the console, with no AutoFetch involved at all - which is why the first
// fix appeared to change nothing.
func TestEveryFormatUsesTheResultsViewOnTheDirectPath(t *testing.T) {
	data := [][]string{{"id", "name"}, {"1", "alice"}, {"2", "bob"}}

	for _, format := range []config.OutputFormat{
		config.OutputFormatTable,
		config.OutputFormatASCII,
		config.OutputFormatExpand,
		config.OutputFormatJSON,
	} {
		m := queryResultModel(t)
		before := m.fullHistoryContent
		require.NoError(t, m.sessionManager.SetOutputFormat(format))

		m.showQueryResult(data, nil)

		assert.Equal(t, "table", m.viewMode, "%s", format)
		assert.True(t, m.hasTable, "%s", format)
		assert.Equal(t, before, m.fullHistoryContent,
			"%s: nothing should have been written to the console", format)
		assert.Contains(t, stripAnsiForTest(m.tableViewport.GetContent()), "alice", "%s", format)
		assert.Positive(t, m.tableWidth, "%s: horizontal scrolling needs a width", format)
	}
}

// TestJSONIsOneObjectPerRow.
func TestJSONIsOneObjectPerRow(t *testing.T) {
	out := formatRowsAsJSON([][]string{{"id", "name"}, {"1", "alice"}, {"2", "bob"}}, nil)

	lines := strings.Split(strings.TrimSpace(out), "\n")
	require.Len(t, lines, 2)
	assert.JSONEq(t, `{"id":"1","name":"alice"}`, lines[0])
	assert.JSONEq(t, `{"id":"2","name":"bob"}`, lines[1])
}

// TestSelectJSONIsPassedThrough rather than wrapped in another object.
func TestSelectJSONIsPassedThrough(t *testing.T) {
	out := formatRowsAsJSON([][]string{{"[json]"}, {`{"id": 1}`}, {`{"id": 2}`}}, nil)

	assert.Equal(t, "{\"id\": 1}\n{\"id\": 2}\n", out)
}

func TestNoRowsIsNoJSON(t *testing.T) {
	assert.Empty(t, formatRowsAsJSON([][]string{{"id"}}, nil))
	assert.Empty(t, formatRowsAsJSON(nil, nil))
}

// resultsModel puts content into the Results view as a given format, the way a
// query would.
func resultsModel(t *testing.T, format config.OutputFormat, data [][]string) *MainModel {
	t.Helper()

	m := queryResultModel(t)
	m.columnWidths = []int{8, 8} // left over from an earlier table render
	require.NoError(t, m.sessionManager.SetOutputFormat(format))
	m.showQueryResult(data, nil)
	return m
}

func wideData() [][]string {
	return [][]string{
		{"id", "payload"},
		{"1", strings.Repeat("a", 300)},
		{"2", strings.Repeat("b", 300)},
	}
}

// TestOnlyATableGetsAFrozenHeader.
//
// The header is drawn over the first rows once the view is scrolled. On JSON or
// ASCII that puts columns on screen that nothing below them lines up with -
// which is why the header appeared out of nowhere on scrolling down.
func TestOnlyATableGetsAFrozenHeader(t *testing.T) {
	for _, format := range []config.OutputFormat{
		config.OutputFormatJSON,
		config.OutputFormatASCII,
		config.OutputFormatExpand,
	} {
		m := resultsModel(t, format, wideData())
		m.tableViewport.SetYOffset(3)

		assert.Zero(t, m.stickyHeaderRows(), "%s has no columns to freeze", format)
	}

	m := resultsModel(t, config.OutputFormatTable, wideData())
	m.tableViewport.SetYOffset(3)
	assert.Equal(t, m.headerRowCount(m.tableHeaders), m.stickyHeaderRows())
}

// TestScrollingTextSidewaysDoesNotRebuildItAsATable. Scrolling right used to
// re-render lastTableData as a boxed table, so the JSON you were reading was
// replaced by a mangled table.
func TestScrollingTextSidewaysDoesNotRebuildItAsATable(t *testing.T) {
	for _, format := range []config.OutputFormat{config.OutputFormatJSON, config.OutputFormatASCII} {
		m := resultsModel(t, format, wideData())
		before := m.tableViewport.GetContent()

		m.handleHorizontalScrollRight()

		assert.Equal(t, before, m.tableViewport.GetContent(),
			"%s: the content itself should not have been rebuilt", format)
		assert.Positive(t, m.tableViewport.XOffset(), "%s: it should have slid sideways", format)
		assert.Equal(t, horizontalStep, m.horizontalOffset, "%s", format)
	}
}

// TestScrollingBackLeftReturnsToTheStart.
func TestScrollingBackLeftReturnsToTheStart(t *testing.T) {
	m := resultsModel(t, config.OutputFormatJSON, wideData())

	for range 4 {
		m.handleHorizontalScrollRight()
	}
	require.Positive(t, m.tableViewport.XOffset())

	for range 10 {
		m.handleHorizontalScrollLeft()
	}
	assert.Zero(t, m.horizontalOffset)
	assert.Zero(t, m.tableViewport.XOffset())
}

// TestSidewaysScrollingStopsAtTheEnd rather than running off into blank space.
func TestSidewaysScrollingStopsAtTheEnd(t *testing.T) {
	m := resultsModel(t, config.OutputFormatJSON, wideData())

	for range 100 {
		m.handleHorizontalScrollRight()
	}

	assert.LessOrEqual(t, m.horizontalOffset, m.tableWidth)
}

// TestNarrowContentDoesNotScrollSideways: there is nothing out there to see.
func TestNarrowContentDoesNotScrollSideways(t *testing.T) {
	m := resultsModel(t, config.OutputFormatJSON, [][]string{{"id"}, {"1"}})

	m.handleMouseWheelRight()

	assert.Zero(t, m.horizontalOffset)
}

// rowValues is a row as the values behind it, which a real result carries and a
// test has to make up.
func rowValues(headers []string, row []string) map[string]interface{} {
	values := make(map[string]interface{}, len(headers))
	for i, header := range headers {
		if i < len(row) {
			values[header] = row[i]
		}
	}
	return values
}
