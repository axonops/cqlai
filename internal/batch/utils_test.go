package batch

import (
	"testing"
)

func TestStripComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no comments",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "line comment at end",
			input:    "SELECT * FROM users -- this is a comment",
			expected: "SELECT * FROM users ",
		},
		{
			name:     "double slash comment",
			input:    "SELECT * FROM users // comment",
			expected: "SELECT * FROM users ",
		},
		{
			name:     "block comment",
			input:    "SELECT /* comment */ * FROM users",
			expected: "SELECT  * FROM users",
		},
		{
			name:     "dash inside single quotes - should NOT strip",
			input:    "SELECT * FROM users WHERE name = 'hello--world'",
			expected: "SELECT * FROM users WHERE name = 'hello--world'",
		},
		{
			name:     "URL inside single quotes - should NOT strip",
			input:    "INSERT INTO urls (url) VALUES ('http://example.com')",
			expected: "INSERT INTO urls (url) VALUES ('http://example.com')",
		},
		{
			name:     "block comment markers inside quotes - should NOT strip",
			input:    "SELECT * FROM logs WHERE msg = 'code /* here */ end'",
			expected: "SELECT * FROM logs WHERE msg = 'code /* here */ end'",
		},
		{
			name:     "escaped quote inside string",
			input:    "INSERT INTO users (name) VALUES ('O''Brien')",
			expected: "INSERT INTO users (name) VALUES ('O''Brien')",
		},
		{
			name:     "comment after quoted string with dashes",
			input:    "SELECT 'hello--world' FROM t -- real comment",
			expected: "SELECT 'hello--world' FROM t ",
		},
		{
			name:     "multiline with comments",
			input:    "SELECT *\n-- comment line\nFROM users",
			expected: "SELECT *\n\nFROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripComments(tt.input)
			if result != tt.expected {
				t.Errorf("stripComments(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single statement",
			input:    "SELECT * FROM users;",
			expected: []string{"SELECT * FROM users;"},
		},
		{
			name:     "two statements",
			input:    "SELECT * FROM users; SELECT * FROM orders;",
			expected: []string{"SELECT * FROM users;", "SELECT * FROM orders;"},
		},
		{
			name:     "semicolon inside quotes - should NOT split",
			input:    "INSERT INTO users (bio) VALUES ('Hello; I am here');",
			expected: []string{"INSERT INTO users (bio) VALUES ('Hello; I am here');"},
		},
		{
			name:     "multiple semicolons inside quotes",
			input:    "INSERT INTO t (x) VALUES ('a;b;c'); SELECT 1;",
			expected: []string{"INSERT INTO t (x) VALUES ('a;b;c');", "SELECT 1;"},
		},
		{
			name:     "escaped quotes with semicolon",
			input:    "INSERT INTO t (x) VALUES ('it''s; here'); SELECT 1;",
			expected: []string{"INSERT INTO t (x) VALUES ('it''s; here');", "SELECT 1;"},
		},
		{
			name:     "batch statement",
			input:    "BEGIN BATCH INSERT INTO t (id) VALUES (1); INSERT INTO t (id) VALUES (2); APPLY BATCH;",
			expected: []string{"BEGIN BATCH INSERT INTO t (id) VALUES (1); INSERT INTO t (id) VALUES (2); APPLY BATCH;"},
		},
		{
			name:     "no trailing semicolon",
			input:    "SELECT * FROM users",
			expected: []string{"SELECT * FROM users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitStatements(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitStatements(%q) returned %d statements, want %d\nGot: %v",
					tt.input, len(result), len(tt.expected), result)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitStatements(%q)[%d] = %q, want %q",
						tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}
