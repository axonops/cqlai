package router

import (
	"fmt"
	"strings"
	"github.com/axonops/cqlai/internal/logger"
)

// describeIndex shows detailed information about a specific index
func (v *CqlCommandVisitorImpl) describeIndex(indexName string) interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeIndex", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Parse keyspace.index or just index
		var describeCmd string
		if strings.Contains(indexName, ".") {
			describeCmd = fmt.Sprintf("DESCRIBE INDEX %s", indexName)
		} else {
			currentKeyspace := v.session.CurrentKeyspace()
			if currentKeyspace == "" {
				return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
			}
			describeCmd = fmt.Sprintf("DESCRIBE INDEX %s.%s", currentKeyspace, indexName)
		}
		
		logger.DebugToFile("describeIndex", fmt.Sprintf("Using server-side: %s", describeCmd))
		return v.session.ExecuteCQLQuery(describeCmd)
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := v.session.CurrentKeyspace()
	if currentKeyspace == "" {
		return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
	}

	query := `SELECT table_name, index_name, kind, options 
	          FROM system_schema.indexes 
	          WHERE keyspace_name = ? AND index_name = ?`

	iter := v.session.Query(query, currentKeyspace, indexName).Iter()

	var tableName, idxName, kind string
	var options map[string]string

	if !iter.Scan(&tableName, &idxName, &kind, &options) {
		iter.Close()
		return fmt.Sprintf("Index '%s' not found in keyspace '%s'", indexName, currentKeyspace)
	}
	iter.Close()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Index: %s.%s\n\n", currentKeyspace, indexName))

	// Format CREATE INDEX statement
	target := ""
	if t, ok := options["target"]; ok {
		target = t
	}

	result.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s.%s (%s);\n",
		indexName, currentKeyspace, tableName, target))

	result.WriteString(fmt.Sprintf("\nType: %s\n", kind))
	result.WriteString(fmt.Sprintf("Table: %s\n", tableName))
	result.WriteString(fmt.Sprintf("Target: %s\n", target))

	// Show any additional options
	if len(options) > 1 {
		result.WriteString("\nOptions:\n")
		for k, v := range options {
			if k != "target" {
				result.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
			}
		}
	}

	return result.String()
}
