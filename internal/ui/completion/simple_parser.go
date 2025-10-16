package completion

import (
	"strings"
)

// SimpleCompletionEngine provides completions without using ANTLR
type SimpleCompletionEngine struct {
	*CompletionEngine
}

// NewSimpleCompletionEngine creates a completion engine without ANTLR
func NewSimpleCompletionEngine(ce *CompletionEngine) *SimpleCompletionEngine {
	return &SimpleCompletionEngine{
		CompletionEngine: ce,
	}
}

// GetTokenCompletions provides context-aware completions using simple string parsing
func (sce *SimpleCompletionEngine) GetTokenCompletions(input string) []string {
	// Check for trailing space before trimming
	endsWithSpace := strings.HasSuffix(input, " ")

	// Trim the input for analysis
	trimmed := strings.TrimSpace(input)
	upper := strings.ToUpper(trimmed)

	// If input is empty or only whitespace, return top-level commands
	if trimmed == "" {
		return sce.getTopLevelKeywords()
	}

	// Split into words
	words := strings.Fields(upper)
	if len(words) == 0 {
		return sce.getTopLevelKeywords()
	}

	// Get the last word for partial matching
	lastWord := words[len(words)-1]

	// If we're typing the first word and no space, match top-level commands
	if len(words) == 1 && !endsWithSpace {
		var suggestions []string
		lowerLastWord := strings.ToLower(lastWord)

		// Check all top-level keywords
		for _, keyword := range sce.getTopLevelKeywords() {
			lowerKeyword := strings.ToLower(keyword)
			if strings.HasPrefix(lowerKeyword, lowerLastWord) && lowerKeyword != lowerLastWord {
				suggestions = append(suggestions, keyword)
			}
		}

		// Special handling for DESC -> DESCRIBE
		if lowerLastWord == "desc" {
			// Only suggest DESCRIBE, not DESC (since DESC is already typed)
			suggestions = []string{"DESCRIBE"}
		}

		return suggestions
	}

	// Context-aware completions based on first word
	firstWord := words[0]

	switch firstWord {
	case "DESCRIBE", "DESC":
		return sce.getDescribeCompletions(words, endsWithSpace)
	case "LIST":
		return sce.getListCompletions(words, endsWithSpace)
	case "SHOW":
		return sce.getShowCompletions(words, endsWithSpace)
	case "CONSISTENCY":
		return sce.getConsistencyCompletions(words, endsWithSpace)
	case "OUTPUT":
		return sce.getOutputCompletions(words, endsWithSpace)
	case "CREATE":
		return sce.getCreateCompletions(words, endsWithSpace)
	case "ALTER":
		return sce.getAlterCompletions(words, endsWithSpace)
	case "DROP":
		return sce.getDropCompletions(words, endsWithSpace)
	case "GRANT":
		return sce.getGrantCompletions(words, endsWithSpace)
	case "REVOKE":
		return sce.getRevokeCompletions(words, endsWithSpace)
	case "SELECT":
		return sce.getSelectCompletions(words, endsWithSpace)
	case "INSERT":
		return sce.getInsertCompletions(words, endsWithSpace)
	case "UPDATE":
		return sce.getUpdateCompletions(words, endsWithSpace)
	case "DELETE":
		return sce.getDeleteCompletions(words, endsWithSpace)
	case "USE":
		if endsWithSpace || len(words) > 1 {
			return sce.getKeyspaceNames()
		}
		return nil
	}

	// Default: no completions
	return nil
}

// getDescribeCompletions provides completions for DESCRIBE commands
func (sce *SimpleCompletionEngine) getDescribeCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		// After DESCRIBE, suggest object types
		return []string{
			"KEYSPACE", "KEYSPACES",
			"TABLE", "TABLES",
			"TYPE", "TYPES",
			"FUNCTION", "FUNCTIONS",
			"AGGREGATE", "AGGREGATES",
			"MATERIALIZED",
			"INDEX",
			"CLUSTER",
			"SCHEMA",
		}
	}

	if len(words) == 2 && !endsWithSpace {
		// Partial match on second word
		suggestions := []string{}
		second := strings.ToLower(words[1])

		objects := []string{
			"KEYSPACE", "KEYSPACES",
			"TABLE", "TABLES",
			"TYPE", "TYPES",
			"FUNCTION", "FUNCTIONS",
			"AGGREGATE", "AGGREGATES",
			"MATERIALIZED",
			"INDEX",
			"CLUSTER",
			"SCHEMA",
		}

		for _, obj := range objects {
			if strings.HasPrefix(strings.ToLower(obj), second) && strings.ToLower(obj) != second {
				suggestions = append(suggestions, obj)
			}
		}

		// Also suggest keyspace/table names for direct describe
		if second != "" {
			for _, ks := range sce.getKeyspaceNames() {
				if strings.HasPrefix(strings.ToLower(ks), second) {
					suggestions = append(suggestions, ks)
				}
			}
			for _, tbl := range sce.getTableNames() {
				if strings.HasPrefix(strings.ToLower(tbl), second) {
					suggestions = append(suggestions, tbl)
				}
			}
		}

		return suggestions
	}

	if len(words) == 2 && endsWithSpace {
		switch strings.ToUpper(words[1]) {
		case "KEYSPACE":
			return sce.getKeyspaceNames()
		case "TABLE":
			return sce.getTableAndKeyspaceNames()
		case "TYPE":
			return sce.getTypeNames()
		case "FUNCTION":
			return sce.getFunctionNames()
		case "AGGREGATE":
			return sce.getAggregateNames()
		case "INDEX":
			return sce.getIndexNames()
		case "MATERIALIZED":
			return []string{"VIEW", "VIEWS"}
		}
	}

	if len(words) == 3 && strings.ToUpper(words[1]) == "MATERIALIZED" {
		if !endsWithSpace && strings.HasPrefix(strings.ToUpper(words[2]), "VIEW") {
			return []string{"VIEW"}
		}
		if strings.ToUpper(words[2]) == "VIEW" && endsWithSpace {
			// TODO: implement getMaterializedViewNames
			return nil
		}
	}

	return nil
}

// Helper methods for other completion contexts
func (sce *SimpleCompletionEngine) getListCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return []string{"ROLES", "USERS", "PERMISSIONS"}
	}
	if len(words) == 2 && !endsWithSpace {
		suggestions := []string{}
		second := strings.ToLower(words[1])
		for _, obj := range []string{"ROLES", "USERS", "PERMISSIONS"} {
			if strings.HasPrefix(strings.ToLower(obj), second) && strings.ToLower(obj) != second {
				suggestions = append(suggestions, obj)
			}
		}
		return suggestions
	}
	return nil
}

func (sce *SimpleCompletionEngine) getShowCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return []string{"VERSION", "HOST", "SESSION"}
	}
	if len(words) == 2 && !endsWithSpace {
		suggestions := []string{}
		second := strings.ToLower(words[1])
		for _, obj := range []string{"VERSION", "HOST", "SESSION"} {
			if strings.HasPrefix(strings.ToLower(obj), second) && strings.ToLower(obj) != second {
				suggestions = append(suggestions, obj)
			}
		}
		return suggestions
	}
	return nil
}

func (sce *SimpleCompletionEngine) getConsistencyCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return sce.getConsistencyLevels()
	}
	if len(words) == 2 && !endsWithSpace {
		suggestions := []string{}
		second := strings.ToLower(words[1])
		for _, level := range sce.getConsistencyLevels() {
			if strings.HasPrefix(strings.ToLower(level), second) && strings.ToLower(level) != second {
				suggestions = append(suggestions, level)
			}
		}
		return suggestions
	}
	return nil
}

func (sce *SimpleCompletionEngine) getOutputCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return []string{"TABLE", "ASCII", "EXPAND", "JSON"}
	}
	if len(words) == 2 && !endsWithSpace {
		suggestions := []string{}
		second := strings.ToLower(words[1])
		for _, format := range []string{"TABLE", "ASCII", "EXPAND", "JSON"} {
			if strings.HasPrefix(strings.ToLower(format), second) && strings.ToLower(format) != second {
				suggestions = append(suggestions, format)
			}
		}
		return suggestions
	}
	return nil
}

func (sce *SimpleCompletionEngine) getCreateCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return []string{"KEYSPACE", "TABLE", "INDEX", "TYPE", "TRIGGER", "FUNCTION", "AGGREGATE", "MATERIALIZED", "USER", "ROLE"}
	}
	if len(words) == 2 && !endsWithSpace {
		suggestions := []string{}
		second := strings.ToLower(words[1])
		for _, obj := range []string{"KEYSPACE", "TABLE", "INDEX", "TYPE", "TRIGGER", "FUNCTION", "AGGREGATE", "MATERIALIZED", "USER", "ROLE"} {
			if strings.HasPrefix(strings.ToLower(obj), second) && strings.ToLower(obj) != second {
				suggestions = append(suggestions, obj)
			}
		}
		return suggestions
	}
	// After CREATE TABLE/INDEX, suggest keyspace names for qualified names
	if len(words) == 2 && endsWithSpace && (words[1] == "TABLE" || words[1] == "INDEX") {
		keyspaces := sce.getKeyspaceNames()
		suggestions := []string{"IF"} // Add IF for IF NOT EXISTS
		for _, ks := range keyspaces {
			suggestions = append(suggestions, ks+".")
		}
		return suggestions
	}
	return nil
}

func (sce *SimpleCompletionEngine) getAlterCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return []string{"KEYSPACE", "TABLE", "TYPE", "USER", "ROLE"}
	}
	if len(words) == 2 && !endsWithSpace {
		suggestions := []string{}
		second := strings.ToLower(words[1])
		for _, obj := range []string{"KEYSPACE", "TABLE", "TYPE", "USER", "ROLE"} {
			if strings.HasPrefix(strings.ToLower(obj), second) && strings.ToLower(obj) != second {
				suggestions = append(suggestions, obj)
			}
		}
		return suggestions
	}
	return nil
}

func (sce *SimpleCompletionEngine) getDropCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return []string{"KEYSPACE", "TABLE", "INDEX", "TYPE", "TRIGGER", "FUNCTION", "AGGREGATE", "MATERIALIZED", "USER", "ROLE"}
	}
	if len(words) == 2 && !endsWithSpace {
		suggestions := []string{}
		second := strings.ToLower(words[1])
		for _, obj := range []string{"KEYSPACE", "TABLE", "INDEX", "TYPE", "TRIGGER", "FUNCTION", "AGGREGATE", "MATERIALIZED", "USER", "ROLE"} {
			if strings.HasPrefix(strings.ToLower(obj), second) && strings.ToLower(obj) != second {
				suggestions = append(suggestions, obj)
			}
		}
		return suggestions
	}
	return nil
}

func (sce *SimpleCompletionEngine) getGrantCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return []string{"SELECT", "MODIFY", "CREATE", "ALTER", "DROP", "EXECUTE", "AUTHORIZE"}
	}
	return nil
}

func (sce *SimpleCompletionEngine) getRevokeCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return []string{"SELECT", "MODIFY", "CREATE", "ALTER", "DROP", "EXECUTE", "AUTHORIZE"}
	}
	return nil
}

func (sce *SimpleCompletionEngine) getSelectCompletions(words []string, endsWithSpace bool) []string {
	// For SELECT, we can suggest FROM after fields
	for i, word := range words {
		if word == "FROM" && i < len(words)-1 {
			if (i == len(words)-2 && endsWithSpace) || (i == len(words)-1 && !endsWithSpace) {
				return sce.getTableAndKeyspaceNames()
			}
		}
	}

	// Look for SELECT without FROM yet
	hasFrom := false
	for _, word := range words {
		if word == "FROM" {
			hasFrom = true
			break
		}
	}

	if !hasFrom && endsWithSpace && len(words) > 1 {
		return []string{"FROM"}
	}

	return nil
}

func (sce *SimpleCompletionEngine) getInsertCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return []string{"INTO"}
	}
	if len(words) == 2 && strings.ToUpper(words[1]) == "INTO" && endsWithSpace {
		return sce.getTableAndKeyspaceNames()
	}
	return nil
}

func (sce *SimpleCompletionEngine) getUpdateCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return sce.getTableAndKeyspaceNames()
	}
	return nil
}

func (sce *SimpleCompletionEngine) getDeleteCompletions(words []string, endsWithSpace bool) []string {
	if len(words) == 1 && endsWithSpace {
		return []string{"FROM"}
	}
	if len(words) == 2 && strings.ToUpper(words[1]) == "FROM" && endsWithSpace {
		return sce.getTableAndKeyspaceNames()
	}
	return nil
}

// getTopLevelKeywords returns all top-level CQL keywords
func (sce *SimpleCompletionEngine) getTopLevelKeywords() []string {
	return []string{
		"ALTER", "APPLY", "ASCII", "ASSUME", "BEGIN", "CAPTURE", "CONSISTENCY",
		"COPY", "CREATE", "DELETE", "DESC", "DESCRIBE", "DROP", "EXECUTE", "EXIT",
		"EXPAND", "EXPLAIN", "GRANT", "HELP", "INSERT", "LIST", "OUTPUT", "PAGING",
		"QUIT", "REVOKE", "SELECT", "SHOW", "SOURCE", "TRACING", "TRUNCATE",
		"UPDATE", "USE",
	}
}

// getConsistencyLevels returns valid consistency levels
func (sce *SimpleCompletionEngine) getConsistencyLevels() []string {
	return []string{
		"ALL", "EACH_QUORUM", "QUORUM", "LOCAL_QUORUM", "ONE", "TWO", "THREE",
		"LOCAL_ONE", "ANY", "SERIAL", "LOCAL_SERIAL",
	}
}