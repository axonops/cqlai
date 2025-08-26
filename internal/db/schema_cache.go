package db

import (
	"fmt"
	"sync"
	"time"

	"github.com/axonops/cqlai/internal/logger"
)

// SchemaCache maintains an in-memory cache of Cassandra schema metadata
type SchemaCache struct {
	Keyspaces   []string
	Tables      map[string][]CachedTableInfo        // keyspace -> tables
	Columns     map[string]map[string][]ColumnInfo // keyspace -> table -> columns
	SearchIndex *SearchIndex                       // Pre-computed fuzzy search index
	LastRefresh time.Time
	Mu          sync.RWMutex
	session     *Session
}

// CachedTableInfo extends TableInfo with cache-specific fields
type CachedTableInfo struct {
	TableInfo
	RowCount    int64 // Optional: cached approximate count
	LastUpdated time.Time
}

// SearchIndex contains pre-computed data for fuzzy searching
type SearchIndex struct {
	TableTokens  map[string][]string  // table -> tokens for fuzzy matching
}

// NewSchemaCache creates a new schema cache
func NewSchemaCache(session *Session) *SchemaCache {
	return &SchemaCache{
		session:     session,
		Tables:      make(map[string][]CachedTableInfo),
		Columns:     make(map[string]map[string][]ColumnInfo),
		SearchIndex: &SearchIndex{
			TableTokens:  make(map[string][]string),
		},
	}
}

// executeSchemaQuery executes a query and always returns all data (non-streaming)
func (sc *SchemaCache) executeSchemaQuery(query string) ([][]string, error) {
	iter := sc.session.Query(query).Iter()
	defer iter.Close()
	
	// Get column info
	columns := iter.Columns()
	if len(columns) == 0 {
		return nil, fmt.Errorf("no columns returned")
	}
	
	// Prepare result data
	var data [][]string
	
	// Add header row
	header := make([]string, len(columns))
	for i, col := range columns {
		header[i] = col.Name
	}
	data = append(data, header)
	
	// Fetch all rows
	values := make([]interface{}, len(columns))
	scanDest := make([]interface{}, len(columns))
	for i := range values {
		scanDest[i] = &values[i]
	}
	
	for iter.Scan(scanDest...) {
		row := make([]string, len(columns))
		for i, val := range values {
			if val != nil {
				row[i] = fmt.Sprintf("%v", val)
			} else {
				row[i] = ""
			}
		}
		data = append(data, row)
	}
	
	if err := iter.Close(); err != nil {
		return nil, err
	}
	
	return data, nil
}

// Refresh updates the schema cache from system_schema
func (sc *SchemaCache) Refresh() error {
	sc.Mu.Lock()
	defer sc.Mu.Unlock()

	logger.DebugfToFile("SchemaCache", "Starting schema refresh")

	// Query keyspaces
	keyspaceQuery := "SELECT keyspace_name FROM system_schema.keyspaces"
	keyspaceData, err := sc.executeSchemaQuery(keyspaceQuery)
	if err != nil {
		logger.DebugfToFile("SchemaCache", "Error querying keyspaces: %v", err)
		return fmt.Errorf("failed to query keyspaces: %w", err)
	}
	logger.DebugfToFile("SchemaCache", "Keyspace query returned %d rows", len(keyspaceData))
	
	if len(keyspaceData) > 1 {
		sc.Keyspaces = []string{}
		for _, row := range keyspaceData[1:] { // Skip header row
			if len(row) > 0 {
				sc.Keyspaces = append(sc.Keyspaces, row[0])
				logger.DebugfToFile("SchemaCache", "Added keyspace: %s", row[0])
			}
		}
		logger.DebugfToFile("SchemaCache", "Total keyspaces cached: %d", len(sc.Keyspaces))
	}

	// Query tables
	tableQuery := "SELECT keyspace_name, table_name FROM system_schema.tables"
	tableData, err := sc.executeSchemaQuery(tableQuery)
	if err != nil {
		logger.DebugfToFile("SchemaCache", "Error querying tables: %v", err)
		return fmt.Errorf("failed to query tables: %w", err)
	}
	logger.DebugfToFile("SchemaCache", "Table query returned %d rows", len(tableData))
	
	if len(tableData) > 1 {
		// Clear existing tables
		sc.Tables = make(map[string][]CachedTableInfo)
		
		for _, row := range tableData[1:] { // Skip header row
			if len(row) >= 2 {
				keyspace := row[0]
				tableName := row[1]
				
				tableInfo := CachedTableInfo{
					TableInfo: TableInfo{
						KeyspaceName: keyspace,
						TableName:   tableName,
					},
					LastUpdated: time.Now(),
				}
				
				sc.Tables[keyspace] = append(sc.Tables[keyspace], tableInfo)
				logger.DebugfToFile("SchemaCache", "Added table: %s.%s", keyspace, tableName)
			}
		}
		logger.DebugfToFile("SchemaCache", "Total tables cached: %d keyspaces", len(sc.Tables))
	}

	// Query columns
	columnQuery := "SELECT keyspace_name, table_name, column_name, kind, type, position FROM system_schema.columns"
	columnData, err := sc.executeSchemaQuery(columnQuery)
	if err != nil {
		logger.DebugfToFile("SchemaCache", "Error querying columns: %v", err)
		return fmt.Errorf("failed to query columns: %w", err)
	}
	logger.DebugfToFile("SchemaCache", "Column query returned %d rows", len(columnData))
	
	if len(columnData) > 1 {
		// Clear existing columns
		sc.Columns = make(map[string]map[string][]ColumnInfo)
		
		for _, row := range columnData[1:] { // Skip header row
			if len(row) >= 6 {
				keyspace := row[0]
				tableName := row[1]
				columnName := row[2]
				kind := row[3]
				columnType := row[4]
				position := 0
				if row[5] != "" {
					fmt.Sscanf(row[5], "%d", &position)
				}
				
				columnInfo := ColumnInfo{
					Name:     columnName,
					DataType: columnType,
					Kind:     kind,
					Position: position,
				}
				
				if sc.Columns[keyspace] == nil {
					sc.Columns[keyspace] = make(map[string][]ColumnInfo)
				}
				
				sc.Columns[keyspace][tableName] = append(sc.Columns[keyspace][tableName], columnInfo)
			}
		}
		logger.DebugfToFile("SchemaCache", "Total columns cached across all tables")
	}

	// Build search index
	sc.buildSearchIndex()
	
	sc.LastRefresh = time.Now()
	logger.DebugfToFile("SchemaCache", "Schema refresh completed: %d keyspaces, %d table groups", 
		len(sc.Keyspaces), len(sc.Tables))

	return nil
}

// buildSearchIndex creates tokens for fuzzy searching
func (sc *SchemaCache) buildSearchIndex() {
	sc.SearchIndex.TableTokens = make(map[string][]string)
	
	for keyspace, tables := range sc.Tables {
		for _, table := range tables {
			tableKey := fmt.Sprintf("%s.%s", keyspace, table.TableName)
			
			// Generate tokens from table name only
			tokens := tokenize(table.TableName)
			sc.SearchIndex.TableTokens[tableKey] = tokens
		}
	}
}

// GetTableSchema returns the schema for a specific table
func (sc *SchemaCache) GetTableSchema(keyspace, table string) (*TableSchema, error) {
	sc.Mu.RLock()
	defer sc.Mu.RUnlock()
	
	columns, ok := sc.Columns[keyspace][table]
	if !ok {
		return nil, fmt.Errorf("table %s.%s not found", keyspace, table)
	}
	
	// Convert to ColumnSchema format expected by TableSchema
	columnSchemas := make([]ColumnSchema, len(columns))
	partitionKeys := []string{}
	clusteringKeys := []string{}
	
	for i, col := range columns {
		columnSchemas[i] = ColumnSchema{
			Name:     col.Name,
			Type:     col.DataType,
			Kind:     col.Kind,
			Position: col.Position,
		}
		
		if col.Kind == "partition_key" {
			partitionKeys = append(partitionKeys, col.Name)
		} else if col.Kind == "clustering" {
			clusteringKeys = append(clusteringKeys, col.Name)
		}
	}
	
	schema := &TableSchema{
		Keyspace:       keyspace,
		TableName:      table,
		Columns:        columnSchemas,
		PartitionKeys:  partitionKeys,
		ClusteringKeys: clusteringKeys,
	}
	
	return schema, nil
}

// CountTotalTables returns the total number of tables across all keyspaces
func (sc *SchemaCache) CountTotalTables() int {
	count := 0
	for _, tables := range sc.Tables {
		count += len(tables)
	}
	return count
}

// tokenize splits a string into tokens for fuzzy matching
func tokenize(s string) []string {
	// Simple tokenization - can be enhanced
	tokens := []string{s}
	
	// Add split by underscore
	if parts := splitByDelimiter(s, '_'); len(parts) > 1 {
		tokens = append(tokens, parts...)
	}
	
	// Add split by camelCase
	if parts := splitCamelCase(s); len(parts) > 1 {
		tokens = append(tokens, parts...)
	}
	
	return tokens
}

// Helper functions
func containsAny(s string, patterns []string) bool {
	for _, pattern := range patterns {
		if contains(s, pattern) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		   findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func splitByDelimiter(s string, delimiter rune) []string {
	var parts []string
	var current []rune
	
	for _, r := range s {
		if r == delimiter {
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = []rune{}
			}
		} else {
			current = append(current, r)
		}
	}
	
	if len(current) > 0 {
		parts = append(parts, string(current))
	}
	
	return parts
}

func splitCamelCase(s string) []string {
	var parts []string
	var current []rune
	
	for i, r := range s {
		if i > 0 && isUpper(r) && !isUpper(rune(s[i-1])) {
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = []rune{}
			}
		}
		current = append(current, r)
	}
	
	if len(current) > 0 {
		parts = append(parts, string(current))
	}
	
	return parts
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func deduplicate(strings []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, s := range strings {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	
	return result
}