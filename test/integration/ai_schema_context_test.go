//go:build integration
// +build integration

package integration_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchemaContextNamesTables covers #101.
//
// GetSchemaContext is what the AI is handed as its picture of the database
// (internal/ui/ai_commands.go and internal/ai/client.go). It walks keyspaces and
// loads each table's columns. Loading a table used to fail for every table,
// because the query asked Cassandra to ORDER BY a column it cannot sort by, and
// loadTablesForKeyspace discarded the error and moved on. The result was a
// context listing keyspace names and nothing else - an AI asked to write CQL
// against a database it believes has no tables will invent them.
func TestSchemaContextNamesTables(t *testing.T) {
	dbSession, _, cleanup := getTestSession(t)
	defer cleanup()

	require.NoError(t, dbSession.Query(`CREATE TABLE IF NOT EXISTS ai_ctx (
		a int, b int,
		z int, y int,
		note text,
		PRIMARY KEY ((b, a), z, y)
	)`).Exec())

	ctx, err := dbSession.GetSchemaContext(50)
	require.NoError(t, err)
	require.NotEmpty(t, ctx)

	assert.Contains(t, ctx, "Keyspace: test_roundtrip")
	assert.Contains(t, ctx, "Table: ai_ctx",
		"the AI must be told the table exists, not just the keyspace")

	for _, col := range []string{"a", "b", "z", "y", "note"} {
		assert.Contains(t, ctx, "- "+col+":",
			"column %q is missing from the schema the AI is given", col)
	}

	assert.Contains(t, ctx, "(PK)", "partition keys must be marked")
	assert.Contains(t, ctx, "(CK)", "clustering keys must be marked")
}

// TestTableSchemaKeyOrder guards the same trap that #84 was about. The columns
// come back from system_schema.columns alphabetically, so the key order has to
// be taken after sorting by position, or the AI is told this table is
// partitioned by (a, b) when it is partitioned by (b, a).
func TestTableSchemaKeyOrder(t *testing.T) {
	dbSession, _, cleanup := getTestSession(t)
	defer cleanup()

	require.NoError(t, dbSession.Query(`CREATE TABLE IF NOT EXISTS ai_key_order (
		a int, b int, c int,
		x int, y int, z int,
		v text,
		PRIMARY KEY ((c, a, b), z, y, x)
	)`).Exec())

	ts, err := dbSession.GetTableSchema("test_roundtrip", "ai_key_order")
	require.NoError(t, err)

	assert.Equal(t, []string{"c", "a", "b"}, ts.PartitionKeys,
		"partition keys must follow schema position, not the alphabet")
	assert.Equal(t, []string{"z", "y", "x"}, ts.ClusteringKeys,
		"clustering keys must follow schema position, not the alphabet")

	// Columns are ordered keys-first so the AI reads the primary key correctly.
	var names []string
	for _, c := range ts.Columns {
		names = append(names, c.Name)
	}
	assert.Equal(t, []string{"c", "a", "b", "z", "y", "x", "v"}, names,
		"columns should be ordered partition keys, clustering keys, then the rest")
}

// TestSchemaContextIsNotJustKeyspaces is the symptom in the issue, stated as a
// measurement rather than a list of assertions.
func TestSchemaContextIsNotJustKeyspaces(t *testing.T) {
	dbSession, _, cleanup := getTestSession(t)
	defer cleanup()

	require.NoError(t, dbSession.Query(
		`CREATE TABLE IF NOT EXISTS ai_present (id int PRIMARY KEY, v text)`).Exec())

	ctx, err := dbSession.GetSchemaContext(50)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(ctx), "\n")
	var keyspaceLines, otherLines int
	for _, l := range lines {
		if strings.HasPrefix(l, "Keyspace:") {
			keyspaceLines++
		} else if strings.TrimSpace(l) != "" {
			otherLines++
		}
	}

	assert.Greater(t, otherLines, 0,
		"the schema context was %d keyspace lines and nothing else", keyspaceLines)
}
