package ui

import (
	"fmt"
	"os"
	"strings"
)

// getInsertCompletions returns completions for INSERT commands
func (ce *CompletionEngine) getInsertCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return []string{"INTO"}
	}

	if wordPos == 2 && len(words) > 1 && words[1] == "INTO" {
		return ce.getTableAndKeyspaceNames()
	}

	// Track what we've seen in the command
	hasColumns := false
	hasValues := false
	hasJson := false
	hasIfNotExists := false
	hasUsing := false
	tablePos := -1

	// Find key elements in the command
	for i, word := range words {
		switch word {
		case "INTO":
			if i+1 < len(words) {
				tablePos = i + 1
			}
		case "VALUES":
			hasValues = true
		case "JSON":
			hasJson = true
		case "IF":
			if i+1 < len(words) && words[i+1] == "NOT" && i+2 < len(words) && words[i+2] == "EXISTS" {
				hasIfNotExists = true
			}
		case "USING":
			hasUsing = true
		}

		// Check for column list (parentheses after table name)
		if tablePos > 0 && i == tablePos+1 && strings.HasPrefix(word, "(") {
			hasColumns = true
		}
	}

	// Get the last word to determine context
	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	// Special keyword handling
	switch lastWord {
	case "INTO":
		return ce.getTableAndKeyspaceNames()
	case "VALUES":
		// After VALUES, expect opening parenthesis
		return []string{"("}
	case "IF":
		return []string{"NOT"}
	case "NOT":
		if len(words) > 1 && words[len(words)-2] == "IF" {
			return []string{"EXISTS"}
		}
	case "USING":
		return []string{"TTL", "TIMESTAMP"}
	case "TTL":
		// After TTL, expect a number (no completion)
		return []string{}
	case "TIMESTAMP":
		// After TIMESTAMP, expect a number (no completion)
		return []string{}
	case "JSON":
		// After JSON, expect JSON string (no completion)
		return []string{}
	case "AND":
		// In USING clause
		if hasUsing {
			return []string{"TTL", "TIMESTAMP"}
		}
	}

	// After table name
	if tablePos >= 0 && wordPos == tablePos+1 {
		var suggestions []string

		// Get the table name and suggest formatted column list
		if tablePos < len(words) {
			tableName := words[tablePos]
			columns := ce.getColumnNamesForTable(tableName)

			// If we have columns, suggest the formatted column list first
			if len(columns) > 0 {
				columnList := "(" + strings.Join(columns, ", ") + ")"
				suggestions = append(suggestions, columnList)
			}
		}

		// Can specify columns manually
		suggestions = append(suggestions, "(")

		// Or go directly to VALUES or JSON
		if !hasJson && !hasValues {
			suggestions = append(suggestions, "VALUES", "JSON")
		}

		// Or add IF NOT EXISTS
		if !hasIfNotExists && !hasValues && !hasJson {
			suggestions = append(suggestions, "IF")
		}

		// Or add USING clause
		if !hasUsing && !hasValues && !hasJson {
			suggestions = append(suggestions, "USING")
		}

		return suggestions
	}

	// After column specification or when we need VALUES/JSON
	if tablePos >= 0 && wordPos > tablePos+1 {
		if !hasValues && !hasJson {
			var suggestions []string

			suggestions = append(suggestions, "VALUES")
			if !hasColumns {
				suggestions = append(suggestions, "JSON")
			}

			if !hasIfNotExists {
				suggestions = append(suggestions, "IF")
			}

			if !hasUsing {
				suggestions = append(suggestions, "USING")
			}

			return suggestions
		}

		// After VALUES or JSON
		if hasValues || hasJson {
			if !hasIfNotExists && !hasUsing {
				var suggestions []string
				if !hasIfNotExists {
					suggestions = append(suggestions, "IF")
				}
				if !hasUsing {
					suggestions = append(suggestions, "USING")
				}
				return suggestions
			}
		}
	}

	return []string{}
}

func (pce *ParserBasedCompletionEngine) getInsertSuggestions(tokens []string) []string {
	// Debug logging
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] getInsertSuggestions called with tokens: %v\n", tokens)
		defer debugFile.Close()
	}

	if len(tokens) == 1 {
		return []string{"INTO"}
	}

	// Check what tokens we have
	hasInto := false
	intoIndex := -1
	hasValues := false
	for i, t := range tokens {
		if t == "INTO" {
			hasInto = true
			intoIndex = i
		}
		if t == "VALUES" {
			hasValues = true
		}
	}

	if hasInto && !hasValues {
		// Check if we're right after INTO
		if intoIndex == len(tokens)-1 {
			// After INTO, suggest table names and keyspaces
			return pce.getTableAndKeyspaceNames()
		}

		// Check for keyspace.table pattern
		// Look for pattern: INTO keyspace . table
		tableName := ""
		if intoIndex+3 < len(tokens) && intoIndex+2 < len(tokens) && tokens[intoIndex+2] == "." {
			// We have keyspace.table format
			keyspace := tokens[intoIndex+1]
			table := tokens[intoIndex+3]
			tableName = keyspace + "." + table

			// Check if we're at or right after the table name
			if intoIndex+3 >= len(tokens)-1 || intoIndex+4 == len(tokens) {
				// Get columns for this table
				columns := pce.getColumnNamesForTable(tableName)

				suggestions := []string{}

				// If we have columns, suggest the formatted column list
				if len(columns) > 0 {
					columnList := "(" + strings.Join(columns, ", ") + ")"
					suggestions = append(suggestions, columnList)
				}

				// Also suggest VALUES and opening parenthesis
				suggestions = append(suggestions, "VALUES", "(")
				return suggestions
			}
		} else if intoIndex+1 < len(tokens) {
			// Check if we have exactly one token after INTO that's not a keyword
			if intoIndex+1 == len(tokens)-1 {
				lastToken := tokens[len(tokens)-1]
				// Check if the last token is a CQL keyword - if not, it's likely a partial table name
				if !IsCompleteKeyword(lastToken) {
					// We're typing a table name - return table suggestions
					return pce.getTableAndKeyspaceNames()
				}

				// It's a complete table name without keyspace
				tableName = lastToken
			} else if tokens[len(tokens)-1] == "." && intoIndex+2 == len(tokens) {
				// We have "INSERT INTO keyspace." - suggest tables from that keyspace
				// For now, return empty as we'd need to track the keyspace name
				// This would require more complex parsing
				return []string{}
			}
		}

		// After INTO tablename (without keyspace), suggest column list and VALUES
		if tableName != "" {
			// Get column names for this table
			columns := pce.getColumnNamesForTable(tableName)

			suggestions := []string{}

			// If we have columns, suggest the formatted column list
			if len(columns) > 0 {
				columnList := "(" + strings.Join(columns, ", ") + ")"
				suggestions = append(suggestions, columnList)
			}

			// Also suggest VALUES and opening parenthesis
			suggestions = append(suggestions, "VALUES", "(")
			return suggestions
		}

		// Default suggestions after INTO
		return []string{"(", "VALUES"}
	}

	if hasValues {
		// After VALUES, check if we need to suggest completions
		valuesIndex := -1
		for i, t := range tokens {
			if t == "VALUES" {
				valuesIndex = i
				break
			}
		}

		// Only suggest "(" if VALUES is the very last token
		if valuesIndex == len(tokens)-1 {
			return []string{"("}
		}

		// Don't suggest "(" if there's already content after VALUES
		// The VALUES clause is likely complete if we have tokens after VALUES
		// Just suggest what can come after a complete VALUES clause
		result := []string{"IF", "USING"}
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] getInsertSuggestions returning (after VALUES): %v\n", result)
			debugFile.Close()
		}
		return result
	}

	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] getInsertSuggestions returning empty\n")
		debugFile.Close()
	}
	return []string{}
}
