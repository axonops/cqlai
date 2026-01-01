package mcp

import (
	"os"
	"testing"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/stretchr/testify/assert"
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
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000001",
				"name":  "Test User",
				"email": "test@example.com",
			},
		})
		assertIsError(t, resp, "INSERT should require confirmation")
		assertContains(t, resp, "requires")

		// Verify request created and tracked
		text := extractText(t, resp)
		requestID := extractRequestID(text)
		if requestID != "" {
			// Check it's in pending list
			pendingResp := callToolHTTP(t, ctx, "get_pending_confirmations", map[string]any{})
			if pendingResp != nil {
				assertContains(t, pendingResp, requestID)
			}

			// Check specific state
			stateResp := callToolHTTP(t, ctx, "get_confirmation_state", map[string]any{"request_id": requestID})
			if stateResp != nil {
				assertNotError(t, stateResp, "get_confirmation_state should work")
				assertContains(t, stateResp, "PENDING")
				assertContains(t, stateResp, "INSERT")
			}
		}
	})

	t.Run("DELETE_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
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
		assertIsError(t, resp, "DELETE should require confirmation")
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
		assertIsError(t, resp, "CREATE should require confirmation")
		assertContains(t, resp, "requires")
	})

	t.Run("DROP_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "DROP",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "DROP should require confirmation")
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
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000003",
				"name":  "Multi Test User",
				"email": "multi@example.com",
			},
		})
		assertIsError(t, resp, "INSERT should require confirmation")
	})

	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
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
		assertIsError(t, resp, "CREATE should require confirmation")
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
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "SELECT should require confirmation with overlay")
		assertContains(t, resp, "requires")
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
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "SELECT should require confirmation with ALL")
		text := extractText(t, resp)
		assert.Contains(t, text, "requires")
		assert.Contains(t, text, "req_")
	})

	t.Run("INSERT_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000005",
				"name":  "All Confirm Test",
				"email": "all@example.com",
			},
		})
		assertIsError(t, resp, "INSERT should require confirmation with ALL")
		text := extractText(t, resp)
		assert.Contains(t, text, "requires")
		assert.Contains(t, text, "req_")
	})

	t.Run("DELETE_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
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
		assertIsError(t, resp, "DELETE should require confirmation with ALL")
		text := extractText(t, resp)
		assert.Contains(t, text, "requires")
		assert.Contains(t, text, "req_")
	})

	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
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
		assertIsError(t, resp, "CREATE should require confirmation with ALL")
		text := extractText(t, resp)
		assert.Contains(t, text, "requires")
		assert.Contains(t, text, "req_")
	})

	t.Run("DROP_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "DROP",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "DROP should require confirmation with ALL")
		text := extractText(t, resp)
		assert.Contains(t, text, "requires")
		assert.Contains(t, text, "req_")
	})

	t.Run("GRANT_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
		assertIsError(t, resp, "GRANT should require confirmation with ALL")
		text := extractText(t, resp)
		assert.Contains(t, text, "requires")
		assert.Contains(t, text, "req_")
	})
}
