package router

import (
	"fmt"
	"strings"
	"sort"
	"strconv"
	"github.com/axonops/cqlai/internal/logger"
)

// isVersion4OrHigher checks if the Cassandra version is 4.0 or higher
func isVersion4OrHigher(version string) bool {
	// Parse version string like "4.0.4" or "5.0.4"
	parts := strings.Split(version, ".")
	if len(parts) < 1 {
		return false
	}
	
	majorVersion, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	
	return majorVersion >= 4
}

// describeTable shows detailed information about a specific table
func (v *CqlCommandVisitorImpl) describeTable(tableName string) interface{} {
	// First check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeTable", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Try server-side DESCRIBE
		// For server-side DESCRIBE, we should pass the table name as-is
		// The server will use the current keyspace if none is specified
		describeQuery := fmt.Sprintf("DESCRIBE TABLE %s", tableName)
		logger.DebugToFile("describeTable", fmt.Sprintf("Trying server-side DESCRIBE: %s", describeQuery))
		
		iter := v.session.Query(describeQuery).Iter()
		
		// The server returns a result set with columns like 'keyspace_name', 'type', 'name', 'create_statement'
		// We need to use MapScan to get the create_statement column
		result := make(map[string]interface{})
		if iter.MapScan(result) {
			iter.Close()
			
			// Log all columns returned
			logger.DebugToFile("describeTable", fmt.Sprintf("Server-side DESCRIBE returned columns: %v", result))
			
			if createStmt, ok := result["create_statement"]; ok {
				logger.DebugToFile("describeTable", "Server-side DESCRIBE succeeded with create_statement")
				// Add a comment to show this came from server-side DESCRIBE
				return fmt.Sprintf("%v", createStmt)
			}
			
			// Maybe the column name is different?
			for key, value := range result {
				logger.DebugToFile("describeTable", fmt.Sprintf("Column '%s' = %v", key, value))
			}
		}
		
		err := iter.Close()
		if err != nil {
			logger.DebugToFile("describeTable", fmt.Sprintf("Server-side DESCRIBE error: %v", err))
		} else {
			logger.DebugToFile("describeTable", "Server-side DESCRIBE returned no results")
		}
		// If server-side DESCRIBE failed, fall back to manual construction
	}
	
	// Fall back to manual table description for pre-4.0 or if server-side failed
	logger.DebugToFile("describeTable", "Falling back to manual table description")
	
	// Check if table name includes keyspace qualification
	keyspaceName := v.session.CurrentKeyspace()
	actualTableName := tableName
	
	// Debug: Log what we're working with
	logger.DebugToFile("describeTable", fmt.Sprintf("Manual mode - tableName='%s', currentKeyspace='%s'", tableName, keyspaceName))
	
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		if len(parts) == 2 {
			keyspaceName = parts[0]
			actualTableName = parts[1]
			logger.DebugToFile("describeTable", fmt.Sprintf("Split into: keyspace='%s', table='%s'", keyspaceName, actualTableName))
		}
	} else if keyspaceName == "" {
		return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
	}

	// Debug: Let's see what tables exist in this keyspace
	debugQuery := `SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?`
	logger.DebugToFile("describeTable", fmt.Sprintf("Executing debug query with keyspace='%s'", keyspaceName))
	debugIter := v.session.Query(debugQuery, keyspaceName).Iter()
	var debugTables []string
	var debugTableName string
	for debugIter.Scan(&debugTableName) {
		debugTables = append(debugTables, debugTableName)
	}
	if err := debugIter.Close(); err != nil {
		logger.DebugToFile("describeTable", fmt.Sprintf("Error querying tables: %v", err))
	}
	logger.DebugToFile("describeTable", fmt.Sprintf("Found %d tables in keyspace '%s': %v", len(debugTables), keyspaceName, debugTables))
	
	// Check if our table exists
	tableFound := false
	for _, t := range debugTables {
		logger.DebugToFile("describeTable", fmt.Sprintf("Comparing '%s' with '%s'", t, actualTableName))
		if t == actualTableName {
			tableFound = true
			logger.DebugToFile("describeTable", "Table found!")
			break
		}
	}
	
	if !tableFound {
		availableTables := "none"
		if len(debugTables) > 0 {
			availableTables = strings.Join(debugTables, ", ")
		}
		logger.DebugToFile("describeTable", fmt.Sprintf("Table not found. Returning error"))
		return fmt.Sprintf("Table '%s' not found in keyspace '%s'. Available tables: %s", actualTableName, keyspaceName, availableTables)
	}
	
	logger.DebugToFile("describeTable", "Table found, proceeding to get table details")

	// Get table details - use SELECT * to dynamically handle all columns
	tableQuery := `SELECT * FROM system_schema.tables WHERE keyspace_name = ? AND table_name = ?`

	logger.DebugToFile("describeTable", fmt.Sprintf("Executing table details query for keyspace='%s', table='%s'", keyspaceName, actualTableName))
	iter := v.session.Query(tableQuery, keyspaceName, actualTableName).Iter()

	// Use MapScan to dynamically handle whatever columns are available
	tableProps := make(map[string]interface{})
	if !iter.MapScan(tableProps) {
		if err := iter.Close(); err != nil {
			logger.DebugToFile("describeTable", fmt.Sprintf("Table details query failed with error: %v", err))
			return fmt.Sprintf("Error getting table details: %v", err)
		}
		logger.DebugToFile("describeTable", "Table details query returned no rows")
		return fmt.Sprintf("Table '%s' not found in keyspace '%s'", actualTableName, keyspaceName)
	}
	logger.DebugToFile("describeTable", fmt.Sprintf("Successfully scanned table details: %d properties", len(tableProps)))
	
	// Debug: Log the actual types we're receiving
	for k, v := range tableProps {
		logger.DebugToFile("describeTable", fmt.Sprintf("Property %s has type %T and value %#v", k, v, v))
	}
	
	iter.Close()

	// Get columns
	colQuery := `SELECT column_name, type, kind, position 
	            FROM system_schema.columns 
	            WHERE keyspace_name = ? AND table_name = ?`

	logger.DebugToFile("describeTable", fmt.Sprintf("Executing columns query for keyspace='%s', table='%s'", keyspaceName, actualTableName))
	colIter := v.session.Query(colQuery, keyspaceName, actualTableName).Iter()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Table: %s.%s\n\n", keyspaceName, actualTableName))

	// Format CREATE TABLE statement
	result.WriteString(fmt.Sprintf("CREATE TABLE %s.%s (\n", keyspaceName, actualTableName))

type columnInfo struct {
		name     string
		dataType string
		kind     string
		position int
	}

	var columns []columnInfo
	var partitionKeys []string
	var clusteringKeys []string

	var colName, colType, colKind string
	var colPosition int

	columnCount := 0
	for colIter.Scan(&colName, &colType, &colKind, &colPosition) {
		columnCount++
		logger.DebugToFile("describeTable", fmt.Sprintf("Column %d: name='%s', type='%s', kind='%s', position=%d", columnCount, colName, colType, colKind, colPosition))
		columns = append(columns, columnInfo{
			name:     colName,
			dataType: colType,
			kind:     colKind,
			position: colPosition,
		})

		if colKind == "partition_key" {
			partitionKeys = append(partitionKeys, colName)
		} else if colKind == "clustering" {
			clusteringKeys = append(clusteringKeys, colName)
		}
	}
	if err := colIter.Close(); err != nil {
		logger.DebugToFile("describeTable", fmt.Sprintf("Column query error: %v", err))
	}
	logger.DebugToFile("describeTable", fmt.Sprintf("Found %d columns", columnCount))

	// Sort columns: partition keys first, then clustering keys, then regular columns
	sort.Slice(columns, func(i, j int) bool {
		// Priority: partition_key > clustering > regular
		kindPriority := map[string]int{
			"partition_key": 0,
			"clustering": 1,
			"regular": 2,
		}
		
		iPriority := kindPriority[columns[i].kind]
		jPriority := kindPriority[columns[j].kind]
		
		if iPriority != jPriority {
			return iPriority < jPriority
		}
		
		// Within same kind, sort by position
		return columns[i].position < columns[j].position
	})

	// Write column definitions
	// If there's only one partition key and no clustering keys, put PRIMARY KEY inline
	singlePKNoCluster := len(partitionKeys) == 1 && len(clusteringKeys) == 0
	
	for i, col := range columns {
		result.WriteString(fmt.Sprintf("    %s %s", col.name, col.dataType))
		
		// Add PRIMARY KEY inline if this is the partition key and conditions are met
		if singlePKNoCluster && col.kind == "partition_key" {
			result.WriteString(" PRIMARY KEY")
		}
		
		// Add comma after each column except the last one (unless we need a separate PRIMARY KEY clause)
		if i < len(columns)-1 {
			result.WriteString(",")
		} else if !singlePKNoCluster && len(partitionKeys) > 0 {
			result.WriteString(",")
		}
		result.WriteString("\n")
	}

	// Write PRIMARY KEY as separate line only if we have clustering keys or composite partition key
	if !singlePKNoCluster && len(partitionKeys) > 0 {
		result.WriteString("    PRIMARY KEY (")
		if len(partitionKeys) == 1 {
			result.WriteString(partitionKeys[0])
		} else {
			result.WriteString(fmt.Sprintf("(%s)", strings.Join(partitionKeys, ", ")))
		}
		if len(clusteringKeys) > 0 {
			result.WriteString(", " + strings.Join(clusteringKeys, ", "))
		}
		result.WriteString(")\n")
	}

	result.WriteString(")")

	// Build WITH clause dynamically from table properties
	// Skip certain columns that aren't table options
	skipColumns := map[string]bool{
		"keyspace_name": true,
		"table_name": true,
		"id": true,
		"flags": true,
	}
	
	// Collect property names and sort them alphabetically
	var propNames []string
	for propName := range tableProps {
		if !skipColumns[propName] {
			propNames = append(propNames, propName)
		}
	}
	sort.Strings(propNames)
	
	// Format properties in alphabetical order
	var properties []string
	for _, propName := range propNames {
		if propStr := formatTableProperty(propName, tableProps[propName]); propStr != "" {
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
			"memtable": true,  // Show empty string for null memtable
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

// describeTables lists all tables in the current keyspace.
func (v *CqlCommandVisitorImpl) describeTables() interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeTables", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Use server-side DESCRIBE TABLES
		logger.DebugToFile("describeTables", "Using server-side DESCRIBE TABLES")
		return v.session.ExecuteCQLQuery("DESCRIBE TABLES")
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := v.session.CurrentKeyspace()
	if currentKeyspace == "" {
		return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
	}

	// Query table details
	tableQuery := `SELECT table_name, gc_grace_seconds, compaction, compression 
	               FROM system_schema.tables 
	               WHERE keyspace_name = ?`
	iter := v.session.Query(tableQuery, currentKeyspace).Iter()

type tableDetails struct {
		name        string
		gcGrace     int
		compaction  map[string]string
		compression map[string]string
		primaryKeys []string
	}

	var tables []tableDetails
	var tableName string
	var gcGrace int
	var compaction, compression map[string]string

	// Collect table details
	for iter.Scan(&tableName, &gcGrace, &compaction, &compression) {
		tables = append(tables, tableDetails{
			name:        tableName,
			gcGrace:     gcGrace,
			compaction:  compaction,
			compression: compression,
		})
	}
	if err := iter.Close(); err != nil {
		return fmt.Errorf("error listing tables: %v", err)
	}

	if len(tables) == 0 {
		return fmt.Sprintf("No tables found in keyspace %s", currentKeyspace)
	}

	// Get primary keys for each table
	for i := range tables {
		// First, let's debug what columns exist
		debugQuery := `SELECT column_name, type, kind 
		              FROM system_schema.columns 
		              WHERE keyspace_name = ? AND table_name = ?`

		debugIter := v.session.Query(debugQuery, currentKeyspace, tables[i].name).Iter()

		var colName, colType, colKind string
		var pkNames []string
		var ckNames []string

		// Collect all columns and check their kinds
		hasColumns := false
		for debugIter.Scan(&colName, &colType, &colKind) {
			hasColumns = true
			// Check if this is a primary key column
			if colKind == "partition_key" {
				pkNames = append(pkNames, colName)
			} else if colKind == "clustering" {
				ckNames = append(ckNames, colName)
			}
		}

		// If no columns found at all, the table might be special
		if !hasColumns {
			// Maybe try without the type column which might not exist
			simpleQuery := `SELECT column_name, kind 
			               FROM system_schema.columns 
			               WHERE keyspace_name = ? AND table_name = ?`
			simpleIter := v.session.Query(simpleQuery, currentKeyspace, tables[i].name).Iter()

			var sColName, sColKind string
			for simpleIter.Scan(&sColName, &sColKind) {
				if sColKind == "partition_key" {
					pkNames = append(pkNames, sColName)
				} else if sColKind == "clustering" {
					ckNames = append(ckNames, sColName)
				}
			}
			simpleIter.Close()
		}

		if err := debugIter.Close(); err != nil {
			// If there's an error, maybe columns table doesn't have the expected structure
			// Try a simpler approach - just mark as unknown
			tables[i].primaryKeys = []string{fmt.Sprintf("Error: %v", err)}
		}

		// Format primary key
		if len(pkNames) > 0 {
			if len(pkNames) == 1 && len(ckNames) == 0 {
				// Single partition key only
				tables[i].primaryKeys = pkNames
			} else if len(pkNames) == 1 && len(ckNames) > 0 {
				// Single partition key with clustering keys
				tables[i].primaryKeys = append(pkNames, ckNames...)
			} else {
				// Composite partition key
				tables[i].primaryKeys = append([]string{fmt.Sprintf("(%s)", strings.Join(pkNames, ","))}, ckNames...)
			}
		}
	}

	// Sort by table name
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].name < tables[j].name
	})

	// Build results table
	results := [][]string{{"Table", "Primary Key", "Compaction", "Compression", "GC Grace"}}

	for _, t := range tables {
		// Format primary key (comma-separated if composite)
		pkStr := strings.Join(t.primaryKeys, ", ")
		if pkStr == "" {
			// If no primary keys found, it might be a virtual table or something went wrong
			// Let's try to show something useful
			pkStr = "?"
		}

		// Extract compaction strategy (shorten name)
		compactionStr := "Unknown"
		if class, ok := t.compaction["class"]; ok {
			// Get just the strategy name without the full class path
			parts := strings.Split(class, ".")
			strategy := parts[len(parts)-1]
			// Shorten common strategies
			switch strategy {
			case "SizeTieredCompactionStrategy":
				compactionStr = "STCS"
			case "LeveledCompactionStrategy":
				compactionStr = "LCS"
			case "TimeWindowCompactionStrategy":
				compactionStr = "TWCS"
			case "UnifiedCompactionStrategy":
				compactionStr = "UCS"
			default:
				compactionStr = strategy
			}
		}

		// Extract compression
		compressionStr := "None"
		if class, ok := t.compression["class"]; ok {
			parts := strings.Split(class, ".")
			compressionStr = parts[len(parts)-1]
			// Shorten common compression algorithms
			if compressionStr == "LZ4Compressor" {
				compressionStr = "LZ4"
			} else if compressionStr == "SnappyCompressor" {
				compressionStr = "Snappy"
			} else if compressionStr == "DeflateCompressor" {
				compressionStr = "Deflate"
			}
		}

		// Format GC grace (convert seconds to days/hours)
		gcGraceStr := fmt.Sprintf("%d", t.gcGrace)
		if t.gcGrace >= 86400 {
			days := t.gcGrace / 86400
			gcGraceStr = fmt.Sprintf("%dd", days)
		} else if t.gcGrace >= 3600 {
			hours := t.gcGrace / 3600
			gcGraceStr = fmt.Sprintf("%dh", hours)
		} else if t.gcGrace > 0 {
			gcGraceStr = fmt.Sprintf("%ds", t.gcGrace)
		}

		results = append(results, []string{
			t.name,
			pkStr,
			compactionStr,
			compressionStr,
			gcGraceStr,
		})
	}

	return results
}
