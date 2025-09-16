package db

import (
	"fmt"
	"strings"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/session"
)

// GetKeyspaceMetadata retrieves keyspace metadata using gocql's built-in API
func (s *Session) GetKeyspaceMetadata(keyspace string) (*gocql.KeyspaceMetadata, error) {
	metadata, err := s.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}
	return metadata, nil
}

// GetTableMetadata retrieves table metadata for a specific table
func (s *Session) GetTableMetadata(keyspace, table string) (*gocql.TableMetadata, error) {
	ksMetadata, err := s.GetKeyspaceMetadata(keyspace)
	if err != nil {
		return nil, err
	}

	tableMetadata, ok := ksMetadata.Tables[table]
	if !ok {
		return nil, fmt.Errorf("table %s.%s not found", keyspace, table)
	}

	return tableMetadata, nil
}

// GetUserTypeMetadata retrieves UDT metadata for a specific type
func (s *Session) GetUserTypeMetadata(keyspace, typeName string) (*gocql.UserTypeMetadata, error) {
	ksMetadata, err := s.GetKeyspaceMetadata(keyspace)
	if err != nil {
		return nil, err
	}

	udtMetadata, ok := ksMetadata.UserTypes[typeName]
	if !ok {
		return nil, fmt.Errorf("user type %s.%s not found", keyspace, typeName)
	}

	return udtMetadata, nil
}

// GetMaterializedViewMetadata retrieves materialized view metadata
func (s *Session) GetMaterializedViewMetadata(keyspace, viewName string) (*gocql.MaterializedViewMetadata, error) {
	ksMetadata, err := s.GetKeyspaceMetadata(keyspace)
	if err != nil {
		return nil, err
	}

	mvMetadata, ok := ksMetadata.MaterializedViews[viewName]
	if !ok {
		return nil, fmt.Errorf("materialized view %s.%s not found", keyspace, viewName)
	}

	return mvMetadata, nil
}

// LoadKeyspaceUDTsUsingMetadata delegates to the registry's simplified method
func (s *Session) LoadKeyspaceUDTsUsingMetadata(keyspace string) error {
	if s.udtRegistry == nil {
		return fmt.Errorf("UDT registry not initialized")
	}
	// The new simplified registry uses gocql's cache directly
	return s.udtRegistry.LoadKeyspaceUDTsUsingMetadata(keyspace)
}

// formatTypeInfo converts gocql.TypeInfo to a string representation
func formatTypeInfo(typeInfo gocql.TypeInfo) string {
	if typeInfo == nil {
		return "unknown"
	}

	// Check the base type first
	baseType := typeInfo.Type()

	// Handle simple types
	switch baseType {
	case gocql.TypeList, gocql.TypeSet, gocql.TypeMap:
		// Handle collection types
		if collType, ok := typeInfo.(gocql.CollectionType); ok {
			switch collType.Type() {
			case gocql.TypeList:
				return fmt.Sprintf("list<%s>", formatTypeInfo(collType.Elem))
			case gocql.TypeSet:
				return fmt.Sprintf("set<%s>", formatTypeInfo(collType.Elem))
			case gocql.TypeMap:
				return fmt.Sprintf("map<%s, %s>", formatTypeInfo(collType.Key), formatTypeInfo(collType.Elem))
			}
		}
	case gocql.TypeTuple:
		// Handle tuple types
		if tupleType, ok := typeInfo.(gocql.TupleTypeInfo); ok {
			var elements []string
			for _, elem := range tupleType.Elems {
				elements = append(elements, formatTypeInfo(elem))
			}
			return fmt.Sprintf("tuple<%s>", strings.Join(elements, ", "))
		}
	case gocql.TypeUDT:
		// Handle UDT types
		if udtType, ok := typeInfo.(gocql.UDTTypeInfo); ok {
			// Check if it's a frozen type
			udtName := udtType.Name
			if udtType.Keyspace != "" {
				udtName = fmt.Sprintf("%s.%s", udtType.Keyspace, udtType.Name)
			}
			// Note: gocql doesn't expose frozen status directly in TypeInfo,
			// but we can still return the UDT name which is what we need
			return udtName
		}
	default:
		// Handle native types
		return typeNameFromType(baseType)
	}

	return typeNameFromType(baseType)
}

// typeNameFromType converts gocql.Type to string name
func typeNameFromType(t gocql.Type) string {
	switch t {
	case gocql.TypeCustom:
		return "custom"
	case gocql.TypeAscii:
		return "ascii"
	case gocql.TypeBigInt:
		return "bigint"
	case gocql.TypeBlob:
		return "blob"
	case gocql.TypeBoolean:
		return "boolean"
	case gocql.TypeCounter:
		return "counter"
	case gocql.TypeDecimal:
		return "decimal"
	case gocql.TypeDouble:
		return "double"
	case gocql.TypeFloat:
		return "float"
	case gocql.TypeInt:
		return "int"
	case gocql.TypeText:
		return "text"
	case gocql.TypeTimestamp:
		return "timestamp"
	case gocql.TypeUUID:
		return "uuid"
	case gocql.TypeVarchar:
		return "varchar"
	case gocql.TypeVarint:
		return "varint"
	case gocql.TypeTimeUUID:
		return "timeuuid"
	case gocql.TypeInet:
		return "inet"
	case gocql.TypeDate:
		return "date"
	case gocql.TypeDuration:
		return "duration"
	case gocql.TypeTime:
		return "time"
	case gocql.TypeSmallInt:
		return "smallint"
	case gocql.TypeTinyInt:
		return "tinyint"
	case gocql.TypeList:
		return "list"
	case gocql.TypeMap:
		return "map"
	case gocql.TypeSet:
		return "set"
	case gocql.TypeTuple:
		return "tuple"
	case gocql.TypeUDT:
		return "udt"
	default:
		return "unknown"
	}
}

// GetTableSchemaUsingMetadata retrieves table schema using gocql metadata
func (s *Session) GetTableSchemaUsingMetadata(keyspace, table string) (*TableSchema, error) {
	tableMeta, err := s.GetTableMetadata(keyspace, table)
	if err != nil {
		return nil, err
	}

	ts := &TableSchema{
		Keyspace:       keyspace,
		TableName:      table,
		Columns:        []ColumnSchema{},
		PartitionKeys:  []string{},
		ClusteringKeys: []string{},
	}

	// Process partition keys
	for _, pk := range tableMeta.PartitionKey {
		ts.PartitionKeys = append(ts.PartitionKeys, pk.Name)
	}

	// Process clustering keys
	for _, ck := range tableMeta.ClusteringColumns {
		ts.ClusteringKeys = append(ts.ClusteringKeys, ck.Name)
	}

	// Process all columns
	for name, col := range tableMeta.Columns {
		colSchema := ColumnSchema{
			Name: name,
			Type: formatTypeInfo(col.Type),
		}

		// Determine column kind
		isPartitionKey := false
		isClusteringKey := false

		for _, pk := range tableMeta.PartitionKey {
			if pk.Name == name {
				isPartitionKey = true
				colSchema.Kind = "partition_key"
				colSchema.Position = pk.ComponentIndex
				break
			}
		}

		if !isPartitionKey {
			for _, ck := range tableMeta.ClusteringColumns {
				if ck.Name == name {
					isClusteringKey = true
					colSchema.Kind = "clustering"
					colSchema.Position = ck.ComponentIndex
					break
				}
			}
		}

		if !isPartitionKey && !isClusteringKey {
			// Check if it's a static column
			// Note: gocql doesn't directly expose static columns in TableMetadata
			// We might need to keep using system tables for this specific information
			colSchema.Kind = "regular"
		}

		ts.Columns = append(ts.Columns, colSchema)
	}

	return ts, nil
}

// GetKeyspaceSchemaUsingMetadata retrieves keyspace schema using gocql metadata
func (s *Session) GetKeyspaceSchemaUsingMetadata(keyspace string) (*KeyspaceSchema, error) {
	ksMetadata, err := s.GetKeyspaceMetadata(keyspace)
	if err != nil {
		return nil, err
	}

	ks := &KeyspaceSchema{
		Name:   keyspace,
		Tables: make(map[string]*TableSchema),
	}

	// Load all tables
	for tableName := range ksMetadata.Tables {
		ts, err := s.GetTableSchemaUsingMetadata(keyspace, tableName)
		if err != nil {
			continue // Skip tables we can't load
		}
		ks.Tables[tableName] = ts
	}

	return ks, nil
}

// GetCurrentKeyspaceSchemaUsingMetadata retrieves schema for the current keyspace using metadata
func (s *Session) GetCurrentKeyspaceSchemaUsingMetadata(sessionMgr *session.Manager) (*KeyspaceSchema, error) {
	currentKeyspace := ""
	if sessionMgr != nil {
		currentKeyspace = sessionMgr.CurrentKeyspace()
	}
	if currentKeyspace == "" {
		return nil, fmt.Errorf("no keyspace selected")
	}
	return s.GetKeyspaceSchemaUsingMetadata(currentKeyspace)
}