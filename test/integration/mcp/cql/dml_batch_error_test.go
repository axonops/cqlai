// +build integration

package cql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// BATCH ERROR SCENARIO TESTS
// ============================================================================

// TestDML_Batch_ERR_01_MixedCounterAndRegular tests BATCH mixing counter and non-counter operations
// This should be caught by validation
func TestDML_Batch_ERR_01_MixedCounterAndRegular(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create counter table
	err := createTable(ctx, "page_views", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.page_views (
			page_id text PRIMARY KEY,
			views counter
		)
	`, ctx.Keyspace))
	assert.NoError(ctx.T, err)

	// Create regular table
	err = createTable(ctx, "users", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.users (
			id int PRIMARY KEY,
			name text
		)
	`, ctx.Keyspace))
	assert.NoError(ctx.T, err)

	// Attempt BATCH mixing counter and regular operations
	batchArgs := map[string]any{
		"operation":  "BATCH",
		"batch_type": "LOGGED",
		"batch_statements": []map[string]any{
			{
				"operation": "UPDATE",
				"keyspace":  ctx.Keyspace,
				"table":     "page_views",
				"counter_ops": map[string]string{
					"views": "+1",
				},
				"where": []map[string]any{
					{"column": "page_id", "operator": "=", "value": "home"},
				},
			},
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "users",
				"values": map[string]any{
					"id":   200,
					"name": "Alice",
				},
			},
		},
	}

	result := submitQueryPlanMCP(ctx, batchArgs)

	// Assert EXACT validation error message (improved to be instructive with JSON API)
	expectedError := "Query validation failed: cannot mix counter and non-counter operations in BATCH. Use separate batch requests: batch_type='COUNTER' for counter operations, batch_type='LOGGED' or 'UNLOGGED' for regular operations"
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get instructive error with JSON API guidance")

	// Verify no data was inserted/updated
	// Check counter table
	counterRows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.page_views", ctx.Keyspace))
	assert.Len(ctx.T, counterRows, 0, "Counter table should be empty - BATCH was rejected")

	// Check regular table
	userRows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.users", ctx.Keyspace))
	assert.Len(ctx.T, userRows, 0, "Users table should be empty - BATCH was rejected")

	t.Log("✅ BATCH_ERR_01: Counter/non-counter mixing validation error verified")
}

// TestDML_Batch_ERR_02_LWT_CrossPartition tests BATCH with LWT across different partitions
// This should fail - LWT BATCH cannot span multiple partitions
func TestDML_Batch_ERR_02_LWT_CrossPartition(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table
	err := createTable(ctx, "batch_lwt_err", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.batch_lwt_err (
			id int PRIMARY KEY,
			data text
		)
	`, ctx.Keyspace))
	assert.NoError(ctx.T, err)

	// Attempt BATCH with LWT on DIFFERENT partition keys (71001 and 71002)
	batchArgs := map[string]any{
		"operation":  "BATCH",
		"batch_type": "LOGGED",
		"batch_statements": []map[string]any{
			{
				"operation":     "INSERT",
				"keyspace":      ctx.Keyspace,
				"table":         "batch_lwt_err",
				"values":        map[string]any{"id": 71001, "data": "data1"},
				"if_not_exists": true,
			},
			{
				"operation":     "INSERT",
				"keyspace":      ctx.Keyspace,
				"table":         "batch_lwt_err",
				"values":        map[string]any{"id": 71002, "data": "data2"},
				"if_not_exists": true,
			},
		},
	}

	result := submitQueryPlanMCP(ctx, batchArgs)

	// Should get error - LWT BATCH cannot span partitions
	// Note: This error currently comes from CQL rendering, not our validator
	// TODO: Add validation to catch this earlier
	assertMCPError(ctx.T, result, "partition", "Should fail - LWT BATCH cannot span multiple partitions")

	// Verify no data was inserted
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.batch_lwt_err", ctx.Keyspace))
	assert.Len(ctx.T, rows, 0, "No data should be inserted - BATCH was rejected")

	t.Log("✅ BATCH_ERR_02: LWT cross-partition error verified (currently from CQL renderer)")
}

// TestDML_Batch_ERR_03_EmptyBatch tests BATCH with no statements
// Should be caught by validation
func TestDML_Batch_ERR_03_EmptyBatch(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Attempt BATCH with no statements
	batchArgs := map[string]any{
		"operation":        "BATCH",
		"batch_type":       "LOGGED",
		"batch_statements": []map[string]any{}, // Empty!
	}

	result := submitQueryPlanMCP(ctx, batchArgs)

	// Should get validation error
	expectedError := "Query validation failed: BATCH must contain at least one statement"
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Empty BATCH should be rejected by validation")

	t.Log("✅ BATCH_ERR_03: Empty BATCH validation error verified")
}
