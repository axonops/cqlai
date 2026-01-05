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
