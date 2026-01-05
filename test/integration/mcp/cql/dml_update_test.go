// +build integration

package cql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// DML UPDATE TESTS
// These tests verify UPDATE operations with primary key validation
// ============================================================================

// TestDML_Update_01_FullPrimaryKey tests UPDATE with full primary key in WHERE
// Verifies validation allows UPDATE when full PK (partition + all clustering keys) in WHERE
func TestDML_Update_01_FullPrimaryKey(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table with partition key + clustering keys
	err := createTable(ctx, "events", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.events (
			user_id int,
			timestamp bigint,
			event_type text,
			data text,
			PRIMARY KEY (user_id, timestamp, event_type)
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	// INSERT a row first
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

	// UPDATE with FULL primary key in WHERE clause
	updateArgs := map[string]any{
		"operation": "UPDATE",
		"keyspace":  ctx.Keyspace,
		"table":     "events",
		"values": map[string]any{
			"data": "updated",
		},
		"where": []map[string]any{
			{"column": "user_id", "operator": "=", "value": 1},
			{"column": "timestamp", "operator": "=", "value": 1000},
			{"column": "event_type", "operator": "=", "value": "login"},
		},
	}

	result := submitQueryPlanMCP(ctx, updateArgs)

	// Should succeed - validation allows full PK in WHERE
	assertNoMCPError(ctx.T, result, "UPDATE with full PK should succeed")

	// Assert exact generated CQL
	expectedCQL := fmt.Sprintf("UPDATE %s.events SET data = 'updated' WHERE user_id = 1 AND timestamp = 1000 AND event_type = 'login';", ctx.Keyspace)
	assertCQLEquals(t, result, expectedCQL, "Generated CQL should be correct")

	// Verify data updated in Cassandra
	rows := validateInCassandra(ctx, fmt.Sprintf(
		"SELECT data FROM %s.events WHERE user_id = ? AND timestamp = ? AND event_type = ?",
		ctx.Keyspace,
	), 1, 1000, "login")
	require.Len(t, rows, 1, "Row should exist")
	assert.Equal(t, "updated", rows[0]["data"], "Data should be updated")

	t.Log("✅ UPDATE Test 01: Full primary key UPDATE validated and verified")
}

// TestDML_Update_02_PartialPK_StaticColumn tests UPDATE of static column with partial PK
// Verifies validation allows UPDATE of STATIC column with only partition key (no clustering keys)
func TestDML_Update_02_PartialPK_StaticColumn(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table with static column
	err := createTable(ctx, "users", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.users (
			user_id int,
			session_id int,
			user_type text STATIC,
			data text,
			PRIMARY KEY (user_id, session_id)
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	// INSERT a row first
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     "users",
		"values": map[string]any{
			"user_id":    1,
			"session_id": 100,
			"user_type":  "basic",
			"data":       "session data",
		},
	}
	insertResult := submitQueryPlanMCP(ctx, insertArgs)
	assertNoMCPError(ctx.T, insertResult, "INSERT should succeed")

	// UPDATE static column with PARTIAL PK (only partition key, no clustering key)
	updateArgs := map[string]any{
		"operation": "UPDATE",
		"keyspace":  ctx.Keyspace,
		"table":     "users",
		"values": map[string]any{
			"user_type": "premium", // Static column
		},
		"where": []map[string]any{
			{"column": "user_id", "operator": "=", "value": 1},
			// session_id NOT included - partial PK is OK for static columns
		},
	}

	result := submitQueryPlanMCP(ctx, updateArgs)

	// Should succeed - validation allows partial PK for static column UPDATE
	assertNoMCPError(ctx.T, result, "UPDATE static column with partial PK should succeed")

	// Assert exact generated CQL
	expectedCQL := fmt.Sprintf("UPDATE %s.users SET user_type = 'premium' WHERE user_id = 1;", ctx.Keyspace)
	assertCQLEquals(t, result, expectedCQL, "Generated CQL should be correct")

	// Verify static column updated for entire partition
	// Note: Cannot use clustering key in WHERE when selecting only static columns
	rows := validateInCassandra(ctx, fmt.Sprintf(
		"SELECT user_type FROM %s.users WHERE user_id = ?",
		ctx.Keyspace,
	), 1)
	require.Len(t, rows, 1, "Row should exist")
	assert.Equal(t, "premium", rows[0]["user_type"], "Static column should be updated")

	t.Log("✅ UPDATE Test 02: Static column UPDATE with partial PK validated and verified")
}

