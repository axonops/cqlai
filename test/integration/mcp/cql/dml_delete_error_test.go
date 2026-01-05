// +build integration

package cql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// DELETE ERROR SCENARIO TESTS
// ============================================================================

// TestDML_Delete_ERR_01_MissingPartitionKey tests DELETE without partition key
func TestDML_Delete_ERR_01_MissingPartitionKey(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table with clustering keys
	err := createTable(ctx, "events", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.events (
			user_id int,
			timestamp bigint,
			event_type text,
			data text,
			PRIMARY KEY (user_id, timestamp, event_type)
		)
	`, ctx.Keyspace))
	assert.NoError(ctx.T, err)

	// Insert a row first
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     "events",
		"values": map[string]any{
			"user_id":    1,
			"timestamp":  1000,
			"event_type": "login",
			"data":       "test",
		},
	}
	insertResult := submitQueryPlanMCP(ctx, insertArgs)
	assertNoMCPError(ctx.T, insertResult, "INSERT should succeed")

	// Attempt DELETE without partition key
	deleteArgs := map[string]any{
		"operation": "DELETE",
		"keyspace":  ctx.Keyspace,
		"table":     "events",
		"where": []map[string]any{
			{"column": "timestamp", "operator": "=", "value": 1000},
			{"column": "event_type", "operator": "=", "value": "login"},
			// Missing user_id partition key!
		},
	}

	result := submitQueryPlanMCP(ctx, deleteArgs)

	// Assert EXACT validation error message
	expectedError := "Query validation failed: WHERE clause must include at least one partition key column for DELETE"
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get exact missing partition key error")

	// Verify row still exists (not deleted)
	rows := validateInCassandra(ctx, fmt.Sprintf(
		"SELECT data FROM %s.events WHERE user_id = ? AND timestamp = ? AND event_type = ?",
		ctx.Keyspace,
	), 1, 1000, "login")
	assert.Len(ctx.T, rows, 1, "Row should still exist - DELETE was rejected")

	t.Log("✅ DELETE_ERR_01: Missing partition key validation error verified")
}

// TestDML_Delete_ERR_02_NoWHEREClause tests DELETE without WHERE clause
func TestDML_Delete_ERR_02_NoWHEREClause(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table
	err := createTable(ctx, "users", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.users (
			id int PRIMARY KEY,
			name text
		)
	`, ctx.Keyspace))
	assert.NoError(ctx.T, err)

	// Insert a row
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     "users",
		"values": map[string]any{
			"id":   100,
			"name": "Test User",
		},
	}
	insertResult := submitQueryPlanMCP(ctx, insertArgs)
	assertNoMCPError(ctx.T, insertResult, "INSERT should succeed")

	// Attempt DELETE without WHERE clause
	deleteArgs := map[string]any{
		"operation": "DELETE",
		"keyspace":  ctx.Keyspace,
		"table":     "users",
		"where":     []map[string]any{},
		// Empty WHERE clause
	}

	result := submitQueryPlanMCP(ctx, deleteArgs)

	// Assert EXACT validation error message
	expectedError := "Query validation failed: WHERE clause is required for DELETE (DELETE without WHERE is not allowed)"
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get exact WHERE clause required error")

	// Verify row still exists
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT id, name FROM %s.users WHERE id = ?", ctx.Keyspace), 100)
	assert.Len(ctx.T, rows, 1, "Row should still exist - DELETE was rejected")

	t.Log("✅ DELETE_ERR_02: No WHERE clause validation error verified (exact message)")
}
