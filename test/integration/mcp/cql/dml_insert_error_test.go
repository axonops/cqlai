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
	// Note: Cassandra error just includes table name (not keyspace.table prefix)
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

	// Assert EXACT error message for non-existent table
	// Note: Cassandra error just includes table name (not keyspace.table)
	expectedError1 := fmt.Sprintf("Query execution failed: query failed: table %s does not exist", tableName)
	assertMCPErrorMessageExact(ctx.T, result1, expectedError1, "First attempt should fail with exact table not found error")

	t.Log("Step 1: ✅ INSERT into non-existent table failed as expected (exact error)")

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

	// Assert EXACT validation error message (improved to list ALL partition keys + keyspace.table)
	expectedError := fmt.Sprintf("Query validation failed: missing partition key column(s): device_id. Include all partition keys: user_id, device_id (required for INSERT into %s.pk_test)", ctx.Keyspace)
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get helpful error listing all partition keys and keyspace.table")

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

	// Assert EXACT validation error message (improved to list ALL clustering keys in order + keyspace.table)
	expectedError := fmt.Sprintf("Query validation failed: missing clustering key column(s): month, day. Include all clustering keys in order: year, month, day (required for INSERT into %s.ck_test)", ctx.Keyspace)
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should get helpful error listing all clustering keys in order and keyspace.table")

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

// TestDML_Insert_ERR_05_FrozenUDTFieldUpdate tests that updating frozen UDT fields is correctly rejected
// This is EXPECTED to error - frozen UDTs cannot have individual fields updated
// Moved from Test 53 to error file for better organization
func TestDML_Insert_ERR_05_FrozenUDTFieldUpdate(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	err := ctx.Session.Query(fmt.Sprintf(`
		CREATE TYPE IF NOT EXISTS %s.person (
			name text,
			age int
		)
	`, ctx.Keyspace)).Exec()
	assert.NoError(ctx.T, err)

	err = createTable(ctx, "udt_update", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.udt_update (
			id int PRIMARY KEY,
			info frozen<person>
		)
	`, ctx.Keyspace))
	assert.NoError(ctx.T, err)

	testID := 53000

	// INSERT with frozen UDT
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     "udt_update",
		"values": map[string]any{
			"id": testID,
			"info": map[string]any{
				"name": "Alice",
				"age":  30,
			},
		},
		"value_types": map[string]any{
			"info": "frozen<person>",
		},
	}

	insertResult := submitQueryPlanMCP(ctx, insertArgs)
	assertNoMCPError(ctx.T, insertResult, "INSERT should succeed")

	// ASSERT Generated INSERT CQL
	expectedInsertCQL := fmt.Sprintf("INSERT INTO %s.udt_update (id, info) VALUES (53000, {age: 30, name: 'Alice'});", ctx.Keyspace)
	assertCQLEquals(t, insertResult, expectedInsertCQL, "INSERT CQL should be correct")

	// Attempt to update frozen UDT field: info.age = 31
	// This should fail - frozen UDTs cannot have individual fields updated
	updateArgs := map[string]any{
		"operation": "UPDATE",
		"keyspace":  ctx.Keyspace,
		"table":     "udt_update",
		"collection_ops": map[string]any{
			"info": map[string]any{
				"operation":  "set_field",
				"key":        "age", // Field name
				"value":      31,
				"value_type": "int",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}

	updateResult := submitQueryPlanMCP(ctx, updateArgs)

	// Should get error - frozen UDT fields cannot be updated
	assertMCPError(ctx.T, updateResult, "frozen", "Should error when trying to update frozen UDT field")

	// Verify original data unchanged in Cassandra
	rows := validateInCassandra(ctx,
		fmt.Sprintf("SELECT id FROM %s.udt_update WHERE id = ?", ctx.Keyspace),
		testID)
	assert.Len(ctx.T, rows, 1, "Original row should still exist unchanged")

	// Cleanup
	deleteArgs := map[string]any{
		"operation": "DELETE",
		"keyspace":  ctx.Keyspace,
		"table":     "udt_update",
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}

	deleteResult := submitQueryPlanMCP(ctx, deleteArgs)
	assertNoMCPError(ctx.T, deleteResult, "DELETE should succeed")

	// ASSERT Generated DELETE CQL
	expectedDeleteCQL := fmt.Sprintf("DELETE FROM %s.udt_update WHERE id = 53000;", ctx.Keyspace)
	assertCQLEquals(t, deleteResult, expectedDeleteCQL, "DELETE CQL should be correct")

	validateRowNotExists(ctx, "udt_update", testID)

	t.Log("✅ ERR_05: Frozen UDT field update correctly rejected with error")
}

// TestDML_Insert_ERR_06_ThreeColumnPartitionKey tests INSERT with 3-column partition key missing 2 columns
// This verifies improved error message lists ALL required partition keys
func TestDML_Insert_ERR_06_ThreeColumnPartitionKey(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Create table with 3-column composite partition key
	err := createTable(ctx, "multi_tenant", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.multi_tenant (
			org_id int,
			tenant_id int,
			region text,
			timestamp bigint,
			data text,
			PRIMARY KEY ((org_id, tenant_id, region), timestamp)
		)
	`, ctx.Keyspace))
	assert.NoError(ctx.T, err)

	// Attempt INSERT with only org_id (missing tenant_id and region)
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  ctx.Keyspace,
		"table":     "multi_tenant",
		"values": map[string]any{
			"org_id":    100,
			// tenant_id missing!
			// region missing!
			"timestamp": 1000,
			"data":      "test",
		},
	}

	result := submitQueryPlanMCP(ctx, insertArgs)

	// Assert EXACT error message - should list ALL missing partition keys + keyspace.table
	expectedError := fmt.Sprintf("Query validation failed: missing partition key column(s): tenant_id, region. Include all partition keys: org_id, tenant_id, region (required for INSERT into %s.multi_tenant)", ctx.Keyspace)
	assertMCPErrorMessageExact(ctx.T, result, expectedError, "Should list all missing partition keys and keyspace.table")

	// Verify no data was inserted
	rows := validateInCassandra(ctx, fmt.Sprintf("SELECT * FROM %s.multi_tenant", ctx.Keyspace))
	assert.Len(ctx.T, rows, 0, "No data should be inserted on validation error")

	t.Log("✅ ERR_06: 3-column partition key validation with helpful error listing all required keys")
}

// TestDML_Insert_ERR_07_NonFrozenNestedCollection tests list<list<int>> without frozen (should error)
// Cassandra requires nested collections to be frozen - should get error from Cassandra
func TestDML_Insert_ERR_07_NonFrozenNestedCollection(t *testing.T) {
	ctx := setupCQLTest(t)
	defer teardownCQLTest(ctx)

	// Attempt to create table with non-frozen nested collection
	// Cassandra should reject: "Non-frozen collections are not allowed inside collections"
	err := createTable(ctx, "bad_nesting", fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.bad_nesting (
			id int PRIMARY KEY,
			data list<list<int>>
		)
	`, ctx.Keyspace))

	// Should get error from Cassandra
	assert.Error(ctx.T, err, "CREATE TABLE with list<list<int>> should fail - must be frozen")
	assert.Contains(ctx.T, err.Error(), "frozen", "Error should mention frozen requirement")

	t.Log("✅ ERR_07: Non-frozen nested collection rejected by Cassandra as expected")
}


