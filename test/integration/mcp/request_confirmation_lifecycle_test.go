package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfirmationLifecycle_CreateAndPending tests creating confirmation requests
func TestConfirmationLifecycle_CreateAndPending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start with DBA mode + confirm on DCL
	ctx := startMCPFromConfig(t, "testdata/dba_locked.json") // Has confirm_queries: ["dcl"]
	defer stopMCP(ctx)
	

	// Submit DCL operation (should create confirmation request)
	t.Run("create_request", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
		assertIsError(t, resp, "Should require confirmation")
		text := extractText(t, resp)
		assert.Contains(t, text, "req_")
		assert.Contains(t, text, "requires user confirmation")
	})

	// Get pending confirmations
	t.Run("get_pending", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_pending_confirmations", map[string]any{})
		if resp != nil {
			assertNotError(t, resp, "get_pending should succeed")
			text := extractText(t, resp)

			var pending []map[string]any
			err := json.Unmarshal([]byte(text), &pending)
			if err == nil && len(pending) > 0 {
				req := pending[0]
				assert.Equal(t, "PENDING", req["status"])
				assert.Contains(t, req, "request_id")
				assert.Contains(t, req, "query")
				assert.Contains(t, req, "severity")
				assert.Contains(t, req, "operation")
			}
		}
	})
}

// TestConfirmationLifecycle_GetConfirmationState tests getting specific request state
func TestConfirmationLifecycle_GetConfirmationState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/dba_confirm_all.json") // Confirm ALL
	defer stopMCP(ctx)
	

	// Create a request
	resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
		"operation": "DELETE",
		"keyspace":  "test_mcp",
		"table":     "users",
		"where": []any{
			map[string]any{
				"column":   "id",
				"operator": "=",
				"value":    "00000000-0000-0000-0000-000000000050",
			},
		},
	})

	text := extractText(t, resp)
	requestID := extractRequestID(text)

	if requestID != "" {
		t.Run("get_state_by_id", func(t *testing.T) {
			stateResp := callTool(t, ctx.SocketPath, "get_confirmation_state", map[string]any{
				"request_id": requestID,
			})

			if stateResp != nil {
				assertNotError(t, stateResp, "get_confirmation_state should work")
				stateText := extractText(t, stateResp)

				var state map[string]any
				err := json.Unmarshal([]byte(stateText), &state)
				if err == nil {
					assert.Equal(t, requestID, state["request_id"])
					assert.Equal(t, "PENDING", state["status"])
					assert.Contains(t, state, "query")
					assert.Contains(t, state, "classification")
				}
			}
		})
	}
}

// TestConfirmationLifecycle_CancelRequest tests cancelling requests
func TestConfirmationLifecycle_CancelRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/dba_confirm_all.json")
	defer stopMCP(ctx)
	

	// Create a request
	resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
		"operation": "TRUNCATE",
		"keyspace":  "test_mcp",
		"table":     "users",
	})

	text := extractText(t, resp)
	requestID := extractRequestID(text)

	if requestID != "" {
		// Cancel the request
		t.Run("cancel", func(t *testing.T) {
			cancelResp := callTool(t, ctx.SocketPath, "cancel_confirmation", map[string]any{
				"request_id": requestID,
				"reason":     "Testing cancellation",
			})

			if cancelResp != nil {
				assertNotError(t, cancelResp, "cancel should succeed")
				cancelText := extractText(t, cancelResp)
				assert.Contains(t, cancelText, "cancelled successfully")
			}
		})

		// Verify it appears in cancelled list
		t.Run("verify_in_cancelled_list", func(t *testing.T) {
			cancelledResp := callTool(t, ctx.SocketPath, "get_cancelled_confirmations", map[string]any{})
			if cancelledResp != nil {
				assertNotError(t, cancelledResp, "get_cancelled should work")
				text := extractText(t, cancelledResp)

				var cancelled []map[string]any
				json.Unmarshal([]byte(text), &cancelled)

				found := false
				for _, req := range cancelled {
					if req["request_id"] == requestID {
						found = true
						assert.Equal(t, "CANCELLED", req["status"])
						break
					}
				}
				assert.True(t, found, "Cancelled request should appear in list")
			}
		})

		// Verify no longer in pending list
		t.Run("not_in_pending", func(t *testing.T) {
			pendingResp := callTool(t, ctx.SocketPath, "get_pending_confirmations", map[string]any{})
			if pendingResp != nil {
				text := extractText(t, pendingResp)
				assert.NotContains(t, text, requestID, "Should not be in pending after cancel")
			}
		})
	}
}

// TestConfirmationLifecycle_MultipleRequests tests multiple concurrent requests
func TestConfirmationLifecycle_MultipleRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/dba_confirm_dml_ddl.json") // Confirm DML + DDL
	defer stopMCP(ctx)
	

	// Create multiple requests
	t.Run("INSERT_request", func(t *testing.T) {
		callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000051",
				"name":  "Test User",
				"email": "test@example.com",
			},
		})
	})
	t.Run("UPDATE_request", func(t *testing.T) {
		callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "UPDATE",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{"name": "Updated Name"},
			"where": []any{
				map[string]any{
					"column":   "id",
					"operator": "=",
					"value":    "00000000-0000-0000-0000-000000000051",
				},
			},
		})
	})
	t.Run("DELETE_request", func(t *testing.T) {
		callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "DELETE",
			"keyspace":  "test_mcp",
			"table":     "users",
			"where": []any{
				map[string]any{
					"column":   "id",
					"operator": "=",
					"value":    "00000000-0000-0000-0000-000000000051",
				},
			},
		})
	})
	t.Run("CREATE_request", func(t *testing.T) {
		callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "test_logs_lifecycle",
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
	})
	t.Run("DROP_request", func(t *testing.T) {
		callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "DROP",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
	})

	// Get pending - should have all 5
	t.Run("multiple_pending", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_pending_confirmations", map[string]any{})
		if resp != nil {
			text := extractText(t, resp)
			var pending []map[string]any
			json.Unmarshal([]byte(text), &pending)

			assert.GreaterOrEqual(t, len(pending), 5, "Should have at least 5 pending")
		}
	})
}

// TestConfirmationLifecycle_EmptyLists tests getting confirmations when empty
func TestConfirmationLifecycle_EmptyLists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start fresh server
	ctx := startMCPFromConfig(t, "testdata/readonly.json")
	defer stopMCP(ctx)
	

	// All lists should be empty
	tools := []string{
		"get_pending_confirmations",
		"get_approved_confirmations",
		"get_denied_confirmations",
		"get_cancelled_confirmations",
	}

	for _, tool := range tools {
		t.Run(tool, func(t *testing.T) {
			resp := callTool(t, ctx.SocketPath, tool, map[string]any{})
			if resp != nil {
				assertNotError(t, resp, tool+" should work")
				text := extractText(t, resp)
				var list []any
				json.Unmarshal([]byte(text), &list)
				assert.Equal(t, 0, len(list), tool+" should be empty on fresh server")
			}
		})
	}
}
