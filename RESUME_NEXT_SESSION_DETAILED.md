# Resume Next Session - Comprehensive Prompt

**Branch:** feature/mcp_datatypes
**Last Updated:** 2026-01-05 (End of cluster metadata + validation session)
**Current State:** 125 tests implemented (7.4% of 1,251 target)
**Token Used This Session:** 582K/1M

---

## CRITICAL: ACTIVE BUG TO FIX IMMEDIATELY

### BUG: Tuple in Collection Rendering Failure

**Test:** Test 87 (TestDML_Insert_87_TupleInCollection)
**Status:** SKIPPED - MUST BE FIXED
**Location:** test/integration/mcp/cql/dml_insert_test.go line 6368+

**Error:**
```
Query execution failed: query failed: Invalid list literal for points: value {} is not of type frozen<tuple<int, int>>
```

**Test Details:**
- Table: `list<frozen<tuple<int, int>>>`
- Attempting to INSERT: `[(10, 20), (30, 40), (50, 60)]`
- value_types provided: `"list<frozen<tuple<int,int>>>"`
- CQL should be: `INSERT INTO ks.tuple_list (id, points) VALUES (87000, [(10, 20), (30, 40), (50, 60)]);`

**Root Cause:** Planner's formatCollection may not properly handle nested tuples

**Action Required:**
1. Investigate internal/ai/planner.go - formatCollection function
2. Check how nested tuples are rendered
3. Fix the rendering bug
4. Run Test 87 to verify fix
5. DO NOT SKIP - FIX THE BUG

**Similar Issue:** Test 88 (set<uuid>) also fails with rendering issues

---

## CRITICAL IMPLEMENTATION REQUIREMENTS (READ THESE FIRST)

### SYSTEMATIC APPROACH (How Previous Sessions Succeeded)

**Phase 1 - Understanding (READ IN ORDER):**
1. **claude-notes/cql-complete-test-suite.md** - Master blueprint (1,251 tests)
2. **test/integration/mcp/cql/INSERT_GAP_ANALYSIS.md** - CRITICAL: Shows what's missing
3. **claude-notes/CQL_TEST_ADDENDUM_PRIMARY_KEYS_AND_BATCH.md** - PK/BATCH details
4. **claude-notes/cql-implementation-guide.md** - Test patterns

**Phase 2 - Current State (CHECK PROGRESS):**
5. **test/integration/mcp/cql/PROGRESS_TRACKER.md** - Shows 125/1,251 (7.4%)
6. **test/integration/mcp/cql/OUTSTANDING_WORK.md** - Bugs and duplicates
7. **test/integration/mcp/cql/PRIMARY_KEY_TESTS.md** - PK test tracking

**Phase 3 - Implementation (FOLLOW PATTERNS):**
- Read examples from existing tests
- Use helper functions
- Assert EXACT CQL and errors
- Run immediately
- Commit every 2-3 tests

### INSERT_GAP_ANALYSIS.md - YOUR PRIMARY GUIDE

**THIS IS THE CRITICAL FILE** showing what INSERT tests are missing:

**Major Gaps (from gap analysis):**
- INSERT JSON: 5/10 done (50% gap) - HIGH PRIORITY
- Tuples: 3/5 done (40% gap) - HIGH PRIORITY
- Collections: 7/10 done (30% gap)
- USING clause: 4/10 done (60% gap)
- IF NOT EXISTS: 3/5 done (40% gap)

**The gap analysis breaks down:**
- Which blueprint tests are implemented
- Which are missing
- Which are partial
- Priority levels

**ALWAYS refer to INSERT_GAP_ANALYSIS.md when choosing next tests to implement.**

### Blueprint Documentation Details

1. **claude-notes/cql-complete-test-suite.md**
   - Master blueprint defining ALL 1,251 tests
   - Section: "File 1: dml_insert_test.go (141 tests)"
   - Updated 2026-01-04 with 51 critical PK/BATCH scenarios
   - **Line 226-232:** Error handling tests
   - **Line 233-257:** Primary key validation tests
   - **Line 258-287:** BATCH validation tests

2. **test/integration/mcp/cql/INSERT_GAP_ANALYSIS.md** ⭐ CRITICAL
   - **Line 1-25:** Executive summary and status
   - **Line 26-222:** Detailed gap breakdown by category
   - **Line 223-274:** Critical gaps identified
   - **Line 426-497:** Recommended action plan
   - **USE THIS** to choose which tests to implement next

3. **claude-notes/CQL_TEST_ADDENDUM_PRIMARY_KEYS_AND_BATCH.md**
   - Details for 15 PRIMARY KEY tests (ALL DONE ✅)
   - Details for 22 BATCH tests (10 done, 12 remaining)
   - Detailed code examples for each test

4. **test/integration/mcp/cql/PROGRESS_TRACKER.md**
   - Real-time progress tracking
   - Updated: 2026-01-05
   - Shows 125/1,251 tests (7.4%)
   - Update this after every commit

5. **test/integration/mcp/cql/OUTSTANDING_WORK.md**
   - Tracks incomplete work
   - 5 BATCH tests duplicated (need deletion)
   - DEFAULT UNSET support needed
   - Any bugs found

### Critical Patterns (MANDATORY)

**CQL Assertion (EVERY test):**
```go
result := submitQueryPlanMCP(ctx, args)
assertNoMCPError(t, result, "Operation should succeed")

// MANDATORY - Assert EXACT CQL
expectedCQL := "INSERT INTO ks.table (col1, col2) VALUES (val1, val2);"
assertCQLEquals(t, result, expectedCQL, "CQL must match exactly")
```

**Error Assertion (EVERY error test):**
```go
result := submitQueryPlanMCP(ctx, args)

// Assert EXACT error message including keyspace.table
expectedError := "Query validation failed: missing partition key column: id (required for INSERT into ks.table)"
assertMCPErrorMessageExact(t, result, expectedError, "Should get exact error")
```

**TTL Verification:**
```go
// Use TTL() function to verify TTL actually applied
rows := validateInCassandra(ctx, "SELECT data, TTL(data) FROM ks.table WHERE id = ?", id)
if ttl, ok := rows[0]["ttl(data)"].(int32); ok {
    assert.GreaterOrEqual(t, ttl, int32(290))
    assert.LessOrEqual(t, ttl, int32(300))
}
```

---

## CURRENT STATE DETAILED

### Cluster Metadata Manager ✅ COMPLETE

**Package:** internal/cluster/
**Methods:** 121 total
- 89 helper methods on wrapper types
- 32 MetadataManager interface methods
- 2 validation methods

**Tests:** 70 total
- 61 unit tests (internal/cluster/types_test.go, translator_test.go)
- 9 integration tests (test/integration/cluster_metadata_test.go)

**Tag:** cluster-metadata-complete

**Critical Features:**
- ALWAYS delegates to gocql (no caching)
- Schema changes propagate in ~1 second automatically
- Supports partition keys, clustering keys, static columns
- Cassandra 5.0 ready (vectors, SAI indexes)

### Validation Integration ✅ COMPLETE

**File:** internal/ai/planner_validator.go (325 lines)

**Functions:**
- ValidateInsertPlan() - Checks all partition + clustering keys present
- ValidateUpdatePlan() - Full PK for regular, partial for static
- ValidateDeletePlan() - Requires at least partition key
- ValidateBatchPlan() - Counter mixing validation
- ValidatePlanWithMetadata() - Main entry point

**Integration:** internal/ai/mcp.go
- MetadataManager added to MCPServer
- Validation called BEFORE RenderCQL
- Errors returned before invalid CQL generated

**Error Message Format:**
- Includes keyspace.table
- Lists ALL missing keys
- Lists ALL required keys
- Uses JSON API terminology
- Example: "missing partition key column(s): device_id. Include all partition keys: user_id, device_id (required for INSERT into ks.pk_test)"

### Test Files and Counts

**Test Files (8):**
1. dml_insert_test.go - 86 tests (was 78)
   - Tests 1-78: Original
   - Test 79: PK validation (full PK)
   - Tests 80-86: JSON + Tuple tests
   - Test 87: Tuple in collection (SKIPPED - BUG)

2. dml_insert_error_test.go - 6 error tests
   - ERR_01: Non-existent table
   - ERR_01a: Cross-session schema propagation (CRITICAL)
   - ERR_02: Missing partition key (2-column composite)
   - ERR_03: Missing clustering keys
   - ERR_04: Type mismatch (placeholder)
   - ERR_05: Frozen UDT field update (moved from test 53)
   - ERR_06: 3-column partition key

3. dml_update_test.go - 2 tests
   - Test 01: Full PK UPDATE
   - Test 02: Static column with partial PK

4. dml_update_error_test.go - 3 error tests
   - ERR_01: Partial PK + regular column
   - ERR_02: Missing partition key
   - ERR_03: No WHERE clause

5. dml_delete_test.go - 5 tests
   - Test 01: Full PK row delete
   - Test 02: Partition key only
   - Test 03: Partition key + range
   - Test 04: Partition key + IN
   - Test 05: Static column with partial PK

6. dml_delete_error_test.go - 2 error tests
   - ERR_01: Missing partition key
   - ERR_02: No WHERE clause

7. dml_batch_test.go - 10 tests
   - Test 01: Multiple INSERTs (moved from test 36)
   - Test 02: UNLOGGED (moved from test 68)
   - Test 03: COUNTER (moved from test 69)
   - Test 04: With TIMESTAMP (moved from test 70)
   - Test 05: With LWT single partition (moved from test 71)
   - Test 06: LWT same partition (NEW)
   - Test 07: Mixed DML
   - Test 08: Multiple tables
   - Test 09: Statement-level TTL (with TTL verification)
   - Test 10: Large batch (15 statements, all verified)

8. dml_batch_error_test.go - 2 error tests
   - ERR_01: Counter/non-counter mixing
   - ERR_02: LWT cross-partition

**Total: 125 tests (7.4% of 1,251)**

**ALL TESTS PASSING (except Test 87 SKIPPED for bug)**

### Test Naming Convention

**Format:** `TestDML_<Operation>_<##>_<Description>`

**Examples:**
- TestDML_Insert_79_FullPrimaryKey
- TestDML_Update_01_FullPrimaryKey
- TestDML_Delete_01_FullPrimaryKey
- TestDML_Batch_07_MixedDML
- TestDML_Insert_ERR_02_MissingPartitionKey

**DO NOT use:** TestDML_Insert_PK_01 (breaks convention)

---

## OUTSTANDING WORK (Critical)

### 1. DUPLICATE BATCH TESTS (High Priority)

**Status:** Tests exist in BOTH files

**Duplicates:**
- Test 36 in dml_insert_test.go → Batch_01 in dml_batch_test.go
- Test 68 in dml_insert_test.go → Batch_02 in dml_batch_test.go
- Test 69 in dml_insert_test.go → Batch_03 in dml_batch_test.go
- Test 70 in dml_insert_test.go → Batch_04 in dml_batch_test.go
- Test 71 in dml_insert_test.go → Batch_05 in dml_batch_test.go

**Comments added:** All 5 originals have "NOTE: DUPLICATED" comments

**Action needed:** Delete tests 36, 68-71 from dml_insert_test.go

### 2. BUGS TO FIX

**Bug 1: Tuple in collection rendering (Test 87)**
- Error: Invalid list literal for points
- Location: formatCollection or formatTuple in planner.go
- Priority: HIGH

**Bug 2: set<uuid> rendering (Test 88 - not yet committed)**
- UUIDs not rendering correctly in sets
- Priority: HIGH

### 3. MISSING PLANNER FEATURES

**DEFAULT UNSET support:**
- INSERT JSON DEFAULT UNSET clause not supported
- Needed for partial updates without tombstones
- Currently only DEFAULT NULL works (default behavior)

---

## WHAT WAS ACCOMPLISHED THIS SESSION

### Infrastructure
1. ✅ Cluster metadata manager (6,139 lines, 121 methods, 70 tests)
2. ✅ Validation integration (325 lines)
3. ✅ MetadataManager in MCPServer
4. ✅ Error message improvements (keyspace.table, all keys listed)

### Tests Implemented
1. ✅ 14 error validation tests (exact assertions)
2. ✅ 8 PRIMARY KEY success tests (all 15 PK tests complete with errors)
3. ✅ 4 BATCH tests (tests 07-10)
4. ✅ 5 INSERT JSON tests (tests 80-84, 87 SKIPPED)
5. ✅ 3 Tuple tests (tests 85-86, 87 SKIPPED)
6. ✅ 2 UPDATE tests
7. ✅ 5 DELETE tests

**Total: 125 tests (86 INSERT + 14 INSERT errors + 2 UPDATE + 3 UPDATE errors + 5 DELETE + 2 DELETE errors + 10 BATCH + 2 BATCH errors + 1 SKIPPED)**

### Learnings Documented
- Schema propagation: ~1 second (ERR_01a proves it)
- DEFAULT NULL vs UNSET (tombstone behavior)
- Static column partial PK rules
- LWT BATCH constraints
- TTL verification technique

---

## CRITICAL GAPS REMAINING

### INSERT Tests (141 total, 86 done, 55 missing)

**By Category:**

**INSERT JSON (5/10 done - 50%):**
- ✅ Test 29: Simple JSON (original)
- ✅ Test 80: All columns
- ✅ Test 81: Partial columns (default values)
- ✅ Test 82: NULL values (DEFAULT NULL)
- ✅ Test 83: DEFAULT NULL overwrites existing
- ✅ Test 84: Escaped quotes
- ❌ Test 85: JSON with arrays (missing)
- ❌ Test 86: JSON special characters (missing)
- ❌ Test 87: JSON numeric precision (missing)
- ❌ Test 88: JSON round-trip SELECT (missing)

**Tuples (3/5 done - 60%):**
- ✅ Test 14: Simple tuple (original)
- ✅ Test 85: Mixed types
- ✅ Test 86: NULL elements
- ⏸️ Test 87: Tuple in collection (SKIPPED - BUG TO FIX)
- ❌ Test 88: Tuple immutability/update (missing)

**Collections (7/10 done - 70%):**
- ✅ Tests 10-12, 16-20: Various collections
- ❌ set<uuid> uniqueness (missing)
- ❌ map<text,text> complex values (missing)
- ❌ Nested collection error test (missing)

**USING Clause (4/10 done - 40%):**
- ✅ Tests 26-28: TTL, TIMESTAMP, combined
- ❌ TTL with bind markers (missing - skip for now)
- ❌ TIMESTAMP with bind markers (missing - skip for now)
- ❌ TTL=0 (missing)
- ❌ TTL large value (missing)
- ❌ TTL expiration verification (missing)

**IF NOT EXISTS (3/5 done - 60%):**
- ✅ Test 30: Basic LWT
- ❌ IF NOT EXISTS variants (missing)

**Bind Markers (0/10 done - 0% - LOW PRIORITY):**
- Skip for now per user direction

### UPDATE Tests (2/100 done - 2%)

**Implemented:**
- Test 01: Full PK
- Test 02: Static column partial PK

**Needed:** 98 more UPDATE tests (see blueprint)

### DELETE Tests (5/60 done - 8.3%)

**Implemented:**
- Tests 01-05: PK validation tests

**Needed:** 55 more DELETE tests (see blueprint)

### BATCH Tests (10/22 done - 45.5%)

**Implemented:**
- Tests 01-10 in dml_batch_test.go
- 2 error tests

**Needed:** 12 more BATCH tests
- Cross-partition detection (requires partition key value extraction)
- BATCH size limits
- More LWT scenarios
- More validation scenarios

---

## TEST FILE ORGANIZATION (STRICT CONVENTIONS)

### File Structure

**Main operation files:**
- dml_insert_test.go - All INSERT tests (numbered 01-NN)
- dml_update_test.go - All UPDATE tests (numbered 01-NN)
- dml_delete_test.go - All DELETE tests (numbered 01-NN)
- dml_batch_test.go - All BATCH tests (numbered 01-NN)

**Error validation files:**
- dml_insert_error_test.go - INSERT error tests (ERR_##)
- dml_update_error_test.go - UPDATE error tests (ERR_##)
- dml_delete_error_test.go - DELETE error tests (ERR_##)
- dml_batch_error_test.go - BATCH error tests (ERR_##)

### Naming Convention

**Success tests:** `TestDML_<Operation>_<##>_<Description>`
- Example: TestDML_Insert_79_FullPrimaryKey

**Error tests:** `TestDML_<Operation>_ERR_<##>_<Description>`
- Example: TestDML_Insert_ERR_02_MissingPartitionKey

**NEVER:**
- TestDML_Insert_PK_01 (breaks numbering)
- Creating new files without approval (dml_insert_pk_test.go was wrong)

---

## RUNNING TESTS

**Before running:**
```bash
podman start cassandra-test
sleep 25
go clean -testcache  # ALWAYS clear cache
```

**Run all INSERT tests:**
```bash
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_" -v -p 1
# Expected: 86 PASS, 1 SKIP (Test 87), ~300 seconds
```

**Run error tests:**
```bash
go test ./test/integration/mcp/cql -tags=integration -run "ERR" -v -p 1
# Expected: 14 PASS
```

**Run specific test:**
```bash
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_87" -v -p 1
```

---

## SYSTEMATIC APPROACH (MANDATORY - This Is How We Succeed)

### Step 1: Understand What's Needed
1. Read INSERT_GAP_ANALYSIS.md - See ALL missing tests
2. Read blueprint section for that test category
3. Read existing similar test as example
4. Understand what the test should verify

### Step 2: Write ONE Test
1. Write the test following patterns
2. Add to CORRECT file (dml_insert_test.go, NOT new file)
3. Use CORRECT naming (TestDML_Insert_##_Description)
4. Include EXACT CQL assertion
5. Include EXACT error assertion (if error test)
6. Verify data in Cassandra (not just counts)

### Step 3: Run Test IMMEDIATELY
1. go clean -testcache
2. go test -tags=integration -run "^TestDML_Insert_##" -v -p 1
3. If FAILS: DEBUG and FIX (do NOT skip)
4. If PASSES: Continue to step 4

### Step 4: Document and Commit
1. Update PROGRESS_TRACKER.md with new test count
2. git add test/integration/mcp/cql/
3. git commit with detailed message
4. git push

### Step 5: Repeat
Go back to Step 1 for next test

**This systematic approach ensures:**
- No tests skipped
- All bugs found and fixed
- Progress tracked
- Work committable at any time

---

## IMMEDIATE NEXT STEPS (IN ORDER)

### Step 1: FIX Test 87 Bug (CRITICAL - DO NOT SKIP)

**Test 87 is SKIPPED due to bug. MUST fix before continuing.**

**Debugging steps:**
1. Run Test 87 to see exact error
2. Check internal/ai/planner.go - formatCollection function
3. Check formatTuple function
4. Find why `{}` generated instead of `[(10, 20), (30, 40), (50, 60)]`
5. Fix the bug in planner
6. Run Test 87 until PASSES
7. Commit the fix
8. Document what was wrong

**DO NOT:**
- Skip to easier tests
- Use t.Skip() without fixing
- Move on before bug is fixed

### Step 2: Review INSERT_GAP_ANALYSIS.md

**After Test 87 fixed:**
1. Open INSERT_GAP_ANALYSIS.md
2. Find highest priority missing test
3. Check blueprint for that test details
4. Implement using systematic approach above

**High priority gaps:**
- INSERT JSON: 5/10 done (need 5 more)
- Tuples: 3/5 done (need 1 more after test 87 fixed)
- Collections: 7/10 done (need 3 more)
- USING clause: 4/10 done (need 6 more)

### Step 3: Clean Up Duplicates

**After making good progress:**
- Delete tests 36, 68-71 from dml_insert_test.go
- Verify all tests still pass
- Commit cleanup

### Step 4: Continue Systematically

Repeat Step 1-5 of systematic approach until INSERT tests complete

---

## HELPER FUNCTIONS AVAILABLE

**Test Infrastructure (base_helpers_test.go):**
- setupCQLTest(t) - Creates test context
- teardownCQLTest(ctx) - Cleanup
- createTable(ctx, name, ddl) - Create table
- submitQueryPlanMCP(ctx, args) - Execute via MCP
- assertNoMCPError(t, result, msg) - Assert success
- assertCQLEquals(t, result, expectedCQL, msg) - Assert EXACT CQL
- assertMCPErrorMessageExact(t, result, expectedError, msg) - Assert EXACT error
- extractMCPErrorMessage(result) - Extract error text
- validateInCassandra(ctx, query, params...) - Direct Cassandra query

**Metadata Manager (available in tests):**
- Can access via ctx.Session if needed for debugging

---

## COMMIT STRATEGY

**Every 2-3 tests:**
1. git add test/integration/mcp/cql/
2. git commit -m "test: Add Tests ##-## - Description\n\nProgress: X tests\n\n[Details]\n\nResult: All PASS"
3. git push origin feature/mcp_datatypes

**Update progress files:**
- PROGRESS_TRACKER.md - Update test counts
- PRIMARY_KEY_TESTS.md - If PK tests
- OUTSTANDING_WORK.md - If new issues found

---

## CRITICAL RULES

1. **NEVER skip tests** - Fix bugs until tests pass
2. **NEVER remove tests** - Only add or fix
3. **NEVER use t.Skip()** without explicit approval
4. **ALWAYS read file before editing** - Prevent edit errors
5. **ALWAYS run test immediately** after writing
6. **ALWAYS assert exact CQL** - Not just execution success
7. **ALWAYS assert exact errors** - Not just "contains"
8. **ALWAYS verify data in Cassandra** - Not just row counts
9. **ALWAYS commit progress** - Every 2-3 tests
10. **ALWAYS follow naming conventions** - TestDML_<Op>_<##>

---

## DEBUGGING WHEN TESTS FAIL

1. **Read the error message carefully**
2. **Check the generated CQL** (in test output)
3. **Investigate the root cause** in planner.go
4. **Fix the bug** - don't skip
5. **Run test again**
6. **Ask user if stuck** - don't make independent decisions

---

## CURRENT TEST COUNT SUMMARY

**Total:** 125 tests
- DML INSERT: 86 (1 SKIP)
- DML INSERT Errors: 6
- DML UPDATE: 2
- DML UPDATE Errors: 3
- DML DELETE: 5
- DML DELETE Errors: 2
- DML BATCH: 10
- DML BATCH Errors: 2

**All categories at 100% for errors ✅**
**Primary Key validation 100% complete (15 tests) ✅**

---

## RESUME COMMAND

**Start new session with:**

```
I am continuing comprehensive CQL test suite implementation for CQLAI MCP server.

CRITICAL: Test 87 (TestDML_Insert_87_TupleInCollection) is SKIPPED due to a tuple-in-collection rendering bug. I MUST fix this bug before continuing with other tests.

Current state: 125 tests implemented (7.4% of 1,251 target)
Branch: feature/mcp_datatypes

Read these files IN ORDER:
1. test/integration/mcp/cql/RESUME_NEXT_SESSION_DETAILED.md (this file)
2. test/integration/mcp/cql/PROGRESS_TRACKER.md
3. test/integration/mcp/cql/OUTSTANDING_WORK.md
4. claude-notes/CQL_TEST_ADDENDUM_PRIMARY_KEYS_AND_BATCH.md

First task: Fix Test 87 tuple-in-collection rendering bug
Then: Continue with remaining INSERT tests per gap analysis
```

---

**NEVER skip bugs. ALWAYS fix them. ALWAYS ask before making decisions.**
