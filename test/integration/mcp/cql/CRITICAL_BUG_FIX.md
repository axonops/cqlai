# CRITICAL BUG FIX: Sequential Test Execution

**Date:** 2026-01-04
**Discovered By:** Testing UPDATE suite, found INSERT suite had same issue
**Severity:** HIGH - Prevented running test suite sequentially
**Status:** ✅ FIXED

---

## The Problem

**Tests could NOT run sequentially in same `go test` invocation:**
```bash
go test -run "TestDML_Insert_0[1-5]" -v -p 1
# Test 1: ✅ PASS
# Test 2: ❌ FAIL (401 invalid API key)
# Test 3: ❌ FAIL (401 invalid API key)
# Test 4: ❌ FAIL (401 invalid API key)
# Test 5: ❌ FAIL (401 invalid API key)
```

**But each test passed when run individually:**
```bash
go test -run "^TestDML_Insert_01_$" -v  # ✅ PASS
go test -run "^TestDML_Insert_02_$" -v  # ✅ PASS
go test -run "^TestDML_Insert_03_$" -v  # ✅ PASS
```

---

## Root Cause

**Every test was starting and stopping the MCP HTTP server:**

`setupCQLTest()` (called by EVERY test):
1. Generated a NEW API key
2. Called `.mcp start --config-file ../testdata/dba.json`
3. Tried to start HTTP server on port 8912

`teardownCQLTest()` (called by EVERY test):
1. Called `.mcp stop`
2. Waited 300ms
3. Closed session

**The Issue:**
- Test 1: Start server on port 8912, API key A
- Test 1 teardown: Stop server
- Test 2: Try to start server on port 8912, API key B
- **Problem:** Server not fully stopped, OR new server starts but HTTP init uses old API key

**Result:** Test 2's HTTP initialization failed, fell back to generating a session ID, then got "401 invalid API key" on actual requests.

---

## The Fix

**Start MCP server ONCE for entire test suite using `TestMain()`:**

### Added TestMain

```go
var (
	sharedMCPHandler *router.MCPHandler
	sharedAPIKey     string
	sharedBaseURL    = "http://127.0.0.1:8912/mcp"
)

func TestMain(m *testing.M) {
	exitCode := func() int {
		// Create session
		session, err := db.NewSessionFromCluster(cluster, "cassandra", false)
		if err != nil {
			return 1
		}
		defer session.Close()

		// Create MCP handler (shared)
		sharedMCPHandler = router.NewMCPHandler(session)

		// Generate API key ONCE
		sharedAPIKey, err = ai.GenerateAPIKey()
		os.Setenv("TEST_MCP_API_KEY", sharedAPIKey)

		// Start MCP server ONCE
		result := sharedMCPHandler.HandleMCPCommand(".mcp start --config-file ../testdata/dba.json")

		// Run all tests
		code := m.Run()

		// Stop MCP server ONCE
		sharedMCPHandler.HandleMCPCommand(".mcp stop")

		return code
	}()
	os.Exit(exitCode)
}
```

### Modified setupCQLTest

**Before:**
```go
func setupCQLTest(t *testing.T) *CQLTestContext {
	session, _ := db.NewSessionFromCluster(...)
	mcpHandler := router.NewMCPHandler(session)
	apiKey, _ := ai.GenerateAPIKey()         // ❌ NEW key per test
	result := mcpHandler.HandleMCPCommand(".mcp start ...") // ❌ Start per test
	sessionID := initializeMCPSessionHTTP(t, url, apiKey)
	// ...
}
```

**After:**
```go
func setupCQLTest(t *testing.T) *CQLTestContext {
	session, _ := db.NewSessionFromCluster(...)

	// Use SHARED MCP handler and API key ✅
	require.NotNil(t, sharedMCPHandler)
	require.NotEmpty(t, sharedAPIKey)

	sessionID := initializeMCPSessionHTTP(t, sharedBaseURL, sharedAPIKey)
	// ...
}
```

### Modified teardownCQLTest

**Before:**
```go
func teardownCQLTest(ctx *CQLTestContext) {
	// Drop keyspace
	ctx.Session.Query("DROP KEYSPACE...").Exec()

	// Stop MCP ❌ (every test stops the server!)
	ctx.MCPHandler.HandleMCPCommand(".mcp stop")

	ctx.Session.Close()
}
```

**After:**
```go
func teardownCQLTest(ctx *CQLTestContext) {
	// Drop keyspace
	ctx.Session.Query("DROP KEYSPACE...").Exec()

	// Close session
	ctx.Session.Close()

	// NOTE: Do NOT stop MCP server - shared across tests ✅
}
```

---

## Verification

**Before fix:**
```bash
go test -run "^TestDML_Insert_" -p 1
# FAIL: 1/78 tests pass (1.3%)
# Only test 1 passed, all others failed with "401 invalid API key"
```

**After fix:**
```bash
go clean -testcache
go test -run "^TestDML_Insert_" -p 1
# PASS: 78/78 tests pass (100%)
# Duration: 273 seconds (~3.5s per test)
```

---

## Impact

### Before
- ❌ Could only run tests individually
- ❌ No way to run full suite
- ❌ Massive overhead (server start/stop per test)
- ❌ Flaky due to server lifecycle timing

### After
- ✅ Can run entire suite sequentially
- ✅ 30+ tests in single `go test` invocation
- ✅ Reduced overhead (server starts once)
- ✅ Stable and reliable

---

## Files Modified

**test/integration/mcp/cql/base_helpers_test.go:**
- Added `TestMain()` function
- Added shared variables: `sharedMCPHandler`, `sharedAPIKey`, `sharedBaseURL`
- Modified `setupCQLTest()` to use shared infrastructure
- Modified `teardownCQLTest()` to NOT stop MCP server
- Added `contains()` helper function

---

## Testing Status

✅ **All 78 INSERT tests NOW WORK sequentially**

**Verified with full test run:**
```bash
go clean -testcache
go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_" -v -p 1

# Result:
# 78/78 PASS
# 0 failures
# Duration: 273.2 seconds (4 minutes 33 seconds)
# Average: 3.5 seconds per test
```

**All tests show "MCP session initialized" (success path), not "MCP session generated" (fallback path)**

---

**CRITICAL: This fix is essential for the test suite to function properly!**

**Status: ✅ VERIFIED - Ready to proceed with UPDATE test suite**
