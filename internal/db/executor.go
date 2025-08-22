package db

import (
	"fmt"
	"regexp"
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
	
	// Log column details
	for i, col := range columns {
		logger.DebugfToFile("executeSelectQuery", "Column %d: Name=%s, Type=%v, TypeInfo=%T", 
			i, col.Name, col.TypeInfo.Type(), col.TypeInfo)
	}
	
	if len(columns) == 0 {
		if err := iter.Close(); err != nil {
			logger.DebugfToFile("executeSelectQuery", "Error closing empty iterator: %v", err)
			return fmt.Errorf("query failed: %v", err)
		}
		return "No results"
	}

	// Get key column information
	keyColumns := s.GetKeyColumns(query)
	
	// Prepare headers with key indicators and collect column types
	headers := make([]string, len(columns))
	columnTypes := make([]string, len(columns))
	for i, col := range columns {
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
		
		logger.DebugfToFile("executeSelectQuery", "Successfully scanned row %d", rowNum)
		row := make([]string, len(columns))
		
		for i, col := range columns {
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
	
	// Return QueryResult with metadata
	return QueryResult{
		Data:        results,
		Duration:    duration,
		RowCount:    rowNum, // rowNum already contains the count of data rows (excluding header)
		ColumnTypes: columnTypes,
	}
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