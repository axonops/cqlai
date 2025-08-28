package completion

import (
	"strconv"
	"strings"
)

// completionCache stores database metadata for faster completions
type completionCache struct {
	keyspaces []string
	tables    map[string][]string // keyspace -> tables
	columns   map[string][]string // keyspace.table -> columns
}

// getKeyspaceNames returns cached keyspace names or fetches them
func (ce *CompletionEngine) getKeyspaceNames() []string {
	if len(ce.cache.keyspaces) == 0 {
		ce.refreshKeyspaceCache()
	}
	return ce.cache.keyspaces
}

// getTableNames returns table names for current keyspace
func (ce *CompletionEngine) getTableNames() []string {
	currentKeyspace := ""
	if ce.sessionManager != nil {
		currentKeyspace = ce.sessionManager.CurrentKeyspace()
	}
	if currentKeyspace == "" {
		return []string{}
	}
	
	if _, ok := ce.cache.tables[currentKeyspace]; !ok {
		ce.refreshTableCache(currentKeyspace)
	}
	
	return ce.cache.tables[currentKeyspace]
}

// getTableAndKeyspaceTableNames returns both local table names and keyspace.table combinations
func (ce *CompletionEngine) getTableAndKeyspaceTableNames() []string {
	var suggestions []string
	
	// Add tables from current keyspace (unqualified)
	currentKeyspace := ""
	if ce.sessionManager != nil {
		currentKeyspace = ce.sessionManager.CurrentKeyspace()
	}
	if currentKeyspace != "" {
		suggestions = append(suggestions, ce.getTableNames()...)
	}
	
	// Add keyspace.table combinations for all keyspaces
	keyspaces := ce.getKeyspaceNames()
	for _, ks := range keyspaces {
		tables := ce.getTablesForKeyspace(ks)
		for _, table := range tables {
			suggestions = append(suggestions, ks+"."+table)
		}
	}
	
	return suggestions
}

// getTablesForKeyspace returns table names for a specific keyspace
func (ce *CompletionEngine) getTablesForKeyspace(keyspace string) []string {
	if keyspace == "" {
		return []string{}
	}
	
	if _, ok := ce.cache.tables[keyspace]; !ok {
		ce.refreshTableCache(keyspace)
	}
	
	return ce.cache.tables[keyspace]
}

// getTypeNames returns type names (placeholder for now)
func (ce *CompletionEngine) getTypeNames() []string {
	// TODO: Implement type name caching
	return []string{}
}

// getFunctionNames returns function names (placeholder for now)
func (ce *CompletionEngine) getFunctionNames() []string {
	// TODO: Implement function name caching
	return []string{}
}

// getAggregateNames returns aggregate names (placeholder for now)
func (ce *CompletionEngine) getAggregateNames() []string {
	// TODO: Implement aggregate name caching
	return []string{}
}

// getIndexNames returns index names (placeholder for now)
func (ce *CompletionEngine) getIndexNames() []string {
	// TODO: Implement index name caching
	return []string{}
}

// getViewNames returns view names (placeholder for now)
func (ce *CompletionEngine) getViewNames() []string {
	// TODO: Implement view name caching
	return []string{}
}

// refreshKeyspaceCache updates the keyspace cache
func (ce *CompletionEngine) refreshKeyspaceCache() {
	if ce.session == nil {
		return
	}
	
	iter := ce.session.Query("SELECT keyspace_name FROM system_schema.keyspaces").Iter()
	ce.cache.keyspaces = []string{}
	
	var keyspaceName string
	for iter.Scan(&keyspaceName) {
		ce.cache.keyspaces = append(ce.cache.keyspaces, keyspaceName)
	}
	iter.Close()
}

// refreshTableCache updates the table cache for a specific keyspace
func (ce *CompletionEngine) refreshTableCache(keyspace string) {
	if ce.session == nil {
		return
	}
	
	iter := ce.session.Query(
		"SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?",
		keyspace,
	).Iter()
	
	ce.cache.tables[keyspace] = []string{}
	
	var tableName string
	for iter.Scan(&tableName) {
		ce.cache.tables[keyspace] = append(ce.cache.tables[keyspace], tableName)
	}
	iter.Close()
}

// refreshColumnCache updates the column cache for a specific table
func (ce *CompletionEngine) refreshColumnCache(keyspace, table string) {
	if ce.session == nil {
		return
	}
	
	iter := ce.session.Query(
		"SELECT column_name FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ?",
		keyspace, table,
	).Iter()
	
	cacheKey := keyspace + "." + table
	ce.cache.columns[cacheKey] = []string{}
	
	var columnName string
	for iter.Scan(&columnName) {
		ce.cache.columns[cacheKey] = append(ce.cache.columns[cacheKey], columnName)
	}
	iter.Close()
}

// getColumnNamesForCurrentTable returns column names for the table in FROM clause
func (ce *CompletionEngine) getColumnNamesForCurrentTable(words []string, fromIndex int) []string {
	if fromIndex < 0 || fromIndex+1 >= len(words) {
		return []string{}
	}
	
	tableName := words[fromIndex+1]
	currentKeyspace := ""
	if ce.sessionManager != nil {
		currentKeyspace = ce.sessionManager.CurrentKeyspace()
	}
	
	// Check if table name includes keyspace
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		currentKeyspace = parts[0]
		tableName = parts[1]
	}
	
	if currentKeyspace == "" {
		return []string{}
	}
	
	// Get cached columns or fetch them
	cacheKey := currentKeyspace + "." + tableName
	if columns, ok := ce.cache.columns[cacheKey]; ok {
		return columns
	}
	
	// Fetch columns from database
	ce.refreshColumnCache(currentKeyspace, tableName)
	return ce.cache.columns[cacheKey]
}

// getColumnNamesForTable returns column names for a specific table
func (ce *CompletionEngine) getColumnNamesForTable(tableName string) []string {
	currentKeyspace := ""
	if ce.sessionManager != nil {
		currentKeyspace = ce.sessionManager.CurrentKeyspace()
	}
	
	// Check if table name includes keyspace
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		currentKeyspace = parts[0]
		tableName = parts[1]
	}
	
	if currentKeyspace == "" {
		return []string{}
	}
	
	// Get cached columns or fetch them
	cacheKey := currentKeyspace + "." + tableName
	if columns, ok := ce.cache.columns[cacheKey]; ok {
		return columns
	}
	
	// Fetch columns from database
	ce.refreshColumnCache(currentKeyspace, tableName)
	return ce.cache.columns[cacheKey]
}

// getColumnTypeTemplate returns a template string with data types for the specified columns
func (ce *CompletionEngine) getColumnTypeTemplate(tableName string, columnNames []string) string {
	currentKeyspace := ""
	if ce.sessionManager != nil {
		currentKeyspace = ce.sessionManager.CurrentKeyspace()
	}
	
	// Check if table name includes keyspace
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		currentKeyspace = parts[0]
		tableName = parts[1]
	}
	
	if currentKeyspace == "" || ce.session == nil {
		return ""
	}
	
	// Query column types from system_schema
	iter := ce.session.Query(
		"SELECT column_name, type FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ?",
		currentKeyspace, tableName,
	).Iter()
	
	columnTypes := make(map[string]string)
	var columnName, columnType string
	for iter.Scan(&columnName, &columnType) {
		columnTypes[columnName] = columnType
	}
	iter.Close()
	
	// Build the template
	var templates []string
	for _, col := range columnNames {
		if colType, ok := columnTypes[col]; ok {
			templates = append(templates, ce.getTypeTemplate(colType))
		} else {
			templates = append(templates, "?")
		}
	}
	
	return strings.Join(templates, ", ")
}

// getTypeTemplate returns a template value for a given CQL type
func (ce *CompletionEngine) getTypeTemplate(cqlType string) string {
	// Handle common CQL types
	switch {
	case strings.HasPrefix(cqlType, "text") || strings.HasPrefix(cqlType, "varchar"):
		return "'text'"
	case strings.HasPrefix(cqlType, "int") || cqlType == "varint":
		return "0"
	case strings.HasPrefix(cqlType, "bigint"):
		return "0"
	case strings.HasPrefix(cqlType, "float") || strings.HasPrefix(cqlType, "double") || strings.HasPrefix(cqlType, "decimal"):
		return "0.0"
	case strings.HasPrefix(cqlType, "boolean"):
		return "true"
	case strings.HasPrefix(cqlType, "uuid"):
		return "uuid()"
	case strings.HasPrefix(cqlType, "timestamp"):
		return "'2024-01-01T00:00:00Z'"
	case strings.HasPrefix(cqlType, "date"):
		return "'2024-01-01'"
	case strings.HasPrefix(cqlType, "time"):
		return "'00:00:00'"
	case strings.HasPrefix(cqlType, "blob"):
		return "0x"
	case strings.HasPrefix(cqlType, "vector<"):
		// Extract vector dimensions if possible
		if strings.Contains(cqlType, "float") {
			// Try to extract dimension
			if start := strings.Index(cqlType, ","); start > 0 {
				if end := strings.Index(cqlType[start:], ">"); end > 0 {
					dimStr := strings.TrimSpace(cqlType[start+1 : start+end])
					if dim, err := strconv.Atoi(dimStr); err == nil && dim > 0 {
						// Create a vector template with the right dimension
						values := make([]string, dim)
						for i := range values {
							values[i] = "0.0"
						}
						return "[" + strings.Join(values, ", ") + "]"
					}
				}
			}
			return "[0.0, 0.0, 0.0]"
		}
		return "[]"
	case strings.HasPrefix(cqlType, "list<"):
		return "[]"
	case strings.HasPrefix(cqlType, "set<"):
		return "{}"
	case strings.HasPrefix(cqlType, "map<"):
		return "{}"
	default:
		return "null"
	}
}

// getColumnNamesForCurrentContext returns column names based on current query context
func (ce *CompletionEngine) getColumnNamesForCurrentContext(words []string) []string {
	// Try to find a table reference in the query
	for i, word := range words {
		if strings.EqualFold(word, "FROM") && i+1 < len(words) {
			return ce.getColumnNamesForTable(words[i+1])
		}
		if strings.EqualFold(word, "UPDATE") && i+1 < len(words) {
			return ce.getColumnNamesForTable(words[i+1])
		}
		if strings.EqualFold(word, "INTO") && i+1 < len(words) {
			return ce.getColumnNamesForTable(words[i+1])
		}
	}
	return []string{}
}