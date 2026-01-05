# Outstanding Work - TODO Items

**Date:** 2026-01-05 (Session 2 update)
**Branch:** feature/mcp_datatypes

---

## Session 2 Summary

### Tests Added: 87-110 (24 tests)
- All tests passing ✅
- No skipped tests ✅
- Bugs fixed immediately ✅

### Bugs Fixed This Session
1. ✅ Bug 6: frozen<tuple> not handled in frozen type switch (Test 87)
2. ✅ Bug 7: isFunctionCall matched strings with parentheses (Test 90)
3. ✅ MCP error reporting now includes generated CQL

---

## Current Test Status

**File: dml_insert_test.go**
- 110 INSERT tests (Tests 1-110)
- 5 BATCH duplicates (Tests 36, 68-71) - still need deletion

**Error Test Files:**
- dml_insert_error_test.go: 14 error tests (100% ✅)
- dml_update_error_test.go: 3 error tests (100% ✅)
- dml_delete_error_test.go: 2 error tests (100% ✅)
- dml_batch_error_test.go: 2 error tests (100% ✅)

**Other DML Files:**
- dml_update_test.go: 2 tests
- dml_delete_test.go: 5 tests
- dml_batch_test.go: 10 tests

**Total: 149 tests**

---

## Outstanding Work

### 1. Duplicate BATCH Tests (Still Need Cleanup)

**Status:** Tests 36, 68-71 exist in BOTH files

**Action needed:**
- Delete tests 36, 68, 69, 70, 71 from dml_insert_test.go
- Renumber subsequent tests
- Verify all tests still pass

### 2. Remaining INSERT Tests (31 tests to reach 141)

**Categories needing more tests:**
- USING clause edge cases (3 more)
- More error scenarios (several)
- Bind markers (10 tests - low priority per user)
- Additional edge cases

### 3. DEFAULT UNSET Support

**Status:** Not yet implemented
- INSERT JSON DEFAULT UNSET clause not supported
- Currently only DEFAULT NULL works

---

## Next Session Priorities

1. **Continue INSERT tests** to reach 141 (22% remaining)
2. **Expand UPDATE tests** from 2/100 (98% remaining)
3. **Expand DELETE tests** from 5/60 (92% remaining)
4. **Expand BATCH tests** from 10/22 (54% remaining)
5. **Clean up BATCH duplicates** when convenient

---

**This file tracks work that was started but not completed.**
**Update this file when work is completed or priorities change.**
