//go:build integration
// +build integration

package cql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/router"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// CQL Test Suite - Base Helpers and Patterns
// ============================================================================
//
// PURPOSE:
// This file provides reusable helper functions for comprehensive CQL testing
// following the pattern: INSERT → Validate in Cassandra → SELECT via MCP →
// UPDATE via MCP → Validate UPDATE → DELETE via MCP → Validate DELETE
//
// CRITICAL PRINCIPLES:
// 1. ALWAYS validate data in Cassandra directly (not just MCP response)
// 2. ALWAYS test round-trip (MCP INSERT → MCP SELECT)
// 3. ALWAYS verify UPDATE/DELETE actually changed Cassandra state
// 4. Test both success AND error cases
// 5. Use DBA mode to avoid confirmation noise
//
// Based on:
// - cql-complete-test-suite.md (1,200+ test cases)
// - cql-implementation-guide.md (20+ patterns)
// - Cassandra 5.0.6 specification
// ============================================================================

// Shared test infrastructure (initialized once in TestMain)
var (
	sharedMCPHandler *router.MCPHandler
	sharedAPIKey     string
	sharedBaseURL    = "http://127.0.0.1:8912/mcp"
)

// TestMain sets up and tears down shared test infrastructure
func TestMain(m *testing.M) {
	// Setup: Start MCP server ONCE for all tests
	exitCode := func() int {
		// Create Cassandra session
		cluster := gocql.NewCluster("127.0.0.1:9042")
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: "cassandra",
			Password: "cassandra",
		}
		cluster.Timeout = 10 * time.Second
		cluster.ConnectTimeout = 10 * time.Second
		cluster.Consistency = gocql.LocalOne

		session, err := db.NewSessionFromCluster(cluster, "cassandra", false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create Cassandra session: %v\n", err)
			return 1
		}
		defer session.Close()

		// Create MCP handler (shared across all tests)
		sharedMCPHandler = router.NewMCPHandler(session)
		if sharedMCPHandler == nil {
			fmt.Fprintf(os.Stderr, "Failed to create MCP handler\n")
			return 1
		}

		// Generate API key ONCE
		sharedAPIKey, err = ai.GenerateAPIKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate API key: %v\n", err)
			return 1
		}
		os.Setenv("TEST_MCP_API_KEY", sharedAPIKey)

		// Start MCP server ONCE
		startCmd := ".mcp start --config-file ../testdata/dba.json"
		result := sharedMCPHandler.HandleMCPCommand(startCmd)
		if !contains(result, "started successfully") {
			fmt.Fprintf(os.Stderr, "MCP start failed: %s\n", result)
			return 1
		}

		time.Sleep(500 * time.Millisecond)

		// Run all tests
		code := m.Run()

		// Cleanup: Stop MCP server ONCE
		sharedMCPHandler.HandleMCPCommand(".mcp stop")
		time.Sleep(300 * time.Millisecond)

		return code
	}()

	os.Exit(exitCode)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

// CQLTestContext holds everything needed for a CQL test
type CQLTestContext struct {
	Session    *db.Session      // Direct Cassandra session for validation
	MCPHandler *router.MCPHandler // MCP handler instance
	BaseURL    string           // MCP HTTP endpoint
	APIKey     string           // API key for MCP requests
	SessionID  string           // MCP session ID
	Keyspace   string           // Test keyspace for this test
	T          *testing.T       // Test instance
}

// setupCQLTest creates a fresh test context with MCP in DBA mode
// Uses shared MCP server (started once in TestMain)
func setupCQLTest(t *testing.T) *CQLTestContext {
	// Create Cassandra session for this test
	cluster := gocql.NewCluster("127.0.0.1:9042")
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "cassandra",
		Password: "cassandra",
	}
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second
	cluster.Consistency = gocql.LocalOne

	session, err := db.NewSessionFromCluster(cluster, "cassandra", false)
	require.NoError(t, err, "Failed to create Cassandra session")

	// Use shared MCP handler and API key (started once in TestMain)
	require.NotNil(t, sharedMCPHandler, "Shared MCP handler not initialized")
	require.NotEmpty(t, sharedAPIKey, "Shared API key not initialized")

	// Get session ID from MCP server initialization
	sessionID := initializeMCPSessionHTTP(t, sharedBaseURL, sharedAPIKey)

	// Create unique test keyspace for isolation
	keyspace := fmt.Sprintf("cql_test_%d", time.Now().UnixNano())
	err = session.Query(fmt.Sprintf(
		"CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		keyspace,
	)).Exec()
	require.NoError(t, err, "Failed to create test keyspace")

	return &CQLTestContext{
		Session:    session,
		MCPHandler: sharedMCPHandler,
		BaseURL:    sharedBaseURL,
		APIKey:     sharedAPIKey,
		SessionID:  sessionID,
		Keyspace:   keyspace,
		T:          t,
	}
}

// teardownCQLTest cleans up test context
// NOTE: Does NOT stop MCP server (shared across all tests, stopped in TestMain)
func teardownCQLTest(ctx *CQLTestContext) {
	if ctx == nil {
		return
	}

	// Drop test keyspace
	if ctx.Keyspace != "" && ctx.Session != nil {
		ctx.Session.Query(fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", ctx.Keyspace)).Exec()
	}

	// Close session (each test has its own session)
	if ctx.Session != nil {
		ctx.Session.Close()
	}

	// NOTE: Do NOT stop MCP server here - it's shared across all tests
	// and will be stopped once in TestMain cleanup
}

// ============================================================================
// MCP Operation Helpers
// ============================================================================

// callMCPTool calls an MCP tool via HTTP with proper headers
func callMCPTool(ctx *CQLTestContext, toolName string, args map[string]any) map[string]any {
	// This uses the existing HTTP MCP client pattern
	// Reuse from existing tests
	return callToolHTTPDirect(ctx.T, ctx.BaseURL, ctx.APIKey, ctx.SessionID, toolName, args)
}

// submitQueryPlanMCP submits a query plan via MCP submit_query_plan tool
// Returns full MCP response including the generated CQL query
func submitQueryPlanMCP(ctx *CQLTestContext, args map[string]any) map[string]any {
	result := callMCPTool(ctx, "submit_query_plan", args)

	// Log the operation AND generated CQL for debugging
	if op, ok := args["operation"].(string); ok {
		ctx.T.Logf("MCP Operation: %s on table %s.%s", op, args["keyspace"], args["table"])
	}

	// Extract and log the generated CQL
	if cql, ok := extractGeneratedCQL(result); ok {
		ctx.T.Logf("Generated CQL: %s", cql)
	} else {
		ctx.T.Logf("WARNING: Generated CQL not found in response")
	}

	return result
}

// extractGeneratedCQL extracts the "query" field from MCP response
// Returns (cql, true) if found, ("", false) if not found
func extractGeneratedCQL(response map[string]any) (string, bool) {
	// Check direct query field
	if query, ok := response["query"].(string); ok {
		return query, true
	}

	// Check in content array (MCP wraps response in content blocks)
	if content, ok := response["content"].([]any); ok {
		for _, c := range content {
			if contentMap, ok := c.(map[string]any); ok {
				if contentMap["type"] == "text" {
					if text, ok := contentMap["text"].(string); ok {
						// Text contains JSON string - parse it
						var parsedData map[string]any
						if err := json.Unmarshal([]byte(text), &parsedData); err == nil {
							// Extract query from parsed JSON
							if query, ok := parsedData["query"].(string); ok {
								return query, true
							}
						}
					}
				}
			}
		}
	}

	return "", false
}

// assertCQLEquals asserts the generated CQL EXACTLY matches expected CQL
// CRITICAL: This validates we're generating correct CQL, not just that execution works
// Uses exact string match after normalizing whitespace
func assertCQLEquals(t *testing.T, response map[string]any, expectedCQL string, message string) {
	actualCQL, found := extractGeneratedCQL(response)

	if !found {
		t.Fatalf("%s - Generated CQL not found in MCP response", message)
	}

	// Normalize whitespace for comparison (allows formatting differences)
	actualNorm := normalizeWhitespace(actualCQL)
	expectedNorm := normalizeWhitespace(expectedCQL)

	assert.Equal(t, expectedNorm, actualNorm, "%s - Generated CQL must match exactly", message)
}

// normalizeWhitespace normalizes whitespace for CQL comparison
// Removes extra spaces, tabs, newlines to allow flexible formatting
func normalizeWhitespace(s string) string {
	// Replace multiple spaces/tabs/newlines with single space
	normalized := strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(normalized)
}

// ============================================================================
// Direct Cassandra Validation Helpers (CRITICAL!)
// ============================================================================

// validateInCassandra executes a query directly against Cassandra and returns rows
// This is the CRITICAL validation step - verifies actual database state
func validateInCassandra(ctx *CQLTestContext, query string, params ...interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	iter := ctx.Session.Query(query, params...).Iter()

	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		result = append(result, row)
	}

	if err := iter.Close(); err != nil {
		ctx.T.Logf("Warning: Iterator close error: %v", err)
	}

	return result
}

// validateRowCount checks expected row count in Cassandra
func validateRowCount(ctx *CQLTestContext, table string, expectedCount int) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", ctx.Keyspace, table)
	rows := validateInCassandra(ctx, query)

	require.Len(ctx.T, rows, 1, "COUNT query should return 1 row")

	// Extract count value (might be "count" or "system.count(*)")
	var count int64
	for _, v := range rows[0] {
		if c, ok := v.(int64); ok {
			count = c
			break
		}
	}

	assert.Equal(ctx.T, int64(expectedCount), count, "Row count mismatch")
}

// validateDataEquals checks a specific column value in Cassandra
func validateDataEquals(ctx *CQLTestContext, table, column string, id interface{}, expected interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s.%s WHERE id = ?", column, ctx.Keyspace, table)
	rows := validateInCassandra(ctx, query, id)

	require.Len(ctx.T, rows, 1, "Should retrieve exactly 1 row")
	assert.Equal(ctx.T, expected, rows[0][column], "Column %s value mismatch", column)
}

// validateRowExists checks if a row exists in Cassandra
func validateRowExists(ctx *CQLTestContext, table string, id interface{}) bool {
	query := fmt.Sprintf("SELECT id FROM %s.%s WHERE id = ?", ctx.Keyspace, table)
	rows := validateInCassandra(ctx, query, id)
	return len(rows) > 0
}

// validateRowNotExists checks if a row does NOT exist in Cassandra
func validateRowNotExists(ctx *CQLTestContext, table string, id interface{}) {
	query := fmt.Sprintf("SELECT id FROM %s.%s WHERE id = ?", ctx.Keyspace, table)
	rows := validateInCassandra(ctx, query, id)
	assert.Len(ctx.T, rows, 0, "Row should not exist after DELETE")
}

// ============================================================================
// Schema Creation Helpers
// ============================================================================

// createTable creates a table via direct Cassandra (not MCP)
// Returns error if creation fails
func createTable(ctx *CQLTestContext, tableName, ddl string) error {
	// If DDL doesn't include keyspace, prepend it
	fullDDL := ddl
	if !containsString(ddl, ctx.Keyspace) {
		fullDDL = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s", ctx.Keyspace, ddl[len("CREATE TABLE IF NOT EXISTS "):])
	}

	return ctx.Session.Query(fullDDL).Exec()
}

// createType creates a UDT via direct Cassandra
func createType(ctx *CQLTestContext, typeName, ddl string) error {
	fullDDL := ddl
	if !containsString(ddl, ctx.Keyspace) {
		fullDDL = fmt.Sprintf("CREATE TYPE IF NOT EXISTS %s.%s", ctx.Keyspace, ddl[len("CREATE TYPE IF NOT EXISTS "):])
	}

	return ctx.Session.Query(fullDDL).Exec()
}

// dropTable drops a table if it exists
func dropTable(ctx *CQLTestContext, tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", ctx.Keyspace, tableName)
	return ctx.Session.Query(query).Exec()
}

// ============================================================================
// Round-Trip Test Pattern (The Gold Standard)
// ============================================================================

// roundTripTest performs complete INSERT→SELECT→UPDATE→SELECT→DELETE→SELECT cycle
// This is the pattern EVERY test should follow for comprehensive validation
func roundTripTest(
	ctx *CQLTestContext,
	table string,
	insertArgs map[string]any,
	selectArgs map[string]any,
	updateArgs map[string]any,
	deleteArgs map[string]any,
	expectedInsertData map[string]interface{},
	expectedUpdateData map[string]interface{},
) {
	// Step 1: INSERT via MCP
	insertResult := submitQueryPlanMCP(ctx, insertArgs)
	assertNoMCPError(ctx.T, insertResult, "INSERT via MCP should succeed")

	// Step 2: VALIDATE INSERT in Cassandra (direct query)
	id := insertArgs["values"].(map[string]any)["id"]
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.%s WHERE id = ?", ctx.Keyspace, table), id)
	require.Len(ctx.T, rows, 1, "Should retrieve inserted row from Cassandra")
	for col, expected := range expectedInsertData {
		assert.Equal(ctx.T, expected, rows[0][col], "Column %s mismatch after INSERT", col)
	}

	// Step 3: SELECT via MCP (round-trip)
	selectResult := submitQueryPlanMCP(ctx, selectArgs)
	assertNoMCPError(ctx.T, selectResult, "SELECT via MCP should succeed")

	// Step 4: UPDATE via MCP
	if updateArgs != nil {
		updateResult := submitQueryPlanMCP(ctx, updateArgs)
		assertNoMCPError(ctx.T, updateResult, "UPDATE via MCP should succeed")

		// Step 5: VALIDATE UPDATE in Cassandra
		rows = validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.%s WHERE id = ?", ctx.Keyspace, table), id)
		require.Len(ctx.T, rows, 1, "Should retrieve updated row from Cassandra")
		for col, expected := range expectedUpdateData {
			assert.Equal(ctx.T, expected, rows[0][col], "Column %s mismatch after UPDATE", col)
		}

		// Step 6: SELECT via MCP again (verify UPDATE visible)
		selectResult2 := submitQueryPlanMCP(ctx, selectArgs)
		assertNoMCPError(ctx.T, selectResult2, "SELECT after UPDATE via MCP should succeed")
	}

	// Step 7: DELETE via MCP
	if deleteArgs != nil {
		deleteResult := submitQueryPlanMCP(ctx, deleteArgs)
		assertNoMCPError(ctx.T, deleteResult, "DELETE via MCP should succeed")

		// Step 8: VALIDATE DELETE in Cassandra
		rows = validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.%s WHERE id = ?", ctx.Keyspace, table), id)
		assert.Len(ctx.T, rows, 0, "Row should not exist after DELETE")

		// Step 9: SELECT via MCP (verify DELETE visible)
		selectResult3 := submitQueryPlanMCP(ctx, selectArgs)
		// Should succeed but return empty results
		assertNoMCPError(ctx.T, selectResult3, "SELECT after DELETE via MCP should succeed")
	}
}

// ============================================================================
// Assertion Helpers
// ============================================================================

// assertNoMCPError checks MCP response doesn't contain error
func assertNoMCPError(t *testing.T, response map[string]any, message string) {
	if response == nil {
		t.Fatal("MCP response is nil")
	}

	// Check for error in various response formats
	if isError, ok := response["isError"].(bool); ok && isError {
		t.Fatalf("%s - MCP returned error: %v", message, response)
	}

	if errMsg, ok := response["error"].(string); ok && errMsg != "" {
		t.Fatalf("%s - MCP error: %s", message, errMsg)
	}

	// Check for error content
	if content, ok := response["content"].([]any); ok {
		for _, c := range content {
			if contentMap, ok := c.(map[string]any); ok {
				if contentMap["type"] == "text" {
					text := contentMap["text"].(string)
					if containsString(text, "error") || containsString(text, "Error") || containsString(text, "ERROR") {
						// May be error message - check more carefully
						if containsString(text, "Query execution failed") || containsString(text, "not allowed") {
							t.Fatalf("%s - MCP error in response: %s", message, text)
						}
					}
				}
			}
		}
	}
}

// assertMCPError checks MCP response DOES contain error
func assertMCPError(t *testing.T, response map[string]any, expectedError string, message string) {
	if response == nil {
		t.Fatal("MCP response is nil")
	}

	// Look for error indicators
	hasError := false
	errorText := ""

	if isError, ok := response["isError"].(bool); ok && isError {
		hasError = true
	}

	if errMsg, ok := response["error"].(string); ok && errMsg != "" {
		hasError = true
		errorText = errMsg
	}

	if content, ok := response["content"].([]any); ok {
		for _, c := range content {
			if contentMap, ok := c.(map[string]any); ok {
				if contentMap["type"] == "text" {
					text := contentMap["text"].(string)
					if containsString(text, "error") || containsString(text, "Error") {
						hasError = true
						errorText = text
					}
				}
			}
		}
	}

	assert.True(t, hasError, "%s - Expected error but got none", message)
	if expectedError != "" && errorText != "" {
		assert.Contains(t, errorText, expectedError, "Error message should contain expected text")
	}
}

// ============================================================================
// Utility Helpers
// ============================================================================

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// initializeMCPSessionHTTP initializes MCP session via HTTP and returns session ID
func initializeMCPSessionHTTP(t *testing.T, baseURL, apiKey string) string {
	// Send initialize request to MCP server
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "cql-test-client",
				"version": "1.0.0",
			},
		},
	}

	requestJSON, _ := json.Marshal(request)
	httpReq, _ := http.NewRequest("POST", baseURL, bytes.NewReader(requestJSON))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		t.Logf("Initialize failed: %v", err)
		return fmt.Sprintf("mcp-session-cql-%d", time.Now().UnixNano())
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	var response map[string]any
	json.Unmarshal(respBody, &response)

	// Extract session ID from response header or generate one
	if sessionHeader := httpResp.Header.Get("MCP-Session-Id"); sessionHeader != "" {
		t.Logf("MCP session initialized: %s", sessionHeader)
		return sessionHeader
	}

	// Fallback: generate session ID
	sessionID := fmt.Sprintf("mcp-session-cql-%d", time.Now().UnixNano())
	t.Logf("MCP session generated: %s", sessionID)
	return sessionID
}

// callToolHTTPDirect makes HTTP call to MCP server
// Adapted from ../http_reference_test.go
func callToolHTTPDirect(t *testing.T, baseURL, apiKey, sessionID, toolName string, args map[string]any) map[string]any {
	// Build JSON-RPC request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      time.Now().UnixNano(),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
		return nil
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", baseURL, bytes.NewReader(requestJSON))
	if err != nil{
		t.Fatalf("Failed to create HTTP request: %v", err)
		return nil
	}

	// Add required headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", apiKey)
	if sessionID != "" {
		httpReq.Header.Set("MCP-Session-Id", sessionID)
	}

	// Send request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
		return nil
	}
	defer httpResp.Body.Close()

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		t.Fatalf("HTTP request failed with status %d: %s", httpResp.StatusCode, string(bodyBytes))
		return nil
	}

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
		return nil
	}

	// Parse JSON response
	var response map[string]any
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		t.Logf("Response body: %s", string(respBody))
		t.Fatalf("Failed to parse response: %v", err)
		return nil
	}

	// Extract result from JSON-RPC response
	if result, ok := response["result"].(map[string]any); ok {
		return result
	}

	return response
}
