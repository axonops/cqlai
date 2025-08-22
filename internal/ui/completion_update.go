package ui

// getUpdateCompletions returns completions for UPDATE commands
func (ce *CompletionEngine) getUpdateCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return ce.getTableNames()
	}

	// Track what we've seen
	hasSet := false
	hasWhere := false
	hasIf := false
	hasUsing := false
	tablePos := 0
	setPos := -1
	wherePos := -1

	for i, word := range words {
		switch word {
		case "UPDATE":
			if i+1 < len(words) {
				tablePos = i + 1
			}
		case "SET":
			hasSet = true
			setPos = i
		case "WHERE":
			hasWhere = true
			wherePos = i
		case "IF":
			hasIf = true
		case "USING":
			hasUsing = true
		}
	}

	// Get the last word
	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	// Special keyword handling
	switch lastWord {
	case "UPDATE":
		return ce.getTableAndKeyspaceNames()
	case "SET":
		// After SET, suggest column names
		if tablePos > 0 && tablePos < len(words) {
			return ce.getColumnNamesForTable(words[tablePos])
		}
		return []string{}
	case "WHERE":
		// After WHERE, suggest column names (primary key)
		if tablePos > 0 && tablePos < len(words) {
			return ce.getColumnNamesForTable(words[tablePos])
		}
		return []string{}
	case "IF":
		// After IF, could be condition or EXISTS
		return []string{"EXISTS"}
	case "USING":
		return []string{"TTL", "TIMESTAMP"}
	case "TTL":
		// After TTL, expect a number
		return []string{}
	case "TIMESTAMP":
		// After TIMESTAMP, expect a number
		return []string{}
	case "=":
		// After equals in SET or WHERE clause
		return []string{}
	case "AND":
		// Could be in SET, WHERE, or USING clause
		if hasUsing && !hasWhere {
			return []string{"TTL", "TIMESTAMP"}
		} else if hasWhere && wherePos >= 0 {
			// In WHERE clause
			if tablePos > 0 && tablePos < len(words) {
				return ce.getColumnNamesForTable(words[tablePos])
			}
		} else if hasSet && setPos >= 0 && !hasWhere {
			// In SET clause
			if tablePos > 0 && tablePos < len(words) {
				return ce.getColumnNamesForTable(words[tablePos])
			}
		}
	}

	// After table name
	if tablePos > 0 && wordPos == tablePos+1 {
		var suggestions []string

		if !hasUsing {
			suggestions = append(suggestions, "USING")
		}
		if !hasSet {
			suggestions = append(suggestions, "SET")
		}

		return suggestions
	}

	// After SET clause
	if hasSet && setPos >= 0 && wordPos > setPos && !hasWhere {
		// Check if we're after a column name
		if ce.isColumnNameForTable(lastWord, words[tablePos]) {
			return []string{"=", "[", "."}
		}

		// Otherwise suggest WHERE or IF
		var suggestions []string
		suggestions = append(suggestions, "WHERE")
		if !hasIf {
			suggestions = append(suggestions, "IF")
		}
		return suggestions
	}

	// After WHERE clause
	if hasWhere && wherePos >= 0 && wordPos > wherePos {
		// Check if we're after a column name
		if ce.isColumnNameForTable(lastWord, words[tablePos]) {
			// UPDATE WHERE clause doesn't support CONTAINS operators
			return []string{"=", "!=", "<", ">", "<=", ">=", "IN"}
		}

		// After comparison operator
		if ce.isComparisonOperator(lastWord) {
			return []string{}
		}

		// Otherwise suggest IF
		if !hasIf {
			return []string{"IF", "AND"}
		}

		return []string{"AND"}
	}

	return []string{}
}

func (pce *ParserBasedCompletionEngine) getUpdateSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		// After UPDATE, suggest table names or SET
		return append(pce.getTableNames(), "SET")
	}

	hasSet := false
	hasWhere := false
	for _, t := range tokens {
		if t == "SET" {
			hasSet = true
		}
		if t == "WHERE" {
			hasWhere = true
		}
	}

	if !hasSet {
		return []string{"SET", "USING"}
	}

	if hasSet && !hasWhere {
		return []string{"WHERE", "IF"}
	}

	return []string{"IF", "AND"}
}
