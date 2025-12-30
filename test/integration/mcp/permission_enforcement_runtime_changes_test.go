package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRuntimeChanges_ReadonlyToReadwrite tests escalating from readonly to readwrite
func TestRuntimeChanges_ReadonlyToReadwrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readonly.json")
	defer stopMCP(ctx)
	

	// Verify INSERT blocked
	t.Run("step1_INSERT_blocked", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000040",
				"name":  "Test User",
				"email": "test@example.com",
			},
		})
		assertIsError(t, resp, "INSERT should be blocked")
		assertContains(t, resp, "readwrite") // Should suggest upgrade
	})

	// Change to readwrite
	t.Run("step2_upgrade_to_readwrite", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "readwrite",
			"user_confirmed": true,
		})
		assertNotError(t, resp, "Mode change should succeed")
	})

	// Verify INSERT now works
	t.Run("step3_INSERT_now_works", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000041",
				"name":  "Test User 2",
				"email": "test2@example.com",
			},
		})
		assertNotError(t, resp, "INSERT should work after upgrade")
	})

	// Verify status updated
	t.Run("step4_verify_status", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
		if resp != nil {
			text := extractText(t, resp)
			var status map[string]any
			json.Unmarshal([]byte(text), &status)

			config := status["config"].(map[string]any)
			assert.Equal(t, "readwrite", config["preset_mode"])
		}
	})
}

// TestRuntimeChanges_ReadwriteToDBA tests escalating to DBA mode
func TestRuntimeChanges_ReadwriteToDBA(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readwrite.json")
	defer stopMCP(ctx)
	

	// Verify CREATE blocked
	t.Run("CREATE_blocked_in_readwrite", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "test_logs_runtime",
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
		assertIsError(t, resp, "CREATE should be blocked")
		assertContains(t, resp, "dba") // Should suggest DBA
	})

	// Upgrade to DBA
	t.Run("upgrade_to_dba", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "dba",
			"user_confirmed": true,
		})
		assertNotError(t, resp, "Upgrade should succeed")
	})

	// CREATE now works
	t.Run("CREATE_now_works", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "test_logs_runtime",
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
		assertNotError(t, resp, "CREATE should work in DBA")
	})
}

// TestRuntimeChanges_AddConfirmQueries tests adding confirm-queries overlay
func TestRuntimeChanges_AddConfirmQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/dba.json") // DBA without confirmations
	defer stopMCP(ctx)
	

	// Initially no confirmations
	t.Run("GRANT_works_initially", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
		assertNotError(t, resp, "GRANT should work without confirmation initially")
	})

	// Add DCL confirmations
	t.Run("add_dcl_confirmations", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"confirm_queries": "dcl",
			"user_confirmed":  true,
		})
		assertNotError(t, resp, "Adding confirmations should succeed")
	})

	// Now GRANT requires confirmation
	t.Run("GRANT_now_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
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

// TestRuntimeChanges_DisableConfirmations tests disabling confirmations
func TestRuntimeChanges_DisableConfirmations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/dba_confirm_all.json") // Starts with ALL
	defer stopMCP(ctx)
	

	// Initially requires confirmation
	t.Run("SELECT_requires_confirmation_initially", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "SELECT should require confirmation with ALL")
	})

	// Disable confirmations
	t.Run("disable_confirmations", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"confirm_queries": "disable",
			"user_confirmed":  true,
		})
		assertNotError(t, resp, "Disabling should succeed")
	})

	// Now works without confirmation
	t.Run("SELECT_works_after_disable", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work after disable")
	})
}

// TestRuntimeChanges_PresetToFineGrained tests switching mode types
func TestRuntimeChanges_PresetToFineGrained(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readonly.json") // Preset mode
	defer stopMCP(ctx)
	

	// Verify preset mode
	t.Run("verify_preset_mode", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
		if resp != nil {
			text := extractText(t, resp)
			assert.Contains(t, text, "preset")
		}
	})

	// Switch to fine-grained
	t.Run("switch_to_finegrained", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"skip_confirmation": "dql,dml",
			"user_confirmed":    true,
		})
		assertNotError(t, resp, "Switch to fine-grained should succeed")
	})

	// Verify fine-grained mode
	t.Run("verify_finegrained_mode", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
		if resp != nil {
			text := extractText(t, resp)
			var status map[string]any
			json.Unmarshal([]byte(text), &status)

			config := status["config"].(map[string]any)
			assert.Equal(t, "fine-grained", config["mode"])
		}
	})

	// Switch back to preset
	t.Run("switch_back_to_preset", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "dba",
			"user_confirmed": true,
		})
		assertNotError(t, resp, "Switch back should succeed")
	})
}

// TestRuntimeChanges_UserConfirmedRequired tests user_confirmed enforcement
func TestRuntimeChanges_UserConfirmedRequired(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readonly.json")
	defer stopMCP(ctx)
	

	// Without user_confirmed should fail
	t.Run("without_user_confirmed", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "dba",
			"user_confirmed": false,
		})
		assertIsError(t, resp, "Should require user_confirmed=true")
		assertContains(t, resp, "requires user confirmation")
	})

	// With user_confirmed should succeed
	t.Run("with_user_confirmed", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "dba",
			"user_confirmed": true,
		})
		assertNotError(t, resp, "Should succeed with user_confirmed")
	})
}
