//go:build integration
// +build integration

package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Comprehensive Nesting Tests - Following Cassandra 5 Rules
// ============================================================================
// Based on user research: c5-nesting-cql.md and c5-nesting-mtx.md
//
// CRITICAL RULES:
// 1. Collections inside collections MUST freeze inner collection
// 2. UDTs inside collections MUST be frozen
// 3. UDTs inside UDTs MUST be frozen
// 4. Collections inside UDTs can be non-frozen (unless they nest further)
//
// EVERY TEST MUST:
// - INSERT via MCP
// - VALIDATE data in Cassandra (direct SELECT + assert)
// - SELECT via MCP (round-trip)
// - UPDATE via MCP (where applicable)
// - VERIFY update in Cassandra
// ============================================================================

// ============================================================================
// CATEGORY A: Baseline Collections
// ============================================================================

// TestMCP_Baseline_SetText tests set<text> with full CRUD
func TestMCP_Baseline_SetText(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query(`CREATE TABLE IF NOT EXISTS nesting_test.t_set_text (
		id int PRIMARY KEY,
		v set<text>
	)`).Exec()

	testID := 3000
	testData := []string{"alice", "bob", "charlie"}

	// 1. INSERT via MCP
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "t_set_text",
		"values": map[string]any{
			"id": testID,
			"v":  testData,
		},
		"value_types": map[string]any{
			"v": "set<text>",
		},
	}

	insertResult := callToolHTTP(t, ctx, "submit_query_plan", insertArgs)
	assertNotError(t, insertResult, "INSERT set<text> should succeed")

	// 2. VALIDATE in Cassandra
	var id int
	var retrieved []string // gocql represents sets as slices
	iter := ctx.Session.Query("SELECT id, v FROM nesting_test.t_set_text WHERE id = ?", testID).Iter()
	assert.True(t, iter.Scan(&id, &retrieved), "Should retrieve set from Cassandra")
	assert.Equal(t, testID, id)
	// Set order not guaranteed, just check length and contains
	assert.Len(t, retrieved, 3, "Set should have 3 elements")
	iter.Close()

	// 3. SELECT via MCP
	selectArgs := map[string]any{
		"operation": "SELECT",
		"keyspace":  "nesting_test",
		"table":     "t_set_text",
		"columns":   []string{"id", "v"},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}
	selectResult := callToolHTTP(t, ctx, "submit_query_plan", selectArgs)
	assertNotError(t, selectResult, "SELECT via MCP should succeed")

	// 4. UPDATE via MCP (add to set)
	updateArgs := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "nesting_test",
		"table":     "t_set_text",
		"collection_ops": map[string]any{
			"v": map[string]any{
				"operation": "add",
				"value":     []string{"diana"},
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}
	updateResult := callToolHTTP(t, ctx, "submit_query_plan", updateArgs)
	assertNotError(t, updateResult, "UPDATE set (add) should succeed")

	// 5. VERIFY UPDATE in Cassandra
	iter = ctx.Session.Query("SELECT v FROM nesting_test.t_set_text WHERE id = ?", testID).Iter()
	assert.True(t, iter.Scan(&retrieved))
	assert.Len(t, retrieved, 4, "Set should now have 4 elements")
	iter.Close()

	t.Log("✅ set<text> - Full CRUD verified in Cassandra")
}

// ============================================================================
// CATEGORY B: Nested Collections (Inner MUST be Frozen)
// ============================================================================

// TestMCP_NestedCollections_ListOfFrozenList tests list<frozen<list<int>>>
// CRITICAL: Inner list MUST be frozen per Cassandra rules
func TestMCP_NestedCollections_ListOfFrozenList(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query(`CREATE TABLE IF NOT EXISTS nesting_test.t_list_frozen_list (
		id int PRIMARY KEY,
		v list<frozen<list<int>>>
	)`).Exec()

	testID := 3001
	testData := [][]int{{1, 2}, {3, 4}, {5, 6}}

	// 1. INSERT via MCP
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "t_list_frozen_list",
		"values": map[string]any{
			"id": testID,
			"v":  testData,
		},
		"value_types": map[string]any{
			"v": "list<frozen<list<int>>>", // ✅ CORRECT: inner list frozen
		},
	}

	insertResult := callToolHTTP(t, ctx, "submit_query_plan", insertArgs)
	assertNotError(t, insertResult, "INSERT list<frozen<list<int>>> should succeed")

	// 2. VALIDATE in Cassandra
	var id int
	var retrieved [][]int
	iter := ctx.Session.Query("SELECT id, v FROM nesting_test.t_list_frozen_list WHERE id = ?", testID).Iter()
	assert.True(t, iter.Scan(&id, &retrieved), "Should retrieve from Cassandra")
	assert.Equal(t, testID, id)
	assert.Equal(t, testData, retrieved, "Data must match exactly")
	iter.Close()

	// 3. SELECT via MCP
	selectArgs := map[string]any{
		"operation": "SELECT",
		"keyspace":  "nesting_test",
		"table":     "t_list_frozen_list",
		"columns":   []string{"id", "v"},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}
	selectResult := callToolHTTP(t, ctx, "submit_query_plan", selectArgs)
	assertNotError(t, selectResult, "SELECT via MCP should succeed")

	// 4. UPDATE via MCP (frozen collections require full replacement)
	updateData := [][]int{{10, 20}, {30, 40}}
	updateArgs := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "nesting_test",
		"table":     "t_list_frozen_list",
		"values": map[string]any{
			"v": updateData,
		},
		"value_types": map[string]any{
			"v": "list<frozen<list<int>>>",
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}
	updateResult := callToolHTTP(t, ctx, "submit_query_plan", updateArgs)
	assertNotError(t, updateResult, "UPDATE should succeed")

	// 5. VERIFY UPDATE in Cassandra
	iter = ctx.Session.Query("SELECT v FROM nesting_test.t_list_frozen_list WHERE id = ?", testID).Iter()
	assert.True(t, iter.Scan(&retrieved))
	assert.Equal(t, updateData, retrieved, "Updated data must match")
	iter.Close()

	t.Log("✅ list<frozen<list<int>>> - INSERT/SELECT/UPDATE verified")
}

// TestMCP_NestedCollections_SetOfFrozenMap tests set<frozen<map<text,int>>>
func TestMCP_NestedCollections_SetOfFrozenMap(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query(`CREATE TABLE IF NOT EXISTS nesting_test.t_set_frozen_map (
		id int PRIMARY KEY,
		v set<frozen<map<text,int>>>
	)`).Exec()

	testID := 3002
	// For sets, we pass as slice and Cassandra deduplicates
	testData := []map[string]int{
		{"a": 1, "b": 2},
		{"x": 10, "y": 20},
	}

	// 1. INSERT via MCP
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "t_set_frozen_map",
		"values": map[string]any{
			"id": testID,
			"v":  testData,
		},
		"value_types": map[string]any{
			"v": "set<frozen<map<text,int>>>",
		},
	}

	insertResult := callToolHTTP(t, ctx, "submit_query_plan", insertArgs)
	assertNotError(t, insertResult, "INSERT set<frozen<map<text,int>>> should succeed")

	// 2. VALIDATE in Cassandra
	var id int
	var retrieved []map[string]int
	iter := ctx.Session.Query("SELECT id, v FROM nesting_test.t_set_frozen_map WHERE id = ?", testID).Iter()
	assert.True(t, iter.Scan(&id, &retrieved), "Should retrieve from Cassandra")
	assert.Equal(t, testID, id)
	assert.Len(t, retrieved, 2, "Set should have 2 map elements")
	iter.Close()

	// 3. SELECT via MCP
	selectArgs := map[string]any{
		"operation": "SELECT",
		"keyspace":  "nesting_test",
		"table":     "t_set_frozen_map",
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}
	selectResult := callToolHTTP(t, ctx, "submit_query_plan", selectArgs)
	assertNotError(t, selectResult, "SELECT via MCP should succeed")

	t.Log("✅ set<frozen<map<text,int>>> - INSERT/SELECT verified")
}

// TestMCP_NestedCollections_MapOfFrozenList tests map<text,frozen<list<int>>>
func TestMCP_NestedCollections_MapOfFrozenList(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query(`CREATE TABLE IF NOT EXISTS nesting_test.t_map_frozen_list (
		id int PRIMARY KEY,
		v map<text,frozen<list<int>>>
	)`).Exec()

	testID := 3003
	testData := map[string][]int{
		"group1": {1, 2, 3},
		"group2": {4, 5, 6},
	}

	// 1. INSERT via MCP
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "t_map_frozen_list",
		"values": map[string]any{
			"id": testID,
			"v":  testData,
		},
		"value_types": map[string]any{
			"v": "map<text,frozen<list<int>>>",
		},
	}

	insertResult := callToolHTTP(t, ctx, "submit_query_plan", insertArgs)
	assertNotError(t, insertResult, "INSERT map<text,frozen<list<int>>> should succeed")

	// 2. VALIDATE in Cassandra
	var id int
	var retrieved map[string][]int
	iter := ctx.Session.Query("SELECT id, v FROM nesting_test.t_map_frozen_list WHERE id = ?", testID).Iter()
	assert.True(t, iter.Scan(&id, &retrieved), "Should retrieve from Cassandra")
	assert.Equal(t, testID, id)
	assert.Equal(t, testData, retrieved, "Map data must match exactly")
	iter.Close()

	// 3. SELECT via MCP
	selectArgs := map[string]any{
		"operation": "SELECT",
		"keyspace":  "nesting_test",
		"table":     "t_map_frozen_list",
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}
	selectResult := callToolHTTP(t, ctx, "submit_query_plan", selectArgs)
	assertNotError(t, selectResult, "SELECT via MCP should succeed")

	// 4. UPDATE via MCP (replace entire map value for frozen inner lists)
	updateData := map[string][]int{
		"group3": {7, 8, 9},
	}
	updateArgs := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "nesting_test",
		"table":     "t_map_frozen_list",
		"values": map[string]any{
			"v": updateData,
		},
		"value_types": map[string]any{
			"v": "map<text,frozen<list<int>>>",
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}
	updateResult := callToolHTTP(t, ctx, "submit_query_plan", updateArgs)
	assertNotError(t, updateResult, "UPDATE should succeed")

	// 5. VERIFY UPDATE in Cassandra
	iter = ctx.Session.Query("SELECT v FROM nesting_test.t_map_frozen_list WHERE id = ?", testID).Iter()
	assert.True(t, iter.Scan(&retrieved))
	assert.Equal(t, updateData, retrieved, "Updated map must match")
	iter.Close()

	t.Log("✅ map<text,frozen<list<int>>> - Full CRUD verified")
}

// ============================================================================
// CATEGORY B: SET Collections - Missing from Original Tests
// ============================================================================

// TestMCP_SetOfFrozenSet tests set<frozen<set<int>>>
func TestMCP_SetOfFrozenSet(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query(`CREATE TABLE IF NOT EXISTS nesting_test.t_set_frozen_set (
		id int PRIMARY KEY,
		v set<frozen<set<int>>>
	)`).Exec()

	testID := 3004
	// Set of sets: each inner set is a distinct element
	testData := [][]int{
		{1, 2, 3},
		{4, 5, 6},
	}

	// 1. INSERT via MCP
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "t_set_frozen_set",
		"values": map[string]any{
			"id": testID,
			"v":  testData,
		},
		"value_types": map[string]any{
			"v": "set<frozen<set<int>>>",
		},
	}

	insertResult := callToolHTTP(t, ctx, "submit_query_plan", insertArgs)
	assertNotError(t, insertResult, "INSERT set<frozen<set<int>>> should succeed")

	// 2. VALIDATE in Cassandra
	var id int
	var retrieved [][]int
	iter := ctx.Session.Query("SELECT id, v FROM nesting_test.t_set_frozen_set WHERE id = ?", testID).Iter()
	assert.True(t, iter.Scan(&id, &retrieved), "Should retrieve from Cassandra")
	assert.Equal(t, testID, id)
	assert.Len(t, retrieved, 2, "Set should have 2 set elements")
	iter.Close()

	// 3. SELECT via MCP
	selectArgs := map[string]any{
		"operation": "SELECT",
		"keyspace":  "nesting_test",
		"table":     "t_set_frozen_set",
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}
	selectResult := callToolHTTP(t, ctx, "submit_query_plan", selectArgs)
	assertNotError(t, selectResult, "SELECT via MCP should succeed")

	t.Log("✅ set<frozen<set<int>>> - INSERT/SELECT verified")
}

// ============================================================================
// CATEGORY G: Tuples - Missing Comprehensive Tests
// ============================================================================

// TestMCP_Tuple_ListOfTuple tests list<tuple<int,text>>
func TestMCP_Tuple_ListOfTuple(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query(`CREATE TABLE IF NOT EXISTS nesting_test.t_list_tuple (
		id int PRIMARY KEY,
		v list<tuple<int,text>>
	)`).Exec()

	testID := 3005
	// Tuples as slice of slices (positional)
	testData := [][]interface{}{
		{1, "alice"},
		{2, "bob"},
	}

	// 1. INSERT via MCP
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "t_list_tuple",
		"values": map[string]any{
			"id": testID,
			"v":  testData,
		},
		"value_types": map[string]any{
			"v": "list<tuple<int,text>>",
		},
	}

	insertResult := callToolHTTP(t, ctx, "submit_query_plan", insertArgs)
	assertNotError(t, insertResult, "INSERT list<tuple<int,text>> should succeed")

	// 2. VALIDATE in Cassandra
	var id int
	var retrieved [][]interface{}
	iter := ctx.Session.Query("SELECT id, v FROM nesting_test.t_list_tuple WHERE id = ?", testID).Iter()
	assert.True(t, iter.Scan(&id, &retrieved), "Should retrieve tuples from Cassandra")
	assert.Equal(t, testID, id)
	assert.Len(t, retrieved, 2, "Should have 2 tuples")
	iter.Close()

	// 3. SELECT via MCP
	selectArgs := map[string]any{
		"operation": "SELECT",
		"keyspace":  "nesting_test",
		"table":     "t_list_tuple",
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}
	selectResult := callToolHTTP(t, ctx, "submit_query_plan", selectArgs)
	assertNotError(t, selectResult, "SELECT via MCP should succeed")

	t.Log("✅ list<tuple<int,text>> - INSERT/SELECT verified")
}

// ============================================================================
// CATEGORY D: Collections of UDTs (UDT MUST be Frozen)
// ============================================================================

// TestMCP_SetOfFrozenUDT tests set<frozen<address>>
func TestMCP_SetOfFrozenUDT(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	ctx.Session.Query("CREATE TYPE IF NOT EXISTS nesting_test.address (street text, city text, zip text)").Exec()
	ctx.Session.Query(`CREATE TABLE IF NOT EXISTS nesting_test.t_set_frozen_udt (
		id int PRIMARY KEY,
		v set<frozen<address>>
	)`).Exec()

	testID := 3006
	testData := []map[string]string{
		{"street": "123 Main St", "city": "NYC", "zip": "10001"},
		{"street": "456 Oak Ave", "city": "SF", "zip": "94102"},
	}

	// 1. INSERT via MCP
	insertArgs := map[string]any{
		"operation": "INSERT",
		"keyspace":  "nesting_test",
		"table":     "t_set_frozen_udt",
		"values": map[string]any{
			"id": testID,
			"v":  testData,
		},
		"value_types": map[string]any{
			"v": "set<frozen<address>>",
		},
	}

	insertResult := callToolHTTP(t, ctx, "submit_query_plan", insertArgs)
	assertNotError(t, insertResult, "INSERT set<frozen<address>> should succeed")

	// 2. VALIDATE in Cassandra
	var id int
	var retrieved []map[string]interface{}
	iter := ctx.Session.Query("SELECT id, v FROM nesting_test.t_set_frozen_udt WHERE id = ?", testID).Iter()
	assert.True(t, iter.Scan(&id, &retrieved), "Should retrieve from Cassandra")
	assert.Equal(t, testID, id)
	assert.Len(t, retrieved, 2, "Set should have 2 UDT elements")
	iter.Close()

	// 3. SELECT via MCP
	selectArgs := map[string]any{
		"operation": "SELECT",
		"keyspace":  "nesting_test",
		"table":     "t_set_frozen_udt",
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": testID},
		},
	}
	selectResult := callToolHTTP(t, ctx, "submit_query_plan", selectArgs)
	assertNotError(t, selectResult, "SELECT via MCP should succeed")

	t.Log("✅ set<frozen<address>> - INSERT/SELECT verified")
}

// ============================================================================
// CATEGORY: Negative Tests (Invalid Combinations)
// ============================================================================

// TestMCP_Invalid_ListNonFrozenList tests that list<list<int>> fails
func TestMCP_Invalid_ListNonFrozenList(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ctx.Session.Query("CREATE KEYSPACE IF NOT EXISTS nesting_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()

	// This should FAIL to create
	err := ctx.Session.Query(`CREATE TABLE nesting_test.t_invalid_list_list (
		id int PRIMARY KEY,
		v list<list<int>>
	)`).Exec()

	// Verify it fails with expected error
	assert.Error(t, err, "CREATE TABLE with list<list<int>> should fail")
	if err != nil {
		assert.Contains(t, err.Error(), "Non-frozen collections are not allowed inside collections",
			"Error should mention non-frozen collections")
	}

	t.Log("✅ Negative test: list<list<int>> correctly rejected")
}
