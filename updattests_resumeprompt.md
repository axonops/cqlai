Continue implementing the comprehensive CQL test suite for CQLAI MCP server.

  CURRENT STATE:
  - Branch: feature/mcp_datatypes
  - Status: DML INSERT suite COMPLETE (90/90 tests, all passing)
  - Progress: 90/1,200 tests (7.5%)
  - Tag: session-complete-90-insert-tests

  READ THESE FILES FIRST (in this order):
  1. test/integration/mcp/cql/RESUME_NEXT_SESSION.md - Complete context
  2. test/integration/mcp/cql/PROGRESS_TRACKER.md - Current progress
  3. test/integration/mcp/cql/LWT_TESTING_GUIDELINES.md - CRITICAL LWT rules
  4. test/integration/mcp/cql/BUGS_FOUND_AND_FIXED.md - Bugs found so far

  VERIFY CURRENT STATE:
  Run: go test ./test/integration/mcp/cql -tags=integration -run "TestDML_Insert_01" -v
  Should pass in ~4 seconds

  NEXT TASK:
  Create test/integration/mcp/cql/dml_update_test.go
  Implement UPDATE tests 1-25 (first batch)

  CRITICAL RULES:
  1. EVERY test MUST validate data in Cassandra (direct query)
  2. EVERY test MUST include full CRUD cycle
  3. LWT tests use DELETE IF EXISTS (not regular DELETE)
  4. Non-LWT tests use regular DELETE
  5. RUN tests immediately - don't just write them
  6. Fix bugs as found - no workarounds
  7. Checkpoint every 5-10 tests
  8. Update PROGRESS_TRACKER.md after each checkpoint

  Follow the exact pattern from dml_insert_test.go tests.
  See RESUME_NEXT_SESSION.md for complete details.

  Token budget: 546K remaining (54.6%)
  Estimated: Can complete 25-30 UPDATE tests this session.