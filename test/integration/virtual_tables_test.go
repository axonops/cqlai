//go:build integration
// +build integration

package integration_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Virtual tables.
//
// Cassandra 4.0 added tables the node computes rather than stores -
// system_views.clients, .settings, .thread_pools and forty-odd more. They are
// listed in system_virtual_schema, not system_schema, and every question cqlai
// asked about what exists was asked of system_schema. A SELECT from
// system_views.clients worked, and USE system_views said the keyspace did not
// exist.
//
// These run against every Cassandra in the CI matrix. From 4.0 the virtual
// tables have to appear everywhere a keyspace or a table is listed. Before 4.0
// there is no system_virtual_schema to ask, and the answer has to be "none"
// rather than an error.

// TestVirtualKeyspacesAreListed.
func TestVirtualKeyspacesAreListed(t *testing.T) {
	session, _, cleanup := getTestSession(t)
	defer cleanup()

	virtual := session.VirtualKeyspaces()
	if !session.IsVersion4OrHigher() {
		assert.Empty(t, virtual, "before 4.0 there are none, and asking is not an error")
		return
	}

	require.Contains(t, virtual, "system_views")
	require.Contains(t, virtual, "system_virtual_schema")

	// DESCRIBE KEYSPACES, the schema browser and the keyspace chooser all read
	// this one list.
	listed, err := session.DescribeKeyspacesQuery()
	require.NoError(t, err)

	var names []string
	for _, ks := range listed {
		names = append(names, ks.Name)
		if ks.Name == "system_views" {
			assert.True(t, ks.Virtual, "and it says which ones are virtual")
		}
	}
	assert.Contains(t, names, "system_views", "DESCRIBE KEYSPACES should list it")
	assert.Contains(t, names, "system_schema", "alongside the stored ones")
}

// TestAVirtualKeyspaceCanBeUsed: USE refused it as a keyspace that did not
// exist, while a SELECT from one of its tables worked.
func TestAVirtualKeyspaceCanBeUsed(t *testing.T) {
	session, _, cleanup := getTestSession(t)
	defer cleanup()

	if !session.IsVersion4OrHigher() {
		t.Skip("virtual keyspaces arrived in Cassandra 4.0")
	}

	result := session.ExecuteCQLQuery("USE system_views")

	if err, isErr := result.(error); isErr {
		t.Fatalf("USE system_views failed: %v", err)
	}
	assert.Contains(t, result, "system_views")
}

// TestAVirtualKeyspacesTablesAreListedWithTheirKeys.
//
// The key matters more for a virtual table than a stored one: it can only be
// filtered on its key, so the key is what says how to ask it anything.
func TestAVirtualKeyspacesTablesAreListedWithTheirKeys(t *testing.T) {
	session, _, cleanup := getTestSession(t)
	defer cleanup()

	if !session.IsVersion4OrHigher() {
		assert.Empty(t, session.VirtualTables("system_views"))
		return
	}

	tables, err := session.DescribeTablesQuery("system_views")
	require.NoError(t, err)
	require.NotEmpty(t, tables, "system_views has tables, and asking system_schema found none")

	var clients bool
	for _, table := range tables {
		assert.True(t, table.Virtual, "%s is virtual", table.Name)
		if table.Name == "clients" {
			clients = true
			assert.Equal(t, []string{"address"}, table.PartitionKeys)
			assert.Equal(t, []string{"port"}, table.ClusteringKeys)
		}
	}
	assert.True(t, clients, "system_views.clients exists on every 4.0+ node")
}

// TestAVirtualKeyspaceIsDescribed. Built from system_schema it was described as
// nothing at all; only the server knows it.
func TestAVirtualKeyspaceIsDescribed(t *testing.T) {
	session, _, cleanup := getTestSession(t)
	defer cleanup()

	if !session.IsVersion4OrHigher() {
		t.Skip("virtual keyspaces arrived in Cassandra 4.0")
	}

	described, err := session.DBDescribeFullSchema(nil, "system_views")
	require.NoError(t, err)

	text, ok := described.(string)
	require.True(t, ok, "a description is text")
	assert.Contains(t, text, "virtual keyspace", "the warning cqlsh shows")
	assert.Contains(t, text, "VIRTUAL TABLE system_views.clients", "with its tables")
}

// TestAVirtualTableIsDescribed, which is what the schema browser shows when
// one is picked. It built the definition from system_schema and said the table
// it had just listed was not found.
func TestAVirtualTableIsDescribed(t *testing.T) {
	session, _, cleanup := getTestSession(t)
	defer cleanup()

	if !session.IsVersion4OrHigher() {
		t.Skip("virtual tables arrived in Cassandra 4.0")
	}

	text, err := session.DescribeVirtualTable("system_views", "clients")
	require.NoError(t, err)
	assert.Contains(t, text, "VIRTUAL TABLE system_views.clients")
	assert.Contains(t, text, "address", "with its columns")
}

// TestTheReplayableSchemaStillLeavesThemOut.
//
// DESCRIBE SCHEMA is meant to be replayed against another cluster, and a
// virtual table cannot be created with CQL - so this is the one listing that
// must not gain them. cqlsh leaves them out for the same reason.
func TestTheReplayableSchemaStillLeavesThemOut(t *testing.T) {
	session, _, cleanup := getTestSession(t)
	defer cleanup()

	described, err := session.DBDescribeFullSchema(nil, "")
	require.NoError(t, err)

	text, _ := described.(string)
	assert.False(t, strings.Contains(text, "CREATE TABLE system_views."),
		"DESCRIBE SCHEMA must stay replayable")
}

// TestVirtualColumnsAreKnown, for completion: SELECT * FROM
// system_views.clients WHERE <Tab> offered nothing.
func TestVirtualColumnsAreKnown(t *testing.T) {
	session, _, cleanup := getTestSession(t)
	defer cleanup()

	columns := session.VirtualColumns("system_views", "clients")
	if !session.IsVersion4OrHigher() {
		assert.Empty(t, columns)
		return
	}

	assert.Contains(t, columns, "address")
	assert.Contains(t, columns, "port")
}
