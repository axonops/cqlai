package mcp

import (
	"testing"
	"time"
)

// TestFineGrained_SkipDQL tests skipping only DQL
func TestFineGrained_SkipDQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cmd := startCQLAIWithMCP(t, "testdata/finegrained_skip_dql.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// DQL should skip confirmation
	t.Run("SELECT_no_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work without confirmation")
	})

	// DML should require confirmation
	t.Run("INSERT_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "INSERT should require confirmation")
		assertContains(t, resp, "requires")
	})

	// DDL should require confirmation
	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
		})
		assertIsError(t, resp, "CREATE should require confirmation")
	})
}

// TestFineGrained_SkipDQL_DML tests skipping DQL and DML
func TestFineGrained_SkipDQL_DML(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cmd := startCQLAIWithMCP(t, "testdata/finegrained_skip_dql_dml.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// DQL and DML should work
	t.Run("SELECT_works", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work")
	})

	t.Run("INSERT_works", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "INSERT should work")
	})

	t.Run("DELETE_works", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "DELETE",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "DELETE should work")
	})

	// DDL should require confirmation
	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
		})
		assertIsError(t, resp, "CREATE should require confirmation")
	})

	// DCL should require confirmation
	t.Run("GRANT_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "GRANT should require confirmation")
	})
}

// TestFineGrained_SkipDQL_DML_DDL tests skipping everything except security
func TestFineGrained_SkipDQL_DML_DDL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cmd := startCQLAIWithMCP(t, "testdata/finegrained_skip_dql_dml_ddl.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// DQL, DML, DDL should all work
	t.Run("SELECT_works", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work")
	})

	t.Run("INSERT_works", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "INSERT should work")
	})

	t.Run("CREATE_works", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
		})
		assertNotError(t, resp, "CREATE should work")
	})

	// Only DCL should require confirmation
	t.Run("GRANT_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "GRANT should require confirmation")
	})
}

// TestFineGrained_SkipALL tests skip ALL (no confirmations)
func TestFineGrained_SkipALL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cmd := startCQLAIWithMCP(t, "testdata/finegrained_skip_all.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// Everything should work without confirmation
	allOps := []string{"SELECT", "INSERT", "DELETE", "CREATE", "DROP", "GRANT"}

	for _, op := range allOps {
		t.Run(op+"_works", func(t *testing.T) {
			resp := callTool(t, "submit_query_plan", map[string]any{
				"operation": op,
				"keyspace":  "test_mcp",
				"table":     "users",
			})
			assertNotError(t, resp, op+" should work with skip ALL")
		})
	}
}

// TestFineGrained_SkipNone tests skip none (confirm everything)
func TestFineGrained_SkipNone(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cmd := startCQLAIWithMCP(t, "testdata/finegrained_skip_none.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// Everything should require confirmation (except SESSION)
	allOps := []string{"SELECT", "INSERT", "DELETE", "CREATE", "DROP", "GRANT"}

	for _, op := range allOps {
		t.Run(op+"_requires_confirmation", func(t *testing.T) {
			resp := callTool(t, "submit_query_plan", map[string]any{
				"operation": op,
				"keyspace":  "test_mcp",
				"table":     "users",
			})
			assertIsError(t, resp, op+" should require confirmation with skip none")
			assertContains(t, resp, "requires")
		})
	}
}
