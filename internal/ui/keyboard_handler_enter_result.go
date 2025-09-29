package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)


// processCommandResult processes the result from a command execution
func (m *MainModel) processCommandResult(command string, result interface{}, startTime time.Time) (*MainModel, tea.Cmd) {
	switch v := result.(type) {
	case *router.SaveCommand:
		// Handle SAVE command - check if we have table data
		if len(m.lastTableData) == 0 {
			errorMsg := "No query results available to save. Execute a query first."
			m.fullHistoryContent += "\n" + m.styles.ErrorText.Render("Error: "+errorMsg)
			m.updateHistoryWrapping()
			m.historyViewport.GotoBottom()
			m.input.Reset()
			return m, nil
		}

		// Check if interactive mode
		if v.Interactive {
			// Open save modal
			m.saveModalActive = true
			m.saveModalStep = 0
			m.saveModalFormat = 0
			m.saveModalFilename = ""
			m.input.Reset()
			return m, nil
		}

		// Direct save - execute the save command
		// Check if data is already in JSON format (from OUTPUT JSON mode)
		if v.Format == "JSON" && len(m.lastTableData) > 0 &&
			len(m.lastTableData[0]) == 1 && m.lastTableData[0][0] == "[json]" {
			if v.Options == nil {
				v.Options = make(map[string]interface{})
			}
			v.Options["already_json"] = true
		}
		err := router.HandleSaveCommand(*v, m.lastTableData)
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
	maxInitialRows := 100 // Show first 100 rows immediately

	// Load initial rows from the streaming processor
	initialRows, hasMore, err := streamingResult.LoadMore(ctx, maxInitialRows)
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
	for _, row := range initialRows {
		m.slidingWindow.AddRow(row)
	}

	logger.DebugfToFile("HandleEnterKey", "Loaded %d initial rows", len(initialRows))
	logger.DebugfToFile("HandleEnterKey", "Sliding window has %d rows", len(m.slidingWindow.Rows))

	// Check if we got any data
	if len(initialRows) == 0 {
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
	m.slidingWindow.hasMoreData = hasMore

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
	m.topBar.RowCount = int(m.slidingWindow.TotalRowsSeen)
	m.rowCount = int(m.slidingWindow.TotalRowsSeen)

	logger.DebugfToFile("HandleEnterKey", "TopBar.RowCount set to %d", m.topBar.RowCount)

	// Prepare display based on format
	outputFormat := config.OutputFormatTable
	if m.sessionManager != nil {
		outputFormat = m.sessionManager.GetOutputFormat()
		logger.DebugfToFile("HandleEnterKey", "Got output format from session manager: %v", outputFormat)
	}
	logger.DebugfToFile("HandleEnterKey", "Using output format: %v", outputFormat)
	
	switch outputFormat {
	case config.OutputFormatExpand:
		return m.displayExpandFormat(v.Headers, v.ColumnTypes)
	case config.OutputFormatASCII:
		return m.displayASCIIFormat(v.Headers, v.ColumnTypes)
	case config.OutputFormatJSON:
		return m.displayJSONFormat(v.Headers, v.ColumnTypes, v.ColumnNames)
	default:
		return m.displayTableFormat(v.Headers, v.ColumnTypes)
	}
}

// displayExpandFormat displays results in expanded vertical format
func (m *MainModel) displayExpandFormat(headers []string, columnTypes []string) (*MainModel, tea.Cmd) {
	// EXPAND format - use table viewport for pagination support
	m.tableHeaders = headers
	m.columnTypes = columnTypes
	m.hasTable = true
	m.viewMode = "table"
	m.initialColumnWidths = nil // Reset initial widths for new table
	m.cachedTableLines = nil // Clear cache for new table

	// Format initial data as expanded vertical format
	allData := append([][]string{headers}, m.slidingWindow.Rows...)
	m.lastTableData = allData // Store for pagination
	m.horizontalOffset = 0    // Reset horizontal scroll

	// Format as expanded vertical table
	expandStr := FormatExpandTable(allData, m.styles)
	m.tableViewport.SetContent(expandStr)
	m.tableViewport.GotoTop()
	
	m.input.Reset()
	return m, nil
}

// displayASCIIFormat displays results in ASCII table format
func (m *MainModel) displayASCIIFormat(headers []string, columnTypes []string) (*MainModel, tea.Cmd) {
	logger.DebugToFile("HandleEnterKey", "Formatting output as ASCII")

	// Store headers and types for table view
	m.tableHeaders = headers
	m.columnTypes = columnTypes

	// Check if AutoFetch is ON
	if m.session != nil && m.session.AutoFetch() && m.slidingWindow.hasMoreData && m.slidingWindow.streamingResult != nil {
		logger.DebugToFile("HandleEnterKey", "ASCII format with AutoFetch ON: fetching all remaining pages")

		// Keep loading until we have all data
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
			logger.DebugfToFile("HandleEnterKey", "ASCII format: fetched %d more rows, total fetched: %d",
				loadedRows, totalFetched)
		}
		logger.DebugfToFile("HandleEnterKey", "ASCII format: finished fetching, total rows: %d",
			m.slidingWindow.TotalRowsSeen)

		// Update row count in top bar
		m.topBar.RowCount = int(m.slidingWindow.TotalRowsSeen)
		m.rowCount = int(m.slidingWindow.TotalRowsSeen)

		// When AutoFetch is ON, display all ASCII in history view
		m.hasTable = false
		m.viewMode = "history"

		// Format all data as ASCII table
		allData := append([][]string{headers}, m.slidingWindow.Rows...)
		asciiStr := FormatASCIITable(allData)

		// Add to history
		if asciiStr != "" {
			m.fullHistoryContent += "\n" + asciiStr
		} else {
			m.fullHistoryContent += "\nNo results"
		}
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
	} else {
		// AutoFetch is OFF - use table view for progressive loading
		logger.DebugfToFile("HandleEnterKey", "ASCII format with AutoFetch OFF: using table view for progressive loading")

		m.hasTable = true
		m.viewMode = "table"
		m.horizontalOffset = 0
		m.initialColumnWidths = nil // Reset initial widths for new table
		m.cachedTableLines = nil // Clear cache for new table

		// Build initial ASCII display
		allData := append([][]string{headers}, m.slidingWindow.Rows...)
		m.lastTableData = allData

		// Format as ASCII table
		asciiStr := FormatASCIITable(allData)

		// Add notice about more data if applicable
		if m.slidingWindow.hasMoreData {
			asciiStr += "\n" + m.styles.MutedText.Render(
				fmt.Sprintf("(Showing %d rows. More data available. Use PgDn/Space to load more, or AUTOFETCH ON to fetch all.)",
					len(m.slidingWindow.Rows)))
		}

		// Set content in table viewport
		m.tableViewport.SetContent(asciiStr)
		m.tableViewport.GotoTop()

		// Calculate width for horizontal scrolling
		lines := strings.Split(asciiStr, "\n")
		maxWidth := 0
		for _, line := range lines {
			if width := len(stripAnsi(line)); width > maxWidth {
				maxWidth = width
			}
		}
		m.tableWidth = maxWidth
	}

	m.input.Reset()
	return m, nil
}

// displayJSONFormat displays results in JSON format
func (m *MainModel) displayJSONFormat(headers []string, columnTypes []string, columnNames []string) (*MainModel, tea.Cmd) {
	logger.DebugToFile("HandleEnterKey", "Formatting output as JSON")

	// Store headers and types for table view
	m.tableHeaders = headers
	m.columnTypes = columnTypes

	// Check if AutoFetch is ON
	if m.session != nil && m.session.AutoFetch() && m.slidingWindow.hasMoreData && m.slidingWindow.streamingResult != nil {
		logger.DebugToFile("HandleEnterKey", "JSON format with AutoFetch ON: fetching all remaining pages")

		// Keep loading until we have all data
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
			logger.DebugfToFile("HandleEnterKey", "JSON format: fetched %d more rows, total fetched: %d",
				loadedRows, totalFetched)
		}
		logger.DebugfToFile("HandleEnterKey", "JSON format: finished fetching, total rows: %d",
			m.slidingWindow.TotalRowsSeen)

		// Update row count in top bar
		m.topBar.RowCount = int(m.slidingWindow.TotalRowsSeen)
		m.rowCount = int(m.slidingWindow.TotalRowsSeen)

		// When AutoFetch is ON, display all JSON in history view
		m.hasTable = false
		m.viewMode = "history"

		// Convert all data to JSON
		jsonStr := ""
		if len(headers) == 1 && headers[0] == "[json]" {
			for _, row := range m.slidingWindow.Rows {
				if len(row) > 0 {
					jsonStr += row[0] + "\n"
				}
			}
		} else {
			for _, row := range m.slidingWindow.Rows {
				jsonMap := make(map[string]interface{})
				for i, header := range headers {
					if i < len(row) {
						jsonMap[header] = row[i]
					}
				}
				jsonBytes, err := json.Marshal(jsonMap)
				if err == nil {
					jsonStr += string(jsonBytes) + "\n"
				}
			}
		}

		// Add to history
		if jsonStr != "" {
			m.fullHistoryContent += "\n" + jsonStr
		} else {
			m.fullHistoryContent += "\nNo results"
		}
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
	} else {
		// AutoFetch is OFF - use table view for progressive loading
		logger.DebugfToFile("HandleEnterKey", "JSON format with AutoFetch OFF: using table view for progressive loading")

		m.hasTable = true
		m.viewMode = "table"
		m.horizontalOffset = 0
		m.initialColumnWidths = nil // Reset initial widths for new table
		m.cachedTableLines = nil // Clear cache for new table

		// Build initial JSON display
		allData := append([][]string{headers}, m.slidingWindow.Rows...)
		m.lastTableData = allData

		// Convert to JSON format
		jsonStr := ""
		if len(headers) == 1 && headers[0] == "[json]" {
			for _, row := range m.slidingWindow.Rows {
				if len(row) > 0 {
					jsonStr += row[0] + "\n"
				}
			}
		} else {
			for _, row := range m.slidingWindow.Rows {
				jsonMap := make(map[string]interface{})
				for i, header := range headers {
					if i < len(row) {
						jsonMap[header] = row[i]
					}
				}
				jsonBytes, err := json.Marshal(jsonMap)
				if err == nil {
					jsonStr += string(jsonBytes) + "\n"
				}
			}
		}

		// Add notice about more data if applicable
		if m.slidingWindow.hasMoreData {
			jsonStr += "\n" + m.styles.MutedText.Render(
				fmt.Sprintf("(Showing %d rows. More data available. Use PgDn/Space to load more, or AUTOFETCH ON to fetch all.)",
					len(m.slidingWindow.Rows)))
		}

		// Set content in table viewport
		m.tableViewport.SetContent(jsonStr)
		m.tableViewport.GotoTop()

		// Calculate width for horizontal scrolling
		lines := strings.Split(jsonStr, "\n")
		maxWidth := 0
		for _, line := range lines {
			if width := len(stripAnsi(line)); width > maxWidth {
				maxWidth = width
			}
		}
		m.tableWidth = maxWidth
	}

	m.input.Reset()
	return m, nil
}

// displayTableFormat displays results in table format
func (m *MainModel) displayTableFormat(headers []string, columnTypes []string) (*MainModel, tea.Cmd) {
	// TABLE format - use table viewport
	m.tableHeaders = headers
	m.columnTypes = columnTypes
	m.hasTable = true
	m.viewMode = "table"
	m.initialColumnWidths = nil // Reset initial widths for new table
	m.cachedTableLines = nil // Clear cache for new table

	// Format initial data for display
	allData := append([][]string{headers}, m.slidingWindow.Rows...)
	m.lastTableData = allData // Store for horizontal scrolling
	m.horizontalOffset = 0    // Reset horizontal scroll
	logger.DebugfToFile("HandleEnterKey", "Formatting table with %d rows (including header)", len(allData))
	tableStr := m.formatTableForViewport(allData)
	logger.DebugfToFile("HandleEnterKey", "Table string length: %d", len(tableStr))
	m.tableViewport.SetContent(tableStr)
	m.tableViewport.GotoTop()
	logger.DebugfToFile("HandleEnterKey", "Table viewport content set, viewMode: %s", m.viewMode)
	
	m.input.Reset()
	return m, nil
}

// processQueryResult handles QueryResult type
func (m *MainModel) processQueryResult(command string, v db.QueryResult) (*MainModel, tea.Cmd) {
	// Query result with metadata
	if len(v.Data) > 0 {
		// Update top bar with query metadata
		m.topBar.QueryTime = v.Duration
		m.topBar.RowCount = v.RowCount
		m.topBar.HasQueryData = true

		m.rowCount = v.RowCount

		// Get output format from session manager
		outputFormat := config.OutputFormatTable
		if m.sessionManager != nil {
			outputFormat = m.sessionManager.GetOutputFormat()
		}
		logger.DebugfToFile("keyboard_handler_enter", "QueryResult Format: %v", outputFormat)

		// Check output format
		switch outputFormat {
		case config.OutputFormatASCII:
			// ASCII format - display in CQL view as text
			m.hasTable = false  // No table, just text
			m.viewMode = "history"  // Use history view for text output

			// Format as ASCII table
			asciiOutput := FormatASCIITable(v.Data)
			
			// Add ASCII output to history content
			if asciiOutput != "" {
				m.fullHistoryContent += "\n" + asciiOutput
			} else {
				m.fullHistoryContent += "\nNo results"
			}
			
			// Update with wrapped content
			m.updateHistoryWrapping()
			m.historyViewport.GotoBottom()
		case config.OutputFormatExpand:
			// EXPAND format - use table viewport for scrolling support
			// Store table data and headers
			m.lastTableData = v.Data
			m.tableHeaders = v.Data[0]    // Store the header row
			m.columnTypes = v.ColumnTypes // Store column types
			m.horizontalOffset = 0
			m.hasTable = true
			m.viewMode = "table"
			m.initialColumnWidths = nil // Reset initial widths for new table
			m.cachedTableLines = nil // Clear cache for new table

			// Format as expanded vertical table
			expandOutput := FormatExpandTable(v.Data, m.styles)
			m.tableViewport.SetContent(expandOutput)
			m.tableViewport.GotoTop() // Start at top of table
		case config.OutputFormatJSON:
			// JSON format - display in CQL view as text
			m.hasTable = false  // No table, just text
			m.viewMode = "history"  // Use history view for text output

			// Check if this is already JSON from SELECT JSON
			jsonOutput := ""
			if len(v.Data) > 1 {
				headers := v.Data[0]
				// Check if we have a single [json] column from SELECT JSON
				if len(headers) == 1 && headers[0] == "[json]" {
					// This is already JSON from SELECT JSON - just extract it
					for _, row := range v.Data[1:] {
						if len(row) > 0 {
							jsonOutput += row[0] + "\n"
						}
					}
				} else {
					// Convert regular table data to JSON
					for _, row := range v.Data[1:] {
						jsonMap := make(map[string]interface{})
						for i, header := range headers {
							if i < len(row) {
								jsonMap[header] = row[i]
							}
						}
						jsonBytes, err := json.Marshal(jsonMap)
						if err == nil {
							jsonOutput += string(jsonBytes) + "\n"
						}
					}
				}
			}
			
			// Add JSON output to history content
			if jsonOutput != "" {
				m.fullHistoryContent += "\n" + jsonOutput
			} else {
				m.fullHistoryContent += "\nNo results"
			}
			
			// Update with wrapped content
			m.updateHistoryWrapping()
			m.historyViewport.GotoBottom()
		default:
			// Use table viewport for TABLE format
			// Store table data and headers
			m.lastTableData = v.Data
			m.tableHeaders = v.Data[0]    // Store the header row
			m.columnTypes = v.ColumnTypes // Store column types
			m.horizontalOffset = 0
			m.hasTable = true
			m.viewMode = "table"
			m.initialColumnWidths = nil // Reset initial widths for new table
			m.cachedTableLines = nil // Clear cache for new table

			// Format and display in table viewport
			tableStr := m.formatTableForViewport(v.Data)
			m.tableViewport.SetContent(tableStr)
			m.tableViewport.GotoTop() // Start at top of table
		}

		// Write to capture file if capturing
		metaHandler := router.GetMetaHandler()
		if metaHandler != nil && metaHandler.IsCapturing() && len(v.Data) > 1 {
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
		// Store table data and headers
		m.lastTableData = v
		m.tableHeaders = v[0] // Store the header row
		m.horizontalOffset = 0
		m.hasTable = true
		m.viewMode = "table"
		m.initialColumnWidths = nil // Reset initial widths for new table
		m.cachedTableLines = nil // Clear cache for new table

		// Format and display in table viewport
		tableStr := m.formatTableForViewport(v)
		m.tableViewport.SetContent(tableStr)
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

// processStringResult handles string type results
func (m *MainModel) processStringResult(command string, v string) (*MainModel, tea.Cmd) {
	// Check if this is a USE command result
	if strings.HasPrefix(v, "Now using keyspace ") {
		// Extract keyspace name and update session manager
		keyspace := strings.TrimPrefix(v, "Now using keyspace ")
		keyspace = strings.TrimSpace(keyspace)
		if m.sessionManager != nil {
			m.sessionManager.SetKeyspace(keyspace)
			// Update the status bar
			m.statusBar.Keyspace = keyspace
		}
		// Update the database session's keyspace
		if m.session != nil {
			if err := m.session.SetKeyspace(keyspace); err != nil {
				// Log error but don't fail - the keyspace change was already successful on the server
				logger.DebugfToFile("keyboard_handler_enter", "Failed to update session keyspace: %v", err)
			}
		}
	}
	// Text result - add to history
	m.tableHeaders = nil
	m.columnWidths = nil
	m.hasTable = false
	m.viewMode = "history"
	// Clear query metadata from top bar
	m.topBar.HasQueryData = false
	// Wrap long lines to prevent truncation
	wrappedResult := wrapLongLines(v, m.historyViewport.Width)
	
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