# Cluster Metadata Manager Implementation

**Key Architecture Principle**: This is a thin utility wrapper around gocql's native metadata. The gocql driver handles all caching, refreshes, and invalidation automatically. We simply translate gocql's types to our cleaner wrapper types on-demand. This not just for use in our AI and MCP code, but can be used in any context as it mearly using the go cql driver to return this information. Its just provides lots of useful methods that will allow us to support complex metadata requirents in our query planner.

The goal of this session is to JUST implement this manager - nothing else. We will handel the integration into the broader codebase later. 

When writing tests and saying your done on a test - UNLESS ITS RUN THANS THATS A FUCKING LIE! Test mean shit unless they are run - never say your finished on tests or provide postive "i am finished, yay!!" type of responses when tests are not run! Thats what shit developers do! Test, test, test and make sure they run and pass and if they fail it shoudl be triaged with me.

Always run go clean -testcache before running tests to prevent caching issues on test results

NEVER skip tests or assertions based on things being not implemented in tests as this is probably a bug!

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

---

## Context

I am implementing a cluster metadata manager for CQLAI that will provide schema and topology information directly from the Apache Cassandra gocql driver. This is required to unblock 34+ test scenarios that need schema awareness and support enhanced functionality both within our AI and MCP code and the broader cqlai application.

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
- All caching and refreshing of caching is handled within the go driver - we are just calling its functionality and providing funcyions to access tyhe data
- Handle schema change events (via the Go driver, it does this for us)
- Expose partition keys, clustering keys, static columns
- Expose table/column information
- Expose topology, token rings, datacenter info
- Test that when you connect to a cluster on 1 session, then make a schema change in another session, then the metadata for session 1 is automatically updated via the go cql driver internal functionality.

---

## Critical Requirements

### 0. Exploratory testing and discovery - critical before we do anything else

Before we do anything we need to test and investigate the go driver behaviour is as we expect, what the meta data looks like etc..

- We need to check what the go driver metadata looks like so we understand it. COnnect up via the driver, look at what we can retrieve from the driver metadata and that any assumptions we have made are correct and if theres ant other information that would be worth exposing
- This is a critical and complex test. What you need to do is establish that cluster events get propagated and caches refreshed within the go driver etc.. We are assuming that the driver behaves like this based on docs and such, but if it doesnt we are screwwd. To do this we need a go program that opens a connection to the cluster and is sitting there looking at the cluster metad data, logging it out. Then while that go progam has an active and open session we need to connect seperately to the podman container running cassandra and create a table or something, then we shoudl see in the already running go progam this schema data get propagated automtically via the internals of the go driver. If this does not work - we have a problem so stop and we will need to understand whats happening. This feature in the go driver where cluster events are sent and automtically propagated to everyone connect to the cluster is critical to our architecture. If its not, we have a problem.

Do NO analysis, planning or implementation until we understand the above and have established whats possible and whats not!

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
- No caching is needed - this is internal to the go driver

**Lazy loading:**
- No lazy loading is needed - this is internal to the go driver

---

## Integration Points

### Where It's Needed once we implement it

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

### Step 2: Design Package
- Design types (TableMetadata, ColumnInfo, etc.)
- Design API (GetTableMetadata, etc.)
- Design caching strategy
- Design refresh mechanism

### Step 3: Implement Package
- Create `internal/cluster/` package
- Implement metadata retrieval

### Step 4: Write Unit and Integration Tests
- integration tests as specified above
- Test against real Cassandra
- When in doubt or confused,just connect to a real cassandra cluster and see what it looks like. If you need a multi-dc cluster let me know and I can give you the details
- All must pass


Once we are done, we will start a new session integrating this into the code and tests. The goal is to get this working and tested so we can start using it.

---

## Success Criteria

**Cluster metadata implementation is complete when:**
1. ✅ All integration and unit tests pass
2. ✅ Can retrieve partition keys, clustering keys, static columns
3. ✅ Can detect schema changes (CREATE, ALTER, DROP)
4. ✅ Integrated into planner with validation
5. ✅ Can implement first primary key validation test successfully

---

## Documentation to Update After Implementation

After cluster metadata is complete:
1. Update `RESUME_TESTING_SESSION.md` - Mark Phase 1 complete
2. Update `INSERT_GAP_ANALYSIS.md` - Note schema dependency resolved
3. Create `internal/cluster/README.md` - Document what was built

---

**Ready to start: Read claude-notes/cluster-metadata-requirements.md and begin implementation**
