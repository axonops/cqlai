package mcp

import (
	"strings"
	"testing"
)

// Helper function to build parameters for each operation type
func buildOperationParams(operation, keyspace, table string) map[string]any {
	// Parse compound operations like "LIST ROLES" -> operation="LIST", options={"object_type":"ROLES"}
	parts := strings.Fields(operation)
	mainOp := parts[0]

	params := map[string]any{
		"operation": mainOp,
		"keyspace":  keyspace,
		"table":     table,
	}

	// Handle compound operations (LIST ROLES, SHOW VERSION, etc.)
	if len(parts) > 1 {
		subType := strings.Join(parts[1:], " ")

		switch strings.ToUpper(mainOp) {
		case "LIST":
			params["options"] = map[string]any{
				"object_type": subType,
			}
			return params
		case "SHOW":
			params["options"] = map[string]any{
				"show_type": subType,
			}
			return params
		case "DESCRIBE", "DESC":
			// Keep operation as DESCRIBE and use existing renderDescribe logic
			params["operation"] = "DESCRIBE"
			// Set table to the describe target for specific describes
			if strings.Contains(subType, "TABLE") && !strings.Contains(subType, "TABLES") {
				params["table"] = table
			} else if strings.Contains(subType, "TYPE") && !strings.Contains(subType, "TYPES") {
				params["table"] = "TYPE"
			} else {
				params["table"] = subType
			}
			return params
		case "CREATE", "DROP", "ALTER":
			// Handle CREATE/DROP/ALTER with specific types (INDEX, TYPE, FUNCTION, etc.)
			params["options"] = map[string]any{
				"object_type": subType,
			}
			// Fall through to add type-specific parameters below
		case "GRANT":
			// GRANT ROLE vs GRANT permission
			if strings.Contains(subType, "ROLE") {
				params["options"] = map[string]any{
					"grant_type": "ROLE",
					"role":       "developer",
					"to_role":    "app_admin",
				}
				return params
			}
			// Fall through for regular permission grants
		case "REVOKE":
			// REVOKE ROLE vs REVOKE permission
			if strings.Contains(subType, "ROLE") {
				params["options"] = map[string]any{
					"grant_type": "ROLE",
					"role":       "developer",
					"from_role":  "app_admin",
				}
				return params
			}
			// Fall through for regular permission revokes
		}
	}

	// After compound operation parsing, add type-specific parameters for CREATE/DROP operations
	if opts, ok := params["options"].(map[string]any); ok {
		if objType, ok := opts["object_type"].(string); ok {
			objTypeUpper := strings.ToUpper(objType)
			switch objTypeUpper {
			case "INDEX":
				opts["index_name"] = "test_idx"
				opts["column"] = "email"
				// Don't need schema for INDEX
				delete(params, "schema")
			case "CUSTOM INDEX":
				opts["index_name"] = "test_idx"
				opts["column"] = "email"
				opts["using"] = "org.apache.cassandra.index.sasi.SASIIndex"
				// For CUSTOM INDEX, we need to normalize to INDEX for the renderer
				opts["object_type"] = "INDEX"
				opts["custom"] = true
				delete(params, "schema")
			case "TYPE":
				opts["type_name"] = "test_type"
				// Keep schema for TYPE (UDT fields)
				params["schema"] = map[string]any{
					"street": "text",
					"city":   "text",
				}
			case "FUNCTION", "OR REPLACE FUNCTION":
				opts["function_name"] = "test_func"
				opts["returns"] = "int"
				opts["language"] = "java"
				opts["body"] = "return 0;"
				delete(params, "schema")
			case "AGGREGATE", "OR REPLACE AGGREGATE":
				opts["aggregate_name"] = "test_agg"
				opts["sfunc"] = "state_func"
				opts["stype"] = "int"
				opts["input_type"] = "int"
				delete(params, "schema")
			case "TRIGGER":
				opts["trigger_name"] = "test_trigger"
				opts["trigger_class"] = "org.apache.cassandra.triggers.TestTrigger"
				delete(params, "schema")
			case "MATERIALIZED VIEW":
				opts["view_name"] = "test_mv"
				opts["base_table"] = table
				opts["primary_key"] = "(id, name)"
				params["columns"] = []string{"id", "name"}
				params["where"] = []any{
					map[string]any{
						"column":   "id",
						"operator": "IS NOT NULL",
						"value":    nil,
					},
					map[string]any{
						"column":   "name",
						"operator": "IS NOT NULL",
						"value":    nil,
					},
				}
				delete(params, "schema")
			case "ROLE":
				opts["role_name"] = "test_role"
				opts["password"] = "test_pass"
				opts["login"] = true
				opts["superuser"] = false
				delete(params, "schema")
			case "USER":
				opts["user_name"] = "test_user"
				opts["password"] = "test_pass"
				opts["superuser"] = false
				delete(params, "schema")
			}
		}
	}

	// Add required parameters based on operation
	switch mainOp {
	case "INSERT":
		params["values"] = map[string]any{
			"id":    "00000000-0000-0000-0000-000000000060",
			"name":  "Test User",
			"email": "test@example.com",
		}
	case "UPDATE":
		params["values"] = map[string]any{"name": "Updated Name"}
		params["where"] = []any{
			map[string]any{
				"column":   "id",
				"operator": "=",
				"value":    "00000000-0000-0000-0000-000000000060",
			},
		}
	case "DELETE":
		params["where"] = []any{
			map[string]any{
				"column":   "id",
				"operator": "=",
				"value":    "00000000-0000-0000-0000-000000000060",
			},
		}
	case "CREATE TABLE", "CREATE COLUMNFAMILY", "CREATE", "CREATE KEYSPACE", "CREATE INDEX", "CREATE CUSTOM INDEX",
		"CREATE MATERIALIZED VIEW", "CREATE TYPE", "CREATE FUNCTION", "CREATE OR REPLACE FUNCTION",
		"CREATE AGGREGATE", "CREATE OR REPLACE AGGREGATE", "CREATE TRIGGER", "CREATE ROLE", "CREATE USER":
		// Add IF NOT EXISTS to prevent conflicts when tests run multiple times
		if params["options"] == nil {
			params["options"] = map[string]any{}
		}
		if opts, ok := params["options"].(map[string]any); ok {
			opts["if_not_exists"] = true
		}
		// Default schema for TABLE/KEYSPACE only (others handled above)
		if _, hasSchema := params["schema"]; !hasSchema {
			if opts, ok := params["options"].(map[string]any); ok {
				if objType, ok := opts["object_type"].(string); ok {
					objTypeUpper := strings.ToUpper(objType)
					// Only add schema for TABLE, COLUMNFAMILY, KEYSPACE, or no object_type
					if objTypeUpper == "TABLE" || objTypeUpper == "COLUMNFAMILY" || objTypeUpper == "KEYSPACE" || objTypeUpper == "" {
						params["schema"] = map[string]any{
							"id":   "uuid PRIMARY KEY",
							"name": "text",
						}
					}
				}
			} else {
				// No options, default to table schema
				params["schema"] = map[string]any{
					"id":   "uuid PRIMARY KEY",
					"name": "text",
				}
			}
		}
	case "ALTER TABLE", "ALTER COLUMNFAMILY", "ALTER", "ALTER KEYSPACE", "ALTER MATERIALIZED VIEW",
		"ALTER TYPE", "ALTER ROLE", "ALTER USER":
		params["options"] = map[string]any{
			"object_type": "TABLE",
			"action":      "ADD",
			"column_name": "age",
			"column_type": "int",
		}
	case "GRANT", "GRANT ROLE":
		params["options"] = map[string]any{
			"permission": "SELECT",
			"role":       "app_readonly",
		}
	case "REVOKE":
		params["options"] = map[string]any{
			"permission": "SELECT",
			"role":       "app_readonly",
		}
	case "REVOKE ROLE":
		params["options"] = map[string]any{
			"grant_type": "ROLE",
			"role":       "developer",
			"from_role":  "app_admin",
		}
	case "ADD IDENTITY":
		params["operation"] = "ALTER"
		params["options"] = map[string]any{
			"object_type": "ROLE",
			"action":      "ADD_IDENTITY",
			"identity":    "user@REALM",
			"role_name":   "app_role",
		}
	case "DROP IDENTITY":
		params["operation"] = "ALTER"
		params["options"] = map[string]any{
			"object_type": "ROLE",
			"action":      "DROP_IDENTITY",
			"identity":    "user@REALM",
			"role_name":   "app_role",
		}
	}

	return params
}

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
