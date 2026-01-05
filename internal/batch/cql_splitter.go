package batch

import (
	"fmt"
	"regexp"
	"strings"
)

/**
 * CQL Statement Splitter - Go port from Python cqlsh
 *
 * Ported from:
 * - cqlshlib/cql3handling.py (syntax_rules / terminal definitions)
 * - cqlshlib/cqlhandling.py (cql_massage_tokens, cql_split_statements)
 * - cqlshlib/cqlshhandling.py (commands_end_with_newline)
 */

// TokenType represents the type of a CQL token
type TokenType int

const (
	TokenEndline TokenType = iota
	TokenEndtoken
	TokenIdentifier
	TokenQuotedStringLiteral
	TokenQuotedName
	TokenPgStringLiteral
	TokenUnclosedString
	TokenUnclosedName
	TokenUnclosedPgString
	TokenUnclosedComment
	TokenFloat
	TokenUUID
	TokenBlobLiteral
	TokenWholenumber
	TokenColon
	TokenStar
	TokenOp
	TokenCmp
	TokenBrackets
	TokenJunk // For patterns we want to skip (whitespace, comments)
)

// Token represents a lexed CQL token
type Token struct {
	Type  TokenType
	Value string
	Start int
	End   int
}

// terminalPattern represents a regex pattern for tokenizing
type terminalPattern struct {
	tokenType TokenType // -1 for JUNK (discard)
	pattern   *regexp.Regexp
}

// Commands that terminate on newline instead of semicolon
var commandsEndWithNewline = map[string]bool{
	"help":        true,
	"?":           true,
	"consistency": true,
	"serial":      true,
	"describe":    true,
	"desc":        true,
	"show":        true,
	"source":      true,
	"capture":     true,
	"login":       true,
	"debug":       true,
	"tracing":     true,
	"expand":      true,
	"elapsed":     true,
	"paging":      true,
	"exit":        true,
	"quit":        true,
	"clear":       true,
	"cls":         true,
	"history":     true,
}

// Terminal patterns - order matters (most specific first)
// Note: $$ string patterns are handled separately in Lex() since Go's regexp
// doesn't support negative lookahead (?!...)
var terminalPatterns = []terminalPattern{
	// Endline
	{TokenEndline, regexp.MustCompile(`^\n`)},

	// JUNK: whitespace, line comments, block comments (discard these)
	{TokenJunk, regexp.MustCompile(`^[ \t\r\f\v]+`)},
	{TokenJunk, regexp.MustCompile(`^(--|//)[^\n\r]*`)},
	{TokenJunk, regexp.MustCompile(`^/\*[\s\S]*?\*/`)},

	// Quoted string literals '...'
	{TokenQuotedStringLiteral, regexp.MustCompile(`^'([^']|'')*'`)},

	// Quoted names "..."
	{TokenQuotedName, regexp.MustCompile(`^"([^"]|"")*"`)},

	// Unclosed tokens (for detecting incomplete input)
	{TokenUnclosedString, regexp.MustCompile(`^'([^']|'')*$`)},
	{TokenUnclosedName, regexp.MustCompile(`^"([^"]|"")*$`)},
	{TokenUnclosedComment, regexp.MustCompile(`^/\*[\s\S]*$`)},

	// Numbers and literals
	{TokenFloat, regexp.MustCompile(`^-?[0-9]+\.[0-9]+`)},
	{TokenUUID, regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)},
	{TokenBlobLiteral, regexp.MustCompile(`(?i)^0x[0-9a-f]+`)},
	{TokenWholenumber, regexp.MustCompile(`^[0-9]+`)},

	// Identifiers
	{TokenIdentifier, regexp.MustCompile(`(?i)^[a-z][a-z0-9_]*`)},

	// Operators and punctuation
	{TokenEndtoken, regexp.MustCompile(`^;`)},
	{TokenColon, regexp.MustCompile(`^:`)},
	{TokenStar, regexp.MustCompile(`^\*`)},
	{TokenOp, regexp.MustCompile(`^[-+=%/,().]`)},
	{TokenCmp, regexp.MustCompile(`^[<>!]=?`)},
	{TokenBrackets, regexp.MustCompile(`^[\[\]{}]`)},
}

// Lex tokenizes CQL input text
func Lex(text string) ([]Token, error) {
	var tokens []Token
	pos := 0

	for pos < len(text) {
		matched := false

		// Handle $$ string literals manually (Go regexp doesn't support negative lookahead)
		if strings.HasPrefix(text[pos:], "$$") {
			// Find closing $$
			closeIdx := strings.Index(text[pos+2:], "$$")
			if closeIdx >= 0 {
				// Found closing $$ - complete string literal
				value := text[pos : pos+2+closeIdx+2]
				tokens = append(tokens, Token{
					Type:  TokenPgStringLiteral,
					Value: value,
					Start: pos,
					End:   pos + len(value),
				})
				pos += len(value)
				matched = true
			} else {
				// No closing $$ - unclosed string
				value := text[pos:]
				tokens = append(tokens, Token{
					Type:  TokenUnclosedPgString,
					Value: value,
					Start: pos,
					End:   pos + len(value),
				})
				pos += len(value)
				matched = true
			}
		}

		if !matched {
			for _, tp := range terminalPatterns {
				loc := tp.pattern.FindStringIndex(text[pos:])
				if loc != nil && loc[0] == 0 {
					value := text[pos : pos+loc[1]]

					// Only add non-JUNK tokens
					if tp.tokenType != TokenJunk {
						tokens = append(tokens, Token{
							Type:  tp.tokenType,
							Value: value,
							Start: pos,
							End:   pos + len(value),
						})
					}

					pos += len(value)
					matched = true
					break
				}
			}
		}

		if !matched {
			// Return error with context
			end := pos + 20
			if end > len(text) {
				end = len(text)
			}
			return nil, fmt.Errorf("cannot lex at position %d: %q", pos, text[pos:end])
		}
	}

	return tokens, nil
}

// MassageTokens converts newlines to endtokens for special commands
func MassageTokens(tokens []Token) []Token {
	var output []Token
	var curstmt []Token
	termOnNL := false

	for _, t := range tokens {
		token := t

		if token.Type == TokenEndline {
			if termOnNL {
				// Convert endline to endtoken for newline-terminated commands
				token.Type = TokenEndtoken
			} else {
				// Don't put any 'endline' tokens in output
				continue
			}
		}

		curstmt = append(curstmt, token)

		if token.Type == TokenEndtoken {
			termOnNL = false
			output = append(output, curstmt...)
			curstmt = nil
		} else if len(curstmt) == 1 {
			// First token in statement; command word
			cmd := strings.ToLower(token.Value)
			termOnNL = commandsEndWithNewline[cmd]
		}
	}

	output = append(output, curstmt...)
	return output
}

// hasUnclosedToken checks if tokens contain any unclosed token types
func hasUnclosedToken(tokens []Token) bool {
	for _, t := range tokens {
		switch t.Type {
		case TokenUnclosedString, TokenUnclosedName, TokenUnclosedPgString, TokenUnclosedComment:
			return true
		}
	}
	return false
}

// SplitResult contains the result of splitting CQL statements
type SplitResult struct {
	Statements  [][]Token // Each statement as array of tokens
	Incomplete  bool      // True if input is incomplete
	SourceText  string    // Original source text
	Identifiers []string  // First identifier of each statement (SELECT, INSERT, etc.)
	ExtraTokens []string  // 2nd and 3rd meaningful tokens from first statement
}

// SplitStatements splits CQL input into individual statements
func SplitStatements(text string) (*SplitResult, error) {
	tokens, err := Lex(text)
	if err != nil {
		return nil, err
	}

	massaged := MassageTokens(tokens)

	// Check for unclosed tokens
	inPgString := hasUnclosedToken(massaged)

	// Split on endtoken
	var stmts [][]Token
	var current []Token

	for _, t := range massaged {
		current = append(current, t)
		if t.Type == TokenEndtoken {
			stmts = append(stmts, current)
			current = nil
		}
	}
	if len(current) > 0 {
		stmts = append(stmts, current)
	}

	// Handle BATCH grouping
	var output [][]Token
	inBatch := false

	for _, stmt := range stmts {
		if len(stmt) == 0 {
			continue
		}

		if inBatch {
			// Append to previous statement (we're inside a BATCH)
			if len(output) > 0 {
				output[len(output)-1] = append(output[len(output)-1], stmt...)
			} else {
				output = append(output, stmt)
			}
		} else {
			output = append(output, stmt)
		}

		// Check for BATCH start/end
		if len(stmt) >= 3 {
			// Check for APPLY BATCH at end (position -3 from end, before endtoken)
			if strings.ToUpper(stmt[len(stmt)-3].Value) == "APPLY" {
				inBatch = false
			} else if strings.ToUpper(stmt[0].Value) == "BEGIN" {
				inBatch = true
			}
		} else if len(stmt) >= 1 && strings.ToUpper(stmt[0].Value) == "BEGIN" {
			inBatch = true
		}
	}

	result := &SplitResult{
		Statements:  output,
		Incomplete:  inBatch || inPgString,
		SourceText:  text,
		Identifiers: make([]string, 0),
		ExtraTokens: make([]string, 0),
	}

	// Extract identifiers and extra tokens
	for i, stmt := range output {
		// Find first identifier
		for _, t := range stmt {
			if t.Type == TokenIdentifier {
				result.Identifiers = append(result.Identifiers, strings.ToUpper(t.Value))
				break
			}
		}

		// For first statement, extract 2nd and 3rd meaningful tokens
		if i == 0 {
			meaningfulTokens := getMeaningfulTokens(stmt)
			if len(meaningfulTokens) > 1 {
				result.ExtraTokens = append(result.ExtraTokens, meaningfulTokens[1].Value)
			}
			if len(meaningfulTokens) > 2 {
				result.ExtraTokens = append(result.ExtraTokens, meaningfulTokens[2].Value)
			}
		}
	}

	return result, nil
}

// getMeaningfulTokens filters tokens to only meaningful ones (identifiers, literals, etc.)
func getMeaningfulTokens(tokens []Token) []Token {
	var meaningful []Token
	for _, t := range tokens {
		switch t.Type {
		case TokenIdentifier, TokenStar, TokenQuotedName, TokenQuotedStringLiteral,
			TokenPgStringLiteral, TokenWholenumber, TokenFloat, TokenUUID, TokenBlobLiteral:
			meaningful = append(meaningful, t)
		}
	}
	return meaningful
}

// ExtractStatementText extracts the original text for a token slice
func (sr *SplitResult) ExtractStatementText(tokens []Token) string {
	if len(tokens) == 0 {
		return ""
	}
	return sr.SourceText[tokens[0].Start:tokens[len(tokens)-1].End]
}

// GetStatementStrings returns the statements as strings
func (sr *SplitResult) GetStatementStrings() []string {
	var stmts []string
	for _, tokens := range sr.Statements {
		text := strings.TrimSpace(sr.ExtractStatementText(tokens))
		if text == "" {
			continue
		}

		// For shell commands, remove trailing semicolon if present
		if len(tokens) > 0 {
			firstWord := strings.ToLower(tokens[0].Value)
			if commandsEndWithNewline[firstWord] && strings.HasSuffix(text, ";") {
				text = strings.TrimSpace(text[:len(text)-1])
			}
		}

		if text != "" {
			stmts = append(stmts, text)
		}
	}
	return stmts
}

// Split is the main entry point - splits CQL input into statement strings
func Split(text string) ([]string, error) {
	// Handle empty input
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	result, err := SplitStatements(text)
	if err != nil {
		return nil, err
	}

	return result.GetStatementStrings(), nil
}

// IsShellCommand checks if a command is a shell command that terminates on newline
func IsShellCommand(cmd string) bool {
	return commandsEndWithNewline[strings.ToLower(cmd)]
}
