package db

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/logger"
)

// formatUDTMap formats a UDT map for display
func formatUDTMap(m map[string]interface{}) string {
	if len(m) == 0 {
		return "{}"
	}
	
	var parts []string
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s: %v", k, v))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// captureTracer implements gocql.Tracer to capture trace IDs
type captureTracer struct {
	traceID []byte
}

func (t *captureTracer) Trace(traceID []byte) {
	t.traceID = traceID
}

// ExecuteCQLQuery executes a regular CQL query
func (s *Session) ExecuteCQLQuery(query string) interface{} {
	logger.DebugfToFile("ExecuteCQLQuery", "Called with query: %s", query)

	if s == nil || s.Session == nil {
		return fmt.Errorf("not connected to database")
	}

	// Check if it's a query that returns results
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	switch {
	case strings.HasPrefix(upperQuery, "SELECT") || strings.HasPrefix(upperQuery, "DESCRIBE") || strings.HasPrefix(upperQuery, "LIST"):
		logger.DebugToFile("ExecuteCQLQuery", "Routing to ExecuteSelectQuery for query that returns results")
		return s.ExecuteSelectQuery(query)
	case strings.HasPrefix(upperQuery, "USE "):
		// Handle USE statement - gocql doesn't support USE directly
		// Return the keyspace name for the UI/router layer to handle
		parts := strings.Fields(query)
		if len(parts) >= 2 {
			keyspace := strings.Trim(strings.Trim(parts[1], ";"), "\"")

			// Verify the keyspace exists
			// Use appropriate system table based on Cassandra version
			var exists string
			var iter *gocql.Iter
			
			if s.IsVersion3OrHigher() {
				// Cassandra 3.0+ uses system_schema.keyspaces
				iter = s.Query("SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name = ?", keyspace).Iter()
			} else {
				// Cassandra 2.x uses system.schema_keyspaces
				iter = s.Query("SELECT keyspace_name FROM system.schema_keyspaces WHERE keyspace_name = ?", keyspace).Iter()
			}
			
			if !iter.Scan(&exists) {
				_ = iter.Close()
				return fmt.Errorf("keyspace '%s' does not exist", keyspace)
			}
			_ = iter.Close()

			// Return success - the router/UI will handle updating the current keyspace
			return fmt.Sprintf("Now using keyspace %s", keyspace)
		}
		return "Invalid USE statement"
	default:
		// Execute non-SELECT query
		if err := s.Query(query).Exec(); err != nil {
			// Check if it's a connection error
			errStr := err.Error()
			if strings.Contains(errStr, "connection refused") ||
				strings.Contains(errStr, "no connections") ||
				strings.Contains(errStr, "unable to connect") {
				return fmt.Errorf("connection lost to Cassandra - please check if the server is running")
			}
			return fmt.Errorf("query failed: %v", err)
		}
		return "Query executed successfully"
	}
}

// ExecuteSelectQuery executes a SELECT query and returns formatted results
func (s *Session) ExecuteSelectQuery(query string) interface{} {
	// Add debug logging
	logger.DebugToFile("executeSelectQuery", "Starting executeSelectQuery")

	// Check if we should use streaming for large results
	// This is a simple heuristic - could be made configurable
	useStreaming := s.shouldUseStreaming(query)

	if useStreaming {
		return s.ExecuteStreamingQuery(query)
	}

	// Track query execution time
	startTime := time.Now()

	// Create the query
	q := s.Query(query)
	
	// Enable tracing if needed and capture trace ID
	var tracer *captureTracer
	if s.tracing {
		tracer = &captureTracer{}
		q = q.Trace(tracer)
		defer func() {
			// Store the trace ID for later retrieval
			if tracer != nil && tracer.traceID != nil {
				s.lastTraceID = tracer.traceID
			}
		}()
	}

	iter := q.Iter()

	// Check for connection errors early
	if err := iter.Close(); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "no connections") ||
			strings.Contains(errStr, "unable to connect") {
			return fmt.Errorf("connection lost to Cassandra - please check if the server is running")
		}
		// Re-create the iterator if no connection error
		q = s.Query(query)
		if s.tracing && tracer != nil {
			q = q.Trace(tracer)
		}
		iter = q.Iter()
	} else {
		// Re-create the iterator since we closed it
		q = s.Query(query)
		if s.tracing && tracer != nil {
			q = q.Trace(tracer)
		}
		iter = q.Iter()
	}

	// Get column info
	columns := iter.Columns()
	logger.DebugfToFile("executeSelectQuery", "Number of columns: %d", len(columns))

	// Check if this is a DESCRIBE KEYSPACE or DESCRIBE TABLE query that should filter "type" column
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	shouldFilterType := (strings.HasPrefix(upperQuery, "DESCRIBE KEYSPACE") ||
		strings.HasPrefix(upperQuery, "DESCRIBE TABLE"))

	// Filter out "type" column if needed
	filteredColumns := columns
	if shouldFilterType {
		var newColumns []gocql.ColumnInfo
		for _, col := range columns {
			if col.Name == "type" {
				logger.DebugfToFile("executeSelectQuery", "Filtering out 'type' column")
			} else {
				newColumns = append(newColumns, col)
			}
		}
		filteredColumns = newColumns
	}

	// Log column details
	for i, col := range filteredColumns {
		logger.DebugfToFile("executeSelectQuery", "Column %d: Name=%s, Type=%v, TypeInfo=%T",
			i, col.Name, col.TypeInfo.Type(), col.TypeInfo)
	}

	if len(filteredColumns) == 0 {
		if err := iter.Close(); err != nil {
			logger.DebugfToFile("executeSelectQuery", "Error closing empty iterator: %v", err)
			return fmt.Errorf("query failed: %v", err)
		}
		return "No results"
	}

	// Get key column information
	keyColumns := s.GetKeyColumns(query)

	// Prepare headers with key indicators and collect column types
	headers := make([]string, len(filteredColumns))
	columnTypes := make([]string, len(filteredColumns))
	for i, col := range filteredColumns {
		headers[i] = col.Name
		// Store the column type - use TypeInfoToString to handle custom types
		columnTypes[i] = TypeInfoToString(col.TypeInfo)

		// Add indicators for key columns
		if keyInfo, exists := keyColumns[col.Name]; exists {
			logger.DebugfToFile("executeSelectQuery", "Adding indicator for %s: %s", col.Name, keyInfo.Kind)
			switch keyInfo.Kind {
			case "partition_key":
				headers[i] += " (PK)"
			case "clustering":
				headers[i] += " (C)"
			}
		} else {
			logger.DebugfToFile("executeSelectQuery", "No key info for column %s", col.Name)
		}
	}

	// Collect results - use MapScan for better type handling
	results := [][]string{headers}
	rawData := make([]map[string]interface{}, 0)

	logger.DebugToFile("executeSelectQuery", "Starting row scan with MapScan...")
	rowNum := 0

	// Extract clean column names (without PK/C indicators)
	cleanHeaders := make([]string, len(filteredColumns))
	for i, col := range filteredColumns {
		cleanHeaders[i] = col.Name
	}

	// Use MapScan for better type handling
	// IMPORTANT LIMITATIONS:
	// 1. NULL values: gocql returns zero values (0, false, "") for NULL columns when scanning 
	//    into interface{}. To properly detect NULLs, we would need to scan into typed pointers
	//    (*int, *string, etc), but we don't know column types at compile time.
	// 2. UDTs: gocql often returns empty maps for UDTs when using MapScan or scanning into
	//    interface{}. To properly handle UDTs, you need to scan into specific struct types
	//    that match the UDT schema, which requires compile-time knowledge of the UDT structure.
	// These are fundamental limitations of using dynamic typing with gocql.
	for {
		// First try MapScan for the row
		rowMap := make(map[string]interface{})
		if !iter.MapScan(rowMap) {
			logger.DebugToFile("executeSelectQuery", "MapScan returned false - no more rows or error")
			break
		}

		// Store raw data for JSON export (preserves types)
		rawRow := make(map[string]interface{})
		// Create formatted row for display
		row := make([]string, len(filteredColumns))
		
		for i, col := range filteredColumns {
			val, hasValue := rowMap[col.Name]
			
			if !hasValue || val == nil {
				rawRow[col.Name] = nil
				row[i] = "null"
			} else {
				// Special handling for UDTs
				if col.TypeInfo.Type() == gocql.TypeUDT {
					// When scanning into interface{}, gocql returns a map for UDTs
					// Unfortunately, it often returns an empty map due to how it handles dynamic types
					// This is a known limitation of gocql when not using specific struct types
					
					if m, ok := val.(map[string]interface{}); ok {
						if len(m) > 0 {
							// We got actual UDT data
							rawRow[col.Name] = m
							row[i] = formatUDTMap(m)
						} else {
							// Empty map - common issue with gocql and UDTs
							// This is a known limitation when using MapScan or Scan with interface{}
							// To properly handle UDTs, you need to scan into specific struct types
							logger.DebugfToFile("ExecuteSelectQuery", "UDT %s returned empty map (gocql limitation)", col.Name)
							rawRow[col.Name] = m
							row[i] = "{}"
						}
					} else if bytes, ok := val.([]byte); ok {
						// Sometimes UDTs come as raw bytes
						logger.DebugfToFile("ExecuteSelectQuery", "UDT %s came as bytes: %d bytes", col.Name, len(bytes))
						// We would need the UDT schema to properly unmarshal this
						rawRow[col.Name] = map[string]interface{}{"_raw_bytes": fmt.Sprintf("%x", bytes)}
						row[i] = fmt.Sprintf("{_raw_bytes:%d}", len(bytes))
					} else {
						rawRow[col.Name] = val
						row[i] = fmt.Sprintf("%v", val)
					}
				} else {
					// Store the actual value for JSON
					rawRow[col.Name] = val
					
					// Format for display
					switch v := val.(type) {
					case gocql.UUID:
						row[i] = v.String()
					case []byte:
						row[i] = fmt.Sprintf("0x%x", v)
					case time.Time:
						row[i] = v.Format(time.RFC3339)
					default:
						row[i] = fmt.Sprintf("%v", val)
					}
				}
			}
		}
		rawData = append(rawData, rawRow)
		results = append(results, row)
		rowNum++
	}
	logger.DebugfToFile("executeSelectQuery", "Scan completed. Total rows: %d", rowNum)

	if err := iter.Close(); err != nil {
		logger.DebugfToFile("executeSelectQuery", "Iterator close error: %v", err)
		return fmt.Errorf("query failed: %v", err)
	}

	// Calculate query duration
	duration := time.Since(startTime)

	queryResult := QueryResult{
		Data:        results,
		RawData:     rawData,
		Duration:    duration,
		RowCount:    rowNum, // rowNum already contains the count of data rows (excluding header)
		ColumnTypes: columnTypes,
		Headers:     cleanHeaders,
	}

	// Just pass the result, UI will handle formatting
	logger.DebugfToFile("ExecuteSelectQuery", "Returning QueryResult with %d rows", rowNum)

	return queryResult
}

// shouldUseStreaming determines if a query should use streaming based on heuristics
func (s *Session) shouldUseStreaming(query string) bool {
	// Always use streaming unless there's a small LIMIT
	upperQuery := strings.ToUpper(strings.TrimSpace(query))

	// Check for LIMIT clause
	if strings.Contains(upperQuery, " LIMIT ") {
		// Extract limit value - simple regex for "LIMIT n"
		re := regexp.MustCompile(`LIMIT\s+(\d+)`)
		matches := re.FindStringSubmatch(upperQuery)
		if len(matches) > 1 {
			limit, err := strconv.Atoi(matches[1])
			if err == nil && limit <= 1000 {
				// Small limit, don't use streaming
				logger.DebugfToFile("shouldUseStreaming", "Query has LIMIT %d, not using streaming", limit)
				return false
			}
		}
	}

	// Use streaming for all other SELECT queries
	logger.DebugToFile("shouldUseStreaming", "Using streaming for query")
	return true
}

// ExecuteStreamingQuery executes a query and returns a streaming result
func (s *Session) ExecuteStreamingQuery(query string) interface{} {
	logger.DebugToFile("ExecuteStreamingQuery", "Starting streaming query execution")

	startTime := time.Now()
	// Use the session's page size for pagination
	q := s.Query(query)
	q.PageSize(s.pageSize)
	
	// Enable tracing if needed and capture trace ID
	var tracer *captureTracer
	if s.tracing {
		tracer = &captureTracer{}
		q = q.Trace(tracer)
		defer func() {
			// Store the trace ID for later retrieval
			if tracer != nil && tracer.traceID != nil {
				s.lastTraceID = tracer.traceID
			}
		}()
	}
	
	iter := q.Iter()

	// Get column info
	columns := iter.Columns()
	logger.DebugfToFile("ExecuteStreamingQuery", "Got %d columns from iterator", len(columns))
	if len(columns) == 0 {
		if err := iter.Close(); err != nil {
			return fmt.Errorf("query failed: %v", err)
		}
		return "No results"
	}

	// Check if this is a DESCRIBE query that should filter "type" column
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	shouldFilterType := (strings.HasPrefix(upperQuery, "DESCRIBE KEYSPACE") ||
		strings.HasPrefix(upperQuery, "DESCRIBE TABLE"))

	// Filter columns if needed
	filteredColumns := columns
	if shouldFilterType {
		logger.DebugToFile("ExecuteStreamingQuery", "Filtering type column for DESCRIBE query")
		var newColumns []gocql.ColumnInfo
		for _, col := range columns {
			if col.Name != "type" {
				newColumns = append(newColumns, col)
			}
		}
		filteredColumns = newColumns
	}

	logger.DebugfToFile("ExecuteStreamingQuery", "After filtering: %d columns", len(filteredColumns))

	// Get key column information
	keyColumns := s.GetKeyColumns(query)

	// Prepare headers with key indicators
	headers := make([]string, len(filteredColumns))
	columnNames := make([]string, len(filteredColumns))
	columnTypes := make([]string, len(filteredColumns))
	for i, col := range filteredColumns {
		columnNames[i] = col.Name // Store original name
		headers[i] = col.Name     // Start with original name
		columnTypes[i] = TypeInfoToString(col.TypeInfo)

		// Add indicators for key columns
		if keyInfo, exists := keyColumns[col.Name]; exists {
			switch keyInfo.Kind {
			case "partition_key":
				headers[i] += " (PK)"
			case "clustering":
				headers[i] += " (C)"
			}
		}
	}

	// Return streaming result with iterator
	return StreamingQueryResult{
		Headers:     headers,
		ColumnNames: columnNames,
		ColumnTypes: columnTypes,
		Iterator:    iter,
		StartTime:   startTime,
	}
}

// ConvertToJSONQuery converts a SELECT query to SELECT JSON format
// This is now a public method so it can be called from the router/UI layer when needed
func ConvertToJSONQuery(query string) string {
	upperQuery := strings.ToUpper(strings.TrimSpace(query))

	// Check if it's already a JSON query
	if strings.Contains(upperQuery, "SELECT JSON") || strings.Contains(upperQuery, "SELECT DISTINCT JSON") {
		return query
	}

	// Only convert SELECT queries
	if !strings.HasPrefix(upperQuery, "SELECT") {
		return query
	}

	// Handle SELECT DISTINCT
	if strings.HasPrefix(upperQuery, "SELECT DISTINCT") {
		// Replace "SELECT DISTINCT" with "SELECT DISTINCT JSON"
		re := regexp.MustCompile(`(?i)^SELECT\s+DISTINCT\s+`)
		return re.ReplaceAllString(query, "SELECT DISTINCT JSON ")
	}

	// Handle regular SELECT
	// Replace "SELECT" with "SELECT JSON"
	re := regexp.MustCompile(`(?i)^SELECT\s+`)
	return re.ReplaceAllString(query, "SELECT JSON ")
}

// GetKeyColumns returns information about partition and clustering columns for a table
func (s *Session) GetKeyColumns(query string) map[string]KeyColumnInfo {
	keyColumns := make(map[string]KeyColumnInfo)

	// Try to extract table name from the SELECT query
	// Handle patterns like: SELECT ... FROM keyspace.table or FROM table
	re := regexp.MustCompile(`(?i)FROM\s+(?:([a-zA-Z_][a-zA-Z0-9_]*)\.)?([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := re.FindStringSubmatch(query)

	logger.DebugfToFile("getKeyColumns", "Query: %s", query)
	logger.DebugfToFile("getKeyColumns", "Regex matches: %v", matches)

	if len(matches) < 3 {
		logger.DebugToFile("getKeyColumns", "Could not extract table name from query")
		return keyColumns
	}

	keyspaceName := matches[1] // May be empty
	tableName := matches[2]

	// If no keyspace specified, we can't determine key columns
	// The UI/router layer should track the current keyspace
	if keyspaceName == "" {
		logger.DebugToFile("getKeyColumns", "No keyspace specified")
		return keyColumns
	}

	logger.DebugfToFile("getKeyColumns", "Looking up columns for %s.%s", keyspaceName, tableName)

	// Query system_schema.columns for key column information
	colQuery := `SELECT column_name, kind, position 
	            FROM system_schema.columns 
	            WHERE keyspace_name = ? AND table_name = ?`

	iter := s.Query(colQuery, keyspaceName, tableName).Iter()
	defer iter.Close()

	var columnName, kind string
	var position int

	for iter.Scan(&columnName, &kind, &position) {
		// Only track partition_key and clustering columns
		if kind == "partition_key" || kind == "clustering" {
			keyColumns[columnName] = KeyColumnInfo{
				Kind:     kind,
				Position: position,
			}
			logger.DebugfToFile("getKeyColumns", "Found key column: %s (%s, pos %d)", columnName, kind, position)
		}
	}

	return keyColumns
}
