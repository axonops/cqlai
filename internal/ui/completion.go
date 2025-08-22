package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/axonops/cqlai/internal/db"
)

// CompletionEngine handles tab completion for CQL commands
type CompletionEngine struct {
	session   *db.Session
	cache     *completionCache
	lastInput string // Store last input for context
}

// NewCompletionEngine creates a new completion engine
func NewCompletionEngine(session *db.Session) *CompletionEngine {
	return &CompletionEngine{
		session: session,
		cache: &completionCache{
			tables:  make(map[string][]string),
			columns: make(map[string][]string),
		},
	}
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
		return []string{"VALUES"}
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

// handleKeyspaceTableCompletion handles completions for keyspace.table patterns
func (ce *CompletionEngine) handleKeyspaceTableCompletion(input string) []string {
	// Find the last dot
	lastDot := strings.LastIndex(input, ".")
	beforeDot := input[:lastDot]
	afterDot := input[lastDot+1:]

	// Check if we have a keyspace name pattern (word before the dot)
	words := strings.Fields(beforeDot)
	if len(words) >= 2 {
		// Get the potential keyspace name (last word before the dot)
		potentialKeyspace := words[len(words)-1]

		// Check if this is a valid keyspace
		keyspaces := ce.getKeyspaceNames()
		for _, ks := range keyspaces {
			if strings.EqualFold(ks, potentialKeyspace) {
				// We have a valid keyspace - check tables from that keyspace
				tables := ce.getTablesForKeyspace(ks)

				// If there's text after the dot, check if it's a complete table name or partial
				if afterDot != "" {
					// Check for exact match first
					var isCompleteTable bool
					for _, table := range tables {
						if strings.EqualFold(table, afterDot) {
							isCompleteTable = true
							break
						}
					}

					// If it's a complete table name, don't try to complete it further
					// Let the parser-based completion handle what comes next
					if isCompleteTable {
						// Debug logging
						if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
							fmt.Fprintf(debugFile, "[DEBUG] Complete table name detected: %s.%s, falling through to parser\n", ks, afterDot)
							defer debugFile.Close()
						}
						// Return nil to indicate fall through to parser
						return nil
					}

					// It's a partial table name - filter and return matches
					var filtered []string
					upperAfterDot := strings.ToUpper(afterDot)
					for _, table := range tables {
						if strings.HasPrefix(strings.ToUpper(table), upperAfterDot) {
							// Return just the table name, not the full path
							filtered = append(filtered, table)
						}
					}
					if len(filtered) > 0 {
						return filtered
					}
				} else {
					// No text after dot - return just the table names
					// The keyboard handler will append them correctly
					return tables
				}
			}
		}
	}
	return nil
}

// Complete returns possible completions for the given input
func (ce *CompletionEngine) Complete(input string) []string {
	return ce.CompleteLegacy(input)
}

// CompleteLegacy returns possible completions for the given input using the old implementation
func (ce *CompletionEngine) CompleteLegacy(input string) []string {
	// Debug output
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] Complete called with input: '%s'\n", input)
		defer debugFile.Close()
	}

	// For empty input, always return top-level commands
	if strings.TrimSpace(input) == "" {
		return ce.getTopLevelCommands()
	}

	// Special case: Check for INSERT INTO pattern FIRST
	// This must come before the generic dot handling
	upperInput := strings.ToUpper(strings.TrimSpace(input))
	if strings.HasPrefix(upperInput, "INSERT INTO ") {
		afterInto := input[12:] // Skip "INSERT INTO "
		if completions := ce.handleInsertIntoCompletion(input, afterInto); completions != nil {
			return completions
		}
		// Fall through to parser if no special handling
	}

	// Check if we're completing after a keyspace name with a dot
	// e.g., "SELECT * FROM system." (but not INSERT INTO which was handled above)
	if strings.Contains(input, ".") && !strings.HasPrefix(upperInput, "INSERT INTO ") {
		if completions := ce.handleKeyspaceTableCompletion(input); completions != nil {
			return completions
		}
		// Fall through to parser if handleKeyspaceTableCompletion returns nil
	}

	// For simple partial words (no spaces), check top-level commands first
	if !strings.Contains(input, " ") {
		upperInput := strings.ToUpper(input)
		var matches []string
		for _, cmd := range ce.getTopLevelCommands() {
			if strings.HasPrefix(cmd, upperInput) {
				matches = append(matches, cmd)
			}
		}
		if len(matches) > 0 {
			return matches
		}
	}

	// Check if this looks like a complete INSERT statement
	// If it ends with ) and has VALUES with balanced parens, it's likely complete
	if strings.HasPrefix(upperInput, "INSERT INTO ") && strings.Contains(upperInput, "VALUES") && strings.HasSuffix(strings.TrimSpace(input), ")") {
		// Count parentheses after VALUES to check if balanced
		valuesIdx := strings.Index(upperInput, "VALUES")
		if valuesIdx > 0 {
			afterValues := input[valuesIdx+6:]
			openCount := strings.Count(afterValues, "(")
			closeCount := strings.Count(afterValues, ")")
			if closeCount >= openCount && closeCount > 0 {
				// Complete INSERT statement - return no suggestions
				if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
					fmt.Fprintf(debugFile, "[DEBUG] Complete INSERT statement detected, returning no suggestions\n")
					defer debugFile.Close()
				}
				return []string{}
			}
		}
	}

	// Use the parser-based completion engine
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] Using parser-based completion for: '%s'\n", input)
		defer debugFile.Close()
	}

	parserEngine := NewParserBasedCompletionEngine(ce)
	suggestions := parserEngine.GetTokenCompletions(input)

	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] Parser returned %d suggestions\n", len(suggestions))
		defer debugFile.Close()
	}

	// If parser doesn't return suggestions, fall back to legacy approach
	if len(suggestions) == 0 {
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] Falling back to legacy completion\n")
			defer debugFile.Close()
		}
		legacySuggestions := ce.completeLegacy(input)
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] Legacy completion returned %d suggestions: %v\n", len(legacySuggestions), legacySuggestions)
			defer debugFile.Close()
		}
		return legacySuggestions
	}

	// Return just the suggestions (next words only), not full phrases
	// The keyboard handler will handle how to apply them to the input
	return suggestions
}

// completeLegacy is the fallback completion method
func (ce *CompletionEngine) completeLegacy(input string) []string {
	// Don't trim - we need to know if there's trailing space
	if input == "" {
		return ce.getTopLevelCommands()
	}

	// Check if we're at the end of a word or after a space
	endsWithSpace := strings.HasSuffix(input, " ")

	// Get the word being completed
	var wordToComplete string

	if endsWithSpace {
		// Completing a new word after space
		wordToComplete = ""
	} else {
		// Find the last word
		lastSpace := strings.LastIndex(input, " ")
		if lastSpace == -1 {
			// Single word
			wordToComplete = input
		} else {
			// Multiple words - get the last partial word
			wordToComplete = input[lastSpace+1:]
		}
	}

	// Get all words
	words := strings.Fields(strings.ToUpper(input))

	// If we're completing after a space, we want suggestions for the next position
	afterSpace := endsWithSpace

	// Debug logging
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] completeLegacy: input='%s', words=%v, afterSpace=%v, wordToComplete='%s'\n", 
			input, words, afterSpace, wordToComplete)
		defer debugFile.Close()
	}

	// Get context-aware completions
	var suggestions []string
	if len(words) == 0 {
		suggestions = ce.getTopLevelCommands()
	} else {
		suggestions = ce.getCompletionsForContext(words, afterSpace)
	}

	// Debug what we got from context
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] getCompletionsForContext returned %d suggestions: %v\n", len(suggestions), suggestions)
		defer debugFile.Close()
	}

	// Filter suggestions based on the partial word (case-insensitive)
	if wordToComplete != "" {
		upperWord := strings.ToUpper(wordToComplete)
		var filtered []string
		for _, s := range suggestions {
			if strings.HasPrefix(strings.ToUpper(s), upperWord) {
				filtered = append(filtered, s)
			}
		}
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] After filtering for '%s' (upper: '%s'): %d matches: %v\n", wordToComplete, upperWord, len(filtered), filtered)
			defer debugFile.Close()
		}
		suggestions = filtered
	}

	// Return just the suggestions (next tokens only), not full lines
	// The keyboard handler will handle how to apply them
	return suggestions
}

// getCompletionsForContext returns completions based on the command context
func (ce *CompletionEngine) getCompletionsForContext(words []string, afterSpace bool) []string {
	if len(words) == 0 {
		return ce.getTopLevelCommands()
	}

	// Determine the position we're completing
	wordPos := len(words)
	if !afterSpace && wordPos > 0 {
		wordPos-- // We're still on the last word, not after it
	}

	// Get the first word to determine command type
	firstWord := words[0]

	switch firstWord {
	case "SELECT":
		return ce.getSelectCompletions(words, wordPos)
	case "INSERT":
		return ce.getInsertCompletions(words, wordPos)
	case "UPDATE":
		return ce.getUpdateCompletions(words, wordPos)
	case "DELETE":
		return ce.getDeleteCompletions(words, wordPos)
	case "CREATE":
		return ce.getCreateCompletions(words, wordPos)
	case "DROP":
		return ce.getDropCompletions(words, wordPos)
	case "ALTER":
		return ce.getAlterCompletions(words, wordPos)
	case "TRUNCATE":
		return ce.getTruncateCompletions(words, wordPos)
	case "GRANT":
		return ce.getGrantCompletions(words, wordPos)
	case "REVOKE":
		return ce.getRevokeCompletions(words, wordPos)
	case "DESCRIBE", "DESC":
		return ce.getDescribeCompletions(words, wordPos)
	case "USE":
		return ce.getUseCompletions(words, wordPos)
	case "SHOW":
		return ce.getShowCompletions(words, wordPos)
	case "BEGIN":
		if wordPos == 1 {
			return []string{"BATCH", "UNLOGGED", "COUNTER"}
		}
		if wordPos == 2 && len(words) > 1 {
			switch words[1] {
			case "UNLOGGED", "COUNTER":
				return []string{"BATCH"}
			}
		}
	case "APPLY":
		if wordPos == 1 {
			return []string{"BATCH"}
		}
	case "LIST":
		if wordPos == 1 {
			return []string{"USERS", "ROLES", "PERMISSIONS"}
		}
	}

	// If we don't recognize the command, return empty
	return []string{}
}
