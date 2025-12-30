package mcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)
// TestConfirmQueries_ReadwriteWithDML tests readwrite + confirm dml
func TestConfirmQueries_ReadwriteWithDML(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cmd := startCQLAIWithMCP(t, "testdata/readwrite_confirm_dml.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// DQL should work without confirmation
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

	t.Run("DELETE_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "DELETE",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "DELETE should require confirmation")
	})

	// DDL still blocked (not allowed in readwrite)
	t.Run("CREATE_blocked", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
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

	cmd := startCQLAIWithMCP(t, "testdata/dba_confirm_ddl.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// DML should work without confirmation
	t.Run("INSERT_no_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "INSERT should work without confirmation in DBA")
	})

	// DDL should require confirmation
	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
		})
		assertIsError(t, resp, "CREATE should require confirmation")
		assertContains(t, resp, "requires")
	})

	t.Run("DROP_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "DROP",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "DROP should require confirmation")
	})

	// DCL should work without confirmation
	t.Run("GRANT_no_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "GRANT should work without confirmation")
	})
}

// TestConfirmQueries_DBA_ConfirmMultiple tests dba + confirm dml,ddl
func TestConfirmQueries_DBA_ConfirmMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cmd := startCQLAIWithMCP(t, "testdata/dba_confirm_dml_ddl.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// DQL should work
	t.Run("SELECT_works", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work")
	})

	// Both DML and DDL should require confirmation
	t.Run("INSERT_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "INSERT should require confirmation")
	})

	t.Run("CREATE_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
		})
		assertIsError(t, resp, "CREATE should require confirmation")
	})

	// DCL should work
	t.Run("GRANT_works", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "GRANT should work without confirmation")
	})
}

// TestConfirmQueries_ReadonlyWithDQL tests readonly + confirm dql (overlay on allowed)
func TestConfirmQueries_ReadonlyWithDQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cmd := startCQLAIWithMCP(t, "testdata/readonly_confirm_dql.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// Even SELECT should require confirmation
	t.Run("SELECT_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "SELECT should require confirmation with overlay")
		assertContains(t, resp, "requires")
	})

	// DML still blocked (not in allowed list)
	t.Run("INSERT_still_blocked", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
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

	cmd := startCQLAIWithMCP(t, "testdata/dba_confirm_all.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// ALL operations should require confirmation (except SESSION)
	operations := []string{"SELECT", "INSERT", "DELETE", "CREATE", "DROP", "GRANT"}

	for _, op := range operations {
		t.Run(op+"_requires_confirmation", func(t *testing.T) {
			resp := callTool(t, "submit_query_plan", map[string]any{
				"operation": op,
				"keyspace":  "test_mcp",
				"table":     "users",
			})
			assertIsError(t, resp, op+" should require confirmation with ALL")
			text := extractText(t, resp)
			assert.Contains(t, text, "requires")
			assert.Contains(t, text, "req_") // Should have request ID
		})
	}
}
