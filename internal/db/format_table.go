package db

import (
	"fmt"
	"sort"
	"strings"
)

// FormatTableCreateStatement formats a CREATE TABLE statement
// If includeHeader is true, adds "Table: keyspace.table" header (for DESCRIBE TABLE)
func FormatTableCreateStatement(tableInfo *TableInfo, includeHeader bool) string {
	var result strings.Builder

	if includeHeader {
		result.WriteString(fmt.Sprintf("Table: %s.%s\n\n", tableInfo.KeyspaceName, tableInfo.TableName))
	}

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

	// Filter out internal system properties that shouldn't be displayed
	internalProps := map[string]bool{
		"keyspace_name": true,
		"table_name":    true,
		"id":            true,
		"flags":         true,
		"dclocal_read_repair_chance": true, // Deprecated property
		"read_repair_chance":         true, // Deprecated property
	}

	for key, value := range tableInfo.TableProps {
		if internalProps[key] {
			continue
		}
		prop := formatTableProperty(key, value)
		if prop != "" {
			properties = append(properties, prop)
		}
	}

	// Sort properties for consistent output
	sort.Strings(properties)

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
		// Special handling for memtable empty string
		if name == "memtable" && v == "" {
			return fmt.Sprintf("%s = 'default'", name)
		}
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
		// Special handling for crc_check_chance to show decimal point
		if name == "crc_check_chance" {
			return fmt.Sprintf("%s = %.1f", name, v)
		}
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
		// Check if it's an empty map (from MapScan)
		if fmt.Sprintf("%v", value) == "map[]" {
			return fmt.Sprintf("%s = {}", name)
		}
		// Default formatting
		return fmt.Sprintf("%s = %v", name, v)
	}
}

// formatMapProperty formats map-type properties
func formatMapProperty(name string, m map[string]string) string {
	if len(m) == 0 {
		return fmt.Sprintf("%s = {}", name)
	}

	var parts []string
	for k, v := range m {
		// Always quote both keys and values
		parts = append(parts, fmt.Sprintf("'%s': '%s'", k, v))
	}
	sort.Strings(parts)

	return fmt.Sprintf("%s = {%s}", name, strings.Join(parts, ", "))
}

// FormatMapForCQL formats a map for CQL output (for keyspaces, indexes, etc.)
func FormatMapForCQL(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}

	var parts []string
	for k, v := range m {
		// Always quote both keys and values
		parts = append(parts, fmt.Sprintf("'%s': '%s'", k, v))
	}
	sort.Strings(parts)

	return "{" + strings.Join(parts, ", ") + "}"
}