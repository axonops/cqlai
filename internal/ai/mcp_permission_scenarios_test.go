package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPermissionMatrix_ReadonlyMode tests all operations in readonly mode
func TestPermissionMatrix_ReadonlyMode(t *testing.T) {
	config := &MCPServerConfig{
		Mode:       ConfigModePreset,
		PresetMode: "readonly",
	}
	server := &MCPServer{config: config}

	tests := []struct {
		operation       string
		command         string
		expectAllowed   bool
		expectConfirm   bool
	}{
		// DQL - should be allowed
		{"SELECT", "SELECT * FROM users", true, false},
		{"LIST ROLES", "LIST ROLES", true, false},
		{"DESCRIBE", "DESCRIBE TABLES", true, false},
		{"SHOW", "SHOW VERSION", true, false},

		// SESSION - always allowed
		{"CONSISTENCY", "CONSISTENCY QUORUM", true, false},
		{"PAGING", "PAGING 100", true, false},

		// DML - should be blocked
		{"INSERT", "INSERT INTO users VALUES (...)", false, false},
		{"UPDATE", "UPDATE users SET name=?", false, false},
		{"DELETE", "DELETE FROM users", false, false},

		// DDL - should be blocked
		{"CREATE TABLE", "CREATE TABLE users (...)", false, false},
		{"ALTER TABLE", "ALTER TABLE users ADD col", false, false},
		{"DROP TABLE", "DROP TABLE users", false, false},
		{"TRUNCATE", "TRUNCATE users", false, false},

		// DCL - should be blocked
		{"CREATE ROLE", "CREATE ROLE app", false, false},
		{"GRANT", "GRANT SELECT ON ks.table TO role", false, false},

		// FILE - export allowed, import blocked
		{"COPY TO", "COPY users TO 'data.csv'", true, false},
		{"COPY FROM", "COPY users FROM 'data.csv'", false, false},
		{"SOURCE", "SOURCE 'script.cql'", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			opInfo := ClassifyOperation(tt.command)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			assert.Equal(t, tt.expectAllowed, allowed,
				"Permission mismatch for %s in readonly mode", tt.operation)
			if allowed {
				assert.Equal(t, tt.expectConfirm, needsConf,
					"Confirmation mismatch for %s in readonly mode", tt.operation)
			}
		})
	}
}

// TestPermissionMatrix_ReadwriteMode tests all operations in readwrite mode
func TestPermissionMatrix_ReadwriteMode(t *testing.T) {
	config := &MCPServerConfig{
		Mode:       ConfigModePreset,
		PresetMode: "readwrite",
	}
	server := &MCPServer{config: config}

	tests := []struct {
		operation       string
		command         string
		expectAllowed   bool
		expectConfirm   bool
	}{
		// DQL - allowed
		{"SELECT", "SELECT * FROM users", true, false},
		{"LIST ROLES", "LIST ROLES", true, false},

		// SESSION - always allowed
		{"CONSISTENCY", "CONSISTENCY QUORUM", true, false},

		// DML - allowed
		{"INSERT", "INSERT INTO users VALUES (...)", true, false},
		{"UPDATE", "UPDATE users SET name=?", true, false},
		{"DELETE", "DELETE FROM users", true, false},
		{"BATCH", "BEGIN BATCH", true, false},

		// DDL - blocked
		{"CREATE TABLE", "CREATE TABLE users (...)", false, false},
		{"ALTER TABLE", "ALTER TABLE users ADD col", false, false},
		{"DROP TABLE", "DROP TABLE users", false, false},
		{"TRUNCATE", "TRUNCATE users", false, false},

		// DCL - blocked
		{"CREATE ROLE", "CREATE ROLE app", false, false},
		{"GRANT", "GRANT SELECT ON ks.table TO role", false, false},

		// FILE - all allowed
		{"COPY TO", "COPY users TO 'data.csv'", true, false},
		{"COPY FROM", "COPY users FROM 'data.csv'", true, false},
		{"SOURCE", "SOURCE 'script.cql'", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			opInfo := ClassifyOperation(tt.command)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			assert.Equal(t, tt.expectAllowed, allowed,
				"Permission mismatch for %s in readwrite mode", tt.operation)
			if allowed {
				assert.Equal(t, tt.expectConfirm, needsConf,
					"Confirmation mismatch for %s in readwrite mode", tt.operation)
			}
		})
	}
}

// TestPermissionMatrix_DBAMode tests all operations in dba mode
func TestPermissionMatrix_DBAMode(t *testing.T) {
	config := &MCPServerConfig{
		Mode:       ConfigModePreset,
		PresetMode: "dba",
	}
	server := &MCPServer{config: config}

	// In DBA mode, EVERYTHING should be allowed
	commands := []string{
		"SELECT * FROM users",
		"INSERT INTO users VALUES (...)",
		"UPDATE users SET name=?",
		"DELETE FROM users",
		"CREATE TABLE users (...)",
		"ALTER TABLE users ADD col",
		"DROP TABLE users",
		"TRUNCATE users",
		"CREATE ROLE app",
		"GRANT SELECT ON ks.table TO role",
		"COPY users TO 'data.csv'",
		"COPY users FROM 'data.csv'",
		"SOURCE 'script.cql'",
		"CONSISTENCY QUORUM",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			opInfo := ClassifyOperation(cmd)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			assert.True(t, allowed,
				"All operations should be allowed in dba mode: %s", cmd)
			assert.False(t, needsConf,
				"No confirmations by default in dba mode: %s", cmd)
		})
	}
}

// TestPermissionMatrix_ConfirmQueriesOverlay tests confirm-queries overlay
func TestPermissionMatrix_ConfirmQueriesOverlay(t *testing.T) {
	tests := []struct {
		name           string
		presetMode     string
		confirmQueries []string
		command        string
		expectAllowed  bool
		expectConfirm  bool
	}{
		{
			name:           "readwrite + confirm dml - INSERT needs confirmation",
			presetMode:     "readwrite",
			confirmQueries: []string{"dml"},
			command:        "INSERT INTO users VALUES (...)",
			expectAllowed:  true,
			expectConfirm:  true,
		},
		{
			name:           "readwrite + confirm dml - SELECT no confirmation",
			presetMode:     "readwrite",
			confirmQueries: []string{"dml"},
			command:        "SELECT * FROM users",
			expectAllowed:  true,
			expectConfirm:  false,
		},
		{
			name:           "dba + confirm dcl - CREATE ROLE needs confirmation",
			presetMode:     "dba",
			confirmQueries: []string{"dcl"},
			command:        "CREATE ROLE app",
			expectAllowed:  true,
			expectConfirm:  true,
		},
		{
			name:           "dba + confirm ALL - SELECT needs confirmation",
			presetMode:     "dba",
			confirmQueries: []string{"ALL"},
			command:        "SELECT * FROM users",
			expectAllowed:  true,
			expectConfirm:  true,
		},
		{
			name:           "dba + confirm none - DROP TABLE no confirmation",
			presetMode:     "dba",
			confirmQueries: []string{"none"},
			command:        "DROP TABLE users",
			expectAllowed:  true,
			expectConfirm:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &MCPServerConfig{
				Mode:           ConfigModePreset,
				PresetMode:     tt.presetMode,
				ConfirmQueries: tt.confirmQueries,
			}
			server := &MCPServer{config: config}

			opInfo := ClassifyOperation(tt.command)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			assert.Equal(t, tt.expectAllowed, allowed,
				"Permission mismatch")
			if allowed {
				assert.Equal(t, tt.expectConfirm, needsConf,
					"Confirmation mismatch")
			}
		})
	}
}

// TestPermissionMatrix_FineGrainedMode tests fine-grained skip-confirmation mode
func TestPermissionMatrix_FineGrainedMode(t *testing.T) {
	tests := []struct {
		name             string
		skipConfirmation []string
		command          string
		expectConfirm    bool
	}{
		{
			name:             "skip dql - SELECT no confirmation",
			skipConfirmation: []string{"dql", "session"},
			command:          "SELECT * FROM users",
			expectConfirm:    false,
		},
		{
			name:             "skip dql - INSERT needs confirmation",
			skipConfirmation: []string{"dql", "session"},
			command:          "INSERT INTO users VALUES (...)",
			expectConfirm:    true,
		},
		{
			name:             "skip dql,dml - INSERT no confirmation",
			skipConfirmation: []string{"dql", "dml", "session"},
			command:          "INSERT INTO users VALUES (...)",
			expectConfirm:    false,
		},
		{
			name:             "skip dql,dml - DROP TABLE needs confirmation",
			skipConfirmation: []string{"dql", "dml", "session"},
			command:          "DROP TABLE users",
			expectConfirm:    true,
		},
		{
			name:             "skip ALL - DROP TABLE no confirmation",
			skipConfirmation: []string{"ALL"},
			command:          "DROP TABLE users",
			expectConfirm:    false,
		},
		{
			name:             "skip none - SELECT needs confirmation",
			skipConfirmation: []string{"session"},
			command:          "SELECT * FROM users",
			expectConfirm:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &MCPServerConfig{
				Mode:             ConfigModeFineGrained,
				SkipConfirmation: tt.skipConfirmation,
			}
			server := &MCPServer{config: config}

			opInfo := ClassifyOperation(tt.command)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			// In fine-grained mode, all operations allowed
			assert.True(t, allowed,
				"All operations should be allowed in fine-grained mode: %s", tt.command)

			assert.Equal(t, tt.expectConfirm, needsConf,
				"Confirmation mismatch for: %s", tt.command)
		})
	}
}

// TestPermissionMatrix_SessionAlwaysAllowed tests SESSION category is always allowed
func TestPermissionMatrix_SessionAlwaysAllowed(t *testing.T) {
	modes := []struct {
		name   string
		config *MCPServerConfig
	}{
		{
			name: "readonly mode",
			config: &MCPServerConfig{
				Mode:       ConfigModePreset,
				PresetMode: "readonly",
			},
		},
		{
			name: "readwrite mode",
			config: &MCPServerConfig{
				Mode:       ConfigModePreset,
				PresetMode: "readwrite",
			},
		},
		{
			name: "dba mode",
			config: &MCPServerConfig{
				Mode:       ConfigModePreset,
				PresetMode: "dba",
			},
		},
		{
			name: "fine-grained skip none",
			config: &MCPServerConfig{
				Mode:             ConfigModeFineGrained,
				SkipConfirmation: []string{"session"},
			},
		},
		{
			name: "fine-grained skip dql only",
			config: &MCPServerConfig{
				Mode:             ConfigModeFineGrained,
				SkipConfirmation: []string{"dql", "session"},
			},
		},
	}

	sessionCommands := []string{
		"CONSISTENCY QUORUM",
		"PAGING 100",
		"TRACING ON",
		"AUTOFETCH OFF",
		"EXPAND ON",
		"OUTPUT JSON",
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			server := &MCPServer{config: mode.config}

			for _, cmd := range sessionCommands {
				opInfo := ClassifyOperation(cmd)
				allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

				assert.True(t, allowed,
					"SESSION operation should always be allowed in %s: %s", mode.name, cmd)
				assert.False(t, needsConf,
					"SESSION operation should never need confirmation in %s: %s", mode.name, cmd)
			}
		})
	}
}

// TestRuntimeConfigChange_ModeTransitions tests changing modes at runtime
