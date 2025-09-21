package completion

func (pce *ParserBasedCompletionEngine) getDescribeSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return DescribeObjectsBasic
	}

	if len(tokens) == 2 && tokens[1] == "MATERIALIZED" {
		return MaterializedViews
	}

	return []string{}
}

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
