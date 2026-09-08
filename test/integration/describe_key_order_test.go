//go:build integration
// +build integration

package integration_test

import (
	"testing"

	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDescribeKeyOrderFollowsSchemaPosition covers #84.
//
// DescribeTableQuery is the pre-4.0 fallback: on Cassandra 4.0 and later
// DBDescribeTable asks the server for the CREATE statement instead, which is
// why this only ever showed up for people on 3.x. Calling it directly exercises
// the fallback whatever the server version.
//
// system_schema.columns is clustered by column_name, so it comes back
// alphabetically. Taking the key columns in that order rewrites the primary
// key: a table declared ((c, a, b), z, y, x) was described as
// ((a, b, c), x, y, z), which is a different table. The partition key order
// decides how rows are distributed, and the clustering order decides how they
// are sorted.
func TestDescribeKeyOrderFollowsSchemaPosition(t *testing.T) {
	dbSession, _, cleanup := getTestSession(t)
	defer cleanup()

	// Declared deliberately so that alphabetical order differs from schema
	// position for both halves of the key.
	require.NoError(t, dbSession.Query(`CREATE TABLE IF NOT EXISTS key_order (
		a int, b int, c int,
		x int, y int, z int,
		v text,
		PRIMARY KEY ((c, a, b), z, y, x)
	)`).Exec())

	info, err := dbSession.DescribeTableQuery("test_roundtrip", "key_order")
	require.NoError(t, err)

	assert.Equal(t, []string{"c", "a", "b"}, info.PartitionKeys,
		"partition keys must follow schema position, not the alphabet")
	assert.Equal(t, []string{"z", "y", "x"}, info.ClusteringKeys,
		"clustering keys must follow schema position, not the alphabet")

	ddl := db.FormatTableCreateStatement(info, false)
	assert.Contains(t, ddl, "PRIMARY KEY ((c, a, b), z, y, x)",
		"the generated DDL must recreate the same table")
	assert.NotContains(t, ddl, "PRIMARY KEY ((a, b, c), x, y, z)")
}
