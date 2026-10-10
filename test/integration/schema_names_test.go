//go:build integration
// +build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAllSchemaNamesIsWhatEachKeyspaceHolds: the filter reads every
// keyspace's names in one query a kind, and has to find what opening each
// keyspace finds.
func TestAllSchemaNamesIsWhatEachKeyspaceHolds(t *testing.T) {
	sess, _, cleanup := getTestSession(t)
	defer cleanup()

	tables, objects := sess.AllSchemaNames()
	require.NotEmpty(t, tables)

	iter := sess.Query(`SELECT keyspace_name FROM system_schema.keyspaces`).Iter()
	var keyspace string
	checked := 0
	for iter.Scan(&keyspace) {
		listed, err := sess.DescribeTablesQuery(keyspace)
		require.NoError(t, err, keyspace)
		var names []string
		for _, table := range listed {
			names = append(names, table.Name)
		}
		assert.ElementsMatch(t, names, tables[keyspace], "tables of %s", keyspace)

		one := sess.KeyspaceObjects(keyspace)
		for kind, want := range one {
			assert.Equal(t, want, objects[keyspace][kind], "%s of %s", kind, keyspace)
		}
		for kind, got := range objects[keyspace] {
			assert.Equal(t, one[kind], got, "%s of %s", kind, keyspace)
		}
		checked++
	}
	require.NoError(t, iter.Close())
	assert.Greater(t, checked, 3)
}
