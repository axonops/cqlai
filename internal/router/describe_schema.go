package router

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
)

// describeSchema returns CREATE statements for all keyspaces and their contents (excluding system keyspaces)
func (v *CqlCommandVisitorImpl) describeSchema() interface{} {
	logger.DebugToFile("describeSchema", "Called")

	// For Cassandra 4.0+, we can use server-side DESCRIBE SCHEMA
	if v.session.IsVersion4OrHigher() {
		logger.DebugToFile("describeSchema", "Using server-side DESCRIBE SCHEMA")
		result := v.session.ExecuteCQLQuery("DESCRIBE SCHEMA")
		
		// Extract create_statement column if it's a table result
		return v.extractCreateStatements(result, "DESCRIBE SCHEMA")
	}

	// For older versions, we need to manually build the schema by describing all keyspaces
	logger.DebugToFile("describeSchema", "Building schema manually for pre-4.0 Cassandra")
	
	// Get all non-system keyspaces
	iter := v.session.Query("SELECT keyspace_name FROM system_schema.keyspaces").Iter()
	var keyspaces []string
	var keyspaceName string
	
	for iter.Scan(&keyspaceName) {
		// Exclude system keyspaces
		if !strings.HasPrefix(keyspaceName, "system") {
			keyspaces = append(keyspaces, keyspaceName)
		}
	}
	if err := iter.Close(); err != nil {
		return fmt.Sprintf("Error listing keyspaces: %v", err)
	}
	
	if len(keyspaces) == 0 {
		return "No user keyspaces found"
	}
	
	// Build the complete schema by describing each keyspace
	var schemaStatements []string
	for _, ks := range keyspaces {
		// Get keyspace definition
		_, keyspaceInfo, err := v.session.DBDescribeKeyspace(ks)
		if err != nil {
			logger.DebugfToFile("describeSchema", "Error describing keyspace %s: %v", ks, err)
			continue
		}
		
		if keyspaceInfo != nil {
			// Format keyspace CREATE statement
			var ksStmt strings.Builder
			ksStmt.WriteString(fmt.Sprintf("CREATE KEYSPACE %s WITH replication = {\n", keyspaceInfo.Name))
			
			first := true
			for k, val := range keyspaceInfo.Replication {
				if !first {
					ksStmt.WriteString(",\n")
				}
				ksStmt.WriteString(fmt.Sprintf("    '%s': '%s'", k, val))
				first = false
			}
			
			ksStmt.WriteString(fmt.Sprintf("\n} AND durable_writes = %v;", keyspaceInfo.DurableWrites))
			schemaStatements = append(schemaStatements, ksStmt.String())
		}
		
		// Get all tables in this keyspace
		tableIter := v.session.Query(
			"SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?",
			ks,
		).Iter()
		
		var tableName string
		var tables []string
		for tableIter.Scan(&tableName) {
			tables = append(tables, tableName)
		}
		_ = tableIter.Close()
		
		// Describe each table
		for _, table := range tables {
			_, tableInfo, err := v.session.DBDescribeTable(sessionManager, fmt.Sprintf("%s.%s", ks, table))
			if err != nil {
				logger.DebugfToFile("describeSchema", "Error describing table %s.%s: %v", ks, table, err)
				continue
			}
			
			if tableInfo != nil {
				tableStmt := v.formatTableCreateStatement(tableInfo)
				schemaStatements = append(schemaStatements, tableStmt)
			}
		}
		
		// TODO: Add support for indexes, materialized views, types, functions, aggregates
	}
	
	if len(schemaStatements) == 0 {
		return "No schema to describe"
	}
	
	return strings.Join(schemaStatements, "\n\n")
}

// TODO: Add support for DESCRIBE FULL SCHEMA when grammar is updated
// describeFullSchema would return CREATE statements for ALL keyspaces including system keyspaces

// formatTableCreateStatement formats a table's CREATE TABLE statement from TableInfo
func (v *CqlCommandVisitorImpl) formatTableCreateStatement(tableInfo *db.TableInfo) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("CREATE TABLE %s.%s (\n", tableInfo.KeyspaceName, tableInfo.TableName))
	
	// Check if we have a simple primary key (single partition key, no clustering keys)
	singlePKNoCluster := len(tableInfo.PartitionKeys) == 1 && len(tableInfo.ClusteringKeys) == 0
	
	// Write column definitions
	for i, col := range tableInfo.Columns {
		result.WriteString(fmt.Sprintf("    %s %s", col.Name, col.DataType))
		
		// Add PRIMARY KEY inline if this is the partition key and conditions are met
		if singlePKNoCluster && col.Kind == "partition_key" {
			result.WriteString(" PRIMARY KEY")
		}
		
		// Add STATIC for static columns
		if col.Kind == "static" {
			result.WriteString(" STATIC")
		}
		
		if i < len(tableInfo.Columns)-1 || !singlePKNoCluster {
			result.WriteString(",")
		}
		result.WriteString("\n")
	}
	
	// Add PRIMARY KEY clause if not inline
	if !singlePKNoCluster {
		result.WriteString("    PRIMARY KEY (")
		
		// Partition key(s)
		if len(tableInfo.PartitionKeys) > 1 {
			result.WriteString("(")
		}
		for i, pk := range tableInfo.PartitionKeys {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(pk)
		}
		if len(tableInfo.PartitionKeys) > 1 {
			result.WriteString(")")
		}
		
		// Clustering key(s)
		for _, ck := range tableInfo.ClusteringKeys {
			result.WriteString(", ")
			result.WriteString(ck)
		}
		
		result.WriteString(")\n")
	}
	
	result.WriteString(")")
	
	// Add table properties if any
	if len(tableInfo.TableProps) > 0 {
		result.WriteString(" WITH ")
		first := true
		for key, value := range tableInfo.TableProps {
			if !first {
				result.WriteString("\n    AND ")
			} else {
				first = false
			}
			result.WriteString(fmt.Sprintf("%s = %v", key, value))
		}
	}
	
	result.WriteString(";")
	return result.String()
}