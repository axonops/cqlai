package router

import (
	"fmt"
	"sort"
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/validation"
)

// Abbreviation maps for compaction and compression strategies
var compactionStrategyAbbrev = map[string]string{
	"SizeTieredCompactionStrategy": "STCS",
	"LeveledCompactionStrategy":    "LCS",
	"DateTieredCompactionStrategy": "DTCS",
	"TimeWindowCompactionStrategy": "TWCS",
	"UnifiedCompactionStrategy":    "UCS",
}

var compressionAbbrev = map[string]string{
	"LZ4Compressor":     "LZ4",
	"SnappyCompressor":  "Snappy",
	"DeflateCompressor": "Deflate",
	"ZstdCompressor":    "Zstd",
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
			return fmt.Errorf("no keyspace selected: USE a keyspace, or name it as keyspace.name")
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
	results := p.tableListRows(tables)

	return results
}

// describeTablesPattern describes tables matching a pattern (e.g., "test*")
func (p *CommandParser) describeTablesPattern(pattern string) interface{} {
	// Get current keyspace
	currentKeyspace := ""
	if p.sessionManager != nil {
		currentKeyspace = p.sessionManager.CurrentKeyspace()
	}

	if currentKeyspace == "" {
		return fmt.Errorf("no keyspace selected: USE a keyspace, or name it as keyspace.name")
	}

	serverResult, tables, err := p.session.DBDescribeTables(p.sessionManager)
	if err != nil {
		if err.Error() == "no keyspace selected" {
			return fmt.Errorf("no keyspace selected: USE a keyspace, or name it as keyspace.name")
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// Convert wildcard pattern to matching function
	pattern = strings.ToLower(pattern)
	matchFunc := func(tableName string) bool {
		name := strings.ToLower(tableName)
		// Simple wildcard matching: * matches any sequence
		switch {
		case strings.HasSuffix(pattern, "*"):
			prefix := strings.TrimSuffix(pattern, "*")
			return strings.HasPrefix(name, prefix)
		case strings.HasPrefix(pattern, "*"):
			suffix := strings.TrimPrefix(pattern, "*")
			return strings.HasSuffix(name, suffix)
		case strings.Contains(pattern, "*"):
			parts := strings.Split(pattern, "*")
			if len(parts) == 2 {
				return strings.HasPrefix(name, parts[0]) && strings.HasSuffix(name, parts[1])
			}
			return false
		default:
			return name == pattern
		}
	}

	// If we have server result, we can't easily filter it, so use manual filtering
	_ = serverResult

	// Manual filtering
	filteredTables := []db.TableListInfo{}
	for _, t := range tables {
		if matchFunc(t.Name) {
			filteredTables = append(filteredTables, t)
		}
	}

	if len(filteredTables) == 0 {
		return fmt.Sprintf("No tables matching pattern '%s' found in keyspace %s", pattern, currentKeyspace)
	}

	// Build results table
	results := p.tableListRows(filteredTables)

	return results
}

// describeTable shows detailed information about a specific table
func (p *CommandParser) describeTable(typed string) interface{} {
	name, err := validation.ParseName(typed)
	if err != nil {
		return err
	}
	tableName := name.String()

	serverResult, tableInfo, err := p.session.DBDescribeTable(p.sessionManager, tableName)
	if err != nil {
		if err.Error() == "no keyspace selected" {
			return fmt.Errorf("no keyspace selected: USE a keyspace, or name it as keyspace.name")
		}
		if strings.Contains(err.Error(), "not found") {
			return err
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// Manual format - build CREATE TABLE statement
	if tableInfo == nil {
		return fmt.Errorf("table '%s' not found", tableName)
	}

	return db.FormatTableCreateStatement(tableInfo, true)
}

// describeKeyspace shows detailed information about a keyspace
func (p *CommandParser) describeKeyspace(typed string) interface{} {
	name, err := validation.ParseName(typed)
	if err != nil {
		return err
	}
	if name.Qualified {
		return fmt.Errorf("keyspace '%s' not found: a keyspace is one name", name)
	}
	keyspaceName := name.Table

	// Use DBDescribeFullSchema to get the complete schema for this keyspace
	// This matches cqlsh behavior which shows CREATE statements for all objects
	schema, err := p.session.DBDescribeFullSchema(p.sessionManager, keyspaceName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("keyspace '%s' not found", keyspaceName)
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
func (p *CommandParser) describeType(typed string) interface{} {
	name, err := validation.ParseName(typed)
	if err != nil {
		return err
	}
	typeName := name.String()

	serverResult, typeInfo, err := p.session.DBDescribeType(p.sessionManager, typeName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("type '%s' not found", typeName)
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
		return fmt.Errorf("type '%s' not found", typeName)
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
func (p *CommandParser) describeFunction(typed string) interface{} {
	name, err := validation.ParseName(typed)
	if err != nil {
		return err
	}

	// In the keyspace named, or the one in use, or with neither in all of them.
	functionName, keyspaceName := name.Table, name.Keyspace
	if !name.Qualified && p.sessionManager != nil {
		keyspaceName = p.sessionManager.CurrentKeyspace()
	}

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
func (p *CommandParser) describeAggregate(typed string) interface{} {
	name, err := validation.ParseName(typed)
	if err != nil {
		return err
	}
	aggregateName := name.String()

	serverResult, aggregateInfo, err := p.session.DBDescribeAggregate(p.sessionManager, aggregateName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("aggregate '%s' not found", aggregateName)
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// For manual format, return basic info
	if aggregateInfo == nil {
		return fmt.Errorf("aggregate '%s' not found", aggregateName)
	}

	return fmt.Sprintf("Aggregate: %s", aggregateName)
}

// describeIndex shows detailed information about an index
func (p *CommandParser) describeIndex(typed string) interface{} {
	name, err := validation.ParseName(typed)
	if err != nil {
		return err
	}
	indexName := name.String()

	serverResult, indexInfo, err := p.session.DBDescribeIndex(p.sessionManager, indexName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("index '%s' not found", indexName)
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// For manual format, return basic info
	if indexInfo == nil {
		return fmt.Errorf("index '%s' not found", indexName)
	}

	return fmt.Sprintf("Index: %s", indexName)
}

// describeMaterializedView shows detailed information about a materialized view
func (p *CommandParser) describeMaterializedView(typed string) interface{} {
	name, err := validation.ParseName(typed)
	if err != nil {
		return err
	}
	viewName := name.String()

	serverResult, mvInfo, err := p.session.DBDescribeMaterializedView(p.sessionManager, viewName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("materialized view '%s' not found", viewName)
		}
		return fmt.Sprintf("Error: %v", err)
	}

	// If server-side DESCRIBE is available
	if serverResult != nil {
		return serverResult
	}

	// For manual format, return basic info
	if mvInfo == nil {
		return fmt.Errorf("materialized view '%s' not found", viewName)
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
	name, err := validation.ParseName(identifier)
	if err != nil {
		return err
	}

	// keyspace.name is a table.
	if name.Qualified {
		if result := p.describeTable(identifier); !notFound(result) {
			return result
		}
		return fmt.Errorf("'%s' not found", name)
	}

	// A name alone is a keyspace, or a table in the keyspace in use.
	if result := p.describeKeyspace(identifier); !notFound(result) {
		return result
	}
	if p.sessionManager != nil && p.sessionManager.CurrentKeyspace() != "" {
		if result := p.describeTable(identifier); !notFound(result) {
			return result
		}
	}

	return fmt.Errorf("'%s' not found", name)
}

// notFound reports whether a DESCRIBE answered that there is no such object.
// It answers with an error or with text, and asserting text of an error
// panicked: DESCRIBE k.nosuch, a name that was neither, took the shell down.
func notFound(result interface{}) bool {
	switch r := result.(type) {
	case error:
		return strings.Contains(r.Error(), "not found")
	case string:
		return strings.Contains(r, "not found")
	}
	return false
}

// tableListRows is DESCRIBE TABLES as rows: a header, then a table a row.
//
// It was written out twice, once for each way of asking, and a virtual table
// listed through either said its compaction was "Unknown" and its gc_grace_seconds
// 0. Neither is unknown or zero: a virtual table is computed when it is asked,
// so there is nothing on disk for compaction, compression or a grace period to
// apply to. It says so, the same way in both.
func (p *CommandParser) tableListRows(tables []db.TableListInfo) [][]string {
	rows := [][]string{{"Table", "Primary Key", "Compaction", "Compression", "GC Grace"}}

	for _, t := range tables {
		key := p.formatPrimaryKey(t.PartitionKeys, t.ClusteringKeys)

		if t.Virtual {
			rows = append(rows, []string{t.Name, key, "virtual", "-", "-"})
			continue
		}

		rows = append(rows, []string{
			t.Name,
			key,
			p.formatCompaction(t.Compaction),
			p.formatCompression(t.Compression),
			p.formatGCGrace(t.GcGrace),
		})
	}
	return rows
}
