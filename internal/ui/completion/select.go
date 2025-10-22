package completion

import (
	"fmt"
	"os"
)

// getSelectCompletions returns completions for SELECT commands
func (ce *CompletionEngine) getSelectCompletions(words []string, wordPos int) []string {
	// Track what we've seen in the SELECT statement
	hasDistinct := false
	hasJson := false
	hasFrom := false
	hasWhere := false
	hasOrderBy := false
	hasLimit := false
	hasAllowFiltering := false
	hasGroupBy := false
	hasPerPartitionLimit := false

	fromIndex := -1
	whereIndex := -1
	orderByIndex := -1

	for i, word := range words {
		switch word {
		case "DISTINCT":
			hasDistinct = true
		case "JSON":
			hasJson = true
		case "FROM":
			hasFrom = true
			fromIndex = i
		case "WHERE":
			hasWhere = true
			whereIndex = i
		case "ORDER":
			if i+1 < len(words) && words[i+1] == "BY" {
				hasOrderBy = true
				orderByIndex = i
			}
		case "GROUP":
			if i+1 < len(words) && words[i+1] == "BY" {
				hasGroupBy = true
			}
		case "LIMIT":
			hasLimit = true
		case "PER":
			if i+1 < len(words) && words[i+1] == "PARTITION" && i+2 < len(words) && words[i+2] == "LIMIT" {
				hasPerPartitionLimit = true
			}
		case "ALLOW":
			if i+1 < len(words) && words[i+1] == "FILTERING" {
				hasAllowFiltering = true
			}
		}
	}

	// Get the last word to determine context
	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	// Handle special keywords that need immediate completion
	switch lastWord {
	case "SELECT":
		return append([]string{"*", "DISTINCT", "JSON"}, ce.getFunctionSuggestions()...)
	case "DISTINCT":
		if !hasJson {
			return append([]string{"*", "JSON"}, ce.getFunctionSuggestions()...)
		}
		return append([]string{"*"}, ce.getFunctionSuggestions()...)
	case "JSON":
		return append([]string{"*"}, ce.getFunctionSuggestions()...)
	case "AS":
		// After AS, expect an alias name (no completion)
		return []string{}
	case "FROM":
		return ce.getTableAndKeyspaceNames()
	case "WHERE":
		// After WHERE, suggest column names if we know the table
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] select.go: lastWord=WHERE, calling getColumnNamesForCurrentTable with fromIndex=%d\n", fromIndex)
			defer debugFile.Close()
		}
		columns := ce.getColumnNamesForCurrentTable(words, fromIndex)
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] select.go: getColumnNamesForCurrentTable returned %d columns\n", len(columns))
			defer debugFile.Close()
		}
		return columns
	case "ORDER":
		return ByKeyword
	case "BY":
		if len(words) > 1 {
			prevWord := words[len(words)-2]
			switch prevWord {
			case "ORDER":
				// After ORDER BY, suggest column names
				return ce.getColumnNamesForCurrentTable(words, fromIndex)
			case "GROUP":
				// After GROUP BY, suggest column names
				return ce.getColumnNamesForCurrentTable(words, fromIndex)
			case "PARTITION":
				// After PARTITION BY (in window functions)
				return ce.getColumnNamesForCurrentTable(words, fromIndex)
			}
		}
	case "GROUP":
		return ByKeyword
	case "LIMIT":
		// After LIMIT, expect a number (no completion)
		return []string{}
	case "ALLOW":
		return FilteringKeyword
	case "PER":
		return PartitionKeyword
	case "COUNT(":
		// After COUNT(, suggest *, 1, or column names
		suggestions := []string{"*", "1"}
		columns := ce.getColumnNamesForCurrentTable(words, fromIndex)
		suggestions = append(suggestions, columns...)
		return suggestions
	case "MAX(", "MIN(", "AVG(", "SUM(":
		// Aggregate functions need column names
		return ce.getColumnNamesForCurrentTable(words, fromIndex)
	case "TTL(", "WRITETIME(":
		// These functions need column names
		return ce.getColumnNamesForCurrentTable(words, fromIndex)
	case "TOKEN(":
		// TOKEN function needs partition key columns
		// For now, return all columns
		return ce.getColumnNamesForCurrentTable(words, fromIndex)
	case "CAST(":
		// CAST needs a value/column followed by AS type
		return ce.getColumnNamesForCurrentTable(words, fromIndex)
	case "PARTITION":
		if len(words) > 1 && words[len(words)-2] == "PER" {
			return LimitKeyword
		}
	case "ASC", "DESC":
		// After sort order, might have more columns or next clause
		if orderByIndex >= 0 {
			return ce.getSuggestionsAfterOrderBy(words, hasWhere, hasLimit, hasAllowFiltering, hasPerPartitionLimit)
		}
	}

	// Before FROM clause
	if !hasFrom {
		if wordPos == 1 {
			// First position after SELECT
			if !hasDistinct && !hasJson {
				return append([]string{"*", "DISTINCT", "JSON"}, ce.getFunctionSuggestions()...)
			}
		}

		// Check if we're after column specification
		if wordPos > 1 || (wordPos == 1 && (hasDistinct || hasJson)) {
			// Check if the previous word looks like a column or function
			if lastWord != "DISTINCT" && lastWord != "JSON" && lastWord != "SELECT" {
				// After * or column names, suggest FROM
				if lastWord == "*" || lastWord == "COUNT(*)" {
					return FromKeyword
				}
				return FromCommaAs
			}
		}
	}

	// After FROM clause but before WHERE
	if hasFrom && fromIndex >= 0 {
		// Immediately after FROM
		if wordPos == fromIndex+1 {
			return ce.getTableAndKeyspaceNames()
		}

		// After table name
		if wordPos > fromIndex+1 {
			// Don't suggest clauses if we're inside a WHERE clause
			if hasWhere && whereIndex >= 0 && wordPos > whereIndex {
				// We're inside WHERE clause, let later logic handle it
			} else {
				var suggestions []string

				if !hasWhere {
					suggestions = append(suggestions, "WHERE")
				}
				if !hasGroupBy && hasWhere {
					suggestions = append(suggestions, "GROUP")
				}
				if !hasOrderBy && (hasWhere || !hasWhere) {
					suggestions = append(suggestions, "ORDER")
				}
				if !hasLimit && (hasWhere || hasOrderBy || !hasWhere) {
					suggestions = append(suggestions, "LIMIT")
				}
				if !hasPerPartitionLimit && hasOrderBy {
					suggestions = append(suggestions, "PER")
				}
				if !hasAllowFiltering && hasWhere {
					suggestions = append(suggestions, "ALLOW")
				}

				return suggestions
			}
		}
	}

	// After WHERE clause
	if hasWhere && whereIndex >= 0 && wordPos > whereIndex {
		// Check if we're in the middle of WHERE conditions
		if lastWord == "AND" || lastWord == "OR" {
			return ce.getColumnNamesForCurrentTable(words, fromIndex)
		}

		// Check for comparison operators
		if ce.isColumnName(lastWord, words, fromIndex) {
			return ComparisonOperators
		}

		// After comparison operator, we expect a value (no completion)
		if ce.isComparisonOperator(lastWord) {
			return []string{}
		}

		// Otherwise suggest next clauses
		return ce.getSuggestionsAfterWhere(words, hasOrderBy, hasGroupBy, hasLimit, hasAllowFiltering, hasPerPartitionLimit)
	}

	// After ORDER BY clause
	if hasOrderBy && orderByIndex >= 0 && wordPos > orderByIndex+1 {
		// If last word is a column name, suggest sort order
		if ce.isColumnName(lastWord, words, fromIndex) {
			return AscDescComma
		}

		// After comma in ORDER BY, suggest more columns
		if lastWord == "," {
			return ce.getColumnNamesForCurrentTable(words, fromIndex)
		}

		return ce.getSuggestionsAfterOrderBy(words, hasWhere, hasLimit, hasAllowFiltering, hasPerPartitionLimit)
	}

	return []string{}
}

// Context-specific suggestion methods
