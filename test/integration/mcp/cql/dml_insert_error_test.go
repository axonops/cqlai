// +build integration

package cql

import (
	"fmt"
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Helper Functions
// ============================================================================

// createSeparateCassandraSession creates a NEW gocql session independent of MCP
// This simulates another user/connection to Cassandra
func createSeparateCassandraSession(t *testing.T) *gocql.Session {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "system"
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.Timeout = 10 * time.Second
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "cassandra",
		Password: "cassandra",
	}
	cluster.DisableInitialHostLookup = true

	session, err := cluster.CreateSession()
	assert.NoError(t, err, "Failed to create separate session")

	return session
}

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

	// Assert EXACT error message from Cassandra
	expectedError := "Query execution failed: query failed: table nonexistent_table does not exist"
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get exact table not found error from Cassandra")

	t.Log("✅ Test ERR_01: Non-existent table error verified (exact message)")
}

// TestDML_Insert_ERR_01a_TableCreatedInSeparateSession tests schema propagation across sessions
// CRITICAL TEST: Verifies that when table is created in SEPARATE session, MCP metadata updates
func TestDML_Insert_ERR_01a_TableCreatedInSeparateSession(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	tableName := "cross_session_test"

	// 1. First attempt - INSERT into non-existent table (should error)
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     tableName,
		"values": map[string]any{
			"id":   2000,
			"name": "TestUser",
		},
	}

	result1 := submitQueryPlanMCP(ctx, insertArgs)

	// Should get error
	assertMCPError(ctx.T, result1, "table", "First attempt should fail - table doesn't exist")

	t.Log("Step 1: ✅ INSERT into non-existent table failed as expected")

	// 2. Create table in SEPARATE session (NOT the MCP session)
	// This simulates another user/connection creating the table
	separateSession := createSeparateCassandraSession(t)
	defer separateSession.Close()

	createQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s (
			id int PRIMARY KEY,
			name text
		)
	`, ctx.Keyspace, tableName)

	err := separateSession.Query(createQuery).Exec()
	assert.NoError(t, err, "Failed to create table in separate session")

	t.Log("Step 2: ✅ Table created in SEPARATE session")

	// 3. Wait for schema propagation (gocql detects changes in ~1-2 seconds)
	t.Log("Step 3: Waiting for schema propagation...")
	time.Sleep(3 * time.Second)

	// 4. Retry same INSERT - should succeed now (metadata updated automatically)
	result2 := submitQueryPlanMCP(ctx, insertArgs)

	// Should succeed now!
	assertNoMCPError(ctx.T, result2, "Second attempt should succeed - table now exists")

	// Verify CQL was generated
	expectedCQL := fmt.Sprintf("INSERT INTO %s.%s (id, name) VALUES (2000, 'TestUser');", ctx.Keyspace, tableName)
	assertCQLEquals(t, result2, expectedCQL, "CQL should be generated correctly")

	// Verify data was inserted
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT id, name FROM %s.%s WHERE id = ?", ctx.Keyspace, tableName), 2000)
	assert.Len(t, rows, 1, "Row should be inserted")
	assert.Equal(t, 2000, rows[0]["id"])
	assert.Equal(t, "TestUser", rows[0]["name"])

	t.Log("Step 4: ✅ INSERT succeeded after table created in separate session")
	t.Log("✅ CRITICAL TEST PASSED: Schema propagation works across sessions!")
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

	// Assert EXACT validation error message
	expectedError := "Query validation failed: missing partition key column: device_id (required for INSERT)"
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get exact missing partition key error")

	// Verify no data was inserted
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.pk_test", ctx.Keyspace))
	assert.Len(ctx.T, rows, 0, "No data should be inserted on validation error")

	t.Log("✅ Test ERR_02: Missing partition key validation error verified (exact message)")
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

	// Assert EXACT validation error message
	// Note: Validation reports first missing clustering key found
	expectedError := "Query validation failed: missing clustering key column: month (required for INSERT)"
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get exact missing clustering key error")

	// Verify no data was inserted
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.ck_test", ctx.Keyspace))
	assert.Len(ctx.T, rows, 0, "No data should be inserted on validation error")

	t.Log("✅ Test ERR_03: Missing clustering key validation error verified (exact message)")
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
