//go:build integration
// +build integration

package mcp_test

import (
        "os"
        "github.com/axonops/cqlai/internal/ai"
)

func init() {
        key, _ := ai.GenerateAPIKey()
        os.Setenv("TEST_MCP_API_KEY", key)
}

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPEndToEnd_FullProtocol tests the complete MCP protocol flow
func TestMCPEndToEnd_FullProtocol(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	// 1. Create REPL session (simulating user login)
	replSession, err := createTestREPLSession(t)
	require.NoError(t, err, "Failed to create REPL session")
	defer replSession.Close()

	t.Log("✓ REPL session created")

	// 2. Initialize MCP handler (like UI does)
	err = router.InitMCPHandler(replSession)
	require.NoError(t, err, "Failed to initialize MCP handler")

	mcpHandler := router.GetMCPHandler()
	require.NotNil(t, mcpHandler, "MCP handler should not be nil")

	t.Log("✓ MCP handler initialized")

	// 3. Start MCP server (simulating ".mcp start")
	result := mcpHandler.HandleMCPCommand(".mcp start --socket-path /tmp/cqlai-test-e2e.sock")
	require.Contains(t, result, "MCP server started successfully", "Start command should succeed")
	require.Contains(t, result, "/tmp/cqlai-test-e2e.sock", "Should show socket path")

	t.Log("✓ MCP server started")
	t.Logf("Result: %s", result)

	// Give server time to initialize
	time.Sleep(500 * time.Millisecond)

	// 4. Verify status command works
	statusResult := mcpHandler.HandleMCPCommand(".mcp status")
	require.Contains(t, statusResult, "State: RUNNING", "Status should show running")

	t.Log("✓ Status command works")

	// 5. Connect to MCP socket as a client
	conn, err := net.Dial("unix", "/tmp/cqlai-test-e2e.sock")
	require.NoError(t, err, "Should be able to connect to MCP socket")
	defer conn.Close()

	t.Log("✓ Connected to MCP socket")

	// 6. Send initialize request (MCP protocol)
	initRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	writer := bufio.NewWriter(conn)
	encoder := json.NewEncoder(writer)
	err = encoder.Encode(initRequest)
	require.NoError(t, err, "Should encode initialize request")
	writer.Flush()

	t.Log("✓ Sent initialize request")

	// 7. Read initialize response
	reader := bufio.NewReader(conn)
	decoder := json.NewDecoder(reader)

	var initResponse map[string]any
	err = decoder.Decode(&initResponse)
	require.NoError(t, err, "Should receive initialize response")

	t.Logf("✓ Received initialize response: %v", initResponse)

	// 8. Request tools list
	listToolsRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	}

	encoder = json.NewEncoder(writer)
	err = encoder.Encode(listToolsRequest)
	require.NoError(t, err, "Should encode tools/list request")
	writer.Flush()

	t.Log("✓ Sent tools/list request")

	// 9. Read tools list response
	var toolsResponse map[string]any
	err = decoder.Decode(&toolsResponse)
	require.NoError(t, err, "Should receive tools list")

	t.Logf("✓ Received tools response: %v", toolsResponse)

	// Verify tools are present
	if result, ok := toolsResponse["result"].(map[string]any); ok {
		if tools, ok := result["tools"].([]any); ok {
			assert.GreaterOrEqual(t, len(tools), 9, "Should have at least 9 tools")
			t.Logf("✓ Found %d tools", len(tools))
		}
	}

	// 10. Call a tool (list_keyspaces)
	callToolRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "list_keyspaces",
			"arguments": map[string]any{},
		},
	}

	encoder = json.NewEncoder(writer)
	err = encoder.Encode(callToolRequest)
	require.NoError(t, err, "Should encode tools/call request")
	writer.Flush()

	t.Log("✓ Sent tools/call request (list_keyspaces)")

	// 11. Read tool response
	var toolResponse map[string]any
	err = decoder.Decode(&toolResponse)
	require.NoError(t, err, "Should receive tool response")

	t.Logf("✓ Received tool response: %v", toolResponse)

	// 12. Verify metrics were updated
	metricsResult := mcpHandler.HandleMCPCommand(".mcp metrics")
	require.Contains(t, metricsResult, "Total Requests", "Metrics should show requests")

	t.Log("✓ Metrics updated")
	t.Logf("Metrics: %s", metricsResult)

	// 13. Stop MCP server
	stopResult := mcpHandler.HandleMCPCommand(".mcp stop")
	require.Contains(t, stopResult, "MCP server stopped", "Stop command should succeed")

	t.Log("✓ MCP server stopped")
	t.Logf("Stop result: %s", stopResult)

	// 14. Verify server is stopped
	statusAfterStop := mcpHandler.HandleMCPCommand(".mcp status")
	require.Contains(t, statusAfterStop, "not running", "Status should show not running")

	t.Log("✓ Server confirmed stopped")
}

// TestMCPEndToEnd_ToolExecution tests executing each tool via MCP
func TestMCPEndToEnd_ToolExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tool execution test")
	}

	// Setup
	replSession, err := createTestREPLSession(t)
	require.NoError(t, err)
	defer replSession.Close()

	err = router.InitMCPHandler(replSession)
	require.NoError(t, err)

	mcpHandler := router.GetMCPHandler()

	// Start MCP server
	result := mcpHandler.HandleMCPCommand(".mcp start --socket-path /tmp/cqlai-test-tools.sock")
	require.Contains(t, result, "started successfully")
	defer mcpHandler.HandleMCPCommand(".mcp stop")

	time.Sleep(500 * time.Millisecond)

	// Connect as client
	conn, err := net.Dial("unix", "/tmp/cqlai-test-tools.sock")
	require.NoError(t, err)
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)
	decoder := json.NewDecoder(reader)
	encoder := json.NewEncoder(writer)

	// Initialize
	encoder.Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
		},
	})
	writer.Flush()

	var initResp map[string]any
	decoder.Decode(&initResp)

	// Test each tool
	tools := []struct {
		name string
		args map[string]any
	}{
		{"list_keyspaces", map[string]any{}},
		{"list_tables", map[string]any{"keyspace": "test_mcp"}},
		{"get_schema", map[string]any{"keyspace": "test_mcp", "table": "users"}},
		{"fuzzy_search", map[string]any{"query": "user"}},
	}

	for i, tool := range tools {
		t.Run(tool.name, func(t *testing.T) {
			encoder.Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      i + 10,
				"method":  "tools/call",
				"params": map[string]any{
					"name":      tool.name,
					"arguments": tool.args,
				},
			})
			writer.Flush()

			var response map[string]any
			err := decoder.Decode(&response)
			assert.NoError(t, err, "Should receive response for %s", tool.name)

			t.Logf("Tool %s response: %v", tool.name, response)
		})
	}

	// Verify metrics show all tool calls
	metricsResult := mcpHandler.HandleMCPCommand(".mcp metrics")
	assert.Contains(t, metricsResult, "Tool Breakdown")
	t.Logf("Final metrics:\n%s", metricsResult)
}

// TestMCPEndToEnd_DangerousQuery tests dangerous query detection
func TestMCPEndToEnd_DangerousQuery(t *testing.T) {
	// Test dangerous query classification
	dangerousQueries := []string{
		"DELETE FROM test_mcp.users WHERE id = uuid()",
		"DROP TABLE test_mcp.events",
		"TRUNCATE test_mcp.orders",
		"CREATE ROLE hacker WITH PASSWORD = 'bad'",
	}

	for _, query := range dangerousQueries {
		t.Run(query, func(t *testing.T) {
			classification := ai.ClassifyQuery(query)
			assert.True(t, classification.IsDangerous, "Query should be dangerous: %s", query)
			assert.Contains(t, []string{ai.SeverityCritical, ai.SeverityHigh}, classification.Severity)
			t.Logf("Query: %s → Severity: %s, Operation: %s", query, classification.Severity, classification.Operation)
		})
	}
}
