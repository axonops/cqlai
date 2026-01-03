# Resume Prompt for Next Session - CQL Test Suite

**Use this EXACT prompt to resume work on the comprehensive CQL test suite.**

---

## RESUME PROMPT (Copy This Exactly)

```
Continue implementing comprehensive CQL test suite for CQLAI MCP server.

I am continuing a systematic test implementation that has been very successful.
We've completed 90 DML INSERT tests with full validation and found/fixed 5 bugs.

CURRENT STATE:
- Branch: feature/mcp_datatypes
- Status: DML INSERT suite COMPLETE (90/90 tests, 100% passing)
- Progress: 90/1,200 tests (7.5%)
- Tag: session-complete-90-insert-tests
- Token budget: 546K remaining

CRITICAL: READ ALL THESE FILES IN ORDER BEFORE STARTING:

Phase 1 - Understanding the Blueprint (MUST READ):
1. claude-notes/cql-complete-test-suite.md - Master blueprint defining ALL 1,200+ tests
2. claude-notes/cql-implementation-guide.md - 20+ reusable test patterns
3. claude-notes/test-suite-summary.md - 8-week execution roadmap
4. claude-notes/c5-nesting-cql.md - Cassandra 5 nesting rules (CRITICAL for frozen keyword)
5. claude-notes/c5-nesting-mtx.md - Complete nesting test matrix

Phase 2 - Understanding Current Progress (MUST READ):
6. test/integration/mcp/cql/PROGRESS_TRACKER.md - Real-time progress (90/1,200 done)
7. test/integration/mcp/cql/FINAL_SESSION_SUMMARY.md - Last session accomplishments
8. test/integration/mcp/cql/BUGS_FOUND_AND_FIXED.md - 5 bugs found and how they were fixed

Phase 3 - Critical Testing Rules (MUST READ):
9. test/integration/mcp/cql/LWT_TESTING_GUIDELINES.md - LWT vs non-LWT separation (CRITICAL!)
10. test/integration/mcp/cql/README.md - Test suite structure and principles

Phase 4 - Analysis and Context:
11. claude-notes/cql_test_matrix.md - Analysis of existing 73 tests
12. claude-notes/cql_coverage_gaps.md - 479 lines of gap analysis

Phase 5 - Implementation Reference:
13. test/integration/mcp/cql/dml_insert_test.go - 90 completed tests to use as template

VERIFY INFRASTRUCTURE:
Run ONE test to confirm everything works:
  cd /Users/johnny/Development/cqlai
  git checkout feature/mcp_datatypes
  go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_01_" -v

Should complete in ~4 seconds and show:
  ✅ Test 1: Simple text - Full CRUD cycle verified
  --- PASS: TestDML_Insert_01_SimpleText

NEXT TASK: Create DML UPDATE test suite
File: test/integration/mcp/cql/dml_update_test.go
Target: Start with tests 1-25 (first checkpoint)
Blueprint: claude-notes/cql-complete-test-suite.md section "DML UPDATE Tests"
```

---

## WHY These Documents Matter

### claude-notes/cql-complete-test-suite.md (1,200+ tests defined)
**CRITICAL** - This is the master plan. Every test case is pre-defined with:
- What to test
- Expected behavior
- Edge cases to cover
- Error conditions

**For UPDATE suite:** Section defines all 100 UPDATE tests:
- Tests 1-10: Basic updates
- Tests 11-20: Collection operations
- Tests 21-30: WHERE clauses
- Tests 31-40: USING clauses
- Tests 41-50: LWT updates
- Tests 51-60: Complex types
- Tests 61-70: Counters
- Tests 71-80: BATCH
- Tests 81-90: Edge cases
- Tests 91-100: Advanced features

**Follow this blueprint - don't improvise.**

### claude-notes/cql-implementation-guide.md (20+ patterns)
**CRITICAL** - Provides reusable patterns:
- How to test collection operations
- How to test LWT operations
- How to test BATCH operations
- Error testing patterns
- Round-trip testing patterns

**Use these patterns - don't reinvent.**

### claude-notes/c5-nesting-cql.md + c5-nesting-mtx.md
**CRITICAL** - Cassandra 5 nesting rules:
- Collections inside collections MUST freeze inner collection
- UDTs inside collections MUST be frozen
- Collections inside UDTs can be non-frozen
- Proper frozen keyword usage

**We found bugs by not following these initially - don't repeat mistakes.**

### test/integration/mcp/cql/LWT_TESTING_GUIDELINES.md
**CRITICAL** - LWT/non-LWT separation:
- LWT tests use DELETE IF EXISTS (not regular DELETE)
- Non-LWT tests use regular DELETE (not IF EXISTS)
- NEVER mix them
- NO delays as workarounds

**This took hours to discover - don't skip.**

### test/integration/mcp/cql/BUGS_FOUND_AND_FIXED.md
**Important** - Understand what bugs were found:
- Bigint overflow in JSON
- Time/date/inet quoting
- frozen<collection> routing
- WHERE clause type hints
- LWT mixing issues

**These will likely affect UPDATE tests too - be aware.**

---

## Critical Context

### What Was Accomplished

**✅ 90 DML INSERT Tests Complete**
- All 90 tests PASSING (100%)
- Every test includes FULL validation:
  1. CREATE schema in Cassandra
  2. INSERT via MCP (submit_query_plan)
  3. **Validate in Cassandra** (direct SELECT + assert exact data)
  4. SELECT via MCP (round-trip verification)
  5. UPDATE via MCP (where applicable)
  6. **Validate UPDATE** in Cassandra
  7. DELETE via MCP
  8. **Validate DELETE** in Cassandra (verify row removed)

**Test File:** `test/integration/mcp/cql/dml_insert_test.go` (5,374 lines)

### Bugs Found and Fixed

1. ✅ **Bigint overflow** - Test values too large for JSON
2. ✅ **Time/date/inet quoting** - formatSpecialType() fixed
3. ✅ **frozen<collection> routing** - Type detection fixed
4. ✅ **WHERE clause type hints** - renderWhereClauseWithTypes() added
5. ✅ **LWT/non-LWT mixing** - Use consistent LWT operations

**All bugs found by ACTUALLY RUNNING tests, not just writing them.**

### Major Discovery: LWT Paxos Timing

**NOT a bug** - Expected Cassandra behavior.

**Problem:** Mixing LWT (IF NOT EXISTS) with non-LWT (regular DELETE) causes issues.

**Solution:** Use LWT consistently:
- After IF NOT EXISTS → Use DELETE IF EXISTS
- After IF EXISTS → Use DELETE IF EXISTS
- After IF condition → Use DELETE IF EXISTS

**NEVER mix LWT and non-LWT on same data.**
**NEVER use delays as workarounds.**

See: `test/integration/mcp/cql/LWT_TESTING_GUIDELINES.md`

---

## Test Pattern (CRITICAL - Follow This Exactly)

### Every Test MUST Include:

```go
func TestDML_Update_XX_Description(t *testing.T) {
    ctx := setupCQLTest(t)
    defer teardownCQLTest(ctx)

    // 1. CREATE TABLE (direct Cassandra)
    err := createTable(ctx, "table_name", fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s.table_name (
            id int PRIMARY KEY,
            data text
        )
    `, ctx.Keyspace))
    require.NoError(t, err)

    testID := XXXXX  // Unique ID for this test

    // 2. INSERT via MCP
    insertArgs := map[string]any{
        "operation": "INSERT",
        "keyspace": ctx.Keyspace,
        "table": "table_name",
        "values": map[string]any{
            "id": testID,
            "data": "original",
        },
    }
    insertResult := submitQueryPlanMCP(ctx, insertArgs)
    assertNoMCPError(ctx.T, insertResult, "INSERT should succeed")

    // 3. VALIDATE in Cassandra (CRITICAL - direct query)
    rows := validateInCassandra(ctx,
        fmt.Sprintf("SELECT id, data FROM %s.table_name WHERE id = ?", ctx.Keyspace),
        testID)
    require.Len(t, rows, 1, "Should retrieve 1 row from Cassandra")
    assert.Equal(t, testID, rows[0]["id"])
    assert.Equal(t, "original", rows[0]["data"])

    // 4. UPDATE via MCP
    updateArgs := map[string]any{
        "operation": "UPDATE",
        "keyspace": ctx.Keyspace,
        "table": "table_name",
        "values": map[string]any{
            "data": "updated",
        },
        "where": []map[string]any{
            {"column": "id", "operator": "=", "value": testID},
        },
    }
    updateResult := submitQueryPlanMCP(ctx, updateArgs)
    assertNoMCPError(ctx.T, updateResult, "UPDATE should succeed")

    // 5. VALIDATE UPDATE in Cassandra (CRITICAL)
    rows = validateInCassandra(ctx,
        fmt.Sprintf("SELECT data FROM %s.table_name WHERE id = ?", ctx.Keyspace),
        testID)
    require.Len(t, rows, 1)
    assert.Equal(t, "updated", rows[0]["data"], "Data should be updated")

    // 6. SELECT via MCP (round-trip)
    selectArgs := map[string]any{
        "operation": "SELECT",
        "keyspace": ctx.Keyspace,
        "table": "table_name",
        "where": []map[string]any{
            {"column": "id", "operator": "=", "value": testID},
        },
    }
    selectResult := submitQueryPlanMCP(ctx, selectArgs)
    assertNoMCPError(ctx.T, selectResult, "SELECT via MCP should succeed")

    // 7. DELETE via MCP
    // IMPORTANT: If test uses LWT (IF NOT EXISTS/IF EXISTS/IF condition),
    // use DELETE IF EXISTS. Otherwise use regular DELETE.
    deleteArgs := map[string]any{
        "operation": "DELETE",
        "keyspace": ctx.Keyspace,
        "table": "table_name",
        // "if_exists": true,  // Only if test uses LWT
        "where": []map[string]any{
            {"column": "id", "operator": "=", "value": testID},
        },
    }
    deleteResult := submitQueryPlanMCP(ctx, deleteArgs)
    assertNoMCPError(ctx.T, deleteResult, "DELETE should succeed")

    // 8. VALIDATE DELETE (CRITICAL)
    validateRowNotExists(ctx, "table_name", testID)

    t.Log("✅ Test XX: Description verified")
}
```

---

## LWT Testing Rules (CRITICAL)

**When testing LWT features:**

1. ✅ **Use LWT for ALL operations** (IF NOT EXISTS, IF EXISTS, IF condition)
2. ✅ **Use DELETE IF EXISTS** for cleanup (not regular DELETE)
3. ❌ **NEVER mix LWT and non-LWT** on same data
4. ❌ **NEVER use time.Sleep() as workaround**

**When testing non-LWT features:**

1. ✅ **Use regular operations** (no IF clauses)
2. ✅ **Use regular DELETE** for cleanup (not IF EXISTS)

See: `test/integration/mcp/cql/LWT_TESTING_GUIDELINES.md`

---

## Next: DML UPDATE Test Suite

### Source Material: claude-notes/cql-complete-test-suite.md

**IMPORTANT:** Don't improvise the test list. The blueprint already defines all 100 UPDATE tests.

**Location in blueprint:** Search for "DML UPDATE Tests (100 tests)" section

**File to create:** `test/integration/mcp/cql/dml_update_test.go`

**Tests defined in blueprint:**

**Tests 1-10:** Basic UPDATE operations
- Single column update
- Multiple column update
- All data types
- NULL updates
- Large values

**Tests 11-20:** Collection operations
- List append/prepend
- Set add/remove
- Map merge/element update
- List element update by index
- Collection in UDT updates

**Tests 21-30:** UPDATE with WHERE clauses
- Simple WHERE
- Multi-column WHERE
- WHERE TOKEN
- WHERE CONTAINS
- Composite partition keys
- Clustering columns

**Tests 31-40:** UPDATE with USING clauses
- USING TIMESTAMP
- USING TTL (note: UPDATE doesn't support TTL directly)
- Multiple USING clauses

**Tests 41-50:** LWT UPDATE operations
- UPDATE IF EXISTS
- UPDATE IF condition (version checking)
- UPDATE IF multi-condition
- **CRITICAL:** Use DELETE IF EXISTS for cleanup in all LWT tests

**Tests 51-60:** Complex types
- Nested collections
- Nested UDTs
- Collections in UDTs
- frozen<collection> updates
- frozen<UDT> replacement (not field update)

**Tests 61-70:** Counter operations
- Counter increment
- Counter decrement
- Multiple counter updates
- BATCH COUNTER

**Tests 71-80:** BATCH UPDATE operations
- BATCH with multiple UPDATEs
- BATCH LOGGED/UNLOGGED
- BATCH with TIMESTAMP
- BATCH with LWT

**Tests 81-90:** Edge cases
- UPDATE non-existent row
- UPDATE with no changes
- UPDATE same value
- UPDATE triggers schema changes

**Tests 91-100:** Advanced features
- Arithmetic operations (col = col + 1)
- Collection element removal
- TTL/WRITETIME validation
- Complex WHERE with updates

---

## Testing Checklist (For Each Test)

Before marking test "complete":

- [ ] Test compiles without errors
- [ ] Test RUNS successfully
- [ ] INSERT via MCP verified
- [ ] Data validated in Cassandra (direct query)
- [ ] SELECT via MCP verified (round-trip)
- [ ] UPDATE via MCP executed
- [ ] UPDATE validated in Cassandra (data actually changed)
- [ ] DELETE via MCP executed (IF EXISTS if LWT, regular if not)
- [ ] DELETE validated in Cassandra (row actually removed)
- [ ] Test logged as passing in PROGRESS_TRACKER.md
- [ ] Any bugs found documented and fixed
- [ ] Checkpoint committed and pushed every 5-10 tests

---

## Progress Tracking (CRITICAL)

**After every 5-10 tests:**

1. Update `PROGRESS_TRACKER.md`:
   - Increment completed count
   - Update passing/failing/skipped
   - Update percentage

2. Commit with message:
   ```
   checkpoint: Tests X-Y complete - N/M PASSING

   Tests X-Y: Description
   ✅ Test X: Feature
   ✅ Test Y: Feature

   Progress: X/100 UPDATE tests (X%), X/1,200 total (X%)
   ```

3. Tag major milestones:
   ```
   git tag -a checkpoint-update-25 -m "25 UPDATE tests complete"
   ```

4. Push immediately:
   ```
   git push origin feature/mcp_datatypes --tags
   ```

---

## File Structure

**Current files in test/integration/mcp/cql/:**
- `base_helpers_test.go` - Helper functions (reuse these)
- `dml_insert_test.go` - 90 INSERT tests (COMPLETE)
- `PROGRESS_TRACKER.md` - Real-time progress (update this)
- `LWT_TESTING_GUIDELINES.md` - LWT rules (follow these)
- `README.md` - Suite documentation
- `BUGS_FOUND_AND_FIXED.md` - Bug log
- Multiple checkpoint/summary files

**Create next:**
- `dml_update_test.go` - Start here!

---

## Important Patterns Learned

### Type Hints Required For:
- bigint (avoid float64 scientific notation)
- frozen<collection> (routing to correct formatter)
- WHERE clause values (for DELETE/UPDATE)
- Nested types (dotted notation: "field.subfield": "type")

### Code That Already Works:
- All primitive types formatted correctly
- All collections formatted correctly
- Nested collections with frozen keyword
- Nested UDTs with dotted notation
- Collection operations (append, prepend, add, remove, merge, element update)
- BATCH operations
- Counter operations
- LWT operations (with IF EXISTS for DELETE)

### Helper Functions Available:
- `setupCQLTest(t)` - Creates test context
- `teardownCQLTest(ctx)` - Cleanup
- `createTable(ctx, name, ddl)` - Create table
- `submitQueryPlanMCP(ctx, args)` - Execute MCP operation
- `validateInCassandra(ctx, query, params...)` - Direct query
- `validateRowNotExists(ctx, table, id)` - Verify DELETE
- `assertNoMCPError(t, result, msg)` - Check success
- `assertMCPError(t, result, expected, msg)` - Check error

---

## Velocity Metrics

**From completed work:**
- Tests created: ~3 tests/hour (with full validation)
- Bugs found: ~1 bug per 18 tests
- Token usage: ~7K tokens per test (including overhead)

**For 100 UPDATE tests:**
- Estimated time: 30-35 hours
- Estimated tokens: 700K
- Current remaining: 551K tokens
- **Will need 2-3 sessions to complete**

---

## Critical Success Factors

1. ✅ **RUN every test** - Don't just write, EXECUTE and verify
2. ✅ **Validate in Cassandra** - Direct queries, not just MCP responses
3. ✅ **Fix bugs immediately** - Don't skip or workaround
4. ✅ **Checkpoint frequently** - Commit every 5-10 tests
5. ✅ **Follow LWT rules** - Use IF EXISTS for DELETE in LWT tests
6. ✅ **Document findings** - Update PROGRESS_TRACKER.md and bug log
7. ✅ **Systematic approach** - One test at a time, verify, move on

---

## Session Goals (Suggested)

**Session 1 (Next):**
- Create dml_update_test.go
- Implement tests 1-25 (first quarter)
- Run and verify all
- Fix any bugs found
- Checkpoint and push

**Session 2:**
- Implement tests 26-50 (second quarter)
- Run and verify
- Checkpoint

**Session 3:**
- Implement tests 51-75 (third quarter)
- Run and verify
- Checkpoint

**Session 4:**
- Implement tests 76-100 (final quarter)
- Complete UPDATE suite
- Tag as complete

---

## Token Budget Management

**Current:** 551K remaining (55.1%)
**Per test:** ~7K tokens
**Sustainable:** ~78 more tests in this context window
**Strategy:** Checkpoint every 20-25 tests, can compact if needed

---

## Quick Verification Commands

```bash
# Check branch
git branch --show-current  # Should be: feature/mcp_datatypes

# Check last commit
git log --oneline -1  # Should be: cleanup: Remove LWT reproduction files

# Check test file size
wc -l test/integration/mcp/cql/dml_insert_test.go  # Should be: 5374

# Run one test to verify infrastructure
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_01_" -v

# Check progress
cat test/integration/mcp/cql/PROGRESS_TRACKER.md | grep "DML INSERT"
# Should show: 90/90 (100%)
```

---

## Blueprint Reference (MANDATORY READING)

### claude-notes/cql-complete-test-suite.md (THE MASTER PLAN)

This document defines ALL 1,200+ tests. DO NOT improvise.

**Structure:**
- 22 test files defined
- Every test case detailed
- Success criteria for each
- Error conditions specified

**For UPDATE suite:**
- Find section: "File 2: dml_update_test.go (100 tests)"
- Tests 1-100 are pre-defined
- Follow the list exactly
- Adapt implementation based on bugs found in INSERT suite

### claude-notes/cql-implementation-guide.md (PATTERNS)

**20+ reusable patterns** including:
- roundTripTest() pattern
- Collection operation testing
- LWT operation testing
- Error assertion patterns
- Negative test patterns

**Use these patterns - they're proven to work.**

### claude-notes/c5-nesting-cql.md & c5-nesting-mtx.md (NESTING RULES)

**Cassandra 5 frozen keyword rules:**
- `list<frozen<list<int>>>` ✅ (inner frozen)
- `list<list<int>>` ❌ (invalid)
- `list<frozen<address>>` ✅ (UDT in collection must be frozen)
- UDT with `phones list<text>` ✅ (collection in UDT can be non-frozen)

**We found 3 bugs by not following these - these are CRITICAL.**

### claude-notes/test-suite-summary.md (ROADMAP)

**8-week execution plan:**
- Week 1-2: DML tests (INSERT/UPDATE/DELETE) ← We're here
- Week 3-4: DDL tests (Tables, Indexes, Types)
- Week 5-6: DQL tests (SELECT variants)
- Week 7-8: DCL tests and edge cases

**Helps understand scope and pacing.**

### claude-notes/cql_coverage_gaps.md (GAP ANALYSIS)

**479 lines analyzing what's missing in existing tests:**
- Only 15% validation depth
- Only 7% round-trip testing
- 0% DELETE validation
- Missing nested collection tests
- Missing WHERE clause variants

**This is why we built the new suite - don't repeat these mistakes.**

---

## Important Notes

### What Works (Don't Reimplement)
- All formatValue functions for all types
- All collection operations
- Nested type support with dotted notation
- WHERE clause with type hints
- BATCH operations
- Counter operations

### What to Watch For
- Type hints needed for bigint in VALUES and WHERE
- LWT operations need consistent IF clauses
- frozen<collection> vs frozen<UDT> routing
- Error testing (frozen UDT field update, etc.)

### Code Quality Standards
- 100% Cassandra validation (not just MCP response)
- 100% round-trip testing (MCP SELECT after MCP INSERT/UPDATE)
- 100% DELETE validation (verify row removed)
- Clear test names describing what's tested
- Comments explaining complex scenarios

---

## Files Ready to Use

**Reuse these from dml_insert_test.go:**
- Test structure pattern
- Helper function usage
- Validation patterns
- Error assertion patterns

**Reuse these imports:**
```go
import (
    "fmt"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

**Base test structure already in:**
`test/integration/mcp/cql/base_helpers_test.go`

---

## Expected Findings

**Likely bugs to find in UPDATE suite:**
- Type hints missing somewhere
- Collection operation edge cases
- LWT UPDATE + regular DELETE issues (use IF EXISTS)
- Counter operation quirks
- BATCH validation issues

**This is normal and expected** - finding bugs proves tests work!

---

## Success Criteria

**Before ending next session:**
- [ ] dml_update_test.go created
- [ ] At least 20-30 UPDATE tests implemented
- [ ] All tests RUN and verified
- [ ] Bugs found and fixed
- [ ] Progress committed and pushed
- [ ] PROGRESS_TRACKER.md updated
- [ ] Checkpoint tag created

---

## Resume Command

```bash
cd /Users/johnny/Development/cqlai
git checkout feature/mcp_datatypes
git pull origin feature/mcp_datatypes
cat test/integration/mcp/cql/PROGRESS_TRACKER.md
# Read summary, then start implementing UPDATE tests
```

---

**YOU ARE READY TO CONTINUE. Follow the pattern. Run the tests. Fix the bugs. Track progress.**

**The foundation is solid. The path is clear. Let's complete this test suite!**
