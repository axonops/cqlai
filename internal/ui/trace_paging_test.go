package ui

import (
	"testing"

	"charm.land/bubbles/v2/viewport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/db"
)

// The trace of a query read a page at a time, against a cluster.
//
// This is the test the unit tests could not be: every part of it was right on
// its own - the driver traced every page, the session kept the last one, the
// view asked for it - and what you saw was the first page's trace, because the
// view had two copies of its drawing and the one that ran last was the one
// without the page number on it.
//
// It skips without a cluster, the way the Cassandra checks in the completion
// package do: a test that runs where the answer is available.

// tracingSession is a session to whatever is on the loopback address, with
// tracing on and a page size small enough to make several pages of a small
// table.
func tracingSession(t *testing.T) *db.Session {
	t.Helper()

	session, err := db.NewSessionWithOptions(db.SessionOptions{
		Host: "127.0.0.1", Port: 9042, Keyspace: "system_schema",
	})
	if err != nil {
		t.Skipf("no cluster on 127.0.0.1:9042: %v", err)
	}
	t.Cleanup(session.Close)

	session.SetTracing(true)
	session.SetPageSize(5)
	return session
}

// TestTheTraceFollowsThePageBeingRead.
func TestTheTraceFollowsThePageBeingRead(t *testing.T) {
	session := tracingSession(t)

	m := helpModel()
	m.windowWidth, m.windowHeight = 120, 30
	m.session = session
	m.historyViewport = viewport.New(viewport.WithWidth(120), viewport.WithHeight(20))
	m.traceViewport = viewport.New(viewport.WithWidth(120), viewport.WithHeight(20))
	m.trace.viewport = viewport.New(viewport.WithWidth(120), viewport.WithHeight(5))

	query := "SELECT keyspace_name, table_name, column_name FROM system_schema.columns"
	raw, ok := session.ExecuteStreamingQuery(query).(db.StreamingQueryResult)
	require.True(t, ok, "the query should be answered a page at a time")

	m.slidingWindow = NewSlidingWindowTable(1000, 10)
	m.slidingWindow.Headers = session.ProcessStreamingQuery(raw).Headers
	m.slidingWindow.streamingResult = session.ProcessStreamingQuery(raw)
	m.slidingWindow.hasMoreData = true

	m.captureTraceData(query)
	require.True(t, m.hasTrace, "the query was traced")
	first := m.traceInfo.Pages

	// Read on, a page at a time. The trace follows what was last fetched.
	pages := []int{}
	for range 3 {
		require.Positive(t, m.slidingWindow.LoadMoreRows(5), "there are more rows to read")

		m.viewMode = "results"
		updated, _ := m.showTrace()
		m = updated

		pages = append(pages, m.traceInfo.Pages)
	}

	assert.Greater(t, pages[len(pages)-1], first,
		"the trace took in the pages read since, rather than staying on the first")
	for i, page := range pages {
		assert.GreaterOrEqual(t, page, 1, "page %d", i)
	}
}

// TestTheViewSaysWhichPageTheTraceIsOf.
//
// The view is drawn in one place. It was drawn in two, and the copy that ran
// last built its own summary line without the page on it - so the trace
// changed under the reader with nothing on screen to say it had.
func TestTheViewSaysWhichPageTheTraceIsOf(t *testing.T) {
	session := tracingSession(t)

	m := helpModel()
	m.windowWidth, m.windowHeight = 120, 30
	m.session = session
	m.historyViewport = viewport.New(viewport.WithWidth(120), viewport.WithHeight(20))
	m.traceViewport = viewport.New(viewport.WithWidth(120), viewport.WithHeight(20))

	query := "SELECT keyspace_name, table_name, column_name FROM system_schema.columns"
	raw, ok := session.ExecuteStreamingQuery(query).(db.StreamingQueryResult)
	require.True(t, ok)

	m.slidingWindow = NewSlidingWindowTable(1000, 10)
	m.slidingWindow.streamingResult = session.ProcessStreamingQuery(raw)
	m.slidingWindow.hasMoreData = true

	m.captureTraceData(query)
	for range 3 {
		m.slidingWindow.LoadMoreRows(5)
	}
	m.fetchTrace()
	require.Greater(t, m.traceInfo.Pages, 1, "more than one page has been fetched")

	drawn := stripAnsi(m.traceViewport.GetContent())
	assert.Contains(t, drawn, "Trace Session")
	assert.Contains(t, drawn, tracePagesSaid(m.traceInfo))

	// And the drawing is the same whichever way it is asked for.
	m.refreshTraceView()
	assert.Equal(t, drawn, stripAnsi(m.traceViewport.GetContent()))
}
