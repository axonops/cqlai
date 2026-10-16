package db

import (
	"fmt"

	"github.com/axonops/cqlai/internal/session"
)

// TypeInfo holds user-defined type information for manual describe
type TypeInfo struct {
	Name       string
	FieldNames []string
	FieldTypes []string
}

// TypeListInfo holds type list information for manual describe
type TypeListInfo struct {
	Name string
}

// DescribeTypesQuery executes the query to list all types (for pre-4.0)
func (s *Session) DescribeTypesQuery(keyspace string) ([]TypeListInfo, error) {
	query := `SELECT type_name FROM system_schema.types WHERE keyspace_name = ?`
	iter := s.Query(query, keyspace).Iter()

	var types []TypeListInfo
	var typeName string

	for iter.Scan(&typeName) {
		types = append(types, TypeListInfo{
			Name: typeName,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return types, nil
}

// DescribeTypeQuery executes the query to get type information (for pre-4.0)
func (s *Session) DescribeTypeQuery(keyspace string, typeName string) (*TypeInfo, error) {
	query := `SELECT type_name, field_names, field_types 
	          FROM system_schema.types 
	          WHERE keyspace_name = ? AND type_name = ?`

	iter := s.Query(query, keyspace, typeName).Iter()

	var name string
	var fieldNames []string
	var fieldTypes []string

	if !iter.Scan(&name, &fieldNames, &fieldTypes) {
		_ = iter.Close()
		return nil, fmt.Errorf("type '%s' not found in keyspace '%s'", typeName, keyspace)
	}
	_ = iter.Close()

	return &TypeInfo{
		Name:       name,
		FieldNames: fieldNames,
		FieldTypes: fieldTypes,
	}, nil
}

// DBDescribeTypes handles version detection and returns appropriate data
func (s *Session) DBDescribeTypes(sessionMgr *session.Manager) (interface{}, []TypeListInfo, error) {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	if s.IsVersion4OrHigher() {
		// Use server-side DESCRIBE TYPES
		result := s.ExecuteCQLQuery("DESCRIBE TYPES")
		return result, nil, nil // Server-side result, no TypeListInfo needed
	}

	// Fall back to manual construction for pre-4.0
	currentKeyspace := ""
	if sessionMgr != nil {
		currentKeyspace = sessionMgr.CurrentKeyspace()
	}
	if currentKeyspace == "" {
		return nil, nil, fmt.Errorf("no keyspace selected")
	}

	types, err := s.DescribeTypesQuery(currentKeyspace)
	if err != nil {
		return nil, nil, err
	}

	return nil, types, nil // Manual query result, return TypeListInfo for formatting
}

// DBDescribeType handles version detection and returns appropriate data
func (s *Session) DBDescribeType(sessionMgr *session.Manager, typeName string) (interface{}, *TypeInfo, error) {
	keyspace, name, err := qualify(sessionMgr, typeName)
	if err != nil {
		return nil, nil, err
	}

	// Cassandra 4.0 and later describe it themselves.
	if s.IsVersion4OrHigher() {
		return s.ExecuteCQLQuery("DESCRIBE TYPE " + quoteQualified(keyspace, name)), nil, nil
	}

	typeInfo, err := s.DescribeTypeQuery(keyspace, name)
	if err != nil {
		return nil, nil, err
	}
	return nil, typeInfo, nil
}
