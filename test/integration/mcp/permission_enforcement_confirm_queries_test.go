package mcp

import (
	"os"
	"testing"

	"github.com/axonops/cqlai/internal/ai"
)

func init() {
	key, _ := ai.GenerateAPIKey()
	os.Setenv("TEST_MCP_API_KEY", key)
}

// TestConfirmQueries_ReadwriteWithDML tests readwrite + confirm dml
func TestConfirmQueries_ReadwriteWithDML(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite_confirm_dml.json")
	defer stopMCPHTTP(ctx)

	// DQL should work without confirmation
	t.Run("SELECT_no_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work without confirmation")
	})

	// DML should require confirmation
	t.Run("INSERT_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000001",
				"name":  "Test User",
				"email": "test@example.com",
			},
		})
	})

	t.Run("DELETE_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "DELETE",
			"keyspace":  "test_mcp",
			"table":     "users",
			"where": []any{
				map[string]any{
					"column":   "id",
					"operator": "=",
					"value":    "00000000-0000-0000-0000-000000000001",
				},
			},
		})
	})

	// DDL still blocked (not allowed in readwrite)
	t.Run("CREATE_blocked", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
		assertIsError(t, resp, "CREATE should be blocked")
		assertContains(t, resp, "not allowed")
	})
}

// TestConfirmQueries_DBA_ConfirmDDL tests dba + confirm ddl
func TestConfirmQueries_DBA_ConfirmDDL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfigHTTP(t, "testdata/dba_confirm_ddl.json")
	defer stopMCPHTTP(ctx)

	// Ensure test data exists (in case previous tests dropped tables)
	ensureTestDataExists(t, ctx.Session)

	// DML should work without confirmation
	t.Run("INSERT_no_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000002",
				"name":  "DBA Test User",
				"email": "dba@example.com",
			},
		})
		assertNotError(t, resp, "INSERT should work without confirmation in DBA")
	})

	// DDL should require confirmation
	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
	})

	t.Run("DROP_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "DROP",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
	})

	// DCL should work without confirmation
	t.Run("GRANT_no_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
		assertNotError(t, resp, "GRANT should work without confirmation")
	})
}

// TestConfirmQueries_DBA_ConfirmMultiple tests dba + confirm dml,ddl
func TestConfirmQueries_DBA_ConfirmMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfigHTTP(t, "testdata/dba_confirm_dml_ddl.json")
	defer stopMCPHTTP(ctx)

	// DQL should work
	t.Run("SELECT_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work")
	})

	// Both DML and DDL should require confirmation
	t.Run("INSERT_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000003",
				"name":  "Multi Test User",
				"email": "multi@example.com",
			},
		})
	})

	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
	})

	// DCL should work
	t.Run("GRANT_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
		assertNotError(t, resp, "GRANT should work without confirmation")
	})
}

// TestConfirmQueries_ReadonlyWithDQL tests readonly + confirm dql (overlay on allowed)
func TestConfirmQueries_ReadonlyWithDQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfigHTTP(t, "testdata/readonly_confirm_dql.json")
	defer stopMCPHTTP(ctx)

	// Even SELECT should require confirmation
	t.Run("SELECT_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
	})

	// DML still blocked (not in allowed list)
	t.Run("INSERT_still_blocked", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000004",
				"name":  "Readonly Test",
				"email": "readonly@example.com",
			},
		})
		assertIsError(t, resp, "INSERT should still be blocked")
		assertContains(t, resp, "not allowed")
	})
}

// TestConfirmQueries_ConfirmALL tests confirm ALL (most restrictive)
func TestConfirmQueries_ConfirmALL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfigHTTP(t, "testdata/dba_confirm_all.json")
	defer stopMCPHTTP(ctx)

	// ALL operations should require confirmation (except SESSION)
	t.Run("SELECT_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
	})

	t.Run("INSERT_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000005",
				"name":  "All Confirm Test",
				"email": "all@example.com",
			},
		})
	})

	t.Run("DELETE_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "DELETE",
			"keyspace":  "test_mcp",
			"table":     "users",
			"where": []any{
				map[string]any{
					"column":   "id",
					"operator": "=",
					"value":    "00000000-0000-0000-0000-000000000005",
				},
			},
		})
	})

	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
	})

	t.Run("DROP_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "DROP",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
	})

	t.Run("GRANT_requires_confirmation", func(t *testing.T) {
		assertRequiresConfirmation(t, ctx, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
	})
}
