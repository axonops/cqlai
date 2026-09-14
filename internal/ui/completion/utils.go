package completion

import "strings"

// isColumnName checks if a word could be a column name
func (ce *CompletionEngine) isColumnName(word string, words []string, fromIndex int) bool {
	if fromIndex < 0 {
		return false
	}

	columns := ce.getColumnNamesForCurrentTable(words, fromIndex)
	for _, col := range columns {
		if strings.EqualFold(col, word) {
			return true
		}
	}
	return false
}

// isColumnNameForTable checks if a word is a column name for a specific table
func (ce *CompletionEngine) isColumnNameForTable(word string, tableName string) bool {
	columns := ce.getColumnNamesForTable(tableName)
	for _, col := range columns {
		if strings.EqualFold(col, word) {
			return true
		}
	}
	return false
}

// isComparisonOperator checks if a word is a comparison operator
func (ce *CompletionEngine) isComparisonOperator(word string) bool {
	for _, op := range ComparisonOperators {
		if word == op {
			return true
		}
	}
	return false
}

// IsCompleteKeyword checks if a word is a complete CQL keyword
// This is exported for use by other packages (e.g., ui package)
func IsCompleteKeyword(word string) bool {
	upperWord := strings.ToUpper(word)

	// Check against all keyword lists from constants
	// Combine relevant keyword lists for comprehensive checking
	allKeywords := append([]string{}, TopLevelKeywords...)
	allKeywords = append(allKeywords, SelectKeywords...)
	allKeywords = append(allKeywords, FromKeyword...)
	allKeywords = append(allKeywords, WhereKeyword...)
	allKeywords = append(allKeywords, SetKeyword...)
	allKeywords = append(allKeywords, ValuesKeyword...)
	allKeywords = append(allKeywords, IntoKeyword...)
	allKeywords = append(allKeywords, ByKeyword...)
	allKeywords = append(allKeywords, LimitKeyword...)
	allKeywords = append(allKeywords, LogicalOperators...)
	allKeywords = append(allKeywords, SortOrders...)
	allKeywords = append(allKeywords, UsingOptions...)
	allKeywords = append(allKeywords, IfClauseKeywords...)
	allKeywords = append(allKeywords, DDLObjectTypes...)
	allKeywords = append(allKeywords, BatchTypes...)
	allKeywords = append(allKeywords, FilteringKeyword...)
	allKeywords = append(allKeywords, ConsistencyLevels...)
	allKeywords = append(allKeywords, AggregateFunctions...)
	allKeywords = append(allKeywords, "ALLOW", "PRIMARY", "KEY", "WITH", "DISTINCT", "TOKEN", "JSON", "PER", "PARTITION", "ASCII", "HAVING")

	for _, kw := range allKeywords {
		if upperWord == kw {
			return true
		}
	}
	return false
}

// getSuggestionsAfterWhere returns suggestions after WHERE clause
func (ce *CompletionEngine) getSuggestionsAfterWhere(words []string, hasOrderBy, hasGroupBy, hasLimit, hasAllowFiltering, hasPerPartitionLimit bool) []string {
	var suggestions []string

	// Can always add more conditions
	suggestions = append(suggestions, "AND", "OR")

	if !hasGroupBy {
		suggestions = append(suggestions, "GROUP")
	}
	if !hasOrderBy {
		suggestions = append(suggestions, "ORDER")
	}
	if !hasLimit {
		suggestions = append(suggestions, "LIMIT")
	}
	if !hasPerPartitionLimit && hasOrderBy {
		suggestions = append(suggestions, "PER")
	}
	if !hasAllowFiltering {
		suggestions = append(suggestions, "ALLOW")
	}

	return suggestions
}

// getSuggestionsAfterOrderBy returns suggestions after ORDER BY clause
func (ce *CompletionEngine) getSuggestionsAfterOrderBy(words []string, hasWhere, hasLimit, hasAllowFiltering, hasPerPartitionLimit bool) []string {
	var suggestions []string

	if !hasLimit {
		suggestions = append(suggestions, "LIMIT")
	}
	if !hasPerPartitionLimit {
		suggestions = append(suggestions, "PER")
	}
	if !hasAllowFiltering && hasWhere {
		suggestions = append(suggestions, "ALLOW")
	}

	return suggestions
}

// getTopLevelCommands returns top-level CQL commands
func (ce *CompletionEngine) getTopLevelCommands() []string {
	return TopLevelCommands
}

// IsObjectType checks if a word is a database object type
func IsObjectType(word string) bool {
	upperWord := strings.ToUpper(word)
	for _, obj := range DDLObjectTypes {
		if upperWord == obj {
			return true
		}
	}
	return false
}

// GetFollowKeywords returns keywords that can follow a given keyword
func GetFollowKeywords(keyword string) []string {
	switch strings.ToUpper(keyword) {
	case "CREATE":
		return CreateObjectTypes
	case "DROP":
		return DDLObjectTypes
	case "ALTER":
		return AlterObjectTypes
	case "MATERIALIZED":
		return MaterializedKeyword
	case "IF":
		return IfClauseKeywords
	case "NOT":
		return ExistsKeyword
	case "ORDER":
		return ByKeyword
	case "GROUP":
		return ByKeyword
	case "PRIMARY":
		return KeyKeyword
	case "ALLOW":
		return FilteringKeyword
	case "WITH":
		return KeyspaceOptions
	case "GRANT":
		return CQLPermissions
	case "REVOKE":
		return CQLPermissions
	case "USING":
		return UsingOptions
	case "TRUNCATE":
		return TableKeyword
	default:
		return []string{}
	}
}

// GetCommandObjects returns object names for DDL commands
func GetCommandObjects(command, objectType string) []string {
	switch strings.ToUpper(command) {
	case "CREATE", "DROP", "ALTER":
		switch strings.ToUpper(objectType) {
		case "TABLE":
			return []string{} // Will be filled with actual table names
		case "KEYSPACE":
			return []string{} // Will be filled with actual keyspace names
		default:
			return []string{}
		}
	default:
		return []string{}
	}
}

// takesNoArguments are the functions written with an empty pair of brackets.
//
// Everything else is completed with the bracket open, waiting for a column.
var takesNoArguments = map[string]bool{
	"now": true, "uuid": true,
	"currentTimeUUID": true, "currentTimestamp": true, "currentDate": true,
	"CURRENT_DATE": true, "CURRENT_TIME": true,
	"CURRENT_TIMESTAMP": true, "CURRENT_TIMEUUID": true,
}

// functionFamilies is every function a SELECT can use, in the families
// Cassandra groups them in: one place to add to when Cassandra gains another.
func functionFamilies() [][]string {
	return [][]string{
		AggregateFunctions,
		ScalarMathFunctions,
		CollectionFunctions,
		MaskFunctions,
		TimeFunctions,
		SystemFunctions,
	}
}

// getFunctionSuggestions returns CQL function suggestions for SELECT clause
func (ce *CompletionEngine) getFunctionSuggestions() []string {
	suggestions := make([]string, 0, 64)

	for _, family := range functionFamilies() {
		for _, fn := range family {
			if takesNoArguments[fn] {
				suggestions = append(suggestions, fn+"()")
				continue
			}
			suggestions = append(suggestions, fn+"(")
		}
	}
	return suggestions
}

// getTableAndKeyspaceNames returns both table names and keyspace names
func (ce *CompletionEngine) getTableAndKeyspaceNames() []string {
	var suggestions []string

	// Add table names from current keyspace
	suggestions = append(suggestions, ce.getTableNames()...)

	// Add keyspace names (for keyspace.table syntax)
	keyspaces := ce.getKeyspaceNames()
	for _, ks := range keyspaces {
		// Add keyspace name with a dot to indicate it expects a table name after
		suggestions = append(suggestions, ks+".")
	}

	return suggestions
}

// FindCommonPrefix finds the common prefix of all completion options
func FindCommonPrefix(completions []string) string {
	if len(completions) == 0 {
		return ""
	}
	if len(completions) == 1 {
		return completions[0]
	}

	prefix := completions[0]
	for _, s := range completions[1:] {
		for len(prefix) > 0 && !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
		if len(prefix) == 0 {
			break
		}
	}
	return prefix
}

// getTableNames returns table names from the completion engine

// getKeyspaceNames returns keyspace names from the completion engine

// getColumnNamesForTable returns column names for a specific table

// getTableAndKeyspaceNames returns both table names and keyspace names for INSERT/SELECT

// getTablesForKeyspace returns table names for a specific keyspace

// isCQLKeyword checks if a token is a CQL keyword

// getTopLevelKeywords returns all top-level CQL keywords
