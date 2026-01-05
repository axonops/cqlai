// +build integration

package cql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// UPDATE ERROR SCENARIO TESTS
// ============================================================================

// TestDML_Update_ERR_01_PartialPK_RegularColumn tests UPDATE with partial PK on regular column
// This should be caught by validation - need full PK for regular columns
func TestDML_Update_ERR_01_PartialPK_RegularColumn(t *testing.T) {
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
			"data":       "original",
		},
	}
	insertResult := submitQueryPlanMCP(ctx, insertArgs)
	assertNoMCPError(ctx.T, insertResult, "INSERT should succeed")

	// Attempt UPDATE with partial PK (only partition key, missing clustering keys)
	updateArgs := map[string]any{
		"operation": "UPDATE",
		"keyspace":  ctx.Keyspace,
		"table":     "events",
		"values": map[string]any{
			"data": "updated", // Regular column
		},
		"where": []map[string]any{
			{"column": "user_id", "operator": "=", "value": 1},
			// Missing timestamp and event_type!
		},
	}

	result := submitQueryPlanMCP(ctx, updateArgs)

	// Assert EXACT validation error message (includes keyspace.table)
	expectedError := fmt.Sprintf("Query validation failed: missing clustering key column(s) in WHERE clause: timestamp, event_type. Include all clustering keys in order: timestamp, event_type (required for UPDATE of regular columns on %s.events)", ctx.Keyspace)
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get exact missing clustering keys error with keyspace.table")

	// Verify original data unchanged
	rows := validateInCassandra(ctx, fmt.Sprintf(
		"SELECT data FROM %s.events WHERE user_id = ? AND timestamp = ? AND event_type = ?",
		ctx.Keyspace,
	), 1, 1000, "login")
	assert.Len(ctx.T, rows, 1)
	assert.Equal(ctx.T, "original", rows[0]["data"], "Data should be unchanged on validation error")

	t.Log("✅ UPDATE_ERR_01: Partial PK on regular column validation error verified")
}

// TestDML_Update_ERR_02_MissingPartitionKey tests UPDATE without partition key
func TestDML_Update_ERR_02_MissingPartitionKey(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table
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

	// Attempt UPDATE missing partition key
	updateArgs := map[string]any{
		"operation": "UPDATE",
		"keyspace":  ctx.Keyspace,
		"table":     "events",
		"values": map[string]any{
			"data": "updated",
		},
		"where": []map[string]any{
			{"column": "timestamp", "operator": "=", "value": 1000},
			{"column": "event_type", "operator": "=", "value": "login"},
			// Missing user_id partition key!
		},
	}

	result := submitQueryPlanMCP(ctx, updateArgs)

	// Assert EXACT validation error message (includes keyspace.table)
	expectedError := fmt.Sprintf("Query validation failed: missing partition key in WHERE clause: user_id (required for UPDATE on %s.events)", ctx.Keyspace)
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get exact missing partition key error with keyspace.table")

	t.Log("✅ UPDATE_ERR_02: Missing partition key validation error verified (exact message)")
}

// TestDML_Update_ERR_03_NoWHEREClause tests UPDATE without WHERE clause
func TestDML_Update_ERR_03_NoWHEREClause(t *testing.T) {
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

	// Attempt UPDATE without WHERE clause
	updateArgs := map[string]any{
		"operation": "UPDATE",
		"keyspace":  ctx.Keyspace,
		"table":     "users",
		"values": map[string]any{
			"name": "updated",
		},
		"where": []map[string]any{},
		// Empty WHERE clause
	}

	result := submitQueryPlanMCP(ctx, updateArgs)

	// Assert EXACT validation error message (includes keyspace.table)
	expectedError := fmt.Sprintf("Query validation failed: WHERE clause is required for UPDATE on %s.users", ctx.Keyspace)
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get exact WHERE clause required error with keyspace.table")

	t.Log("✅ UPDATE_ERR_03: No WHERE clause validation error verified (exact message)")
}
