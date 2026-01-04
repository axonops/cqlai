# Resume Prompt: Cluster Metadata Manager Implementation

**Use this prompt to start cluster metadata manager implementation**

---

## Context

I am implementing a cluster metadata manager for CQLAI that will provide schema and topology information from the Apache Cassandra gocql driver. This is required to unblock 34+ test scenarios that need schema awareness.

**Current branch:** feature/mcp_datatypes

**Background:**
- We have 78 INSERT tests with CQL assertions working
- 96 INSERT test scenarios are blocked waiting for schema metadata
- Primary key validation, BATCH cross-partition detection, static column semantics all require schema information
- Tests are designed and documented, waiting for metadata implementation

---

## Requirements

**Read this file FIRST:**
`claude-notes/cluster-metadata-requirements.md`

This file will contain:
- Detailed requirements for cluster metadata manager
- gocql driver metadata documentation
- API design requirements
- Integration test requirements
- Performance requirements

---

## Implementation Location

**Package:** `internal/cluster/`

**Purpose:**
- Encapsulate gocql metadata (schema + topology)
- Provide clean API for schema lookups
- Cache metadata with refresh capabilities
- Handle schema change events
- Expose partition keys, clustering keys, static columns
- Expose table/column information
- Eventually: topology, token rings, datacenter info

---

## Critical Requirements

### 1. Integration Tests FIRST

**Before any planner integration, write integration tests:**

**Test: Cluster metadata retrieval**
- Create table in Cassandra
- Retrieve metadata
- Verify partition keys, clustering keys identified correctly
- Verify static columns identified
- Verify regular columns identified

**Test: Schema change detection**
- Get table metadata
- ALTER TABLE (add column)
- Refresh metadata
- Verify new column appears

**Test: CREATE/DROP table**
- Verify table doesn't exist
- CREATE TABLE
- Refresh, verify table exists
- DROP TABLE
- Refresh, verify table gone

**Test: Composite partition keys**
- Table with (col1, col2) as partition key
- Verify both identified as partition keys

**Test: Multiple clustering keys**
- Table with 3 clustering columns
- Verify order preserved
- Verify ASC/DESC captured

**ALL 5 TESTS MUST PASS** before proceeding to planner integration!

---

### 2. Clean Abstraction

**Hide gocql specifics:**
- Don't expose gocql types directly
- Wrap in our own types
- Make it easy to use throughout codebase

**Reusable everywhere:**
- Query planner
- MCP server
- Session
- Tests

---

### 3. Performance

**Caching:**
- Cache schema metadata
- Invalidate on schema changes
- Manual refresh for now (events later)

**Lazy loading:**
- Only fetch when needed
- Don't pre-load all keyspaces

---

## Integration Points

### Where It's Needed

**1. Query Planner (internal/ai/planner.go)**
- Validate INSERT has all PK columns
- Validate UPDATE has full PK (or partial for static)
- Validate DELETE has at least partition key
- BATCH cross-partition detection
- Return errors BEFORE generating CQL

**2. MCP Server (internal/ai/mcp.go)**
- Schema lookups for validation
- Expose metadata via tools

**3. Tests (test/integration/mcp/cql/)**
- 34 tests blocked on schema metadata
- Primary key validation (15)
- BATCH validation (partial - 16)
- Error scenarios (partial - 10)
- Static column tests (5)

---

## Test Scenarios That Need This

**From INSERT_GAP_ANALYSIS.md:**

**Primary Key Validation (15 tests):**
- INSERT missing partition key → Error
- INSERT missing clustering key → Error
- UPDATE with partial PK + regular column → Error
- UPDATE with partial PK + static column → Success
- DELETE with no WHERE → Error
- DELETE with partial PK → Success (partition/range delete)

**BATCH Cross-Partition (4 tests):**
- Detect when BATCH spans multiple partitions → Warning
- Detect when BATCH LWT crosses partitions → Error
- Composite partition key validation

**Static Column Tests (5 tests):**
- Identify static columns
- Allow partial PK for static UPDATE
- Reject partial PK for regular UPDATE

**Error Tests (10 tests):**
- Missing PK validation
- Invalid WHERE validation

---

## Implementation Sequence

### Step 1: Read Requirements
- Read `claude-notes/cluster-metadata-requirements.md` (user will provide)
- Understand gocql metadata API
- Understand caching requirements
- Understand event handling requirements

### Step 2: Design Package
- Design types (TableMetadata, ColumnInfo, etc.)
- Design API (GetTableMetadata, etc.)
- Design caching strategy
- Design refresh mechanism

### Step 3: Implement Package
- Create `internal/cluster/` package
- Implement metadata retrieval
- Implement caching
- Implement refresh

### Step 4: Write Integration Tests
- 5 integration tests as specified above
- Test against real Cassandra
- All must pass

### Step 5: Integrate into Planner
- Add metadata to planner
- Implement validation functions
- Update callers

### Step 6: Resume Testing
- Return to RESUME_PROMPT_TESTING.md
- Implement 96 missing test scenarios
- Use cluster metadata for validation

---

## Success Criteria

**Cluster metadata implementation is complete when:**
1. ✅ All 5 integration tests pass
2. ✅ Can retrieve partition keys, clustering keys, static columns
3. ✅ Can detect schema changes (CREATE, ALTER, DROP)
4. ✅ Integrated into planner with validation
5. ✅ Can implement first primary key validation test successfully

---

## Documentation to Update After Implementation

After cluster metadata is complete:
1. Update `RESUME_TESTING_SESSION.md` - Mark Phase 1 complete
2. Update `INSERT_GAP_ANALYSIS.md` - Note schema dependency resolved
3. Create `internal/cluster/IMPLEMENTATION.md` - Document what was built

---

**Ready to start: Read claude-notes/cluster-metadata-requirements.md and begin implementation**
