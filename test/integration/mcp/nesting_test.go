//go:build integration
// +build integration

package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMCP_NestedCollections tests collections of collections
func TestMCP_NestedCollections_ListOfLists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	// Ensure nesting_test keyspace and types exist
	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query(`CREATE TABLE IF NOT EXISTS nesting_test.test_nested (
		id int PRIMARY KEY,
		list_of_lists frozen<list<list<int>>>,
		map_of_lists frozen<map<text, list<int>>>,
		set_of_sets frozen<set<set<int>>>
	)`).Exec()

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "test_nested",
		"values": map[string]any{
			"id":            2000,
			"list_of_lists": [][]int{{1, 2}, {3, 4}, {5, 6}},
		},
		"value_types": map[string]any{
			"list_of_lists": "list<list<int>>",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with list<list<int>> should succeed")

	// Verify in Cassandra
	var id int
	var listOfLists [][]int
	iter := ctx.Session.Query("SELECT id, list_of_lists FROM nesting_test.test_nested WHERE id = 2000").Iter()
	if assert.True(t, iter.Scan(&id, &listOfLists), "Should retrieve nested list") {
		assert.Equal(t, 2000, id)
		assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5, 6}}, listOfLists)
	}
	iter.Close()

	t.Log("✅ list<list<int>> via MCP")
}

// TestMCP_NestedCollections_MapOfLists tests map<text, list<int>>
func TestMCP_NestedCollections_MapOfLists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query(`CREATE TABLE IF NOT EXISTS nesting_test.test_nested (
		id int PRIMARY KEY,
		map_of_lists frozen<map<text, list<int>>>
	)`).Exec()

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "test_nested",
		"values": map[string]any{
			"id": 2001,
			"map_of_lists": map[string]any{
				"group1": []int{1, 2, 3},
				"group2": []int{4, 5, 6},
			},
		},
		"value_types": map[string]any{
			"map_of_lists": "map<text,list<int>>",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with map<text, list<int>> should succeed")

	t.Log("✅ map<text, list<int>> via MCP")
}

// TestMCP_NestedUDTs tests UDTs containing other UDTs
func TestMCP_NestedUDTs(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query("CREATE TYPE IF NOT EXISTS nesting_test.address (street text, city text, zip text)").Exec()
	ctx.Session.Query("CREATE TYPE IF NOT EXISTS nesting_test.person (name text, addr frozen<address>)").Exec()
	ctx.Session.Query("CREATE TABLE IF NOT EXISTS nesting_test.people (id int PRIMARY KEY, info frozen<person>)").Exec()

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "people",
		"values": map[string]any{
			"id": 2002,
			"info": map[string]any{
				"name": "Alice",
				"addr": map[string]any{
					"street": "123 Main St",
					"city":   "NYC",
					"zip":    "10001",
				},
			},
		},
		"value_types": map[string]any{
			"info":      "frozen<person>",
			"info.addr": "frozen<address>", // Type hint for nested UDT
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with nested UDT should succeed")

	t.Log("✅ Nested UDT via MCP")
}

// TestMCP_ListOfUDTs tests list<udt_type>
func TestMCP_ListOfUDTs(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query("CREATE TYPE IF NOT EXISTS nesting_test.address (street text, city text, zip text)").Exec()
	ctx.Session.Query("CREATE TABLE IF NOT EXISTS nesting_test.addresses (id int PRIMARY KEY, addrs frozen<list<address>>)").Exec()

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "addresses",
		"values": map[string]any{
			"id": 2003,
			"addrs": []map[string]any{
				{
					"street": "123 Main St",
					"city":   "NYC",
					"zip":    "10001",
				},
				{
					"street": "456 Oak Ave",
					"city":   "SF",
					"zip":    "94102",
				},
			},
		},
		"value_types": map[string]any{
			"addrs": "list<frozen<address>>",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with list<udt> should succeed")

	t.Log("✅ list<frozen<address>> via MCP")
}

// TestMCP_UDT_WithCollections tests UDT containing list, set, and map
func TestMCP_UDT_WithCollections(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query("CREATE TYPE IF NOT EXISTS nesting_test.contact (name text, phones list<text>, emails set<text>, metadata map<text, text>)").Exec()
	ctx.Session.Query("CREATE TABLE IF NOT EXISTS nesting_test.contacts_udt (id int PRIMARY KEY, info frozen<contact>)").Exec()

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "contacts_udt",
		"values": map[string]any{
			"id": 2004,
			"info": map[string]any{
				"name":   "Charlie",
				"phones": []string{"555-1111", "555-2222", "555-3333"},
				"emails": []string{"charlie@example.com", "chuck@example.com"},
				"metadata": map[string]any{
					"dept":   "sales",
					"region": "west",
					"level":  "senior",
				},
			},
		},
		"value_types": map[string]any{
			"info":          "frozen<contact>",
			"info.phones":   "list<text>",
			"info.emails":   "set<text>",
			"info.metadata": "map<text,text>",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with UDT containing collections should succeed")

	t.Log("✅ UDT with list/set/map via MCP")
}

// TestMCP_MapOfMaps tests nested map<text, map<text, int>>
func TestMCP_MapOfMaps(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query("CREATE TABLE IF NOT EXISTS nesting_test.nested_maps_table (id int PRIMARY KEY, data frozen<map<text, map<text, int>>>)").Exec()

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "nested_maps_table",
		"values": map[string]any{
			"id": 2005,
			"data": map[string]any{
				"group1": map[string]any{"a": 1, "b": 2},
				"group2": map[string]any{"x": 10, "y": 20},
			},
		},
		"value_types": map[string]any{
			"data": "map<text,map<text,int>>",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with map<text, map<text,int>> should succeed")

	t.Log("✅ map<text, map<text,int>> via MCP")
}

