package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetMCPStatus_ReadonlyMode tests status in readonly mode
func TestGetMCPStatus_ReadonlyMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readonly.json")
	defer stopMCP(ctx)
	

	resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
	assertNotError(t, resp, "get_mcp_status should work")

	text := extractText(t, resp)
	var status map[string]any
	err := json.Unmarshal([]byte(text), &status)
	require.NoError(t, err, "Status should be valid JSON")

	// Verify top-level fields
	assert.Equal(t, "RUNNING", status["state"])
	assert.Contains(t, status, "config")
	assert.Contains(t, status, "connection")
	assert.Contains(t, status, "metrics")

	// Verify config details
	config := status["config"].(map[string]any)
	assert.Equal(t, "preset", config["mode"])
	assert.Equal(t, "readonly", config["preset_mode"])
	assert.False(t, config["disable_runtime_permission_changes"].(bool))

	// Verify connection details
	connection := status["connection"].(map[string]any)
	assert.Contains(t, connection, "contact_point")
	assert.Equal(t, "cassandra", connection["username"])

	// Verify metrics
	metrics := status["metrics"].(map[string]any)
	assert.Contains(t, metrics, "total_requests")
	assert.Contains(t, metrics, "success_rate")
}

// TestGetMCPStatus_AllModes tests status for all preset modes
func TestGetMCPStatus_AllModes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	modes := []struct {
		config       string
		expectedMode string
	}{
		{"testdata/readonly.json", "readonly"},
		{"testdata/readwrite.json", "readwrite"},
		{"testdata/dba.json", "dba"},
	}

	for _, mode := range modes {
		t.Run(mode.expectedMode, func(t *testing.T) {
			ctx := startMCPFromConfig(t, mode.config)
			defer stopMCP(ctx)
			

			resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
			if resp != nil {
				text := extractText(t, resp)
				var status map[string]any
				json.Unmarshal([]byte(text), &status)

				config := status["config"].(map[string]any)
				assert.Equal(t, mode.expectedMode, config["preset_mode"])
			}
		})
	}
}

// TestGetMCPStatus_WithConfirmQueries tests status shows confirm-queries
func TestGetMCPStatus_WithConfirmQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/dba_locked.json") // Has confirm_queries: ["dcl"]
	defer stopMCP(ctx)
	

	resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
	if resp != nil {
		text := extractText(t, resp)
		var status map[string]any
		json.Unmarshal([]byte(text), &status)

		config := status["config"].(map[string]any)
		confirmQueries := config["confirm_queries"].([]any)
		assert.Contains(t, confirmQueries, "dcl")
	}
}

// TestGetMCPStatus_PermissionLockdown tests status shows lockdown
func TestGetMCPStatus_PermissionLockdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/dba_locked.json") // Has disable_runtime_permission_changes: true
	defer stopMCP(ctx)
	

	resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
	if resp != nil {
		text := extractText(t, resp)
		var status map[string]any
		json.Unmarshal([]byte(text), &status)

		config := status["config"].(map[string]any)
		disabled := config["disable_runtime_permission_changes"].(bool)
		assert.True(t, disabled, "Status should show runtime changes disabled")
	}
}

// TestGetMCPStatus_FineGrainedMode tests status in fine-grained mode
func TestGetMCPStatus_FineGrainedMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/finegrained_skip_dql_dml.json")
	defer stopMCP(ctx)
	

	resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
	if resp != nil {
		text := extractText(t, resp)
		var status map[string]any
		json.Unmarshal([]byte(text), &status)

		config := status["config"].(map[string]any)
		assert.Equal(t, "fine-grained", config["mode"])

		skipList := config["skip_confirmation"].([]any)
		assert.Contains(t, skipList, "dql")
		assert.Contains(t, skipList, "dml")
		assert.Contains(t, skipList, "session") // Auto-added
	}
}

// TestGetMCPStatus_Metrics tests metrics are tracked
func TestGetMCPStatus_Metrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/dba.json")
	defer stopMCP(ctx)
	

	// Make a few tool calls
	callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
		"operation": "SELECT",
		"keyspace":  "test_mcp",
		"table":     "users",
	})

	callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
		"operation": "INSERT",
		"keyspace":  "test_mcp",
		"table":     "users",
	})

	// Check metrics
	resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
	if resp != nil {
		text := extractText(t, resp)
		var status map[string]any
		json.Unmarshal([]byte(text), &status)

		metrics := status["metrics"].(map[string]any)
		totalRequests := metrics["total_requests"].(float64)
		assert.GreaterOrEqual(t, totalRequests, float64(2), "Should have at least 2 requests")

		toolCalls := metrics["tool_calls"].(map[string]any)
		assert.Contains(t, toolCalls, "submit_query_plan")
	}
}
