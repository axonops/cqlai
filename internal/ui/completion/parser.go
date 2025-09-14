package completion

import (
	"fmt"
	"os"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/axonops/cqlai/internal/parser/grammar"
)

// ParserBasedCompletionEngine provides completions using the ANTLR parser
type ParserBasedCompletionEngine struct {
	*CompletionEngine
}

// NewParserBasedCompletionEngine creates a completion engine that uses the ANTLR parser
func NewParserBasedCompletionEngine(ce *CompletionEngine) *ParserBasedCompletionEngine {
	return &ParserBasedCompletionEngine{
		CompletionEngine: ce,
	}
}

// GetTokenCompletions analyzes the partial input using ANTLR to provide context-aware completions
func (pce *ParserBasedCompletionEngine) GetTokenCompletions(input string) []string {
	// Create lexer and tokenize the input
	is := antlr.NewInputStream(input)
	lexer := grammar.NewCqlLexer(is)

	// Collect all tokens
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	stream.Fill()
	tokens := stream.GetAllTokens()

	// If input is empty or only whitespace, return top-level commands
	if len(tokens) == 0 || (len(tokens) == 1 && tokens[0].GetTokenType() == antlr.TokenEOF) {
		return pce.getTopLevelKeywords()
	}

	// Get the last meaningful token (ignoring EOF)
	var lastToken antlr.Token
	for i := len(tokens) - 1; i >= 0; i-- {
		if tokens[i].GetTokenType() != antlr.TokenEOF {
			lastToken = tokens[i]
			break
		}
	}

	// Check if we're in the middle of typing a keyword
	endsWithSpace := strings.HasSuffix(input, " ")

	// Debug logging
	if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
		tokenStrs := []string{}
		for _, t := range tokens {
			if t.GetTokenType() != antlr.TokenEOF {
				tokenStrs = append(tokenStrs, t.GetText())
			}
		}
		fmt.Fprintf(debugFile, "[DEBUG] GetTokenCompletions: input='%s', endsWithSpace=%v, tokens=%v\n", input, endsWithSpace, tokenStrs)
		defer debugFile.Close()
	}

	// Special case: Check if we have a complete INSERT INTO table statement
	// Pattern: INSERT INTO [keyspace.]table - should suggest column list
	if len(tokens) >= 3 {
		// Check for INSERT INTO pattern
		tokenStrs := []string{}
		for _, t := range tokens {
			if t.GetTokenType() != antlr.TokenEOF {
				tokenStrs = append(tokenStrs, strings.ToUpper(t.GetText()))
			}
		}

		// Check if we have INSERT INTO followed by table name
		if len(tokenStrs) >= 3 && tokenStrs[0] == "INSERT" && tokenStrs[1] == "INTO" {
			// Check if we have keyspace.table (5 tokens) or just table (3 tokens)
			// And we're at the end of the table name (no space or at end of input)
			if !endsWithSpace && (len(tokenStrs) == 3 || (len(tokenStrs) == 5 && tokenStrs[3] == ".")) {
				// We have a complete table specification, return what comes next
				return pce.getNextTokenSuggestions(tokens)
			}
		}
	}

	// Special case: Check for SELECT FROM keyspace.table pattern
	if len(tokens) >= 6 {
		tokenStrs := []string{}
		for _, t := range tokens {
			if t.GetTokenType() != antlr.TokenEOF {
				tokenStrs = append(tokenStrs, strings.ToUpper(t.GetText()))
			}
		}

		// Check for SELECT * FROM keyspace.table pattern
		if len(tokenStrs) >= 6 && tokenStrs[0] == "SELECT" && tokenStrs[2] == "FROM" && tokenStrs[4] == "." {
			// We have SELECT ... FROM keyspace . table - this is complete
			if !endsWithSpace {
				// Return suggestions for what comes after the table
				return pce.getNextTokenSuggestions(tokens)
			}
		}
	}

	// If the input ends with space, we're starting a new token
	if endsWithSpace {
		return pce.getNextTokenSuggestions(tokens)
	}

	// Otherwise, we're completing the current token
	if lastToken != nil {
		partial := lastToken.GetText()

		// Debug logging
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] Checking last token: '%s', total tokens: %d\n", partial, len(tokens))
			if len(tokens) >= 2 {
				fmt.Fprintf(debugFile, "[DEBUG] Second to last token: '%s'\n", tokens[len(tokens)-2].GetText())
			}
			defer debugFile.Close()
		}

		// Check if the last token is a complete special token (like *, COUNT(*), etc.)
		if partial == "*" || partial == "COUNT(*)" || partial == ")" {
			// These are complete tokens, get suggestions for what comes next
			return pce.getNextTokenSuggestions(tokens)
		}

		// Check if we have a keyspace.table pattern - if so, the table name is complete
		if len(tokens) >= 4 {
			// Look for pattern: ... keyspace . table
			// Check if the token before last is a dot (excluding EOF)
			secondToLast := ""
			if len(tokens) >= 3 {
				// tokens includes EOF, so we need to check len(tokens)-3 for the actual second-to-last token
				for i := len(tokens) - 2; i >= 0; i-- {
					if tokens[i].GetTokenType() != antlr.TokenEOF {
						if i > 0 && tokens[i-1].GetTokenType() != antlr.TokenEOF {
							secondToLast = tokens[i-1].GetText()
							break
						}
					}
				}
			}

			if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
				fmt.Fprintf(debugFile, "[DEBUG] Checking for keyspace.table: secondToLast='%s'\n", secondToLast)
				defer debugFile.Close()
			}

			if secondToLast == "." {
				// We have pattern: something . table
				// This is a complete table reference in keyspace.table format
				// Treat it as complete and get suggestions for what comes next
				if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
					fmt.Fprintf(debugFile, "[DEBUG] Detected complete keyspace.table pattern, getting next suggestions\n")
					defer debugFile.Close()
				}
				return pce.getNextTokenSuggestions(tokens)
			}
		}

		// Check if the last token is an opening function call
		if strings.HasSuffix(partial, "(") {
			// This is a function call, handle it specially
			if strings.HasPrefix(strings.ToUpper(partial), "COUNT(") {
				// After COUNT(, suggest *, 1, or column names
				return []string{"*", "1"}
			}
			// For other functions, return empty for now
			return []string{}
		}

		// If we have no previous tokens, we're at the start - get top-level commands
		var suggestions []string
		if len(tokens) <= 1 {
			suggestions = pce.getTopLevelKeywords()
		} else {
			suggestions = pce.getNextTokenSuggestions(tokens[:len(tokens)-1])
		}

		// Debug: log what we got
		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] Partial token '%s', got %d suggestions before filtering\n", partial, len(suggestions))
			if len(suggestions) > 0 && len(suggestions) <= 10 {
				fmt.Fprintf(debugFile, "[DEBUG] Suggestions: %v\n", suggestions)
			}
			defer debugFile.Close()
		}

		// Filter suggestions that match the partial input
		var filtered []string
		upperPartial := strings.ToUpper(partial)
		for _, s := range suggestions {
			if strings.HasPrefix(strings.ToUpper(s), upperPartial) {
				filtered = append(filtered, s)
			}
		}

		if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			fmt.Fprintf(debugFile, "[DEBUG] After filtering for '%s' (upper: '%s'): %d matches\n", partial, upperPartial, len(filtered))
			defer debugFile.Close()
		}

		// If no matches from parser, try top-level commands as fallback
		if len(filtered) == 0 && len(tokens) <= 1 {
			for _, cmd := range pce.getTopLevelKeywords() {
				if strings.HasPrefix(strings.ToUpper(cmd), upperPartial) {
					filtered = append(filtered, cmd)
				}
			}
		}

		return filtered
	}

	return []string{}
}

// getNextTokenSuggestions returns possible next tokens based on the current token sequence
func (pce *ParserBasedCompletionEngine) getNextTokenSuggestions(tokens []antlr.Token) []string {
	if len(tokens) == 0 {
		return pce.getTopLevelKeywords()
	}

	// Build a simplified representation of the token sequence
	var tokenTypes []string
	for _, token := range tokens {
		tokenType := token.GetTokenType()

		// Map token types to keyword names
		switch tokenType {
		case grammar.CqlLexerK_SELECT:
			tokenTypes = append(tokenTypes, "SELECT")
		case grammar.CqlLexerK_INSERT:
			tokenTypes = append(tokenTypes, "INSERT")
		case grammar.CqlLexerK_UPDATE:
			tokenTypes = append(tokenTypes, "UPDATE")
		case grammar.CqlLexerK_DELETE:
			tokenTypes = append(tokenTypes, "DELETE")
		case grammar.CqlLexerK_CREATE:
			tokenTypes = append(tokenTypes, "CREATE")
		case grammar.CqlLexerK_DROP:
			tokenTypes = append(tokenTypes, "DROP")
		case grammar.CqlLexerK_ALTER:
			tokenTypes = append(tokenTypes, "ALTER")
		case grammar.CqlLexerK_DESCRIBE, grammar.CqlLexerK_DESC:
			tokenTypes = append(tokenTypes, "DESCRIBE")
		case grammar.CqlLexerK_USE:
			tokenTypes = append(tokenTypes, "USE")
		case grammar.CqlLexerK_GRANT:
			tokenTypes = append(tokenTypes, "GRANT")
		case grammar.CqlLexerK_REVOKE:
			tokenTypes = append(tokenTypes, "REVOKE")
		case grammar.CqlLexerK_FROM:
			tokenTypes = append(tokenTypes, "FROM")
		case grammar.CqlLexerK_WHERE:
			tokenTypes = append(tokenTypes, "WHERE")
		case grammar.CqlLexerK_INTO:
			tokenTypes = append(tokenTypes, "INTO")
		case grammar.CqlLexerK_VALUES:
			tokenTypes = append(tokenTypes, "VALUES")
		case grammar.CqlLexerK_SET:
			tokenTypes = append(tokenTypes, "SET")
		case grammar.CqlLexerK_TABLE:
			tokenTypes = append(tokenTypes, "TABLE")
		case grammar.CqlLexerK_KEYSPACE:
			tokenTypes = append(tokenTypes, "KEYSPACE")
		case grammar.CqlLexerSTAR:
			tokenTypes = append(tokenTypes, "*")
		case grammar.CqlLexerOBJECT_NAME:
			// For identifiers (table names, column names, etc.), use the actual text
			tokenTypes = append(tokenTypes, token.GetText())
		case grammar.CqlLexerDOT:
			tokenTypes = append(tokenTypes, ".")
		case grammar.CqlLexerK_INDEX:
			tokenTypes = append(tokenTypes, "INDEX")
		case grammar.CqlLexerK_TYPE:
			tokenTypes = append(tokenTypes, "TYPE")
		case grammar.CqlLexerK_ROLE:
			tokenTypes = append(tokenTypes, "ROLE")
		case grammar.CqlLexerK_USER:
			tokenTypes = append(tokenTypes, "USER")
		case grammar.CqlLexerK_FUNCTION:
			tokenTypes = append(tokenTypes, "FUNCTION")
		case grammar.CqlLexerK_AGGREGATE:
			tokenTypes = append(tokenTypes, "AGGREGATE")
		case grammar.CqlLexerK_MATERIALIZED:
			tokenTypes = append(tokenTypes, "MATERIALIZED")
		case grammar.CqlLexerK_VIEW:
			tokenTypes = append(tokenTypes, "VIEW")
		case grammar.CqlLexerK_TRIGGER:
			tokenTypes = append(tokenTypes, "TRIGGER")
		case grammar.CqlLexerK_CONSISTENCY:
			tokenTypes = append(tokenTypes, "CONSISTENCY")
		case grammar.CqlLexerK_ASCII:
			tokenTypes = append(tokenTypes, "ASCII")
		case grammar.CqlLexerK_BATCH:
			tokenTypes = append(tokenTypes, "BATCH")
		case grammar.CqlLexerK_KEYSPACES:
			tokenTypes = append(tokenTypes, "KEYSPACES")
		case grammar.CqlLexerK_TABLES:
			tokenTypes = append(tokenTypes, "TABLES")
		}
	}

	// Provide context-aware suggestions based on the token sequence
	if len(tokenTypes) > 0 {
		switch tokenTypes[0] {
		case "SELECT":
			return pce.getSelectSuggestions(tokenTypes)
		case "INSERT":
			return pce.getInsertSuggestions(tokenTypes)
		case "UPDATE":
			return pce.getUpdateSuggestions(tokenTypes)
		case "DELETE":
			return pce.getDeleteSuggestions(tokenTypes)
		case "CREATE":
			return pce.getCreateSuggestions(tokenTypes)
		case "DROP":
			return pce.getDropSuggestions(tokenTypes)
		case "ALTER":
			return pce.getAlterSuggestions(tokenTypes)
		case "DESCRIBE":
			return pce.getDescribeSuggestions(tokenTypes)
		case "USE":
			return pce.getUseSuggestions(tokenTypes)
		case "GRANT":
			return pce.getGrantSuggestions(tokenTypes)
		case "REVOKE":
			return pce.getRevokeSuggestions(tokenTypes)
		case "CONSISTENCY":
			return pce.getConsistencySuggestions(tokenTypes)
		case "OUTPUT":
			return pce.getOutputSuggestions(tokenTypes)
		case "COPY":
			return pce.getCopySuggestions(tokenTypes)
		}
	}

	return []string{}
}
