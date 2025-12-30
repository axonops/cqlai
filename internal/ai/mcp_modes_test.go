package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsQueryAllowed tests the query permission logic for all confirmation modes
func TestIsQueryAllowed(t *testing.T) {
	tests := []struct {
		name               string
		mode               string
		operation          string
		isDangerous        bool
		severity           string
		wantAllowed        bool
		wantConfirmation   bool
	}{
		// READONLY mode
		{
			name:             "readonly - SELECT allowed",
			mode:             "readonly",
			operation:        "SELECT",
			isDangerous:      false,
			severity:         "SAFE",
			wantAllowed:      true,
			wantConfirmation: false,
		},
		{
			name:             "readonly - DESCRIBE allowed",
			mode:             "readonly",
			operation:        "DESCRIBE",
			isDangerous:      false,
			severity:         "SAFE",
			wantAllowed:      true,
			wantConfirmation: false,
		},
		{
			name:             "readonly - INSERT blocked",
			mode:             "readonly",
			operation:        "INSERT",
			isDangerous:      false,
			severity:         "SAFE",
			wantAllowed:      false,
			wantConfirmation: false,
		},
		{
			name:             "readonly - DELETE blocked",
			mode:             "readonly",
			operation:        "DELETE",
			isDangerous:      true,
			severity:         "HIGH",
			wantAllowed:      false,
			wantConfirmation: false,
		},
		{
			name:             "readonly - DROP blocked",
			mode:             "readonly",
			operation:        "DROP",
			isDangerous:      true,
			severity:         "CRITICAL",
			wantAllowed:      false,
			wantConfirmation: false,
		},

		// READ_WRITE mode
		{
			name:             "read_write - SELECT allowed",
			mode:             "read_write",
			operation:        "SELECT",
			isDangerous:      false,
			severity:         "SAFE",
			wantAllowed:      true,
			wantConfirmation: false,
		},
		{
			name:             "read_write - INSERT allowed",
			mode:             "read_write",
			operation:        "INSERT",
			isDangerous:      false,
			severity:         "SAFE",
			wantAllowed:      true,
			wantConfirmation: false,
		},
		{
			name:             "read_write - UPDATE allowed",
			mode:             "read_write",
			operation:        "UPDATE",
			isDangerous:      true,
			severity:         "HIGH",
			wantAllowed:      true,
			wantConfirmation: false,
		},
		{
			name:             "read_write - DELETE allowed (no confirmation)",
			mode:             "read_write",
			operation:        "DELETE",
			isDangerous:      true,
			severity:         "HIGH",
			wantAllowed:      true,
			wantConfirmation: false,
		},
		{
			name:             "read_write - DROP requires confirmation",
			mode:             "read_write",
			operation:        "DROP",
			isDangerous:      true,
			severity:         "CRITICAL",
			wantAllowed:      true,
			wantConfirmation: true,
		},
		{
			name:             "read_write - TRUNCATE requires confirmation",
			mode:             "read_write",
			operation:        "TRUNCATE",
			isDangerous:      true,
			severity:         "CRITICAL",
			wantAllowed:      true,
			wantConfirmation: true,
		},
		{
			name:             "read_write - CREATE blocked",
			mode:             "read_write",
			operation:        "CREATE",
			isDangerous:      true,
			severity:         "CRITICAL",
			wantAllowed:      false,
			wantConfirmation: false,
		},
		{
			name:             "read_write - ALTER blocked",
			mode:             "read_write",
			operation:        "ALTER",
			isDangerous:      true,
			severity:         "CRITICAL",
			wantAllowed:      false,
			wantConfirmation: false,
		},
		{
			name:             "read_write - GRANT blocked",
			mode:             "read_write",
			operation:        "GRANT",
			isDangerous:      true,
			severity:         "CRITICAL",
			wantAllowed:      false,
			wantConfirmation: false,
		},

		// DANGEROUS_ONLY mode (default)
		{
			name:             "dangerous_only - SELECT allowed no confirmation",
			mode:             "dangerous_only",
			operation:        "SELECT",
			isDangerous:      false,
			severity:         "SAFE",
			wantAllowed:      true,
			wantConfirmation: false,
		},
		{
			name:             "dangerous_only - DELETE requires confirmation",
			mode:             "dangerous_only",
			operation:        "DELETE",
			isDangerous:      true,
			severity:         "HIGH",
			wantAllowed:      true,
			wantConfirmation: true,
		},
		{
			name:             "dangerous_only - DROP requires confirmation",
			mode:             "dangerous_only",
			operation:        "DROP",
			isDangerous:      true,
			severity:         "CRITICAL",
			wantAllowed:      true,
			wantConfirmation: true,
		},
		{
			name:             "dangerous_only - INSERT allowed no confirmation",
			mode:             "dangerous_only",
			operation:        "INSERT",
			isDangerous:      false,
			severity:         "SAFE",
			wantAllowed:      true,
			wantConfirmation: false,
		},

		// ALL mode
		{
			name:             "all - SELECT requires confirmation",
			mode:             "all",
			operation:        "SELECT",
			isDangerous:      false,
			severity:         "SAFE",
			wantAllowed:      true,
			wantConfirmation: true,
		},
		{
			name:             "all - DELETE requires confirmation",
			mode:             "all",
			operation:        "DELETE",
			isDangerous:      true,
			severity:         "HIGH",
			wantAllowed:      true,
			wantConfirmation: true,
		},
		{
			name:             "all - DROP requires confirmation",
			mode:             "all",
			operation:        "DROP",
			isDangerous:      true,
			severity:         "CRITICAL",
			wantAllowed:      true,
			wantConfirmation: true,
		},

		// NONE mode
		{
			name:             "none - SELECT no confirmation",
			mode:             "none",
			operation:        "SELECT",
			isDangerous:      false,
			severity:         "SAFE",
			wantAllowed:      true,
			wantConfirmation: false,
		},
		{
			name:             "none - DELETE no confirmation",
			mode:             "none",
			operation:        "DELETE",
			isDangerous:      true,
			severity:         "HIGH",
			wantAllowed:      true,
			wantConfirmation: false,
		},
		{
			name:             "none - DROP no confirmation",
			mode:             "none",
			operation:        "DROP",
			isDangerous:      true,
			severity:         "CRITICAL",
			wantAllowed:      true,
			wantConfirmation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock MCP server with specific mode
			server := &MCPServer{
				config: &MCPServerConfig{
					ConfirmationMode: tt.mode,
				},
			}

			classification := QueryClassification{
				IsDangerous: tt.isDangerous,
				Severity:    tt.severity,
				Operation:   tt.operation,
			}

			allowed, needsConfirmation := server.isQueryAllowed(tt.operation, classification)

			assert.Equal(t, tt.wantAllowed, allowed,
				"Query allowed mismatch for %s in %s mode", tt.operation, tt.mode)
			assert.Equal(t, tt.wantConfirmation, needsConfirmation,
				"Confirmation needed mismatch for %s in %s mode", tt.operation, tt.mode)
		})
	}
}

// TestGetConfirmationModeDescription tests mode descriptions
func TestGetConfirmationModeDescription(t *testing.T) {
	tests := []struct {
		mode     string
		contains string
	}{
		{"readonly", "SELECT and DESCRIBE"},
		{"read_write", "DELETE allowed"},
		{"dangerous_only", "default"},
		{"all", "most restrictive"},
		{"none", "NOT RECOMMENDED"},
		{"unknown_mode", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			desc := GetConfirmationModeDescription(tt.mode)
			assert.Contains(t, desc, tt.contains,
				"Description for mode %s should contain %q", tt.mode, tt.contains)
		})
	}
}

// TestBuildCQLFromParams tests CQL query building from parameters
func TestBuildCQLFromParams(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]any
		expected string
	}{
		{
			name: "SELECT with columns",
			params: map[string]any{
				"operation": "SELECT",
				"keyspace":  "test_mcp",
				"table":     "users",
				"columns":   []interface{}{"email", "name"},
			},
			expected: "SELECT email, name FROM test_mcp.users",
		},
		{
			name: "SELECT all columns",
			params: map[string]any{
				"operation": "SELECT",
				"keyspace":  "test_mcp",
				"table":     "users",
				"columns":   []interface{}{"*"},
			},
			expected: "SELECT * FROM test_mcp.users",
		},
		{
			name: "DELETE",
			params: map[string]any{
				"operation": "DELETE",
				"keyspace":  "test_mcp",
				"table":     "events",
			},
			expected: "DELETE FROM test_mcp.events",
		},
		{
			name: "DROP TABLE",
			params: map[string]any{
				"operation": "DROP",
				"keyspace":  "test_mcp",
				"table":     "orders",
			},
			expected: "DROP TABLE test_mcp.orders",
		},
		{
			name: "DROP KEYSPACE",
			params: map[string]any{
				"operation": "DROP",
				"keyspace":  "test_mcp",
			},
			expected: "DROP KEYSPACE test_mcp",
		},
		{
			name: "TRUNCATE",
			params: map[string]any{
				"operation": "TRUNCATE",
				"keyspace":  "test_mcp",
				"table":     "users",
			},
			expected: "TRUNCATE test_mcp.users",
		},
		{
			name: "INSERT",
			params: map[string]any{
				"operation": "INSERT",
				"keyspace":  "test_mcp",
				"table":     "users",
			},
			expected: "INSERT INTO test_mcp.users",
		},
		{
			name: "UPDATE",
			params: map[string]any{
				"operation": "UPDATE",
				"keyspace":  "test_mcp",
				"table":     "users",
			},
			expected: "UPDATE test_mcp.users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCQLFromParams(tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}
}
