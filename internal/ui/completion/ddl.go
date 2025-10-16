package completion

import "strings"

// getCreateCompletions returns completions for CREATE commands
func (ce *CompletionEngine) getCreateCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return CreateDropObjectTypes
	}

	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	if lastWord == "MATERIALIZED" {
		return MaterializedKeyword
	}

	if wordPos == 2 && len(words) > 1 {
		// Check if it's a valid object type (excluding MATERIALIZED which needs VIEW after it)
		for _, objType := range CreateDropObjectTypesNoMaterialized {
			if words[1] == objType {
				// After CREATE <type>, suggest IF keyword and optionally keyspace names for qualified names
				// For TABLE, INDEX, allow keyspace.object syntax
				if words[1] == "TABLE" || words[1] == "INDEX" {
					// Suggest both IF keyword and keyspace names with dots
					keyspaces := ce.getKeyspaceNames()
					suggestions := append([]string{}, IfKeyword...)
					for _, ks := range keyspaces {
						suggestions = append(suggestions, ks+".")
					}
					return suggestions
				}
				// For all other types (KEYSPACE, TYPE, ROLE, etc.), suggest IF
				return IfKeyword
			}
		}
	}

	// After "CREATE <type> IF", suggest "NOT"
	if wordPos == 3 && len(words) > 2 && words[2] == "IF" {
		return []string{"NOT"}
	}

	// After "CREATE <type> IF NOT", suggest "EXISTS"
	if wordPos == 4 && len(words) > 3 && words[2] == "IF" && words[3] == "NOT" {
		return []string{"EXISTS"}
	}

	// After "CREATE TABLE IF NOT EXISTS", suggest keyspace names
	if wordPos == 5 && len(words) > 4 && words[2] == "IF" && words[3] == "NOT" && words[4] == "EXISTS" {
		if words[1] == "TABLE" || words[1] == "INDEX" {
			keyspaces := ce.getKeyspaceNames()
			suggestions := make([]string, 0, len(keyspaces))
			for _, ks := range keyspaces {
				suggestions = append(suggestions, ks+".")
			}
			return suggestions
		}
	}

	// After "CREATE KEYSPACE [IF NOT EXISTS] <name>", suggest WITH
	if wordPos >= 3 && len(words) > 2 && words[1] == "KEYSPACE" {
		// Could be: CREATE KEYSPACE name (wordPos=3)
		// Or: CREATE KEYSPACE IF NOT EXISTS name (wordPos=6)
		if (wordPos == 3 && words[2] != "IF") || (wordPos == 6 && words[2] == "IF" && words[3] == "NOT" && words[4] == "EXISTS") {
			return WithKeyword
		}
	}

	// After "CREATE KEYSPACE ... WITH", suggest REPLICATION
	if wordPos >= 4 && len(words) > 3 && words[1] == "KEYSPACE" {
		for i := 2; i < len(words); i++ {
			if words[i] == "WITH" && wordPos == i+1 {
				return []string{"REPLICATION"}
			}
		}
	}

	// After "CREATE KEYSPACE ... REPLICATION", suggest "="
	if wordPos >= 5 && len(words) > 4 && words[1] == "KEYSPACE" {
		for i := 2; i < len(words); i++ {
			if words[i] == "REPLICATION" && wordPos == i+1 {
				return []string{"="}
			}
		}
	}

	// After "CREATE KEYSPACE ... REPLICATION =", suggest replication map syntax
	if wordPos >= 6 && len(words) > 5 && words[1] == "KEYSPACE" {
		for i := 2; i < len(words)-1; i++ {
			if words[i] == "REPLICATION" && words[i+1] == "=" && wordPos == i+2 {
				return []string{
					"{'class': 'SimpleStrategy', 'replication_factor': 1}",
					"{'class': 'NetworkTopologyStrategy', 'datacenter1': 3}",
				}
			}
		}
	}

	// Handle CREATE TABLE column type suggestions
	// Detect if we're inside column definitions (after opening parenthesis)
	if len(words) > 1 && words[1] == "TABLE" {
		// Look for opening parenthesis in any word
		openParenIdx := -1
		for i, word := range words {
			if strings.Contains(word, "(") {
				openParenIdx = i
				break
			}
		}

		// If we found an opening paren and we're past it
		if openParenIdx != -1 && wordPos > openParenIdx {
			// Check if we're expecting a data type
			// Simple heuristic: after column name, before comma or PRIMARY KEY
			// Count words after the opening paren
			wordsAfterParen := wordPos - openParenIdx

			// If we're at an odd position after the paren, we're likely at a type position
			// Position pattern: (col1 type1, col2 type2, col3 type3, PRIMARY KEY...)
			// After paren: 1=col, 2=type, 3=comma, 4=col, 5=type, etc.
			if wordsAfterParen%2 == 0 {
				// Even position = type position
				// But exclude if the current or previous word is a keyword
				currentWord := ""
				if wordPos < len(words) {
					currentWord = words[wordPos]
				}
				prevWord := ""
				if wordPos > 0 && wordPos-1 < len(words) {
					prevWord = words[wordPos-1]
				}

				// Don't suggest types if we're at PRIMARY, KEY, etc.
				excludedKeywords := []string{"PRIMARY", "KEY", "WITH", "CLUSTERING", "COMPACT"}
				isExcluded := false
				for _, kw := range excludedKeywords {
					if currentWord == kw || prevWord == kw {
						isExcluded = true
						break
					}
				}

				if !isExcluded {
					return CQLDataTypes
				}
			}
		}
	}

	return []string{}
}

// getDropCompletions returns completions for DROP commands
func (ce *CompletionEngine) getDropCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return CreateDropObjectTypes
	}

	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	if lastWord == "MATERIALIZED" {
		return MaterializedKeyword
	}

	// After DROP <type>
	if wordPos == 2 && len(words) > 1 {
		switch words[1] {
		case "KEYSPACE":
			return append(IfKeyword, ce.getKeyspaceNames()...) // "IF"
		case "TABLE":
			return append(IfKeyword, ce.getTableAndKeyspaceNames()...) // "IF"
		case "INDEX":
			return append(IfKeyword, ce.getIndexNames()...) // "IF"
		case "TYPE":
			return append(IfKeyword, ce.getTypeNames()...) // "IF"
		case "FUNCTION":
			return append(IfKeyword, ce.getFunctionNames()...) // "IF"
		case "AGGREGATE":
			return append(IfKeyword, ce.getAggregateNames()...) // "IF"
		case "ROLE", "USER":
			return IfKeyword
		}
	}

	// After DROP MATERIALIZED VIEW
	if wordPos == 3 && len(words) > 2 && words[1] == "MATERIALIZED" && words[2] == "VIEW" {
		return append(IfKeyword, ce.getViewNames()...)
	}

	// After "DROP <type> IF", suggest "EXISTS"
	if wordPos == 3 && len(words) > 2 && words[2] == "IF" {
		return []string{"EXISTS"}
	}

	// After "DROP MATERIALIZED VIEW IF", suggest "EXISTS"
	if wordPos == 4 && len(words) > 3 && words[1] == "MATERIALIZED" && words[2] == "VIEW" && words[3] == "IF" {
		return []string{"EXISTS"}
	}

	// After "DROP TABLE IF EXISTS", suggest keyspace.table names
	if wordPos == 4 && len(words) > 3 && words[2] == "IF" && words[3] == "EXISTS" {
		switch words[1] {
		case "KEYSPACE":
			return ce.getKeyspaceNames()
		case "TABLE":
			return ce.getTableAndKeyspaceNames()
		case "INDEX":
			return ce.getIndexNames()
		case "TYPE":
			return ce.getTypeNames()
		case "FUNCTION":
			return ce.getFunctionNames()
		case "AGGREGATE":
			return ce.getAggregateNames()
		}
	}

	// After "DROP MATERIALIZED VIEW IF EXISTS", suggest view names
	if wordPos == 5 && len(words) > 4 && words[1] == "MATERIALIZED" && words[2] == "VIEW" && words[3] == "IF" && words[4] == "EXISTS" {
		return ce.getViewNames()
	}

	return []string{}
}

// getAlterCompletions returns completions for ALTER commands
func (ce *CompletionEngine) getAlterCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return AlterObjectTypes
	}

	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	if lastWord == "MATERIALIZED" {
		return MaterializedKeyword
	}

	// After ALTER <type>
	if wordPos == 2 && len(words) > 1 {
		switch words[1] {
		case "KEYSPACE":
			return ce.getKeyspaceNames()
		case "TABLE":
			return ce.getTableAndKeyspaceNames()
		case "TYPE":
			return ce.getTypeNames()
		case "ROLE", "USER":
			return []string{}
		}
	}

	// After ALTER MATERIALIZED VIEW
	if wordPos == 3 && len(words) > 2 && words[1] == "MATERIALIZED" && words[2] == "VIEW" {
		return ce.getViewNames()
	}

	// After the object name, suggest modification keywords
	if wordPos == 3 && len(words) > 2 {
		switch words[1] {
		case "TABLE":
			return AlterTableOperations
		case "KEYSPACE":
			return WithKeyword
		case "TYPE":
			return AlterTypeOperations
		}
	}

	return []string{}
}

// getTruncateCompletions returns completions for TRUNCATE commands
func (ce *CompletionEngine) getTruncateCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		// After TRUNCATE, can specify TABLE or just table name
		tables := ce.getTableAndKeyspaceNames()
		return append(TableKeyword, tables...)
	}

	if wordPos == 2 && len(words) > 1 && words[1] == "TABLE" {
		return ce.getTableAndKeyspaceNames()
	}

	return []string{}
}



