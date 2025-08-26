package completion

import (
	"fmt"
	"os"
	"strings"
)

// getInsertCompletions returns completions for INSERT commands
func (ce *CompletionEngine) getInsertCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return IntoKeyword
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
		return NotKeyword
	case "NOT":
		if len(words) > 1 && words[len(words)-2] == "IF" {
			return ExistsKeyword
		}
	case "USING":
		return UsingOptions
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
			return UsingOptions
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
		return IntoKeyword
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

// handleInsertIntoCompletion handles completions for INSERT INTO statements
func (ce *CompletionEngine) handleInsertIntoCompletion(input string, afterInto string) []string {
	afterIntoTrimmed := strings.TrimSpace(afterInto)

	// Debug output
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] INSERT INTO pattern detected. Input: '%s', After INTO: '%s' (trimmed: '%s')\n", input, afterInto, afterIntoTrimmed)
		defer debugFile.Close()
	}

	// Check different states of INSERT INTO statement
	if strings.Contains(afterIntoTrimmed, "VALUES") {
		return ce.handleInsertValuesCompletion(input, afterIntoTrimmed)
	} else if strings.Contains(afterIntoTrimmed, "(") && strings.Contains(afterIntoTrimmed, ")") &&
		!strings.Contains(afterIntoTrimmed, "VALUES") {
		// We have columns specified but no VALUES yet - suggest VALUES
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] Columns specified in INSERT INTO, no VALUES yet\n")
			defer debugFile.Close()
		}
		return ValuesKeyword
	}

	// Check if it's in keyspace.table format
	if afterIntoTrimmed != "" && strings.Contains(afterIntoTrimmed, ".") &&
		!strings.Contains(afterIntoTrimmed, " ") && !strings.Contains(afterIntoTrimmed, "(") {
		return ce.handleInsertKeyspaceTableCompletion(input, afterIntoTrimmed)
	}

	return nil // No special handling, fall through to parser
}

// handleInsertValuesCompletion handles completions after VALUES keyword
func (ce *CompletionEngine) handleInsertValuesCompletion(input string, afterIntoTrimmed string) []string {
	// We have VALUES keyword - check what comes after
	afterValues := afterIntoTrimmed[strings.Index(afterIntoTrimmed, "VALUES")+6:]
	afterValues = strings.TrimSpace(afterValues)

	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] VALUES found. After VALUES: '%s'\n", afterValues)
		defer debugFile.Close()
	}

	if afterValues == "" {
		// Just "VALUES" - suggest opening parenthesis
		// Return just the next token, not the full input
		return []string{"("}
	} else if afterValues == "(" {
		// After "VALUES (", suggest data type template
		return ce.handleInsertValueTemplateCompletion(input, afterIntoTrimmed)
	} else if strings.Contains(afterValues, ")") {
		// Check if VALUES clause has balanced parentheses
		openCount := strings.Count(afterValues, "(")
		closeCount := strings.Count(afterValues, ")")

		// If we have more or equal closing parens than opening ones,
		// the VALUES clause is likely complete
		if closeCount >= openCount {
			// Complete VALUES clause - don't suggest another opening paren
			// Let the parser handle what comes after a complete INSERT
			return nil // Let parser handle post-VALUES completions
		}
	}
	return nil
}

// handleInsertValueTemplateCompletion builds value template after VALUES (
func (ce *CompletionEngine) handleInsertValueTemplateCompletion(input string, afterIntoTrimmed string) []string {
	// Extract table name and column names from the INSERT statement
	var tableName string
	var columnNames []string

	// Find table name (before the first opening paren)
	beforeFirstParen := afterIntoTrimmed
	if idx := strings.Index(afterIntoTrimmed, "("); idx > 0 {
		beforeFirstParen = afterIntoTrimmed[:idx]
	}
	tableName = strings.TrimSpace(beforeFirstParen)

	// Extract column names from between parentheses
	if startIdx := strings.Index(afterIntoTrimmed, "("); startIdx >= 0 {
		if endIdx := strings.Index(afterIntoTrimmed[startIdx:], ")"); endIdx > 0 {
			columnsPart := afterIntoTrimmed[startIdx+1 : startIdx+endIdx]
			// Split by comma and clean up
			for _, col := range strings.Split(columnsPart, ",") {
				col = strings.TrimSpace(col)
				if col != "" {
					columnNames = append(columnNames, col)
				}
			}
		}
	}

	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] After VALUES (, table='%s', columns=%v\n", tableName, columnNames)
		defer debugFile.Close()
	}

	// Get column types and build template
	if tableName != "" && len(columnNames) > 0 {
		typeTemplate := ce.getColumnTypeTemplate(tableName, columnNames)
		if typeTemplate != "" {
			// Return the template wrapped in parentheses
			return []string{"(" + typeTemplate + ")"}
		}
	}
	return nil
}

// handleInsertKeyspaceTableCompletion handles keyspace.table patterns in INSERT INTO
func (ce *CompletionEngine) handleInsertKeyspaceTableCompletion(input string, afterIntoTrimmed string) []string {
	// Split to check if we have keyspace.table or just keyspace.
	parts := strings.Split(afterIntoTrimmed, ".")

	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] Keyspace.table check: '%s', parts=%v\n", afterIntoTrimmed, parts)
		defer debugFile.Close()
	}

	if len(parts) == 2 {
		keyspaceName := parts[0]
		tableNamePart := parts[1]

		if tableNamePart == "" {
			// This is "keyspace." - should suggest table names
			if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				fmt.Fprintf(debugFile, "[DEBUG] Incomplete keyspace. pattern, keyspace='%s'\n", keyspaceName)
				defer debugFile.Close()
			}

			// Get tables for this keyspace
			tables := ce.getTablesForKeyspace(keyspaceName)

			if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				fmt.Fprintf(debugFile, "[DEBUG] Returning %d table completions for keyspace %s\n", len(tables), keyspaceName)
				defer debugFile.Close()
			}
			// Return just the table names, not the full input
			return tables
		} else {
			// We have some text after the dot - could be partial or complete table name
			return ce.handleInsertTableNameCompletion(input, keyspaceName, tableNamePart, afterIntoTrimmed)
		}
	}
	return nil
}

// handleInsertTableNameCompletion handles table name completion after keyspace.
func (ce *CompletionEngine) handleInsertTableNameCompletion(input string, keyspaceName string, tableNamePart string, afterIntoTrimmed string) []string {
	// First check if it's a partial table name
	tables := ce.getTablesForKeyspace(keyspaceName)
	var matchingTables []string
	var exactMatch bool

	upperTablePart := strings.ToUpper(tableNamePart)
	for _, table := range tables {
		if strings.ToUpper(table) == upperTablePart {
			exactMatch = true
			break
		}
		if strings.HasPrefix(strings.ToUpper(table), upperTablePart) {
			matchingTables = append(matchingTables, table)
		}
	}

	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] Table part '%s', exact match: %v, matching tables: %v\n",
			tableNamePart, exactMatch, matchingTables)
		defer debugFile.Close()
	}

	if exactMatch {
		// This is a complete, valid table name - suggest what comes after
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] Complete keyspace.table: '%s'\n", afterIntoTrimmed)
			defer debugFile.Close()
		}

		columns := ce.getColumnNamesForTable(afterIntoTrimmed)
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] Got %d columns for table %s\n", len(columns), afterIntoTrimmed)
			defer debugFile.Close()
		}

		// Return completions for what comes after the table name
		var completions []string

		// If we have columns, suggest the formatted column list
		if len(columns) > 0 {
			columnList := "(" + strings.Join(columns, ", ") + ")"
			completions = append(completions, columnList)
		}

		// Always suggest VALUES and opening parenthesis
		completions = append(completions, "VALUES")
		completions = append(completions, "(")
		completions = append(completions, "IF")
		completions = append(completions, "JSON")
		completions = append(completions, "USING")

		// Return these completions (just the next tokens, not full lines)
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] Returning %d completions for complete keyspace.table\n", len(completions))
			defer debugFile.Close()
		}
		return completions
	} else if len(matchingTables) > 0 {
		// This is a partial table name - suggest matching table names
		// Return just the table names
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] Returning %d partial table completions\n", len(matchingTables))
			defer debugFile.Close()
		}
		return matchingTables
	}
	return nil
}
