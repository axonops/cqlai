package splitter_test

import (
	"reflect"
	"testing"

	"github.com/axonops/cqlai/internal/batch"
)

func TestSplitForNode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		// Basic splitting (semicolons are preserved in output)
		{
			name:     "basic multi-statement",
			input:    "SELECT * FROM t1; SELECT * FROM t2;",
			expected: []string{"SELECT * FROM t1;", "SELECT * FROM t2;"},
		},
		{
			name:     "single statement with semicolon",
			input:    "SELECT * FROM t1;",
			expected: []string{"SELECT * FROM t1;"},
		},
		{
			name:     "single statement no trailing semicolon",
			input:    "SELECT * FROM t1",
			expected: []string{"SELECT * FROM t1"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   \t\n  ",
			expected: nil,
		},

		// Semicolons in strings (should NOT split on these)
		{
			name:     "semicolon in single quotes",
			input:    "INSERT INTO t (a) VALUES ('hello; world');",
			expected: []string{"INSERT INTO t (a) VALUES ('hello; world');"},
		},
		{
			name:     "semicolon in double quotes",
			input:    `SELECT * FROM "table;name";`,
			expected: []string{`SELECT * FROM "table;name";`},
		},
		{
			name:     "multiple semicolons in string",
			input:    "INSERT INTO t (a) VALUES ('a;b;c;d');",
			expected: []string{"INSERT INTO t (a) VALUES ('a;b;c;d');"},
		},
		{
			name:     "escaped quotes in string",
			input:    "INSERT INTO t (a) VALUES ('it''s; fine');",
			expected: []string{"INSERT INTO t (a) VALUES ('it''s; fine');"},
		},

		// Dollar-quoted strings
		{
			name:     "dollar string with semicolon",
			input:    "INSERT INTO t (a) VALUES ($$test;test$$);",
			expected: []string{"INSERT INTO t (a) VALUES ($$test;test$$);"},
		},
		{
			name:     "dollar string with multiple semicolons",
			input:    "INSERT INTO t (a) VALUES ($$a;b;c$$);",
			expected: []string{"INSERT INTO t (a) VALUES ($$a;b;c$$);"},
		},

		// Comments (semicolons in comments should be ignored)
		{
			name:     "line comment with semicolon",
			input:    "SELECT * FROM t1; -- comment; here\nSELECT * FROM t2;",
			expected: []string{"SELECT * FROM t1;", "SELECT * FROM t2;"},
		},
		{
			name:     "block comment with semicolon",
			input:    "SELECT * FROM t1; /* comment; here */ SELECT * FROM t2;",
			expected: []string{"SELECT * FROM t1;", "SELECT * FROM t2;"},
		},
		{
			name:     "double-dash comment",
			input:    "SELECT * FROM t1; -- ignore this\nSELECT * FROM t2;",
			expected: []string{"SELECT * FROM t1;", "SELECT * FROM t2;"},
		},
		{
			name:     "double-slash comment",
			input:    "SELECT * FROM t1; // ignore this\nSELECT * FROM t2;",
			expected: []string{"SELECT * FROM t1;", "SELECT * FROM t2;"},
		},

		// BATCH grouping (entire batch as single statement)
		{
			name:  "simple batch",
			input: "BEGIN BATCH INSERT INTO t (a) VALUES (1); INSERT INTO t (a) VALUES (2); APPLY BATCH;",
			expected: []string{
				"BEGIN BATCH INSERT INTO t (a) VALUES (1); INSERT INTO t (a) VALUES (2); APPLY BATCH;",
			},
		},
		{
			name:  "unlogged batch",
			input: "BEGIN UNLOGGED BATCH INSERT INTO t (a) VALUES (1); APPLY BATCH;",
			expected: []string{
				"BEGIN UNLOGGED BATCH INSERT INTO t (a) VALUES (1); APPLY BATCH;",
			},
		},
		{
			name:  "counter batch",
			input: "BEGIN COUNTER BATCH UPDATE t SET c = c + 1 WHERE id = 1; APPLY BATCH;",
			expected: []string{
				"BEGIN COUNTER BATCH UPDATE t SET c = c + 1 WHERE id = 1; APPLY BATCH;",
			},
		},
		{
			name:  "batch with timestamp",
			input: "BEGIN BATCH USING TIMESTAMP 12345 INSERT INTO t (a) VALUES (1); APPLY BATCH;",
			expected: []string{
				"BEGIN BATCH USING TIMESTAMP 12345 INSERT INTO t (a) VALUES (1); APPLY BATCH;",
			},
		},
		{
			name:  "statements before and after batch",
			input: "SELECT * FROM t1; BEGIN BATCH INSERT INTO t (a) VALUES (1); APPLY BATCH; SELECT * FROM t2;",
			expected: []string{
				"SELECT * FROM t1;",
				"BEGIN BATCH INSERT INTO t (a) VALUES (1); APPLY BATCH;",
				"SELECT * FROM t2;",
			},
		},

		// Shell commands (newline-terminated, semicolons stripped)
		{
			name:     "describe command",
			input:    "DESCRIBE keyspaces\nSELECT * FROM t1;",
			expected: []string{"DESCRIBE keyspaces", "SELECT * FROM t1;"},
		},
		{
			name:     "help command",
			input:    "HELP\nSELECT * FROM t1;",
			expected: []string{"HELP", "SELECT * FROM t1;"},
		},
		{
			name:     "show command",
			input:    "SHOW VERSION\nSELECT * FROM t1;",
			expected: []string{"SHOW VERSION", "SELECT * FROM t1;"},
		},

		// Edge cases
		{
			name:     "multiple spaces between statements",
			input:    "SELECT * FROM t1;    SELECT * FROM t2;",
			expected: []string{"SELECT * FROM t1;", "SELECT * FROM t2;"},
		},
		{
			name:     "newlines between statements",
			input:    "SELECT * FROM t1;\n\n\nSELECT * FROM t2;",
			expected: []string{"SELECT * FROM t1;", "SELECT * FROM t2;"},
		},
		{
			name:     "mixed whitespace",
			input:    "SELECT * FROM t1;\t\n  SELECT * FROM t2;",
			expected: []string{"SELECT * FROM t1;", "SELECT * FROM t2;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := batch.SplitForNode(tt.input)
			if err != nil {
				t.Fatalf("SplitForNode() error = %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SplitForNode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSplitStatements(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantIncomplete  bool
		wantIdentifiers []string
	}{
		{
			name:            "complete statements",
			input:           "SELECT * FROM t1; INSERT INTO t2 (a) VALUES (1);",
			wantIncomplete:  false,
			wantIdentifiers: []string{"SELECT", "INSERT"},
		},
		{
			name:            "unclosed single quote",
			input:           "SELECT * FROM t1 WHERE a = 'unclosed",
			wantIncomplete:  true,
			wantIdentifiers: []string{"SELECT"},
		},
		{
			name:            "unclosed double quote",
			input:           `SELECT * FROM "unclosed`,
			wantIncomplete:  true,
			wantIdentifiers: []string{"SELECT"},
		},
		{
			name:            "unclosed dollar string",
			input:           "SELECT * FROM t1 WHERE a = $$unclosed",
			wantIncomplete:  true,
			wantIdentifiers: []string{"SELECT"},
		},
		{
			name:            "unclosed block comment",
			input:           "SELECT * FROM t1 /* unclosed comment",
			wantIncomplete:  true,
			wantIdentifiers: []string{"SELECT"},
		},
		{
			name:            "incomplete batch",
			input:           "BEGIN BATCH INSERT INTO t (a) VALUES (1);",
			wantIncomplete:  true,
			wantIdentifiers: []string{"BEGIN"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := batch.SplitStatements(tt.input)
			if err != nil {
				t.Fatalf("SplitStatements() error = %v", err)
			}
			if result.Incomplete != tt.wantIncomplete {
				t.Errorf("SplitStatements().Incomplete = %v, want %v", result.Incomplete, tt.wantIncomplete)
			}
			if !reflect.DeepEqual(result.Identifiers, tt.wantIdentifiers) {
				t.Errorf("SplitStatements().Identifiers = %v, want %v", result.Identifiers, tt.wantIdentifiers)
			}
		})
	}
}

func TestLex(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTypes []batch.TokenType
		wantErr   bool
	}{
		{
			name:      "simple select",
			input:     "SELECT * FROM t1",
			wantTypes: []batch.TokenType{batch.TokenIdentifier, batch.TokenStar, batch.TokenIdentifier, batch.TokenIdentifier},
			wantErr:   false,
		},
		{
			name:      "string literal",
			input:     "'hello world'",
			wantTypes: []batch.TokenType{batch.TokenQuotedStringLiteral},
			wantErr:   false,
		},
		{
			name:      "quoted identifier",
			input:     `"column name"`,
			wantTypes: []batch.TokenType{batch.TokenQuotedName},
			wantErr:   false,
		},
		{
			name:      "dollar string",
			input:     "$$code block$$",
			wantTypes: []batch.TokenType{batch.TokenPgStringLiteral},
			wantErr:   false,
		},
		{
			name:      "uuid",
			input:     "550e8400-e29b-41d4-a716-446655440000",
			wantTypes: []batch.TokenType{batch.TokenUUID},
			wantErr:   false,
		},
		{
			name:      "blob literal",
			input:     "0xDEADBEEF",
			wantTypes: []batch.TokenType{batch.TokenBlobLiteral},
			wantErr:   false,
		},
		{
			name:      "numbers",
			input:     "123 45.67",
			wantTypes: []batch.TokenType{batch.TokenWholenumber, batch.TokenFloat},
			wantErr:   false,
		},
		{
			name:      "semicolon",
			input:     ";",
			wantTypes: []batch.TokenType{batch.TokenEndtoken},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := batch.Lex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Lex() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			gotTypes := make([]batch.TokenType, len(tokens))
			for i, tok := range tokens {
				gotTypes[i] = tok.Type
			}
			if !reflect.DeepEqual(gotTypes, tt.wantTypes) {
				t.Errorf("Lex() types = %v, want %v", gotTypes, tt.wantTypes)
			}
		})
	}
}

func TestIsShellCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"describe", true},
		{"DESCRIBE", true},
		{"desc", true},
		{"help", true},
		{"?", true},
		{"show", true},
		{"consistency", true},
		{"tracing", true},
		{"exit", true},
		{"quit", true},
		{"SELECT", false},
		{"INSERT", false},
		{"UPDATE", false},
		{"DELETE", false},
		{"CREATE", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if got := batch.IsShellCommand(tt.cmd); got != tt.want {
				t.Errorf("IsShellCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestMassageTokens(t *testing.T) {
	// Test that DESCRIBE command gets newline converted to endtoken
	input := "DESCRIBE keyspaces\nSELECT * FROM t1"
	tokens, err := batch.Lex(input)
	if err != nil {
		t.Fatalf("Lex() error = %v", err)
	}

	massaged := batch.MassageTokens(tokens)

	// Find the first endtoken - should be after DESCRIBE keyspaces
	foundEndtoken := false
	endtokenIndex := -1
	for i, tok := range massaged {
		if tok.Type == batch.TokenEndtoken {
			foundEndtoken = true
			endtokenIndex = i
			break
		}
	}

	if !foundEndtoken {
		t.Error("MassageTokens() did not convert newline to endtoken for DESCRIBE command")
	}

	// The endtoken should come after "DESCRIBE" and "keyspaces"
	if endtokenIndex < 2 {
		t.Errorf("MassageTokens() endtoken at wrong position: %d", endtokenIndex)
	}
}

// TestBatchDetectionRobustness verifies that APPLY alone doesn't terminate a batch
func TestBatchDetectionRobustness(t *testing.T) {
	// This input has "APPLY BATCH" which should correctly terminate the batch
	input := "BEGIN BATCH INSERT INTO t (a) VALUES (1); UPDATE t SET a = 2; APPLY BATCH;"
	result, err := batch.SplitForNode(input)
	if err != nil {
		t.Fatalf("SplitForNode() error = %v", err)
	}

	// Should be a single statement (the whole batch)
	if len(result) != 1 {
		t.Errorf("SplitForNode() returned %d statements, want 1", len(result))
	}
}
