package router

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parquet"
	"github.com/axonops/cqlai/internal/session"
)

// MetaCommandHandler handles non-CQL meta commands
type MetaCommandHandler struct {
	session                  *db.Session
	sessionManager           *session.Manager
	expandMode               bool
	captureFile              string
	captureOutput            io.WriteCloser
	captureFormat            string // "text", "json", "csv", or "parquet"
	csvWriter                *csv.Writer
	parquetWriter            *parquet.ParquetCaptureWriter
	captureHeaders           []string // Store headers for parquet writer
	partitionedWriter        *parquet.PartitionedParquetWriter // For partitioned Parquet capture
	captureOptions           map[string]string // Capture options (compression, partition, etc.)
	capturePartitionColumns  []string // Partition columns for capture
	captureColumnTypes       []string // Column types for partitioned capture
}

// NewMetaCommandHandler creates a new meta command handler
func NewMetaCommandHandler(session *db.Session, sessionMgr *session.Manager) *MetaCommandHandler {
	return &MetaCommandHandler{
		session:        session,
		sessionManager: sessionMgr,
		expandMode:     false,
		captureFormat:  "text",
	}
}

// HandleMetaCommand processes meta commands that aren't CQL
func (h *MetaCommandHandler) HandleMetaCommand(command string) interface{} {
	upperCommand := strings.ToUpper(strings.TrimSpace(command))
	parts := strings.Fields(upperCommand)

	if len(parts) == 0 {
		return ""
	}

	switch parts[0] {
	case "CONSISTENCY":
		return h.handleConsistency(command)
	case "SHOW":
		return h.handleShow(command)
	case "TRACING":
		return h.handleTracing(command)
	case "PAGING":
		return h.handlePaging(command)
	case "AUTOFETCH":
		return h.handleAutoFetch(command)
	case "EXPAND":
		return h.handleExpand(command)
	case "SOURCE":
		return h.handleSource(command)
	case "CAPTURE":
		return h.handleCapture(command)
	case "COPY":
		return h.handleCopy(command)
	case "HELP":
		return h.handleHelp()
	default:
		return fmt.Sprintf("Unknown meta command: %s", parts[0])
	}
}

// handleConsistency handles CONSISTENCY command
func (h *MetaCommandHandler) handleConsistency(command string) interface{} {
	parts := strings.Fields(strings.ToUpper(command))

	if len(parts) == 1 {
		// Show current consistency
		return fmt.Sprintf("Current consistency level: %s", h.session.Consistency())
	}

	if len(parts) >= 2 {
		// Set consistency level - handle both "CONSISTENCY LOCAL_QUORUM" and "CONSISTENCY LOCAL QUORUM"
		level := parts[1]
		// Handle multi-word consistency levels (e.g., LOCAL_QUORUM might be split as LOCAL QUORUM)
		if len(parts) == 3 && parts[1] == "LOCAL" {
			level = parts[1] + "_" + parts[2]
		} else if len(parts) == 3 && parts[1] == "EACH" {
			level = parts[1] + "_" + parts[2]
		}

		if err := h.session.SetConsistency(level); err != nil {
			return fmt.Sprintf("Error setting consistency: %v", err)
		}
		return fmt.Sprintf("Consistency level set to %s", level)
	}

	return "Usage: CONSISTENCY [level]\nValid levels: ANY, ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM, LOCAL_ONE"
}

// handleShow handles SHOW commands
func (h *MetaCommandHandler) handleShow(command string) interface{} {
	upperCommand := strings.ToUpper(command)

	if strings.Contains(upperCommand, "VERSION") {
		// Show Cassandra version
		iter := h.session.Query("SELECT release_version FROM system.local").Iter()
		var version string
		if iter.Scan(&version) {
			_ = iter.Close()
			return fmt.Sprintf("Cassandra version: %s", version)
		}
		_ = iter.Close()
		return "Unable to get Cassandra version"
	}

	if strings.Contains(upperCommand, "HOST") {
		// Show current host connection
		iter := h.session.Query("SELECT rpc_address, data_center, rack FROM system.local").Iter()
		var host, datacenter, rack string
		if iter.Scan(&host, &datacenter, &rack) {
			_ = iter.Close()
			result := fmt.Sprintf("Connected to: %s\n", host)
			result += fmt.Sprintf("Datacenter: %s\n", datacenter)
			result += fmt.Sprintf("Rack: %s", rack)
			return result
		}
		_ = iter.Close()
		return "Unable to get host information"
	}

	if strings.Contains(upperCommand, "SESSION") {
		// Show session information
		currentKeyspace := ""
		if sessionManager != nil {
			currentKeyspace = sessionManager.CurrentKeyspace()
		}
		result := fmt.Sprintf("Current keyspace: %s\n", currentKeyspace)
		result += fmt.Sprintf("Consistency: %s\n", h.session.Consistency())
		result += fmt.Sprintf("Page size: %d\n", h.session.PageSize())
		result += fmt.Sprintf("Tracing: %v\n", h.session.Tracing())
		result += fmt.Sprintf("Auto-fetch: %v\n", h.session.AutoFetch())
		result += fmt.Sprintf("Expand mode: %v", h.expandMode)
		return result
	}

	return "Usage: SHOW VERSION | SHOW HOST | SHOW SESSION"
}

// handleTracing handles TRACING command
func (h *MetaCommandHandler) handleTracing(command string) interface{} {
	parts := strings.Fields(strings.ToUpper(command))

	if len(parts) < 2 {
		if h.session.Tracing() {
			return "Tracing is currently ON"
		}
		return "Tracing is currently OFF"
	}

	switch parts[1] {
	case "ON":
		h.session.SetTracing(true)
		return "Tracing turned ON"
	case "OFF":
		h.session.SetTracing(false)
		return "Tracing turned OFF"
	default:
		return "Usage: TRACING ON | OFF"
	}
}

// handlePaging handles PAGING command
func (h *MetaCommandHandler) handlePaging(command string) interface{} {
	parts := strings.Fields(command)

	switch len(parts) {
	case 1:
		// Show current page size
		pageSize := h.session.PageSize()
		if pageSize == 0 {
			return "Paging disabled (using server defaults)"
		}
		return fmt.Sprintf("Current page size: %d", pageSize)
	case 2:
		// Check if it's "PAGING OFF"
		if strings.ToUpper(parts[1]) == "OFF" {
			// Set to a large value to effectively disable paging
			h.session.SetPageSize(10000)
			return "Paging disabled (set to 10000)"
		}

		// Try to parse the page size
		var pageSize int
		if _, err := fmt.Sscanf(parts[1], "%d", &pageSize); err != nil {
			return fmt.Sprintf("Invalid page size: %s", parts[1])
		}

		if pageSize < 1 {
			return "Page size must be at least 1"
		}

		h.session.SetPageSize(pageSize)
		return fmt.Sprintf("Page size set to %d", pageSize)
	default:
		return "Usage: PAGING [size] | PAGING OFF"
	}
}

// handleAutoFetch handles AUTOFETCH command for auto-fetching all pages
func (h *MetaCommandHandler) handleAutoFetch(command string) interface{} {
	parts := strings.Fields(command)

	switch len(parts) {
	case 1:
		// Show current auto-fetch status
		if h.session.AutoFetch() {
			return "Auto-fetch is ON (all pages fetched automatically)"
		}
		return "Auto-fetch is OFF (pages fetched on scroll)"
	case 2:
		switch strings.ToUpper(parts[1]) {
		case "ON":
			h.session.SetAutoFetch(true)
			return "Auto-fetch enabled (all pages will be fetched automatically)"
		case "OFF":
			h.session.SetAutoFetch(false)
			return "Auto-fetch disabled (pages fetched on scroll)"
		default:
			return "Usage: AUTOFETCH ON | OFF"
		}
	default:
		return "Usage: AUTOFETCH ON | OFF"
	}
}

// handleExpand handles EXPAND command for vertical output
func (h *MetaCommandHandler) handleExpand(command string) interface{} {
	parts := strings.Fields(strings.ToUpper(command))

	switch len(parts) {
	case 1:
		if h.expandMode {
			return "Expand mode is currently ON (vertical output)"
		}
		return "Expand mode is currently OFF (table output)"
	case 2:
		switch parts[1] {
		case "ON":
			h.expandMode = true
			return "Expand mode turned ON - results will be shown vertically"
		case "OFF":
			h.expandMode = false
			return "Expand mode turned OFF - results will be shown as tables"
		default:
			return "Usage: EXPAND ON | OFF"
		}
	default:
		return "Usage: EXPAND ON | OFF"
	}
}

// handleSource handles SOURCE command to execute CQL from file
func (h *MetaCommandHandler) handleSource(command string) interface{} {
	parts := strings.Fields(command)

	if len(parts) < 2 {
		return "Usage: SOURCE 'filename'"
	}

	// Extract filename (remove quotes if present)
	filename := strings.Join(parts[1:], " ")
	filename = strings.Trim(filename, "'\"")

	// Expand home directory if needed
	if strings.HasPrefix(filename, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			filename = filepath.Join(home, filename[2:])
		}
	}

	// Read the file
	content, err := os.ReadFile(filename) // #nosec G304 - User-provided source filename
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}

	// Split into statements (simple split by semicolon)
	statements := strings.Split(string(content), ";")

	results := []string{}
	successCount := 0
	errorCount := 0

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Execute the statement
		result := ProcessCommand(stmt+";", h.session, h.sessionManager)
		logger.DebugfToFile("SOURCE", "Result type for '%s': %T", stmt, result)

		// Check if it's an error
		if err, ok := result.(error); ok {
			errorCount++
			results = append(results, fmt.Sprintf("Error in statement: %v", err))
		} else {
			successCount++

			// If capturing is enabled and this is a query result, write to capture file
			isCapturing := h.IsCapturing()
			logger.DebugfToFile("SOURCE", "IsCapturing=%v, captureOutput=%v, partitionedWriter=%v, result type: %T",
				isCapturing, h.captureOutput != nil, h.partitionedWriter != nil, result)
			if isCapturing {
				logger.DebugfToFile("SOURCE", "Capturing is enabled, result type: %T", result)
				switch v := result.(type) {
				case db.QueryResult:
					logger.DebugfToFile("SOURCE", "QueryResult with %d rows", len(v.Data))
					// Convert QueryResult to string arrays for capture
					if len(v.Data) > 0 {
						headers := v.Data[0]
						rows := [][]string{}
						if len(v.Data) > 1 {
							rows = v.Data[1:]
						}
						// Use WriteCaptureResultWithTypes if we have column types
						switch {
						case len(v.ColumnTypes) > 0:
							_ = h.WriteCaptureResultWithTypes(stmt, headers, v.ColumnTypes, rows, v.RawData)
						case len(v.RawData) > 0:
							_ = h.WriteCaptureResultWithRawData(stmt, headers, rows, v.RawData)
						default:
							_ = h.WriteCaptureResult(stmt, headers, rows)
						}
					}
				case db.StreamingQueryResult:
					// For streaming results, we need to fetch all rows
					defer v.Iterator.Close()

					rows := [][]string{}
					rawRows := []map[string]interface{}{}

					// Fetch rows from iterator
					for {
						row := make(map[string]interface{})
						if !v.Iterator.MapScan(row) {
							break
						}

						// Convert to string array
						stringRow := make([]string, len(v.ColumnNames))
						for i, col := range v.ColumnNames {
							if val, ok := row[col]; ok {
								stringRow[i] = fmt.Sprintf("%v", val)
							} else {
								stringRow[i] = "null"
							}
						}
						rows = append(rows, stringRow)
						rawRows = append(rawRows, row)
					}

					// Write to capture file
					if len(rows) > 0 {
						switch {
						case len(v.ColumnTypes) > 0:
							_ = h.WriteCaptureResultWithTypes(stmt, v.Headers, v.ColumnTypes, rows, rawRows)
						default:
							_ = h.WriteCaptureResultWithRawData(stmt, v.Headers, rows, rawRows)
						}
					}
				}
			}
		}
	}

	// Refresh schema cache after SOURCE command completes
	// This ensures cache is up-to-date after executing multiple DDL statements
	logger.DebugfToFile("SOURCE", "SOURCE command completed, refreshing schema cache")
	if cache := h.session.GetSchemaCache(); cache != nil {
		if err := cache.Refresh(); err != nil {
			logger.DebugfToFile("SOURCE", "Failed to refresh schema cache: %v", err)
		} else {
			logger.DebugfToFile("SOURCE", "Schema cache refreshed successfully")
		}
	}

	summary := fmt.Sprintf("\nExecuted %d statements from %s", successCount+errorCount, filename)
	if errorCount > 0 {
		summary += fmt.Sprintf(" (%d successful, %d failed)", successCount, errorCount)
	}

	if len(results) > 0 {
		return strings.Join(results, "\n") + "\n" + summary
	}

	return summary
}


// batchEntry holds a prepared statement template and its bound values
type batchEntry struct {
	query  string
	values []interface{}
}

// executeBatchWithValues executes a batch of INSERT queries using prepared statements
// and returns the number of errors
func (h *MetaCommandHandler) executeBatchWithValues(entries []batchEntry) int {
	if len(entries) == 0 {
		return 0
	}

	// Use UNLOGGED batch for better performance (like cqlsh COPY)
	batch := h.session.CreateBatch(gocql.UnloggedBatch)
	for _, entry := range entries {
		batch.Query(entry.query, entry.values...)
	}

	err := h.session.ExecuteBatch(batch)
	if err != nil {
		// If batch fails, try individual queries to count actual errors
		errors := 0
		for _, entry := range entries {
			if execErr := h.session.Query(entry.query, entry.values...).Exec(); execErr != nil {
				errors++
			}
		}
		return errors
	}
	return 0
}

// getTableColumns retrieves column names for a table
func (h *MetaCommandHandler) getTableColumns(table string) []string {
	// Parse table name (could be keyspace.table)
	parts := strings.Split(table, ".")
	var keyspace, tableName string

	if len(parts) == 2 {
		keyspace = parts[0]
		tableName = parts[1]
	} else {
		// Use current keyspace
		keyspace = h.sessionManager.CurrentKeyspace()
		if keyspace == "" {
			return []string{}
		}
		tableName = parts[0]
	}

	// Build query with values inline
	query := fmt.Sprintf(`SELECT column_name FROM system_schema.columns
	          WHERE keyspace_name = '%s' AND table_name = '%s'
	          ORDER BY position`, keyspace, tableName)

	result := h.session.ExecuteCQLQuery(query)

	switch v := result.(type) {
	case db.QueryResult:
		columns := make([]string, 0, len(v.Data))
		for _, row := range v.Data {
			if len(row) > 0 {
				columns = append(columns, row[0])
			}
		}
		return columns
	default:
		return []string{}
	}
}

// parseValueForBinding converts a CSV string value to the appropriate Go type
// for gocql prepared statement binding
func (h *MetaCommandHandler) parseValueForBinding(value string, _ string, _ string) interface{} {
	// Handle empty string as empty string, not null
	if value == "" {
		return ""
	}

	// Try to parse as integer
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	// Try to parse as boolean
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// Return as string (gocql will handle UUIDs, timestamps, etc.)
	return value
}

// IsExpandMode returns whether expand mode is on
func (h *MetaCommandHandler) IsExpandMode() bool {
	return h.expandMode
}

// FormatResultAsJSONWithRawData formats query results as JSON with optional raw data
func FormatResultAsJSONWithRawData(headers []string, rows [][]string, rawData []map[string]interface{}) (string, error) {
	type QueryResult struct {
		Query   string                   `json:"query,omitempty"`
		Columns []string                 `json:"columns"`
		Rows    []map[string]interface{} `json:"rows"`
		Count   int                      `json:"row_count"`
	}

	result := QueryResult{
		Columns: headers,
		Rows:    make([]map[string]interface{}, 0, len(rows)),
		Count:   len(rows),
	}

	// Use raw data if provided, otherwise fall back to string parsing
	if rawData != nil && len(rawData) == len(rows) {
		// Use the raw data directly - it preserves types
		result.Rows = rawData
	} else {
		// Fall back to parsing strings (backward compatibility)
		for _, row := range rows {
			rowMap := make(map[string]interface{})
			for i, col := range headers {
				if i < len(row) {
					// Try to parse as number or boolean
					value := row[i]
					if value == "null" {
						rowMap[col] = nil
					} else if value == "true" || value == "false" {
						rowMap[col] = value == "true"
					} else if num, err := json.Number(value).Float64(); err == nil {
						rowMap[col] = num
					} else {
						rowMap[col] = value
					}
				}
			}
			result.Rows = append(result.Rows, rowMap)
		}
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// Close closes any open resources
func (h *MetaCommandHandler) Close() {
	if h.captureOutput != nil {
		// If JSON format, close the array
		if h.captureFormat == "json" {
			// Close the JSON array
			_, _ = h.captureOutput.Write([]byte("\n]\n"))
		} else if h.captureFormat == "csv" && h.csvWriter != nil {
			// Flush CSV writer
			h.csvWriter.Flush()
			h.csvWriter = nil
		}
		_ = h.captureOutput.Close()
		h.captureOutput = nil
		h.captureFile = ""
		h.captureFormat = "text"
	}
}
