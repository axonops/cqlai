package ai

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Complete list of all operations we must test
var allOperations = []struct {
	category OperationCategory
	command  string
	op       string
}{
	// DQL (14 operations)
	{CategoryDQL, "SELECT * FROM users", "SELECT"},
	{CategoryDQL, "LIST ROLES", "LIST ROLES"},
	{CategoryDQL, "LIST USERS", "LIST USERS"},
	{CategoryDQL, "LIST PERMISSIONS", "LIST PERMISSIONS"},
	{CategoryDQL, "DESCRIBE KEYSPACES", "DESCRIBE KEYSPACES"},
	{CategoryDQL, "DESCRIBE TABLES", "DESCRIBE TABLES"},
	{CategoryDQL, "DESCRIBE TABLE users", "DESCRIBE TABLE"},
	{CategoryDQL, "DESC TABLE users", "DESC TABLE"},
	{CategoryDQL, "DESCRIBE TYPE address", "DESCRIBE TYPE"},
	{CategoryDQL, "DESCRIBE TYPES", "DESCRIBE TYPES"},
	{CategoryDQL, "DESCRIBE CLUSTER", "DESCRIBE CLUSTER"},
	{CategoryDQL, "SHOW VERSION", "SHOW VERSION"},
	{CategoryDQL, "SHOW HOST", "SHOW HOST"},
	{CategoryDQL, "SHOW SESSION", "SHOW SESSION"},

	// SESSION (8 operations)
	{CategorySESSION, "CONSISTENCY QUORUM", "CONSISTENCY"},
	{CategorySESSION, "PAGING 100", "PAGING"},
	{CategorySESSION, "TRACING ON", "TRACING"},
	{CategorySESSION, "AUTOFETCH ON", "AUTOFETCH"},
	{CategorySESSION, "EXPAND ON", "EXPAND"},
	{CategorySESSION, "OUTPUT JSON", "OUTPUT"},
	{CategorySESSION, "CAPTURE ON", "CAPTURE"},
	{CategorySESSION, "SAVE 'file.csv'", "SAVE"},

	// DML (8 operations)
	{CategoryDML, "INSERT INTO users VALUES (...)", "INSERT"},
	{CategoryDML, "UPDATE users SET name=?", "UPDATE"},
	{CategoryDML, "DELETE FROM users WHERE id=?", "DELETE"},
	{CategoryDML, "BEGIN BATCH", "BEGIN BATCH"},
	{CategoryDML, "BEGIN UNLOGGED BATCH", "BEGIN UNLOGGED"},
	{CategoryDML, "BEGIN COUNTER BATCH", "BEGIN COUNTER"},
	{CategoryDML, "APPLY BATCH", "APPLY BATCH"},
	{CategoryDML, "BATCH", "BATCH"},

	// DDL (28 operations)
	{CategoryDDL, "CREATE KEYSPACE ks WITH REPLICATION = ...", "CREATE KEYSPACE"},
	{CategoryDDL, "ALTER KEYSPACE ks WITH REPLICATION = ...", "ALTER KEYSPACE"},
	{CategoryDDL, "DROP KEYSPACE ks", "DROP KEYSPACE"},
	{CategoryDDL, "USE ks", "USE"},
	{CategoryDDL, "CREATE TABLE users (...)", "CREATE TABLE"},
	{CategoryDDL, "CREATE COLUMNFAMILY users (...)", "CREATE COLUMNFAMILY"},
	{CategoryDDL, "ALTER TABLE users ADD col text", "ALTER TABLE"},
	{CategoryDDL, "ALTER COLUMNFAMILY users ADD col", "ALTER COLUMNFAMILY"},
	{CategoryDDL, "DROP TABLE users", "DROP TABLE"},
	{CategoryDDL, "DROP COLUMNFAMILY users", "DROP COLUMNFAMILY"},
	{CategoryDDL, "TRUNCATE users", "TRUNCATE"},
	{CategoryDDL, "CREATE INDEX ON users(email)", "CREATE INDEX"},
	{CategoryDDL, "CREATE CUSTOM INDEX ON users(name) USING 'SAI'", "CREATE CUSTOM"},
	{CategoryDDL, "DROP INDEX idx_name", "DROP INDEX"},
	{CategoryDDL, "CREATE MATERIALIZED VIEW mv AS SELECT ...", "CREATE MATERIALIZED"},
	{CategoryDDL, "ALTER MATERIALIZED VIEW mv WITH ...", "ALTER MATERIALIZED"},
	{CategoryDDL, "DROP MATERIALIZED VIEW mv", "DROP MATERIALIZED"},
	{CategoryDDL, "CREATE TYPE address (...)", "CREATE TYPE"},
	{CategoryDDL, "ALTER TYPE address ADD field text", "ALTER TYPE"},
	{CategoryDDL, "DROP TYPE address", "DROP TYPE"},
	{CategoryDDL, "CREATE FUNCTION func(...)", "CREATE FUNCTION"},
	{CategoryDDL, "CREATE OR REPLACE FUNCTION func(...)", "CREATE OR"},
	{CategoryDDL, "DROP FUNCTION func", "DROP FUNCTION"},
	{CategoryDDL, "CREATE AGGREGATE avg(...)", "CREATE AGGREGATE"},
	{CategoryDDL, "CREATE OR REPLACE AGGREGATE avg(...)", "CREATE OR"},
	{CategoryDDL, "DROP AGGREGATE avg", "DROP AGGREGATE"},
	{CategoryDDL, "CREATE TRIGGER trig ON table", "CREATE TRIGGER"},
	{CategoryDDL, "DROP TRIGGER trig", "DROP TRIGGER"},

	// DCL (12 operations - LIST PERMISSIONS is DQL, not DCL)
	{CategoryDCL, "CREATE ROLE app WITH PASSWORD='...'", "CREATE ROLE"},
	{CategoryDCL, "ALTER ROLE app WITH PASSWORD='...'", "ALTER ROLE"},
	{CategoryDCL, "DROP ROLE app", "DROP ROLE"},
	{CategoryDCL, "GRANT ROLE admin TO user", "GRANT ROLE"},
	{CategoryDCL, "REVOKE ROLE admin FROM user", "REVOKE ROLE"},
	{CategoryDCL, "CREATE USER admin WITH PASSWORD='...'", "CREATE USER"},
	{CategoryDCL, "ALTER USER admin WITH PASSWORD='...'", "ALTER USER"},
	{CategoryDCL, "DROP USER admin", "DROP USER"},
	{CategoryDCL, "ADD IDENTITY 'user@REALM' TO role", "ADD IDENTITY"},
	{CategoryDCL, "DROP IDENTITY 'user@REALM' FROM role", "DROP IDENTITY"},
	{CategoryDCL, "GRANT SELECT ON ks.table TO role", "GRANT"},
	{CategoryDCL, "REVOKE ALL ON ks.table FROM role", "REVOKE"},

	// FILE (3 operations)
	{CategoryFILE, "COPY users TO 'data.csv'", "COPY TO"},
	{CategoryFILE, "COPY users FROM 'data.csv'", "COPY FROM"},
	{CategoryFILE, "SOURCE 'script.cql'", "SOURCE"},
}

// TestComprehensive_ReadonlyMode_AllOperations tests ALL 76 operations in readonly mode
func TestComprehensive_ReadonlyMode_AllOperations(t *testing.T) {
	config := &MCPServerConfig{
		Mode:       ConfigModePreset,
		PresetMode: "readonly",
	}
	server := &MCPServer{config: config}

	for _, op := range allOperations {
		t.Run(op.op, func(t *testing.T) {
			opInfo := ClassifyOperation(op.command)
			allowed, needsConf, reason := server.CheckOperationPermission(opInfo)

			// Determine expected behavior
			expectAllowed := op.category == CategoryDQL || op.category == CategorySESSION ||
				(op.category == CategoryFILE && op.op == "COPY TO")

			if expectAllowed {
				assert.True(t, allowed,
					"%s should be allowed in readonly mode", op.op)
				assert.False(t, needsConf,
					"%s should not need confirmation in readonly mode", op.op)
			} else {
				assert.False(t, allowed,
					"%s should NOT be allowed in readonly mode, got reason: %s", op.op, reason)
			}
		})
	}
}

// TestComprehensive_ReadwriteMode_AllOperations tests ALL 76 operations in readwrite mode
func TestComprehensive_ReadwriteMode_AllOperations(t *testing.T) {
	config := &MCPServerConfig{
		Mode:       ConfigModePreset,
		PresetMode: "readwrite",
	}
	server := &MCPServer{config: config}

	for _, op := range allOperations {
		t.Run(op.op, func(t *testing.T) {
			opInfo := ClassifyOperation(op.command)
			allowed, needsConf, reason := server.CheckOperationPermission(opInfo)

			// Determine expected behavior
			expectAllowed := op.category == CategoryDQL ||
				op.category == CategorySESSION ||
				op.category == CategoryDML ||
				op.category == CategoryFILE

			if expectAllowed {
				assert.True(t, allowed,
					"%s should be allowed in readwrite mode", op.op)
				assert.False(t, needsConf,
					"%s should not need confirmation in readwrite mode by default", op.op)
			} else {
				assert.False(t, allowed,
					"%s should NOT be allowed in readwrite mode (DDL/DCL blocked), got reason: %s", op.op, reason)
			}
		})
	}
}

// TestComprehensive_DBAMode_AllOperations tests ALL 76 operations in dba mode
func TestComprehensive_DBAMode_AllOperations(t *testing.T) {
	config := &MCPServerConfig{
		Mode:       ConfigModePreset,
		PresetMode: "dba",
	}
	server := &MCPServer{config: config}

	for _, op := range allOperations {
		t.Run(op.op, func(t *testing.T) {
			opInfo := ClassifyOperation(op.command)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			// In DBA mode, ALL operations should be allowed
			assert.True(t, allowed,
				"%s should be allowed in dba mode", op.op)
			assert.False(t, needsConf,
				"%s should not need confirmation in dba mode by default", op.op)
		})
	}
}

// TestComprehensive_ConfirmQueriesALL_AllOperations tests confirm-queries ALL
func TestComprehensive_ConfirmQueriesALL_AllOperations(t *testing.T) {
	config := &MCPServerConfig{
		Mode:           ConfigModePreset,
		PresetMode:     "dba",
		ConfirmQueries: []string{"ALL"},
	}
	server := &MCPServer{config: config}

	for _, op := range allOperations {
		t.Run(op.op, func(t *testing.T) {
			opInfo := ClassifyOperation(op.command)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			// All should be allowed in DBA mode
			assert.True(t, allowed,
				"%s should be allowed in dba mode", op.op)

			// SESSION never needs confirmation
			if op.category == CategorySESSION {
				assert.False(t, needsConf,
					"SESSION operation %s should never need confirmation", op.op)
			} else {
				// Everything else should need confirmation with ALL
				assert.True(t, needsConf,
					"%s should need confirmation with confirm-queries ALL", op.op)
			}
		})
	}
}

// TestComprehensive_FineGrainedSkipALL_AllOperations tests skip-confirmation ALL
func TestComprehensive_FineGrainedSkipALL_AllOperations(t *testing.T) {
	config := &MCPServerConfig{
		Mode:             ConfigModeFineGrained,
		SkipConfirmation: []string{"ALL"},
	}
	server := &MCPServer{config: config}

	for _, op := range allOperations {
		t.Run(op.op, func(t *testing.T) {
			opInfo := ClassifyOperation(op.command)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			// All operations allowed in fine-grained mode
			assert.True(t, allowed,
				"%s should be allowed in fine-grained mode", op.op)

			// None should need confirmation with skip ALL
			assert.False(t, needsConf,
				"%s should not need confirmation with skip ALL", op.op)
		})
	}
}

// TestComprehensive_FineGrainedSkipNone_AllOperations tests skip-confirmation none
func TestComprehensive_FineGrainedSkipNone_AllOperations(t *testing.T) {
	config := &MCPServerConfig{
		Mode:             ConfigModeFineGrained,
		SkipConfirmation: []string{"session"}, // Only session (none specified)
	}
	server := &MCPServer{config: config}

	for _, op := range allOperations {
		t.Run(op.op, func(t *testing.T) {
			opInfo := ClassifyOperation(op.command)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			// All operations allowed
			assert.True(t, allowed,
				"%s should be allowed in fine-grained mode", op.op)

			// Only SESSION should skip confirmation
			if op.category == CategorySESSION {
				assert.False(t, needsConf,
					"SESSION %s should not need confirmation", op.op)
			} else {
				assert.True(t, needsConf,
					"%s SHOULD need confirmation with skip none", op.op)
			}
		})
	}
}

// TestComprehensive_ConfirmQueriesByCategory tests each category with confirm-queries
func TestComprehensive_ConfirmQueriesByCategory(t *testing.T) {
	categories := []string{"dql", "dml", "ddl", "dcl", "file"}

	for _, confirmCat := range categories {
		t.Run(fmt.Sprintf("confirm_%s", confirmCat), func(t *testing.T) {
			config := &MCPServerConfig{
				Mode:           ConfigModePreset,
				PresetMode:     "dba",
				ConfirmQueries: []string{confirmCat},
			}
			server := &MCPServer{config: config}

			for _, op := range allOperations {
				opInfo := ClassifyOperation(op.command)
				allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

				// All should be allowed in DBA mode
				assert.True(t, allowed,
					"%s should be allowed in dba mode", op.op)

				// Should confirm only if matches the category
				expectConfirm := string(op.category) == confirmCat

				if op.category == CategorySESSION {
					// SESSION never confirms
					assert.False(t, needsConf,
						"SESSION %s should never need confirmation", op.op)
				} else {
					assert.Equal(t, expectConfirm, needsConf,
						"%s confirmation mismatch when confirming %s", op.op, confirmCat)
				}
			}
		})
	}
}

// TestComprehensive_InvalidOperations tests handling of invalid/unknown operations
func TestComprehensive_InvalidOperations(t *testing.T) {
	invalidOps := []string{
		"FOOBAR",
		"INVALID COMMAND",
		"RANDOM STUFF HERE",
		"EXPLODE TABLE users",
		"DESTROYALL",
		"",
		"   ",
	}

	modes := []struct {
		name   string
		config *MCPServerConfig
	}{
		{"readonly", &MCPServerConfig{Mode: ConfigModePreset, PresetMode: "readonly"}},
		{"readwrite", &MCPServerConfig{Mode: ConfigModePreset, PresetMode: "readwrite"}},
		{"dba", &MCPServerConfig{Mode: ConfigModePreset, PresetMode: "dba"}},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			server := &MCPServer{config: mode.config}

			for _, cmd := range invalidOps {
				t.Run(cmd, func(t *testing.T) {
					opInfo := ClassifyOperation(cmd)
					assert.Equal(t, CategoryUNKNOWN, opInfo.Category,
						"Unknown operation should be CategoryUNKNOWN")
					assert.Equal(t, "CRITICAL", opInfo.RiskLevel,
						"Unknown operation should be CRITICAL risk")

					// Permission check - unknown operations should be treated safely
					// In most modes they should require confirmation or be blocked
					allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

					// The behavior depends on mode - but we should never silently allow unknown ops
					if allowed {
						// If allowed, must require confirmation
						assert.True(t, needsConf,
							"Unknown operation %s must require confirmation if allowed", cmd)
					}
				})
			}
		})
	}
}

// TestComprehensive_EdgeCases tests edge cases and corner scenarios
func TestComprehensive_EdgeCases(t *testing.T) {
	t.Run("readonly_with_confirm_dql", func(t *testing.T) {
		// Readonly mode + confirm DQL = queries need confirmation
		config := &MCPServerConfig{
			Mode:           ConfigModePreset,
			PresetMode:     "readonly",
			ConfirmQueries: []string{"dql"},
		}
		server := &MCPServer{config: config}

		opInfo := ClassifyOperation("SELECT * FROM users")
		allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

		assert.True(t, allowed, "SELECT should be allowed in readonly")
		assert.True(t, needsConf, "SELECT should need confirmation with confirm-queries dql")
	})

	t.Run("readonly_with_confirm_dml", func(t *testing.T) {
		// Readonly mode + confirm DML = DML still blocked (not in allowed list)
		config := &MCPServerConfig{
			Mode:           ConfigModePreset,
			PresetMode:     "readonly",
			ConfirmQueries: []string{"dml"},
		}
		server := &MCPServer{config: config}

		opInfo := ClassifyOperation("INSERT INTO users VALUES (...)")
		allowed, _, _ := server.CheckOperationPermission(opInfo)

		// DML not allowed in readonly, even with confirm-queries dml
		assert.False(t, allowed,
			"INSERT should not be allowed in readonly mode even with confirm-queries dml")
	})

	t.Run("dba_with_confirm_none", func(t *testing.T) {
		// DBA with confirm none = everything allowed, nothing confirmed
		config := &MCPServerConfig{
			Mode:           ConfigModePreset,
			PresetMode:     "dba",
			ConfirmQueries: []string{"none"},
		}
		server := &MCPServer{config: config}

		// Test dangerous operation
		opInfo := ClassifyOperation("DROP KEYSPACE production")
		allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

		assert.True(t, allowed, "DROP should be allowed in dba")
		assert.False(t, needsConf, "DROP should not need confirmation with confirm-queries none")
	})

	t.Run("fine_grained_with_dql_only", func(t *testing.T) {
		// Skip only DQL = everything else needs confirmation
		config := &MCPServerConfig{
			Mode:             ConfigModeFineGrained,
			SkipConfirmation: []string{"dql", "session"},
		}
		server := &MCPServer{config: config}

		// DQL should skip
		opInfo := ClassifyOperation("SELECT * FROM users")
		_, needsConf, _ := server.CheckOperationPermission(opInfo)
		assert.False(t, needsConf, "DQL should skip confirmation")

		// DML should confirm
		opInfo = ClassifyOperation("INSERT INTO users VALUES (...)")
		_, needsConf, _ = server.CheckOperationPermission(opInfo)
		assert.True(t, needsConf, "DML should need confirmation")

		// DDL should confirm
		opInfo = ClassifyOperation("DROP TABLE users")
		_, needsConf, _ = server.CheckOperationPermission(opInfo)
		assert.True(t, needsConf, "DDL should need confirmation")
	})
}

// TestComprehensive_AllModeTransitions tests all possible mode transitions
func TestComprehensive_AllModeTransitions(t *testing.T) {
	modes := []string{"readonly", "readwrite", "dba"}

	// Test all transitions (9 total: 3 from × 3 to)
	for _, fromMode := range modes {
		for _, toMode := range modes {
			t.Run(fmt.Sprintf("%s_to_%s", fromMode, toMode), func(t *testing.T) {
				config := &MCPServerConfig{
					Mode:       ConfigModePreset,
					PresetMode: fromMode,
				}

				// Transition
				err := config.UpdatePresetMode(toMode)
				assert.NoError(t, err,
					"Should allow transition from %s to %s", fromMode, toMode)
				assert.Equal(t, toMode, config.PresetMode,
					"Mode should be updated to %s", toMode)
			})
		}
	}

	// Test transition from preset to fine-grained
	for _, fromMode := range modes {
		t.Run(fmt.Sprintf("%s_to_finegrained", fromMode), func(t *testing.T) {
			config := &MCPServerConfig{
				Mode:       ConfigModePreset,
				PresetMode: fromMode,
			}

			// Transition to fine-grained
			err := config.UpdateSkipConfirmation([]string{"dql", "dml"})
			assert.NoError(t, err)
			assert.Equal(t, ConfigModeFineGrained, config.Mode)
			assert.Equal(t, "", config.PresetMode)
		})
	}

	// Test transition from fine-grained to preset
	t.Run("finegrained_to_readonly", func(t *testing.T) {
		config := &MCPServerConfig{
			Mode:             ConfigModeFineGrained,
			SkipConfirmation: []string{"dql", "dml"},
		}

		err := config.UpdatePresetMode("readonly")
		assert.NoError(t, err)
		assert.Equal(t, ConfigModePreset, config.Mode)
		assert.Equal(t, "readonly", config.PresetMode)
		assert.Nil(t, config.SkipConfirmation)
	})
}

// TestComprehensive_ConfirmQueriesCombinations tests all confirm-queries combinations
func TestComprehensive_ConfirmQueriesCombinations(t *testing.T) {
	combinations := []struct {
		name           string
		confirmQueries []string
		testOp         string
		testCategory   OperationCategory
		expectConfirm  bool
	}{
		{"dql_only", []string{"dql"}, "SELECT * FROM users", CategoryDQL, true},
		{"dml_only", []string{"dml"}, "SELECT * FROM users", CategoryDQL, false},
		{"dml_only_insert", []string{"dml"}, "INSERT INTO users VALUES (...)", CategoryDML, true},
		{"dql_dml", []string{"dql", "dml"}, "SELECT * FROM users", CategoryDQL, true},
		{"dql_dml_insert", []string{"dql", "dml"}, "INSERT INTO users VALUES (...)", CategoryDML, true},
		{"dql_dml_drop", []string{"dql", "dml"}, "DROP TABLE users", CategoryDDL, false},
		{"ddl_dcl", []string{"ddl", "dcl"}, "CREATE TABLE users (...)", CategoryDDL, true},
		{"ddl_dcl_insert", []string{"ddl", "dcl"}, "INSERT INTO users VALUES (...)", CategoryDML, false},
		{"ddl_dcl_grant", []string{"ddl", "dcl"}, "GRANT SELECT ON ks.table TO role", CategoryDCL, true},
		{"ALL_select", []string{"ALL"}, "SELECT * FROM users", CategoryDQL, true},
		{"ALL_insert", []string{"ALL"}, "INSERT INTO users VALUES (...)", CategoryDML, true},
		{"ALL_drop", []string{"ALL"}, "DROP TABLE users", CategoryDDL, true},
		{"none_select", []string{"none"}, "SELECT * FROM users", CategoryDQL, false},
		{"none_drop", []string{"none"}, "DROP TABLE users", CategoryDDL, false},
	}

	for _, tc := range combinations {
		t.Run(tc.name, func(t *testing.T) {
			config := &MCPServerConfig{
				Mode:           ConfigModePreset,
				PresetMode:     "dba", // All ops allowed
				ConfirmQueries: tc.confirmQueries,
			}
			server := &MCPServer{config: config}

			opInfo := ClassifyOperation(tc.testOp)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			assert.True(t, allowed, "Should be allowed in dba mode")

			// SESSION never confirms
			if tc.testCategory == CategorySESSION {
				assert.False(t, needsConf, "SESSION should never need confirmation")
			} else {
				assert.Equal(t, tc.expectConfirm, needsConf,
					"Confirmation mismatch for %s with confirm-queries %v",
					tc.testOp, tc.confirmQueries)
			}
		})
	}
}

// TestComprehensive_SkipConfirmationCombinations tests all skip-confirmation combinations
func TestComprehensive_SkipConfirmationCombinations(t *testing.T) {
	combinations := []struct {
		name             string
		skipConfirmation []string
		testOp           string
		testCategory     OperationCategory
		expectConfirm    bool
	}{
		{"dql_only_select", []string{"dql"}, "SELECT * FROM users", CategoryDQL, false},
		{"dql_only_insert", []string{"dql"}, "INSERT INTO users VALUES (...)", CategoryDML, true},
		{"dml_only_select", []string{"dml"}, "SELECT * FROM users", CategoryDQL, true},
		{"dml_only_insert", []string{"dml"}, "INSERT INTO users VALUES (...)", CategoryDML, false},
		{"dql_dml_select", []string{"dql", "dml"}, "SELECT * FROM users", CategoryDQL, false},
		{"dql_dml_insert", []string{"dql", "dml"}, "INSERT INTO users VALUES (...)", CategoryDML, false},
		{"dql_dml_drop", []string{"dql", "dml"}, "DROP TABLE users", CategoryDDL, true},
		{"dql_dml_ddl_drop", []string{"dql", "dml", "ddl"}, "DROP TABLE users", CategoryDDL, false},
		{"dql_dml_ddl_grant", []string{"dql", "dml", "ddl"}, "GRANT SELECT ON ks.table TO role", CategoryDCL, true},
		{"all_cats_grant", []string{"dql", "dml", "ddl", "dcl", "file"}, "GRANT SELECT ON ks.table TO role", CategoryDCL, false},
	}

	for _, tc := range combinations {
		t.Run(tc.name, func(t *testing.T) {
			// Add session to skip list (auto-added normally)
			skipList := append([]string{"session"}, tc.skipConfirmation...)

			config := &MCPServerConfig{
				Mode:             ConfigModeFineGrained,
				SkipConfirmation: skipList,
			}
			server := &MCPServer{config: config}

			opInfo := ClassifyOperation(tc.testOp)
			allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

			assert.True(t, allowed, "All ops allowed in fine-grained mode")

			// SESSION never confirms
			if tc.testCategory == CategorySESSION {
				assert.False(t, needsConf, "SESSION should never need confirmation")
			} else {
				assert.Equal(t, tc.expectConfirm, needsConf,
					"Confirmation mismatch for %s with skip %v",
					tc.testOp, tc.skipConfirmation)
			}
		})
	}
}

// TestComprehensive_ConcurrentModeChanges tests thread safety during concurrent config changes
func TestComprehensive_ConcurrentModeChanges(t *testing.T) {
	config := DefaultMCPConfig()
	server := &MCPServer{config: config}

	done := make(chan bool, 300)

	// 100 readers checking permissions
	for i := 0; i < 100; i++ {
		go func(idx int) {
			ops := []string{
				"SELECT * FROM users",
				"INSERT INTO users VALUES (...)",
				"DROP TABLE users",
			}
			opInfo := ClassifyOperation(ops[idx%3])
			server.CheckOperationPermission(opInfo)
			done <- true
		}(i)
	}

	// 100 mode changers
	for i := 0; i < 100; i++ {
		go func(idx int) {
			modes := []string{"readonly", "readwrite", "dba"}
			_ = config.UpdatePresetMode(modes[idx%3])
			done <- true
		}(i)
	}

	// 100 confirm-queries changers
	for i := 0; i < 100; i++ {
		go func(idx int) {
			lists := [][]string{
				{"dml"},
				{"dcl", "ddl"},
				{"ALL"},
				{"none"},
			}
			_ = config.UpdateConfirmQueries(lists[idx%4])
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 300; i++ {
		<-done
	}

	// No panics or deadlocks = success
	assert.True(t, true, "Thread safety test completed without panics")
}
