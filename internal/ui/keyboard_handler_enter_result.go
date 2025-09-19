package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// extractTableNameFromQuery extracts the table name from a SELECT query
func extractTableNameFromQuery(query string) (keyspace, table string) {
	// Simple extraction - look for FROM tablename pattern
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	fromIndex := strings.Index(upperQuery, "FROM ")
	if fromIndex == -1 {
		return "", ""
	}

	// Get the part after FROM
	afterFrom := strings.TrimSpace(query[fromIndex+5:])

	// Split by whitespace or special characters to get the table name
	parts := strings.FieldsFunc(afterFrom, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ';' || r == '('
	})

	if len(parts) > 0 {
		fullName := parts[0]
		if strings.Contains(fullName, ".") {
			// Has keyspace prefix
			tableParts := strings.Split(fullName, ".")
			if len(tableParts) == 2 {
				return tableParts[0], tableParts[1]
			}
			return "", tableParts[len(tableParts)-1]
		}
		return "", fullName
	}

	return "", ""
}

// processCommandResult processes the result from a command execution
func (m *MainModel) processCommandResult(command string, result interface{}, startTime time.Time) (*MainModel, tea.Cmd) {
	switch v := result.(type) {
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
	// Handle streaming query result
	logger.DebugfToFile("HandleEnterKey", "Got StreamingQueryResult with %d headers", len(v.Headers))
	logger.DebugfToFile("HandleEnterKey", "Headers: %v", v.Headers)
	logger.DebugfToFile("HandleEnterKey", "ColumnNames: %v", v.ColumnNames)

	// Initialize sliding window with configured memory limit
	maxMemoryMB := 10 // default
	if m.config != nil && m.config.MaxMemoryMB > 0 {
		maxMemoryMB = m.config.MaxMemoryMB
	}
	m.slidingWindow = NewSlidingWindowTable(10000, maxMemoryMB)
	m.slidingWindow.Headers = v.Headers
	m.slidingWindow.ColumnNames = v.ColumnNames
	m.slidingWindow.ColumnTypes = v.ColumnTypes

	// Load initial batch of rows
	initialRows := 0
	maxInitialRows := 100 // Show first 100 rows immediately
	
	// Create type handler once for all rows
	typeHandler := db.NewCQLTypeHandler()

	// Get column information for UDT detection
	cols := v.Iterator.Columns()
	currentKeyspace := ""
	tableName := ""
	if m.lastCommand != "" {
		// Extract keyspace and table from the query
		queryKeyspace, queryTable := extractTableNameFromQuery(m.lastCommand)
		tableName = queryTable

		// Use query keyspace if present, otherwise use session keyspace
		if queryKeyspace != "" {
			currentKeyspace = queryKeyspace
		} else if m.session != nil {
			currentKeyspace = m.session.Keyspace()
		}
	}

	logger.DebugfToFile("HandleEnterKey", "Processing streaming result: currentKeyspace=%s, tableName=%s, columns=%d",
		currentKeyspace, tableName, len(cols))
	for _, col := range cols {
		logger.DebugfToFile("HandleEnterKey", "Column: %s, Type: %v", col.Name, col.TypeInfo.Type())
	}

	// Initialize UDT registry if needed
	var udtRegistry *db.UDTRegistry
	var decoder *db.BinaryDecoder
	if m.session != nil {
		if m.session.GetUDTRegistry() == nil {
			m.session.SetUDTRegistry(db.NewUDTRegistry(m.session.Session))
		}
		udtRegistry = m.session.GetUDTRegistry()
		decoder = db.NewBinaryDecoder(udtRegistry)
	}

	for initialRows < maxInitialRows {
		// Create scan destinations - use RawBytes for UDT columns
		scanDest := make([]interface{}, len(v.ColumnNames))
		for i, col := range cols {
			if col.TypeInfo.Type() == gocql.TypeUDT {
				// Use RawBytes for UDT columns to get raw data
				scanDest[i] = new(db.RawBytes)
			} else {
				// Use regular interface{} for other types
				scanDest[i] = new(interface{})
			}
		}

		if !v.Iterator.Scan(scanDest...) {
			// Check for iterator error
			if err := v.Iterator.Close(); err != nil {
				logger.DebugfToFile("HandleEnterKey", "Iterator error: %v", err)
				// Show error to user
				m.fullHistoryContent += "\n" + m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", err))
				m.updateHistoryWrapping()
				m.historyViewport.GotoBottom()
				m.viewMode = "history"
				m.hasTable = false
				m.input.Reset()
				return m, nil
			}
			logger.DebugfToFile("HandleEnterKey", "Scan returned false after %d rows", initialRows)
			break
		}

		// Convert row to string array using original column names
		row := make([]string, len(v.ColumnNames))
		for i, colName := range v.ColumnNames {
			var val interface{}

			// Get column info
			var col *gocql.ColumnInfo
			for _, c := range cols {
				if c.Name == colName {
					col = &c
					break
				}
			}

			// Extract value based on type
			if col != nil && col.TypeInfo.Type() == gocql.TypeUDT {
				// For UDT columns, we used RawBytes
				rawBytes := scanDest[i].(*db.RawBytes)
				if *rawBytes == nil {
					val = nil
				} else {
					// We have raw bytes for the UDT
					val = []byte(*rawBytes)
				}
			} else {
				// For other columns, we used interface{}
				val = *(scanDest[i].(*interface{}))
			}

			if val == nil {
				row[i] = typeHandler.NullString
			} else {
				// Check if it's a UDT and we have raw bytes
				if col != nil && col.TypeInfo.Type() == gocql.TypeUDT {
					logger.DebugfToFile("HandleEnterKey", "Column %s is UDT, val type: %T, val: %v", colName, val, val)

					if bytes, ok := val.([]byte); ok && len(bytes) > 0 && decoder != nil {
						logger.DebugfToFile("HandleEnterKey", "Got %d bytes for UDT column %s", len(bytes), colName)

						// Get full type definition from system tables if possible
						var fullType string
						if currentKeyspace != "" && tableName != "" && m.session != nil {
							fullType = m.session.GetColumnTypeFromSystemTable(currentKeyspace, tableName, colName)
							logger.DebugfToFile("HandleEnterKey", "Full type for %s: %s", colName, fullType)
						}

						// Try to decode the UDT
						if fullType != "" {
							if typeInfo, err := db.ParseCQLType(fullType); err == nil {
								logger.DebugfToFile("HandleEnterKey", "Parsed type info for %s: %+v", colName, typeInfo)
								if decoded, err := decoder.Decode(bytes, typeInfo, currentKeyspace); err == nil {
									logger.DebugfToFile("HandleEnterKey", "Successfully decoded UDT %s: %v", colName, decoded)
									val = decoded
								} else {
									logger.DebugfToFile("HandleEnterKey", "Failed to decode UDT %s: %v", colName, err)
								}
							} else {
								logger.DebugfToFile("HandleEnterKey", "Failed to parse type for %s: %v", colName, err)
							}
						} else {
							logger.DebugfToFile("HandleEnterKey", "No full type found for %s (keyspace=%s, table=%s)", colName, currentKeyspace, tableName)
						}
					} else if m, ok := val.(map[string]interface{}); ok {
						logger.DebugfToFile("HandleEnterKey", "UDT %s came as map with %d entries", colName, len(m))
						// gocql might have already decoded it as a map
						if len(m) == 0 {
							logger.DebugfToFile("HandleEnterKey", "Empty map for UDT %s - this is the gocql issue", colName)
						}
					} else {
						logger.DebugfToFile("HandleEnterKey", "UDT %s has unexpected type: %T", colName, val)
					}
				}

				// Format the value
				if col != nil {
					row[i] = typeHandler.FormatValue(val, col.TypeInfo)
				} else {
					row[i] = fmt.Sprintf("%v", val)
				}
			}
		}

		m.slidingWindow.AddRow(row)
		initialRows++
	}

	logger.DebugfToFile("HandleEnterKey", "Loaded %d initial rows", initialRows)
	logger.DebugfToFile("HandleEnterKey", "Sliding window has %d rows", len(m.slidingWindow.Rows))

	// Check if we got any data
	if initialRows == 0 {
		// No data returned
		_ = v.Iterator.Close()
		m.fullHistoryContent += "\n" + "No results"
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		m.viewMode = "history"
		m.hasTable = false
		m.input.Reset()
		return m, nil
	}

	// Check if there's more data by trying to peek at the next row
	// Store the iterator for later use
	m.slidingWindow.iterator = v.Iterator
	m.slidingWindow.hasMoreData = true // Assume more data until proven otherwise

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
	// ASCII format - display in CQL view as text
	m.hasTable = false  // No table, just text
	m.viewMode = "history"  // Use history view for text output

	// Format initial data as ASCII table
	allData := append([][]string{headers}, m.slidingWindow.Rows...)
	
	// Format as ASCII table
	asciiStr := FormatASCIITable(allData)
	
	// Add ASCII output to history content
	if asciiStr != "" {
		m.fullHistoryContent += "\n" + asciiStr
	} else {
		m.fullHistoryContent += "\nNo results"
	}
	
	// Update with wrapped content
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()
	
	m.input.Reset()
	return m, nil
}

// displayJSONFormat displays results in JSON format
func (m *MainModel) displayJSONFormat(headers []string, columnTypes []string, columnNames []string) (*MainModel, tea.Cmd) {
	logger.DebugToFile("HandleEnterKey", "Formatting output as JSON")
	
	// JSON format - display in CQL view as text
	m.hasTable = false  // No table, just text
	m.viewMode = "history"  // Use history view for text output

	// Convert table data to JSON format
	jsonStr := ""
	// Check if we have a single [json] column from SELECT JSON
	if len(headers) == 1 && headers[0] == "[json]" {
		// This is already JSON from SELECT JSON - just extract it
		for _, row := range m.slidingWindow.Rows {
			if len(row) > 0 {
				jsonStr += row[0] + "\n"
			}
		}
	} else {
		// Convert regular table data to JSON
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
			} else {
				logger.DebugfToFile("HandleEnterKey", "Error marshaling row to JSON: %v", err)
			}
		}
	}
	
	logger.DebugfToFile("HandleEnterKey", "Generated JSON with %d characters", len(jsonStr))
	
	// Add JSON output to history content
	if jsonStr != "" {
		m.fullHistoryContent += "\n" + jsonStr
	} else {
		m.fullHistoryContent += "\nNo results"
	}
	
	// Update with wrapped content
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()
	
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