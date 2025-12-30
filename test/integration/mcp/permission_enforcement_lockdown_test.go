package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLockdown_ReadonlyMode tests lockdown in readonly mode
func TestLockdown_ReadonlyMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readonly_locked.json")
	defer stopMCP(ctx)

	// Verify status shows lockdown
	t.Run("status_shows_disabled", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
		if resp != nil {
			text := extractText(t, resp)
			var status map[string]any
			json.Unmarshal([]byte(text), &status)

			config := status["config"].(map[string]any)
			disabled := config["disable_runtime_permission_changes"].(bool)
			assert.True(t, disabled, "Lockdown should be enabled")
		}
	})

	// Try to change mode (should be blocked)
	t.Run("mode_change_blocked", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "readwrite",
			"user_confirmed": true,
		})
		assertIsError(t, resp, "Mode change should be blocked")
		assertContains(t, resp, "disabled")
		assertContains(t, resp, "locked")
	})

	// Try to change confirm-queries (should be blocked)
	t.Run("confirm_queries_blocked", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"confirm_queries": "dql",
			"user_confirmed":  true,
		})
		assertIsError(t, resp, "Confirm-queries change should be blocked")
	})

	// Try to change skip-confirmation (should be blocked)
	t.Run("skip_confirmation_blocked", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"skip_confirmation": "dql,dml",
			"user_confirmed":    true,
		})
		assertIsError(t, resp, "Skip-confirmation change should be blocked")
	})

	// But operations should still work normally
	t.Run("operations_still_work", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should still work when locked")
	})
}

// TestLockdown_AllModes tests lockdown works with all modes
func TestLockdown_AllModes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	configs := []struct {
		file string
		mode string
	}{
		{"testdata/readonly_locked.json", "readonly"},
		{"testdata/readwrite_locked.json", "readwrite"},
		{"testdata/dba_locked.json", "dba"},
	}

	for _, cfg := range configs {
		t.Run(cfg.mode+"_locked", func(t *testing.T) {
			ctx := startMCPFromConfig(t, cfg.file)
			defer stopMCP(ctx)

			// Verify lockdown in status
			resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
			if resp != nil {
				text := extractText(t, resp)
				var status map[string]any
				json.Unmarshal([]byte(text), &status)

				config := status["config"].(map[string]any)
				assert.True(t, config["disable_runtime_permission_changes"].(bool))
			}

			// Verify update_mcp_permissions is blocked
			updateResp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
				"mode":           "dba",
				"user_confirmed": true,
			})
			assertIsError(t, updateResp, "Updates should be blocked")
		})
	}
}

// TestLockdown_ErrorMessage tests lockdown error message quality
func TestLockdown_ErrorMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readonly_locked.json")
	defer stopMCP(ctx)

	resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
		"mode":           "dba",
		"user_confirmed": true,
	})

	text := extractText(t, resp)

	// Error should explain what's wrong and how to fix
	requiredStrings := []string{
		"disabled",
		"locked",
		"--disable-runtime-permission-changes",
		"stop the server",
		"restart",
	}

	for _, s := range requiredStrings {
		assert.Contains(t, text, s, "Error should contain: "+s)
	}
}
