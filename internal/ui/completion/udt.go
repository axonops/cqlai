package completion

import (
	"strings"

	"github.com/axonops/cqlai/internal/db"
)

// handleUDTFieldCompletion handles completion for UDT field access (column.field)
func (ce *CompletionEngine) handleUDTFieldCompletion(input string) []string {
	// Look for pattern: column_name.partial_field
	// This needs to be context-aware - we need to know which table we're working with

	// Split the input to find the last word with a dot
	words := strings.Fields(input)
	if len(words) == 0 {
		return nil
	}

	lastWord := words[len(words)-1]
	if !strings.Contains(lastWord, ".") {
		return nil
	}

	// Split the last word by dot
	parts := strings.Split(lastWord, ".")
	if len(parts) != 2 {
		return nil // Only handle single-level field access for now
	}

	columnName := parts[0]
	partialField := parts[1]

	// Find the table context from the query
	tableName := ce.findTableContext(words)
	if tableName == "" {
		return nil
	}

	// Get UDT field completions
	fields := ce.getUDTFieldsForColumn(tableName, columnName)
	if len(fields) == 0 {
		return nil
	}

	// Filter fields based on partial match
	var suggestions []string
	upperPartial := strings.ToUpper(partialField)
	for _, field := range fields {
		if strings.HasPrefix(strings.ToUpper(field), upperPartial) {
			// Return just the field name, not the full column.field
			suggestions = append(suggestions, field)
		}
	}

	return suggestions
}

// findTableContext finds the table being referenced in the current query
func (ce *CompletionEngine) findTableContext(words []string) string {
	// Look for common patterns: FROM table, UPDATE table, INSERT INTO table
	for i := 0; i < len(words)-1; i++ {
		upperWord := strings.ToUpper(words[i])
		switch upperWord {
		case "FROM", "UPDATE":
			if i+1 < len(words) {
				return strings.Trim(words[i+1], ",;")
			}
		case "INTO":
			if i > 0 && strings.ToUpper(words[i-1]) == "INSERT" && i+1 < len(words) {
				return strings.Trim(words[i+1], ",;")
			}
		}
	}
	return ""
}

// getUDTFieldsForColumn returns the field names for a UDT column
func (ce *CompletionEngine) getUDTFieldsForColumn(tableName, columnName string) []string {
	if ce.session == nil {
		return nil
	}

	currentKeyspace := ce.sessionManager.CurrentKeyspace()

	// Check if table name includes keyspace
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		currentKeyspace = parts[0]
		tableName = parts[1]
	}

	if currentKeyspace == "" {
		return nil
	}

	// Get column type information
	var columnType string
	query := `SELECT type FROM system_schema.columns
	          WHERE keyspace_name = ? AND table_name = ? AND column_name = ?`

	iter := ce.session.Query(query, currentKeyspace, tableName, columnName).Iter()
	if !iter.Scan(&columnType) {
		_ = iter.Close()
		return nil
	}
	_ = iter.Close()

	// Parse the column type to check if it's a UDT
	typeInfo, err := db.ParseCQLType(columnType)
	if err != nil || typeInfo.BaseType != "udt" {
		return nil
	}

	// Get UDT field names from system_schema.types
	var fieldNames []string
	udtQuery := `SELECT field_names FROM system_schema.types
	             WHERE keyspace_name = ? AND type_name = ?`

	udtKeyspace := currentKeyspace
	if typeInfo.Keyspace != "" {
		udtKeyspace = typeInfo.Keyspace
	}

	iter = ce.session.Query(udtQuery, udtKeyspace, typeInfo.UDTName).Iter()
	if !iter.Scan(&fieldNames) {
		_ = iter.Close()
		return nil
	}
	_ = iter.Close()

	return fieldNames
}

// isUDTFieldAccess checks if the input is trying to access a UDT field
func (ce *CompletionEngine) isUDTFieldAccess(input string) bool {
	// Look for pattern: word.partial_word at the end of input
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return false
	}

	// Get the last word
	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return false
	}

	lastWord := words[len(words)-1]

	// Check if it contains a dot but is not a keyspace.table pattern
	if !strings.Contains(lastWord, ".") {
		return false
	}

	// Make sure it's not a keyspace.table pattern (those are handled elsewhere)
	// UDT field access would be in SELECT, WHERE, SET contexts
	for i := len(words) - 2; i >= 0; i-- {
		upperWord := strings.ToUpper(words[i])
		if upperWord == "FROM" || upperWord == "UPDATE" || upperWord == "INTO" {
			// This is likely a keyspace.table pattern, not UDT field access
			return false
		}
		if upperWord == "SELECT" || upperWord == "WHERE" || upperWord == "SET" {
			// This could be UDT field access
			return true
		}
	}

	return false
}