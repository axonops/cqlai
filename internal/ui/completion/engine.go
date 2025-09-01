package completion

import (
	"fmt"
	"os"
	"strings"

	"github.com/axonops/cqlai/internal/db"
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

// Complete returns possible completions for the given input
func (ce *CompletionEngine) Complete(input string) []string {
	return ce.CompleteLegacy(input)
}

// CompleteLegacy returns possible completions for the given input using the old implementation
func (ce *CompletionEngine) CompleteLegacy(input string) []string {
	// Debug output
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
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
				if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
					fmt.Fprintf(debugFile, "[DEBUG] Complete INSERT statement detected, returning no suggestions\n")
					defer debugFile.Close()
				}
				return []string{}
			}
		}
	}

	// Use the parser-based completion engine
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] Using parser-based completion for: '%s'\n", input)
		defer debugFile.Close()
	}

	parserEngine := NewParserBasedCompletionEngine(ce)
	suggestions := parserEngine.GetTokenCompletions(input)

	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] Parser returned %d suggestions\n", len(suggestions))
		defer debugFile.Close()
	}

	// If parser doesn't return suggestions, fall back to legacy approach
	if len(suggestions) == 0 {
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] Falling back to legacy completion\n")
			defer debugFile.Close()
		}
		legacySuggestions := ce.completeLegacy(input)
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
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
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] completeLegacy: input='%s', words=%v, afterSpace=%v, wordToComplete='%s'\n",
			input, words, afterSpace, wordToComplete)
		defer debugFile.Close()
	}

	// Get context-aware completions
	var suggestions []string
	
	// Debug the decision logic
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
		fmt.Fprintf(debugFile, "[DEBUG] COMPLETION LOGIC: len(words)=%d, afterSpace=%v\n", len(words), afterSpace)
		defer debugFile.Close()
	}
	
	switch {
	case len(words) == 0:
		suggestions = ce.getTopLevelCommands()
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] BRANCH: No words - using top-level commands: %d suggestions\n", len(suggestions))
			defer debugFile.Close()
		}
	case len(words) == 1 && !afterSpace:
		// Single partial word - check against top-level commands
		suggestions = ce.getTopLevelCommands()
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] BRANCH: Single partial word '%s' - using top-level commands: %d suggestions\n", words[0], len(suggestions))
			if len(suggestions) > 0 {
				fmt.Fprintf(debugFile, "[DEBUG] First few commands: %v\n", suggestions[:5])
			}
			defer debugFile.Close()
		}
	default:
		suggestions = ce.getCompletionsForContext(words, afterSpace)
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] BRANCH: Using context - getCompletionsForContext returned %d suggestions\n", len(suggestions))
			defer debugFile.Close()
		}
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
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
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
	}

	// If we don't recognize the command, return empty
	return []string{}
}
