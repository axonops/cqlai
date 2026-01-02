# MCP Integration Tests

## Running Tests

### Individual Tests (Recommended)
Each test can be run individually and will pass:
```bash
go test ./test/integration/mcp -tags=integration -run "TestMCP_PrimitiveTypes_Text$" -v
go test ./test/integration/mcp -tags=integration -run "TestMCP_PrimitiveTypes_Integers$" -v
# etc.
```

### All Tests Sequentially
To run all tests, use `-p 1` to force sequential execution:
```bash
go test ./test/integration/mcp -tags=integration -p 1 -v
```

### Specific Test Groups
```bash
# All primitive type tests
go test ./test/integration/mcp -tags=integration -run "TestMCP_PrimitiveTypes" -p 1 -v

# All data type tests
go test ./test/integration/mcp -tags=integration -run "TestMCP.*Types" -p 1 -v

# All feature tests
go test ./test/integration/mcp -tags=integration -run "TestMCP_(Counter|List|Set|Map|Batch)" -p 1 -v
```

## Test Isolation Status

### Current Limitation
Tests must run sequentially (`-p 1`) due to shared HTTP port configuration. When tests run concurrently, they interfere with each other because:
- All tests use the same HTTP port (8911) from `testdata/readwrite.json`
- HTTP server cleanup between tests takes time
- Port binding conflicts occur

### What's Been Done
- ✅ Each test creates its own Cassandra session
- ✅ Each test creates its own MCP handler instance (via `router.NewMCPHandler()`)
- ✅ Tests are isolated at the session/handler level

### What's Needed for Full Concurrent Support
- [ ] Dynamic port allocation per test (random available ports)
- [ ] Proper HTTP server shutdown synchronization
- [ ] Test-specific config files with unique ports
- [ ] Or: Refactor to use in-process testing without HTTP

## Prerequisites

### Cassandra
Tests require a running Cassandra instance:
```bash
podman start cassandra-test
# or
docker start cassandra-test
```

**Expected:**
- Host: 127.0.0.1:9042
- Username: cassandra
- Password: cassandra
- Version: 5.0+ (for vector type support)

### Test Data
Tests use these keyspaces:
- `type_test` - For data type tests
- `cqlai_test` - For feature tests
- `test_mcp` - For general MCP tests

Tables are created automatically with `IF NOT EXISTS`.

## Test Structure

### Data Type Tests
- `TestMCP_PrimitiveTypes_*` - Individual primitive types
- `TestMCP_ComplexTypes_*` - Complex types (tuple, UDT)
- `TestMCP_DataTypes_*` - Collection types (list, set, map)

All 28 Cassandra data types are tested:
- ✅ Text types (text, ascii, varchar)
- ✅ Integer types (tinyint, smallint, int, bigint, varint)
- ✅ Float types (float, double, decimal)
- ✅ Boolean
- ✅ Blob
- ✅ Date/time types (date, time, timestamp, duration)
- ✅ UUID types (uuid, timeuuid)
- ✅ Network (inet)
- ✅ Counter
- ✅ Vector (Cassandra 5.0+)
- ✅ Collections (list, set, map)
- ✅ Complex (tuple, UDT, frozen)

### Feature Tests
- Counter operations (increment, decrement)
- Collection operations (append, prepend, add, remove, merge)
- LWT tests (IF NOT EXISTS, IF EXISTS, IF conditions)
- BATCH operations (LOGGED, UNLOGGED, COUNTER, with LWT)
- DDL operations with IF clauses
- Advanced query features (TOKEN, tuples, aggregates)

## Troubleshooting

### "session has been closed" errors
**Cause:** Tests running concurrently
**Fix:** Run with `-p 1` flag

### "bind: address already in use"
**Cause:** Previous test HTTP server still shutting down
**Fix:** Wait a few seconds and retry, or run with `-p 1`

### "connection refused" on port 9042
**Cause:** Cassandra not running
**Fix:**
```bash
podman start cassandra-test
podman logs cassandra-test  # wait for "Starting listening for CQL clients"
```

### Test failures after code changes
1. Verify unit tests pass first: `go test ./internal/ai -count=1`
2. Run individual integration test: `go test ./test/integration/mcp -tags=integration -run "TestMCP_PrimitiveTypes_Text$" -v`
3. If individual test passes, it's likely a concurrency issue - use `-p 1`

## Verification

### Verify All Data Types Work
```bash
#!/bin/bash
# Run all data type tests individually

for test in Text Integers Floats Boolean Blob DateTime UUID Inet Duration Vector; do
    echo "Testing: $test"
    go test ./test/integration/mcp -tags=integration -run "TestMCP_PrimitiveTypes_${test}$" -v || exit 1
done

for test in UDT Tuple; do
    echo "Testing: $test"
    go test ./test/integration/mcp -tags=integration -run "TestMCP_ComplexTypes_${test}$" -v || exit 1
done

echo "✅ All data type tests passed!"
```

## Future Improvements

1. **Dynamic Port Allocation**
   - Generate random available port per test
   - Update config on-the-fly
   - Eliminates port conflicts

2. **Test Containers**
   - Use testcontainers-go for embedded Cassandra
   - Each test gets fresh Cassandra instance
   - True isolation, no shared state

3. **In-Process Testing**
   - Test MCP server without HTTP layer
   - Direct function calls instead of HTTP requests
   - Faster, no port conflicts

4. **Parallel Test Support**
   - Implement one of above solutions
   - Update docs to remove `-p 1` requirement
