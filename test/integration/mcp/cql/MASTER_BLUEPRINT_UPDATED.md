# Master Blueprint Consolidated - All Test Scenarios

**Date:** 2026-01-04
**Action:** Consolidated all test scenarios into master blueprint
**File:** `claude-notes/cql-complete-test-suite.md` (Note: in .gitignore, changes not committed to git)

---

## What Was Updated in Master Blueprint

### 1. INSERT Test Count Updated
**Was:** 90 tests
**Now:** 141 tests (+51 new scenarios)

### 2. Three New Sections Added

**Section 1: Primary Key Validation Tests (Test 1.91-1.108)**
- 18 tests covering INSERT/UPDATE/DELETE with full vs partial primary keys
- Validates Cassandra's fundamental PK requirements
- Tests static column special semantics

**Section 2: BATCH Advanced Validation (Test 1.109-1.130)**
- 22 tests covering BATCH statement validation
- Counter/non-counter mixing errors
- Cross-partition detection and warnings
- Mixed DML operations
- USING clause combinations
- Size limits and atomicity

**Section 3: Additional Error Validation (Test 1.131-1.141)**
- 10 tests for enhanced error coverage
- Missing PK component errors
- Invalid WHERE clause errors

### 3. Total Suite Count Updated
**Was:** 1,200+ tests
**Now:** 1,251+ tests

---

## All 96 Missing INSERT Scenarios Documented

**In master blueprint `cql-complete-test-suite.md`:**
- Tests 1-90: Original scenarios (45 implemented, 45 missing)
- Tests 91-141: New scenarios (0 implemented, 51 missing)

**Total:** 141 scenarios needed, 45 implemented = **96 missing**

**Breakdown:**
- Bind markers: 10 (Tests 66-75)
- INSERT JSON: 8 (Tests 57-65)
- Error scenarios: 17 (Tests 81-90 + new tests 131-141)
- Primary key validation: 18 (Tests 91-108)
- BATCH validation: 22 (Tests 109-130)
- Tuples: 4 (Tests 42-45)
- USING variants: 6 (Tests 47-48, 53-55)
- Collections, UDTs, primitives: 7 (Tests 29, 33-34, 37, 40)
- Edge cases: 3 (distributed)

---

## Why This Matters

**Before:** Test scenarios scattered across multiple documents
**After:** ALL scenarios in ONE master blueprint

**For future sessions:**
1. Read `claude-notes/cql-complete-test-suite.md` (master)
2. Check `test/integration/mcp/cql/INSERT_GAP_ANALYSIS.md` (current status)
3. All 141 scenarios documented - nothing missed

---

## References

**Master blueprint (all scenarios):**
- `claude-notes/cql-complete-test-suite.md`

**Current status (what's implemented):**
- `test/integration/mcp/cql/INSERT_GAP_ANALYSIS.md`

**Detailed new scenario descriptions:**
- `claude-notes/CQL_TEST_ADDENDUM_PRIMARY_KEYS_AND_BATCH.md`

**Note:** claude-notes directory is in .gitignore - changes to master blueprint exist locally but don't commit to git. This file serves as the commit record.
