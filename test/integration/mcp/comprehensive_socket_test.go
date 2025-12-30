package mcp

// DEPRECATED: This file is superseded by the focused test files:
// - permission_enforcement_preset_modes_test.go
// - permission_enforcement_confirm_queries_test.go
// - permission_enforcement_finegrained_test.go
// - permission_enforcement_lockdown_test.go
// - permission_enforcement_runtime_changes_test.go
// - status_test.go
// - request_confirmation_lifecycle_test.go
// - all_operations_matrix_test.go
// - cql_query_validation_test.go
//
// These tests don't use auto-start and were written before the complete redesign.
// They are kept for reference but should eventually be removed.

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// callMCPTool calls an MCP tool via the Unix socket and returns the response
func callMCPTool(t *testing.T, toolName string, args map[string]any) map[string]any {
	requestID := time.Now().UnixNano()

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      requestID,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	}

	requestJSON, err := json.Marshal(request)
	require.NoError(t, err)

	// Call via nc
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo '%s' | nc -U /tmp/cqlai-mcp.sock", string(requestJSON)))
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to call MCP tool %s: %v", toolName, err)

	// Parse response
	var response map[string]any
	err = json.Unmarshal(output, &response)
	require.NoError(t, err, "Failed to parse response: %s", string(output))

	return response
}

// Test get_mcp_status tool
func TestSocket_GetMCPStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	response := callMCPTool(t, "get_mcp_status", map[string]any{})

	// Check response structure
	result, ok := response["result"].(map[string]any)
	require.True(t, ok, "Response should have result")

	content, ok := result["content"].([]any)
	require.True(t, ok)
	require.Greater(t, len(content), 0)

	textContent := content[0].(map[string]any)
	text := textContent["text"].(string)

	// Parse the JSON status
	var status map[string]any
	err := json.Unmarshal([]byte(text), &status)
	require.NoError(t, err)

	// Verify status fields
	assert.Equal(t, "RUNNING", status["state"])
	assert.Contains(t, status, "config")
	assert.Contains(t, status, "connection")
	assert.Contains(t, status, "metrics")
}

// Test get_pending_confirmations tool
func TestSocket_GetPendingConfirmations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	response := callMCPTool(t, "get_pending_confirmations", map[string]any{})

	// Should return a list (possibly empty)
	result, ok := response["result"].(map[string]any)
	require.True(t, ok)

	content, ok := result["content"].([]any)
	require.True(t, ok)
	require.Greater(t, len(content), 0)

	// Parse the list
	textContent := content[0].(map[string]any)
	text := textContent["text"].(string)

	var confirmations []any
	err := json.Unmarshal([]byte(text), &confirmations)
	require.NoError(t, err)

	// Should be valid array (may be empty)
	assert.NotNil(t, confirmations)
}

// TestSocket_PermissionDenied_ReadonlyMode tests operations blocked in readonly mode
func TestSocket_PermissionDenied_ReadonlyMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// These operations should be blocked in readonly mode
	blockedOps := []struct {
		operation string
		keyspace  string
		table     string
	}{
		{"INSERT", "test_mcp", "users"},
		{"UPDATE", "test_mcp", "users"},
		{"DELETE", "test_mcp", "users"},
		{"CREATE", "test_mcp", "logs"},
		{"DROP", "test_mcp", "users"},
		{"TRUNCATE", "test_mcp", "users"},
		{"GRANT", "test_mcp", "users"},
	}

	for _, op := range blockedOps {
		t.Run(op.operation, func(t *testing.T) {
			response := callMCPTool(t, "submit_query_plan", map[string]any{
				"operation": op.operation,
				"keyspace":  op.keyspace,
				"table":     op.table,
			})

			// Should be an error
			result, ok := response["result"].(map[string]any)
			require.True(t, ok)

			isError, _ := result["isError"].(bool)
			assert.True(t, isError, "%s should be blocked in readonly mode", op.operation)

			// Check error contains helpful message
			content := result["content"].([]any)
			textContent := content[0].(map[string]any)
			errorText := textContent["text"].(string)

			assert.Contains(t, errorText, "not allowed")
			assert.Contains(t, errorText, "readonly")
		})
	}
}

// TestSocket_AllowedOperations_ReadonlyMode tests allowed operations in readonly
func TestSocket_AllowedOperations_ReadonlyMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// These should be allowed in readonly mode
	response := callMCPTool(t, "submit_query_plan", map[string]any{
		"operation": "SELECT",
		"keyspace":  "test_mcp",
		"table":     "users",
		"columns":   []string{"*"},
	})

	// Should succeed
	result, ok := response["result"].(map[string]any)
	require.True(t, ok)

	isError, _ := result["isError"].(bool)
	assert.False(t, isError, "SELECT should be allowed in readonly mode")
}

// TestSocket_UpdatePermissions_UserConfirmedRequired tests user_confirmed enforcement
func TestSocket_UpdatePermissions_UserConfirmedRequired(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Try without user_confirmed
	response := callMCPTool(t, "update_mcp_permissions", map[string]any{
		"mode":           "readwrite",
		"user_confirmed": false,
	})

	// Should be error
	result, ok := response["result"].(map[string]any)
	require.True(t, ok)

	isError, _ := result["isError"].(bool)
	assert.True(t, isError, "Should require user_confirmed=true")

	content := result["content"].([]any)
	textContent := content[0].(map[string]any)
	errorText := textContent["text"].(string)

	assert.Contains(t, errorText, "requires user confirmation")
}

// TestSocket_UpdatePermissions_ModeChange tests changing mode via socket
func TestSocket_UpdatePermissions_ModeChange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Change to readwrite mode
	response := callMCPTool(t, "update_mcp_permissions", map[string]any{
		"mode":           "readwrite",
		"user_confirmed": true,
	})

	// Should succeed
	result, ok := response["result"].(map[string]any)
	require.True(t, ok)

	isError, _ := result["isError"].(bool)
	assert.False(t, isError, "Mode change should succeed")

	// Verify status shows new mode
	statusResp := callMCPTool(t, "get_mcp_status", map[string]any{})
	statusResult := statusResp["result"].(map[string]any)
	statusContent := statusResult["content"].([]any)
	statusText := statusContent[0].(map[string]any)["text"].(string)

	var status map[string]any
	json.Unmarshal([]byte(statusText), &status)

	config := status["config"].(map[string]any)
	assert.Equal(t, "readwrite", config["preset_mode"])
}

// TestSocket_ConfirmationRequired_WithRequestID tests confirmation flow
func TestSocket_ConfirmationRequired_WithRequestID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Set confirmations on DML
	callMCPTool(t, "update_mcp_permissions", map[string]any{
		"confirm_queries": "dml",
		"user_confirmed":  true,
	})

	// Try INSERT (should require confirmation)
	response := callMCPTool(t, "submit_query_plan", map[string]any{
		"operation": "INSERT",
		"keyspace":  "test_mcp",
		"table":     "users",
	})

	// Should be error with confirmation required
	result := response["result"].(map[string]any)
	isError, _ := result["isError"].(bool)
	assert.True(t, isError)

	content := result["content"].([]any)
	errorText := content[0].(map[string]any)["text"].(string)

	// Should contain request ID
	assert.Contains(t, errorText, "req_")
	assert.Contains(t, errorText, "requires user confirmation")

	// Extract request ID
	if strings.Contains(errorText, "req_") {
		// Get pending confirmations to verify request exists
		pendingResp := callMCPTool(t, "get_pending_confirmations", map[string]any{})
		pendingResult := pendingResp["result"].(map[string]any)
		pendingContent := pendingResult["content"].([]any)
		pendingText := pendingContent[0].(map[string]any)["text"].(string)

		var pending []map[string]any
		json.Unmarshal([]byte(pendingText), &pending)

		assert.Greater(t, len(pending), 0, "Should have at least one pending request")

		if len(pending) > 0 {
			request := pending[0]
			assert.Contains(t, request, "request_id")
			assert.Equal(t, "PENDING", request["status"])
			assert.Equal(t, "INSERT", request["operation"])
		}
	}
}
