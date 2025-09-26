package router

import (
	"fmt"
	"sort"
	"strings"

	"github.com/axonops/cqlai/internal/db"
)

// Abbreviation maps for compaction and compression strategies
var compactionStrategyAbbrev = map[string]string{
	"SizeTieredCompactionStrategy":     "STCS",
	"LeveledCompactionStrategy":        "LCS",
	"DateTieredCompactionStrategy":     "DTCS",
	"TimeWindowCompactionStrategy":     "TWCS",
	"UnifiedCompactionStrategy":        "UCS",
}

var compressionAbbrev = map[string]string{
	"LZ4Compressor":      "LZ4",
	"SnappyCompressor":   "Snappy",
	"DeflateCompressor":  "Deflate",
	"ZstdCompressor":     "Zstd",
}

// describeKeyspaces returns a list of all keyspaces
func (p *CommandParser) describeKeyspaces() interface{} {
	serverResult, keyspaces, err := p.session.DBDescribeKeyspaces()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available (Cassandra 4.0+)
	if serverResult != nil {
		return serverResult
	}

	// Format keyspaces list for display
	if len(keyspaces) == 0 {
		return "No keyspaces found"
	}

	var names []string
	for _, ks := range keyspaces {
		names = append(names, ks.Name)
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

// describeTables returns a table of all tables in the current keyspace
// MUST return [][]string to display as a table in the UI
func (p *CommandParser) describeTables() interface{} {
	// Get current keyspace
	currentKeyspace := ""
	if p.sessionManager != nil {
		currentKeyspace = p.sessionManager.CurrentKeyspace()
	}

	serverResult, tables, err := p.session.DBDescribeTables(p.sessionManager)
	if err != nil {
		if err.Error() == "no keyspace selected" {
			return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		// Check if we need to filter for current keyspace (Cassandra 5.0 behavior)
		if streamResult, ok := serverResult.(db.StreamingQueryResult); ok {
			return p.filterStreamingTablesResult(streamResult, currentKeyspace)
		}
		if queryResult, ok := serverResult.(db.QueryResult); ok {
			return p.filterQueryTablesResult(queryResult, currentKeyspace)
		}
		return serverResult
	}

	// Manual formatting - MUST return [][]string for table display
	if len(tables) == 0 {
		return fmt.Sprintf("No tables found in keyspace %s", currentKeyspace)
	}

	// Build results table - this format MUST be maintained for UI
	results := [][]string{{"Table", "Primary Key", "Compaction", "Compression", "GC Grace"}}

	for _, t := range tables {
		// Format primary key
		pkStr := p.formatPrimaryKey(t.PartitionKeys, t.ClusteringKeys)

		// Format compaction strategy
		compactionStr := p.formatCompaction(t.Compaction)

		// Format compression
		compressionStr := p.formatCompression(t.Compression)

		// Format GC grace
		gcGraceStr := p.formatGCGrace(t.GcGrace)

		results = append(results, []string{
			t.Name,
			pkStr,
			compactionStr,
			compressionStr,
			gcGraceStr,
		})
	}

	return results
}

// describeTable shows detailed information about a specific table
func (p *CommandParser) describeTable(tableName string) interface{} {
	// Remove quotes if present
	tableName = strings.Trim(tableName, "'\"")

	serverResult, tableInfo, err := p.session.DBDescribeTable(p.sessionManager, tableName)
	if err != nil {
		if err.Error() == "no keyspace selected" {
			return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
		}
		if strings.Contains(err.Error(), "not found") {
			return err.Error()
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// Manual format - build CREATE TABLE statement
	if tableInfo == nil {
		return fmt.Sprintf("Table '%s' not found", tableName)
	}

	return db.FormatTableCreateStatement(tableInfo, true)
}

// describeKeyspace shows detailed information about a keyspace
func (p *CommandParser) describeKeyspace(keyspaceName string) interface{} {
	// Remove quotes if present
	keyspaceName = strings.Trim(keyspaceName, "'\"")

	// Use DBDescribeFullSchema to get the complete schema for this keyspace
	// This matches cqlsh behavior which shows CREATE statements for all objects
	schema, err := p.session.DBDescribeFullSchema(p.sessionManager, keyspaceName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Sprintf("Keyspace '%s' not found", keyspaceName)
		}
		return fmt.Sprintf("Error: %v", err)
	}

	return schema
}

// describeCluster shows cluster information
func (p *CommandParser) describeCluster() interface{} {
	serverResult, clusterInfo, err := p.session.DBDescribeCluster()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// Manual format
	if clusterInfo == nil {
		return "Unable to retrieve cluster information"
	}

	return fmt.Sprintf("Cluster: %s\nPartitioner: %s",
		clusterInfo.ClusterName, clusterInfo.Partitioner)
}

// describeTypes lists all user-defined types
func (p *CommandParser) describeTypes() interface{} {
	serverResult, types, err := p.session.DBDescribeTypes(p.sessionManager)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// Manual format
	if len(types) == 0 {
		return "No types found"
	}

	var names []string
	for _, t := range types {
		names = append(names, t.Name)
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

// describeType shows detailed information about a specific type
func (p *CommandParser) describeType(typeName string) interface{} {
	// Remove quotes if present
	typeName = strings.Trim(typeName, "'\"")

	serverResult, typeInfo, err := p.session.DBDescribeType(p.sessionManager, typeName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Sprintf("Type '%s' not found", typeName)
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// For manual format, we'd need to build CREATE TYPE statement
	// For now, return basic info
	if typeInfo == nil {
		return fmt.Sprintf("Type '%s' not found", typeName)
	}

	return fmt.Sprintf("Type: %s", typeName)
}

// describeFunctions lists all user-defined functions
func (p *CommandParser) describeFunctions() interface{} {
	serverResult, _, err := p.session.DBDescribeFunctions(p.sessionManager)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// DBDescribeFunctions returns the result directly
	if serverResult != nil {
		return serverResult
	}

	return "No functions found"
}

// describeFunction shows detailed information about a specific function
func (p *CommandParser) describeFunction(functionName string) interface{} {
	// Functions are complex - pass through a query
	// Remove quotes if present
	functionName = strings.Trim(functionName, "'\"")

	// Parse keyspace.function if present
	parts := strings.Split(functionName, ".")
	var keyspaceName string

	currentKeyspace := ""
	if p.sessionManager != nil {
		currentKeyspace = p.sessionManager.CurrentKeyspace()
	}

	if len(parts) == 2 {
		// keyspace.function format
		keyspaceName = parts[0]
		functionName = parts[1]
	} else if currentKeyspace != "" {
		// Use current keyspace
		keyspaceName = currentKeyspace
	}
	// If no keyspace, keyspaceName remains empty string

	result, err := p.session.DBDescribeFunctionByName(functionName, keyspaceName)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return result
}

// describeAggregates lists all user-defined aggregates
func (p *CommandParser) describeAggregates() interface{} {
	serverResult, aggregates, err := p.session.DBDescribeAggregates(p.sessionManager)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// Manual format
	if len(aggregates) == 0 {
		return "No aggregates found"
	}

	var names []string
	for _, a := range aggregates {
		names = append(names, a.Name)
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

// describeAggregate shows detailed information about a specific aggregate
func (p *CommandParser) describeAggregate(aggregateName string) interface{} {
	// Remove quotes if present
	aggregateName = strings.Trim(aggregateName, "'\"")

	serverResult, aggregateInfo, err := p.session.DBDescribeAggregate(p.sessionManager, aggregateName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Sprintf("Aggregate '%s' not found", aggregateName)
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// For manual format, return basic info
	if aggregateInfo == nil {
		return fmt.Sprintf("Aggregate '%s' not found", aggregateName)
	}

	return fmt.Sprintf("Aggregate: %s", aggregateName)
}

// describeIndex shows detailed information about an index
func (p *CommandParser) describeIndex(indexName string) interface{} {
	// Remove quotes if present
	indexName = strings.Trim(indexName, "'\"")

	serverResult, indexInfo, err := p.session.DBDescribeIndex(p.sessionManager, indexName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Sprintf("Index '%s' not found", indexName)
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// For manual format, return basic info
	if indexInfo == nil {
		return fmt.Sprintf("Index '%s' not found", indexName)
	}

	return fmt.Sprintf("Index: %s", indexName)
}

// describeMaterializedView shows detailed information about a materialized view
func (p *CommandParser) describeMaterializedView(viewName string) interface{} {
	// Remove quotes if present
	viewName = strings.Trim(viewName, "'\"")

	serverResult, mvInfo, err := p.session.DBDescribeMaterializedView(p.sessionManager, viewName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Sprintf("Materialized view '%s' not found", viewName)
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// For manual format, return basic info
	if mvInfo == nil {
		return fmt.Sprintf("Materialized view '%s' not found", viewName)
	}

	return fmt.Sprintf("Materialized view: %s", viewName)
}

// describeMaterializedViews lists all materialized views
func (p *CommandParser) describeMaterializedViews() interface{} {
	result, err := p.session.DBDescribeMaterializedViews(p.sessionManager)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return result
}

// describeSchema shows all CREATE statements for ALL keyspaces
func (p *CommandParser) describeSchema() interface{} {
	// DESCRIBE SCHEMA should show the entire cluster schema (all keyspaces)
	// Not just the current keyspace
	schema, err := p.session.DBDescribeFullSchema(p.sessionManager, "")
	if err != nil {
		return fmt.Sprintf("Error describing schema: %v", err)
	}

	return schema
}

// describeIdentifier tries to determine what type of object an identifier is
func (p *CommandParser) describeIdentifier(identifier string) interface{} {
	// Remove quotes if present
	identifier = strings.Trim(identifier, "'\"")

	// Check if it contains a dot (keyspace.object notation)
	if strings.Contains(identifier, ".") {
		parts := strings.Split(identifier, ".")
		if len(parts) == 2 {
			// Try as keyspace.table first
			result := p.describeTable(identifier)
			if !strings.Contains(result.(string), "not found") {
				return result
			}
		}
	}

	// Try as keyspace
	result := p.describeKeyspace(identifier)
	if resultStr, ok := result.(string); !ok || !strings.Contains(resultStr, "not found") {
		return result
	}

	// Try as table in current keyspace
	if p.sessionManager != nil && p.sessionManager.CurrentKeyspace() != "" {
		result = p.describeTable(identifier)
		if resultStr, ok := result.(string); !ok || !strings.Contains(resultStr, "not found") {
			return result
		}
	}

	return fmt.Sprintf("'%s' not found", identifier)
}