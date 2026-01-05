// +build integration

package cql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// DML DELETE TESTS
// These tests verify DELETE operations with primary key validation
// ============================================================================

// TestDML_Delete_01_FullPrimaryKey tests DELETE with full primary key
// Verifies deletion of single row when full PK specified
func TestDML_Delete_01_FullPrimaryKey(t *testing.T) {
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
	require.NoError(t, err)

	// INSERT a row
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

	// DELETE with full PK - deletes single row
	deleteArgs := map[string]any{
		"operation": "DELETE",
		"keyspace":  ctx.Keyspace,
		"table":     "events",
		"where": []map[string]any{
			{"column": "user_id", "operator": "=", "value": 1},
			{"column": "timestamp", "operator": "=", "value": 1000},
			{"column": "event_type", "operator": "=", "value": "login"},
		},
	}

	result := submitQueryPlanMCP(ctx, deleteArgs)

	// Should succeed
	assertNoMCPError(ctx.T, result, "DELETE with full PK should succeed")

	// Assert exact CQL
	expectedCQL := fmt.Sprintf("DELETE FROM %s.events WHERE user_id = 1 AND timestamp = 1000 AND event_type = 'login';", ctx.Keyspace)
	assertCQLEquals(t, result, expectedCQL, "Generated CQL should be correct")

	// Verify row deleted
	rows := validateInCassandra(ctx, fmt.Sprintf(
		"SELECT * FROM %s.events WHERE user_id = ? AND timestamp = ? AND event_type = ?",
		ctx.Keyspace,
	), 1, 1000, "login")
	assert.Len(t, rows, 0, "Row should be deleted")

	t.Log("✅ DELETE Test 01: Full PK row delete validated")
}

// TestDML_Delete_02_PartitionKeyOnly tests DELETE with only partition key
// Verifies deletion of entire partition
func TestDML_Delete_02_PartitionKeyOnly(t *testing.T) {
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
	require.NoError(t, err)

	// INSERT multiple rows in same partition
	for i, ts := range []int{1000, 2000, 3000} {
		insertArgs := map[string]any{
			"operation": "INSERT",
			"keyspace":  ctx.Keyspace,
			"table":     "events",
			"values": map[string]any{
				"user_id":    1,
				"timestamp":  ts,
				"event_type": fmt.Sprintf("event%d", i),
				"data":       "test",
			},
		}
		insertResult := submitQueryPlanMCP(ctx, insertArgs)
		assertNoMCPError(ctx.T, insertResult, "INSERT should succeed")
	}

	// Verify 3 rows inserted
	allRows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.events WHERE user_id = ?", ctx.Keyspace), 1)
	assert.Len(t, allRows, 3, "Should have 3 rows before DELETE")

	// DELETE with partition key only - deletes entire partition
	deleteArgs := map[string]any{
		"operation": "DELETE",
		"keyspace":  ctx.Keyspace,
		"table":     "events",
		"where": []map[string]any{
			{"column": "user_id", "operator": "=", "value": 1},
			// No clustering keys - deletes entire partition
		},
	}

	result := submitQueryPlanMCP(ctx, deleteArgs)

	// Should succeed
	assertNoMCPError(ctx.T, result, "DELETE partition should succeed")

	// Assert exact CQL
	expectedCQL := fmt.Sprintf("DELETE FROM %s.events WHERE user_id = 1;", ctx.Keyspace)
	assertCQLEquals(t, result, expectedCQL, "Generated CQL should be correct")

	// Verify entire partition deleted
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.events WHERE user_id = ?", ctx.Keyspace), 1)
	assert.Len(t, rows, 0, "Entire partition should be deleted")

	t.Log("✅ DELETE Test 02: Partition delete validated")
}

// TestDML_Delete_03_PartitionKey_Plus_Range tests DELETE with partition key + range
// Verifies range deletion within partition
func TestDML_Delete_03_PartitionKey_Plus_Range(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table
	err := createTable(ctx, "events", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.events (
			user_id int,
			timestamp bigint,
			data text,
			PRIMARY KEY (user_id, timestamp)
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	// INSERT multiple rows
	for _, ts := range []int{1000, 2000, 3000, 4000} {
		insertArgs := map[string]any{
			"operation": "INSERT",
			"keyspace":  ctx.Keyspace,
			"table":     "events",
			"values": map[string]any{
				"user_id":   1,
				"timestamp": ts,
				"data":      fmt.Sprintf("ts%d", ts),
			},
		}
		insertResult := submitQueryPlanMCP(ctx, insertArgs)
		assertNoMCPError(ctx.T, insertResult, "INSERT should succeed")
	}

	// DELETE with range: timestamp > 2000
	deleteArgs := map[string]any{
		"operation": "DELETE",
		"keyspace":  ctx.Keyspace,
		"table":     "events",
		"where": []map[string]any{
			{"column": "user_id", "operator": "=", "value": 1},
			{"column": "timestamp", "operator": ">", "value": 2000},
		},
	}

	result := submitQueryPlanMCP(ctx, deleteArgs)

	// Should succeed
	assertNoMCPError(ctx.T, result, "DELETE range should succeed")

	// Assert exact CQL
	expectedCQL := fmt.Sprintf("DELETE FROM %s.events WHERE user_id = 1 AND timestamp > 2000;", ctx.Keyspace)
	assertCQLEquals(t, result, expectedCQL, "Generated CQL should be correct")

	// Verify range deleted (rows with ts > 2000 deleted: 3000, 4000)
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT timestamp FROM %s.events WHERE user_id = ? ORDER BY timestamp", ctx.Keyspace), 1)
	assert.Len(t, rows, 2, "Should have 2 rows remaining (1000, 2000)")

	t.Log("✅ DELETE Test 03: Range delete validated")
}

// TestDML_Delete_04_PartitionKey_Plus_IN tests DELETE with partition key + IN clause
// Verifies deletion of multiple specific rows
func TestDML_Delete_04_PartitionKey_Plus_IN(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table
	err := createTable(ctx, "events", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.events (
			user_id int,
			timestamp bigint,
			data text,
			PRIMARY KEY (user_id, timestamp)
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	// INSERT multiple rows
	for _, ts := range []int{1000, 2000, 3000, 4000, 5000} {
		insertArgs := map[string]any{
			"operation": "INSERT",
			"keyspace":  ctx.Keyspace,
			"table":     "events",
			"values": map[string]any{
				"user_id":   1,
				"timestamp": ts,
				"data":      fmt.Sprintf("ts%d", ts),
			},
		}
		insertResult := submitQueryPlanMCP(ctx, insertArgs)
		assertNoMCPError(ctx.T, insertResult, "INSERT should succeed")
	}

	// DELETE with IN clause: timestamp IN (2000, 4000)
	deleteArgs := map[string]any{
		"operation": "DELETE",
		"keyspace":  ctx.Keyspace,
		"table":     "events",
		"where": []map[string]any{
			{"column": "user_id", "operator": "=", "value": 1},
			{"column": "timestamp", "operator": "IN", "values": []interface{}{2000, 4000}},
		},
	}

	result := submitQueryPlanMCP(ctx, deleteArgs)

	// Should succeed
	assertNoMCPError(ctx.T, result, "DELETE with IN should succeed")

	// Assert exact CQL
	expectedCQL := fmt.Sprintf("DELETE FROM %s.events WHERE user_id = 1 AND timestamp IN (2000, 4000);", ctx.Keyspace)
	assertCQLEquals(t, result, expectedCQL, "Generated CQL should be correct")

	// Verify specific rows deleted (2000, 4000 gone; 1000, 3000, 5000 remain)
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT timestamp FROM %s.events WHERE user_id = ? ORDER BY timestamp", ctx.Keyspace), 1)
	assert.Len(t, rows, 3, "Should have 3 rows remaining")

	t.Log("✅ DELETE Test 04: IN clause delete validated")
}

// TestDML_Delete_05_StaticColumn_PartialPK tests DELETE of static column with partial PK
// Verifies deletion of specific static column for partition
func TestDML_Delete_05_StaticColumn_PartialPK(t *testing.T) {
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

	// INSERT a row
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     "users",
		"values": map[string]any{
			"user_id":    1,
			"session_id": 100,
			"user_type":  "premium",
			"data":       "test",
		},
	}
	insertResult := submitQueryPlanMCP(ctx, insertArgs)
	assertNoMCPError(ctx.T, insertResult, "INSERT should succeed")

	// DELETE specific static column with partial PK
	deleteArgs := map[string]any{
		"operation": "DELETE",
		"keyspace":  ctx.Keyspace,
		"table":     "users",
		"columns":   []string{"user_type"}, // Delete specific column
		"where": []map[string]any{
			{"column": "user_id", "operator": "=", "value": 1},
			// session_id NOT included - partial PK OK for static column
		},
	}

	result := submitQueryPlanMCP(ctx, deleteArgs)

	// Should succeed
	assertNoMCPError(ctx.T, result, "DELETE static column with partial PK should succeed")

	// Assert exact CQL
	expectedCQL := fmt.Sprintf("DELETE user_type FROM %s.users WHERE user_id = 1;", ctx.Keyspace)
	assertCQLEquals(t, result, expectedCQL, "Generated CQL should be correct")

	// Verify static column deleted, row still exists
	rows := validateInCassandra(ctx, fmt.Sprintf(
		"SELECT data FROM %s.users WHERE user_id = ?",
		ctx.Keyspace,
	), 1)
	assert.Len(t, rows, 1, "Row should still exist")

	t.Log("✅ DELETE Test 05: Static column delete with partial PK validated")
}
