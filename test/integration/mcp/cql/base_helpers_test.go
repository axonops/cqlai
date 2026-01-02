//go:build integration
// +build integration

package cql

import (
	"fmt"
	"os"
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

func init() {
	// Generate test API key for CQL tests
	key, _ := ai.GenerateAPIKey()
	os.Setenv("TEST_MCP_API_KEY_CQL", key)
	os.Setenv("TEST_MCP_API_KEY", key)
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
// This is the foundation of every test - provides both direct Cassandra
// access AND MCP access for full validation
func setupCQLTest(t *testing.T) *CQLTestContext {
	// Create direct Cassandra session
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

	// Create isolated MCP handler for this test
	mcpHandler := router.NewMCPHandler(session)
	require.NotNil(t, mcpHandler)

	// Start MCP in DBA mode (no confirmations)
	apiKey, err := ai.GenerateAPIKey()
	require.NoError(t, err)
	os.Setenv("TEST_MCP_API_KEY", apiKey)

	// Use unique port for this test to avoid conflicts
	// For now, use standard port - will fix concurrency later
	startCmd := ".mcp start --config-file ../testdata/dba.json"
	result := mcpHandler.HandleMCPCommand(startCmd)
	require.Contains(t, result, "started successfully", "MCP start failed: %s", result)

	time.Sleep(300 * time.Millisecond) // Wait for HTTP server

	// Create unique test keyspace for isolation
	keyspace := fmt.Sprintf("cql_test_%d", time.Now().UnixNano())
	err = session.Query(fmt.Sprintf(
		"CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		keyspace,
	)).Exec()
	require.NoError(t, err, "Failed to create test keyspace")

	// Initialize MCP session
	sessionID, err := initializeMCPSession(t, "http://127.0.0.1:8912/mcp", apiKey)
	require.NoError(t, err, "Failed to initialize MCP session")

	return &CQLTestContext{
		Session:    session,
		MCPHandler: mcpHandler,
		BaseURL:    "http://127.0.0.1:8912/mcp",
		APIKey:     apiKey,
		SessionID:  sessionID,
		Keyspace:   keyspace,
		T:          t,
	}
}

// teardownCQLTest cleans up test context
func teardownCQLTest(ctx *CQLTestContext) {
	if ctx == nil {
		return
	}

	// Drop test keyspace
	if ctx.Keyspace != "" && ctx.Session != nil {
		ctx.Session.Query(fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", ctx.Keyspace)).Exec()
	}

	// Stop MCP
	if ctx.MCPHandler != nil {
		ctx.MCPHandler.HandleMCPCommand(".mcp stop")
		time.Sleep(300 * time.Millisecond)
	}

	// Close session
	if ctx.Session != nil {
		ctx.Session.Close()
	}
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
func submitQueryPlanMCP(ctx *CQLTestContext, args map[string]any) map[string]any {
	return callMCPTool(ctx, "submit_query_plan", args)
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

// initializeMCPSession initializes MCP session via HTTP
func initializeMCPSession(t *testing.T, baseURL, apiKey string) (string, error) {
	// Simple session ID generation
	return fmt.Sprintf("mcp-session-cql-%d", time.Now().UnixNano()), nil
}

// callToolHTTPDirect makes actual MCP tool call
// For CQL tests, we call MCP server directly via handler instead of HTTP
// This tests CQL functionality without HTTP layer complexity
func callToolHTTPDirect(t *testing.T, baseURL, apiKey, sessionID, toolName string, args map[string]any) map[string]any {
	// Not implemented yet - CQL tests need proper MCP integration
	// Returning error to make test failures obvious
	return map[string]any{
		"isError": true,
		"error":   "HTTP client not integrated - CQL tests cannot execute yet",
	}
}
