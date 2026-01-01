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

// TestFineGrained_SkipDQL tests skipping only DQL
func TestFineGrained_SkipDQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfigHTTP(t, "testdata/finegrained_skip_dql.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// DQL should skip confirmation
	t.Run("SELECT_no_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work without confirmation")
	})

	// DML should require confirmation (streaming - returns notification immediately, then blocks)
	t.Run("INSERT_requires_confirmation", func(t *testing.T) {
		// Skip this test - INSERT now BLOCKS waiting for confirmation (streaming behavior)
		// To properly test, would need concurrent thread to approve (see TestHTTP_StreamingConfirmation)
		t.Skip("INSERT now uses streaming confirmation (blocks until approved) - see TestHTTP_StreamingConfirmation for proper pattern")
	})

	// DDL should require confirmation
	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "test_logs_skipdql",
			"options": map[string]any{
				"if_not_exists": true,
			},
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
		assertIsError(t, resp, "CREATE should require confirmation")
	})
}

// TestFineGrained_SkipDQL_DML tests skipping DQL and DML
func TestFineGrained_SkipDQL_DML(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfigHTTP(t, "testdata/finegrained_skip_dql_dml.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// DQL and DML should work
	t.Run("SELECT_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work")
	})

	t.Run("INSERT_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000002",
				"name":  "Test User 2",
				"email": "test2@example.com",
			},
		})
		assertNotError(t, resp, "INSERT should work")
	})

	t.Run("DELETE_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "DELETE",
			"keyspace":  "test_mcp",
			"table":     "users",
			"where": []any{
				map[string]any{
					"column":   "id",
					"operator": "=",
					"value":    "00000000-0000-0000-0000-000000000002",
				},
			},
		})
		assertNotError(t, resp, "DELETE should work")
	})

	// DDL should require confirmation
	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "test_logs_dqlml",
			"options": map[string]any{
				"if_not_exists": true,
			},
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
		assertIsError(t, resp, "CREATE should require confirmation")
	})

	// DCL should require confirmation
	t.Run("GRANT_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
		assertIsError(t, resp, "GRANT should require confirmation")
	})
}

// TestFineGrained_SkipDQL_DML_DDL tests skipping everything except security
func TestFineGrained_SkipDQL_DML_DDL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfigHTTP(t, "testdata/finegrained_skip_dql_dml_ddl.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// DQL, DML, DDL should all work
	t.Run("SELECT_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work")
	})

	t.Run("INSERT_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000003",
				"name":  "Test User 3",
				"email": "test3@example.com",
			},
		})
		assertNotError(t, resp, "INSERT should work")
	})

	t.Run("CREATE_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "test_logs_dmlddl",
			"options": map[string]any{
				"if_not_exists": true,
			},
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
		assertNotError(t, resp, "CREATE should work")
	})

	// Only DCL should require confirmation
	t.Run("GRANT_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
		assertIsError(t, resp, "GRANT should require confirmation")
	})
}

// TestFineGrained_SkipALL tests skip ALL (no confirmations)
func TestFineGrained_SkipALL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfigHTTP(t, "testdata/finegrained_skip_all.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// Everything should work without confirmation
	t.Run("SELECT_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work with skip ALL")
	})

	t.Run("INSERT_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000004",
				"name":  "Test User 4",
				"email": "test4@example.com",
			},
		})
		assertNotError(t, resp, "INSERT should work with skip ALL")
	})

	t.Run("DELETE_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "DELETE",
			"keyspace":  "test_mcp",
			"table":     "users",
			"where": []any{
				map[string]any{
					"column":   "id",
					"operator": "=",
					"value":    "00000000-0000-0000-0000-000000000004",
				},
			},
		})
		assertNotError(t, resp, "DELETE should work with skip ALL")
	})

	t.Run("CREATE_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "test_table_skipall",
			"options": map[string]any{
				"if_not_exists": true,
			},
			"schema": map[string]any{
				"id":   "uuid PRIMARY KEY",
				"data": "text",
			},
		})
		assertNotError(t, resp, "CREATE should work with skip ALL")
	})

	t.Run("DROP_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "DROP",
			"keyspace":  "test_mcp",
			"table":     "test_table_skipall",
		})
		assertNotError(t, resp, "DROP should work with skip ALL")
	})

	t.Run("GRANT_works", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
		assertNotError(t, resp, "GRANT should work with skip ALL")
	})
}

// TestFineGrained_SkipNone tests skip none (confirm everything)
func TestFineGrained_SkipNone(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfigHTTP(t, "testdata/finegrained_skip_none.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// Everything should require confirmation (except SESSION)
	t.Run("SELECT_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "SELECT should require confirmation with skip none")
		assertContains(t, resp, "requires")
	})

	t.Run("INSERT_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000005",
				"name":  "Test User 5",
				"email": "test5@example.com",
			},
		})
		assertIsError(t, resp, "INSERT should require confirmation with skip none")
		assertContains(t, resp, "requires")
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
		assertIsError(t, resp, "DELETE should require confirmation with skip none")
		assertContains(t, resp, "requires")
	})

	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "test_table_skipnone",
			"options": map[string]any{
				"if_not_exists": true,
			},
			"schema": map[string]any{
				"id":   "uuid PRIMARY KEY",
				"data": "text",
			},
		})
		assertIsError(t, resp, "CREATE should require confirmation with skip none")
		assertContains(t, resp, "requires")
	})

	t.Run("DROP_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "DROP",
			"keyspace":  "test_mcp",
			"table":     "test_table_skipnone",
		})
		assertIsError(t, resp, "DROP should require confirmation with skip none")
		assertContains(t, resp, "requires")
	})

	t.Run("GRANT_requires_confirmation", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
		assertIsError(t, resp, "GRANT should require confirmation with skip none")
		assertContains(t, resp, "requires")
	})
}
