# Outstanding Work - TODO Items

**Date:** 2026-01-05
**Branch:** feature/mcp_datatypes

---

## Duplicate Tests (Need Cleanup)

### BATCH Tests Duplicated

**Status:** Tests exist in BOTH dml_insert_test.go AND dml_batch_test.go

**Duplicates:**
- Test 36 (BatchMultipleInserts) → dml_batch_test.go::Batch_01
- Test 68 (BatchUnlogged) → dml_batch_test.go::Batch_02
- Test 69 (BatchCounter) → dml_batch_test.go::Batch_03
- Test 70 (BatchWithTimestamp) → dml_batch_test.go::Batch_04
- Test 71 (BatchWithLWT) → dml_batch_test.go::Batch_05

**Action needed:**
- Delete tests 36, 68, 69, 70, 71 from dml_insert_test.go
- Add renumbering comments where they were removed
- Verify all tests still pass
- Commit cleanup

**Comments added:** ✅ All 5 tests now have "NOTE: DUPLICATED" comments

---

## Test File Organization

**Current state:**
- dml_insert_test.go: 78 tests (including 5 BATCH duplicates)
- dml_insert_error_test.go: 6 error tests
- dml_update_error_test.go: 3 error tests
- dml_delete_error_test.go: 2 error tests
- dml_batch_test.go: 6 BATCH tests
- dml_batch_error_test.go: 2 BATCH error tests

**After cleanup:**
- dml_insert_test.go: 73 tests (remove 5 BATCH duplicates)
- Error files: 13 error tests
- dml_batch_test.go: 6 BATCH tests

**Total after cleanup:** 92 tests (73 + 13 + 6)

---

## Priority After Cleanup

1. **Remove duplicate BATCH tests** (high priority - affects test count)
2. **Primary key validation tests** (8 more success scenarios)
3. **More BATCH validation tests** (cross-partition detection, etc.)
4. **Bind marker tests** (10 tests)
5. **INSERT JSON tests** (8 tests)

---

**This file tracks work that was started but not completed.**
**Update this file when work is completed or priorities change.**
