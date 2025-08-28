package db

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/session"
)

// TableSchema represents a table's schema information
type TableSchema struct {
	Keyspace       string
	TableName      string
	Columns        []ColumnSchema
	PartitionKeys  []string
	ClusteringKeys []string
}

// ColumnSchema represents a column's schema
type ColumnSchema struct {
	Name     string
	Type     string
	Kind     string // 'partition_key', 'clustering', 'regular', 'static'
	Position int    // Position for partition/clustering keys
}

// KeyspaceSchema represents a keyspace's schema
type KeyspaceSchema struct {
	Name   string
	Tables map[string]*TableSchema
}

// SchemaCatalog holds all schema information
type SchemaCatalog struct {
	Keyspaces map[string]*KeyspaceSchema
}

// GetSchemaCatalog retrieves the complete schema catalog from Cassandra
func (s *Session) GetSchemaCatalog() (*SchemaCatalog, error) {
	catalog := &SchemaCatalog{
		Keyspaces: make(map[string]*KeyspaceSchema),
	}

	// Get all keyspaces
	keyspaceQuery := `SELECT keyspace_name FROM system_schema.keyspaces`
	iter := s.Query(keyspaceQuery).Iter()
	
	var keyspaceName string
	for iter.Scan(&keyspaceName) {
		// Skip system keyspaces unless explicitly requested
		if strings.HasPrefix(keyspaceName, "system") {
			continue
		}
		
		ks := &KeyspaceSchema{
			Name:   keyspaceName,
			Tables: make(map[string]*TableSchema),
		}
		
		// Get tables for this keyspace
		if err := s.loadTablesForKeyspace(ks); err != nil {
			iter.Close()
			return nil, fmt.Errorf("failed to load tables for keyspace %s: %v", keyspaceName, err)
		}
		
		catalog.Keyspaces[keyspaceName] = ks
	}
	
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to retrieve keyspaces: %v", err)
	}
	
	return catalog, nil
}

// GetKeyspaceSchema retrieves schema for a specific keyspace
func (s *Session) GetKeyspaceSchema(keyspace string) (*KeyspaceSchema, error) {
	ks := &KeyspaceSchema{
		Name:   keyspace,
		Tables: make(map[string]*TableSchema),
	}
	
	if err := s.loadTablesForKeyspace(ks); err != nil {
		return nil, fmt.Errorf("failed to load tables for keyspace %s: %v", keyspace, err)
	}
	
	return ks, nil
}

// GetTableSchema retrieves schema for a specific table
func (s *Session) GetTableSchema(keyspace, table string) (*TableSchema, error) {
	ts := &TableSchema{
		Keyspace:       keyspace,
		TableName:      table,
		Columns:        []ColumnSchema{},
		PartitionKeys:  []string{},
		ClusteringKeys: []string{},
	}
	
	// Get columns
	columnQuery := `
		SELECT column_name, type, kind, position 
		FROM system_schema.columns 
		WHERE keyspace_name = ? AND table_name = ?
		ORDER BY position`
	
	iter := s.Query(columnQuery, keyspace, table).Iter()
	
	var colName, colType, colKind string
	var position int
	
	for iter.Scan(&colName, &colType, &colKind, &position) {
		col := ColumnSchema{
			Name:     colName,
			Type:     colType,
			Kind:     colKind,
			Position: position,
		}
		ts.Columns = append(ts.Columns, col)
		
		// Track partition and clustering keys
		switch colKind {
		case "partition_key":
			ts.PartitionKeys = append(ts.PartitionKeys, colName)
		case "clustering":
			ts.ClusteringKeys = append(ts.ClusteringKeys, colName)
		}
	}
	
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to retrieve columns: %v", err)
	}
	
	if len(ts.Columns) == 0 {
		return nil, fmt.Errorf("table %s.%s not found", keyspace, table)
	}
	
	return ts, nil
}

// GetCurrentKeyspaceSchema retrieves schema for the current keyspace
func (s *Session) GetCurrentKeyspaceSchema(sessionMgr *session.Manager) (*KeyspaceSchema, error) {
	currentKeyspace := ""
	if sessionMgr != nil {
		currentKeyspace = sessionMgr.CurrentKeyspace()
	}
	if currentKeyspace == "" {
		return nil, fmt.Errorf("no keyspace selected")
	}
	return s.GetKeyspaceSchema(currentKeyspace)
}

// loadTablesForKeyspace loads all tables for a keyspace
func (s *Session) loadTablesForKeyspace(ks *KeyspaceSchema) error {
	tableQuery := `SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?`
	iter := s.Query(tableQuery, ks.Name).Iter()
	
	var tableName string
	for iter.Scan(&tableName) {
		ts, err := s.GetTableSchema(ks.Name, tableName)
		if err != nil {
			continue // Skip tables we can't load
		}
		ks.Tables[tableName] = ts
	}
	
	return iter.Close()
}

// GetSchemaContext returns a compact representation for AI context
func (s *Session) GetSchemaContext(limit int) (string, error) {
	catalog, err := s.GetSchemaCatalog()
	if err != nil {
		return "", err
	}
	
	var sb strings.Builder
	count := 0
	
	for ksName, ks := range catalog.Keyspaces {
		if count >= limit {
			break
		}
		
		sb.WriteString(fmt.Sprintf("Keyspace: %s\n", ksName))
		
		for tableName, table := range ks.Tables {
			if count >= limit {
				break
			}
			
			sb.WriteString(fmt.Sprintf("  Table: %s\n", tableName))
			sb.WriteString("    Columns:\n")
			
			for _, col := range table.Columns {
				marker := ""
				if col.Kind == "partition_key" {
					marker = " (PK)"
				} else if col.Kind == "clustering" {
					marker = " (CK)"
				}
				sb.WriteString(fmt.Sprintf("      - %s: %s%s\n", col.Name, col.Type, marker))
			}
			count++
		}
	}
	
	return sb.String(), nil
}