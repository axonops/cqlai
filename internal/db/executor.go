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
	case []string:
		// Format list/set of strings
		if len(v) == 0 {
			return "[]"
		}
		parts := v // v is already []string, no need to copy element by element
		return "[" + strings.Join(parts, " ") + "]"
	case map[string]string:
		// Format map<text, text>
		if len(v) == 0 {
			return "{}"
		}
		var parts []string
		for key, val := range v {
			parts = append(parts, fmt.Sprintf("%s:%s", key, val))
		}
		return "map[" + strings.Join(parts, " ") + "]"
	case map[string]int:
		// Format map<text, int>
		if len(v) == 0 {
			return "{}"
		}
		var parts []string
		for key, val := range v {
			parts = append(parts, fmt.Sprintf("%s:%d", key, val))
		}
		return "map[" + strings.Join(parts, " ") + "]"
	case []int, []int32, []int64:
		// Format list/set of integers
		return fmt.Sprintf("%v", v)
	case []float32, []float64:
		// Format list/set of floats (including vectors)
		return fmt.Sprintf("%v", v)
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
// This method is kept for backward compatibility but getColumnTypeUsingMetadata is preferred
func (s *Session) getColumnTypeFromSystemTable(keyspace, table, column string) string {
	if s.Session == nil {
		return ""
	}

	query := `SELECT type FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ? AND column_name = ?`

	var columnType string
	iter := s.Query(query, keyspace, table, column).Iter()
	if !iter.Scan(&columnType) {
		_ = iter.Close()
		return ""
	}
	_ = iter.Close()

	return columnType
}

// getColumnTypeUsingMetadata gets the full type definition for a column using gocql metadata API
func (s *Session) getColumnTypeUsingMetadata(keyspace, table, column string) string {
	if s.Session == nil {
		return ""
	}

	// Try to get table metadata
	tableMeta, err := s.GetTableMetadata(keyspace, table)
	if err != nil {
		// Fall back to system table approach
		return s.getColumnTypeFromSystemTable(keyspace, table, column)
	}

	// Look for the column in the metadata
	if colMeta, exists := tableMeta.Columns[column]; exists {
		// Log type information for debugging
		logger.DebugfToFile("getColumnTypeUsingMetadata", "Column %s: TypeInfo=%T, Type=%v",
			column, colMeta.Type, colMeta.Type.Type())

		typeStr := formatTypeInfo(colMeta.Type)

		// For UDT types, ensure we have the fully qualified name
		if colMeta.Type.Type() == gocql.TypeUDT || colMeta.Type.Type() == gocql.TypeCustom {
			// Try to cast to UDTTypeInfo
			if udtInfo, ok := colMeta.Type.(gocql.UDTTypeInfo); ok {
				// Return the fully qualified UDT name
				if udtInfo.Keyspace != "" {
					typeStr = fmt.Sprintf("%s.%s", udtInfo.Keyspace, udtInfo.Name)
				} else {
					typeStr = udtInfo.Name
				}
				logger.DebugfToFile("getColumnTypeUsingMetadata", "UDT cast successful: %s", typeStr)
			} else {
				logger.DebugfToFile("getColumnTypeUsingMetadata", "UDT cast failed for %s, falling back to system table", column)
				// Fall back to system table approach for UDT type name
				return s.getColumnTypeFromSystemTable(keyspace, table, column)
			}
		}
		return typeStr
	}

	// Column not found, fall back to system table
	return s.getColumnTypeFromSystemTable(keyspace, table, column)
}

// captureTracer implements gocql.Tracer to capture trace IDs
type captureTracer struct {
	traceID []byte
}

func (t *captureTracer) Trace(traceID []byte) {
	t.traceID = traceID
}

// ExecuteCQLQueryWithValues executes a statement with bind parameters.
//
// Use this for anything built from data rather than typed by the user: values
// travel to Cassandra as parameters instead of being pasted into the statement
// text, so they cannot alter it. It returns the same shapes as ExecuteCQLQuery,
// an error or a QueryResult, so callers can handle both the same way.
//
// This is for statements that do not return rows. SELECTs still go through
// ExecuteCQLQuery, which routes them to the streaming path.
func (s *Session) ExecuteCQLQueryWithValues(query string, values ...interface{}) interface{} {
	if s == nil || s.Session == nil {
		return fmt.Errorf("not connected to database")
	}

	if err := s.Query(query, values...).Exec(); err != nil {
		return err
	}
	return QueryResult{}
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

	// Get column info
	columns := iter.Columns()
	logger.DebugfToFile("executeSelectQuery", "Number of columns: %d", len(columns))

	// Check if this is a virtual table query (system_views)
	isVirtualTable := strings.Contains(strings.ToLower(query), "system_views.")
	if isVirtualTable {
		logger.DebugToFile("executeSelectQuery", "Detected virtual table query")
	}

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

	// Log column details and validate TypeInfo
	for i, col := range filteredColumns {
		if col.TypeInfo != nil {
			logger.DebugfToFile("executeSelectQuery", "Column %d: Name=%s, Type=%v, TypeInfo=%T",
				i, col.Name, col.TypeInfo.Type(), col.TypeInfo)
		} else {
			logger.DebugfToFile("executeSelectQuery", "Column %d: Name=%s has nil TypeInfo (virtual table?)",
				i, col.Name)
		}
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
	columnTypeInfos := make([]gocql.TypeInfo, len(filteredColumns))

	// For UDT columns, we need to get the full type definition from system tables
	queryKeyspace, tableName := extractTableName(query)
	currentKeyspace := queryKeyspace
	if currentKeyspace == "" {
		currentKeyspace = s.Keyspace()
	}

	for i, col := range filteredColumns {
		headers[i] = col.Name
		// Store the TypeInfo for proper type handling (especially UDTs)
		columnTypeInfos[i] = col.TypeInfo

		// Store the column type - use formatTypeInfo to get full type info including collection element types
		if col.TypeInfo == nil {
			columnTypes[i] = "unknown"
		} else {
			// Use formatTypeInfo for all columns to get proper type with element types
			fullType := formatTypeInfo(col.TypeInfo)

			// For UDTs, we might need additional metadata
			if col.TypeInfo.Type() == gocql.TypeUDT && currentKeyspace != "" && tableName != "" {
				// Try to get the UDT name from metadata if formatTypeInfo didn't get it
				if fullType == "udt" || fullType == "" {
					udtType := s.getColumnTypeUsingMetadata(currentKeyspace, tableName, col.Name)
					if udtType != "" {
						fullType = udtType
					}
				}
			}
			columnTypes[i] = fullType
		}

		headers[i] += keyColumns.Marker(col.Name)
	}

	// Collect results - use MapScan for better type handling
	results := [][]string{headers}
	rawData := make([]map[string]interface{}, 0)

	logger.DebugToFile("executeSelectQuery", "Starting row scan with MapScan...")

	// Extract clean column names (without PK/C indicators)
	cleanHeaders := make([]string, len(filteredColumns))
	for i, col := range filteredColumns {
		cleanHeaders[i] = col.Name
	}

	// MapScan for every table: scanning a NULL into an interface{} with Scan
	// panics, and MapScan leaves the key out of the map instead.
	scanned := make([][]string, 0)
	for {
		rowMap := make(map[string]interface{})
		if !iter.MapScan(rowMap) {
			break
		}

		// Convert map to row array in column order
		row := make([]string, len(filteredColumns))
		rawRow := make(map[string]interface{})

		for i, col := range filteredColumns {
			val, exists := rowMap[col.Name]
			if !exists {
				val = nil
			}
			rawRow[col.Name] = val
			row[i] = FormatValue(val)
		}

		scanned = append(scanned, row)
		rawData = append(rawData, rawRow)
	}
	results = append(results, scanned...)

	// Count what was collected. This used to be a counter incremented beside
	// the append, in the other half of an "if true ... else": the rows came
	// back from the half that runs and the count from the half that cannot, so
	// every non-streaming SELECT reported no rows however many it returned.
	rowNum := len(scanned)
	logger.DebugfToFile("executeSelectQuery", "Scan completed. Total rows: %d", rowNum)

	if err := iter.Close(); err != nil {
		logger.DebugfToFile("executeSelectQuery", "Iterator close error: %v", err)
		return fmt.Errorf("query failed: %v", err)
	}

	// Calculate query duration
	duration := time.Since(startTime)

	queryResult := QueryResult{
		Data:            results,
		RawData:         rawData,
		Duration:        duration,
		RowCount:        rowNum, // rowNum already contains the count of data rows (excluding header)
		ColumnTypes:     columnTypes,
		ColumnTypeInfos: columnTypeInfos,
		Headers:         cleanHeaders,
	}

	// Just pass the result, UI will handle formatting
	logger.DebugfToFile("ExecuteSelectQuery", "Returning QueryResult with %d rows", rowNum)

	return queryResult
}

// shouldUseStreaming determines if a query should use streaming based on heuristics
func (s *Session) shouldUseStreaming(query string) bool {
	// Always use streaming unless there's a small LIMIT
	upperQuery := strings.ToUpper(strings.TrimSpace(query))

	// A LIMIT small enough to fit in one page has nothing to page through, so
	// fetch it in one go.
	//
	// This used to compare the limit against a hardcoded 1000, which had no
	// relationship to the page size: PAGING 100 with LIMIT 300 fetched all 300
	// at once and PAGING was ignored. Measuring against the page size is what
	// makes PAGING mean something.
	if pageSize := s.PageSize(); pageSize > 0 && strings.Contains(upperQuery, " LIMIT ") {
		re := regexp.MustCompile(`LIMIT\s+(\d+)`)
		matches := re.FindStringSubmatch(upperQuery)
		if len(matches) > 1 {
			limit, err := strconv.Atoi(matches[1])
			if err == nil && limit <= pageSize {
				logger.DebugfToFile("shouldUseStreaming", "LIMIT %d fits in a page of %d, not using streaming", limit, pageSize)
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
	// Only set page size if it's greater than 0
	// Setting to 0 or not setting at all disables client-side paging
	if s.pageSize > 0 {
		q.PageSize(s.pageSize)
	}

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
	columnTypeInfos := make([]gocql.TypeInfo, len(filteredColumns))

	// For UDT columns, we need to get the full type definition from system tables
	queryKeyspace, tableName := extractTableName(query)
	currentKeyspace := queryKeyspace
	if currentKeyspace == "" {
		currentKeyspace = s.Keyspace()
	}

	for i, col := range filteredColumns {
		columnNames[i] = col.Name // Store original name
		headers[i] = col.Name     // Start with original name

		// Store the TypeInfo for proper type handling (especially UDTs)
		columnTypeInfos[i] = col.TypeInfo

		// Store the column type - use formatTypeInfo to get full type info including collection element types
		if col.TypeInfo == nil {
			columnTypes[i] = "unknown"
		} else {
			// Use formatTypeInfo for all columns to get proper type with element types
			fullType := formatTypeInfo(col.TypeInfo)

			// For UDTs, we might need additional metadata
			if col.TypeInfo.Type() == gocql.TypeUDT && currentKeyspace != "" && tableName != "" {
				// Try to get the UDT name from metadata if formatTypeInfo didn't get it
				if fullType == "udt" || fullType == "" {
					udtType := s.getColumnTypeUsingMetadata(currentKeyspace, tableName, col.Name)
					if udtType != "" {
						fullType = udtType
					}
				}
			}
			columnTypes[i] = fullType
		}

		headers[i] += keyColumns.Marker(col.Name)
	}

	// Return streaming result with iterator
	return StreamingQueryResult{
		Headers:         headers,
		ColumnNames:     columnNames,
		ColumnTypes:     columnTypes,
		ColumnTypeInfos: columnTypeInfos,
		Iterator:        iter,
		StartTime:       startTime,
		Keyspace:        currentKeyspace,
	}
}

// GetKeyColumns returns information about partition and clustering columns for a table
func (s *Session) GetKeyColumns(query string) KeyColumns {
	keyColumns := make(KeyColumns)

	// Try to extract table name from the SELECT query
	// Handle patterns like: SELECT ... FROM keyspace.table or FROM table,
	// with either part quoted.
	matches := fromClause.FindStringSubmatch(query)

	logger.DebugfToFile("getKeyColumns", "Query: %s", query)
	logger.DebugfToFile("getKeyColumns", "Regex matches: %v", matches)

	if len(matches) < 3 {
		logger.DebugToFile("getKeyColumns", "Could not extract table name from query")
		return keyColumns
	}

	keyspaceName := cqlIdentifier(matches[1]) // May be empty
	tableName := cqlIdentifier(matches[2])

	// An unqualified query means the keyspace the session is on, which is what
	// USE sets. Giving up here is why the markers only ever appeared when the
	// query happened to name the keyspace.
	if keyspaceName == "" {
		keyspaceName = s.Keyspace()
	}
	if keyspaceName == "" {
		logger.DebugToFile("getKeyColumns", "No keyspace in the query and none set on the session")
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
