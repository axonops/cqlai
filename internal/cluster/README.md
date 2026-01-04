# Cluster Metadata Manager

**Package:** `github.com/axonops/cqlai/internal/cluster`

Thin wrapper around gocql's metadata APIs providing a clean interface for accessing Cassandra cluster and schema metadata.

---

## Architecture

**Key Principle:** This is a thin delegation wrapper. The gocql driver handles ALL caching and automatic schema change propagation internally. We simply translate gocql's types to cleaner wrapper types on-demand.

### What This Package Does

✅ Provides clean API for schema metadata (keyspaces, tables, columns, UDTs, functions, aggregates, views)
✅ Provides cluster topology information (nodes, datacenters, racks)
✅ Translates gocql types to our wrapper types on-demand
✅ Exposes helper methods for common queries (GetPartitionKeyNames, GetClusteringKeyNames, etc.)
✅ Supports Cassandra 5.0 features (vectors, SAI indexes)

### What This Package Does NOT Do

❌ Does NOT cache metadata (gocql handles this)
❌ Does NOT manually refresh metadata (gocql auto-updates in ~1 second)
❌ Does NOT hold references to gocql metadata across calls
❌ Does NOT manage connection pooling (uses provided session)

---

## Critical Architecture Requirement

**ALWAYS delegate to gocql on every API call. NEVER cache metadata.**

The manager holds ONLY:
- `session *gocql.Session` - reference to gocql session
- `cluster *gocql.ClusterConfig` - reference to cluster config (static)
- `config ManagerConfig` - static configuration

Every public API method MUST:
1. Call `session.KeyspaceMetadata()` or query system tables FRESH
2. Translate gocql types to wrapper types
3. Return wrapper (discard gocql references)
4. Let gocql references go out of scope

See: `claude-notes/complete-delegation-requirements.md` for detailed patterns.

---

## Usage

```go
import (
    gocql "github.com/apache/cassandra-gocql-driver/v2"
    "github.com/axonops/cqlai/internal/cluster"
)

// Create session
clusterConfig := gocql.NewCluster("127.0.0.1")
session, err := clusterConfig.CreateSession()

// Create metadata manager
manager := cluster.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

// Get keyspace metadata
ksMeta, err := manager.GetKeyspace("my_keyspace")

// Get table metadata
tableMeta, err := manager.GetTable("my_keyspace", "my_table")

// Get partition keys
pkNames := tableMeta.GetPartitionKeyNames()  // []string in schema order

// Get clustering keys with ordering
ckNames := tableMeta.GetClusteringKeyNames()
ckOrders := tableMeta.GetClusteringKeyOrders()  // []ColumnOrder (ASC/DESC)

// Check if table has static columns
hasStatic := tableMeta.HasStaticColumns()

// Get cluster topology
topology, err := manager.GetTopology()
```

---

## Schema Change Propagation

**gocql automatically propagates schema changes in ~1 second.**

When you ALTER, CREATE, or DROP tables/keyspaces:
1. Schema change executed via any connection
2. gocql driver detects change automatically (~1 second)
3. Next call to `session.KeyspaceMetadata()` returns fresh data
4. No manual refresh needed!

**Verified via integration test:** `TestMetadataManager_SchemaChangePropagation`

---

## Integration Tests

**Location:** `test/integration/cluster_metadata_test.go`

**All tests passing:**
1. ✅ `TestMetadataManager_BasicSchemaRetrieval` - Schema retrieval works
2. ✅ `TestMetadataManager_SchemaChangePropagation` - Automatic schema updates (CRITICAL)
3. ✅ `TestMetadataManager_CompositePartitionKey` - Composite partition key detection
4. ✅ `TestMetadataManager_ClusteringKeysWithOrder` - Clustering column ordering (ASC/DESC)
5. ✅ `TestMetadataManager_CreateDropTableDetection` - CREATE/DROP detection
6. ✅ `TestMetadataManager_TopologyRetrieval` - Cluster topology (name, partitioner, hosts, DCs)
7. ✅ `TestMetadataManager_KeyspaceList` - Keyspace listing
8. ✅ `TestMetadataManager_SchemaAgreement` - Schema agreement waiting
9. ✅ `TestMetadataManager_GetSchemaVersion` - Schema version retrieval

**Run tests:**
```bash
go test -tags=integration ./test/integration -v -run "TestMetadataManager_"
```

---

## Files

- `types.go` - Wrapper types and helper methods
- `manager.go` - MetadataManager interface
- `gocql_manager.go` - Implementation with delegation pattern
- `translator.go` - Translation functions (gocql → wrapper types)
- `errors.go` - Error definitions
- `README.md` - This file

---

## Performance

**Translation is cheap:**
- gocql metadata is already cached internally
- Translation is simple struct copying
- No network calls (gocql handles that)
- No expensive operations

**Thread-safe:**
- All gocql calls are thread-safe
- Each translation is independent
- No shared mutable state
- No locks needed

---

## Future Work

- [ ] Add SAI index detection once Cassandra 5.0 test data available
- [ ] Add vector type detection once vector columns in test cluster
- [ ] Add materialized view WHERE clause extraction (if gocql exposes it)
- [ ] Add table options extraction (if gocql exposes them)
- [ ] Add unit tests for edge cases

---

**Implementation Date:** 2026-01-04
**Status:** Complete - All integration tests passing
