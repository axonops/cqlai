package ui

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/viewport"
	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rowCountModel(t *testing.T) *MainModel {
	t.Helper()

	return &MainModel{
		styles:          DefaultStyles(),
		windowWidth:     100,
		windowHeight:    30,
		viewMode:        "history",
		input:           newTestInput(),
		historyViewport: viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
		tableViewport:   viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}
}

// shownRows is what the info bar says, after View has synced it.
func shownRows(m *MainModel) string {
	m.topBar.RowCount = m.rowCount // what View does
	for _, seg := range m.topBar.segments() {
		if seg.label == "Rows: " {
			return seg.value
		}
	}
	return ""
}

// TestASelectReportsItsRowCount. The count used to be incremented in the dead
// half of an "if true ... else", so every non-streaming SELECT said 0 however
// many rows it returned.
func TestASelectReportsItsRowCount(t *testing.T) {
	m := rowCountModel(t)

	m.processQueryResult("SELECT * FROM t", db.QueryResult{
		Data:     [][]string{{"id"}, {"1"}, {"2"}, {"3"}},
		RowCount: 3,
		Duration: 12 * time.Millisecond,
		Headers:  []string{"id"},
	})

	assert.Equal(t, 3, m.rowCount)
	assert.Equal(t, "3", shownRows(m))
}

// TestDescribeReportsItsRowCount: it counted its rows and never told the bar,
// so Rows read "-" after a DESCRIBE.
func TestDescribeReportsItsRowCount(t *testing.T) {
	m := rowCountModel(t)

	m.processTableResult("DESCRIBE TABLES", [][]string{{"name"}, {"users"}, {"events"}, {"logs"}})

	assert.Equal(t, 3, m.rowCount)
	assert.Equal(t, "3", shownRows(m))
}

// TestNoQueryYetShowsADash rather than a zero that looks like a real answer.
func TestNoQueryYetShowsADash(t *testing.T) {
	m := rowCountModel(t)

	assert.Equal(t, infoPlaceholder, shownRows(m))
}

// TestAQueryReturningNothingShowsZero, which is a real answer.
func TestAQueryReturningNothingShowsZero(t *testing.T) {
	m := rowCountModel(t)

	m.processQueryResult("SELECT * FROM t", db.QueryResult{
		Data:     [][]string{{"id"}},
		RowCount: 0,
		Duration: time.Millisecond,
		Headers:  []string{"id"},
	})

	assert.Equal(t, "0", shownRows(m))
}

// TestTheBarFollowsTheModel: the count is derived in View rather than pushed in
// by each handler, which is how one of six paths came to forget.
func TestTheBarFollowsTheModel(t *testing.T) {
	m := rowCountModel(t)
	m.topBar.HasQueryData = true

	for _, n := range []int{0, 1, 42, 1000} {
		m.rowCount = n
		require.Equal(t, n, m.rowCount)
		assert.Equal(t, itoa(n), shownRows(m))
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

// TestMoreToFetchIsMarked: the "+" says the number is what has been fetched so
// far, not what the query would return.
func TestMoreToFetchIsMarked(t *testing.T) {
	m := rowCountModel(t)
	m.topBar.HasQueryData = true
	m.rowCount = 100
	m.topBar.HasMoreData = true

	assert.Equal(t, "100+", shownRows(m))

	m.topBar.AutoFetch = true
	assert.Equal(t, "100", shownRows(m), "AutoFetch means there is nothing left to fetch")
}

// TestTheFirstBatchIsOnePage. It was a hardcoded 100, so PAGING 500 still
// showed 100 first and PAGING 50 still showed 100.
func TestTheFirstBatchIsOnePage(t *testing.T) {
	for _, pageSize := range []int{50, 100, 500} {
		m := rowCountModel(t)
		m.session = &db.Session{}
		m.session.SetPageSize(pageSize)

		assert.Equal(t, pageSize, m.initialRowsToLoad())
	}
}

// TestWithPagingOffTheBatchFallsBack, since there is no page size to take.
func TestWithPagingOffTheBatchFallsBack(t *testing.T) {
	m := rowCountModel(t)
	m.session = &db.Session{}
	m.session.SetPageSize(0)

	assert.Equal(t, rowsWithoutPaging, m.initialRowsToLoad())

	m.session = nil
	assert.Equal(t, rowsWithoutPaging, m.initialRowsToLoad())
}
