package ui

// getDescribeCompletions returns completions for DESCRIBE commands
func (ce *CompletionEngine) getDescribeCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		// First word after DESCRIBE/DESC
		return DescribeObjects
	}

	// Get the last word to determine context
	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	// Handle MATERIALIZED VIEW
	if lastWord == "MATERIALIZED" {
		return MaterializedKeyword
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
		return append(OnKeyword, ",")
	}

	// After ON
	if hasOn && onPos >= 0 {
		if wordPos == onPos+1 {
			// Resource types
			return ResourceTypes
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
				return MaterializedKeyword
			case "ALL":
				return AllResourceTargets
			}
		}

		// After MATERIALIZED VIEW
		if wordPos == onPos+3 && words[onPos+1] == "MATERIALIZED" && words[onPos+2] == "VIEW" {
			return ce.getViewNames()
		}

		// After resource name, suggest TO
		if !hasTo && wordPos > onPos+2 {
			return ToKeyword
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
		return append(OnKeyword, ",")
	}

	// After ON
	if hasOn && onPos >= 0 {
		if wordPos == onPos+1 {
			// Resource types
			return ResourceTypes
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
				return MaterializedKeyword
			case "ALL":
				return AllResourceTargets
			}
		}

		// After MATERIALIZED VIEW
		if wordPos == onPos+3 && words[onPos+1] == "MATERIALIZED" && words[onPos+2] == "VIEW" {
			return ce.getViewNames()
		}

		// After resource name, suggest FROM
		if !hasFrom && wordPos > onPos+2 {
			return FromKeyword
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
func (ce *CompletionEngine) getUseCompletions(_ []string, wordPos int) []string {
	if wordPos == 1 {
		return ce.getKeyspaceNames()
	}
	return []string{}
}

// getShowCompletions returns completions for SHOW commands
func (ce *CompletionEngine) getShowCompletions(_ []string, wordPos int) []string {
	if wordPos == 1 {
		return ShowCommands
	}
	return []string{}
}

func (pce *ParserBasedCompletionEngine) getConsistencySuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return ConsistencyLevels
	}
	return []string{}
}

func (pce *ParserBasedCompletionEngine) getOutputSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		// After OUTPUT, suggest format types
		return OutputFormats
	}
	return []string{}
}
