package router

import (
	"fmt"
	"strings"
	"github.com/axonops/cqlai/internal/logger"
)

// describeTypes lists all user-defined types
func (v *CqlCommandVisitorImpl) describeTypes() interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeTypes", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Use server-side DESCRIBE TYPES
		logger.DebugToFile("describeTypes", "Using server-side DESCRIBE TYPES")
		return v.session.ExecuteCQLQuery("DESCRIBE TYPES")
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := v.session.CurrentKeyspace()
	if currentKeyspace == "" {
		return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
	}

	query := `SELECT type_name FROM system_schema.types WHERE keyspace_name = ?`
	iter := v.session.Query(query, currentKeyspace).Iter()

	results := [][]string{{"Type Name"}}
	var typeName string

	for iter.Scan(&typeName) {
		results = append(results, []string{typeName})
	}
	iter.Close()

	if len(results) == 1 {
		return fmt.Sprintf("No user-defined types in keyspace %s", currentKeyspace)
	}

	return results
}

// describeType shows detailed information about a user-defined type
func (v *CqlCommandVisitorImpl) describeType(typeName string) interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeType", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Parse keyspace.type or just type
		var describeCmd string
		if strings.Contains(typeName, ".") {
			describeCmd = fmt.Sprintf("DESCRIBE TYPE %s", typeName)
		} else {
			currentKeyspace := v.session.CurrentKeyspace()
			if currentKeyspace == "" {
				return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
			}
			describeCmd = fmt.Sprintf("DESCRIBE TYPE %s.%s", currentKeyspace, typeName)
		}
		
		logger.DebugToFile("describeType", fmt.Sprintf("Using server-side: %s", describeCmd))
		return v.session.ExecuteCQLQuery(describeCmd)
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := v.session.CurrentKeyspace()
	if currentKeyspace == "" {
		return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
	}

	query := `SELECT type_name, field_names, field_types 
	          FROM system_schema.types 
	          WHERE keyspace_name = ? AND type_name = ?`

	iter := v.session.Query(query, currentKeyspace, typeName).Iter()

	var name string
	var fieldNames []string
	var fieldTypes []string

	if !iter.Scan(&name, &fieldNames, &fieldTypes) {
		iter.Close()
		return fmt.Sprintf("Type '%s' not found in keyspace '%s'", typeName, currentKeyspace)
	}
	iter.Close()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("CREATE TYPE %s.%s (\n", currentKeyspace, typeName))

	for i := 0; i < len(fieldNames) && i < len(fieldTypes); i++ {
		result.WriteString(fmt.Sprintf("    %s %s", fieldNames[i], fieldTypes[i]))
		if i < len(fieldNames)-1 {
			result.WriteString(",")
		}
		result.WriteString("\n")
	}

	result.WriteString(");")

	return result.String()
}
