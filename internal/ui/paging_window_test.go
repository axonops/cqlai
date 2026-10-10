package ui

import (
	"context"
	"regexp"
	"strconv"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/router"
	session2 "github.com/axonops/cqlai/internal/session"
)

// pagingModel is a result of total rows, of which the window keeps at most
// keep, delivered a page at a time; asked records the size of each request.
func pagingModel(t *testing.T, total, keep, pageSize int, asked *[]int) *MainModel {
	t.Helper()
	sess := &db.Session{}
	sess.SetPageSize(pageSize)
	manager := session2.NewManager(&config.Config{})
	require.NoError(t, manager.SetOutputFormat(config.OutputFormatTable))
	router.InitRouter(manager)
	t.Cleanup(func() { router.InitRouter(nil) })

	window := NewSlidingWindowTable(keep, 100)
	window.Headers = []string{"id"}
	window.ColumnNames = window.Headers
	next := 0
	load := func(n int) db.Page {
		var page db.Page
		for ; n > 0 && next < total; n-- {
			row := []string{strconv.Itoa(next)}
			page.Rows = append(page.Rows, row)
			page.Raw = append(page.Raw, rowValues(window.Headers, row))
			next++
		}
		page.HasMore = next < total
		return page
	}
	first := load(20)
	for i, row := range first.Rows {
		window.AddRow(row, first.Raw[i])
	}
	window.hasMoreData = true
	window.streamingResult = &db.StreamingResult{
		LoadMore: func(_ context.Context, n int) (db.Page, error) {
			*asked = append(*asked, n)
			return load(n), nil
		},
	}

	m := &MainModel{
		styles:          DefaultStyles(),
		session:         sess,
		sessionManager:  manager,
		windowWidth:     60,
		windowHeight:    20,
		viewMode:        "history",
		slidingWindow:   window,
		input:           newTestInput(),
		historyViewport: viewport.New(viewport.WithWidth(60), viewport.WithHeight(10)),
		tableViewport:   viewport.New(viewport.WithWidth(60), viewport.WithHeight(10)),
	}
	m.displayResult(window.Headers, nil)
	m.viewMode = "table"
	m.hasTable = true
	return m
}

var cellNumber = regexp.MustCompile(`│\s*(\d+)\s*│`)

// pageThrough presses PageDown until the screen stops changing, and returns
// every row that was on it.
func pageThrough(m *MainModel) map[int]bool {
	seen := map[int]bool{}
	last := ""
	for i := 0; i < 500; i++ {
		screen := stripAnsiForTest(m.tableViewport.View())
		for _, match := range cellNumber.FindAllStringSubmatch(screen, -1) {
			n, _ := strconv.Atoi(match[1])
			seen[n] = true
		}
		if screen == last && !m.slidingWindow.hasMoreData {
			break
		}
		last = screen
		m, _ = m.handlePageDown(tea.KeyPressMsg{Code: tea.KeyPgDown})
	}
	return seen
}

// TestPagingShowsEveryRowWhenTheWindowIsFull: once the window holds all it
// may, rows are dropped from its top, which moves every line up. The screen
// was not moved with them, so each page down landed further on and the rows
// in between were never shown.
func TestPagingShowsEveryRowWhenTheWindowIsFull(t *testing.T) {
	var asked []int
	m := pagingModel(t, 400, 60, 25, &asked)
	seen := pageThrough(m)
	for i := 0; i < 400; i++ {
		if !seen[i] {
			t.Fatalf("row %d was never on the screen (%d of 400 seen)", i, len(seen))
		}
	}
	assert.Greater(t, m.slidingWindow.FirstRowIndex, int64(0), "the window dropped rows on the way")
}

// TestPagingOffStillLoadsMore: PAGING OFF is a page size of 0, and asking for
// 0 more rows loaded none, so nothing past the first page could be reached.
func TestPagingOffStillLoadsMore(t *testing.T) {
	var asked []int
	m := pagingModel(t, 250, 10000, 0, &asked)
	seen := pageThrough(m)
	assert.Len(t, seen, 250)
	require.NotEmpty(t, asked)
	for _, n := range asked {
		assert.Equal(t, rowsWithoutPaging, n)
	}
}

// TestLookingAtTheTraceLeavesResultsScrolling: drawing the trace filled the
// results' row boundaries with the trace's, and paging in RESULTS snaps to
// them, so it stuck at the top.
func TestLookingAtTheTraceLeavesResultsScrolling(t *testing.T) {
	var asked []int
	m := pagingModel(t, 400, 10000, 400, &asked)
	before := append([]int(nil), m.tableRowBoundaries...)

	m.hasTrace = true
	m.traceData = [][]string{{"activity", "source"}, {"Parsing", "127.0.0.1"}, {"Executing", "127.0.0.1"}}
	m.refreshTraceView()
	assert.Equal(t, before, m.tableRowBoundaries, "the results' boundaries are the results'")

	m, _ = m.handlePageDown(tea.KeyPressMsg{Code: tea.KeyPgDown})
	m, _ = m.handlePageDown(tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Greater(t, m.tableViewport.YOffset(), 8, "paging still moves through the results")
}
