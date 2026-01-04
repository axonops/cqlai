# DML INSERT Tests - Gap Analysis

**Date:** 2026-01-04
**Blueprint Target:** 90 tests
**Actually Implemented:** 78 tests
**Missing:** 12 tests (13.3% gap)
**Status:** ⚠️ INCOMPLETE - Gaps identified

---

## Executive Summary

The blueprint defined **90 specific INSERT test scenarios**. The implementation has **78 test functions**, with test #78 being a catch-all `TestDML_Insert_Remaining` that bundles multiple scenarios.

**CRITICAL GAPS:**
- Missing dedicated tests for bind markers (10 scenarios defined)
- Missing dedicated tests for TTL variants (several scenarios)
- Missing dedicated tests for error scenarios (10 scenarios defined)
- Some scenarios merged or simplified

---

## Detailed Gap Analysis

### Blueprint Section 1: Basic INSERT Operations (Tests 1-5)

| Blueprint # | Scenario | Implemented | Test Name | Status |
|-------------|----------|-------------|-----------|--------|
| 1.1 | Single column table INSERT | ✅ | Test 01: SimpleText | DONE |
| 1.2 | Multi-column INSERT | ✅ | Test 02: MultipleColumns | DONE |
| 1.3 | Partial column INSERT (sparse) | ⚠️ | Test 02 (partial) | PARTIAL |
| 1.4 | INSERT with all NULL values (except key) | ✅ | Test 32: NullValues | DONE |
| 1.5 | INSERT IF NOT EXISTS (success/failure) | ✅ | Test 30: IfNotExists | DONE |

**Coverage: 5/5 ✅**

---

### Blueprint Section 2: INSERT with Primitives (Tests 6-25, 20 types)

| Blueprint # | Type | Implemented | Test Name | Status |
|-------------|------|-------------|-----------|--------|
| 1.6 | ascii | ⚠️ | Merged into Test 02? | MISSING |
| 1.7 | bigint | ✅ | Test 03: AllIntegerTypes | DONE |
| 1.8 | blob | ✅ | Test 06: Blob | DONE |
| 1.9 | boolean | ✅ | Test 05: Boolean | DONE |
| 1.10 | counter | ✅ | Test 37: CounterColumn | DONE |
| 1.11 | date | ✅ | Test 08: DateTimeTypes | DONE |
| 1.12 | decimal | ✅ | Test 04: AllFloatTypes | DONE |
| 1.13 | double | ✅ | Test 04: AllFloatTypes | DONE |
| 1.14 | duration | ✅ | Test 73: DurationAllFormats | DONE |
| 1.15 | float | ✅ | Test 04: AllFloatTypes | DONE |
| 1.16 | inet | ✅ | Test 09: Inet, Test 72: IPv6 | DONE |
| 1.17 | int | ✅ | Test 03: AllIntegerTypes | DONE |
| 1.18 | smallint | ✅ | Test 03: AllIntegerTypes | DONE |
| 1.19 | text | ✅ | Test 01, 33, 41, 42 | DONE |
| 1.20 | time | ✅ | Test 08: DateTimeTypes, Test 77 | DONE |
| 1.21 | timestamp | ✅ | Test 08: DateTimeTypes | DONE |
| 1.22 | timeuuid | ✅ | Test 07: UUIDTypes | DONE |
| 1.23 | tinyint | ✅ | Test 03: AllIntegerTypes | DONE |
| 1.24 | uuid | ✅ | Test 07: UUIDTypes | DONE |
| 1.25 | varchar | ⚠️ | Text aliases | MISSING dedicated test |
| 1.26 | varint | ✅ | Test 76: VarIntLargeValue | DONE |
| 1.27 | vector | ✅ | Test 15, Test 74 | DONE |

**Coverage: 20/22 primitive scenarios** (ascii, varchar missing dedicated tests)

---

### Blueprint Section 3: INSERT with Collections (Tests 26-35)

| Blueprint # | Scenario | Implemented | Test Name | Status |
|-------------|----------|-------------|-----------|--------|
| 1.26 | list<int>: empty, single, multiple | ✅ | Test 10, 31 | DONE |
| 1.27 | list<text>: special chars, Unicode | ✅ | Test 41, 42 | DONE |
| 1.28 | set<text>: uniqueness, ordering | ✅ | Test 11 | DONE |
| 1.29 | set<uuid>: UUID uniqueness | ⚠️ | | MISSING |
| 1.30 | map<text, int>: key-value pairs | ✅ | Test 12 | DONE |
| 1.31 | map<text, text>: complex values | ⚠️ | | MISSING |
| 1.32 | map<int, double>: numeric keys | ✅ | Test 43 | DONE |
| 1.33 | Nested collection without freezing (error) | ⚠️ | | MISSING |
| 1.34 | Collection with NULL elements (error) | ⚠️ | Test 31? | PARTIAL |

**Coverage: 6/9 scenarios** (3 missing dedicated tests)

---

### Blueprint Section 4: INSERT with UDTs (Tests 36-40)

| Blueprint # | Scenario | Implemented | Test Name | Status |
|-------------|----------|-------------|-----------|--------|
| 1.36 | Simple UDT (all fields) | ✅ | Test 13: SimpleUDT | DONE |
| 1.37 | UDT with NULL fields | ⚠️ | | MISSING |
| 1.38 | Nested UDT (frozen required) | ✅ | Test 21: NestedUDT | DONE |
| 1.39 | UDT in collection (frozen<udt> in list) | ✅ | Test 25 | DONE |
| 1.40 | Invalid UDT insert (error test) | ⚠️ | | MISSING |

**Coverage: 3/5 scenarios** (2 missing)

---

### Blueprint Section 5: INSERT with Tuples (Tests 41-45)

| Blueprint # | Scenario | Implemented | Test Name | Status |
|-------------|----------|-------------|-----------|--------|
| 1.41 | Simple tuple (2-3 elements) | ✅ | Test 14: Tuple | DONE |
| 1.42 | Tuple with mixed types | ⚠️ | | MISSING |
| 1.43 | Tuple in collection | ⚠️ | | MISSING |
| 1.44 | Tuple with NULL elements | ⚠️ | | MISSING |
| 1.45 | Tuple update (immutable) | ⚠️ | | MISSING |

**Coverage: 1/5 scenarios** (4 missing - significant gap!)

---

### Blueprint Section 6: INSERT with USING Clause (Tests 46-55)

| Blueprint # | Scenario | Implemented | Test Name | Status |
|-------------|----------|-------------|-----------|--------|
| 1.46 | USING TTL with integer | ✅ | Test 26: UsingTTL | DONE |
| 1.47 | USING TTL with bind marker (?) | ⚠️ | | MISSING |
| 1.48 | USING TTL with named bind marker (:ttl) | ⚠️ | | MISSING |
| 1.49 | USING TIMESTAMP with integer | ✅ | Test 27: UsingTimestamp | DONE |
| 1.50 | USING TIMESTAMP with bind marker | ⚠️ | | MISSING |
| 1.51 | Combined USING TTL AND TIMESTAMP | ✅ | Test 28 | DONE |
| 1.52 | USING with multiple statements in batch | ✅ | Test 70 | DONE |
| 1.53 | TTL=0 (no expiration) | ⚠️ | | MISSING |
| 1.54 | TTL with very large value | ⚠️ | | MISSING |
| 1.55 | Verify TTL actually expires (wait test) | ⚠️ | | MISSING |

**Coverage: 4/10 scenarios** (6 missing - major gap!)

---

### Blueprint Section 7: INSERT JSON (Tests 56-65)

| Blueprint # | Scenario | Implemented | Test Name | Status |
|-------------|----------|-------------|-----------|--------|
| 1.56 | Simple JSON object INSERT | ✅ | Test 29: InsertJSON | DONE |
| 1.57 | JSON with all columns | ⚠️ | Test 29? | PARTIAL |
| 1.58 | JSON with partial columns (NULL) | ⚠️ | | MISSING |
| 1.59 | JSON with nested objects (error) | ⚠️ | | MISSING |
| 1.60 | JSON with array (maps to list) | ⚠️ | | MISSING |
| 1.61 | JSON with NULL values | ⚠️ | | MISSING |
| 1.62 | JSON with escaped quotes | ⚠️ | | MISSING |
| 1.63 | JSON with special chars/emoji | ⚠️ | | MISSING |
| 1.64 | JSON numeric precision | ⚠️ | | MISSING |
| 1.65 | JSON INSERT + SELECT comparison | ⚠️ | Test 48? | PARTIAL |

**Coverage: 1/10 scenarios** (9 missing - CRITICAL gap!)

---

### Blueprint Section 8: INSERT with Bind Markers (Tests 66-75)

| Blueprint # | Scenario | Implemented | Test Name | Status |
|-------------|----------|-------------|-----------|--------|
| 1.66 | Anonymous bind marker (?) | ⚠️ | | MISSING |
| 1.67 | Named bind markers (:col1, :col2) | ⚠️ | | MISSING |
| 1.68 | Mixed anonymous/named (error) | ⚠️ | | MISSING |
| 1.69 | Bind marker for TTL | ⚠️ | | MISSING |
| 1.70 | Bind marker for TIMESTAMP | ⚠️ | | MISSING |
| 1.71 | Multiple rows with prepared statement | ⚠️ | | MISSING |
| 1.72 | NULL values through bind markers | ⚠️ | | MISSING |
| 1.73 | Type mismatch in bind markers (error) | ⚠️ | | MISSING |
| 1.74 | Bind marker validation | ⚠️ | | MISSING |

**Coverage: 0/10 scenarios** (10 missing - ENTIRE SECTION MISSING!)

---

### Blueprint Section 9: INSERT IF NOT EXISTS (Tests 76-80)

| Blueprint # | Scenario | Implemented | Test Name | Status |
|-------------|----------|-------------|-----------|--------|
| 1.76 | IF NOT EXISTS succeeds | ✅ | Test 30: IfNotExists | DONE |
| 1.77 | IF NOT EXISTS fails (duplicate) | ✅ | Test 30: IfNotExists | DONE |
| 1.78 | IF NOT EXISTS returns [applied] | ✅ | Test 30: IfNotExists | DONE |
| 1.79 | IF NOT EXISTS with TTL | ⚠️ | | MISSING |
| 1.80 | Verify error type matches spec | ⚠️ | | MISSING |

**Coverage: 3/5 scenarios** (2 missing)

---

### Blueprint Section 10: Edge Cases and Error Handling (Tests 81-90)

| Blueprint # | Scenario | Implemented | Test Name | Status |
|-------------|----------|-------------|-----------|--------|
| 1.81 | INSERT duplicate keys (last write wins) | ⚠️ | | MISSING |
| 1.82 | INSERT missing partition key (error) | ⚠️ | | MISSING |
| 1.83 | INSERT missing clustering column (error) | ⚠️ | | MISSING |
| 1.84 | INSERT into non-existent table (error) | ⚠️ | | MISSING |
| 1.85 | INSERT too many columns (ignored) | ⚠️ | | MISSING |
| 1.86 | INSERT type mismatch (error) | ⚠️ | | MISSING |
| 1.87 | INSERT very large blob (multi-MB) | ⚠️ | | MISSING |
| 1.88 | INSERT MAX_INT, MIN_INT boundaries | ✅ | Test 03 (partial) | PARTIAL |
| 1.89 | INSERT empty text field | ✅ | Test 33 | DONE |
| 1.90 | INSERT extremely long string | ✅ | Test 33: LargeText | DONE |

**Coverage: 3/10 scenarios** (7 missing - CRITICAL gap!)

---

## Summary by Category

| Category | Blueprint Tests | Implemented | Missing | % Complete |
|----------|----------------|-------------|---------|------------|
| Basic Operations (1-5) | 5 | 5 | 0 | 100% |
| Primitives (6-25) | 20 | 18 | 2 | 90% |
| Collections (26-35) | 10 | 7 | 3 | 70% |
| UDTs (36-40) | 5 | 3 | 2 | 60% |
| Tuples (41-45) | 5 | 1 | 4 | 20% ⚠️ |
| USING Clause (46-55) | 10 | 4 | 6 | 40% ⚠️ |
| INSERT JSON (56-65) | 10 | 1 | 9 | 10% ⚠️ |
| **Bind Markers (66-75)** | **10** | **0** | **10** | **0%** ❌ |
| IF NOT EXISTS (76-80) | 5 | 3 | 2 | 60% |
| **Error Handling (81-90)** | **10** | **3** | **7** | **30%** ⚠️ |
| **TOTAL** | **90** | **45** | **45** | **50%** |

---

## Critical Gaps (High Priority)

### 1. ❌ BIND MARKERS - ENTIRE SECTION MISSING (10 tests)

**Blueprint defined:**
- Test 66: Anonymous bind marker (?)
- Test 67: Named bind markers (:col1, :col2)
- Test 68: Mixed use (should error)
- Test 69: Bind marker for TTL
- Test 70: Bind marker for TIMESTAMP
- Test 71: Multiple rows with prepared statement
- Test 72: NULL values through bind markers
- Test 73: Type mismatch (error)
- Test 74-75: Bind marker validation

**Actually implemented:** NONE

**Impact:**
- No testing of prepared statements
- No testing of bind marker functionality
- No testing of bind marker error scenarios
- This is a MAJOR feature gap

---

### 2. ⚠️ INSERT JSON - 90% MISSING (9/10 tests)

**Blueprint defined 10 specific JSON scenarios:**
- Simple JSON (implemented in Test 29)
- JSON with all columns (missing)
- JSON with partial columns (missing)
- JSON with nested objects/error (missing)
- JSON with arrays (missing)
- JSON with NULL values (missing)
- JSON with escaped quotes (missing)
- JSON with special chars (missing)
- JSON numeric precision (missing)
- JSON round-trip SELECT (partially in Test 48)

**Actually implemented:** 1-2 tests

**Impact:**
- JSON INSERT barely tested
- No error scenario testing for JSON
- No complex JSON structure testing

---

### 3. ⚠️ ERROR HANDLING - 70% MISSING (7/10 tests)

**Blueprint defined error scenarios:**
- Duplicate keys (missing)
- Missing partition key (missing)
- Missing clustering column (missing)
- Non-existent table (missing)
- Too many columns (missing)
- Type mismatch (missing)
- Very large blob (missing)
- Boundary values (partial)
- Empty text (done)
- Long string (done)

**Actually implemented:** 3/10

**Impact:**
- Error scenarios not validated
- Don't know if errors are detected correctly
- Missing negative testing

---

### 4. ⚠️ TUPLES - 80% MISSING (4/5 tests)

**Blueprint defined:**
- Simple tuple (done in Test 14)
- Tuple with mixed types (missing)
- Tuple in collection (missing)
- Tuple with NULL elements (missing)
- Tuple update semantics (missing)

**Actually implemented:** 1/5

**Impact:**
- Tuple functionality barely tested
- Complex tuple scenarios untested

---

### 5. ⚠️ USING CLAUSE - 60% MISSING (6/10 tests)

**Blueprint defined:**
- Basic TTL (done)
- Basic TIMESTAMP (done)
- Combined TTL+TIMESTAMP (done)
- BATCH with USING (done)
- TTL with bind markers (missing)
- TIMESTAMP with bind markers (missing)
- TTL=0 (missing)
- TTL large value (missing)
- TTL expiration verification (missing)

**Actually implemented:** 4/10

---

## What Was Actually Tested (78 implemented tests)

Looking at the test names, the implementation focused on:
- ✅ Core data types (primitives, collections, UDTs)
- ✅ Nested collections (frozen rules)
- ✅ Basic USING clauses
- ✅ Some LWT scenarios
- ✅ BATCH operations
- ✅ Some WHERE clause variants
- ✅ Some SELECT features
- ✅ Special characters and Unicode
- ✅ Some boundary values

**But skipped/merged:**
- ❌ All bind marker tests
- ❌ Most JSON tests
- ❌ Most error scenarios
- ❌ Most tuple tests
- ❌ Several USING clause variants

---

## Test 78: "TestDML_Insert_Remaining"

**What it claims to test:** "Tests 78-90: All final INSERT tests"

**What it actually does:**
```go
// Just runs 13 iterations of basic INSERT/DELETE with:
CREATE TABLE final_tests (id int PRIMARY KEY, data text)

for id in [78000, 79000, 80000, ..., 90000] {
    INSERT INTO final_tests (id, data) VALUES (id, "testXX")
    SELECT to verify
    DELETE
}
```

**What it's NOT testing:**
- No bind markers (blueprint tests 66-75)
- No error scenarios (blueprint tests 81-90)
- No TTL variants (blueprint tests 53-55)
- No JSON variants (blueprint tests 58-65)
- No tuple variants (blueprint tests 42-45)

**This is a placeholder that does NOT fulfill the blueprint requirements.**

---

## CRITICAL FINDINGS

### 1. Actual Coverage is ~50%, NOT 100%

**Blueprint:** 90 specific test scenarios
**Implemented:** ~45 scenarios actually tested
**Placeholder:** Test 78 loops 13 times with same simple INSERT

### 2. Major Feature Gaps

| Feature | Blueprint Tests | Actually Tested | Gap |
|---------|----------------|-----------------|-----|
| **Bind Markers** | 10 tests | 0 tests | 100% missing |
| **INSERT JSON** | 10 tests | 1-2 tests | 80-90% missing |
| **Error Scenarios** | 10 tests | 3 tests | 70% missing |
| **Tuples** | 5 tests | 1 test | 80% missing |
| **USING Variants** | 10 tests | 4 tests | 60% missing |

### 3. What This Means

**User's concern is VALID:**
> "I suspect there are a LOT of gaps now, which is an issue"

✅ **Confirmed: 45/90 scenarios missing (50% gap)**

**The goal:**
> "Complete coverage of ALL of the CQL language... there is nothing we skip, leave for later or decide not to bother testing"

❌ **Not achieved: Major features (bind markers, JSON, errors) barely tested**

---

## Prioritized Gap List

### CRITICAL Priority (Must Implement)

**1. Bind Markers (Tests 66-75) - 10 missing tests**
- Anonymous bind markers (?)
- Named bind markers (:name)
- Mixed use error testing
- Bind markers in TTL/TIMESTAMP
- Prepared statement reuse
- NULL binding
- Type mismatch errors
- Bind marker validation

**2. Error Scenarios (Tests 81-90) - 7 missing tests**
- Missing partition key (should error)
- Missing clustering column (should error)
- Non-existent table (should error)
- Type mismatch (should error)
- Too many columns (behavior validation)
- Duplicate keys (last write wins validation)
- Very large blob (boundary test)

**3. INSERT JSON (Tests 57-65) - 8 missing tests**
- JSON with partial columns
- JSON with arrays → list mapping
- JSON with NULL values
- JSON with escaped quotes
- JSON with special characters
- JSON numeric precision handling
- JSON nested objects error testing
- JSON round-trip validation

### HIGH Priority (Should Implement)

**4. Tuples (Tests 42-45) - 4 missing tests**
- Tuple with mixed types
- Tuple in collection (list<tuple<>>)
- Tuple with NULL elements
- Tuple immutability testing

**5. USING Clause Variants (Tests 47-48, 53-55) - 6 missing tests**
- TTL with bind markers
- TIMESTAMP with bind markers
- TTL=0 (no expiration)
- TTL with very large value
- TTL expiration verification (wait test)

### MEDIUM Priority (Nice to Have)

**6. Collection Variants (Tests 29, 33-34) - 3 missing tests**
- set<uuid> uniqueness
- map<text,text> complex values
- Nested collection without freezing (error test)

**7. UDT Variants (Tests 37, 40) - 2 missing tests**
- UDT with NULL fields
- Invalid UDT insert (error test)

**8. Primitive Types (Tests for ascii, varchar) - 2 missing tests**
- Dedicated ascii test
- Dedicated varchar test

---

## Recommended Action Plan

### Phase 1: Fill Critical Gaps (35 tests)
1. Implement bind marker tests (10 tests) - **CRITICAL**
2. Implement error scenario tests (7 tests) - **CRITICAL**
3. Implement INSERT JSON tests (8 tests) - **CRITICAL**
4. Implement tuple tests (4 tests) - **HIGH**
5. Implement USING clause variants (6 tests) - **HIGH**

### Phase 2: Complete Coverage (10 tests)
6. Implement collection variants (3 tests)
7. Implement UDT variants (2 tests)
8. Implement primitive type variants (2 tests)
9. Implement remaining edge cases (3 tests)

### Total Additional Work Required
- **45 missing tests** to reach 100% blueprint coverage
- **35 tests** are HIGH/CRITICAL priority
- **10 tests** are MEDIUM priority
- Estimated time: 15-20 hours (3-4 tests/hour)

---

## Impact Assessment

**Current State:**
- ✅ Core functionality tested (primitives, collections, basic operations)
- ✅ Nested collections tested (frozen rules validated)
- ✅ Some advanced features tested (LWT, BATCH, some USING clauses)

**Missing:**
- ❌ Prepared statements and bind markers (critical feature)
- ❌ JSON INSERT variants (important feature)
- ❌ Error handling and validation (important for robustness)
- ❌ Tuple edge cases (moderate feature)
- ❌ USING clause variants (moderate feature)

**Risk:**
- Bind markers are widely used in production → not testing them is HIGH RISK
- Error scenarios not tested → don't know if errors are handled correctly
- JSON is a common use case → minimal testing is MEDIUM RISK

---

## Conclusion

**User's assessment is CORRECT:**
> "I suspect there are a LOT of gaps now"

**Gaps confirmed:** 45/90 tests missing (50%)

**Recommendation:** Implement the 35 CRITICAL/HIGH priority missing tests before proceeding to UPDATE suite.

**Next step:** Shall I create detailed test implementations for the missing tests, starting with bind markers (10 tests)?

