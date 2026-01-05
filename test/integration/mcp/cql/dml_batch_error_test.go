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
