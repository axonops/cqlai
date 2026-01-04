# Resume Testing Session - Complete Guide

**Created:** 2026-01-04
**Branch:** feature/mcp_datatypes
**Purpose:** Resume INSERT test implementation after cluster metadata feature is complete
**Current Status:** 45/141 INSERT scenarios implemented (31.9%)

---

## Quick Resume

**To resume INSERT testing work:**

1. Read this file (RESUME_TESTING_SESSION.md)
2. Read master blueprint: `claude-notes/cql-complete-test-suite.md`
3. Check current status: `INSERT_GAP_ANALYSIS.md`
4. Review schema architecture: `internal/schema/ARCHITECTURE.md`
5. Start implementing missing tests (96 scenarios remaining)

**Current test command:**
```bash
podman start cassandra-test
sleep 25
go clean -testcache
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_" -v -p 1
```

---

## What's Been Accomplished

### ✅ Infrastructure Complete

**1. All 78 existing tests have CQL assertions (167 assertions total)**
- Every `submitQueryPlanMCP()` call validates EXACT generated CQL
- INSERT, SELECT, UPDATE, DELETE operations all verified
- File: `test/integration/mcp/cql/dml_insert_test.go`

**2. Deterministic CQL Rendering**
- Column names sorted alphabetically
- Map keys sorted (type-aware: numeric or lexicographic)
- UDT fields sorted alphabetically
- Set elements sorted (type-aware: numeric or lexicographic)
- BATCH statements have proper semicolons
- File: `internal/ai/planner.go`

**3. TestMain Pattern (Shared MCP Server)**
- MCP server starts ONCE for all tests
- No more "401 invalid API key" errors
- Tests run sequentially without issues
- File: `test/integration/mcp/cql/base_helpers_test.go`

**4. WHERE Clause Improvements**
- IN operator parsing and rendering
- TOKEN() wrapper support
- Tuple notation support
- File: `internal/ai/mcp.go`, `internal/ai/planner.go`

**5. Unit Tests Complete (245 tests)**
- Deterministic rendering tests (7 tests)
- WHERE clause tests (IN, TOKEN, tuple)
- Fixed wrong BATCH semicolon test
- File: `internal/ai/planner_deterministic_test.go`

### ✅ Documentation Complete

**Master Blueprint:**
- `claude-notes/cql-complete-test-suite.md` - ALL 141 INSERT scenarios documented

**Gap Analysis:**
- `test/integration/mcp/cql/INSERT_GAP_ANALYSIS.md` - Current status (45/141)

**New Scenarios:**
- `claude-notes/CQL_TEST_ADDENDUM_PRIMARY_KEYS_AND_BATCH.md` - 51 new scenarios detailed

**Implementation Guides:**
- `claude-notes/cql-implementation-guide.md` - Patterns and helpers
- `claude-notes/CRITICAL_TESTMAIN_PATTERN.md` - TestMain requirement
- `internal/schema/ARCHITECTURE.md` - Schema metadata design

### ✅ 9 Bugs Found and Fixed

1. Non-deterministic ordering (columns, maps, UDTs, sets)
2. BATCH missing semicolons
3. WHERE IN operator broken
4. TOKEN() wrapper not applied
5. Tuple notation broken
6. BATCH validation inadequate
7. Test 47 loop missing assertions
8. Decimal quoting mismatch
9. Bigint JSON precision loss

---

## What's Missing (96 Scenarios)

### **Tier 1: CRITICAL (52 tests)**

**1. Bind Markers (10 tests) - 0% coverage**
- File location: Tests 66-75 in blueprint
- Requires: Prepared statement support in MCP
- Complexity: Medium
- Implementation time: ~5 hours

**2. Primary Key Validation (15 tests) - 0% coverage**
- File location: Tests 91-108 in blueprint
- **Requires: Cluster metadata manager (internal/cluster/ - to be implemented)**
- Complexity: High (schema-aware)
- Implementation time: ~8-10 hours
- **MUST implement cluster metadata first!**

**3. Error Scenarios (14 tests) - ~21% coverage**
- File location: Tests 87-90, plus integrated in PK/BATCH tests
- 4 simple errors (no schema needed)
- 10 schema-dependent errors (missing PK, invalid WHERE)
- Complexity: Medium to High
- Implementation time: ~6 hours

**4. INSERT JSON (8 tests) - 12% coverage**
- File location: Tests 57-65 in blueprint
- Requires: JSON parsing, NULL handling, escaping
- Complexity: Medium
- Implementation time: ~4 hours

### **Tier 2: HIGH (21 tests)**

**5. BATCH Validation (21 tests) - 24% coverage (5/21)**
- File location: Tests 109-130 in blueprint
- **Requires: Cluster metadata for cross-partition detection**
- Counter mixing (2 tests)
- Cross-partition warnings (4 tests)
- Mixed DML (6 tests)
- USING clauses (3 tests)
- Size/atomicity (5 tests)
- LWT multi-row (1 test)
- Complexity: High (schema-aware)
- Implementation time: ~10 hours

### **Tier 3: MEDIUM (23 tests)**

**6-11. Various** (tuples, USING, collections, UDTs, primitives, edge cases)
- Mostly independent tests
- Varying complexity
- Implementation time: ~8 hours

---

## Implementation Priority (Your Directive)

**You specified this order:**

1. ✅ **Error Scenarios** (14 tests)
   - Start with 4 simple errors (no schema)
   - Then 10 schema-dependent errors

2. **Primary Key Validation** (15 tests) - **REQUIRES SCHEMA PACKAGE**
   - **BLOCKER:** Must implement `internal/schema/` first
   - Write 5 integration tests for schema package
   - Then implement PK validation tests

3. **BATCH Validation** (21 tests) - **REQUIRES SCHEMA PACKAGE**
   - Counter/non-counter mixing
   - Cross-partition detection needs schema

4. **Bind Markers** (10 tests)
   - Prepared statements

---

## Critical Dependencies

### Schema Metadata Package (MUST IMPLEMENT FIRST)

**Why needed:**
- Primary key validation (15 tests)
- BATCH cross-partition detection (4 tests)
- Static column semantics (5 tests)
- Error validation (10 tests)
- **Total blocked: 34 tests**

**Implementation sequence:**
1. Create `internal/schema/` package
2. Write 5 integration tests (schema retrieval, refresh, CREATE/DROP)
3. All tests must PASS
4. Then integrate into planner
5. Then implement blocked tests

**Estimated time:** 2-3 days for schema package + integration

---

## Test Infrastructure Ready

**Helper functions available:**
- `setupCQLTest(t)` - Creates test context with MCP server
- `teardownCQLTest(ctx)` - Cleans up test keyspace
- `createTable(ctx, name, ddl)` - Creates table in Cassandra
- `submitQueryPlanMCP(ctx, args)` - Executes via MCP, returns CQL
- `assertNoMCPError(t, result, msg)` - Asserts operation succeeded
- `assertCQLEquals(t, result, expectedCQL, msg)` - **Asserts EXACT CQL**
- `validateInCassandra(ctx, query, params...)` - Direct Cassandra validation

**Test pattern (established in 78 tests):**
```go
func TestDML_Insert_XX_Description(t *testing.T) {
    ctx := setupCQLTest(t)
    defer teardownCQLTest(ctx)

    // 1. CREATE TABLE
    createTable(ctx, "table_name", ddl)

    // 2. Operation via MCP
    result := submitQueryPlanMCP(ctx, args)
    assertNoMCPError(t, result, "Operation should succeed")

    // 2a. ASSERT exact CQL
    expectedCQL := fmt.Sprintf("...", ctx.Keyspace)
    assertCQLEquals(t, result, expectedCQL, "CQL must match exactly")

    // 3. Validate in Cassandra
    rows := validateInCassandra(ctx, query, params...)
    assert.Equal(t, expected, rows[0]["column"])

    // 4. DELETE cleanup
    ...
}
```

**For error tests:**
```go
func TestDML_Insert_ERR_Description(t *testing.T) {
    ctx := setupCQLTest(t)
    defer teardownCQLTest(ctx)

    // Attempt invalid operation
    result := submitQueryPlanMCP(ctx, args)

    // Should get error back
    assertMCPError(t, result, "expected error substring", "Should fail with error")

    // Verify no data in Cassandra
    rows := validateInCassandra(ctx, "SELECT * FROM table")
    assert.Len(t, rows, 0, "No data should be inserted on error")
}
```

---

## Files to Read on Resume

**PRIORITY ORDER:**

1. **This file** - `RESUME_TESTING_SESSION.md` (overview)

2. **Master blueprint** - `claude-notes/cql-complete-test-suite.md`
   - Section: "File 1: dml_insert_test.go (141 tests)"
   - Tests 1-90: Original scenarios
   - Tests 91-108: Primary key validation
   - Tests 109-130: BATCH validation
   - Tests 131-141: Additional errors

3. **Gap analysis** - `test/integration/mcp/cql/INSERT_GAP_ANALYSIS.md`
   - Shows 96 missing tests by category
   - Current status: 45/141 (31.9%)

4. **New scenario details** - `claude-notes/CQL_TEST_ADDENDUM_PRIMARY_KEYS_AND_BATCH.md`
   - 51 new test scenarios with code examples
   - Primary key validation rules
   - BATCH validation rules

5. **Cluster metadata requirements** - `claude-notes/cluster-metadata-requirements.md`
   - Requirements for cluster metadata manager
   - Will be created by user
   - **Read this before implementing cluster metadata!**

6. **Implementation guide** - `claude-notes/cql-implementation-guide.md`
   - TestMain pattern
   - CQL assertion pattern
   - Helper functions

7. **Test file** - `test/integration/mcp/cql/dml_insert_test.go`
   - 78 existing tests (reference examples)
   - Line 5995: Where to add new tests

---

## How to Run Tests

**Before running:**
```bash
podman start cassandra-test
sleep 25
podman exec cassandra-test cqlsh -u cassandra -p cassandra -e "SELECT release_version FROM system.local"
# Should show: 5.0.6
```

**Run tests:**
```bash
go clean -testcache  # ALWAYS clear cache first
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_" -v -p 1
# Expected: 78/78 PASS in ~274 seconds
```

**Run specific test:**
```bash
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_01_" -v -p 1
```

**Run unit tests:**
```bash
go test ./internal/ai -v
# Expected: 245 tests PASS in <1 second
```

---

## Commit History (This Session)

1. `a75cf18` - Tests 1-30 + deterministic rendering + TestMain
2. `9efdfda` - Tests 31-40 + BATCH semicolons
3. `74c5c13` - Tests 41-44
4. `3de52f5` - Tests 45-78 + WHERE IN/TOKEN/tuple fixes
5. `0423aaa` - Session summary
6. `5f234d0` - Unit tests for all implementation changes
7. `06f93a7` - Type-aware numeric sorting
8. `33c2eab` - Gap analysis (47 new scenarios)
9. `af79ecf` - Gap analysis correction
10. `2b739f3` - Master blueprint consolidated
11. `b529e77` - Error vs upsert clarification
12. `ef5b9ea` - Schema architecture design

**Tags:**
- `checkpoint-cql-assertions-30-tests`
- `checkpoint-cql-assertions-40-tests`
- `milestone-all-78-tests-cql-assertions`
- `unit-tests-complete`

---

## Key Patterns Established

### CQL Assertion Pattern (MANDATORY)

```go
result := submitQueryPlanMCP(ctx, args)
assertNoMCPError(t, result, "Operation should succeed")

// MANDATORY: Assert exact CQL
expectedCQL := fmt.Sprintf("INSERT INTO %s.table (col1, col2) VALUES (val1, val2);", ctx.Keyspace)
assertCQLEquals(t, result, expectedCQL, "CQL must match exactly")
```

**Applies to:** Every INSERT, SELECT, UPDATE, DELETE, BATCH

### Deterministic CQL Format

**Columns:** Alphabetical
```sql
INSERT INTO t (age, email, id, name) VALUES (...)  -- a, e, i, n
```

**Map keys:** Type-aware
```sql
-- Text: Alphabetical
{'key1': 'v1', 'key2': 'v2'}

-- Numeric: Numeric order
{1: 'first', 2: 'second', 10: 'tenth'}  -- 1 < 2 < 10 (not "1" < "10" < "2")
```

**UDT fields:** Alphabetical
```sql
{city: 'NYC', street: '123 Main', zip: '10001'}  -- c, s, z
```

**Set elements:** Type-aware
```sql
-- Text: Alphabetical
{'apple', 'banana', 'cherry'}

-- Numeric: Numeric order
{7, 13, 21}  -- 7 < 13 < 21 (not "13" < "21" < "7")
```

### TestMain Pattern (REQUIRED)

All tests share ONE MCP server started in `TestMain()`:
- `sharedMCPHandler`, `sharedAPIKey`, `sharedBaseURL`
- Each test creates unique keyspace for isolation
- Tests run sequentially with `-p 1` flag

---

## Implementation Roadmap

### Phase 1: Cluster Metadata Manager (PREREQUISITE)

**BLOCKER:** 34 tests require cluster metadata (primary keys, BATCH validation, static columns, error validation)

**Requirements will be provided in:** `claude-notes/cluster-metadata-requirements.md`

**Implementation location:** `internal/cluster/` package

**What it needs to provide:**
- Table schema metadata (partition keys, clustering keys, static columns)
- Column type information
- Schema refresh on CREATE/ALTER/DROP
- Caching with invalidation

**Must have integration tests validating:**
- Initial schema retrieval works
- Schema changes are detected (ALTER TABLE, CREATE, DROP)
- All tests pass before planner integration

**Estimated time:** TBD (based on requirements)

---

### Phase 2: Planner Integration

**After cluster metadata is implemented:**
- Integrate metadata into planner
- Add validation before CQL rendering
- Return errors early (before sending to Cassandra)

**Estimated time:** 1-2 days

---

### Phase 3: Implement Missing Tests

**Priority order (as you specified):**

**Round 1: Error Scenarios (14 tests) - ~2 days**
- 4 simple errors (no schema): non-existent table, type mismatch
- 10 schema-dependent errors (missing PK, invalid WHERE)

**Round 2: Primary Key Validation (15 tests) - ~3 days**
- INSERT with full/partial PK
- UPDATE with full/partial PK, static columns
- DELETE with full/partial/range PK

**Round 3: BATCH Validation (21 tests) - ~4 days**
- Counter/non-counter mixing
- Cross-partition detection (needs schema)
- Mixed DML operations
- Size, atomicity, LWT

**Round 4: Bind Markers (10 tests) - ~2 days**
- Anonymous/named markers
- Prepared statements

**Round 5: Remaining (36 tests) - ~5 days**
- INSERT JSON variants (8)
- Tuples (4)
- USING variants (6)
- Collections, UDTs, primitives (18)

**Total estimated time:** 16-18 days for all 96 tests

---

## Critical Files Reference

| File | Purpose | Lines |
|------|---------|-------|
| `dml_insert_test.go` | All INSERT tests | 6,100+ |
| `base_helpers_test.go` | Test infrastructure, TestMain | 600+ |
| `planner.go` | CQL rendering, deterministic sorting | 2,900+ |
| `mcp.go` | MCP server, WHERE clause parsing | 1,700+ |
| `planner_deterministic_test.go` | Deterministic rendering unit tests | 180+ |
| `planner_where_test.go` | WHERE clause unit tests | 100+ |

---

## Common Issues and Solutions

### Issue: Tests fail sequentially
**Cause:** TestMain not working
**Check:** Is MCP server started once in base_helpers_test.go?
**Solution:** Verify sharedMCPHandler, sharedAPIKey are used

### Issue: CQL assertion fails
**Cause:** Non-deterministic ordering or wrong expected CQL
**Check:** Are columns/maps/sets sorted correctly?
**Solution:** Update expected CQL to match deterministic format

### Issue: "WHERE IN null"
**Cause:** Not parsing "values" field
**Check:** mcp.go line 1426-1429
**Solution:** Already fixed

### Issue: Set elements in wrong order
**Cause:** Numeric sets sorting lexicographically
**Check:** Using sortStringsNumericAware with isNumericType?
**Solution:** Already fixed

---

## Next Steps on Resume

1. **Review schema architecture** - `internal/schema/ARCHITECTURE.md`
2. **Implement schema package** - `internal/schema/*.go`
3. **Write 5 integration tests** - Verify schema retrieval works
4. **Integrate into planner** - Add validation before rendering
5. **Start Phase 3** - Implement error scenarios (4 simple ones first)

---

**Status: Infrastructure complete, 78 tests with CQL assertions, ready for schema-aware implementation**
