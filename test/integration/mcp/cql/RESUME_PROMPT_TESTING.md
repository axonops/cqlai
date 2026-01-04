# Resume Prompt: CQL INSERT Testing Implementation

**Use this prompt to resume INSERT test implementation work**

---

## Context

I am continuing comprehensive CQL test suite implementation for CQLAI MCP server testing against Cassandra 5.0.6.

**Current branch:** feature/mcp_datatypes

**Current status:**
- 78 INSERT test functions implemented
- 45 test scenarios actually covered
- 96 test scenarios missing (68.1% gap)
- All 78 tests have EXACT CQL assertions (167 assertions total)
- Infrastructure complete: TestMain pattern, deterministic rendering, unit tests

**What's been done:**
- Deterministic CQL rendering (alphabetical + type-aware numeric sorting)
- TestMain pattern (shared MCP server for sequential execution)
- 167 CQL assertions across all 78 tests
- 9 bugs found and fixed by CQL assertions
- 245 unit tests (all passing)
- WHERE IN, TOKEN, tuple notation support added
- BATCH semicolon formatting fixed

---

## Files to Read (IN ORDER)

**CRITICAL - Read these first:**

1. `test/integration/mcp/cql/RESUME_TESTING_SESSION.md`
   - Complete overview of current status
   - What's done, what's missing
   - Test patterns and infrastructure
   - Common issues and solutions

2. `claude-notes/cql-complete-test-suite.md`
   - **MASTER BLUEPRINT** - All 141 INSERT scenarios documented
   - Section: "File 1: dml_insert_test.go (141 tests)"
   - Tests 1-90: Original scenarios
   - Tests 91-108: Primary key validation (NEW)
   - Tests 109-130: BATCH validation (NEW)
   - Tests 131-141: Additional errors (NEW)

3. `test/integration/mcp/cql/INSERT_GAP_ANALYSIS.md`
   - Current implementation status
   - 96 missing tests broken down by category
   - Priority levels

4. `claude-notes/CQL_TEST_ADDENDUM_PRIMARY_KEYS_AND_BATCH.md`
   - Detailed code examples for 51 new scenarios
   - Primary key validation rules
   - BATCH validation rules

**Supporting documentation:**

5. `claude-notes/cql-implementation-guide.md`
   - TestMain pattern (REQUIRED)
   - CQL assertion pattern (MANDATORY)
   - Helper function reference

6. `test/integration/mcp/cql/dml_insert_test.go`
   - 78 existing tests (use as reference)
   - Line 5995: Where to add new tests
   - See Tests 1-30 for best examples

7. `test/integration/mcp/cql/base_helpers_test.go`
   - TestMain implementation
   - CQL assertion helpers
   - Test infrastructure

---

## Current Test Infrastructure

**Before running tests:**
```bash
podman start cassandra-test
sleep 25
go clean -testcache
```

**Run all INSERT tests:**
```bash
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_" -v -p 1
# Expected: 78/78 PASS in ~274 seconds
```

**Run unit tests:**
```bash
go test ./internal/ai -v
# Expected: 245/245 PASS in <1 second
```

---

## Implementation Priority

**You specified this order:**

1. **Error Scenarios** (14 tests)
   - Start with 4 simple errors (no cluster metadata needed)
   - Then 10 that need cluster metadata

2. **Primary Key Validation** (15 tests)
   - **BLOCKER:** Requires cluster metadata manager
   - Must implement `internal/cluster/` first
   - Read requirements: `claude-notes/cluster-metadata-requirements.md`

3. **BATCH Validation** (21 tests)
   - **BLOCKER:** Requires cluster metadata for cross-partition detection

4. **Bind Markers** (10 tests)
   - Can implement independently

---

## Critical Patterns to Follow

### 1. CQL Assertion (MANDATORY)

**After EVERY submitQueryPlanMCP:**
```go
result := submitQueryPlanMCP(ctx, args)
assertNoMCPError(t, result, "Operation should succeed")

// MANDATORY: Assert exact CQL
expectedCQL := fmt.Sprintf("INSERT INTO %s.table (col1, col2) VALUES (val1, val2);", ctx.Keyspace)
assertCQLEquals(t, result, expectedCQL, "CQL must match exactly")
```

### 2. Deterministic CQL Format

- Columns: Alphabetical
- Map keys: Type-aware (numeric or alphabetical)
- UDT fields: Alphabetical
- Set elements: Type-aware (numeric or alphabetical)
- BATCH: Multi-line with semicolons

### 3. Error Test Pattern

```go
func TestDML_Insert_ERR_Description(t *testing.T) {
    ctx := setupCQLTest(t)
    defer teardownCQLTest(ctx)

    // Attempt invalid operation
    result := submitQueryPlanMCP(ctx, args)

    // Should get error
    assertMCPError(t, result, "expected error text", "Should fail")

    // Verify no data persisted
    rows := validateInCassandra(ctx, "SELECT * FROM table")
    assert.Len(t, rows, 0)
}
```

---

## Next Steps

**IMMEDIATE:**
1. Wait for cluster metadata implementation (internal/cluster/)
2. Cluster metadata must have passing integration tests
3. Cluster metadata integrated into planner

**THEN RESUME TESTING:**
1. Implement 4 simple error tests (no cluster metadata)
2. Implement 10 schema-dependent error tests (with cluster metadata)
3. Implement 15 primary key validation tests
4. Implement 21 BATCH validation tests
5. Implement 10 bind marker tests
6. Implement remaining 36 tests

**Total remaining:** 96 tests

---

## Verification Checklist

**Before claiming a test is done:**
- ✅ CQL assertion added for EVERY submitQueryPlanMCP()
- ✅ Test passes when run individually
- ✅ Test passes in full suite (sequential run)
- ✅ Data validated in Cassandra (not just MCP response)
- ✅ For error tests: Verify error message received
- ✅ For error tests: Verify no data persisted
- ✅ Commit after every 5-10 tests

---

## Common Pitfalls

❌ **Don't skip CQL assertions** - Every operation must have one
❌ **Don't use t.Skip()** - Implement the test properly
❌ **Don't assume execution success means CQL is correct** - Assert the CQL!
❌ **Don't forget to validate in Cassandra** - Direct query required
❌ **Don't mix LWT and non-LWT** - Use IF EXISTS consistently

---

**Status: Ready to resume testing after cluster metadata is complete**
