package router

import (
	"fmt"
	"sort"
	"strings"
)

// describeTable shows detailed information about a specific table
func (v *CqlCommandVisitorImpl) describeTable(tableName string) interface{} {
	serverResult, tableInfo, err := v.session.DBDescribeTable(sessionManager, tableName)

	if err != nil {
		if err.Error() == "no keyspace selected" {
			return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
		}
		if strings.Contains(err.Error(), "not found") {
			return err.Error()
		}
		return fmt.Sprintf("Error: %v", err)
	}

	if serverResult != nil {
		// Server-side DESCRIBE result, return as-is
		return serverResult
	}

	// Manual query result - format it
	if tableInfo == nil {
		return fmt.Sprintf("Table '%s' not found", tableName)
	}

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

		// Add comma after each column except the last one (unless we need a separate PRIMARY KEY clause)
		if i < len(tableInfo.Columns)-1 {
			result.WriteString(",")
		} else if !singlePKNoCluster && len(tableInfo.PartitionKeys) > 0 {
			result.WriteString(",")
		}
		result.WriteString("\n")
	}

	// Write PRIMARY KEY as separate line only if we have clustering keys or composite partition key
	if !singlePKNoCluster && len(tableInfo.PartitionKeys) > 0 {
		result.WriteString("    PRIMARY KEY (")
		if len(tableInfo.PartitionKeys) == 1 {
			result.WriteString(tableInfo.PartitionKeys[0])
		} else {
			result.WriteString(fmt.Sprintf("(%s)", strings.Join(tableInfo.PartitionKeys, ", ")))
		}
		if len(tableInfo.ClusteringKeys) > 0 {
			result.WriteString(", " + strings.Join(tableInfo.ClusteringKeys, ", "))
		}
		result.WriteString(")\n")
	}

	result.WriteString(")")

	// Build WITH clause dynamically from table properties
	// Skip certain columns that aren't table options
	skipColumns := map[string]bool{
		"keyspace_name": true,
		"table_name":    true,
		"id":            true,
		"flags":         true,
	}

	// Collect property names and sort them alphabetically
	var propNames []string
	for propName := range tableInfo.TableProps {
		if !skipColumns[propName] {
			propNames = append(propNames, propName)
		}
	}
	sort.Strings(propNames)

	// Format properties in alphabetical order
	var properties []string
	for _, propName := range propNames {
		if propStr := formatTableProperty(propName, tableInfo.TableProps[propName]); propStr != "" {
			properties = append(properties, propStr)
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
		// For pre-4.0, we can't determine the actual server config value
		// So we'll show what we know from the system tables
		// Note: In 4.0+, server-side DESCRIBE handles this correctly
		showNullFor := map[string]bool{
			"memtable": true, // Show empty string for null memtable
		}
		if showNullFor[name] {
			if name == "memtable" {
				return fmt.Sprintf("%s = ''", name)
			}
			return fmt.Sprintf("%s = null", name)
		}
		// Skip properties that are null (they use server defaults)
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
		// Unquoted strings (shouldn't happen but just in case)
		return fmt.Sprintf("%s = %s", name, v)

	case bool:
		return fmt.Sprintf("%s = %v", name, v)

	case int, int32, int64:
		return fmt.Sprintf("%s = %v", name, v)

	case float32, float64:
		return fmt.Sprintf("%s = %g", name, v)

	case map[string]string:
		// Handle frozen<map<text, text>> columns from Cassandra
		if len(v) == 0 {
			return fmt.Sprintf("%s = {}", name)
		}
		return fmt.Sprintf("%s = %s", name, formatMap(v))

	case map[string][]byte:
		// Handle frozen<map<text, blob>> columns like extensions
		if len(v) == 0 {
			return fmt.Sprintf("%s = {}", name)
		}
		// Convert to string map for display
		strMap := make(map[string]string)
		for k, val := range v {
			// Convert bytes to hex string for display
			strMap[k] = fmt.Sprintf("0x%x", val)
		}
		return fmt.Sprintf("%s = %s", name, formatMap(strMap))

	case map[string]interface{}:
		// Convert to map[string]string for formatting
		if len(v) == 0 {
			return fmt.Sprintf("%s = {}", name)
		}
		strMap := make(map[string]string)
		for k, val := range v {
			strMap[k] = fmt.Sprint(val)
		}
		return fmt.Sprintf("%s = %s", name, formatMap(strMap))

	case []string:
		// Handle frozen<set<text>> columns
		if len(v) == 0 {
			return fmt.Sprintf("%s = {}", name)
		}
		var items []string
		for _, item := range v {
			items = append(items, fmt.Sprintf("'%s'", item))
		}
		return fmt.Sprintf("%s = {%s}", name, strings.Join(items, ", "))

	case []interface{}:
		// Handle sets like flags
		if len(v) == 0 {
			return fmt.Sprintf("%s = {}", name)
		}
		var items []string
		for _, item := range v {
			items = append(items, fmt.Sprintf("'%v'", item))
		}
		return fmt.Sprintf("%s = {%s}", name, strings.Join(items, ", "))

	default:
		// For any other type, use default formatting
		return fmt.Sprintf("%s = %v", name, v)
	}
}

// formatMap formats a map for CQL output
func formatMap(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}

	// Define the standard property order for better readability
	propertyOrder := []string{"class"}

	var pairs []string
	processed := make(map[string]bool)

	// First add properties in preferred order
	for _, key := range propertyOrder {
		if value, exists := m[key]; exists {
			pairs = append(pairs, fmt.Sprintf("'%s': '%s'", key, value))
			processed[key] = true
		}
	}

	// Then add remaining properties in sorted order
	var remainingKeys []string
	for k := range m {
		if !processed[k] {
			remainingKeys = append(remainingKeys, k)
		}
	}
	sort.Strings(remainingKeys)

	for _, k := range remainingKeys {
		pairs = append(pairs, fmt.Sprintf("'%s': '%s'", k, m[k]))
	}

	// Join all pairs - ensure we include everything
	return "{" + strings.Join(pairs, ", ") + "}"
}

// Map common compaction strategies to their abbreviations
var compactionStrategyAbbrev = map[string]string{
	"SizeTieredCompactionStrategy": "STCS",
	"LeveledCompactionStrategy":    "LCS",
	"TimeWindowCompactionStrategy": "TWCS",
	"UnifiedCompactionStrategy":    "UCS",
}

// Map common compression algorithms to their abbreviations
var compressionAbbrev = map[string]string{
	"LZ4Compressor":     "LZ4",
	"SnappyCompressor":  "Snappy",
	"DeflateCompressor": "Deflate",
}

// describeTables lists all tables in the current keyspace.
func (v *CqlCommandVisitorImpl) describeTables() interface{} {
	serverResult, tables, err := v.session.DBDescribeTables(sessionManager)

	if err != nil {
		if err.Error() == "no keyspace selected" {
			return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
		}
		return fmt.Sprintf("Error: %v", err)
	}

	if serverResult != nil {
		// Server-side DESCRIBE result, return as-is
		return serverResult
	}

	// Manual query result - format it
	if len(tables) == 0 {
		currentKeyspace := ""
		if sessionManager != nil {
			currentKeyspace = sessionManager.CurrentKeyspace()
		}
		return fmt.Sprintf("No tables found in keyspace %s", currentKeyspace)
	}

	// Build results table
	results := [][]string{{"Table", "Primary Key", "Compaction", "Compression", "GC Grace"}}

	for _, t := range tables {
		// Format primary key
		var primaryKeys []string
		if len(t.PartitionKeys) > 0 {
			if len(t.PartitionKeys) == 1 && len(t.ClusteringKeys) == 0 { //nolint:gocritic // more readable as if
				// Single partition key only
				primaryKeys = t.PartitionKeys
			} else if len(t.PartitionKeys) == 1 && len(t.ClusteringKeys) > 0 {
				// Single partition key with clustering keys
				primaryKeys = make([]string, 0, len(t.PartitionKeys)+len(t.ClusteringKeys))
				primaryKeys = append(primaryKeys, t.PartitionKeys...)
				primaryKeys = append(primaryKeys, t.ClusteringKeys...)
			} else {
				// Composite partition key
				primaryKeys = append([]string{fmt.Sprintf("(%s)", strings.Join(t.PartitionKeys, ","))}, t.ClusteringKeys...)
			}
		}

		pkStr := strings.Join(primaryKeys, ", ")
		if pkStr == "" {
			pkStr = "?"
		}

		// Extract compaction strategy (shorten name)
		compactionStr := "Unknown"
		if class, ok := t.Compaction["class"]; ok {
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

		// Extract compression
		compressionStr := "None"
		if class, ok := t.Compression["class"]; ok {
			parts := strings.Split(class, ".")
			algorithm := parts[len(parts)-1]
			// Use abbreviation if available, otherwise use the algorithm name
			if abbrev, found := compressionAbbrev[algorithm]; found {
				compressionStr = abbrev
			} else {
				compressionStr = algorithm
			}
		}

		// Format GC grace (convert seconds to days/hours)
		gcGraceStr := fmt.Sprintf("%d", t.GcGrace)
		if t.GcGrace >= 86400 { //nolint:gocritic // more readable as if
			days := t.GcGrace / 86400
			gcGraceStr = fmt.Sprintf("%dd", days)
		} else if t.GcGrace >= 3600 {
			hours := t.GcGrace / 3600
			gcGraceStr = fmt.Sprintf("%dh", hours)
		} else if t.GcGrace > 0 {
			gcGraceStr = fmt.Sprintf("%ds", t.GcGrace)
		}

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
