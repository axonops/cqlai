package router

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/db"
)

// formatPrimaryKey formats partition and clustering keys for display
func (p *CommandParser) formatPrimaryKey(partitionKeys []string, clusteringKeys []string) string {
	if len(partitionKeys) == 0 {
		return "?"
	}

	var primaryKeys []string
	switch {
	case len(partitionKeys) == 1 && len(clusteringKeys) == 0:
		// Single partition key only
		primaryKeys = partitionKeys
	case len(partitionKeys) == 1 && len(clusteringKeys) > 0:
		// Single partition key with clustering keys
		primaryKeys = make([]string, 0, len(partitionKeys)+len(clusteringKeys))
		primaryKeys = append(primaryKeys, partitionKeys...)
		primaryKeys = append(primaryKeys, clusteringKeys...)
	default:
		// Composite partition key
		primaryKeys = append([]string{fmt.Sprintf("(%s)", strings.Join(partitionKeys, ","))}, clusteringKeys...)
	}

	return strings.Join(primaryKeys, ", ")
}

// formatCompaction formats compaction strategy for display
func (p *CommandParser) formatCompaction(compaction map[string]string) string {
	compactionStr := "Unknown"
	if class, ok := compaction["class"]; ok {
		// Get just the strategy name without the full class path
		parts := strings.Split(class, ".")
		strategy := parts[len(parts)-1]
		// Use abbreviation if available, otherwise use the strategy name
		if abbrev, found := compactionStrategyAbbrev[strategy]; found {
			compactionStr = abbrev
		} else {
			compactionStr = strategy
		}
	}
	return compactionStr
}

// formatCompression formats compression strategy for display
func (p *CommandParser) formatCompression(compression map[string]string) string {
	compressionStr := "None"
	if class, ok := compression["class"]; ok {
		parts := strings.Split(class, ".")
		algorithm := parts[len(parts)-1]
		// Use abbreviation if available, otherwise use the algorithm name
		if abbrev, found := compressionAbbrev[algorithm]; found {
			compressionStr = abbrev
		} else {
			compressionStr = algorithm
		}
	}
	return compressionStr
}

// formatGCGrace formats GC grace seconds for display
func (p *CommandParser) formatGCGrace(gcGrace int) string {
	gcGraceStr := fmt.Sprintf("%d", gcGrace)
	switch {
	case gcGrace >= 86400:
		days := gcGrace / 86400
		gcGraceStr = fmt.Sprintf("%dd", days)
	case gcGrace >= 3600:
		hours := gcGrace / 3600
		gcGraceStr = fmt.Sprintf("%dh", hours)
	case gcGrace > 0:
		gcGraceStr = fmt.Sprintf("%ds", gcGrace)
	}
	return gcGraceStr
}

// formatCreateTableStatement builds a CREATE TABLE statement from TableInfo
func (p *CommandParser) formatCreateTableStatement(tableInfo *db.TableInfo) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Table: %s.%s\n\n", tableInfo.KeyspaceName, tableInfo.TableName))

	// Format CREATE TABLE statement
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

		if i < len(tableInfo.Columns)-1 || !singlePKNoCluster {
			result.WriteString(",")
		}
		result.WriteString("\n")
	}

	// Add composite PRIMARY KEY if needed
	if !singlePKNoCluster && len(tableInfo.PartitionKeys) > 0 {
		result.WriteString("    PRIMARY KEY (")

		// Format partition keys
		if len(tableInfo.PartitionKeys) > 1 {
			result.WriteString(fmt.Sprintf("(%s)", strings.Join(tableInfo.PartitionKeys, ", ")))
		} else if len(tableInfo.PartitionKeys) == 1 {
			result.WriteString(tableInfo.PartitionKeys[0])
		}

		// Add clustering keys
		if len(tableInfo.ClusteringKeys) > 0 {
			result.WriteString(", ")
			result.WriteString(strings.Join(tableInfo.ClusteringKeys, ", "))
		}

		result.WriteString(")\n")
	}

	result.WriteString(")")

	// Add table properties if available
	var properties []string
	for key, value := range tableInfo.TableProps {
		prop := formatTableProperty(key, value)
		if prop != "" {
			properties = append(properties, prop)
		}
	}

	// Write WITH clause
	if len(properties) > 0 {
		result.WriteString(" WITH ")
		for i, prop := range properties {
			if i > 0 {
				result.WriteString("\n    AND ")
			}
			result.WriteString(prop)
		}
	}
	result.WriteString(";")

	return result.String()
}




// formatTableProperty formats a single table property for the WITH clause
func formatTableProperty(name string, value interface{}) string {
	// Handle nil values
	if value == nil {
		showNullFor := map[string]bool{
			"memtable": true,
		}
		if showNullFor[name] {
			if name == "memtable" {
				return fmt.Sprintf("%s = ''", name)
			}
			return fmt.Sprintf("%s = null", name)
		}
		return ""
	}

	// Handle different types
	switch v := value.(type) {
	case string:
		// String properties that should be quoted
		if name == "comment" || name == "speculative_retry" || name == "additional_write_policy" ||
			name == "memtable" || name == "read_repair" {
			return fmt.Sprintf("%s = '%s'", name, v)
		}
		// Unquoted strings
		return fmt.Sprintf("%s = %s", name, v)

	case bool:
		return fmt.Sprintf("%s = %v", name, v)

	case int, int32, int64:
		return fmt.Sprintf("%s = %v", name, v)

	case float32, float64:
		return fmt.Sprintf("%s = %g", name, v)

	case map[string]string:
		// Format map properties (like compaction, compression, caching)
		return formatMapProperty(name, v)

	case map[string]interface{}:
		// Convert to string map and format
		strMap := make(map[string]string)
		for k, val := range v {
			strMap[k] = fmt.Sprintf("%v", val)
		}
		return formatMapProperty(name, strMap)

	default:
		// Default formatting
		return fmt.Sprintf("%s = %v", name, v)
	}
}

// formatMapProperty formats map-type properties
func formatMapProperty(name string, m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	var parts []string
	for k, v := range m {
		// Quote string values, but not numeric or boolean values
		if needsQuoting(v) {
			parts = append(parts, fmt.Sprintf("'%s': '%s'", k, v))
		} else {
			parts = append(parts, fmt.Sprintf("'%s': %s", k, v))
		}
	}

	return fmt.Sprintf("%s = {%s}", name, strings.Join(parts, ", "))
}

// needsQuoting determines if a value needs quotes
func needsQuoting(value string) bool {
	// Don't quote numbers
	if _, err := fmt.Sscanf(value, "%f", new(float64)); err == nil {
		return false
	}
	// Don't quote booleans
	if value == "true" || value == "false" {
		return false
	}
	// Quote everything else
	return true
}

// filterStreamingTablesResult filters streaming results for current keyspace
func (p *CommandParser) filterStreamingTablesResult(streamResult db.StreamingQueryResult, currentKeyspace string) interface{} {
	// Find keyspace_name column index
	keyspaceColIdx := -1
	nameColIdx := -1
	for idx, colName := range streamResult.ColumnNames {
		switch colName {
		case "keyspace_name":
			keyspaceColIdx = idx
		case "name":
			nameColIdx = idx
		}
	}

	if keyspaceColIdx >= 0 && nameColIdx >= 0 && currentKeyspace != "" {
		// Consume the iterator and collect filtered table names
		filteredResults := [][]string{{"name"}}

		for {
			rowMap := make(map[string]interface{})
			if !streamResult.Iterator.MapScan(rowMap) {
				break
			}

			// Check if this row is for current keyspace
			if ks, ok := rowMap["keyspace_name"].(string); ok && ks == currentKeyspace {
				if name, ok := rowMap["name"].(string); ok {
					filteredResults = append(filteredResults, []string{name})
				}
			}
		}

		return db.QueryResult{
			Data:        filteredResults,
			ColumnTypes: []string{"text"},
			RowCount:    len(filteredResults) - 1,
		}
	}

	// Return original if we can't filter
	return streamResult
}

// filterQueryTablesResult filters query results for current keyspace
func (p *CommandParser) filterQueryTablesResult(queryResult db.QueryResult, currentKeyspace string) interface{} {
	if currentKeyspace == "" || len(queryResult.Data) == 0 {
		return queryResult
	}

	// Find column indices
	headers := queryResult.Data[0]
	keyspaceColIdx := -1
	nameColIdx := -1

	for idx, header := range headers {
		switch header {
		case "keyspace_name":
			keyspaceColIdx = idx
		case "name":
			nameColIdx = idx
		}
	}

	if keyspaceColIdx >= 0 && nameColIdx >= 0 {
		// Filter results
		filteredResults := [][]string{{"name"}}

		for i := 1; i < len(queryResult.Data); i++ {
			row := queryResult.Data[i]
			if len(row) > keyspaceColIdx && row[keyspaceColIdx] == currentKeyspace {
				if len(row) > nameColIdx {
					filteredResults = append(filteredResults, []string{row[nameColIdx]})
				}
			}
		}

		return db.QueryResult{
			Data:        filteredResults,
			ColumnTypes: []string{"text"},
			RowCount:    len(filteredResults) - 1,
		}
	}

	return queryResult
}