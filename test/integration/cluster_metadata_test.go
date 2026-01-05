// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	clustermd "github.com/axonops/cqlai/internal/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration
const (
	testCassandraHost = "127.0.0.1"
	testKeyspace      = "cluster_metadata_test"
)

// setupTestSession creates a test session to Cassandra
func setupTestSession(t *testing.T) (*gocql.Session, *gocql.ClusterConfig) {
	clusterConfig := gocql.NewCluster(testCassandraHost)
	clusterConfig.Keyspace = "system"
	clusterConfig.Consistency = gocql.Quorum
	clusterConfig.ProtoVersion = 4
	clusterConfig.Timeout = 10 * time.Second
	clusterConfig.Authenticator = gocql.PasswordAuthenticator{
		Username: "cassandra",
		Password: "cassandra",
	}
	clusterConfig.DisableInitialHostLookup = true

	session, err := clusterConfig.CreateSession()
	require.NoError(t, err, "Failed to connect to Cassandra")

	return session, clusterConfig
}

// cleanupTestKeyspace drops the test keyspace
func cleanupTestKeyspace(session *gocql.Session) {
	session.Query(fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", testKeyspace)).Exec()
}

// createTestKeyspace creates the test keyspace
func createTestKeyspace(t *testing.T, session *gocql.Session) {
	cleanupTestKeyspace(session)

	query := fmt.Sprintf(
		"CREATE KEYSPACE %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		testKeyspace,
	)
	err := session.Query(query).Exec()
	require.NoError(t, err, "Failed to create test keyspace")

	// Wait for schema agreement
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = session.AwaitSchemaAgreement(ctx)
	require.NoError(t, err, "Schema agreement failed")
}

// TestMetadataManager_BasicSchemaRetrieval verifies basic schema metadata retrieval
func TestMetadataManager_BasicSchemaRetrieval(t *testing.T) {
	session, clusterConfig := setupTestSession(t)
	defer session.Close()

	createTestKeyspace(t, session)
	defer cleanupTestKeyspace(session)

	// Create test table
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE %s.users (
			id int,
			email text,
			age int,
			created_at timestamp,
			PRIMARY KEY (id)
		)
	`, testKeyspace)
	err := session.Query(createTableQuery).Exec()
	require.NoError(t, err, "Failed to create test table")

	// Wait for schema propagation
	time.Sleep(2 * time.Second)

	// Create metadata manager
	manager := clustermd.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

	t.Run("GetKeyspace returns metadata", func(t *testing.T) {
		ksMeta, err := manager.GetKeyspace(testKeyspace)
		require.NoError(t, err)
		require.NotNil(t, ksMeta)

		assert.Equal(t, testKeyspace, ksMeta.Name)
		assert.True(t, ksMeta.DurableWrites)
		assert.NotNil(t, ksMeta.Replication)
		assert.Equal(t, clustermd.ReplicationStrategySimple, ksMeta.Replication.Class)
		assert.Equal(t, 1, ksMeta.Replication.ReplicationFactor)
	})

	t.Run("GetTable returns table metadata", func(t *testing.T) {
		tableMeta, err := manager.GetTable(testKeyspace, "users")
		require.NoError(t, err)
		require.NotNil(t, tableMeta)

		assert.Equal(t, "users", tableMeta.Name)
		assert.Equal(t, testKeyspace, tableMeta.Keyspace)
		assert.Len(t, tableMeta.Columns, 4)
		assert.Len(t, tableMeta.PartitionKeys, 1)
		assert.Len(t, tableMeta.ClusteringKeys, 0)

		// Verify partition key
		assert.Equal(t, "id", tableMeta.PartitionKeys[0].Name)
		assert.True(t, tableMeta.PartitionKeys[0].IsPartitionKey())
		assert.Equal(t, 0, tableMeta.PartitionKeys[0].ComponentIndex)
	})

	t.Run("GetPartitionKeyNames returns correct keys", func(t *testing.T) {
		tableMeta, err := manager.GetTable(testKeyspace, "users")
		require.NoError(t, err)

		pkNames := tableMeta.GetPartitionKeyNames()
		assert.Equal(t, []string{"id"}, pkNames)
	})

	t.Run("GetRegularColumns filters correctly", func(t *testing.T) {
		tableMeta, err := manager.GetTable(testKeyspace, "users")
		require.NoError(t, err)

		regularCols := tableMeta.GetRegularColumns()
		assert.Len(t, regularCols, 3) // email, age, created_at
	})
}

// TestMetadataManager_SchemaChangePropagation verifies schema changes propagate
//
// CRITICAL TEST: Verifies that gocql automatically updates metadata on schema changes
func TestMetadataManager_SchemaChangePropagation(t *testing.T) {
	session, clusterConfig := setupTestSession(t)
	defer session.Close()

	createTestKeyspace(t, session)
	defer cleanupTestKeyspace(session)

	// Create test table with 2 columns
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE %s.test_table (
			id int PRIMARY KEY,
			name text
		)
	`, testKeyspace)
	err := session.Query(createTableQuery).Exec()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Create metadata manager
	manager := clustermd.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

	// Get initial metadata
	tableMeta1, err := manager.GetTable(testKeyspace, "test_table")
	require.NoError(t, err)
	require.NotNil(t, tableMeta1)
	assert.Len(t, tableMeta1.Columns, 2, "Should have 2 columns initially")

	// ALTER TABLE - add new column
	alterQuery := fmt.Sprintf("ALTER TABLE %s.test_table ADD new_column text", testKeyspace)
	err = session.Query(alterQuery).Exec()
	require.NoError(t, err, "Failed to alter table")

	// Wait for schema propagation (gocql detects changes in ~1 second)
	time.Sleep(2 * time.Second)

	// Get metadata again - should see new column WITHOUT manual refresh
	tableMeta2, err := manager.GetTable(testKeyspace, "test_table")
	require.NoError(t, err)
	require.NotNil(t, tableMeta2)
	assert.Len(t, tableMeta2.Columns, 3, "Should have 3 columns after ALTER")

	// Verify new column exists
	_, hasNewColumn := tableMeta2.Columns["new_column"]
	assert.True(t, hasNewColumn, "New column should be present")

	t.Log("✅ CRITICAL TEST PASSED: Schema changes propagate automatically!")
}

// TestMetadataManager_CompositePartitionKey verifies composite partition key detection
func TestMetadataManager_CompositePartitionKey(t *testing.T) {
	session, clusterConfig := setupTestSession(t)
	defer session.Close()

	createTestKeyspace(t, session)
	defer cleanupTestKeyspace(session)

	// Create table with composite partition key
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE %s.composite_pk (
			user_id int,
			device_id int,
			timestamp bigint,
			data text,
			PRIMARY KEY ((user_id, device_id), timestamp)
		)
	`, testKeyspace)
	err := session.Query(createTableQuery).Exec()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	manager := clustermd.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

	tableMeta, err := manager.GetTable(testKeyspace, "composite_pk")
	require.NoError(t, err)
	require.NotNil(t, tableMeta)

	// Verify partition keys (order matters!)
	assert.Len(t, tableMeta.PartitionKeys, 2)
	assert.Equal(t, "user_id", tableMeta.PartitionKeys[0].Name)
	assert.Equal(t, "device_id", tableMeta.PartitionKeys[1].Name)

	// Verify clustering key
	assert.Len(t, tableMeta.ClusteringKeys, 1)
	assert.Equal(t, "timestamp", tableMeta.ClusteringKeys[0].Name)

	// Verify GetPartitionKeyNames preserves order
	pkNames := tableMeta.GetPartitionKeyNames()
	assert.Equal(t, []string{"user_id", "device_id"}, pkNames)
}

// TestMetadataManager_ClusteringKeysWithOrder verifies clustering column ordering
func TestMetadataManager_ClusteringKeysWithOrder(t *testing.T) {
	session, clusterConfig := setupTestSession(t)
	defer session.Close()

	createTestKeyspace(t, session)
	defer cleanupTestKeyspace(session)

	// Create table with multiple clustering columns and DESC order
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE %s.timeseries (
			sensor_id int,
			year int,
			month int,
			day int,
			value double,
			PRIMARY KEY (sensor_id, year, month, day)
		) WITH CLUSTERING ORDER BY (year DESC, month DESC, day DESC)
	`, testKeyspace)
	err := session.Query(createTableQuery).Exec()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	manager := clustermd.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

	tableMeta, err := manager.GetTable(testKeyspace, "timeseries")
	require.NoError(t, err)
	require.NotNil(t, tableMeta)

	// Verify clustering keys count and order
	assert.Len(t, tableMeta.ClusteringKeys, 3)
	assert.Equal(t, "year", tableMeta.ClusteringKeys[0].Name)
	assert.Equal(t, "month", tableMeta.ClusteringKeys[1].Name)
	assert.Equal(t, "day", tableMeta.ClusteringKeys[2].Name)

	// Verify ordering (all DESC)
	assert.Equal(t, clustermd.ColumnOrderDESC, tableMeta.ClusteringKeys[0].ClusteringOrder)
	assert.Equal(t, clustermd.ColumnOrderDESC, tableMeta.ClusteringKeys[1].ClusteringOrder)
	assert.Equal(t, clustermd.ColumnOrderDESC, tableMeta.ClusteringKeys[2].ClusteringOrder)

	// Verify helper methods
	ckNames := tableMeta.GetClusteringKeyNames()
	assert.Equal(t, []string{"year", "month", "day"}, ckNames)

	ckOrders := tableMeta.GetClusteringKeyOrders()
	assert.Equal(t, []clustermd.ColumnOrder{clustermd.ColumnOrderDESC, clustermd.ColumnOrderDESC, clustermd.ColumnOrderDESC}, ckOrders)
}

// TestMetadataManager_CreateDropTableDetection verifies CREATE/DROP detection
func TestMetadataManager_CreateDropTableDetection(t *testing.T) {
	session, clusterConfig := setupTestSession(t)
	defer session.Close()

	createTestKeyspace(t, session)
	defer cleanupTestKeyspace(session)

	manager := clustermd.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

	// Verify table doesn't exist
	exists, err := manager.TableExists(testKeyspace, "dynamic_table")
	require.NoError(t, err)
	assert.False(t, exists, "Table should not exist initially")

	// CREATE TABLE
	createQuery := fmt.Sprintf(`
		CREATE TABLE %s.dynamic_table (
			id int PRIMARY KEY,
			data text
		)
	`, testKeyspace)
	err = session.Query(createQuery).Exec()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Verify table now exists
	exists, err = manager.TableExists(testKeyspace, "dynamic_table")
	require.NoError(t, err)
	assert.True(t, exists, "Table should exist after CREATE")

	// DROP TABLE
	dropQuery := fmt.Sprintf("DROP TABLE %s.dynamic_table", testKeyspace)
	err = session.Query(dropQuery).Exec()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Verify table no longer exists
	exists, err = manager.TableExists(testKeyspace, "dynamic_table")
	require.NoError(t, err)
	assert.False(t, exists, "Table should not exist after DROP")
}

// TestMetadataManager_TopologyRetrieval verifies cluster topology retrieval
func TestMetadataManager_TopologyRetrieval(t *testing.T) {
	session, clusterConfig := setupTestSession(t)
	defer session.Close()

	manager := clustermd.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

	t.Run("GetClusterName returns name", func(t *testing.T) {
		name, err := manager.GetClusterName()
		require.NoError(t, err)
		assert.NotEmpty(t, name)
		t.Logf("Cluster name: %s", name)
	})

	t.Run("GetPartitioner returns partitioner", func(t *testing.T) {
		partitioner, err := manager.GetPartitioner()
		require.NoError(t, err)
		assert.NotEmpty(t, partitioner)
		assert.Contains(t, partitioner, "Murmur3Partitioner")
		t.Logf("Partitioner: %s", partitioner)
	})

	t.Run("GetTopology returns topology", func(t *testing.T) {
		topology, err := manager.GetTopology()
		require.NoError(t, err)
		require.NotNil(t, topology)

		assert.NotEmpty(t, topology.ClusterName)
		assert.NotEmpty(t, topology.Partitioner)
		assert.NotEmpty(t, topology.Hosts)
		assert.Len(t, topology.SchemaVersions, 1) // Single node, single schema version

		// Verify at least one host (local node)
		assert.GreaterOrEqual(t, len(topology.Hosts), 1)
		assert.GreaterOrEqual(t, len(topology.DataCenters), 1)

		t.Logf("Cluster: %s, Hosts: %d, DCs: %d",
			topology.ClusterName, len(topology.Hosts), len(topology.DataCenters))
	})

	t.Run("GetUpNodes returns nodes", func(t *testing.T) {
		nodes, err := manager.GetUpNodes()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(nodes), 1)

		for _, node := range nodes {
			assert.NotEmpty(t, node.HostID)
			assert.NotEmpty(t, node.DataCenter)
			assert.NotEmpty(t, node.ReleaseVersion)
			assert.Equal(t, clustermd.HostStateUp, node.State)
		}
	})
}

// TestMetadataManager_KeyspaceList verifies keyspace listing
func TestMetadataManager_KeyspaceList(t *testing.T) {
	session, clusterConfig := setupTestSession(t)
	defer session.Close()

	manager := clustermd.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

	keyspaces, err := manager.GetKeyspaceNames()
	require.NoError(t, err)
	assert.NotEmpty(t, keyspaces)

	// Should include system keyspaces
	hasSystem := false
	for _, ks := range keyspaces {
		if ks == "system" {
			hasSystem = true
			break
		}
	}
	assert.True(t, hasSystem, "Should include 'system' keyspace")

	t.Logf("Found %d keyspaces", len(keyspaces))
}

// TestMetadataManager_SchemaAgreement verifies schema agreement waiting
func TestMetadataManager_SchemaAgreement(t *testing.T) {
	session, clusterConfig := setupTestSession(t)
	defer session.Close()

	manager := clustermd.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := manager.WaitForSchemaAgreement(ctx)
	assert.NoError(t, err, "Schema agreement should succeed")
}

// TestMetadataManager_GetSchemaVersion verifies schema version retrieval
func TestMetadataManager_GetSchemaVersion(t *testing.T) {
	session, clusterConfig := setupTestSession(t)
	defer session.Close()

	manager := clustermd.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

	versions, err := manager.GetSchemaVersion()
	require.NoError(t, err)
	assert.Len(t, versions, 1) // Single node should have single schema version
	assert.NotEmpty(t, versions[0])

	t.Logf("Schema version: %s", versions[0])
}

// TestMetadataManager_RefreshKeyspace verifies RefreshKeyspace waits for schema agreement
func TestMetadataManager_RefreshKeyspace(t *testing.T) {
	session, clusterConfig := setupTestSession(t)
	defer session.Close()

	createTestKeyspace(t, session)
	defer cleanupTestKeyspace(session)

	manager := clustermd.NewGocqlMetadataManagerWithDefaults(session, clusterConfig)

	// Create a table
	createQuery := fmt.Sprintf("CREATE TABLE %s.refresh_test (id int PRIMARY KEY, data text)", testKeyspace)
	err := session.Query(createQuery).Exec()
	require.NoError(t, err)

	// Call RefreshKeyspace - should wait for schema agreement
	ctx := context.Background()
	err = manager.RefreshKeyspace(ctx, testKeyspace)
	require.NoError(t, err, "RefreshKeyspace should succeed")

	// Verify metadata is available (gocql auto-refreshed)
	tableMeta, err := manager.GetTable(testKeyspace, "refresh_test")
	require.NoError(t, err)
	require.NotNil(t, tableMeta)
	assert.Equal(t, "refresh_test", tableMeta.Name)

	t.Log("✅ RefreshKeyspace successfully waited for schema agreement")
}
