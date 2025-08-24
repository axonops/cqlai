package ui

import "strings"

// IsCompleteKeyword checks if a word is a complete CQL keyword
func IsCompleteKeyword(word string) bool {
	// List of all CQL keywords that should trigger auto-space
	keywords := []string{
		// Top-level commands
		"SELECT", "INSERT", "UPDATE", "DELETE",
		"CREATE", "DROP", "ALTER", "TRUNCATE",
		"GRANT", "REVOKE", "USE",
		"DESCRIBE", "DESC", "BEGIN", "APPLY", "LIST", "OUTPUT",
		// Common clauses and keywords
		"FROM", "WHERE", "AND", "OR", "NOT",
		"SET", "VALUES", "INTO", "IF", "EXISTS",
		"TABLE", "KEYSPACE", "INDEX", "TYPE", "ROLE", "USER",
		"FUNCTION", "AGGREGATE", "MATERIALIZED", "VIEW", "TRIGGER",
		"PRIMARY", "KEY", "WITH", "USING", "TTL", "TIMESTAMP",
		"ORDER", "BY", "GROUP", "LIMIT", "ALLOW", "FILTERING",
		"ASC", "DESC", "DISTINCT", "COUNT", "TOKEN",
		"BATCH", "UNLOGGED", "COUNTER", "JSON",
		"CONSISTENCY", "SERIAL", "QUORUM", "ALL", "ONE", "TWO", "THREE",
		"LOCAL_ONE", "LOCAL_QUORUM", "EACH_QUORUM", "LOCAL_SERIAL",
		"ANY", "PER", "PARTITION", "ASCII",
	}

	upperWord := strings.ToUpper(word)
	for _, k := range keywords {
		if upperWord == k {
			return true
		}
	}
	return false
}

// isCompleteKeyword is a method wrapper for backward compatibility
func (m MainModel) isCompleteKeyword(word string) bool {
	return IsCompleteKeyword(word)
}
