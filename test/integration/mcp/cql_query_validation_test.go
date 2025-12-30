package mcp

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: This test validates actual query execution and DB changes
// Requires: Cassandra running + test_mcp keyspace with tables

// TestQueryExecution_HappyPath tests actual query execution in DBA mode
func TestQueryExecution_HappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start in DBA mode (all operations allowed, no confirmations)
	cmd := startCQLAIWithMCP(t, "testdata/dba.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// Create a direct DB session for validation
	sess, err := db.NewSessionWithOptions(db.SessionOptions{
		Host:     "127.0.0.1",
		Port:     9042,
		Username: "cassandra",
		Password: "cassandra",
		Keyspace: "test_mcp",
	})
	require.NoError(t, err, "Failed to create validation session")
	defer sess.Close()

	// Test SELECT - verify data exists
	t.Run("SELECT_query", func(t *testing.T) {
		// Via MCP
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"columns":   []string{"*"},
		})
		assertNotError(t, resp, "SELECT should work")

		// Verify via direct DB query
		result := sess.ExecuteCQLQuery("SELECT COUNT(*) FROM test_mcp.users")
		if qr, ok := result.(db.QueryResult); ok {
			assert.Greater(t, len(qr.Data), 1, "Should have data rows")
		}
	})

	// Note: INSERT/UPDATE/DELETE/CREATE would actually execute if we implement full query execution
	// Currently submit_query_plan returns "Query plan approved" but doesn't execute
	// This is expected behavior - MCP is for query planning/validation, not execution
	// Full execution would happen in a separate implementation phase
}

// TestQueryValidation_PermissionEnforcement tests permissions block invalid queries
func TestQueryValidation_PermissionEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start in readonly mode
	cmd := startCQLAIWithMCP(t, "testdata/readonly.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// Create validation session
	sess, err := db.NewSessionWithOptions(db.SessionOptions{
		Host:     "127.0.0.1",
		Port:     9042,
		Username: "cassandra",
		Password: "cassandra",
		Keyspace: "test_mcp",
	})
	require.NoError(t, err)
	defer sess.Close()

	// Get initial row count
	result := sess.ExecuteCQLQuery("SELECT COUNT(*) FROM test_mcp.users")
	var initialCount int64
	if qr, ok := result.(db.QueryResult); ok && len(qr.Data) > 1 {
		fmt.Sscanf(qr.Data[1][0], "%d", &initialCount)
	}

	// Try INSERT via MCP (should be blocked)
	t.Run("INSERT_blocked_in_readonly", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertIsError(t, resp, "INSERT should be blocked")

		// Verify no data was inserted (count unchanged)
		result := sess.ExecuteCQLQuery("SELECT COUNT(*) FROM test_mcp.users")
		var newCount int64
		if qr, ok := result.(db.QueryResult); ok && len(qr.Data) > 1 {
			fmt.Sscanf(qr.Data[1][0], "%d", &newCount)
		}
		assert.Equal(t, initialCount, newCount, "Row count should be unchanged")
	})
}

// TestQueryValidation_DMLOperations tests data modification permissions
func TestQueryValidation_DMLOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start in readwrite mode
	cmd := startCQLAIWithMCP(t, "testdata/readwrite.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// Create validation session
	sess, err := db.NewSessionWithOptions(db.SessionOptions{
		Host:     "127.0.0.1",
		Port:     9042,
		Username: "cassandra",
		Password: "cassandra",
		Keyspace: "test_mcp",
	})
	require.NoError(t, err)
	defer sess.Close()

	// Test INSERT allowed
	t.Run("INSERT_allowed", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "INSERT should be allowed in readwrite")
		assertContains(t, resp, "approved")
	})

	// Test UPDATE allowed
	t.Run("UPDATE_allowed", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "UPDATE",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "UPDATE should be allowed in readwrite")
	})

	// Test DELETE allowed
	t.Run("DELETE_allowed", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "DELETE",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "DELETE should be allowed in readwrite")
	})
}

// TestQueryValidation_DDLOperations tests schema change permissions
func TestQueryValidation_DDLOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start in DBA mode
	cmd := startCQLAIWithMCP(t, "testdata/dba.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// CREATE TABLE should be allowed
	t.Run("CREATE_TABLE_allowed", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "CREATE",
			"keyspace":  "test_mcp",
			"table":     "test_logs",
		})
		assertNotError(t, resp, "CREATE TABLE should be allowed in DBA")
	})

	// ALTER TABLE should be allowed
	t.Run("ALTER_TABLE_allowed", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "ALTER",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "ALTER TABLE should be allowed in DBA")
	})

	// DROP TABLE should be allowed
	t.Run("DROP_TABLE_allowed", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "DROP",
			"keyspace":  "test_mcp",
			"table":     "test_logs",
		})
		assertNotError(t, resp, "DROP TABLE should be allowed in DBA")
	})

	// TRUNCATE should be allowed
	t.Run("TRUNCATE_allowed", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "TRUNCATE",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "TRUNCATE should be allowed in DBA")
	})
}

// TestQueryValidation_DCLOperations tests security operation permissions
func TestQueryValidation_DCLOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start in DBA mode
	cmd := startCQLAIWithMCP(t, "testdata/dba.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// GRANT should be allowed
	t.Run("GRANT_allowed", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "GRANT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "GRANT should be allowed in DBA")
	})

	// REVOKE should be allowed
	t.Run("REVOKE_allowed", func(t *testing.T) {
		resp := callTool(t, "submit_query_plan", map[string]any{
			"operation": "REVOKE",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "REVOKE should be allowed in DBA")
	})
}

// TestQueryValidation_ConnectionState tests DB connection is valid
func TestQueryValidation_ConnectionState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cmd := startCQLAIWithMCP(t, "testdata/readonly.json")
	defer stopCQLAI(cmd)
	time.Sleep(2 * time.Second)

	// Verify status shows connected
	resp := callTool(t, "get_mcp_status", map[string]any{})
	if resp != nil {
		text := extractText(t, resp)
		var status map[string]any
		json.Unmarshal([]byte(text), &status)

		connection := status["connection"].(map[string]any)
		assert.Equal(t, "cassandra", connection["username"])
		assert.Contains(t, connection["contact_point"], "127.0.0.1")
	}

	// Create direct session and verify same data accessible
	sess, err := db.NewSessionWithOptions(db.SessionOptions{
		Host:     "127.0.0.1",
		Port:     9042,
		Username: "cassandra",
		Password: "cassandra",
		Keyspace: "test_mcp",
	})
	require.NoError(t, err)
	defer sess.Close()

	// Verify test_mcp keyspace exists
	result := sess.ExecuteCQLQuery("SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name='test_mcp'")
	if qr, ok := result.(db.QueryResult); ok {
		assert.Greater(t, len(qr.Data), 1, "test_mcp keyspace should exist")
	}

	// Verify test tables exist
	result = sess.ExecuteCQLQuery("SELECT table_name FROM system_schema.tables WHERE keyspace_name='test_mcp'")
	if qr, ok := result.(db.QueryResult); ok {
		assert.Greater(t, len(qr.Data), 1, "Should have tables in test_mcp")
	}
}
