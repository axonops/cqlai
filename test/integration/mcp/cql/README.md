# CQL Test Suite - Comprehensive Cassandra 5.0.6 Coverage

**Purpose:** Complete, systematic testing of ALL CQL functionality through the CQLAI MCP server

**Status:** Work in Progress (78 INSERT tests complete)
**Target:** 1,200+ tests covering 100% of Cassandra 5.0.6 CQL specification
**Approach:** Systematic, test-driven, with full validation at every step

---

## ⚠️ CRITICAL: TestMain Pattern (Bug Fix - 2026-01-04)

**ALL test files in this directory share a single MCP server started in `TestMain()`.**

**DO NOT:**
- ❌ Call `.mcp start` in individual tests
- ❌ Call `.mcp stop` in teardown functions
- ❌ Generate new API keys per test

**DO:**
- ✅ Use shared `sharedMCPHandler`, `sharedAPIKey` from `base_helpers_test.go`
- ✅ Create unique keyspace per test for isolation
- ✅ Drop keyspace in teardown (not stop server)

**Why:** Starting server per test causes port conflicts and "401 invalid API key" errors. Only the first test would pass.

**Verified:** All 78 INSERT tests now pass sequentially with this fix.

---

## Directory Structure

```
test/integration/mcp/cql/
├── README.md (this file)
├── base_helpers_test.go          # Reusable helper functions
├── dml_insert_test.go             # INSERT tests (90+ tests)
├── dml_update_test.go             # UPDATE tests (100+ tests)
├── dml_delete_test.go             # DELETE tests (90+ tests)
├── ddl_keyspace_test.go           # Keyspace DDL (60+ tests)
├── ddl_table_test.go              # Table DDL (150+ tests)
├── ddl_types_test.go              # UDT DDL (80+ tests)
├── ddl_index_test.go              # Index DDL (110+ tests)
├── ddl_functions_test.go          # Function/Aggregate DDL (90+ tests)
├── ddl_views_test.go              # Materialized View DDL (50+ tests)
├── ddl_triggers_test.go           # Trigger DDL (20+ tests)
├── dql_select_basic_test.go       # Basic SELECT (80+ tests)
├── dql_select_advanced_test.go    # Advanced SELECT (90+ tests)
├── dql_select_functions_test.go   # Functions in SELECT (60+ tests)
├── dql_prepared_statements_test.go # Prepared statements (70+ tests)
├── dql_aggregates_test.go         # Aggregate functions (40+ tests)
├── dql_json_test.go               # JSON operations (30+ tests)
├── dcl_roles_test.go              # Role management (60+ tests)
├── dcl_permissions_test.go        # Permissions/Grants (65+ tests)
├── dcl_ddm_test.go                # Dynamic Data Masking (40+ tests)
├── spec_round_trip_test.go        # Full CRUD cycles (45+ tests)
├── spec_nesting_test.go           # Complete nesting matrix (60+ tests)
└── spec_datatypes_test.go         # All data type combinations (80+ tests)
```

---

## Test Principles

**Every test MUST:**

1. ✅ **Start in DBA mode** - No confirmation prompts (testing CQL, not permissions)
2. ✅ **Execute operation via MCP** - Test the MCP interface
3. ✅ **ASSERT exact generated CQL** - Verify query builder generates correct CQL (MANDATORY - Added 2026-01-04)
4. ✅ **Validate in Cassandra directly** - Direct query, assert exact data match
5. ✅ **Test round-trip via MCP** - INSERT→SELECT→UPDATE→DELETE all via MCP
6. ✅ **Verify state changes** - Confirm UPDATE/DELETE actually modified Cassandra
7. ✅ **Test error cases** - What should fail and why
8. ✅ **Clean up properly** - No state pollution between tests

### CQL Assertion Pattern (MANDATORY)

**After EVERY submitQueryPlanMCP() call:**
```go
result := submitQueryPlanMCP(ctx, args)
assertNoMCPError(t, result, "Operation should succeed")

// MANDATORY: Assert exact CQL
expectedCQL := "INSERT INTO ks.table (id, name) VALUES (1, 'Alice');"
assertCQLEquals(t, result, expectedCQL, "Generated CQL must match exactly")
```

**Why:** Execution success doesn't prove CQL is correct. We must verify the exact CQL generated.

**See:** Test 1 (TestDML_Insert_01_SimpleText) for complete example.

---

## Test Pattern Template

```go
func TestDML_Insert_SimpleText(t *testing.T) {
    ctx := setupCQLTest(t)
    defer teardownCQLTest(ctx)

    // 1. Create table
    err := createTable(ctx, "users", `
        users (
            id int PRIMARY KEY,
            name text
        )
    `)
    require.NoError(t, err)

    testID := 1
    testName := "Alice"

    // 2. INSERT via MCP
    insertArgs := map[string]any{
        "operation": "INSERT",
        "keyspace": ctx.Keyspace,
        "table": "users",
        "values": map[string]any{
            "id": testID,
            "name": testName,
        },
    }
    insertResult := submitQueryPlanMCP(ctx, insertArgs)
    assertNoMCPError(t, insertResult, "INSERT should succeed")

    // 3. VALIDATE in Cassandra (direct query)
    rows := validateInCassandra(ctx,
        fmt.Sprintf("SELECT id, name FROM %s.users WHERE id = ?", ctx.Keyspace),
        testID)
    require.Len(t, rows, 1, "Should retrieve 1 row from Cassandra")
    assert.Equal(t, testID, rows[0]["id"])
    assert.Equal(t, testName, rows[0]["name"])

    // 4. SELECT via MCP (round-trip)
    selectArgs := map[string]any{
        "operation": "SELECT",
        "keyspace": ctx.Keyspace,
        "table": "users",
        "where": []map[string]any{
            {"column": "id", "operator": "=", "value": testID},
        },
    }
    selectResult := submitQueryPlanMCP(ctx, selectArgs)
    assertNoMCPError(t, selectResult, "SELECT via MCP should succeed")

    // 5. UPDATE via MCP
    updateName := "Alice Updated"
    updateArgs := map[string]any{
        "operation": "UPDATE",
        "keyspace": ctx.Keyspace,
        "table": "users",
        "values": map[string]any{
            "name": updateName,
        },
        "where": []map[string]any{
            {"column": "id", "operator": "=", "value": testID},
        },
    }
    updateResult := submitQueryPlanMCP(ctx, updateArgs)
    assertNoMCPError(t, updateResult, "UPDATE should succeed")

    // 6. VALIDATE UPDATE in Cassandra
    rows = validateInCassandra(ctx,
        fmt.Sprintf("SELECT name FROM %s.users WHERE id = ?", ctx.Keyspace),
        testID)
    require.Len(t, rows, 1)
    assert.Equal(t, updateName, rows[0]["name"], "Updated name should match")

    // 7. DELETE via MCP
    deleteArgs := map[string]any{
        "operation": "DELETE",
        "keyspace": ctx.Keyspace,
        "table": "users",
        "where": []map[string]any{
            {"column": "id", "operator": "=", "value": testID},
        },
    }
    deleteResult := submitQueryPlanMCP(ctx, deleteArgs)
    assertNoMCPError(t, deleteResult, "DELETE should succeed")

    // 8. VALIDATE DELETE in Cassandra
    rows = validateInCassandra(ctx,
        fmt.Sprintf("SELECT id FROM %s.users WHERE id = ?", ctx.Keyspace),
        testID)
    assert.Len(t, rows, 0, "Row should not exist after DELETE")

    t.Log("✅ Full CRUD cycle verified: INSERT→SELECT→UPDATE→DELETE")
}
```

---

## Current Status

### Completed
- ✅ Directory structure created
- ✅ Base helpers file created
- ✅ Analysis completed (cql_test_matrix.md, cql_coverage_gaps.md)
- ✅ Comprehensive blueprint received (1,200+ test cases defined)

### In Progress
- 🔄 Integrating with existing MCP HTTP client
- 🔄 Creating first example test with full validation

### Planned
- 📋 22 test files with 1,200+ test cases
- 📋 Complete CQL 5.0.6 specification coverage
- 📋 Full validation and round-trip testing

---

## Estimated Effort

Based on analysis:
- **High Priority (Critical):** 23-34 hours
- **Medium Priority:** 6-10 hours
- **Low Priority:** 6-9 hours
- **Total:** 35-53 hours for complete suite

---

## How to Run

### Prerequisites
```bash
# 1. Start Cassandra container
podman start cassandra-test
sleep 25  # Wait for startup

# 2. Verify Cassandra is ready
podman exec cassandra-test cqlsh -u cassandra -p cassandra -e "SELECT release_version FROM system.local"
# Should show: 5.0.6

# 3. Clear test cache (ALWAYS do this before running)
go clean -testcache
```

### Run Commands

**CRITICAL: Always use `-p 1` flag (sequential execution required)**

```bash
# Run all INSERT tests (78 tests, ~4-5 minutes)
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_" -v -p 1

# Run specific test
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_01_" -v -p 1

# Run range of tests (tests 1-15)
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_(0[1-9]|1[0-5])_" -v -p 1

# Run specific category (when UPDATE/DELETE tests are added)
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Update_" -v -p 1
```

### Required Flags

- `-tags=integration` — Required (tests won't build without it)
- `-v` — Verbose output (recommended to see progress)
- `-p 1` — **REQUIRED** (sequential execution, we don't support parallel)

### Expected Results

**Successful run shows:**
```
MCP session initialized: mcp-session-XXXX  ✅ (success)
```

**Failed run shows:**
```
MCP session generated: mcp-session-cql-XXX  ❌ (fallback - indicates problem)
HTTP request failed with status 401: invalid API key  ❌
```

If you see the failed pattern, TestMain pattern is not working correctly.

---

## References

- **cql-complete-test-suite.md** - Complete test blueprint (1,200+ tests)
- **cql-implementation-guide.md** - Patterns and helpers
- **test-suite-summary.md** - Execution roadmap
- **cql_test_matrix.md** - Current test analysis
- **cql_coverage_gaps.md** - Gap analysis (479 lines)
- **c5-nesting-cql.md** - Cassandra 5 nesting rules
- **c5-nesting-mtx.md** - Nesting test matrix

---

## Notes

- Uses existing Cassandra instance (cassandra-test container)
- Uses DBA mode for MCP (no confirmation prompts)
- Tests run sequentially for now (port isolation issue remains)
- Each test creates unique keyspace for isolation
- Direct Cassandra validation is MANDATORY for every test
- Round-trip validation (MCP SELECT) required for all DML tests
