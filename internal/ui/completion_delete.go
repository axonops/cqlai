package ui

// getDeleteCompletions returns completions for DELETE commands
func (ce *CompletionEngine) getDeleteCompletions(words []string, wordPos int) []string {
	// Track what we've seen
	hasFrom := false
	hasWhere := false
	hasIf := false
	hasUsing := false
	tablePos := -1

	for i, word := range words {
		switch word {
		case "FROM":
			hasFrom = true
			if i+1 < len(words) {
				tablePos = i + 1
			}
		case "WHERE":
			hasWhere = true
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
	case "DELETE":
		// After DELETE, can specify columns or FROM
		return FromKeyword
	case "FROM":
		return ce.getTableAndKeyspaceNames()
	case "WHERE":
		// After WHERE, suggest column names
		if tablePos >= 0 && tablePos < len(words) {
			return ce.getColumnNamesForTable(words[tablePos])
		}
		return []string{}
	case "IF":
		return ExistsKeyword
	case "USING":
		return TimestampKeyword
	case "TIMESTAMP":
		// After TIMESTAMP, expect a number
		return []string{}
	case "AND":
		// In WHERE clause
		if hasWhere && tablePos >= 0 && tablePos < len(words) {
			return ce.getColumnNamesForTable(words[tablePos])
		}
	}

	// After DELETE
	if wordPos == 1 && !hasFrom {
		// Can specify column names or FROM
		var suggestions []string
		suggestions = append(suggestions, "FROM")
		// Could also add column names for DELETE column FROM syntax
		if len(words) > 1 {
			suggestions = append(suggestions, ce.getColumnNamesForCurrentContext(words)...)
		}
		return suggestions
	}

	// After table name
	if hasFrom && tablePos >= 0 && wordPos == tablePos+1 {
		var suggestions []string

		if !hasUsing {
			suggestions = append(suggestions, "USING")
		}
		if !hasWhere {
			suggestions = append(suggestions, "WHERE")
		}

		return suggestions
	}

	// After WHERE clause
	if hasWhere && !hasIf {
		// Check if we're after a column name
		if tablePos >= 0 && ce.isColumnNameForTable(lastWord, words[tablePos]) {
			// DELETE WHERE clause doesn't support CONTAINS operators
			return []string{"=", "!=", "<", ">", "<=", ">=", "IN"}
		}

		// After comparison operator
		if ce.isComparisonOperator(lastWord) {
			return []string{}
		}

		return IfAnd
	}

	return []string{}
}

func (pce *ParserBasedCompletionEngine) getDeleteSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return FromKeyword
	}

	hasFrom := false
	hasWhere := false
	for _, t := range tokens {
		if t == "FROM" {
			hasFrom = true
		}
		if t == "WHERE" {
			hasWhere = true
		}
	}

	if hasFrom && !hasWhere {
		return WhereUsingIf
	}

	return []string{}
}
