package db

import (
	"fmt"
	"math/big"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/logger"
)

// formatUDTMap formats a UDT map for display
// formatValueInUDT formats a value that appears inside a UDT or collection
// Strings should be quoted in this context
func formatValueInUDT(val interface{}) string {
	switch v := val.(type) {
	case nil:
		return "null"
	case string:
		// Quote strings inside UDTs/collections
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case map[string]interface{}:
		return formatUDTMap(v)
	case map[interface{}]interface{}:
		// Convert to string-keyed map for display
		m := make(map[string]interface{})
		for k, val := range v {
			m[fmt.Sprintf("%v", k)] = val
		}
		return formatUDTMap(m)
	case []interface{}:
		// Format list/set/tuple
		if len(v) == 0 {
			return "[]"
		}
		var parts []string
		for _, item := range v {
			parts = append(parts, formatValueInUDT(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case []map[string]interface{}:
		// Format list of UDT maps
		if len(v) == 0 {
			return "[]"
		}
		var parts []string
		for _, item := range v {
			parts = append(parts, formatUDTMap(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case gocql.UUID:
		return v.String()
	case []byte:
		return fmt.Sprintf("0x%x", v)
	case time.Time:
		return v.Format(time.RFC3339)
	case time.Duration:
		return v.String()
	case net.IP:
		return v.String()
	case *big.Int:
		return v.String()
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatUDTMap(m map[string]interface{}) string {
	if len(m) == 0 {
		return "{}"
	}

	var parts []string
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s: %v", k, formatValueInUDT(v)))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// FormatValue formats any value for display, handling nested structures
// This is called for top-level values, so strings should NOT be quoted
func FormatValue(val interface{}) string {
	switch v := val.(type) {
	case nil:
		return "null"
	case string:
		// Don't quote top-level strings
		return v
	case map[string]interface{}:
		return formatUDTMap(v)
	case map[interface{}]interface{}:
		// Convert to string-keyed map for display
		m := make(map[string]interface{})
		for k, val := range v {
			m[fmt.Sprintf("%v", k)] = val
		}
		return formatUDTMap(m)
	case []interface{}:
		// Format list/set/tuple
		if len(v) == 0 {
			return "[]"
		}
		var parts []string
		for _, item := range v {
			parts = append(parts, FormatValue(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case []map[string]interface{}:
		// Format list of UDT maps (e.g., list<frozen<phone>>)
		if len(v) == 0 {
			return "[]"
		}
		var parts []string
		for _, item := range v {
			parts = append(parts, formatUDTMap(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case gocql.UUID:
		return v.String()
	case []byte:
		return fmt.Sprintf("0x%x", v)
	case time.Time:
		return v.Format(time.RFC3339)
	case time.Duration:
		return v.String()
	case net.IP:
		return v.String()
	case *big.Int:
		return v.String()
	case bool:
		return fmt.Sprintf("%v", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		// For unknown types, treat as string and quote it
		return "'" + strings.ReplaceAll(fmt.Sprintf("%v", val), "'", "''") + "'"
	}
}

// extractTableName extracts the keyspace and table name from a SELECT query
func extractTableName(query string) (keyspace, table string) {
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

// getColumnTypeFromSystemTable gets the full type definition for a column from system tables
func (s *Session) getColumnTypeFromSystemTable(keyspace, table, column string) string {
	if s.Session == nil {
		return ""
	}

	query := `SELECT type FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ? AND column_name = ?`

	var columnType string
	iter := s.Query(query, keyspace, table, column).Iter()
	if !iter.Scan(&columnType) {
		iter.Close()
		return ""
	}
	iter.Close()

	return columnType
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

	// Initialize UDT registry if needed (will be cached)
	if s.udtRegistry == nil {
		s.udtRegistry = NewUDTRegistry(s.Session)
	}

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

	// For UDT columns, we need to get the full type definition from system tables
	queryKeyspace, tableName := extractTableName(query)
	currentKeyspace := queryKeyspace
	if currentKeyspace == "" {
		currentKeyspace = s.Keyspace()
	}

	for i, col := range filteredColumns {
		headers[i] = col.Name
		// Store the column type - use TypeInfoToString to handle custom types
		basicType := TypeInfoToString(col.TypeInfo)

		// If it's a UDT, try to get the full type definition
		if basicType == "udt" && currentKeyspace != "" && tableName != "" {
			fullType := s.getColumnTypeFromSystemTable(currentKeyspace, tableName, col.Name)
			if fullType != "" {
				columnTypes[i] = fullType
			} else {
				columnTypes[i] = basicType
			}
		} else {
			columnTypes[i] = basicType
		}

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

	// Use Scan with interface{} slice instead of MapScan to get raw bytes for UDTs
	// MapScan returns empty maps for UDTs, but we need to use RawBytes for UDT columns
	for {
		// Create a slice to scan into - use RawBytes for UDT columns
		scanDest := make([]interface{}, len(filteredColumns))
		for i, col := range filteredColumns {
			if col.TypeInfo.Type() == gocql.TypeUDT {
				// Use RawBytes for UDT columns to bypass gocql's broken UDT decoding
				scanDest[i] = new(RawBytes)
			} else {
				scanDest[i] = new(interface{})
			}
		}

		// Scan the row
		if !iter.Scan(scanDest...) {
			logger.DebugToFile("executeSelectQuery", "Scan returned false - no more rows or error")
			break
		}

		// Store raw data for JSON export (preserves types)
		rawRow := make(map[string]interface{})
		// Create formatted row for display
		row := make([]string, len(filteredColumns))

		for i, col := range filteredColumns {
			// Extract value based on type
			var val interface{}
			if col.TypeInfo.Type() == gocql.TypeUDT {
				// For UDT columns, we used RawBytes
				rawBytes := scanDest[i].(*RawBytes)
				if rawBytes != nil && *rawBytes != nil {
					val = []byte(*rawBytes)
				} else {
					val = nil
				}
			} else {
				// Regular column - dereference the pointer
				val = *(scanDest[i].(*interface{}))
			}

			if val == nil {
				rawRow[cleanHeaders[i]] = nil
				row[i] = "null"
			} else {
				// Special handling for UDTs and complex types
				typeStr := columnTypes[i]

				// Parse the type string to get structured type information
				typeInfo, parseErr := ParseCQLType(typeStr)

				// Add debug logging to understand what we're getting
				logger.DebugfToFile("ExecuteSelectQuery", "Column %s: typeStr=%s, gocqlType=%v, parsedType=%v, parseErr=%v, valType=%T",
					col.Name, typeStr, col.TypeInfo.Type(), typeInfo, parseErr, val)

				if col.TypeInfo.Type() == gocql.TypeUDT || (typeInfo != nil && typeInfo.BaseType == "udt") {
					// UDT handling - try to decode if we got raw bytes
					if bytes, ok := val.([]byte); ok && len(bytes) > 0 {
						logger.DebugfToFile("ExecuteSelectQuery", "UDT %s came as bytes: %d bytes", col.Name, len(bytes))

						// Use our binary decoder to decode the UDT
						decoder := NewBinaryDecoder(s.udtRegistry)

						// Determine the keyspace - prefer query keyspace, then current
						keyspace := currentKeyspace
						if keyspace == "" {
							keyspace = s.Keyspace()
							if keyspace == "" && s.cluster != nil {
								keyspace = s.cluster.Keyspace
							}
						}

						// Try to decode the UDT
						if typeInfo != nil {
							decoded, err := decoder.Decode(bytes, typeInfo, keyspace)
							if err != nil {
								logger.DebugfToFile("ExecuteSelectQuery", "Failed to decode UDT %s: %v", col.Name, err)
								// Fall back to showing raw bytes info
								rawRow[cleanHeaders[i]] = map[string]interface{}{"_raw_bytes": fmt.Sprintf("%x", bytes)}
								row[i] = fmt.Sprintf("{_raw_bytes:%d}", len(bytes))
							} else {
								// Successfully decoded UDT
								rawRow[cleanHeaders[i]] = decoded
								// Format for display
								if m, ok := decoded.(map[string]interface{}); ok {
									row[i] = formatUDTMap(m)
								} else {
									row[i] = fmt.Sprintf("%v", decoded)
								}
							}
						} else {
							// Couldn't parse type, show raw bytes
							rawRow[cleanHeaders[i]] = map[string]interface{}{"_raw_bytes": fmt.Sprintf("%x", bytes)}
							row[i] = fmt.Sprintf("{_raw_bytes:%d}", len(bytes))
						}
					} else if m, ok := val.(map[string]interface{}); ok {
						// Sometimes gocql returns a map directly
						if len(m) > 0 {
							rawRow[cleanHeaders[i]] = m
							row[i] = formatUDTMap(m)
						} else {
							// Empty map - common issue with gocql and UDTs
							logger.DebugfToFile("ExecuteSelectQuery", "UDT %s returned empty map", col.Name)
							rawRow[cleanHeaders[i]] = m
							row[i] = "{}"
						}
					} else {
						// Other format - just display as is
						rawRow[cleanHeaders[i]] = val
						row[i] = fmt.Sprintf("%v", val)
					}
				} else if typeInfo != nil && (typeInfo.BaseType == "list" || typeInfo.BaseType == "set" ||
					typeInfo.BaseType == "map" || typeInfo.BaseType == "tuple") {
					// Handle collections that might contain UDTs
					if bytes, ok := val.([]byte); ok && len(bytes) > 0 {
						decoder := NewBinaryDecoder(s.udtRegistry)
						keyspace := s.Keyspace()
						if keyspace == "" && s.cluster != nil {
							keyspace = s.cluster.Keyspace
						}

						decoded, err := decoder.Decode(bytes, typeInfo, keyspace)
						if err != nil {
							logger.DebugfToFile("ExecuteSelectQuery", "Failed to decode collection %s: %v", col.Name, err)
							rawRow[cleanHeaders[i]] = val
							row[i] = fmt.Sprintf("%v", val)
						} else {
							rawRow[cleanHeaders[i]] = decoded
							row[i] = FormatValue(decoded)
						}
					} else {
						// Use existing collection formatting
						rawRow[cleanHeaders[i]] = val
						// Debug: log the type of val
						logger.DebugfToFile("ExecuteSelectQuery", "Collection %s value type: %T", col.Name, val)
						row[i] = FormatValue(val)
					}
				} else {
					// Store the actual value for JSON
					rawRow[cleanHeaders[i]] = val

					// Format for display - use formatValue which handles collections properly
					row[i] = FormatValue(val)
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

	// For UDT columns, we need to get the full type definition from system tables
	queryKeyspace, tableName := extractTableName(query)
	currentKeyspace := queryKeyspace
	if currentKeyspace == "" {
		currentKeyspace = s.Keyspace()
	}

	for i, col := range filteredColumns {
		columnNames[i] = col.Name // Store original name
		headers[i] = col.Name     // Start with original name

		// Store the column type - use TypeInfoToString to handle custom types
		basicType := TypeInfoToString(col.TypeInfo)

		// If it's a UDT, try to get the full type definition
		if basicType == "udt" && currentKeyspace != "" && tableName != "" {
			fullType := s.getColumnTypeFromSystemTable(currentKeyspace, tableName, col.Name)
			if fullType != "" {
				columnTypes[i] = fullType
			} else {
				columnTypes[i] = basicType
			}
		} else {
			columnTypes[i] = basicType
		}

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
		Keyspace:    currentKeyspace,
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
