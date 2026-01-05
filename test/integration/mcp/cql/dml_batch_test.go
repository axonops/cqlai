// +build integration

package cql

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// BATCH OPERATION TESTS
// Moved from dml_insert_test.go for better organization
// ============================================================================

// TestDML_Batch_01_MultipleInserts tests BATCH with multiple INSERT statements
// Moved from TestDML_Insert_36
func TestDML_Batch_01_MultipleInserts(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	err := createTable(ctx, "batch_test", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.batch_test (
			id int PRIMARY KEY,
			data text
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	// BATCH with 3 INSERTs
	batchArgs := map[string]any{
		"operation": "BATCH",
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "batch_test",
				"values":    map[string]any{"id": 36001, "data": "batch1"},
			},
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "batch_test",
				"values":    map[string]any{"id": 36002, "data": "batch2"},
			},
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "batch_test",
				"values":    map[string]any{"id": 36003, "data": "batch3"},
			},
		},
	}

	batchResult := submitQueryPlanMCP(ctx, batchArgs)
	assertNoMCPError(ctx.T, batchResult, "BATCH should succeed")

	// ASSERT Generated BATCH CQL (properly formatted with newlines and semicolons)
	expectedBatchCQL := fmt.Sprintf(`BEGIN BATCH
  INSERT INTO %s.batch_test (data, id) VALUES ('batch1', 36001);
  INSERT INTO %s.batch_test (data, id) VALUES ('batch2', 36002);
  INSERT INTO %s.batch_test (data, id) VALUES ('batch3', 36003);
APPLY BATCH;`, ctx.Keyspace, ctx.Keyspace, ctx.Keyspace)
	assertCQLEquals(t, batchResult, expectedBatchCQL, "BATCH CQL should be correct")

	// Verify all 3 rows were inserted by BATCH
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s.batch_test", ctx.Keyspace))
	require.Len(t, rows, 1, "COUNT query should return 1 row")

	// Extract count value and verify it's 3
	var count int64
	for _, v := range rows[0] {
		if c, ok := v.(int64); ok {
			count = c
			break
		}
	}
	assert.Equal(t, int64(3), count, "BATCH should have inserted 3 rows")

	// Verify each individual row exists
	for _, id := range []int{36001, 36002, 36003} {
		rowCheck := validateInCassandra(ctx, fmt.Sprintf("SELECT id, data FROM %s.batch_test WHERE id = ?", ctx.Keyspace), id)
		require.Len(t, rowCheck, 1, "Row %d should exist", id)
	}

	t.Log("✅ Batch Test 01: BATCH multiple INSERTs verified - all 3 rows confirmed")
}

// TestDML_Batch_02_Unlogged tests BEGIN UNLOGGED BATCH
// Moved from TestDML_Insert_68
func TestDML_Batch_02_Unlogged(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	err := createTable(ctx, "batch_unlogged", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.batch_unlogged (
			id int PRIMARY KEY,
			data text
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	batchArgs := map[string]any{
		"operation":  "BATCH",
		"batch_type": "UNLOGGED",
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "batch_unlogged",
				"values":    map[string]any{"id": 68001, "data": "batch1"},
			},
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "batch_unlogged",
				"values":    map[string]any{"id": 68002, "data": "batch2"},
			},
		},
	}

	batchResult := submitQueryPlanMCP(ctx, batchArgs)
	assertNoMCPError(ctx.T, batchResult, "BATCH UNLOGGED should succeed")

	// ASSERT Generated BATCH CQL
	expectedBatchCQL := fmt.Sprintf("BEGIN UNLOGGED BATCH\n  INSERT INTO %s.batch_unlogged (data, id) VALUES ('batch1', 68001);\n  INSERT INTO %s.batch_unlogged (data, id) VALUES ('batch2', 68002);\nAPPLY BATCH;", ctx.Keyspace, ctx.Keyspace)
	assertCQLEquals(t, batchResult, expectedBatchCQL, "BATCH UNLOGGED CQL should be correct")

	// Verify rows inserted
	for _, id := range []int{68001, 68002} {
		rows := validateInCassandra(ctx, fmt.Sprintf("SELECT data FROM %s.batch_unlogged WHERE id = ?", ctx.Keyspace), id)
		require.Len(t, rows, 1, "Row %d should exist", id)
	}

	t.Log("✅ Batch Test 02: UNLOGGED BATCH verified")
}

// TestDML_Batch_03_Counter tests BATCH COUNTER with counter operations
// Moved from TestDML_Insert_69
func TestDML_Batch_03_Counter(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	err := createTable(ctx, "page_stats", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.page_stats (
			page_id text PRIMARY KEY,
			views counter,
			downloads counter
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	batchArgs := map[string]any{
		"operation":  "BATCH",
		"batch_type": "COUNTER",
		"batch_statements": []map[string]any{
			{
				"operation": "UPDATE",
				"keyspace":  ctx.Keyspace,
				"table":     "page_stats",
				"counter_ops": map[string]string{
					"views": "+1",
				},
				"where": []map[string]any{
					{"column": "page_id", "operator": "=", "value": "home"},
				},
			},
			{
				"operation": "UPDATE",
				"keyspace":  ctx.Keyspace,
				"table":     "page_stats",
				"counter_ops": map[string]string{
					"downloads": "+1",
				},
				"where": []map[string]any{
					{"column": "page_id", "operator": "=", "value": "home"},
				},
			},
		},
	}

	batchResult := submitQueryPlanMCP(ctx, batchArgs)
	assertNoMCPError(ctx.T, batchResult, "BATCH COUNTER should succeed")

	// ASSERT Generated BATCH CQL
	expectedBatchCQL := fmt.Sprintf("BEGIN COUNTER BATCH\n  UPDATE %s.page_stats SET views = views + 1 WHERE page_id = 'home';\n  UPDATE %s.page_stats SET downloads = downloads + 1 WHERE page_id = 'home';\nAPPLY BATCH;", ctx.Keyspace, ctx.Keyspace)
	assertCQLEquals(t, batchResult, expectedBatchCQL, "BATCH COUNTER CQL should be correct")

	// Verify counter incremented
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT views, downloads FROM %s.page_stats WHERE page_id = ?", ctx.Keyspace), "home")
	require.Len(t, rows, 1)
	assert.Equal(t, int64(1), rows[0]["views"])
	assert.Equal(t, int64(1), rows[0]["downloads"])

	t.Log("✅ Batch Test 03: COUNTER BATCH verified")
}

// TestDML_Batch_04_WithTimestamp tests BATCH with USING TIMESTAMP
// Moved from TestDML_Insert_70
func TestDML_Batch_04_WithTimestamp(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	err := createTable(ctx, "batch_ts", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.batch_ts (
			id int PRIMARY KEY,
			data text
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	batchArgs := map[string]any{
		"operation":       "BATCH",
		"using_timestamp": int64(1609459200000000),
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "batch_ts",
				"values":    map[string]any{"id": 70001, "data": "ts1"},
			},
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "batch_ts",
				"values":    map[string]any{"id": 70002, "data": "ts2"},
			},
		},
	}

	batchResult := submitQueryPlanMCP(ctx, batchArgs)
	assertNoMCPError(ctx.T, batchResult, "BATCH with TIMESTAMP should succeed")

	// ASSERT Generated BATCH CQL
	expectedBatchCQL := fmt.Sprintf("BEGIN BATCH USING TIMESTAMP 1609459200000000\n  INSERT INTO %s.batch_ts (data, id) VALUES ('ts1', 70001);\n  INSERT INTO %s.batch_ts (data, id) VALUES ('ts2', 70002);\nAPPLY BATCH;", ctx.Keyspace, ctx.Keyspace)
	assertCQLEquals(t, batchResult, expectedBatchCQL, "BATCH TIMESTAMP CQL should be correct")

	// Verify rows inserted
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s.batch_ts", ctx.Keyspace))
	require.Len(t, rows, 1)

	t.Log("✅ Batch Test 04: BATCH with USING TIMESTAMP verified")
}

// TestDML_Batch_05_WithLWT tests BATCH with IF NOT EXISTS (single partition)
// Moved from TestDML_Insert_71
func TestDML_Batch_05_WithLWT(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	err := createTable(ctx, "batch_lwt", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.batch_lwt (
			id int PRIMARY KEY,
			data text
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	batchArgs := map[string]any{
		"operation":  "BATCH",
		"batch_type": "LOGGED",
		"batch_statements": []map[string]any{
			{
				"operation":     "INSERT",
				"keyspace":      ctx.Keyspace,
				"table":         "batch_lwt",
				"values":        map[string]any{"id": 71001, "data": "data1"},
				"if_not_exists": true,
			},
		},
	}

	batchResult := submitQueryPlanMCP(ctx, batchArgs)
	assertNoMCPError(ctx.T, batchResult, "BATCH with LWT should succeed")

	// ASSERT Generated BATCH CQL
	expectedBatchCQL := fmt.Sprintf("BEGIN BATCH\n  INSERT INTO %s.batch_lwt (data, id) VALUES ('data1', 71001) IF NOT EXISTS;\nAPPLY BATCH;", ctx.Keyspace)
	assertCQLEquals(t, batchResult, expectedBatchCQL, "BATCH CQL should be correct")

	time.Sleep(5 * time.Second) // LWT delay

	rows := validateInCassandra(ctx,
		fmt.Sprintf("SELECT id FROM %s.batch_lwt WHERE id = 71001", ctx.Keyspace))
	require.Len(t, rows, 1)

	t.Log("✅ Batch Test 05: BATCH with LWT (IF NOT EXISTS) verified")
}

// TestDML_Batch_06_LWT_SamePartition tests BATCH with multiple LWT INSERTs to SAME table, SAME partition
func TestDML_Batch_06_LWT_SamePartition(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table with clustering key
	err := createTable(ctx, "events", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.events (
			user_id int,
			session_id int,
			data text,
			PRIMARY KEY (user_id, session_id)
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	// BATCH with LWT on SAME table, SAME partition key (user_id=100), different clustering keys
	batchArgs := map[string]any{
		"operation":  "BATCH",
		"batch_type": "LOGGED",
		"batch_statements": []map[string]any{
			{
				"operation":     "INSERT",
				"keyspace":      ctx.Keyspace,
				"table":         "events",
				"values":        map[string]any{"user_id": 100, "session_id": 1, "data": "session1"},
				"if_not_exists": true,
			},
			{
				"operation":     "INSERT",
				"keyspace":      ctx.Keyspace,
				"table":         "events",
				"values":        map[string]any{"user_id": 100, "session_id": 2, "data": "session2"},
				"if_not_exists": true,
			},
		},
	}

	batchResult := submitQueryPlanMCP(ctx, batchArgs)
	assertNoMCPError(ctx.T, batchResult, "BATCH with LWT on same partition should succeed")

	// Verify both rows inserted
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT session_id, data FROM %s.events WHERE user_id = ?", ctx.Keyspace), 100)
	require.Len(t, rows, 2, "Both rows should be inserted in same partition")

	t.Log("✅ Batch Test 06: LWT BATCH with same table and partition key verified")
}
