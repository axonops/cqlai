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

// ExecuteCQLQuery executes a regular CQL query
func (s *Session) ExecuteCQLQuery(query string) interface{} {
	logger.DebugfToFile("ExecuteCQLQuery", "Called with query: %s", query)

	if s == nil || s.Session == nil {
		return fmt.Errorf("not connected to database")
	}

	// Check if it's a SELECT or DESCRIBE query (DESCRIBE returns results in 4.0+)
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	if strings.HasPrefix(upperQuery, "SELECT") || strings.HasPrefix(upperQuery, "DESCRIBE") {
		logger.DebugToFile("ExecuteCQLQuery", "Routing to ExecuteSelectQuery for SELECT/DESCRIBE")
		return s.ExecuteSelectQuery(query)
	} else if strings.HasPrefix(upperQuery, "USE ") {
		// Handle USE statement - gocql doesn't support USE directly
		// We just track the keyspace locally for DESCRIBE commands
		parts := strings.Fields(query)
		if len(parts) >= 2 {
			keyspace := strings.Trim(strings.Trim(parts[1], ";"), "\"")

			// Verify the keyspace exists
			var exists string
			iter := s.Query("SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name = ?", keyspace).Iter()
			if !iter.Scan(&exists) {
				iter.Close()
				return fmt.Errorf("Keyspace '%s' does not exist", keyspace)
			}
			iter.Close()

			// Update the current keyspace in our session wrapper
			s.SetKeyspace(keyspace)
			return fmt.Sprintf("Now using keyspace %s", keyspace)
		}
		return "Invalid USE statement"
	} else {
		// Execute non-SELECT query
		if err := s.Query(query).Exec(); err != nil {
			// Check if it's a connection error
			errStr := err.Error()
			if strings.Contains(errStr, "connection refused") ||
				strings.Contains(errStr, "no connections") ||
				strings.Contains(errStr, "unable to connect") {
				return fmt.Errorf("Connection lost to Cassandra. Please check if the server is running")
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
	
	// Check if JSON format is requested and modify query accordingly
	if s.GetOutputFormat() == OutputFormatJSON {
		query = s.convertToJSONQuery(query)
		logger.DebugfToFile("executeSelectQuery", "Converted to JSON query: %s", query)
	}

	// Check if we should use streaming for large results
	// This is a simple heuristic - could be made configurable
	useStreaming := s.shouldUseStreaming(query)

	if useStreaming {
		return s.ExecuteStreamingQuery(query)
	}

	// Track query execution time
	startTime := time.Now()

	iter := s.Query(query).Iter()

	// Check for connection errors early
	if err := iter.Close(); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "no connections") ||
			strings.Contains(errStr, "unable to connect") {
			return fmt.Errorf("Connection lost to Cassandra. Please check if the server is running")
		}
		// Re-create the iterator if no connection error
		iter = s.Query(query).Iter()
	} else {
		// Re-create the iterator since we closed it
		iter = s.Query(query).Iter()
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
			if keyInfo.Kind == "partition_key" {
				headers[i] += " (PK)"
			} else if keyInfo.Kind == "clustering" {
				headers[i] += " (C)"
			}
		} else {
			logger.DebugfToFile("executeSelectQuery", "No key info for column %s", col.Name)
		}
	}

	// Collect results - use MapScan for better type handling
	results := [][]string{headers}

	logger.DebugToFile("executeSelectQuery", "Starting row scan with MapScan...")
	rowNum := 0

	// Use MapScan which handles types better than Scan with interface{}
	for {
		rowMap := make(map[string]interface{})
		if !iter.MapScan(rowMap) {
			logger.DebugToFile("executeSelectQuery", "MapScan returned false - no more rows or error")
			break
		}

		row := make([]string, len(filteredColumns))

		for i, col := range filteredColumns {
			if val, ok := rowMap[col.Name]; ok {
				if val == nil {
					row[i] = "null"
				} else {
					// Handle different types appropriately
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
			} else {
				row[i] = "null"
			}
		}
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

	// Check the output format
	outputFormat := s.GetOutputFormat()
	logger.DebugfToFile("ExecuteSelectQuery", "Current output format: %v", outputFormat)

	queryResult := QueryResult{
		Data:        results,
		Duration:    duration,
		RowCount:    rowNum, // rowNum already contains the count of data rows (excluding header)
		ColumnTypes: columnTypes,
		Format:      outputFormat,
	}

	// Just pass the format info, UI will handle formatting
	logger.DebugfToFile("ExecuteSelectQuery", "Returning QueryResult with format: %v", outputFormat)

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
			if keyInfo.Kind == "partition_key" {
				headers[i] += " (PK)"
			} else if keyInfo.Kind == "clustering" {
				headers[i] += " (C)"
			}
		}
	}

	// Return streaming result with iterator
	return StreamingQueryResult{
		Headers:     headers,
		ColumnNames: columnNames,
		ColumnTypes: columnTypes,
		Format:      s.GetOutputFormat(),
		Iterator:    iter,
		StartTime:   startTime,
	}
}

// convertToJSONQuery converts a SELECT query to SELECT JSON format
func (s *Session) convertToJSONQuery(query string) string {
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

	// If no keyspace specified, use current keyspace
	if keyspaceName == "" {
		keyspaceName = s.CurrentKeyspace()
		if keyspaceName == "" {
			logger.DebugToFile("getKeyColumns", "No keyspace available")
			return keyColumns
		}
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
