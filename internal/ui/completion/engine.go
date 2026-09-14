package completion

import (
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/session"
)

// CompletionEngine handles tab completion for CQL commands
type CompletionEngine struct {
	session        *db.Session
	sessionManager *session.Manager
	cache          *completionCache
}

// NewCompletionEngine creates a new completion engine
func NewCompletionEngine(dbSession *db.Session, sessionMgr *session.Manager) *CompletionEngine {
	return &CompletionEngine{
		session:        dbSession,
		sessionManager: sessionMgr,
		cache: &completionCache{
			tables:  make(map[string][]string),
			columns: make(map[string][]string),
		},
	}
}

// Complete returns possible completions for the given input, and a note about
// what to type where the next word is the user's to invent.
//
// The note is added here, in the one place every answer passes through, rather
// than in each of the places an answer is worked out.
func (ce *CompletionEngine) Complete(input string) []string {
	suggestions := ce.CompleteNative(input)

	suggestions = withoutASecondOpenBracket(input, suggestions)

	switch hint, placement := hintFor(input); placement {
	case hintInstead:
		return []string{hint}
	case hintAlongside:
		return append(suggestions, hint)
	case hintWhenNothingElse:
		if len(suggestions) == 0 {
			return []string{hint}
		}
	case noHint:
	}
	return suggestions
}

// withoutASecondOpenBracket drops the open bracket offered where one is open
// already.
//
// `INSERT INTO t (` was answered with "(", which is the bracket that has just
// been typed. The answer there is a column name, and the note saying so only
// appears when there is nothing else in the list.
func withoutASecondOpenBracket(input string, suggestions []string) []string {
	if !strings.HasSuffix(strings.TrimRight(input, " "), "(") {
		return suggestions
	}

	kept := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		if suggestion != "(" {
			kept = append(kept, suggestion)
		}
	}
	return kept
}

// CompleteNative returns possible completions for the given input using native pattern matching
func (ce *CompletionEngine) CompleteNative(input string) []string {
	// Debug output
	logger.DebugfToFile("Completion", "Complete called with input: '%s'", input)

	// For empty input, always return top-level commands
	if strings.TrimSpace(input) == "" {
		return ce.getTopLevelCommands()
	}

	// A batch is a run of statements, each finished with a semicolon. The one
	// being typed is completed as the statement it is.
	if inner, inBatch := statementInABatch(input); inBatch {
		return ce.CompleteNative(inner)
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

	// Special case: Handle COPY command with native completion
	// COPY is a meta-command, not part of CQL grammar
	if strings.HasPrefix(upperInput, "COPY ") || upperInput == "COPY" {
		// Use native completion for COPY
		return ce.handleCopyNativeCompletion(input)
	}

	// Check if we're trying to access UDT fields (column.field)
	if ce.isUDTFieldAccess(input) {
		if completions := ce.handleUDTFieldCompletion(input); completions != nil {
			return completions
		}
	}

	// Check if we're completing after a keyspace name with a dot
	// e.g., "SELECT * FROM system." or "COPY system." (but not INSERT INTO which was handled above)
	if strings.Contains(input, ".") && !strings.HasPrefix(upperInput, "INSERT INTO ") && !strings.HasPrefix(upperInput, "COPY ") {
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

	// A table statement is answered here rather than by the parser below.
	//
	// The parser counts words, and a table statement is the one that defeats
	// counting: IF NOT EXISTS moves everything along by three, a column
	// definition is any number of words, and the options are a map. It offered
	// types where the primary key goes and nothing at all after the table name.
	if shape, handled := tableStatementCompletions(input); handled {
		logger.DebugfToFile("Completion", "Table statement: %d suggestions", len(shape))
		return filterByWord(input, shape)
	}

	// An index statement is answered here for the same reason: what it is on
	// is in brackets, what it uses is a quoted name, and what it is given is a
	// map, none of which the parser's word positions describe.
	if shape, handled := indexStatementCompletions(input); handled {
		logger.DebugfToFile("Completion", "Index statement: %d suggestions", len(shape))
		return filterByWord(input, shape)
	}

	// A keyspace and a role are a WITH clause, the same shape as a table's.
	if shape, handled := keyspaceOptionCompletions(input); handled {
		logger.DebugfToFile("Completion", "Keyspace clause: %d suggestions", len(shape))
		return filterByWord(input, shape)
	}
	if shape, handled := roleOptionCompletions(input); handled {
		logger.DebugfToFile("Completion", "Role clause: %d suggestions", len(shape))
		return filterByWord(input, shape)
	}

	// The rest of the CREATE statements are a sequence of words in a fixed
	// order, which is written down once and walked.
	if shape, handled := createStatementCompletions(input); handled {
		logger.DebugfToFile("Completion", "Create statement: %d suggestions", len(shape))
		return filterByWord(input, shape)
	}

	// The statements that read and write rows are walked the same way.
	if shape, handled := ce.dmlCompletions(input); handled {
		logger.DebugfToFile("Completion", "Row statement: %d suggestions", len(shape))
		return filterByWord(input, shape)
	}

	// Use the simple completion engine
	logger.DebugfToFile("Completion", "Using simple completion for: '%s'", input)

	simpleEngine := NewSimpleCompletionEngine(ce)
	suggestions := simpleEngine.GetTokenCompletions(input)

	logger.DebugfToFile("Completion", "Parser returned %d suggestions", len(suggestions))

	// If parser doesn't return suggestions, fall back to native pattern matching
	if len(suggestions) == 0 {
		logger.DebugfToFile("Completion", "Falling back to native completion")
		nativeSuggestions := ce.completeNative(input)
		logger.DebugfToFile("Completion", "Native completion returned %d suggestions: %v", len(nativeSuggestions), nativeSuggestions)
		return nativeSuggestions
	}

	// Filter suggestions based on the partial word being typed
	// Get the word being completed (after the last space)
	var wordToComplete string
	endsWithSpace := strings.HasSuffix(input, " ")
	if !endsWithSpace {
		lastSpace := strings.LastIndex(input, " ")
		if lastSpace >= 0 {
			wordToComplete = input[lastSpace+1:]
		} else {
			wordToComplete = input
		}
	}

	// Apply filtering if we have a partial word
	if wordToComplete != "" {
		upperWord := strings.ToUpper(wordToComplete)

		// Check if any suggestion exactly matches the word being typed
		// If so, user wants the NEXT word, not to complete the current word
		exactMatch := false
		for _, s := range suggestions {
			if strings.ToUpper(s) == upperWord {
				exactMatch = true
				break
			}
		}

		// If there's an exact match, don't filter - show next word suggestions
		// Otherwise, filter suggestions that start with the partial word
		if !exactMatch {
			var filtered []string
			for _, s := range suggestions {
				if strings.HasPrefix(strings.ToUpper(s), upperWord) {
					filtered = append(filtered, s)
				}
			}
			logger.DebugfToFile("Completion", "After filtering for '%s': %d matches: %v", wordToComplete, len(filtered), filtered)
			suggestions = filtered
		} else {
			logger.DebugfToFile("Completion", "Exact match found for '%s', showing next word suggestions: %v", wordToComplete, suggestions)
		}
	}

	// Return just the suggestions (next words only), not full phrases
	// The keyboard handler will handle how to apply them to the input
	return suggestions
}

// completeNative is the native pattern-based completion method
func (ce *CompletionEngine) completeNative(input string) []string {
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
	logger.DebugfToFile("Completion", "completeNative: input='%s', words=%v, afterSpace=%v, wordToComplete='%s'",
		input, words, afterSpace, wordToComplete)

	// Get context-aware completions
	var suggestions []string

	// Debug the decision logic
	logger.DebugfToFile("Completion", "COMPLETION LOGIC: len(words)=%d, afterSpace=%v", len(words), afterSpace)

	switch {
	case len(words) == 0:
		suggestions = ce.getTopLevelCommands()
		logger.DebugfToFile("Completion", "BRANCH: No words - using top-level commands: %d suggestions", len(suggestions))
	case len(words) == 1 && !afterSpace:
		// Single partial word - check against top-level commands
		suggestions = ce.getTopLevelCommands()
		logger.DebugfToFile("Completion", "BRANCH: Single partial word '%s' - using top-level commands: %d suggestions", words[0], len(suggestions))
		if len(suggestions) > 0 {
			logger.DebugfToFile("Completion", "First few commands: %v", suggestions[:min(5, len(suggestions))])
		}
	default:
		suggestions = ce.getCompletionsForContext(input, words, afterSpace)
		logger.DebugfToFile("Completion", "BRANCH: Using context - getCompletionsForContext returned %d suggestions", len(suggestions))
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
		logger.DebugfToFile("Completion", "After filtering for '%s' (upper: '%s'): %d matches: %v", wordToComplete, upperWord, len(filtered), filtered)
		suggestions = filtered
	}

	// Return just the suggestions (next tokens only), not full lines
	// The keyboard handler will handle how to apply them
	return suggestions
}

// handleCopyNativeCompletion handles COPY command completion natively
func (ce *CompletionEngine) handleCopyNativeCompletion(input string) []string {
	// Parse the input more carefully to handle quotes
	upperInput := strings.ToUpper(strings.TrimSpace(input))

	// If just "COPY", return table names
	if upperInput == "COPY" {
		return ce.getTableAndKeyspaceTableNames()
	}

	// Check if we're completing after COPY with a dot (keyspace.table)
	if strings.Contains(input, ".") && !strings.Contains(input, "'") {
		// Let the keyspace.table completion handle it
		if completions := ce.handleKeyspaceTableCompletion(input); completions != nil {
			return completions
		}
	}

	// Parse to find TO/FROM and WITH
	// Check both with trailing space and at end of input
	hasTo := strings.Contains(upperInput, " TO ") || strings.HasSuffix(upperInput, " TO")
	hasFrom := strings.Contains(upperInput, " FROM ") || strings.HasSuffix(upperInput, " FROM")
	hasWith := strings.Contains(upperInput, " WITH ") || strings.HasSuffix(upperInput, " WITH")

	// Also check if the last word is WITH (user just typed it)
	words := strings.Fields(upperInput)
	lastWord := ""
	secondLastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
		if len(words) > 1 {
			secondLastWord = words[len(words)-2]
		}
	}

	// If we have WITH, provide context-aware completions
	if hasWith {
		// Check if we have a complete option assignment (e.g., FORMAT='PARQUET', HEADER=TRUE, etc.)
		// Look for patterns like OPTION=value, OPTION='value', or OPTION=number
		if strings.Contains(lastWord, "=") {
			// If the word contains = and ends with a quote, number, or TRUE/FALSE (case insensitive), it's complete
			upperLastWord := strings.ToUpper(lastWord)
			if strings.HasSuffix(lastWord, "'") ||
				strings.HasSuffix(upperLastWord, "TRUE") ||
				strings.HasSuffix(upperLastWord, "FALSE") ||
				len(lastWord) > 0 && (lastWord[len(lastWord)-1] >= '0' && lastWord[len(lastWord)-1] <= '9') {
				// Complete option assignment, suggest AND
				return []string{"AND"}
			}
		}

		// Check if the input ends with a value that completes an assignment
		trimmedInput := strings.TrimSpace(input)
		if strings.HasSuffix(trimmedInput, "'") ||
			strings.HasSuffix(upperInput, "TRUE") ||
			strings.HasSuffix(upperInput, "FALSE") {
			// Check if this is after an option assignment
			withPart := input[strings.Index(upperInput, "WITH"):]
			if strings.Contains(withPart, "=") {
				// We have an assignment with a value, suggest AND
				return []string{"AND"}
			}
		}

		// Check if last character is a digit (for numeric values like PAGESIZE=1000)
		if len(trimmedInput) > 0 {
			lastChar := trimmedInput[len(trimmedInput)-1]
			if lastChar >= '0' && lastChar <= '9' {
				// Check if this is after an equals sign
				withPart := input[strings.Index(upperInput, "WITH"):]
				if strings.Contains(withPart, "=") {
					return []string{"AND"}
				}
			}
		}

		// Check for specific option completions that need value lists
		if strings.HasSuffix(upperInput, "FORMAT=") || secondLastWord == "FORMAT" {
			return CopyFormats
		}
		if strings.HasSuffix(upperInput, "COMPRESSION=") || secondLastWord == "COMPRESSION" {
			return ParquetCompressionTypes
		}
		if lastWord == "AND" {
			return CopyOptions
		}
		// Default to options
		return CopyOptions
	}

	// If we just typed WITH, suggest options
	if lastWord == "WITH" {
		return CopyOptions
	}

	// If we have TO or FROM but no WITH
	if (hasTo || hasFrom) && !hasWith {
		// Check if we're right after TO or FROM (with or without space)
		trimmed := strings.TrimSpace(input)
		if strings.HasSuffix(trimmed, " TO") || strings.HasSuffix(trimmed, " FROM") ||
			(lastWord == "TO" || lastWord == "FROM") {
			return CopyFileSuggestions
		}

		// Check if we end with a quote (completed filename) or if we're past a filename
		if strings.HasSuffix(trimmed, "'") || strings.HasSuffix(trimmed, "\"") {
			return []string{"WITH"}
		}
		// Check if we have a complete filename (has both opening and closing quotes)
		if (strings.Count(input, "'") >= 2) || (strings.Count(input, "\"") >= 2) {
			return []string{"WITH"}
		}
		// If no closing quote yet, don't suggest anything (user is typing filename)
		return []string{}
	}

	// If we don't have TO/FROM yet
	if !hasTo && !hasFrom {
		// Check if the last word IS "TO" or "FROM" (user just typed it)
		if lastWord == "TO" || lastWord == "FROM" {
			// User just typed TO/FROM, suggest file paths
			return CopyFileSuggestions
		}

		// Check if we have a table name
		words := strings.Fields(upperInput)
		if len(words) >= 2 {
			// We have "COPY tablename", suggest TO/FROM
			return CopyDirections
		}
	}

	return []string{}
}

// getCompletionsForContext returns completions based on the command context
func (ce *CompletionEngine) getCompletionsForContext(input string, words []string, afterSpace bool) []string {
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

	// A table statement is answered by the same function that answers it
	// higher up, rather than by a second copy of the rules. The copy here
	// decided a materialized view was a table and offered the bracket its
	// columns would go in.
	if suggestions, handled := tableStatementCompletions(input); handled {
		return suggestions
	}

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
			return BatchTypes
		}
		if wordPos == 2 && len(words) > 1 {
			switch words[1] {
			case "UNLOGGED", "COUNTER":
				return []string{BatchTypes[0]} // "BATCH"
			}
		}
	case "APPLY":
		if wordPos == 1 {
			return []string{BatchTypes[0]} // "BATCH"
		}
	case "LIST":
		if wordPos == 1 {
			return ListTargets
		}
	case "OUTPUT":
		if wordPos == 1 {
			return OutputFormats
		}
	case "CONSISTENCY":
		if wordPos == 1 {
			return ConsistencyLevels
		}
	case "AUTOFETCH":
		if wordPos == 1 {
			return []string{"ON", "OFF"}
		}
	case "COPY":
		logger.DebugfToFile("Completion", "getCompletionsForContext: Routing to getCopyCompletions")
		return ce.getCopyCompletions(words, wordPos)
	}

	// If we don't recognize the command, return empty
	return []string{}
}

// isTableStatement reports whether the statement is about a table, which is
// what has a WITH clause of options.
//
// A materialized view takes the same options, and reaches the same place: ALTER
// MATERIALIZED VIEW ... WITH is the table's list.
func isTableStatement(words []string) bool {
	if len(words) < 2 {
		return false
	}

	switch words[1] {
	case "TABLE", "COLUMNFAMILY":
		return true
	case "MATERIALIZED":
		return true
	}
	return false
}
