package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/db"
)

// captureTraceData captures trace data if tracing is enabled
func (m *MainModel) captureTraceData(command string) {
	upperCmd := strings.ToUpper(strings.TrimSpace(command))
	if m.session != nil && m.session.Tracing() &&
		(strings.HasPrefix(upperCmd, "SELECT") ||
			strings.HasPrefix(upperCmd, "LIST") ||
			strings.HasPrefix(upperCmd, "DESCRIBE") ||
			strings.HasPrefix(upperCmd, "DESC")) {
		// Give Cassandra a moment to write trace data
		time.Sleep(50 * time.Millisecond)
		m.fetchTrace()
	}
}

// fetchTrace reads the trace of the request the shell last made, and puts it
// in the view.
//
// A paged query is one traced request per page, so which trace this is depends
// on when it is asked for: after the query it is the first page's, and after
// paging through the results it is the page last fetched. The summary line
// says which.
//
// The drawing is refreshTraceView's. This held a copy of it - the same swap of
// the table renderer's state, the same summary line without the page on it -
// and the copy that ran last was the one without, so the view updated on every
// page and said nothing to show for it.
func (m *MainModel) fetchTrace() {
	if m.session == nil || !m.session.Tracing() {
		return
	}

	traceData, traceHeaders, traceInfo, err := m.session.GetTraceData()
	if err != nil || len(traceData) == 0 {
		return
	}

	// The headers are a row of the table the renderer draws.
	full := make([][]string, 0, len(traceData)+1)
	full = append(full, traceHeaders)
	full = append(full, traceData...)

	m.traceData = full
	m.traceHeaders = traceHeaders
	m.traceInfo = traceInfo
	m.hasTrace = true
	m.traceHorizontalOffset = 0

	m.refreshTraceView()
	m.traceViewport.GotoTop()
}

// tracePagesSaid says how much of the query the trace covers, where it is more
// than one request.
//
// A paged query is traced per page, and the total above it is the sum of them:
// without this the view says a scan took nine milliseconds without saying that
// was six requests, or that there were forty more it is not showing.
func tracePagesSaid(info *db.TraceInfo) string {
	switch {
	case info == nil || info.Pages <= 1:
		return ""
	case info.Shown < info.Pages:
		return fmt.Sprintf(" | %d pages, the last %d shown", info.Pages, info.Shown)
	}
	return fmt.Sprintf(" | %d pages", info.Pages)
}
