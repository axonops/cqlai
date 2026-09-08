package db

import (
	"fmt"
	"sort"
	"strings"

	"github.com/axonops/cqlai/internal/logger"
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
			_ = iter.Close()
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

	// Get columns.
	//
	// Deliberately no ORDER BY: system_schema.columns is clustered by
	// column_name, and Cassandra rejects ordering by anything else with
	// "Order by is currently only supported on the clustered columns of the
	// PRIMARY KEY". Asking for ORDER BY position failed every call, and the
	// caller swallowed the error, so no table was ever loaded. Sort in Go
	// instead.
	columnQuery := `
		SELECT column_name, type, kind, position
		FROM system_schema.columns
		WHERE keyspace_name = ? AND table_name = ?`

	iter := s.Query(columnQuery, keyspace, table).Iter()

	var colName, colType, colKind string
	var position int

	for iter.Scan(&colName, &colType, &colKind, &position) {
		ts.Columns = append(ts.Columns, ColumnSchema{
			Name:     colName,
			Type:     colType,
			Kind:     colKind,
			Position: position,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to retrieve columns: %v", err)
	}

	// Partition keys first, then clustering keys, then the rest; within a kind,
	// by position.
	kindPriority := map[string]int{"partition_key": 0, "clustering": 1, "regular": 2}
	sort.SliceStable(ts.Columns, func(i, j int) bool {
		iPriority, jPriority := kindPriority[ts.Columns[i].Kind], kindPriority[ts.Columns[j].Kind]
		if iPriority != jPriority {
			return iPriority < jPriority
		}
		return ts.Columns[i].Position < ts.Columns[j].Position
	})

	// Collect the key names only once sorted. Taking them during the scan gives
	// alphabetical order, which describes a different table: partition key
	// order decides how rows are distributed, clustering order how they sort.
	for _, col := range ts.Columns {
		switch col.Kind {
		case "partition_key":
			ts.PartitionKeys = append(ts.PartitionKeys, col.Name)
		case "clustering":
			ts.ClusteringKeys = append(ts.ClusteringKeys, col.Name)
		}
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
			// Log rather than skip in silence. A whole schema disappearing
			// with no message is why the broken query above went unnoticed.
			logger.DebugfToFile("Schema", "Skipping table %s.%s: %v", ks.Name, tableName, err)
			continue
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

		fmt.Fprintf(&sb, "Keyspace: %s\n", ksName)

		for tableName, table := range ks.Tables {
			if count >= limit {
				break
			}

			fmt.Fprintf(&sb, "  Table: %s\n", tableName)
			sb.WriteString("    Columns:\n")

			for _, col := range table.Columns {
				marker := ""
				switch col.Kind {
				case "partition_key":
					marker = " (PK)"
				case "clustering":
					marker = " (CK)"
				}
				fmt.Fprintf(&sb, "      - %s: %s%s\n", col.Name, col.Type, marker)
			}
			count++
		}
	}

	return sb.String(), nil
}
