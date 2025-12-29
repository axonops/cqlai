// +build integration

package mcp_test

import (
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to get test cluster config
func getTestCluster(t *testing.T) *gocql.ClusterConfig {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cluster := gocql.NewCluster("127.0.0.1:9042")
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "cassandra",
		Password: "cassandra",
	}
	cluster.Timeout = 5 * time.Second
	cluster.ConnectTimeout = 5 * time.Second
	cluster.Consistency = gocql.LocalOne

	return cluster
}

// TestNewSessionFromCluster verifies that a new session can be created from a cluster config
func TestNewSessionFromCluster(t *testing.T) {
	cluster := getTestCluster(t)

	// Create first session using the public API
	s1, err := db.NewSessionFromCluster(cluster, "cassandra", false)
	require.NoError(t, err, "Failed to connect to Cassandra - is it running on 127.0.0.1:9042?")
	defer s1.Close()

	// Create second session from the cluster using our helper
	s2, err := db.NewSessionFromCluster(s1.GetCluster(), "cassandra", false)
	require.NoError(t, err, "Failed to create second session from cluster")
	defer s2.Close()

	// Verify both sessions are independent (different session objects)
	assert.NotEqual(t, s1, s2, "Expected different session pointers")
	assert.NotEqual(t, s1.Session, s2.Session, "Expected different gocql.Session instances")

	// Verify both sessions can query independently
	var version1, version2 string

	iter1 := s1.Query("SELECT release_version FROM system.local").Iter()
	require.True(t, iter1.Scan(&version1), "Session 1 failed to scan version")
	require.NoError(t, iter1.Close(), "Session 1 close failed")

	iter2 := s2.Query("SELECT release_version FROM system.local").Iter()
	require.True(t, iter2.Scan(&version2), "Session 2 failed to scan version")
	require.NoError(t, iter2.Close(), "Session 2 close failed")

	assert.NotEmpty(t, version1, "Session 1 returned empty version")
	assert.NotEmpty(t, version2, "Session 2 returned empty version")
	assert.Equal(t, version1, version2, "Both sessions should query same Cassandra version")

	t.Logf("Both sessions successfully queried Cassandra version: %s", version1)
}

// TestIndependentSessions verifies that two sessions from the same cluster are isolated
func TestIndependentSessions(t *testing.T) {
	cluster := getTestCluster(t)

	// Create first session
	s1, err := db.NewSessionFromCluster(cluster, "cassandra", false)
	require.NoError(t, err, "Failed to connect to Cassandra - is it running on 127.0.0.1:9042?")
	defer s1.Close()

	// Create second session from same cluster
	s2, err := db.NewSessionFromCluster(cluster, "cassandra", false)
	require.NoError(t, err, "Failed to create second session")
	defer s2.Close()

	// Verify sessions are independent (different session objects)
	assert.NotEqual(t, s1, s2, "Expected different session pointers")
	assert.NotEqual(t, s1.Session, s2.Session, "Expected different gocql.Session instances")

	// Verify each has independent schema cache
	require.NotNil(t, s1.GetSchemaCache(), "Session 1 schema cache is nil")
	require.NotNil(t, s2.GetSchemaCache(), "Session 2 schema cache is nil")
	assert.NotEqual(t, s1.GetSchemaCache(), s2.GetSchemaCache(),
		"Expected different schema cache instances")

	// Verify initial consistency is the same for both (from cluster)
	initial1 := s1.Consistency()
	initial2 := s2.Consistency()
	assert.Equal(t, initial1, initial2, "Initial consistency should match")

	// Modify consistency on one session, verify other is unaffected
	err = s1.SetConsistency("QUORUM")
	require.NoError(t, err, "Failed to set consistency on session 1")

	after1 := s1.Consistency()
	after2 := s2.Consistency()

	assert.Equal(t, "QUORUM", after1, "Session 1 consistency not updated")
	assert.Equal(t, initial2, after2, "Session 2 consistency affected by session 1 change")

	// Modify page size on one session, verify other is unaffected
	s1.SetPageSize(500)
	assert.Equal(t, 500, s1.PageSize(), "Session 1 page size not updated")
	assert.NotEqual(t, 500, s2.PageSize(), "Session 2 page size affected by session 1 change")
}

// TestNewSessionFromClusterBatchMode verifies schema cache skipping in batch mode.
//
// Batch mode is when CQLAI runs non-interactively (e.g., cqlai -e "SELECT * FROM users" or cqlai -f script.cql).
// In batch mode, we skip schema cache initialization because:
// - No AI features needed (just executing CQL, no natural language queries)
// - Faster startup (schema cache queries all keyspaces/tables from system_schema)
// - Simpler operation (scripts don't need schema introspection)
//
// MCP will use batchMode=false (needs schema for AI tools), but we verify both paths work.
func TestNewSessionFromClusterBatchMode(t *testing.T) {
	cluster := getTestCluster(t)

	// Create session in batch mode (batchMode=true skips schema cache initialization)
	s, err := db.NewSessionFromCluster(cluster, "cassandra", true)
	require.NoError(t, err, "Failed to connect to Cassandra - is it running on 127.0.0.1:9042?")
	defer s.Close()

	// Verify schema cache was not initialized (nil when batchMode=true)
	assert.Nil(t, s.GetSchemaCache(), "Expected nil schema cache in batch mode")
}

// TestSessionSharesAuth verifies that authentication is preserved across sessions
func TestSessionSharesAuth(t *testing.T) {
	cluster := getTestCluster(t)

	// Create first session
	s1, err := db.NewSessionFromCluster(cluster, "cassandra", false)
	require.NoError(t, err, "Failed to connect to Cassandra - is it running on 127.0.0.1:9042?")
	defer s1.Close()

	// Get cluster for second session
	retrievedCluster := s1.GetCluster()
	require.NotNil(t, retrievedCluster, "GetCluster() returned nil")

	// Verify auth is preserved
	auth, ok := retrievedCluster.Authenticator.(gocql.PasswordAuthenticator)
	require.True(t, ok, "Authenticator not a PasswordAuthenticator")
	assert.Equal(t, "cassandra", auth.Username, "Username not preserved")
	assert.Equal(t, "cassandra", auth.Password, "Password not preserved")

	// Create second session and verify it can connect (auth works)
	s2, err := db.NewSessionFromCluster(retrievedCluster, "cassandra", false)
	require.NoError(t, err, "Failed to create second session with preserved auth")
	defer s2.Close()

	// Verify second session can query (proof auth worked)
	var version string
	iter := s2.Query("SELECT release_version FROM system.local").Iter()
	require.True(t, iter.Scan(&version), "Failed to query with second session")
	require.NoError(t, iter.Close())
	assert.NotEmpty(t, version, "Version query returned empty result")
}
