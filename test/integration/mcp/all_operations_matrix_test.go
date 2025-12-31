package mcp

import (
	"testing"
)

// Helper function to build parameters for each operation type

// Complete matrix of ALL 76 operations to test
var complete76Operations = []struct {
	operation string
	category  string
	command   string // Sample command for this operation
}{
	// DQL (14 operations)
	{"SELECT", "DQL", "SELECT * FROM users"},
	{"LIST ROLES", "DQL", "LIST ROLES"},
	{"LIST USERS", "DQL", "LIST USERS"},
	{"LIST PERMISSIONS", "DQL", "LIST PERMISSIONS"},
	{"DESCRIBE KEYSPACES", "DQL", "DESCRIBE KEYSPACES"},
	{"DESCRIBE TABLES", "DQL", "DESCRIBE TABLES"},
	{"DESCRIBE TABLE", "DQL", "DESCRIBE TABLE users"},
	{"DESCRIBE TYPE", "DQL", "DESCRIBE TYPE address"},
	{"DESCRIBE TYPES", "DQL", "DESCRIBE TYPES"},
	{"DESCRIBE CLUSTER", "DQL", "DESCRIBE CLUSTER"},
	{"DESC", "DQL", "DESC TABLE users"},
	{"SHOW VERSION", "DQL", "SHOW VERSION"},
	{"SHOW HOST", "DQL", "SHOW HOST"},
	{"SHOW SESSION", "DQL", "SHOW SESSION"},

	// SESSION (8 operations)
	{"CONSISTENCY", "SESSION", "CONSISTENCY QUORUM"},
	{"PAGING", "SESSION", "PAGING 100"},
	{"TRACING", "SESSION", "TRACING ON"},
	{"AUTOFETCH", "SESSION", "AUTOFETCH ON"},
	{"EXPAND", "SESSION", "EXPAND ON"},
	{"OUTPUT", "SESSION", "OUTPUT JSON"},
	{"CAPTURE", "SESSION", "CAPTURE ON"},
	{"SAVE", "SESSION", "SAVE 'file.csv'"},

	// DML (8 operations)
	{"INSERT", "DML", "INSERT INTO users VALUES (...)"},
	{"UPDATE", "DML", "UPDATE users SET name=?"},
	{"DELETE", "DML", "DELETE FROM users"},
	{"BATCH", "DML", "BATCH"},
	{"BEGIN BATCH", "DML", "BEGIN BATCH"},
	{"BEGIN UNLOGGED BATCH", "DML", "BEGIN UNLOGGED BATCH"},
	{"BEGIN COUNTER BATCH", "DML", "BEGIN COUNTER BATCH"},
	{"APPLY BATCH", "DML", "APPLY BATCH"},

	// DDL - Keyspace (4 operations)
	{"CREATE KEYSPACE", "DDL", "CREATE KEYSPACE ks WITH REPLICATION = ..."},
	{"ALTER KEYSPACE", "DDL", "ALTER KEYSPACE ks WITH REPLICATION = ..."},
	{"DROP KEYSPACE", "DDL", "DROP KEYSPACE ks"},
	{"USE", "DDL", "USE keyspace"},

	// DDL - Table (7 operations)
	{"CREATE TABLE", "DDL", "CREATE TABLE users (...)"},
	{"CREATE COLUMNFAMILY", "DDL", "CREATE COLUMNFAMILY users (...)"},
	{"ALTER TABLE", "DDL", "ALTER TABLE users ADD col"},
	{"ALTER COLUMNFAMILY", "DDL", "ALTER COLUMNFAMILY users ADD col"},
	{"DROP TABLE", "DDL", "DROP TABLE users"},
	{"DROP COLUMNFAMILY", "DDL", "DROP COLUMNFAMILY users"},
	{"TRUNCATE", "DDL", "TRUNCATE users"},

	// DDL - Index (3 operations)
	{"CREATE INDEX", "DDL", "CREATE INDEX ON users(email)"},
	{"CREATE CUSTOM INDEX", "DDL", "CREATE CUSTOM INDEX ON users(name)"},
	{"DROP INDEX", "DDL", "DROP INDEX idx_name"},

	// DDL - Materialized View (3 operations)
	{"CREATE MATERIALIZED VIEW", "DDL", "CREATE MATERIALIZED VIEW mv AS SELECT ..."},
	{"ALTER MATERIALIZED VIEW", "DDL", "ALTER MATERIALIZED VIEW mv WITH ..."},
	{"DROP MATERIALIZED VIEW", "DDL", "DROP MATERIALIZED VIEW mv"},

	// DDL - Type/UDT (3 operations)
	{"CREATE TYPE", "DDL", "CREATE TYPE address (...)"},
	{"ALTER TYPE", "DDL", "ALTER TYPE address ADD field"},
	{"DROP TYPE", "DDL", "DROP TYPE address"},

	// DDL - Function (3 operations)
	{"CREATE FUNCTION", "DDL", "CREATE FUNCTION func(...)"},
	{"CREATE OR REPLACE FUNCTION", "DDL", "CREATE OR REPLACE FUNCTION func(...)"},
	{"DROP FUNCTION", "DDL", "DROP FUNCTION func"},

	// DDL - Aggregate (3 operations)
	{"CREATE AGGREGATE", "DDL", "CREATE AGGREGATE avg(...)"},
	{"CREATE OR REPLACE AGGREGATE", "DDL", "CREATE OR REPLACE AGGREGATE avg(...)"},
	{"DROP AGGREGATE", "DDL", "DROP AGGREGATE avg"},

	// DDL - Trigger (2 operations)
	{"CREATE TRIGGER", "DDL", "CREATE TRIGGER trig ON table"},
	{"DROP TRIGGER", "DDL", "DROP TRIGGER trig"},

	// DCL - Role (5 operations)
	{"CREATE ROLE", "DCL", "CREATE ROLE app WITH PASSWORD='...'"},
	{"ALTER ROLE", "DCL", "ALTER ROLE app WITH PASSWORD='...'"},
	{"DROP ROLE", "DCL", "DROP ROLE app"},
	{"GRANT ROLE", "DCL", "GRANT ROLE admin TO user"},
	{"REVOKE ROLE", "DCL", "REVOKE ROLE admin FROM user"},

	// DCL - User (3 operations - legacy)
	{"CREATE USER", "DCL", "CREATE USER admin WITH PASSWORD='...'"},
	{"ALTER USER", "DCL", "ALTER USER admin WITH PASSWORD='...'"},
	{"DROP USER", "DCL", "DROP USER admin"},

	// DCL - Identity (2 operations)
	{"ADD IDENTITY", "DCL", "ADD IDENTITY 'user@REALM' TO role"},
	{"DROP IDENTITY", "DCL", "DROP IDENTITY 'user@REALM' FROM role"},

	// DCL - Permission (2 operations)
	{"GRANT", "DCL", "GRANT SELECT ON ks.table TO role"},
	{"REVOKE", "DCL", "REVOKE ALL ON ks.table FROM role"},

	// FILE (3 operations)
	{"COPY TO", "FILE", "COPY users TO 'data.csv'"},
	{"COPY FROM", "FILE", "COPY users FROM 'data.csv'"},
	{"SOURCE", "FILE", "SOURCE 'script.cql'"},
}

// TestComplete76Operations_ReadonlyMode tests all 76 operations in readonly
func TestComplete76Operations_ReadonlyMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readonly.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)

	for _, op := range complete76Operations {
		t.Run(op.operation, func(t *testing.T) {
			params := buildOperationParams(op.operation, "test_mcp", "users")
			resp := callTool(t, ctx.SocketPath, "submit_query_plan", params)

			// Only DQL and SESSION should be allowed in readonly
			if op.category == "DQL" || op.category == "SESSION" {
				assertNotError(t, resp, op.operation+" should be allowed in readonly")
			} else if op.category == "FILE" && op.operation == "COPY TO" {
				// COPY TO (export) is allowed in readonly
				assertNotError(t, resp, "COPY TO should be allowed in readonly")
			} else {
				// Everything else blocked
				assertIsError(t, resp, op.operation+" should be blocked in readonly")
				assertContains(t, resp, "not allowed")
			}
		})
	}
}

// TestComplete76Operations_ReadwriteMode tests all 76 operations in readwrite
func TestComplete76Operations_ReadwriteMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readwrite.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)

	for _, op := range complete76Operations {
		t.Run(op.operation, func(t *testing.T) {
			params := buildOperationParams(op.operation, "test_mcp", "users")
			resp := callTool(t, ctx.SocketPath, "submit_query_plan", params)

			// DQL, SESSION, DML, FILE should be allowed
			if op.category == "DQL" || op.category == "SESSION" || op.category == "DML" || op.category == "FILE" {
				assertNotError(t, resp, op.operation+" should be allowed in readwrite")
			} else {
				// DDL and DCL blocked
				assertIsError(t, resp, op.operation+" should be blocked in readwrite")
			}
		})
	}
}

// TestComplete76Operations_DBAMode tests all 76 operations in DBA mode
func TestComplete76Operations_DBAMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/dba.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// In DBA mode, ALL operations should be allowed
	for _, op := range complete76Operations {
		t.Run(op.operation, func(t *testing.T) {
			params := buildOperationParams(op.operation, "test_mcp", "users")
			resp := callTool(t, ctx.SocketPath, "submit_query_plan", params)

			assertNotError(t, resp, op.operation+" should be allowed in DBA mode")
		})
	}
}

// TestComplete76Operations_ConfirmALL tests all 76 ops with confirm ALL
func TestComplete76Operations_ConfirmALL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/dba_confirm_all.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)

	for _, op := range complete76Operations {
		t.Run(op.operation, func(t *testing.T) {
			params := buildOperationParams(op.operation, "test_mcp", "users")
			resp := callTool(t, ctx.SocketPath, "submit_query_plan", params)

			// SESSION never requires confirmation
			if op.category == "SESSION" {
				assertNotError(t, resp, "SESSION operations never require confirmation")
			} else {
				// Everything else should require confirmation
				assertIsError(t, resp, op.operation+" should require confirmation with confirm ALL")
				assertContains(t, resp, "requires")
			}
		})
	}
}

// TestComplete76Operations_SkipALL tests all 76 ops with skip ALL
func TestComplete76Operations_SkipALL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/finegrained_skip_all.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// With skip ALL, everything should work without confirmation
	for _, op := range complete76Operations {
		t.Run(op.operation, func(t *testing.T) {
			params := buildOperationParams(op.operation, "test_mcp", "users")
			resp := callTool(t, ctx.SocketPath, "submit_query_plan", params)

			assertNotError(t, resp, op.operation+" should work with skip ALL")
		})
	}
}
