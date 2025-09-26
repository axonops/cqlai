package db

import (
	"fmt"
	"sort"
	"strings"

	"github.com/axonops/cqlai/internal/session"
)

// DBDescribeFullSchema returns all CREATE statements for a keyspace and its objects
// If keyspace is empty, returns schema for all keyspaces
func (s *Session) DBDescribeFullSchema(sessionMgr *session.Manager, keyspace string) (interface{}, error) {
	// For Cassandra 4.0+, DESCRIBE commands are handled server-side
	// Try server-side first, fall back to manual construction for older versions

	// Try executing DESCRIBE SCHEMA as a CQL query
	var describeQuery string
	if keyspace != "" {
		describeQuery = fmt.Sprintf("DESCRIBE KEYSPACE %s", keyspace)
	} else {
		describeQuery = "DESCRIBE SCHEMA"
	}

	queryResult := s.ExecuteCQLQuery(describeQuery)

	// Check if we got a valid result (not an error)
	switch v := queryResult.(type) {
	case string:
		// If it's an error message, fall back to manual construction
		if strings.Contains(v, "SyntaxException") || strings.Contains(v, "Invalid") ||
		   strings.Contains(v, "Unknown") || strings.Contains(v, "Error") {
			// Fall through to manual construction
			break
		}
		return v, nil

	case QueryResult:
		// Server returned a result - format it as text
		if len(v.Data) > 0 && len(v.Data[0]) > 0 {
			// DESCRIBE returns the schema as text in a single column
			var lines []string
			for _, row := range v.Data {
				if len(row) > 0 {
					lines = append(lines, row[0])
				}
			}
			return strings.Join(lines, "\n"), nil
		}

	default:
		// Unknown result type, try manual construction
		break
	}

	// Manual construction for older Cassandra versions (< 4.0)
	var result strings.Builder

	// Determine which keyspaces to describe
	var keyspaces []string
	if keyspace != "" {
		// Describe a specific keyspace
		keyspaces = []string{keyspace}
	} else {
		// Describe all keyspaces (except system keyspaces)
		iter := s.Query("SELECT keyspace_name FROM system_schema.keyspaces").Iter()
		var ksName string
		for iter.Scan(&ksName) {
			// Skip system keyspaces
			if !strings.HasPrefix(ksName, "system") {
				keyspaces = append(keyspaces, ksName)
			}
		}
		_ = iter.Close()
		sort.Strings(keyspaces)
	}

	// Process each keyspace
	for i, ks := range keyspaces {
		if i > 0 {
			result.WriteString("\n\n")
		}

		// 1. Get keyspace CREATE statement
		keyspaceInfo, err := s.DescribeKeyspaceQuery(ks)
		if err != nil {
			// Skip keyspaces we can't describe
			continue
		}

		// Build CREATE KEYSPACE statement
		result.WriteString(fmt.Sprintf("CREATE KEYSPACE %s WITH replication = %s",
			ks, FormatMapForCQL(keyspaceInfo.Replication)))

		// Always show durable_writes like cqlsh does
		if keyspaceInfo.DurableWrites {
			result.WriteString(" AND durable_writes = true")
		} else {
			result.WriteString(" AND durable_writes = false")
		}
		result.WriteString(";\n\n")

		// 2. Get all types
		typesQuery := `SELECT type_name FROM system_schema.types WHERE keyspace_name = ?`
		iter := s.Query(typesQuery, ks).Iter()
	var typeName string
	var typeNames []string
	for iter.Scan(&typeName) {
		typeNames = append(typeNames, typeName)
	}
	_ = iter.Close()

		// Get CREATE statement for each type
		sort.Strings(typeNames)
		for _, name := range typeNames {
			typeInfo, err := s.DescribeTypeQuery(ks, name)
			if err == nil && typeInfo != nil {
				result.WriteString(formatTypeCreateStatement(ks, typeInfo))
				result.WriteString("\n\n")
			}
		}

		// 3. Get all tables and their indexes
		tablesQuery := `SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?`
		iter = s.Query(tablesQuery, ks).Iter()
	var tableName string
	var tableNames []string
	for iter.Scan(&tableName) {
		tableNames = append(tableNames, tableName)
	}
	_ = iter.Close()

		// Get all indexes for this keyspace
		indexQuery := `SELECT index_name, table_name FROM system_schema.indexes WHERE keyspace_name = ?`
		iter = s.Query(indexQuery, ks).Iter()
	var indexName, indexTableName string
	type indexPair struct {
		name  string
		table string
	}
	tableIndexes := make(map[string][]indexPair) // map from table name to its indexes
	for iter.Scan(&indexName, &indexTableName) {
		tableIndexes[indexTableName] = append(tableIndexes[indexTableName], indexPair{name: indexName, table: indexTableName})
	}
	_ = iter.Close()

		// Get CREATE statement for each table and its indexes
		sort.Strings(tableNames)
		for _, name := range tableNames {
			tableInfo, err := s.DescribeTableQuery(ks, name)
			if err == nil && tableInfo != nil {
				result.WriteString(FormatTableCreateStatement(tableInfo, false))
				result.WriteString("\n\n")

				// Add indexes for this table immediately after
				if indexes, ok := tableIndexes[name]; ok {
					sort.Slice(indexes, func(i, j int) bool { return indexes[i].name < indexes[j].name })
					for _, idx := range indexes {
						indexInfo, err := s.DescribeIndexQueryWithTable(ks, idx.table, idx.name)
						if err == nil && indexInfo != nil {
							result.WriteString(formatIndexCreateStatement(ks, indexInfo))
							result.WriteString("\n\n")
						}
					}
				}
			}
		}

		// 5. Get all materialized views
		viewsQuery := `SELECT view_name FROM system_schema.views WHERE keyspace_name = ?`
		iter = s.Query(viewsQuery, ks).Iter()
	var viewName string
	var viewNames []string
	for iter.Scan(&viewName) {
		viewNames = append(viewNames, viewName)
	}
	_ = iter.Close()

		// Get CREATE statement for each materialized view
		sort.Strings(viewNames)
		for _, name := range viewNames {
			mvInfo, err := s.DescribeMaterializedViewQuery(ks, name)
			if err == nil && mvInfo != nil {
				result.WriteString(formatMaterializedViewCreateStatement(ks, mvInfo))
				result.WriteString("\n\n")
			}
		}

		// 6. Get all functions
		functionsQuery := `SELECT function_name, argument_types FROM system_schema.functions WHERE keyspace_name = ?`
		iter = s.Query(functionsQuery, ks).Iter()
	var functionName string
	var argumentTypes []string
	type functionSig struct {
		name string
		args []string
	}
	var functions []functionSig
	for iter.Scan(&functionName, &argumentTypes) {
		functions = append(functions, functionSig{name: functionName, args: argumentTypes})
	}
	_ = iter.Close()

		// Get CREATE statement for each function
		sort.Slice(functions, func(i, j int) bool { return functions[i].name < functions[j].name })
		for _, fn := range functions {
			fnDetails, err := s.DescribeFunctionQuery(ks, fn.name)
			if err == nil && len(fnDetails) > 0 {
				for _, detail := range fnDetails {
					result.WriteString(formatFunctionCreateStatement(ks, &detail))
					result.WriteString("\n\n")
				}
			}
		}

		// 7. Get all aggregates
		aggregatesQuery := `SELECT aggregate_name, argument_types FROM system_schema.aggregates WHERE keyspace_name = ?`
		iter = s.Query(aggregatesQuery, ks).Iter()
	var aggregateName string
	var aggregateArgTypes []string
	type aggregateSig struct {
		name string
		args []string
	}
	var aggregates []aggregateSig
	for iter.Scan(&aggregateName, &aggregateArgTypes) {
		aggregates = append(aggregates, aggregateSig{name: aggregateName, args: aggregateArgTypes})
	}
	_ = iter.Close()

		// Get CREATE statement for each aggregate
		sort.Slice(aggregates, func(i, j int) bool { return aggregates[i].name < aggregates[j].name })
		for _, agg := range aggregates {
			aggInfo, err := s.DescribeAggregateQuery(ks, agg.name)
			if err == nil && aggInfo != nil {
				result.WriteString(formatAggregateCreateStatement(ks, aggInfo))
				result.WriteString("\n\n")
			}
		}
	}

	return result.String(), nil
}

// Helper functions to format CREATE statements


func formatTypeCreateStatement(keyspace string, typeInfo *TypeInfo) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("CREATE TYPE %s.%s (\n", keyspace, typeInfo.Name))

	for i := range typeInfo.FieldNames {
		if i < len(typeInfo.FieldTypes) {
			result.WriteString(fmt.Sprintf("    %s %s", typeInfo.FieldNames[i], typeInfo.FieldTypes[i]))
		}
		if i < len(typeInfo.FieldNames)-1 {
			result.WriteString(",")
		}
		result.WriteString("\n")
	}

	result.WriteString(");")

	return result.String()
}

func formatIndexCreateStatement(keyspace string, index *IndexInfo) string {
	var result strings.Builder

	if index.Kind == "CUSTOM" {
		// For CUSTOM indexes, extract target (column) and class_name
		target := index.Options["target"]
		className := index.Options["class_name"]

		// Build CREATE CUSTOM INDEX statement
		if target != "" {
			result.WriteString(fmt.Sprintf("CREATE CUSTOM INDEX %s ON %s.%s (%s)",
				index.IndexName, keyspace, index.TableName, target))
		} else {
			result.WriteString(fmt.Sprintf("CREATE CUSTOM INDEX %s ON %s.%s",
				index.IndexName, keyspace, index.TableName))
		}

		// Add USING clause
		if className != "" {
			result.WriteString(fmt.Sprintf(" USING '%s'", className))
		}

		// Build OPTIONS excluding target and class_name
		filteredOptions := make(map[string]string)
		for k, v := range index.Options {
			if k != "target" && k != "class_name" {
				filteredOptions[k] = v
			}
		}

		if len(filteredOptions) > 0 {
			result.WriteString(" WITH OPTIONS = ")
			result.WriteString(FormatMapForCQL(filteredOptions))
		}
	} else {
		// For regular (non-CUSTOM) indexes
		if target, ok := index.Options["target"]; ok {
			result.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s.%s (%s)",
				index.IndexName, keyspace, index.TableName, target))
		} else {
			result.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s.%s",
				index.IndexName, keyspace, index.TableName))
		}

		// Regular indexes typically don't have additional options
		// but include them if present (excluding target)
		filteredOptions := make(map[string]string)
		for k, v := range index.Options {
			if k != "target" {
				filteredOptions[k] = v
			}
		}

		if len(filteredOptions) > 0 {
			result.WriteString(" WITH OPTIONS = ")
			result.WriteString(FormatMapForCQL(filteredOptions))
		}
	}

	result.WriteString(";")

	return result.String()
}

func formatMaterializedViewCreateStatement(keyspace string, mv *MaterializedViewInfo) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("CREATE MATERIALIZED VIEW %s.%s AS\n", keyspace, mv.Name))
	result.WriteString(fmt.Sprintf("    SELECT * FROM %s.%s\n", keyspace, mv.BaseTable))
	result.WriteString(fmt.Sprintf("    WHERE %s\n", mv.WhereClause))

	// Add PRIMARY KEY
	result.WriteString("    PRIMARY KEY (")
	if len(mv.PartitionKeys) > 1 {
		result.WriteString("(")
		result.WriteString(strings.Join(mv.PartitionKeys, ", "))
		result.WriteString(")")
	} else if len(mv.PartitionKeys) == 1 {
		result.WriteString(mv.PartitionKeys[0])
	}
	if len(mv.ClusteringKeys) > 0 {
		result.WriteString(", ")
		result.WriteString(strings.Join(mv.ClusteringKeys, ", "))
	}
	result.WriteString(")")

	// Add MV options
	var options []string

	if mv.Comment != "" {
		options = append(options, fmt.Sprintf("comment = '%s'", mv.Comment))
	}

	if len(mv.Compaction) > 0 {
		options = append(options, fmt.Sprintf("compaction = %s", FormatMapForCQL(mv.Compaction)))
	}

	if len(mv.Compression) > 0 {
		options = append(options, fmt.Sprintf("compression = %s", FormatMapForCQL(mv.Compression)))
	}

	if len(options) > 0 {
		result.WriteString("\n    WITH ")
		result.WriteString(strings.Join(options, " AND "))
	}

	result.WriteString(";")

	return result.String()
}

func formatFunctionCreateStatement(keyspace string, fn *FunctionDetails) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("CREATE FUNCTION %s.%s(", keyspace, fn.Name))

	// Add parameters
	params := make([]string, 0, len(fn.ArgumentNames))
	for i := range fn.ArgumentNames {
		if i < len(fn.ArgumentTypes) {
			params = append(params, fmt.Sprintf("%s %s", fn.ArgumentNames[i], fn.ArgumentTypes[i]))
		}
	}
	result.WriteString(strings.Join(params, ", "))
	result.WriteString(")")

	// Add return type
	if fn.CalledOnNull {
		result.WriteString("\n    CALLED ON NULL INPUT")
	} else {
		result.WriteString("\n    RETURNS NULL ON NULL INPUT")
	}

	result.WriteString(fmt.Sprintf("\n    RETURNS %s", fn.ReturnType))
	result.WriteString(fmt.Sprintf("\n    LANGUAGE %s", fn.Language))
	result.WriteString(fmt.Sprintf("\n    AS '%s';", fn.Body))

	return result.String()
}

func formatAggregateCreateStatement(keyspace string, agg *AggregateInfo) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("CREATE AGGREGATE %s.%s(%s)",
		keyspace, agg.Name, strings.Join(agg.ArgumentTypes, ", ")))

	result.WriteString(fmt.Sprintf("\n    SFUNC %s", agg.StateFunc))
	result.WriteString(fmt.Sprintf("\n    STYPE %s", agg.StateType))

	if agg.FinalFunc != "" {
		result.WriteString(fmt.Sprintf("\n    FINALFUNC %s", agg.FinalFunc))
	}

	if agg.InitCond != "" {
		result.WriteString(fmt.Sprintf("\n    INITCOND %s", agg.InitCond))
	}

	result.WriteString(";")

	return result.String()
}