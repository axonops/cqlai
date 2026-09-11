package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
)

// runSaveCommand carries out a parsed SAVE.
//
// Both routes to a save come through here: typing the command, and the Save
// window, which builds the command typing it would. The window went to
// runCommand, which prints a string result and drops anything else - and SAVE
// returns a parsed command rather than a string, so picking a format and a path
// wrote no file, showed no error and said nothing at all. It had been that way
// for every format since the window was shared in #129.
func (m *MainModel) runSaveCommand(v *router.SaveCommand) (*MainModel, tea.Cmd) {
	if !m.hasResults() {
		m.fullHistoryContent += "\n" + m.styles.ErrorText.Render("Error: "+noResultsToSave)
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		m.input.Reset()
		return m, nil
	}

	// SAVE with no arguments asks which format and where, in the window
	// Capture uses. Centred, because it was typed rather than clicked.
	if v.Interactive {
		m.input.Reset()
		return m.openSavePanel()
	}

	// Check if data is already in JSON format (from OUTPUT JSON mode)
	if v.Format == "JSON" && len(m.lastTableData[0]) == 1 && m.lastTableData[0][0] == "[json]" {
		if v.Options == nil {
			v.Options = make(map[string]interface{})
		}
		v.Options["already_json"] = true
	}

	err := router.HandleSaveCommand(*v, m.lastTableData, m.columnTypes)
	if err != nil {
		m.fullHistoryContent += "\n" + m.styles.ErrorText.Render("Error: "+err.Error())
	} else {
		rowCount := len(m.lastTableData) - 1 // Exclude header
		successMsg := fmt.Sprintf("Successfully saved %d rows to %s", rowCount, v.Filename)
		m.fullHistoryContent += "\n" + m.styles.SuccessText.Render(successMsg)
	}
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()
	m.input.Reset()
	return m, nil
}

// processCommandResult processes the result from a command execution
func (m *MainModel) processCommandResult(command string, result interface{}, startTime time.Time) (*MainModel, tea.Cmd) {
	switch v := result.(type) {
	case *router.SaveCommand:
		return m.runSaveCommand(v)
	case db.StreamingQueryResult:
		return m.processStreamingQueryResult(command, v, startTime)
	case db.QueryResult:
		return m.processQueryResult(command, v)
	case [][]string:
		return m.processTableResult(command, v)
	case string:
		return m.processStringResult(command, v)
	case error:
		return m.processErrorResult(v)
	}
	return m, nil
}

// rowsWithoutPaging is the batch to fetch when PAGING is off and there is no
// page size to take it from.
const rowsWithoutPaging = 100

// initialRowsToLoad is how many rows to fetch before drawing anything: one
// page, which is what PAGING sets.
//
// It was a hardcoded 100, so PAGING 500 still showed 100 first and PAGING 50
// still showed 100. It matching the default was a coincidence.
func (m *MainModel) initialRowsToLoad() int {
	if m.session != nil {
		if pageSize := m.session.PageSize(); pageSize > 0 {
			return pageSize
		}
	}
	return rowsWithoutPaging
}

// processStreamingQueryResult handles streaming query results
func (m *MainModel) processStreamingQueryResult(command string, v db.StreamingQueryResult, startTime time.Time) (*MainModel, tea.Cmd) {
	// Use the new StreamingProcessor from the database layer
	streamingResult := m.session.ProcessStreamingQuery(v)
	// Note: Don't defer Close() here since we're storing the streamingResult for progressive loading

	logger.DebugfToFile("HandleEnterKey", "Got StreamingQueryResult with %d headers", len(streamingResult.Headers))
	logger.DebugfToFile("HandleEnterKey", "Headers: %v", streamingResult.Headers)

	// Initialize sliding window with configured memory limit
	maxMemoryMB := 10 // default
	if m.config != nil && m.config.MaxMemoryMB > 0 {
		maxMemoryMB = m.config.MaxMemoryMB
	}
	m.slidingWindow = NewSlidingWindowTable(10000, maxMemoryMB)
	m.slidingWindow.Headers = streamingResult.Headers
	m.slidingWindow.ColumnNames = v.ColumnNames
	m.slidingWindow.ColumnTypes = v.ColumnTypes

	// Load initial batch of rows using the streaming processor
	ctx := context.Background()
	maxInitialRows := m.initialRowsToLoad()

	// Load initial rows from the streaming processor
	page, err := streamingResult.LoadMore(ctx, maxInitialRows)
	if err != nil {
		logger.DebugfToFile("HandleEnterKey", "Error loading initial rows: %v", err)
		m.fullHistoryContent += "\n" + m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", err))
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		m.viewMode = "history"
		m.hasTable = false
		m.input.Reset()
		return m, nil
	}

	// Add rows to sliding window
	for i, row := range page.Rows {
		m.slidingWindow.AddRow(row, rawAt(page.Raw, i))
	}

	logger.DebugfToFile("HandleEnterKey", "Loaded %d initial rows", len(page.Rows))
	logger.DebugfToFile("HandleEnterKey", "Sliding window has %d rows", len(m.slidingWindow.Rows))

	// Check if we got any data
	if len(page.Rows) == 0 {
		// No data returned
		m.fullHistoryContent += "\n" + "No results"
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		m.viewMode = "history"
		m.hasTable = false
		m.input.Reset()
		return m, nil
	}

	// Store the streaming result in sliding window for progressive loading
	m.slidingWindow.streamingResult = streamingResult
	m.slidingWindow.hasMoreData = page.HasMore

	// If auto-fetch is enabled, fetch all remaining pages immediately
	if m.session != nil && m.session.AutoFetch() && m.slidingWindow.hasMoreData {
		logger.DebugToFile("HandleEnterKey", "Auto-fetch enabled, loading all remaining rows")

		// Load all remaining rows
		for m.slidingWindow.hasMoreData {
			loadedRows := m.slidingWindow.LoadMoreRows(10000) // Load in large batches
			if loadedRows == 0 {
				break
			}
			logger.DebugfToFile("HandleEnterKey", "Auto-fetched %d more rows, total: %d",
				loadedRows, m.slidingWindow.TotalRowsSeen)
		}
		logger.DebugfToFile("HandleEnterKey", "Auto-fetch complete, total rows: %d",
			m.slidingWindow.TotalRowsSeen)
	}

	// Write initial rows to capture file if capturing
	metaHandler := router.GetMetaHandler()
	if metaHandler != nil && metaHandler.IsCapturing() && len(m.slidingWindow.Rows) > 0 {
		// If we have column types (for Parquet support), use the new method
		if len(v.ColumnTypes) > 0 {
			_ = metaHandler.WriteCaptureResultWithTypes(command, v.Headers, v.ColumnTypes, m.slidingWindow.Rows, nil)
		} else {
			_ = metaHandler.WriteCaptureResult(command, v.Headers, m.slidingWindow.Rows)
		}
		m.slidingWindow.MarkRowsAsCaptured(len(m.slidingWindow.Rows))
	}

	// Update UI
	m.topBar.HasQueryData = true
	m.topBar.QueryTime = time.Since(v.StartTime)
	m.rowCount = int(m.slidingWindow.TotalRowsSeen)

	logger.DebugfToFile("HandleEnterKey", "Row count set to %d", m.rowCount)

	return m.displayResult(v.Headers, v.ColumnTypes)
}

// displayResult puts a streaming result on screen, in the format OUTPUT is set
// to.
//
// There were four of these, one per format, and between them they were where
// the format got written down a third time: each set m.resultFormat itself,
// and which one ran was decided by a switch somewhere else. They differed in
// two lines - whether to pull the rest of the rows in first - so that is all
// that is left of them.
func (m *MainModel) displayResult(headers, columnTypes []string) (*MainModel, tea.Cmd) {
	m.tableHeaders = headers
	m.columnTypes = columnTypes
	m.hasTable = true
	m.viewMode = "table"
	m.initialColumnWidths = nil // Reset initial widths for new table
	m.cachedTableLines = nil    // Clear cache for new table
	m.resetHorizontalScroll()

	format := m.outputFormat()
	if onePassFormat(format) {
		// AutoFetch only decides how much is pulled in; the result goes to the
		// Results view either way, in every format.
		m.fetchAllPages(string(format))
	}

	logger.DebugfToFile("HandleEnterKey", "Showing %d rows as %v", len(m.slidingWindow.Rows), format)
	m.renderResults(m.resultRows())
	m.tableViewport.GotoTop()

	m.input.Reset()
	return m, nil
}

// showQueryResult draws a result that arrived complete, in whatever format
// OUTPUT is set to.
func (m *MainModel) showQueryResult(data [][]string, columnTypes []string, raw ...[]map[string]interface{}) {
	if len(data) == 0 {
		return
	}

	// Every format goes to the Results view. ASCII and JSON used to be
	// written into the console as text instead, where it is wrapped to the
	// window - which breaks an ASCII table's borders mid-row and gives up
	// the sideways scrolling that wide output needs.
	m.tableHeaders = data[0]
	m.columnTypes = columnTypes
	m.lastRawData = nil
	if len(raw) > 0 {
		m.lastRawData = raw[0]
	}
	m.resetHorizontalScroll()
	m.hasTable = true
	m.viewMode = "table"
	m.initialColumnWidths = nil // Reset initial widths for new table
	m.cachedTableLines = nil    // Clear cache for new table

	// This result arrived whole. Any sliding window belongs to an earlier
	// streaming one, and left in place it would answer for this result: how
	// many rows are showing, whether there are more, and which rows to draw
	// when the format changes.
	m.slidingWindow = nil

	m.renderResults(data)
	m.tableViewport.GotoTop()
}

// jsonLines renders rows as one JSON object per line.
//
// A SELECT JSON already comes back as JSON in a single "[json]" column, so that
// is passed through rather than wrapped in an object a second time.
func jsonLines(headers []string, rows [][]string, raw []map[string]interface{}) string {
	var b strings.Builder

	if len(headers) == 1 && headers[0] == "[json]" {
		for _, row := range rows {
			if len(row) > 0 {
				b.WriteString(row[0] + "\n")
			}
		}
		return b.String()
	}

	for i, row := range rows {
		b.WriteByte('{')
		for c, header := range headers {
			if c > 0 {
				b.WriteByte(',')
			}
			name := db.StripKeyMarker(header)
			key, _ := json.Marshal(name)
			b.Write(key)
			b.WriteByte(':')
			b.Write(jsonCell(name, row, c, rawRow(raw, i)))
		}
		b.WriteString("}\n")
	}
	return b.String()
}

// jsonCell is one value, written from what it was rather than from how it is
// drawn where that is still in hand.
//
// The string in the cell is a fallback: it is what a grid the shell made up has
// instead of values, and what is left of a row whose values were dropped.
func jsonCell(name string, row []string, i int, raw map[string]interface{}) []byte {
	if value, ok := raw[name]; ok {
		if encoded, err := json.Marshal(db.JSONValue(value)); err == nil {
			return encoded
		}
	}

	cell := ""
	if i < len(row) {
		cell = row[i]
	}
	encoded, err := json.Marshal(cell)
	if err != nil {
		return []byte(`""`)
	}
	return encoded
}

// rawRow is the values behind a row, or nil when there are none.
func rawRow(raw []map[string]interface{}, i int) map[string]interface{} {
	if i < len(raw) {
		return raw[i]
	}
	return nil
}

// formatRowsAsJSON renders a result whose first row is its headers.
func formatRowsAsJSON(data [][]string, raw []map[string]interface{}) string {
	if len(data) < 2 {
		return ""
	}
	return jsonLines(data[0], data[1:], raw)
}

// widestLine is the width of the longest line, for horizontal scrolling.
func widestLine(s string) int {
	widest := 0
	for _, line := range strings.Split(s, "\n") {
		if w := lipgloss.Width(line); w > widest {
			widest = w
		}
	}
	return widest
}

// fetchAllPages pulls the rest of a streaming result in, when AutoFetch is on.
//
// It says nothing about where the result is then shown. AutoFetch decides how
// much is fetched; the OUTPUT format decides how it is drawn. Tangling the two
// is what sent ASCII and JSON results to the console rather than the Results
// view, wrapped to the window and with their borders broken in half.
func (m *MainModel) fetchAllPages(format string) {
	if m.session == nil || !m.session.AutoFetch() ||
		!m.slidingWindow.hasMoreData || m.slidingWindow.streamingResult == nil {
		return
	}

	logger.DebugfToFile("HandleEnterKey", "%s format with AutoFetch ON: fetching all remaining pages", format)

	totalFetched := 0
	for m.slidingWindow.hasMoreData {
		pageSize := m.session.PageSize()
		if pageSize == 0 {
			pageSize = 100 // Default page size
		}
		loadedRows := m.slidingWindow.LoadMoreRows(pageSize)
		if loadedRows == 0 {
			break
		}
		totalFetched += loadedRows
	}
	logger.DebugfToFile("HandleEnterKey", "%s format: fetched %d rows, total rows: %d",
		format, totalFetched, m.slidingWindow.TotalRowsSeen)

	m.rowCount = int(m.slidingWindow.TotalRowsSeen)
}

// processQueryResult handles QueryResult type
func (m *MainModel) processQueryResult(command string, v db.QueryResult) (*MainModel, tea.Cmd) {
	// Query result with metadata
	if len(v.Data) > 0 {
		// Update top bar with query metadata
		m.topBar.QueryTime = v.Duration
		m.topBar.HasQueryData = true

		m.rowCount = v.RowCount

		logger.DebugfToFile("keyboard_handler_enter", "QueryResult Format: %v", m.outputFormat())
		m.showQueryResult(v.Data, v.ColumnTypes, v.RawData)

		// Write to the AutoSave file, unless the router already did: a streaming
		// result is drained there and handed on as an ordinary QueryResult, and
		// writing it again put one query into two files.
		metaHandler := router.GetMetaHandler()
		if metaHandler != nil && !v.AlreadySaved && metaHandler.IsCapturing() && len(v.Data) > 1 {
			// Extract headers and rows from data
			headers := v.Data[0]
			rows := v.Data[1:]
			// If we have column types (for Parquet support), use the new method
			switch {
			case len(v.ColumnTypes) > 0:
				_ = metaHandler.WriteCaptureResultWithTypes(command, headers, v.ColumnTypes, rows, v.RawData)
			case len(v.RawData) > 0:
				// Otherwise use raw data if available for better type preservation in JSON
				_ = metaHandler.WriteCaptureResultWithRawData(command, headers, rows, v.RawData)
			default:
				_ = metaHandler.WriteCaptureResult(command, headers, rows)
			}
		}
	}

	m.input.Reset()
	return m, nil
}

// processTableResult handles [][]string type (backward compatibility)
func (m *MainModel) processTableResult(command string, v [][]string) (*MainModel, tea.Cmd) {
	// Table data without metadata (for backward compatibility)
	if len(v) > 0 {
		m.rowCount = len(v) - 1 // Exclude header
		m.topBar.HasQueryData = true
		// Store table data and headers
		m.lastTableData = v
		m.tableHeaders = v[0] // Store the header row
		// These are a grid the shell made up rather than a result over a table -
		// DESCRIBE and the listings - so there are no types to go with them.
		// The last query's would otherwise be left standing beside a result they
		// do not describe, with no guarantee even of the same number of columns:
		// the header's detail row reads them by index and PARQUET builds a
		// schema from them, so both would be quietly wrong.
		m.columnTypes = nil
		m.lastRawData = nil
		m.resetHorizontalScroll()
		m.hasTable = true
		m.viewMode = "table"
		m.initialColumnWidths = nil // Reset initial widths for new table
		m.cachedTableLines = nil    // Clear cache for new table

		// Nothing streamed this, so there is no window of rows to page: what is
		// on screen is all of it.
		m.slidingWindow = nil

		// Drawn in whatever OUTPUT is set to, like every other result. It used
		// to force the boxed table, which was a rule nothing else followed and
		// one that changing the format on screen would have broken anyway: a
		// grid drawn one way and redrawn another.
		m.renderResults(v)
		m.tableViewport.GotoTop() // Start at top of table

		// Write to capture file if capturing
		metaHandler := router.GetMetaHandler()
		if metaHandler != nil && metaHandler.IsCapturing() && len(v) > 1 {
			// Extract headers and rows from data
			headers := v[0]
			rows := v[1:]
			_ = metaHandler.WriteCaptureResult(command, headers, rows)
		}
	}

	m.input.Reset()
	return m, nil
}

// adoptKeyspace records a keyspace change the server has already accepted.
//
// The change is done by the time this runs; what is left is bookkeeping, and
// there are three places that track the current keyspace separately - the
// session manager, the gocql session, and the status line. Missing one leaves
// the bottom line claiming a keyspace the queries are not going to.
//
// It takes the result string rather than a name because that is what says the
// USE succeeded. Both routes to a USE - typing it, and picking a keyspace from
// the status line - come through here, so neither can drift from the other.
func (m *MainModel) adoptKeyspace(result string) {
	const prefix = "Now using keyspace "
	if !strings.HasPrefix(result, prefix) {
		return
	}
	keyspace := strings.TrimSpace(strings.TrimPrefix(result, prefix))

	if m.sessionManager != nil {
		if err := m.sessionManager.SetKeyspace(keyspace); err != nil {
			logger.DebugfToFile("keyboard_handler_enter", "Failed to update session manager keyspace: %v", err)
		}
		m.statusBar.Keyspace = keyspace
	}
	if m.session != nil {
		if err := m.session.SetKeyspace(keyspace); err != nil {
			// Log it but carry on: the server has already switched, so
			// refusing here would only make the two disagree.
			logger.DebugfToFile("keyboard_handler_enter", "Failed to update session keyspace: %v", err)
		}
	}
}

// processStringResult handles string type results
func (m *MainModel) processStringResult(command string, v string) (*MainModel, tea.Cmd) {
	m.adoptKeyspace(v)

	// Text result - add to history
	m.tableHeaders = nil
	m.columnWidths = nil
	m.hasTable = false
	m.viewMode = "history"
	// Clear query metadata from top bar
	m.topBar.HasQueryData = false
	// Wrap long lines to prevent truncation
	wrappedResult := wrapLongLines(v, m.historyViewport.Width())

	m.fullHistoryContent += "\n" + wrappedResult
	m.updateHistoryWrapping()

	// Write to capture file if capturing
	metaHandler := router.GetMetaHandler()
	if metaHandler != nil && metaHandler.IsCapturing() {
		_ = metaHandler.WriteCaptureText(command, v)
	}

	// Always scroll to bottom for consistent behavior
	// Users can scroll up if they need to see earlier parts
	m.historyViewport.GotoBottom()

	m.input.Reset()
	return m, nil
}

// processErrorResult handles error type results
func (m *MainModel) processErrorResult(v error) (*MainModel, tea.Cmd) {
	// Error result - add to history
	m.tableHeaders = nil
	m.columnWidths = nil
	// Clear query metadata from top bar
	m.topBar.HasQueryData = false
	m.hasTable = false
	m.viewMode = "history"
	errorMsg := m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", v))
	m.fullHistoryContent += "\n" + errorMsg
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()

	m.input.Reset()
	return m, nil
}
