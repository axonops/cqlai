# Deterministic CQL Rendering - Critical Implementation

**Date:** 2026-01-04
**File:** internal/ai/planner.go
**Purpose:** Make CQL generation deterministic for exact testing and consistent behavior
**Status:** ✅ IMPLEMENTED AND VERIFIED

---

## The Problem

Go maps iterate in **random order**. When rendering CQL from maps, we got non-deterministic output:

**Example - INSERT columns (random order):**
```sql
-- Run 1:
INSERT INTO table (tiny_val, var_val, big_val, id) VALUES (...)

-- Run 2:
INSERT INTO table (big_val, id, tiny_val, var_val) VALUES (...)
```

**Example - Map literals (random key order):**
```sql
-- Run 1:
{'group1': [1,2], 'group2': [3,4]}

-- Run 2:
{'group2': [3,4], 'group1': [1,2]}
```

**Example - UDT fields (random field order):**
```sql
-- Run 1:
{street: 'Main', city: 'NYC', zip: '10001'}

-- Run 2:
{city: 'NYC', zip: '10001', street: 'Main'}
```

### Impact

- ❌ CQL assertions impossible (expected CQL never matches actual)
- ❌ Audit logs inconsistent
- ❌ Query caching broken (same logical query has different string representation)
- ❌ Debugging difficult (same query looks different each time)

---

## The Solution

**Sort EVERYTHING alphabetically before rendering CQL.**

This ensures:
- ✅ Same logical query always produces identical CQL string
- ✅ CQL assertions work (can compare exact strings)
- ✅ Logs and debugging are consistent
- ✅ Query caching can work (same string = same query)

---

## Changes Made

### 1. INSERT Column Names (renderInsert - Line 237-255)

**Before:**
```go
columns := make([]string, 0, len(plan.Values))
values := make([]string, 0, len(plan.Values))

for col, val := range plan.Values {  // ❌ Random order
    columns = append(columns, col)
    values = append(values, formatValue(val))
}
```

**After:**
```go
// SORT columns alphabetically for deterministic output
columns := make([]string, 0, len(plan.Values))
for col := range plan.Values {
    columns = append(columns, col)
}
sort.Strings(columns)  // ✅ Alphabetical order

// Build values in same sorted order as columns
values := make([]string, 0, len(columns))
for _, col := range columns {
    val := plan.Values[col]
    values = append(values, formatValue(val))
}
```

**Result:**
```sql
-- Always produces:
INSERT INTO table (big_val, id, int_val, small_val, tiny_val, var_val) VALUES (...)
```

---

### 2. UPDATE SET Clause Columns (renderUpdate - Lines 314-331, 337-344, 395-410)

**Sorted 3 types of columns:**

**A. Counter operations:**
```go
counterCols := make([]string, 0, len(plan.CounterOps))
for col := range plan.CounterOps {
    counterCols = append(counterCols, col)
}
sort.Strings(counterCols)  // ✅ Sorted

for _, col := range counterCols {
    // Render counter operation
}
```

**B. Collection operations:**
```go
collCols := make([]string, 0, len(plan.CollectionOps))
for col := range plan.CollectionOps {
    collCols = append(collCols, col)
}
sort.Strings(collCols)  // ✅ Sorted

for _, col := range collCols {
    // Render collection operation
}
```

**C. Regular value updates:**
```go
valueCols := make([]string, 0, len(plan.Values))
for col := range plan.Values {
    valueCols = append(valueCols, col)
}
sort.Strings(valueCols)  // ✅ Sorted

for _, col := range valueCols {
    // Render value update
}
```

**Result:**
```sql
-- Always produces (alphabetical SET clauses):
UPDATE table SET age = 31, email = 'new@email', is_active = false WHERE ...
```

---

### 3. Map Literal Keys (formatMapWithContext - Lines 2553-2579)

**Before:**
```go
pairs := make([]string, 0, len(m))
for key, value := range m {  // ❌ Random order
    formattedKey := formatValue(key)
    formattedValue := formatValue(value)
    pairs = append(pairs, fmt.Sprintf("%s: %s", formattedKey, formattedValue))
}
```

**After:**
```go
// SORT map keys alphabetically for deterministic output
keys := make([]string, 0, len(m))
for key := range m {
    keys = append(keys, fmt.Sprintf("%v", key))
}
sort.Strings(keys)  // ✅ Sorted

pairs := make([]string, 0, len(keys))
for _, keyStr := range keys {
    // Find value for this key
    value := m[keyStr]
    formattedKey := formatValue(keyStr)
    formattedValue := formatValue(value)
    pairs = append(pairs, fmt.Sprintf("%s: %s", formattedKey, formattedValue))
}
```

**Result:**
```sql
-- Always produces:
{'group1': [1,2,3], 'group2': [4,5,6]}  (alphabetical keys)
```

---

### 4. UDT Field Names (formatUDTWithContext - Lines 2601-2633)

**Before:**
```go
pairs := make([]string, 0, len(m))
for field, value := range m {  // ❌ Random order
    formattedValue := formatValue(value)
    pairs = append(pairs, fmt.Sprintf("%s: %s", field, formattedValue))
}
```

**After:**
```go
// SORT UDT fields alphabetically for deterministic output
fields := make([]string, 0, len(m))
for field := range m {
    fields = append(fields, field)
}
sort.Strings(fields)  // ✅ Sorted

pairs := make([]string, 0, len(fields))
for _, field := range fields {
    value := m[field]
    formattedValue := formatValue(value)
    pairs = append(pairs, fmt.Sprintf("%s: %s", field, formattedValue))
}
```

**Result:**
```sql
-- Always produces:
{city: 'NYC', street: '123 Main St', zip: '10001'}  (alphabetical fields)
```

---

### 5. Set Elements (formatSetWithContext - Lines 2527-2536)

**Before:**
```go
formatted := make([]string, len(unique))
for i, elem := range unique {
    formatted[i] = formatValue(elem)
}
return fmt.Sprintf("{%s}", strings.Join(formatted, ", "))  // ❌ Random order
```

**After:**
```go
formatted := make([]string, len(unique))
for i, elem := range unique {
    formatted[i] = formatValue(elem)
}

// SORT set elements alphabetically for deterministic output
sort.Strings(formatted)  // ✅ Sorted

return fmt.Sprintf("{%s}", strings.Join(formatted, ", "))
```

**Result:**
```sql
-- Always produces:
{'admin', 'premium', 'verified'}  (alphabetical elements)
```

---

## Import Added

```go
import (
    "encoding/json"
    "fmt"
    "sort"  // ← Added for sorting
    "strings"
)
```

---

## Verification

**Tested with 20 INSERT tests containing:**
- Multi-column INSERTs (sorted column order)
- Map literals (sorted keys)
- Set literals (sorted elements)
- UDT literals (sorted fields)
- Nested collections (all levels sorted)

**Result:** ✅ 20/20 tests PASS with exact CQL assertions

---

## Benefits

### 1. Exact Testing
- Can assert exact CQL string match
- Catches rendering bugs immediately
- Validates correct quoting, spacing, syntax

### 2. Consistent Behavior
- Same query always produces same CQL string
- Logs are consistent across runs
- Easier debugging

### 3. Query Caching Possible
- Same logical query = same string representation
- Can cache by CQL string
- Performance optimization enabled

### 4. Better Audit Trail
- CQL in logs is consistent
- Can grep/search for specific queries reliably
- Audit compliance improved

---

## Files Modified

1. ✅ **internal/ai/planner.go**
   - Added `sort` import
   - Sorted INSERT columns (renderInsert)
   - Sorted UPDATE SET clause columns (renderUpdate - 3 places)
   - Sorted map keys (formatMapWithContext)
   - Sorted UDT fields (formatUDTWithContext)
   - Sorted set elements (formatSetWithContext)

2. ✅ **test/integration/mcp/cql/dml_insert_test.go**
   - Updated Test 3 expected CQL (bigint value + column order)
   - Updated Test 13 expected CQL (UDT field order)
   - All 20 tests now pass with exact CQL assertions

---

## Critical Notes

### This is NOT Just for Testing

**This change affects PRODUCTION CQL generation:**
- The CQL sent to Cassandra is now deterministic
- Not just test assertions - actual executed queries
- Same logical operation always produces identical CQL statement

### Cassandra Doesn't Care About Order

**Cassandra treats these as equivalent:**
```sql
INSERT INTO t (a, b, c) VALUES (1, 2, 3)
INSERT INTO t (c, b, a) VALUES (3, 2, 1)
```

**But we care because:**
- Testing requires exact match
- Logging requires consistency
- Caching requires string equality
- Debugging requires predictability

---

**Status: CQL generation is now fully deterministic - ready for comprehensive testing!**
