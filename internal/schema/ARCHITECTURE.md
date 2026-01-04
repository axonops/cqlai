# Schema Metadata Architecture

**Date:** 2026-01-04
**Purpose:** Design schema-aware metadata layer for query validation and planning
**Status:** Design phase - implementation pending

---

## Requirements

1. **Encapsulate gocql metadata** - Hide driver-specific details
2. **Reusable everywhere** - Session, planner, MCP server, tests
3. **Event-driven updates** - Refresh on schema changes
4. **Performance** - Cached, only refresh when needed
5. **Testing first** - Integration tests before planner integration

---

## Architecture

### Package: `internal/schema`

```
internal/schema/
├── manager.go          # MetadataManager - main API
├── table.go            # TableMetadata wrapper
├── column.go           # ColumnInfo types
├── events.go           # Schema change event handling
├── cache.go            # Caching layer
├── manager_test.go     # Unit tests
└── integration_test.go # Integration tests (schema refresh)
```

---

## Core Types

### MetadataManager

```go
type MetadataManager struct {
    session *db.Session
    cache   *sync.Map  // map[string]*TableMetadata (key: "keyspace.table")
    mu      sync.RWMutex
}

// Constructor
func NewMetadataManager(session *db.Session) *MetadataManager

// Primary API
func (m *MetadataManager) GetTableMetadata(keyspace, table string) (*TableMetadata, error)
func (m *MetadataManager) GetKeyspaceMetadata(keyspace string) (*KeyspaceMetadata, error)

// Cache management
func (m *MetadataManager) RefreshTable(keyspace, table string) error
func (m *MetadataManager) RefreshKeyspace(keyspace string) error
func (m *MetadataManager) InvalidateTable(keyspace, table string)
func (m *MetadataManager) InvalidateKeyspace(keyspace string)
func (m *MetadataManager) ClearAll()

// Event handling (future)
func (m *MetadataManager) HandleSchemaEvent(event *SchemaChangeEvent) error
```

### TableMetadata

```go
type TableMetadata struct {
    Keyspace string
    Table    string

    // Primary key structure
    PartitionKeys  []ColumnInfo  // In schema-defined order
    ClusteringKeys []ColumnInfo  // In schema-defined order with ASC/DESC

    // Column categories
    RegularColumns map[string]*ColumnInfo  // By name
    StaticColumns  map[string]*ColumnInfo  // By name
    AllColumns     map[string]*ColumnInfo  // All columns by name

    // Table properties
    Options map[string]interface{}  // compaction, compression, etc.
}

// Helper methods
func (t *TableMetadata) IsPartitionKey(columnName string) bool
func (t *TableMetadata) IsClusteringKey(columnName string) bool
func (t *TableMetadata) IsStaticColumn(columnName string) bool
func (t *TableMetadata) HasColumn(columnName string) bool
func (t *TableMetadata) GetColumnType(columnName string) (string, error)
func (t *TableMetadata) GetPartitionKeyColumns() []string
func (t *TableMetadata) GetAllPrimaryKeyColumns() []string  // Partition + clustering
func (t *TableMetadata) ValidateFullPrimaryKey(columns []string) error
```

### ColumnInfo

```go
type ColumnInfo struct {
    Name     string
    Type     string  // CQL type: "int", "text", "frozen<address>", etc.
    Kind     ColumnKind
    Position int     // Position in partition/clustering keys (0-based)

    // For clustering columns only
    ClusteringOrder string  // "ASC" or "DESC"
}

type ColumnKind int

const (
    PartitionKeyColumn ColumnKind = iota
    ClusteringKeyColumn
    RegularColumn
    StaticColumn
)

func (k ColumnKind) String() string
```

---

## Implementation Strategy

### Phase 1: Basic Metadata Retrieval (No Events)

**File:** `internal/schema/manager.go`

```go
func (m *MetadataManager) GetTableMetadata(keyspace, table string) (*TableMetadata, error) {
    // 1. Check cache
    cacheKey := keyspace + "." + table
    if cached, ok := m.cache.Load(cacheKey); ok {
        return cached.(*TableMetadata), nil
    }

    // 2. Fetch from Cassandra via gocql
    meta, err := m.fetchTableMetadataFromDriver(keyspace, table)
    if err != nil {
        return nil, err
    }

    // 3. Store in cache
    m.cache.Store(cacheKey, meta)

    return meta, nil
}

func (m *MetadataManager) fetchTableMetadataFromDriver(keyspace, table string) (*TableMetadata, error) {
    // Use gocql's KeyspaceMetadata
    ksMeta, err := m.session.KeyspaceMetadata(keyspace)
    if err != nil {
        return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
    }

    tableMeta, ok := ksMeta.Tables[table]
    if !ok {
        return nil, fmt.Errorf("table %s.%s not found in schema", keyspace, table)
    }

    // Convert gocql.TableMetadata to our TableMetadata
    return convertGoCQLTableMetadata(tableMeta), nil
}

func convertGoCQLTableMetadata(gocqlTable *gocql.TableMetadata) *TableMetadata {
    meta := &TableMetadata{
        Keyspace:       gocqlTable.Keyspace,
        Table:          gocqlTable.Name,
        PartitionKeys:  make([]ColumnInfo, len(gocqlTable.PartitionKey)),
        ClusteringKeys: make([]ColumnInfo, len(gocqlTable.ClusteringColumns)),
        RegularColumns: make(map[string]*ColumnInfo),
        StaticColumns:  make(map[string]*ColumnInfo),
        AllColumns:     make(map[string]*ColumnInfo),
    }

    // Extract partition key columns
    for i, col := range gocqlTable.PartitionKey {
        colInfo := &ColumnInfo{
            Name:     col.Name,
            Type:     col.Type,
            Kind:     PartitionKeyColumn,
            Position: i,
        }
        meta.PartitionKeys[i] = *colInfo
        meta.AllColumns[col.Name] = colInfo
    }

    // Extract clustering columns
    for i, col := range gocqlTable.ClusteringColumns {
        colInfo := &ColumnInfo{
            Name:            col.Name,
            Type:            col.Type,
            Kind:            ClusteringKeyColumn,
            Position:        i,
            ClusteringOrder: col.Order,  // "ASC" or "DESC"
        }
        meta.ClusteringKeys[i] = *colInfo
        meta.AllColumns[col.Name] = colInfo
    }

    // Extract regular and static columns
    for name, col := range gocqlTable.Columns {
        // Skip if already processed as PK
        if meta.AllColumns[name] != nil {
            continue
        }

        colInfo := &ColumnInfo{
            Name: name,
            Type: col.Type,
            Kind: RegularColumn,
        }

        // Check if static (gocql has this info)
        if col.Kind == gocql.ColumnStatic {
            colInfo.Kind = StaticColumn
            meta.StaticColumns[name] = colInfo
        } else {
            meta.RegularColumns[name] = colInfo
        }

        meta.AllColumns[name] = colInfo
    }

    return meta
}
```

---

## Integration Tests (FIRST - Before Planner Integration)

### Test 1: Basic Retrieval

```go
func TestSchemaManager_BasicRetrieval(t *testing.T) {
    // 1. Connect to Cassandra
    session := setupCassandraSession(t)
    defer session.Close()

    // 2. Create test schema
    session.Query(`CREATE KEYSPACE IF NOT EXISTS schema_test
        WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec()
    session.Query(`CREATE TABLE IF NOT EXISTS schema_test.users (
        user_id int,
        session_id int,
        name text,
        email text STATIC,
        PRIMARY KEY (user_id, session_id)
    )`).Exec()

    // 3. Get metadata
    mgr := schema.NewMetadataManager(session)
    meta, err := mgr.GetTableMetadata("schema_test", "users")

    require.NoError(t, err)

    // 4. Validate partition keys
    assert.Len(t, meta.PartitionKeys, 1)
    assert.Equal(t, "user_id", meta.PartitionKeys[0].Name)
    assert.Equal(t, PartitionKeyColumn, meta.PartitionKeys[0].Kind)

    // 5. Validate clustering keys
    assert.Len(t, meta.ClusteringKeys, 1)
    assert.Equal(t, "session_id", meta.ClusteringKeys[0].Name)
    assert.Equal(t, ClusteringKeyColumn, meta.ClusteringKeys[0].Kind)

    // 6. Validate static columns
    assert.Len(t, meta.StaticColumns, 1)
    emailCol, ok := meta.StaticColumns["email"]
    require.True(t, ok)
    assert.Equal(t, StaticColumn, emailCol.Kind)

    // 7. Validate regular columns
    nameCol, ok := meta.RegularColumns["name"]
    require.True(t, ok)
    assert.Equal(t, RegularColumn, nameCol.Kind)

    // 8. Test helper methods
    assert.True(t, meta.IsPartitionKey("user_id"))
    assert.True(t, meta.IsClusteringKey("session_id"))
    assert.True(t, meta.IsStaticColumn("email"))
    assert.False(t, meta.IsStaticColumn("name"))
}
```

### Test 2: Schema Change Refresh

```go
func TestSchemaManager_SchemaChangeRefresh(t *testing.T) {
    session := setupCassandraSession(t)
    defer session.Close()

    // 1. Create initial schema
    session.Query(`CREATE KEYSPACE IF NOT EXISTS schema_test
        WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec()
    session.Query(`CREATE TABLE IF NOT EXISTS schema_test.test_table (
        id int PRIMARY KEY,
        col1 text
    )`).Exec()

    mgr := schema.NewMetadataManager(session)

    // 2. Get initial metadata
    meta, _ := mgr.GetTableMetadata("schema_test", "test_table")
    assert.False(t, meta.HasColumn("col2"), "col2 should not exist yet")

    // 3. ALTER TABLE - add column
    session.Query(`ALTER TABLE schema_test.test_table ADD col2 text`).Exec()

    // Wait for schema propagation
    time.Sleep(2 * time.Second)

    // 4. Refresh metadata
    err := mgr.RefreshTable("schema_test", "test_table")
    require.NoError(t, err)

    // 5. Verify new column appears
    meta, _ = mgr.GetTableMetadata("schema_test", "test_table")
    assert.True(t, meta.HasColumn("col2"), "col2 should exist after ALTER")

    // 6. Verify column details
    col2, ok := meta.RegularColumns["col2"]
    require.True(t, ok)
    assert.Equal(t, "text", col2.Type)
}
```

### Test 3: CREATE/DROP Table

```go
func TestSchemaManager_CreateDropTable(t *testing.T) {
    session := setupCassandraSession(t)
    defer session.Close()

    session.Query(`CREATE KEYSPACE IF NOT EXISTS schema_test
        WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec()

    mgr := schema.NewMetadataManager(session)

    // 1. Table doesn't exist yet
    _, err := mgr.GetTableMetadata("schema_test", "new_table")
    assert.Error(t, err, "Should error on non-existent table")

    // 2. CREATE TABLE
    session.Query(`CREATE TABLE schema_test.new_table (
        id int PRIMARY KEY,
        data text
    )`).Exec()

    time.Sleep(2 * time.Second)

    // 3. Refresh and verify table exists
    mgr.RefreshKeyspace("schema_test")
    meta, err := mgr.GetTableMetadata("schema_test", "new_table")
    require.NoError(t, err)
    assert.Equal(t, "new_table", meta.Table)

    // 4. DROP TABLE
    session.Query(`DROP TABLE schema_test.new_table`).Exec()

    time.Sleep(2 * time.Second)

    // 5. Invalidate cache and verify table gone
    mgr.InvalidateTable("schema_test", "new_table")
    _, err = mgr.GetTableMetadata("schema_test", "new_table")
    assert.Error(t, err, "Should error after DROP TABLE")
}
```

### Test 4: Composite Partition Keys

```go
func TestSchemaManager_CompositePartitionKey(t *testing.T) {
    session := setupCassandraSession(t)
    defer session.Close()

    session.Query(`CREATE KEYSPACE IF NOT EXISTS schema_test
        WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec()
    session.Query(`CREATE TABLE IF NOT EXISTS schema_test.multi_tenant (
        tenant_id int,
        user_id int,
        data text,
        PRIMARY KEY ((tenant_id, user_id))  -- Composite partition key
    )`).Exec()

    mgr := schema.NewMetadataManager(session)
    meta, _ := mgr.GetTableMetadata("schema_test", "multi_tenant")

    // Verify both columns are partition keys
    assert.Len(t, meta.PartitionKeys, 2)
    assert.Equal(t, "tenant_id", meta.PartitionKeys[0].Name)
    assert.Equal(t, "user_id", meta.PartitionKeys[1].Name)

    // Verify both marked as partition keys
    assert.True(t, meta.IsPartitionKey("tenant_id"))
    assert.True(t, meta.IsPartitionKey("user_id"))

    // Verify no clustering keys
    assert.Len(t, meta.ClusteringKeys, 0)
}
```

### Test 5: Multiple Clustering Keys

```go
func TestSchemaManager_MultipleClusteringKeys(t *testing.T) {
    session := setupCassandraSession(t)
    defer session.Close()

    session.Query(`CREATE KEYSPACE IF NOT EXISTS schema_test
        WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec()
    session.Query(`CREATE TABLE IF NOT EXISTS schema_test.events (
        user_id int,
        year int,
        month int,
        day int,
        event text,
        PRIMARY KEY (user_id, year, month, day)
    ) WITH CLUSTERING ORDER BY (year DESC, month DESC, day DESC)`).Exec()

    mgr := schema.NewMetadataManager(session)
    meta, _ := mgr.GetTableMetadata("schema_test", "events")

    // Verify clustering keys in order
    assert.Len(t, meta.ClusteringKeys, 3)
    assert.Equal(t, "year", meta.ClusteringKeys[0].Name)
    assert.Equal(t, "month", meta.ClusteringKeys[1].Name)
    assert.Equal(t, "day", meta.ClusteringKeys[2].Name)

    // Verify clustering order
    assert.Equal(t, "DESC", meta.ClusteringKeys[0].ClusteringOrder)
    assert.Equal(t, "DESC", meta.ClusteringKeys[1].ClusteringOrder)
    assert.Equal(t, "DESC", meta.ClusteringKeys[2].ClusteringOrder)

    // Verify positions
    assert.Equal(t, 0, meta.ClusteringKeys[0].Position)
    assert.Equal(t, 1, meta.ClusteringKeys[1].Position)
    assert.Equal(t, 2, meta.ClusteringKeys[2].Position)
}
```

---

## Integration with Query Planner

### Updated RenderCQL Signature

**Current:**
```go
func RenderCQL(plan *AIResult) (string, error)
```

**New:**
```go
func RenderCQL(plan *AIResult, schemaMgr *schema.MetadataManager) (string, error)
```

**OR use planner struct:**
```go
type CQLPlanner struct {
    schemaMgr *schema.MetadataManager
}

func (p *CQLPlanner) RenderCQL(plan *AIResult) (string, error)
```

### Validation Before Rendering

```go
func (p *CQLPlanner) RenderCQL(plan *AIResult) (string, error) {
    // 1. Get table metadata (if operation requires it)
    if requiresTableMetadata(plan.Operation) {
        meta, err := p.schemaMgr.GetTableMetadata(plan.Keyspace, plan.Table)
        if err != nil {
            return "", fmt.Errorf("schema lookup failed: %w", err)
        }

        // 2. Validate before rendering
        if err := validateAgainstSchema(plan, meta); err != nil {
            return "", err  // Error before generating CQL
        }

        // 3. Render with schema awareness
        return renderWithSchema(plan, meta)
    }

    // Fallback for operations without schema requirements
    return renderWithoutSchema(plan)
}

func validateAgainstSchema(plan *AIResult, meta *TableMetadata) error {
    switch plan.Operation {
    case "INSERT":
        return validateInsert(plan, meta)
    case "UPDATE":
        return validateUpdate(plan, meta)
    case "DELETE":
        return validateDelete(plan, meta)
    case "BATCH":
        return validateBatch(plan, meta)
    }
    return nil
}

func validateInsert(plan *AIResult, meta *TableMetadata) error {
    // Check all partition keys present
    for _, pk := range meta.PartitionKeys {
        if _, ok := plan.Values[pk.Name]; !ok {
            return fmt.Errorf("INSERT missing partition key column: %s", pk.Name)
        }
    }

    // Check all clustering keys present
    for _, ck := range meta.ClusteringKeys {
        if _, ok := plan.Values[ck.Name]; !ok {
            return fmt.Errorf("INSERT missing clustering key column: %s", ck.Name)
        }
    }

    return nil
}
```

---

## Implementation Sequence

### Step 1: Create Schema Package (Day 1)
- Create `internal/schema/` package
- Implement `manager.go`, `table.go`, `column.go`
- NO event handling yet (manual refresh only)

### Step 2: Integration Tests (Day 1-2)
- Write 5 integration tests as shown above
- Test basic retrieval
- Test schema refresh (manual)
- Test CREATE/DROP/ALTER
- Test composite keys
- Test multiple clustering keys
- **All tests must PASS before proceeding**

### Step 3: Integrate into Planner (Day 2)
- Add MetadataManager to planner
- Implement validation functions
- Update RenderCQL signature
- Update all callers

### Step 4: Add Primary Key Validation Tests (Day 3-4)
- Implement 15 PK validation tests
- Use schema metadata to validate PK requirements
- Test INSERT/UPDATE/DELETE with full/partial PK

### Step 5: Event Handling (Future)
- Listen for schema change events
- Auto-refresh on CREATE/ALTER/DROP
- This can come later

---

## Questions to Resolve

1. **Where to instantiate MetadataManager?**
   - In Session? (one per session)
   - In MCPServer? (one per MCP server)
   - Pass through everywhere?

2. **Cache invalidation strategy?**
   - Manual refresh only (simple)
   - TTL-based (automatic but may be stale)
   - Event-driven (complex but always fresh)

3. **Error handling:**
   - Fail fast if schema not found?
   - Allow rendering without schema validation?
   - Configurable validation level?

4. **Testing mode:**
   - Mock schema in unit tests?
   - Require real Cassandra for schema tests?

---

## My Recommendations

1. **MetadataManager in Session** - Already has DB connection
2. **Manual refresh for now** - Simple, explicit
3. **Fail fast on schema errors** - Catch issues early
4. **Real Cassandra for integration tests** - More accurate

**Next step:** Implement `internal/schema/` package with integration tests, THEN integrate into planner.

**Should I proceed with creating the schema package?**
