When writing tests and saying your done on a test - UNLESS ITS RUN THANS THATS A FUCKING LIE! Test mean shit unless they are run - never say your finished on tests or provide postive "i am finished, yay!!" type of responses when tests are not run! Thats what shit developers do! Test, test, test and make sure they run and pass and if they fail it shoudl be triaged with me.

Always run go clean -testcache before running tests to prevent caching issues on test results

NEVER skip tests or assertions based on things being not implemented in tests as this is probably a bug! The point of these tests is that we are testing we havd 100% complete coverage on everything you can do with CQL, if you start skipping things your hidng a bug.

## CRITICAL: Shell Command Rules
When appending or writing content to files, always use the Write or Edit tools instead of bash commands like `cat >> file <<'EOF'`. Never use heredocs or shell redirection for file modifications.

When running CQL commands via podman, write the CQL to a temp .cql file first, then execute with: podman exec cassandra-test cqlsh -f /path/to/file.cql . Never use heredocs or pipes with podman commands.

Run git commands separately, never chain with &&

NEVER use these shell operators in bash commands:
- No pipes: `|`
- No redirections: `>`, `>>`, `<`, `<<`
- No heredocs: `<< 'EOF'`
- No chaining: `&&`, `||`, `;`

Instead:
- Run commands separately one at a time
- Use Write/Edit tools for file content
- Let me see full command output (no grep filtering)


Continue implementing comprehensive CQL test suite for CQLAI MCP server.

I am continuing a systematic test implementation that has been very successful.
We've completed 90 DML INSERT tests with full validation and found/fixed 5 bugs.

CURRENT STATE:
- Branch: feature/mcp_datatypes
- Status: DML INSERT suite COMPLETE (90/90 tests, 100% passing)
- Progress: 90/1,200 tests (7.5%)
- Tag: session-complete-90-insert-tests
- Token budget: 546K remaining

CRITICAL: READ ALL THESE FILES IN ORDER BEFORE STARTING:

Phase 1 - Understanding the Blueprint (MUST READ):
1. claude-notes/cql-complete-test-suite.md - Master blueprint defining ALL 1,200+ tests
2. claude-notes/cql-implementation-guide.md - 20+ reusable test patterns
3. claude-notes/test-suite-summary.md - 8-week execution roadmap
4. claude-notes/c5-nesting-cql.md - Cassandra 5 nesting rules (CRITICAL for frozen keyword)
5. claude-notes/c5-nesting-mtx.md - Complete nesting test matrix

Phase 2 - Understanding Current Progress (MUST READ):
6. test/integration/mcp/cql/PROGRESS_TRACKER.md - Real-time progress (90/1,200 done)
7. test/integration/mcp/cql/FINAL_SESSION_SUMMARY.md - Last session accomplishments
8. test/integration/mcp/cql/BUGS_FOUND_AND_FIXED.md - 5 bugs found and how they were fixed

Phase 3 - Critical Testing Rules (MUST READ):
9. test/integration/mcp/cql/LWT_TESTING_GUIDELINES.md - LWT vs non-LWT separation (CRITICAL!)
10. test/integration/mcp/cql/README.md - Test suite structure and principles

Phase 4 - Analysis and Context:
11. claude-notes/cql_test_matrix.md - Analysis of existing 73 tests
12. claude-notes/cql_coverage_gaps.md - 479 lines of gap analysis

Phase 5 - Implementation Reference:
13. test/integration/mcp/cql/dml_insert_test.go - 90 completed tests to use as template

VERIFY INFRASTRUCTURE:
Run ONE test to confirm everything works:
  cd /Users/johnny/Development/cqlai
  git checkout feature/mcp_datatypes
  go test ./test/integration/mcp/cql -tags=integration -run "^TestDML_Insert_01_" -v

Should complete in ~4 seconds and show:
  ✅ Test 1: Simple text - Full CRUD cycle verified
  --- PASS: TestDML_Insert_01_SimpleText

NEXT TASK: Create DML UPDATE test suite
File: test/integration/mcp/cql/dml_update_test.go
Target: Start with tests 1-25 (first checkpoint)
Blueprint: claude-notes/cql-complete-test-suite.md section "DML UPDATE Tests"