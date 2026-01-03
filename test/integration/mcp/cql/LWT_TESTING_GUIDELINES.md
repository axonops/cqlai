# LWT (Lightweight Transaction) Testing Guidelines

**CRITICAL:** When testing LWT operations, maintain consistency in the Paxos clock domain.

---

## The Problem

**Mixing LWT and non-LWT operations on the same data causes timing issues.**

### Why This Happens

1. **LWT operations** (IF NOT EXISTS, IF EXISTS, IF conditions) use **Paxos consensus** with a **hybrid-logical clock**
2. **Non-LWT operations** (regular INSERT, UPDATE, DELETE) use **regular timestamps**
3. **These clocks are NOT interchangeable**
4. Mixing them causes timing/visibility issues where operations succeed but don't take effect immediately

### Example of the Problem

```go
// ❌ WRONG: Mixing LWT and non-LWT
INSERT INTO table (...) VALUES (...) IF NOT EXISTS;  // LWT (Paxos clock)
DELETE FROM table WHERE id = ?;                      // Non-LWT (regular clock)
// Result: DELETE succeeds but row may still exist!
```

---

## The Solution

**Use LWT consistently throughout the test - don't mix with non-LWT.**

### Correct Pattern for LWT Tests

```go
// ✅ CORRECT: All LWT operations
INSERT INTO table (...) VALUES (...) IF NOT EXISTS;  // LWT
DELETE FROM table WHERE id = ? IF EXISTS;            // LWT
// Result: DELETE works immediately, no delays needed
```

---

## Testing Guidelines

### Rule 1: LWT Tests Use LWT for ALL Operations

**When testing any LWT feature:**
- INSERT IF NOT EXISTS → Use DELETE IF EXISTS (not regular DELETE)
- UPDATE IF EXISTS → Use DELETE IF EXISTS (not regular DELETE)
- UPDATE IF condition → Use DELETE IF EXISTS (not regular DELETE)

**Example:**
```go
// Test: INSERT IF NOT EXISTS
insertArgs := map[string]any{
    "operation": "INSERT",
    "values": map[string]any{...},
    "if_not_exists": true,  // LWT
}

// Cleanup: Use LWT for DELETE too
deleteArgs := map[string]any{
    "operation": "DELETE",
    "if_exists": true,  // CRITICAL: Keep in LWT domain
    "where": []map[string]any{...},
}
```

### Rule 2: Non-LWT Tests Use Non-LWT for ALL Operations

**When testing regular operations (no IF clauses):**
- Regular INSERT → Use regular DELETE (no IF EXISTS)
- Regular UPDATE → Use regular DELETE (no IF EXISTS)

**Example:**
```go
// Test: Regular INSERT
insertArgs := map[string]any{
    "operation": "INSERT",
    "values": map[string]any{...},
    // No if_not_exists
}

// Cleanup: Regular DELETE
deleteArgs := map[string]any{
    "operation": "DELETE",
    // No if_exists
    "where": []map[string]any{...},
}
```

### Rule 3: NEVER Mix LWT and Non-LWT on Same Data

**❌ DO NOT:**
```go
// Bad: LWT INSERT, regular DELETE
INSERT ... IF NOT EXISTS;
DELETE ... WHERE id = ?;  // Will have timing issues

// Bad: Regular INSERT, LWT DELETE
INSERT ... VALUES (...);
DELETE ... IF EXISTS;  // Unnecessary and potentially confusing
```

**✅ DO:**
```go
// Good: All LWT
INSERT ... IF NOT EXISTS;
DELETE ... IF EXISTS;

// Good: All non-LWT
INSERT ... VALUES (...);
DELETE ... WHERE id = ?;
```

### Rule 4: NO Delays as Workarounds

**❌ DO NOT add delays:**
```go
// Bad workaround
INSERT ... IF NOT EXISTS;
time.Sleep(5 * time.Second);  // NO!
DELETE ... WHERE id = ?;
```

**✅ DO use consistent LWT:**
```go
// Proper solution
INSERT ... IF NOT EXISTS;
DELETE ... IF EXISTS;  // No delay needed
```

---

## Tests Following These Guidelines

**LWT Tests (use IF EXISTS for DELETE):**
- Test 30: INSERT IF NOT EXISTS → DELETE IF EXISTS ✅
- Test 63: UPDATE IF EXISTS → DELETE IF EXISTS ✅
- Test 64: UPDATE IF condition → DELETE IF EXISTS ✅
- Test 65: DELETE IF EXISTS (tests the DELETE itself) ✅
- Test 71: BATCH with LWT → (validates data only) ✅

**Non-LWT Tests (regular DELETE):**
- Tests 1-29, 31-62, 66-70, 72-90: Regular operations → Regular DELETE ✅

---

## Technical Explanation

### Paxos Hybrid-Logical Clock

LWT operations in Cassandra use Paxos consensus which maintains a separate hybrid-logical clock to ensure linearizability. This clock advances independently from regular write timestamps.

**Implications:**
- LWT writes record their commit in the Paxos clock
- Regular reads/writes use the regular timestamp clock
- Cross-clock operations have undefined visibility guarantees
- This is by design for correctness, not a bug

**Result:**
When you mix LWT and non-LWT:
- Operations may succeed but not be immediately visible
- Requires waiting for clock synchronization
- Unsafe and unpredictable timing

**Solution:**
Stay within one clock domain - use LWT consistently throughout.

---

## Performance Impact

### With Delays (Old Approach)
```
Test 30: 6.78s (5s delay)
Test 63: 14.8s (two 5s delays)
```

### Without Delays (Correct LWT)
```
Test 30: 1.87s (no delay)
Test 63: 9.5s (no delay - time is test overhead)
```

**Time saved: 4-5 seconds per LWT test**

---

## References

- Cassandra LWT documentation: Uses Paxos consensus
- DELETE succeeds whether data exists or not (returns success in both cases)
- Mixing LWT and non-LWT is generally unsafe
- Can be done if timing caveats are understood, but not recommended

---

**Summary:**
- LWT tests: Use IF clauses consistently (IF NOT EXISTS, IF EXISTS, IF conditions)
- Non-LWT tests: Use regular operations consistently
- NEVER mix them on the same data
- NO delays as workarounds
