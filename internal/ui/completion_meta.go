package ui

// getDescribeCompletions returns completions for DESCRIBE commands
func (ce *CompletionEngine) getDescribeCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		// First word after DESCRIBE/DESC
		return []string{
			"KEYSPACE", "KEYSPACES",
			"TABLE", "TABLES",
			"TYPE", "TYPES",
			"FUNCTION", "FUNCTIONS",
			"AGGREGATE", "AGGREGATES",
			"MATERIALIZED",
			"INDEX",
			"SCHEMA",
			"CLUSTER",
		}
	}

	// Get the last word to determine context
	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	// Handle MATERIALIZED VIEW
	if lastWord == "MATERIALIZED" {
		return []string{"VIEW"}
	}

	// If we're at position 2 (after DESCRIBE <type>), suggest names
	// This handles both "DESCRIBE KEYSPACE " and "DESCRIBE KEYSPACE p"
	if wordPos == 2 {
		// Check what type of object we're describing
		if len(words) > 1 {
			switch words[1] {
			case "KEYSPACE":
				return ce.getKeyspaceNames()
			case "TABLE":
				// For DESCRIBE TABLE, return both local tables and keyspace.table combinations
				return ce.getTableAndKeyspaceTableNames()
			case "TYPE":
				return ce.getTypeNames()
			case "FUNCTION":
				return ce.getFunctionNames()
			case "AGGREGATE":
				return ce.getAggregateNames()
			case "INDEX":
				return ce.getIndexNames()
			}
		}
	}

	// After DESCRIBE MATERIALIZED VIEW
	if wordPos == 3 && len(words) > 2 && words[1] == "MATERIALIZED" && words[2] == "VIEW" {
		// After DESCRIBE MATERIALIZED VIEW
		return ce.getViewNames()
	}

	return []string{}
}

// getGrantCompletions returns completions for GRANT commands
func (ce *CompletionEngine) getGrantCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		// After GRANT, suggest permissions
		return CQLPermissions
	}

	// Track what we've seen
	hasOn := false
	hasTo := false
	onPos := -1

	for i, word := range words {
		if word == "ON" {
			hasOn = true
			onPos = i
		} else if word == "TO" {
			hasTo = true
		}
	}

	// Get the last word
	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	// After permission(s) and before ON
	if !hasOn && wordPos > 1 {
		// Could have multiple permissions separated by commas
		if lastWord == "," {
			return CQLPermissions
		}
		return []string{"ON", ","}
	}

	// After ON
	if hasOn && onPos >= 0 {
		if wordPos == onPos+1 {
			// Resource types
			return []string{"ALL", "KEYSPACE", "TABLE", "ROLE", "FUNCTION", "AGGREGATE", "INDEX", "MATERIALIZED"}
		}

		// After resource type
		if wordPos == onPos+2 {
			switch words[onPos+1] {
			case "KEYSPACE":
				return ce.getKeyspaceNames()
			case "TABLE":
				return ce.getTableNames()
			case "FUNCTION":
				return ce.getFunctionNames()
			case "AGGREGATE":
				return ce.getAggregateNames()
			case "INDEX":
				return ce.getIndexNames()
			case "MATERIALIZED":
				return []string{"VIEW"}
			case "ALL":
				return []string{"KEYSPACES", "FUNCTIONS", "ROLES"}
			}
		}

		// After MATERIALIZED VIEW
		if wordPos == onPos+3 && words[onPos+1] == "MATERIALIZED" && words[onPos+2] == "VIEW" {
			return ce.getViewNames()
		}

		// After resource name, suggest TO
		if !hasTo && wordPos > onPos+2 {
			return []string{"TO"}
		}
	}

	// After TO
	if hasTo {
		// Suggest role/user name (no completion for now)
		return []string{}
	}

	return []string{}
}

// getRevokeCompletions returns completions for REVOKE commands
func (ce *CompletionEngine) getRevokeCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		// After REVOKE, suggest permissions
		return CQLPermissions
	}

	// Track what we've seen
	hasOn := false
	hasFrom := false
	onPos := -1

	for i, word := range words {
		if word == "ON" {
			hasOn = true
			onPos = i
		} else if word == "FROM" {
			hasFrom = true
		}
	}

	// Get the last word
	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	// After permission(s) and before ON
	if !hasOn && wordPos > 1 {
		// Could have multiple permissions separated by commas
		if lastWord == "," {
			return CQLPermissions
		}
		return []string{"ON", ","}
	}

	// After ON
	if hasOn && onPos >= 0 {
		if wordPos == onPos+1 {
			// Resource types
			return []string{"ALL", "KEYSPACE", "TABLE", "ROLE", "FUNCTION", "AGGREGATE", "INDEX", "MATERIALIZED"}
		}

		// After resource type
		if wordPos == onPos+2 {
			switch words[onPos+1] {
			case "KEYSPACE":
				return ce.getKeyspaceNames()
			case "TABLE":
				return ce.getTableNames()
			case "FUNCTION":
				return ce.getFunctionNames()
			case "AGGREGATE":
				return ce.getAggregateNames()
			case "INDEX":
				return ce.getIndexNames()
			case "MATERIALIZED":
				return []string{"VIEW"}
			case "ALL":
				return []string{"KEYSPACES", "FUNCTIONS", "ROLES"}
			}
		}

		// After MATERIALIZED VIEW
		if wordPos == onPos+3 && words[onPos+1] == "MATERIALIZED" && words[onPos+2] == "VIEW" {
			return ce.getViewNames()
		}

		// After resource name, suggest FROM
		if !hasFrom && wordPos > onPos+2 {
			return []string{"FROM"}
		}
	}

	// After FROM
	if hasFrom {
		// Suggest role/user name (no completion for now)
		return []string{}
	}

	return []string{}
}

// getUseCompletions returns completions for USE commands
func (ce *CompletionEngine) getUseCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return ce.getKeyspaceNames()
	}
	return []string{}
}

// getShowCompletions returns completions for SHOW commands
func (ce *CompletionEngine) getShowCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return []string{"VERSION", "HOST", "SESSION"}
	}
	return []string{}
}

func (pce *ParserBasedCompletionEngine) getConsistencySuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return []string{
			"ALL",
			"ANY",
			"EACH_QUORUM",
			"LOCAL_ONE",
			"LOCAL_QUORUM",
			"LOCAL_SERIAL",
			"ONE",
			"QUORUM",
			"SERIAL",
			"THREE",
			"TWO",
		}
	}
	return []string{}
}
