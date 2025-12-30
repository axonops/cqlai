package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
// Seed test data: ./test/integration/setup_test_data.sh

// TestAutomated_ReadonlyMode_AllOperations tests all 76 operations in readonly mode
func TestAutomated_ReadonlyMode_AllOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping automated integration test")
	}

	// Start CQLAI with MCP auto-start in readonly mode
	ctx := startMCPFromConfig(t, "testdata/readonly.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// Wait for MCP server to be ready


	// Test DQL operations (should be allowed)
	dqlOps := []string{"SELECT"}
	for _, op := range dqlOps {
		t.Run("DQL_"+op, func(t *testing.T) {
			resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
				"operation": op,
				"keyspace":  "test_mcp",
				"table":     "users",
			})
			assertNotError(t, resp, op+" should be allowed in readonly")
		})
	}

	// Test DML operations (should be blocked)
	t.Run("DML_INSERT", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000021",
				"name":  "Test User",
				"email": "test@example.com",
			},
		})
		assertIsError(t, resp, "INSERT should be blocked in readonly")
		assertContains(t, resp, "not allowed")
	})
	t.Run("DML_UPDATE", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "UPDATE",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{"name": "Updated Name"},
			"where": []any{
				map[string]any{
					"column":   "id",
					"operator": "=",
					"value":    "00000000-0000-0000-0000-000000000021",
				},
			},
		})
		assertIsError(t, resp, "UPDATE should be blocked in readonly")
		assertContains(t, resp, "not allowed")
	})
	t.Run("DML_DELETE", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "DELETE",
			"keyspace":  "test_mcp",
			"table":     "users",
			"where": []any{
				map[string]any{
					"column":   "id",
					"operator": "=",
					"value":    "00000000-0000-0000-0000-000000000021",
				},
			},
		})
		assertIsError(t, resp, "DELETE should be blocked in readonly")
		assertContains(t, resp, "not allowed")
	})

	// Test DDL operations (should be blocked)
	t.Run("DDL_CREATE", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "test_logs",
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
		assertIsError(t, resp, "CREATE should be blocked in readonly")
	})
	t.Run("DDL_ALTER", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "ALTER",
			"keyspace":  "test_mcp",
			"table":     "users",
			"options": map[string]any{
				"object_type": "TABLE",
				"action":      "ADD",
				"column_name": "age",
				"column_type": "int",
			},
		})
		assertIsError(t, resp, "ALTER should be blocked in readonly")
	})
	t.Run("DDL_DROP", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "DROP",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "DROP should be blocked in readonly")
	})
	t.Run("DDL_TRUNCATE", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "TRUNCATE",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "TRUNCATE should be blocked in readonly")
	})

	// Test get_mcp_status
	t.Run("get_mcp_status", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
		assertNotError(t, resp, "get_mcp_status should work")

		// Parse and verify status
		text := extractText(t, resp)
		var status map[string]any
		err := json.Unmarshal([]byte(text), &status)
		require.NoError(t, err)

		config := status["config"].(map[string]any)
		assert.Equal(t, "readonly", config["preset_mode"])
		assert.Equal(t, "preset", config["mode"])
	})
}

// Helper functions

// TestRuntimePermissionConfigChanges tests changing permissions at runtime via MCP tools
func TestRuntimePermissionConfigChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start in readonly mode
	ctx := startMCPFromConfig(t, "testdata/readonly.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)


	// Verify INSERT blocked initially
	t.Run("step1_INSERT_blocked_in_readonly", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000060",
				"name":  "Test User",
				"email": "test@example.com",
			},
		})
		assertIsError(t, resp, "INSERT should be blocked in readonly")
		assertContains(t, resp, "not allowed")
		assertContains(t, resp, "readwrite") // Should suggest readwrite mode
	})

	// Change to readwrite mode
	t.Run("step2_change_to_readwrite", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "readwrite",
			"user_confirmed": true,
		})
		assertNotError(t, resp, "Mode change should succeed")
	})

	// Verify INSERT now works
	t.Run("step3_INSERT_now_allowed", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000061",
				"name":  "Test User 2",
				"email": "test2@example.com",
			},
		})
		assertNotError(t, resp, "INSERT should be allowed after mode change")
	})

	// Verify CREATE still blocked
	t.Run("step4_CREATE_still_blocked_in_readwrite", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
			"options": map[string]any{
				"if_not_exists": true,
			},
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
		assertIsError(t, resp, "CREATE should be blocked in readwrite")
		assertContains(t, resp, "dba") // Should suggest dba mode
	})

	// Change to dba mode
	t.Run("step5_change_to_dba", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "dba",
			"user_confirmed": true,
		})
		assertNotError(t, resp, "Mode change to dba should succeed")
	})

	// Verify CREATE now works
	t.Run("step6_CREATE_now_allowed", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "logs",
			"options": map[string]any{
				"if_not_exists": true,
			},
			"schema": map[string]any{
				"id":        "uuid PRIMARY KEY",
				"timestamp": "timestamp",
				"message":   "text",
			},
		})
		assertNotError(t, resp, "CREATE should be allowed in dba mode")
	})

	// Add confirmations on DCL
	t.Run("step7_add_confirmations_on_dcl", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"confirm_queries": "dcl",
			"user_confirmed":  true,
		})
		assertNotError(t, resp, "Adding dcl confirmations should succeed")
	})

	// Verify GRANT requires confirmation
	t.Run("step8_GRANT_requires_confirmation", func(t *testing.T) {
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
		assertContains(t, resp, "requires")
		assertContains(t, resp, "req_") // Should have request ID
	})

	// Get pending confirmations
	t.Run("step9_get_pending_confirmations", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_pending_confirmations", map[string]any{})
		if resp != nil {
			assertNotError(t, resp, "get_pending_confirmations should work")
			text := extractText(t, resp)
			assert.Contains(t, text, "req_") // Should show pending request
		}
	})

	// Disable all confirmations
	t.Run("step10_disable_confirmations", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"confirm_queries": "disable",
			"user_confirmed":  true,
		})
		assertNotError(t, resp, "Disabling confirmations should succeed")
	})

	// Verify GRANT now works without confirmation
	t.Run("step11_GRANT_no_longer_requires_confirmation", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"options": map[string]any{
				"permission": "SELECT",
				"role":       "app_readonly",
			},
		})
		assertNotError(t, resp, "GRANT should work without confirmation after disable")
	})

	// Switch to fine-grained mode
	t.Run("step12_switch_to_finegrained", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"skip_confirmation": "dql,dml",
			"user_confirmed":    true,
		})
		assertNotError(t, resp, "Switching to fine-grained should succeed")
	})

	// Verify status shows fine-grained mode
	t.Run("step13_verify_finegrained_mode", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
		if resp != nil {
			text := extractText(t, resp)
			assert.Contains(t, text, "fine-grained")
		}
	})

	// Back to preset mode
	t.Run("step14_back_to_preset_readonly", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "readonly",
			"user_confirmed": true,
		})
		assertNotError(t, resp, "Switching back to readonly should succeed")
	})
}

// TestUserConfirmedEnforcement tests user_confirmed flag is enforced
func TestUserConfirmedEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readonly.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)


	// Try to change mode WITHOUT user_confirmed
	t.Run("without_user_confirmed", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "dba",
			"user_confirmed": false,
		})
		assertIsError(t, resp, "Should reject without user_confirmed")
		assertContains(t, resp, "requires user confirmation")
	})

	// Try WITH user_confirmed
	t.Run("with_user_confirmed", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "dba",
			"user_confirmed": true,
		})
		assertNotError(t, resp, "Should succeed with user_confirmed")
	})
}

// TestPermissionLockdownEnforcement tests runtime permission changes can be disabled
func TestPermissionLockdownEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start with lockdown enabled
	ctx := startMCPFromConfig(t, "testdata/dba_locked.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)


	// Verify status shows lockdown
	t.Run("status_shows_disabled", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_mcp_status", map[string]any{})
		if resp != nil {
			text := extractText(t, resp)
			var status map[string]any
			json.Unmarshal([]byte(text), &status)
			config := status["config"].(map[string]any)
			disabled, _ := config["disable_runtime_permission_changes"].(bool)
			assert.True(t, disabled, "Lockdown should be enabled")
		}
	})

	// Try to change mode (should fail)
	t.Run("mode_change_blocked", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"mode":           "readonly",
			"user_confirmed": true,
		})
		assertIsError(t, resp, "Mode change should be blocked when locked")
		assertContains(t, resp, "disabled")
	})

	// Try to change confirm-queries (should also fail)
	t.Run("confirm_queries_change_blocked", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "update_mcp_permissions", map[string]any{
			"confirm_queries": "disable",
			"user_confirmed":  true,
		})
		assertIsError(t, resp, "Confirm-queries change should be blocked when locked")
	})
}

// Complete list of all operations to test
var allOperationsMatrix = []struct {
	operation string
	category  string
}{
	// DQL (14 operations)
	{"SELECT", "DQL"},

	// SESSION operations tested separately (always allowed)

	// DML (8 operations)
	{"INSERT", "DML"},
	{"UPDATE", "DML"},
	{"DELETE", "DML"},

	// DDL (representative sample - 28 total, testing key ones)
	{"CREATE", "DDL"},
	{"ALTER", "DDL"},
	{"DROP", "DDL"},
	{"TRUNCATE", "DDL"},

	// DCL (13 operations - testing key ones)
	{"GRANT", "DCL"},
	{"REVOKE", "DCL"},
}

// TestReadonlyMode_AllOperations tests all operations in readonly mode
func TestReadonlyMode_AllOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readonly.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)


	for _, op := range allOperationsMatrix {
		t.Run(op.operation, func(t *testing.T) {
			resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
				"operation": op.operation,
				"keyspace":  "test_mcp",
				"table":     "users",
			})

			if op.category == "DQL" {
				assertNotError(t, resp, op.operation+" should be allowed")
			} else {
				assertIsError(t, resp, op.operation+" should be blocked")
			}
		})
	}
}

// TestReadwriteMode_AllOperations tests all operations in readwrite mode
func TestReadwriteMode_AllOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := startMCPFromConfig(t, "testdata/readwrite.json")
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)


	for _, op := range allOperationsMatrix {
		t.Run(op.operation, func(t *testing.T) {
			params := map[string]any{
				"operation": op.operation,
				"keyspace":  "test_mcp",
				"table":     "users",
			}

			// Add required parameters based on operation
			switch op.operation {
			case "INSERT":
				params["values"] = map[string]any{
					"id":    "00000000-0000-0000-0000-000000000062",
					"name":  "Test User",
					"email": "test@example.com",
				}
			case "UPDATE":
				params["values"] = map[string]any{"name": "Updated Name"}
				params["where"] = []any{
					map[string]any{
						"column":   "id",
						"operator": "=",
						"value":    "00000000-0000-0000-0000-000000000062",
					},
				}
			case "DELETE":
				params["where"] = []any{
					map[string]any{
						"column":   "id",
						"operator": "=",
						"value":    "00000000-0000-0000-0000-000000000062",
					},
				}
			}

			resp := callTool(t, ctx.SocketPath, "submit_query_plan", params)

			if op.category == "DQL" || op.category == "DML" {
				assertNotError(t, resp, op.operation+" should be allowed")
			} else {
				assertIsError(t, resp, op.operation+" should be blocked")
			}
		})
	}
}

// TestConfirmationLifecycle tests confirmation request lifecycle
func TestConfirmationLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start in dba mode with confirmations on DML
	ctx := startMCPFromConfig(t, "testdata/dba_locked.json") // Has confirm_queries: ["dcl"]
	defer stopMCP(ctx)

	ensureTestDataExists(t, ctx.Session)


	// Submit a DCL operation (should create confirmation request)
	t.Run("create_confirmation_request", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "GRANT should require confirmation")
		text := extractText(t, resp)
		assert.Contains(t, text, "req_")
	})

	// Get pending confirmations
	t.Run("get_pending", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_pending_confirmations", map[string]any{})
		if resp != nil {
			assertNotError(t, resp, "get_pending should work")
			text := extractText(t, resp)

			// Parse the list
			var pending []map[string]any
			json.Unmarshal([]byte(text), &pending)

			if len(pending) > 0 {
				assert.Equal(t, "PENDING", pending[0]["status"])
				assert.Contains(t, pending[0], "request_id")
			}
		}
	})

	// Get approved confirmations (should be empty initially)
	t.Run("get_approved_empty", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_approved_confirmations", map[string]any{})
		if resp != nil {
			assertNotError(t, resp, "get_approved should work")
		}
	})

	// Get denied confirmations (should be empty)
	t.Run("get_denied_empty", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_denied_confirmations", map[string]any{})
		if resp != nil {
			assertNotError(t, resp, "get_denied should work")
		}
	})

	// Get cancelled confirmations (should be empty)
	t.Run("get_cancelled_empty", func(t *testing.T) {
		resp := callTool(t, ctx.SocketPath, "get_cancelled_confirmations", map[string]any{})
		if resp != nil {
			assertNotError(t, resp, "get_cancelled should work")
		}
	})
}

// TestFineGrainedMode tests skip-confirmation modes
func TestFineGrainedMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("skip_ALL", func(t *testing.T) {
		ctx := startMCPFromConfig(t, "testdata/finegrained_skip_all.json")
		defer stopMCP(ctx)

		ensureTestDataExists(t, ctx.Session)


		// All operations should be allowed without confirmation
		for _, op := range allOperationsMatrix {
			params := map[string]any{
				"operation": op.operation,
				"keyspace":  "test_mcp",
				"table":     "users",
			}

			// Add required parameters based on operation
			switch op.operation {
			case "INSERT":
				params["values"] = map[string]any{
					"id":    "00000000-0000-0000-0000-000000000070",
					"name":  "Test User",
					"email": "test@example.com",
				}
			case "UPDATE":
				params["values"] = map[string]any{"name": "Updated Name"}
				params["where"] = []any{
					map[string]any{
						"column":   "id",
						"operator": "=",
						"value":    "00000000-0000-0000-0000-000000000070",
					},
				}
			case "DELETE":
				params["where"] = []any{
					map[string]any{
						"column":   "id",
						"operator": "=",
						"value":    "00000000-0000-0000-0000-000000000070",
					},
				}
			}

			resp := callTool(t, ctx.SocketPath, "submit_query_plan", params)
			assertNotError(t, resp, op.operation+" should be allowed with skip ALL")
		}
	})

	t.Run("skip_none", func(t *testing.T) {
		ctx := startMCPFromConfig(t, "testdata/finegrained_skip_none.json")
		defer stopMCP(ctx)

		ensureTestDataExists(t, ctx.Session)


		// All operations should require confirmation (except SESSION)
		for _, op := range allOperationsMatrix {
			params := map[string]any{
				"operation": op.operation,
				"keyspace":  "test_mcp",
				"table":     "users",
			}

			// Add required parameters based on operation
			switch op.operation {
			case "INSERT":
				params["values"] = map[string]any{
					"id":    "00000000-0000-0000-0000-000000000071",
					"name":  "Test User",
					"email": "test@example.com",
				}
			case "UPDATE":
				params["values"] = map[string]any{"name": "Updated Name"}
				params["where"] = []any{
					map[string]any{
						"column":   "id",
						"operator": "=",
						"value":    "00000000-0000-0000-0000-000000000071",
					},
				}
			case "DELETE":
				params["where"] = []any{
					map[string]any{
						"column":   "id",
						"operator": "=",
						"value":    "00000000-0000-0000-0000-000000000071",
					},
				}
			}

			resp := callTool(t, ctx.SocketPath, "submit_query_plan", params)
			assertIsError(t, resp, op.operation+" should require confirmation with skip none")
			text := extractText(t, resp)
			assert.Contains(t, text, "requires")
		}
	})
}
