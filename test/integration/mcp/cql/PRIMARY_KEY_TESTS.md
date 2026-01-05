# Primary Key Validation Tests - Implementation Tracker

**Date:** 2026-01-05
**Blueprint:** claude-notes/CQL_TEST_ADDENDUM_PRIMARY_KEYS_AND_BATCH.md
**Total PK Tests:** 15 (from blueprint)

---

## Status Summary

**Already implemented as error tests:** 7
**Need to implement (SUCCESS tests):** 8
**Total:** 15

---

## Tests Already Implemented

### INSERT Errors (2/3)
- ✅ INSERT_PK_02: Missing partition key → ERR_02 (dml_insert_error_test.go)
- ✅ INSERT_PK_03: Missing clustering key → ERR_03 (dml_insert_error_test.go)

### UPDATE Errors (2/5)
- ✅ UPDATE_PK_02: Partial PK + regular column → UPDATE_ERR_01 (dml_update_error_test.go)
- ✅ UPDATE_PK_04: Missing partition key → UPDATE_ERR_02 (dml_update_error_test.go)

### DELETE Errors (2/7)
- ✅ DELETE_PK_05: Missing partition key → DELETE_ERR_01 (dml_delete_error_test.go)
- ✅ DELETE_PK_06: No WHERE clause → DELETE_ERR_02 (dml_delete_error_test.go)

### Additional
- ✅ UPDATE_ERR_03: No WHERE clause (bonus)

---

## Tests To Implement (8 SUCCESS tests)

### INSERT Tests (1)
- ✅ INSERT_PK_01: Full PK (valid) - DONE as Test 79

### UPDATE Tests (2)
- ✅ UPDATE_PK_01: Full PK (valid) - DONE as UPDATE Test 01
- ❌ UPDATE_PK_03: Partial PK + static column (valid) - TODO

### DELETE Tests (5)
- ✅ DELETE_PK_01: Full PK - row delete - DONE as DELETE Test 01
- ✅ DELETE_PK_02: Partition key only - partition delete - DONE as DELETE Test 02
- ✅ DELETE_PK_03: Partition key + range - DONE as DELETE Test 03
- ✅ DELETE_PK_04: Partition key + IN - DONE as DELETE Test 04
- ✅ DELETE_PK_07: Static column with partial PK - DONE as DELETE Test 05

---

## Implementation Plan

**File locations:**
- INSERT_PK_01 → Create new file: dml_insert_pk_test.go
- UPDATE_PK_01, 03 → Create new file: dml_update_pk_test.go
- DELETE_PK_01-04, 07 → Create new file: dml_delete_pk_test.go

**Approach:**
1. Implement ONE test
2. RUN it immediately
3. Verify it PASSES
4. Document result here
5. Move to next test
6. Commit every 2-3 tests

---

## Implementation Log

**None yet - starting now**

---

**This file will be updated as each test is implemented.**
