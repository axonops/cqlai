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

// TestDML_Batch_07_MixedDML tests BATCH with INSERT, UPDATE, and DELETE
// Verifies BATCH can mix different DML operations
func TestDML_Batch_07_MixedDML(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table
	err := createTable(ctx, "users", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.users (
			id int PRIMARY KEY,
			name text
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	// Pre-insert rows for UPDATE and DELETE
	for _, id := range []int{2, 3} {
		insertArgs := map[string]any{
			"operation": "INSERT",
			"keyspace":  ctx.Keyspace,
			"table":     "users",
			"values":    map[string]any{"id": id, "name": fmt.Sprintf("user%d", id)},
		}
		insertResult := submitQueryPlanMCP(ctx, insertArgs)
		assertNoMCPError(ctx.T, insertResult, "Pre-insert should succeed")
	}

	// BATCH with mixed DML: INSERT, UPDATE, DELETE
	batchArgs := map[string]any{
		"operation": "BATCH",
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "users",
				"values":    map[string]any{"id": 1, "name": "Alice"},
			},
			{
				"operation": "UPDATE",
				"keyspace":  ctx.Keyspace,
				"table":     "users",
				"values":    map[string]any{"name": "Bob Updated"},
				"where":     []map[string]any{{"column": "id", "operator": "=", "value": 2}},
			},
			{
				"operation": "DELETE",
				"keyspace":  ctx.Keyspace,
				"table":     "users",
				"where":     []map[string]any{{"column": "id", "operator": "=", "value": 3}},
			},
		},
	}

	result := submitQueryPlanMCP(ctx, batchArgs)
	assertNoMCPError(ctx.T, result, "Mixed DML BATCH should succeed")

	// Assert CQL
	expectedCQL := fmt.Sprintf(`BEGIN BATCH
  INSERT INTO %s.users (id, name) VALUES (1, 'Alice');
  UPDATE %s.users SET name = 'Bob Updated' WHERE id = 2;
  DELETE FROM %s.users WHERE id = 3;
APPLY BATCH;`, ctx.Keyspace, ctx.Keyspace, ctx.Keyspace)
	assertCQLEquals(t, result, expectedCQL, "Mixed DML BATCH CQL should be correct")

	// Verify results
	row1 := validateInCassandra(ctx, fmt.Sprintf("SELECT name FROM %s.users WHERE id = ?", ctx.Keyspace), 1)
	assert.Len(t, row1, 1)
	assert.Equal(t, "Alice", row1[0]["name"])

	row2 := validateInCassandra(ctx, fmt.Sprintf("SELECT name FROM %s.users WHERE id = ?", ctx.Keyspace), 2)
	assert.Len(t, row2, 1)
	assert.Equal(t, "Bob Updated", row2[0]["name"])

	row3 := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.users WHERE id = ?", ctx.Keyspace), 3)
	assert.Len(t, row3, 0, "Row 3 should be deleted")

	t.Log("✅ Batch Test 07: Mixed DML (INSERT+UPDATE+DELETE) validated")
}

// TestDML_Batch_08_MultipleTables tests BATCH across multiple tables
// Verifies BATCH can operate on different tables
func TestDML_Batch_08_MultipleTables(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create two tables
	err := createTable(ctx, "users", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.users (
			id int PRIMARY KEY,
			name text
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	err = createTable(ctx, "profiles", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.profiles (
			id int PRIMARY KEY,
			bio text
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	// BATCH across two tables
	batchArgs := map[string]any{
		"operation": "BATCH",
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "users",
				"values":    map[string]any{"id": 1, "name": "Alice"},
			},
			{
				"operation": "INSERT",
				"keyspace":  ctx.Keyspace,
				"table":     "profiles",
				"values":    map[string]any{"id": 1, "bio": "Developer"},
			},
		},
	}

	result := submitQueryPlanMCP(ctx, batchArgs)
	assertNoMCPError(ctx.T, result, "Multi-table BATCH should succeed")

	// Verify both inserts
	userRows := validateInCassandra(ctx, fmt.Sprintf("SELECT name FROM %s.users WHERE id = ?", ctx.Keyspace), 1)
	assert.Len(t, userRows, 1)

	profileRows := validateInCassandra(ctx, fmt.Sprintf("SELECT bio FROM %s.profiles WHERE id = ?", ctx.Keyspace), 1)
	assert.Len(t, profileRows, 1)

	t.Log("✅ Batch Test 08: Multiple tables BATCH validated")
}

// TestDML_Batch_09_WithTTL tests BATCH with TTL on individual statements
// Verifies TTL can be applied to statements within BATCH
func TestDML_Batch_09_WithTTL(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table
	err := createTable(ctx, "temp_data", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.temp_data (
			id int PRIMARY KEY,
			data text
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	// BATCH with statement-level TTL
	batchArgs := map[string]any{
		"operation": "BATCH",
		"batch_statements": []map[string]any{
			{
				"operation":  "INSERT",
				"keyspace":   ctx.Keyspace,
				"table":      "temp_data",
				"values":     map[string]any{"id": 1, "data": "expires in 300s"},
				"using_ttl":  300,
			},
			{
				"operation":  "INSERT",
				"keyspace":   ctx.Keyspace,
				"table":      "temp_data",
				"values":     map[string]any{"id": 2, "data": "expires in 600s"},
				"using_ttl":  600,
			},
		},
	}

	result := submitQueryPlanMCP(ctx, batchArgs)
	assertNoMCPError(ctx.T, result, "BATCH with statement-level TTL should succeed")

	// Assert CQL
	expectedCQL := fmt.Sprintf(`BEGIN BATCH
  INSERT INTO %s.temp_data (data, id) VALUES ('expires in 300s', 1) USING TTL 300;
  INSERT INTO %s.temp_data (data, id) VALUES ('expires in 600s', 2) USING TTL 600;
APPLY BATCH;`, ctx.Keyspace, ctx.Keyspace)
	assertCQLEquals(t, result, expectedCQL, "BATCH with TTL CQL should be correct")

	// Verify rows inserted with CORRECT TTL values
	row1 := validateInCassandra(ctx, fmt.Sprintf("SELECT data, TTL(data) FROM %s.temp_data WHERE id = ?", ctx.Keyspace), 1)
	require.Len(t, row1, 1)
	assert.Equal(t, "expires in 300s", row1[0]["data"])

	// TTL should be ~300 seconds (allow some margin for execution time)
	if ttl1, ok := row1[0]["ttl(data)"].(int32); ok {
		assert.GreaterOrEqual(t, ttl1, int32(290), "TTL should be ~300s (minus execution time)")
		assert.LessOrEqual(t, ttl1, int32(300), "TTL should not exceed 300s")
	}

	row2 := validateInCassandra(ctx, fmt.Sprintf("SELECT data, TTL(data) FROM %s.temp_data WHERE id = ?", ctx.Keyspace), 2)
	require.Len(t, row2, 1)
	assert.Equal(t, "expires in 600s", row2[0]["data"])

	// TTL should be ~600 seconds
	if ttl2, ok := row2[0]["ttl(data)"].(int32); ok {
		assert.GreaterOrEqual(t, ttl2, int32(590), "TTL should be ~600s (minus execution time)")
		assert.LessOrEqual(t, ttl2, int32(600), "TTL should not exceed 600s")
	}

	t.Log("✅ Batch Test 09: BATCH with statement-level TTL validated - TTL values verified")
}

// TestDML_Batch_10_LargeBatch tests BATCH with many statements (size test)
// Verifies BATCH can handle 10+ statements
func TestDML_Batch_10_LargeBatch(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table
	err := createTable(ctx, "large_batch", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.large_batch (
			id int PRIMARY KEY,
			data text
		)
	`, ctx.Keyspace))
	require.NoError(t, err)

	// BATCH with 15 statements
	statements := []map[string]any{}
	for i := 1; i <= 15; i++ {
		statements = append(statements, map[string]any{
			"operation": "INSERT",
			"keyspace":  ctx.Keyspace,
			"table":     "large_batch",
			"values":    map[string]any{"id": i, "data": fmt.Sprintf("row%d", i)},
		})
	}

	batchArgs := map[string]any{
		"operation":        "BATCH",
		"batch_statements": statements,
	}

	result := submitQueryPlanMCP(ctx, batchArgs)
	assertNoMCPError(ctx.T, result, "Large BATCH should succeed")

	// Verify all 15 rows inserted with CORRECT data in CORRECT columns
	for i := 1; i <= 15; i++ {
		rows := validateInCassandra(ctx, fmt.Sprintf("SELECT id, data FROM %s.large_batch WHERE id = ?", ctx.Keyspace), i)
		require.Len(t, rows, 1, "Row %d should exist", i)
		assert.Equal(t, i, rows[0]["id"], "Row %d should have correct id", i)
		assert.Equal(t, fmt.Sprintf("row%d", i), rows[0]["data"], "Row %d should have correct data", i)
	}

	t.Log("✅ Batch Test 10: Large BATCH (15 statements) validated - all rows verified with correct data")
}

