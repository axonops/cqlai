package db

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/axonops/cqlai/internal/logger"
)

// SchemaCache provides schema information using gocql's metadata API
// This replaces the old implementation that maintained its own cache
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
	TableTokens map[string][]string // table -> tokens for fuzzy matching
}

// NewSchemaCache creates a new schema cache using gocql metadata
func NewSchemaCache(session *Session) *SchemaCache {
	return &SchemaCache{
		session:     session,
		Tables:      make(map[string][]CachedTableInfo),
		Columns:     make(map[string]map[string][]ColumnInfo),
		SearchIndex: &SearchIndex{
			TableTokens: make(map[string][]string),
		},
		Mu: sync.RWMutex{},
	}
}

// GetAllKeyspaces returns all non-system keyspaces
// Note: gocql doesn't provide a direct method to list all keyspaces,
// so we need to query system tables for this specific case
func (sc *SchemaCache) GetAllKeyspaces() ([]string, error) {
	if sc.session == nil || sc.session.Session == nil {
		return nil, fmt.Errorf("no session available")
	}

	// For listing all keyspaces, we still need to query system tables
	// as gocql requires knowing the keyspace name to get its metadata
	query := "SELECT keyspace_name FROM system_schema.keyspaces"
	iter := sc.session.Query(query).Iter()
	defer iter.Close()

	var keyspaces []string
	var keyspace string
	for iter.Scan(&keyspace) {
		// Skip system keyspaces
		if !strings.HasPrefix(keyspace, "system") {
			keyspaces = append(keyspaces, keyspace)
		}
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get keyspaces: %w", err)
	}

	return keyspaces, nil
}

// GetKeyspaceTables returns all tables for a specific keyspace using gocql metadata
func (sc *SchemaCache) GetKeyspaceTables(keyspace string) ([]CachedTableInfo, error) {
	if sc.session == nil || sc.session.Session == nil {
		return nil, fmt.Errorf("no session available")
	}

	// Use gocql's metadata API
	ksMetadata, err := sc.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	var tables []CachedTableInfo
	for tableName, tableMeta := range ksMetadata.Tables {
		// Convert gocql metadata to our format
		tableInfo := CachedTableInfo{
			TableInfo: TableInfo{
				KeyspaceName: keyspace,
				TableName:    tableName,
			},
			LastUpdated: time.Now(),
		}

		// Get partition and clustering keys
		for _, pk := range tableMeta.PartitionKey {
			tableInfo.PartitionKeys = append(tableInfo.PartitionKeys, pk.Name)
		}
		for _, ck := range tableMeta.ClusteringColumns {
			tableInfo.ClusteringKeys = append(tableInfo.ClusteringKeys, ck.Name)
		}

		tables = append(tables, tableInfo)
	}

	return tables, nil
}

// GetTableColumns returns columns for a specific table using gocql metadata
func (sc *SchemaCache) GetTableColumns(keyspace, table string) ([]ColumnInfo, error) {
	if sc.session == nil || sc.session.Session == nil {
		return nil, fmt.Errorf("no session available")
	}

	// Get table metadata from gocql
	tableMeta, err := sc.session.GetTableMetadata(keyspace, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}

	var columns []ColumnInfo

	// Process partition keys
	for i, pk := range tableMeta.PartitionKey {
		if colMeta, exists := tableMeta.Columns[pk.Name]; exists {
			columns = append(columns, ColumnInfo{
				Name:     pk.Name,
				DataType: formatTypeInfo(colMeta.Type),
				Kind:     "partition_key",
				Position: i,
			})
		}
	}

	// Process clustering keys
	for i, ck := range tableMeta.ClusteringColumns {
		if colMeta, exists := tableMeta.Columns[ck.Name]; exists {
			columns = append(columns, ColumnInfo{
				Name:     ck.Name,
				DataType: formatTypeInfo(colMeta.Type),
				Kind:     "clustering",
				Position: i,
			})
		}
	}

	// Process regular columns
	for colName, colMeta := range tableMeta.Columns {
		// Skip if already processed as key
		isKey := false
		for _, pk := range tableMeta.PartitionKey {
			if pk.Name == colName {
				isKey = true
				break
			}
		}
		if !isKey {
			for _, ck := range tableMeta.ClusteringColumns {
				if ck.Name == colName {
					isKey = true
					break
				}
			}
		}

		if !isKey {
			columns = append(columns, ColumnInfo{
				Name:     colName,
				DataType: formatTypeInfo(colMeta.Type),
				Kind:     "regular",
				Position: -1,
			})
		}
	}

	return columns, nil
}

// Refresh loads or refreshes the schema metadata
// With gocql's metadata API, this mainly updates the search index
func (sc *SchemaCache) Refresh() error {
	sc.Mu.Lock()
	defer sc.Mu.Unlock()

	logger.DebugToFile("SchemaCache", "Starting schema refresh using metadata API")

	// Get all keyspaces
	keyspaces, err := sc.GetAllKeyspaces()
	if err != nil {
		return fmt.Errorf("failed to get keyspaces: %w", err)
	}

	sc.Keyspaces = keyspaces

	// Clear and rebuild the search-related caches
	sc.Tables = make(map[string][]CachedTableInfo)
	sc.Columns = make(map[string]map[string][]ColumnInfo)
	sc.SearchIndex = &SearchIndex{
		TableTokens: make(map[string][]string),
	}

	// Populate tables and columns for each keyspace
	for _, ks := range keyspaces {
		tables, err := sc.GetKeyspaceTables(ks)
		if err != nil {
			logger.DebugfToFile("SchemaCache", "Failed to get tables for keyspace %s: %v", ks, err)
			continue
		}

		sc.Tables[ks] = tables

		// Initialize column map for keyspace
		if sc.Columns[ks] == nil {
			sc.Columns[ks] = make(map[string][]ColumnInfo)
		}

		// Get columns for each table
		for _, table := range tables {
			columns, err := sc.GetTableColumns(ks, table.TableName)
			if err != nil {
				logger.DebugfToFile("SchemaCache", "Failed to get columns for %s.%s: %v", ks, table.TableName, err)
				continue
			}

			sc.Columns[ks][table.TableName] = columns

			// Build search tokens for fuzzy matching
			tokens := buildSearchTokens(table.TableName)
			sc.SearchIndex.TableTokens[fmt.Sprintf("%s.%s", ks, table.TableName)] = tokens
		}
	}

	sc.LastRefresh = time.Now()
	logger.DebugfToFile("SchemaCache", "Schema refresh completed. Found %d keyspaces", len(keyspaces))

	return nil
}

// RefreshIfNeeded refreshes the cache if it's older than the specified duration
func (sc *SchemaCache) RefreshIfNeeded(maxAge time.Duration) error {
	sc.Mu.RLock()
	needsRefresh := time.Since(sc.LastRefresh) > maxAge
	sc.Mu.RUnlock()

	if needsRefresh {
		return sc.Refresh()
	}
	return nil
}

// GetTableInfo retrieves information about a specific table
func (sc *SchemaCache) GetTableInfo(keyspace, table string) (*TableInfo, error) {
	if sc.session == nil || sc.session.Session == nil {
		return nil, fmt.Errorf("no session available")
	}

	tableMeta, err := sc.session.GetTableMetadata(keyspace, table)
	if err != nil {
		return nil, err
	}

	tableInfo := &TableInfo{
		KeyspaceName: keyspace,
		TableName:    table,
	}

	// Get partition and clustering keys
	for _, pk := range tableMeta.PartitionKey {
		tableInfo.PartitionKeys = append(tableInfo.PartitionKeys, pk.Name)
	}
	for _, ck := range tableMeta.ClusteringColumns {
		tableInfo.ClusteringKeys = append(tableInfo.ClusteringKeys, ck.Name)
	}

	// Get columns
	tableInfo.Columns, err = sc.GetTableColumns(keyspace, table)
	if err != nil {
		return nil, err
	}

	return tableInfo, nil
}

// buildSearchTokens creates tokens for fuzzy searching
func buildSearchTokens(tableName string) []string {
	tokens := []string{strings.ToLower(tableName)}

	// Split on underscores and camelCase
	parts := strings.Split(tableName, "_")
	for _, part := range parts {
		if part != "" {
			tokens = append(tokens, strings.ToLower(part))
		}
	}

	// Add camelCase splits
	// Simple implementation - can be enhanced
	for i := 1; i < len(tableName); i++ {
		if tableName[i] >= 'A' && tableName[i] <= 'Z' {
			if i > 0 {
				tokens = append(tokens, strings.ToLower(tableName[:i]))
				tokens = append(tokens, strings.ToLower(tableName[i:]))
			}
		}
	}

	return tokens
}

// Compatibility methods to maintain the same interface

// GetCachedTableCount returns approximate count (not cached with metadata API)
func (sc *SchemaCache) GetCachedTableCount(keyspace string) int {
	sc.Mu.RLock()
	defer sc.Mu.RUnlock()

	if tables, exists := sc.Tables[keyspace]; exists {
		return len(tables)
	}
	return 0
}

// IsInitialized checks if the schema cache has been populated
func (sc *SchemaCache) IsInitialized() bool {
	sc.Mu.RLock()
	defer sc.Mu.RUnlock()
	return len(sc.Keyspaces) > 0
}

// CountTotalTables returns the total number of tables across all keyspaces
func (sc *SchemaCache) CountTotalTables() int {
	sc.Mu.RLock()
	defer sc.Mu.RUnlock()

	count := 0
	for _, tables := range sc.Tables {
		count += len(tables)
	}
	return count
}

// GetTableSchema returns schema information for a specific table
// This method is used by AI components for query context
func (sc *SchemaCache) GetTableSchema(keyspace, table string) (*TableSchema, error) {
	sc.Mu.RLock()
	defer sc.Mu.RUnlock()

	// Check if we have the table in our cache
	tables, ok := sc.Tables[keyspace]
	if !ok {
		return nil, fmt.Errorf("keyspace %s not found", keyspace)
	}

	// Find the table
	var foundTable *CachedTableInfo
	for _, t := range tables {
		if t.TableName == table {
			foundTable = &t
			break
		}
	}

	if foundTable == nil {
		return nil, fmt.Errorf("table %s.%s not found", keyspace, table)
	}

	// Get columns from cache
	columns, ok := sc.Columns[keyspace][table]
	if !ok {
		return nil, fmt.Errorf("columns for %s.%s not found", keyspace, table)
	}

	// Build TableSchema with proper types
	schema := &TableSchema{
		Keyspace:       keyspace,
		TableName:      table,
		PartitionKeys:  foundTable.PartitionKeys,
		ClusteringKeys: foundTable.ClusteringKeys,
	}

	// Convert ColumnInfo to ColumnSchema
	for _, col := range columns {
		schema.Columns = append(schema.Columns, ColumnSchema{
			Name:     col.Name,
			Type:     col.DataType,
			Kind:     col.Kind,
			Position: col.Position,
		})
	}

	return schema, nil
}