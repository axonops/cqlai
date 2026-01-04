# Unit Test Coverage Analysis - Implementation Changes

**Date:** 2026-01-04
**Purpose:** Ensure unit tests cover all implementation changes made during CQL assertion work
**Status:** ⚠️ GAPS IDENTIFIED - Unit tests need updates

---

## Implementation Changes Made

### File 1: internal/ai/planner.go

| Change | Lines | Description | Unit Test Status |
|--------|-------|-------------|------------------|
| Add sort import | 6 | Added "sort" package | ✅ N/A |
| **Sort INSERT columns** | 237-255 | Alphabetical column ordering | ❌ NOT TESTED |
| **Sort UPDATE counter columns** | 314-331 | Alphabetical counter op ordering | ❌ NOT TESTED |
| **Sort UPDATE collection columns** | 337-344 | Alphabetical collection op ordering | ❌ NOT TESTED |
| **Sort UPDATE value columns** | 395-410 | Alphabetical value ordering | ❌ NOT TESTED |
| **Sort map keys** | 2553-2579 | Alphabetical map key ordering | ❌ NOT TESTED |
| **Sort UDT fields** | 2601-2633 | Alphabetical UDT field ordering | ❌ NOT TESTED |
| **Sort set elements** | 2533-2536 | Alphabetical set element ordering | ❌ NOT TESTED |
| **BATCH semicolons** | 1661-1666 | Keep semicolons (was removing) | ❌ **WRONG TEST** |
| **WHERE IN operator** | 702-710, 735-743 | Handle Values field for IN | ❌ NOT TESTED |

### File 2: internal/ai/mcp.go

| Change | Lines | Description | Unit Test Status |
|--------|-------|-------------|------------------|
| **Parse WHERE values field** | 1426-1429 | For IN operator | ❌ NOT TESTED |
| **Parse WHERE is_token field** | 1430-1433 | For TOKEN() wrapper | ❌ NOT TESTED |
| **Parse WHERE columns field** | 1434-1442 | For tuple notation | ❌ NOT TESTED |

---

## Existing Unit Tests Found

**Planner tests:**
- `planner_test.go` - Main rendering tests
- `planner_where_test.go` - WHERE clause tests
- `planner_batch_test.go` - BATCH tests
- `planner_types_test.go` - Type rendering tests
- `planner_udt_test.go` - UDT tests
- `planner_advanced_test.go` - Advanced features
- `planner_delete_test.go` - DELETE tests
- `planner_ddl_test.go` - DDL tests
- `planner_batch_lwt_test.go` - BATCH LWT tests
- `planner_mcp_simulation_test.go` - MCP simulation

**MCP tests:**
- Various MCP-related test files

---

## Critical Gaps Identified

### Gap 1: NO Deterministic Rendering Tests ❌

**What's missing:**
- Test that INSERT columns are sorted alphabetically
- Test that UPDATE SET clauses are sorted alphabetically
- Test that map keys are sorted alphabetically
- Test that UDT fields are sorted alphabetically
- Test that set elements are sorted alphabetically
- Test that same logical query produces identical string

**Impact:** Changes could be broken without detection

**Example needed:**
```go
func TestRenderInsert_DeterministicColumnOrder(t *testing.T) {
    plan := &AIResult{
        Operation: "INSERT",
        Table: "users",
        Values: map[string]any{
            "z_col": "last",
            "a_col": "first",
            "m_col": "middle",
        },
    }

    got, _ := RenderCQL(plan)

    // Columns should be alphabetically sorted: a_col, m_col, z_col
    assert.Contains(t, got, "(a_col, m_col, z_col)")
    assert.Contains(t, got, "VALUES ('first', 'middle', 'last')")
}
```

---

### Gap 2: BATCH Semicolon Test is WRONG ❌

**File:** `internal/ai/planner_batch_test.go` line 45

**Current test (INCORRECT):**
```go
assert.NotContains(t, line, ";", "Batch statements should not have semicolons")
```

**This is WRONG!** BATCH statements **MUST** have semicolons.

**Should be:**
```go
assert.Contains(t, line, ";", "Batch statements MUST have semicolons")
```

**Impact:** Test enforces incorrect behavior

**Fix needed:**
1. Update test to expect semicolons
2. Add test for compacted BATCH (single line)
3. Verify BATCH is valid CQL

---

### Gap 3: NO WHERE IN Tests ❌

**File:** `internal/ai/planner_where_test.go`

**What's missing:**
- Test WHERE IN with single value
- Test WHERE IN with multiple values
- Test WHERE IN with empty list
- Test that Values field (plural) is used

**Example needed:**
```go
func TestRenderSelect_WhereIN(t *testing.T) {
    plan := &AIResult{
        Operation: "SELECT",
        Table: "users",
        Where: []WhereClause{
            {
                Column: "id",
                Operator: "IN",
                Values: []any{1, 2, 3},
            },
        },
    }

    got, _ := RenderCQL(plan)

    assert.Contains(t, got, "WHERE id IN (1, 2, 3)")
}
```

---

### Gap 4: NO TOKEN() Wrapper Tests ❌

**File:** `internal/ai/planner_where_test.go`

**What's missing:**
- Test WHERE with TOKEN()
- Test that IsToken field applies TOKEN() wrapper

**Example needed:**
```go
func TestRenderSelect_WhereToken(t *testing.T) {
    plan := &AIResult{
        Operation: "SELECT",
        Table: "users",
        Where: []WhereClause{
            {
                Column: "id",
                Operator: ">",
                Value: 100,
                IsToken: true,
            },
        },
    }

    got, _ := RenderCQL(plan)

    assert.Contains(t, got, "WHERE TOKEN(id) > 100")
}
```

---

### Gap 5: NO Tuple Notation Tests ❌

**File:** `internal/ai/planner_where_test.go`

**What's missing:**
- Test WHERE with tuple notation
- Test (col1, col2) > (val1, val2) syntax

**Example needed:**
```go
func TestRenderSelect_WhereTupleNotation(t *testing.T) {
    plan := &AIResult{
        Operation: "SELECT",
        Table: "users",
        Where: []WhereClause{
            {
                Columns: []string{"user_id", "timestamp"},
                Operator: ">",
                Values: []any{1000, 500},
            },
        },
    }

    got, _ := RenderCQL(plan)

    assert.Contains(t, got, "WHERE (user_id, timestamp) > (1000, 500)")
}
```

---

### Gap 6: NO MCP WHERE Parsing Tests ❌

**What's missing:**
- Test parseSubmitQueryPlanParams parses "values" field
- Test parseSubmitQueryPlanParams parses "is_token" field
- Test parseSubmitQueryPlanParams parses "columns" field

**File:** Need to check if `internal/ai/mcp_test.go` or similar exists

---

## Summary of Required Unit Tests

### CRITICAL Priority (Must Add)

**1. Deterministic Rendering Tests (7 tests)**
- TestRenderInsert_DeterministicColumnOrder
- TestRenderUpdate_DeterministicSetClauseOrder
- TestFormatMap_DeterministicKeyOrder
- TestFormatUDT_DeterministicFieldOrder
- TestFormatSet_DeterministicElementOrder
- TestRenderInsert_MultipleRunsSameOutput (idempotency)
- TestRenderUpdate_CounterOpsSorted

**2. BATCH Semicolon Tests (2 tests)**
- Fix existing test to expect semicolons (line 45 in planner_batch_test.go)
- TestRenderBatch_ProperSemicolons (verify semicolons present)

**3. WHERE Clause Tests (3 tests)**
- TestRenderSelect_WhereIN
- TestRenderSelect_WhereToken
- TestRenderSelect_WhereTupleNotation

**4. MCP Parsing Tests (3 tests)**
- TestParseWhereClause_ValuesField
- TestParseWhereClause_IsTokenField
- TestParseWhereClause_ColumnsField

**Total:** 15 new unit tests needed

---

## Action Plan

### Step 1: Fix Wrong BATCH Test
File: `internal/ai/planner_batch_test.go` line 40-48

**Change from:**
```go
// Statements should NOT have semicolons inside batch
assert.NotContains(t, line, ";", "Batch statements should not have semicolons")
```

**Change to:**
```go
// Statements MUST have semicolons for valid CQL
assert.Contains(t, line, ";", "Batch statements MUST have semicolons")
```

### Step 2: Add WHERE Clause Tests
File: `internal/ai/planner_where_test.go`

Add 3 new tests:
- TestRenderSelect_WhereIN
- TestRenderSelect_WhereToken
- TestRenderSelect_WhereTupleNotation

### Step 3: Add Deterministic Rendering Tests
File: `internal/ai/planner_test.go` or new file `internal/ai/planner_deterministic_test.go`

Add 7 tests for alphabetical sorting

### Step 4: Add MCP Parsing Tests
File: Check if `internal/ai/mcp_test.go` exists, or add to appropriate test file

Add 3 tests for WHERE clause field parsing

### Step 5: Run All Unit Tests
```bash
go test ./internal/ai/... -v
```

Verify:
- All new tests pass
- Fixed BATCH test passes
- No regressions

---

## Why This Matters

1. **Faster feedback** - Unit tests run in seconds vs minutes for integration tests
2. **Precise isolation** - Know exactly what broke
3. **CI/CD** - Unit tests run on every commit
4. **Documentation** - Tests document expected behavior
5. **Regression prevention** - Changes won't break silently

**Without unit tests:** Next developer could accidentally revert deterministic sorting and integration tests would fail mysteriously.

**With unit tests:** Unit test fails immediately: "TestDeterministicColumnOrder FAILED - columns not sorted"

---

**NEXT: Implement the 15 missing unit tests**
