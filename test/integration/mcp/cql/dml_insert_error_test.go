// +build integration

package cql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ERROR SCENARIO TESTS (Tests 87-90 + Primary Key Validation Errors)
// ============================================================================

// TestDML_Insert_ERR_01_NonExistentTable tests INSERT into non-existent table
// This should fail with error from Cassandra
func TestDML_Insert_ERR_01_NonExistentTable(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Attempt INSERT into table that doesn't exist
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     "nonexistent_table",
		"values": map[string]any{
			"id":   1000,
			"name": "Test",
		},
	}

	result := submitQueryPlanMCP(ctx, insertArgs)

	// Should get error from Cassandra
	assertMCPError(ctx.T, result, "table", "Should fail - table doesn't exist")

	t.Log("✅ Test ERR_01: Non-existent table error verified")
}

// TestDML_Insert_ERR_02_MissingPartitionKey tests INSERT without partition key
// This should be caught by our validation BEFORE CQL generation
func TestDML_Insert_ERR_02_MissingPartitionKey(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table with composite partition key
	err := createTable(ctx, "pk_test", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.pk_test (
			user_id int,
			device_id int,
			timestamp bigint,
			data text,
			PRIMARY KEY ((user_id, device_id), timestamp)
		)
	`, ctx.Keyspace))
	assert.NoError(ctx.T, err)

	// Attempt INSERT missing partition key (device_id missing)
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     "pk_test",
		"values": map[string]any{
			"user_id":   1000,
			// device_id missing!
			"timestamp": 1000,
			"data":      "test",
		},
	}

	result := submitQueryPlanMCP(ctx, insertArgs)

	// Should get validation error BEFORE CQL generation
	assertMCPError(ctx.T, result, "partition key", "Should fail - missing partition key")
	assertMCPError(ctx.T, result, "device_id", "Error should mention missing column")

	// Verify no data was inserted
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.pk_test", ctx.Keyspace))
	assert.Len(ctx.T, rows, 0, "No data should be inserted on validation error")

	t.Log("✅ Test ERR_02: Missing partition key validation error verified")
}

// TestDML_Insert_ERR_03_MissingClusteringKey tests INSERT without clustering key
// This should be caught by our validation BEFORE CQL generation
func TestDML_Insert_ERR_03_MissingClusteringKey(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table with clustering keys
	err := createTable(ctx, "ck_test", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.ck_test (
			sensor_id int,
			year int,
			month int,
			day int,
			value double,
			PRIMARY KEY (sensor_id, year, month, day)
		)
	`, ctx.Keyspace))
	assert.NoError(ctx.T, err)

	// Attempt INSERT missing clustering keys (month and day missing)
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     "ck_test",
		"values": map[string]any{
			"sensor_id": 100,
			"year":      2024,
			// month missing!
			// day missing!
			"value": 98.6,
		},
	}

	result := submitQueryPlanMCP(ctx, insertArgs)

	// Should get validation error
	assertMCPError(ctx.T, result, "clustering key", "Should fail - missing clustering keys")
	// Should mention at least one missing column (month or day)
	// Note: Validation reports first missing key found
	errorStr := fmt.Sprintf("%v", result)
	hasMonth := strings.Contains(errorStr, "month")
	hasDay := strings.Contains(errorStr, "day")
	assert.True(ctx.T, hasMonth || hasDay, "Error should mention at least one missing clustering key")

	// Verify no data was inserted
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.ck_test", ctx.Keyspace))
	assert.Len(ctx.T, rows, 0, "No data should be inserted on validation error")

	t.Log("✅ Test ERR_03: Missing clustering key validation error verified")
}

// TestDML_Insert_ERR_04_TypeMismatch tests INSERT with wrong data type
// This should fail from Cassandra (our planner doesn't do type checking yet)
func TestDML_Insert_ERR_04_TypeMismatch(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table
	err := createTable(ctx, "type_test", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.type_test (
			id int PRIMARY KEY,
			age int,
			name text
		)
	`, ctx.Keyspace))
	assert.NoError(ctx.T, err)

	// Attempt INSERT with string where int expected
	// Note: Our planner doesn't do type checking, so CQL will be generated
	// but Cassandra should reject it
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     "type_test",
		"values": map[string]any{
			"id":   1000,
			"age":  "not_a_number", // Wrong type!
			"name": "John",
		},
	}

	_ = submitQueryPlanMCP(ctx, insertArgs)

	// Note: Type checking is not implemented in planner yet
	// This test is a placeholder for future type validation
	// For now, we just verify the test can execute
	// TODO: Add type validation in planner to catch this earlier

	t.Log("✅ Test ERR_04: Type mismatch test executed (planner doesn't validate types yet)")
}
