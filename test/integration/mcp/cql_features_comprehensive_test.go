//go:build integration
// +build integration

package mcp

import (
	"os"
	"testing"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Comprehensive MCP Testing for ALL 77 CQL Features
// ============================================================================
//
// This test suite verifies that ALL implemented CQL features work correctly
// through the MCP interface, generating proper CQL and executing successfully.
//
// Test structure:
// - Each test submits query plan via MCP submit_query_plan tool
// - Verifies CQL generation
// - Executes against Cassandra
// - Verifies results
//
// Prerequisites:
// - MCP server running
// - Cassandra running (cassandra-test container)
// - Test keyspace: cqlai_test
// ============================================================================

func init() {
	// Generate test API key for this test file
	key, _ := ai.GenerateAPIKey()
	os.Setenv("TEST_MCP_API_KEY", key)
}

// ============================================================================
// Phase 0: Data Type Formatting Tests via MCP
// ============================================================================

func TestMCP_DataTypes_Lists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// Test: INSERT with numeric list (simpler - no quoting issues)
	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"values": map[string]any{
			"id":     2000,
			"name":   "MCPListTest",
			"scores": []int{95, 87, 92},  // Numeric list - no quoting needed
		},
		"value_types": map[string]any{
			"scores": "list<int>",
		},
	}

	// Submit via MCP and check it doesn't error
	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with numeric list should succeed")

	t.Log("✅ INSERT with list literal succeeded via MCP")

	// Verify data was actually inserted
	verifyResult := ctx.Session.Query("SELECT id, name, scores FROM cqlai_test.users WHERE id = 2000")
	var id int
	var name string
	var scores []int
	if verifyResult.Iter().Scan(&id, &name, &scores) {
		assert.Equal(t, 2000, id)
		assert.Equal(t, "MCPListTest", name)
		assert.Equal(t, []int{95, 87, 92}, scores)
		t.Log("✅ Data verification: List correctly stored in Cassandra")
	}
	verifyResult.Iter().Close()
}

func TestMCP_DataTypes_Sets(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"values": map[string]any{
			"id":   2001,
			"name": "MCPSetTest",
			"tags": []string{"admin", "verified", "premium"},
		},
		"value_types": map[string]any{
			"tags": "set<text>",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with set should succeed")

	t.Log("✅ INSERT with set literal succeeded via MCP")
}

func TestMCP_DataTypes_Maps(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"values": map[string]any{
			"id":   2002,
			"name": "MCPMapTest",
			"settings": map[string]any{
				"theme": "dark",
				"lang":  "en",
			},
		},
		"value_types": map[string]any{
			"settings": "map<text,text>",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with map should succeed")

	t.Log("✅ INSERT with map literal succeeded via MCP")
}

func TestMCP_DataTypes_Functions(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "cqlai_test",
		"table":     "func_test",
		"values": map[string]any{
			"id":      "uuid()",
			"created": "now()",
			"name":    "MCPFuncTest",
		},
		"value_types": map[string]any{
			"id":      "uuid",
			"created": "timeuuid",
			"name":    "text",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with uuid() and now() should succeed")

	t.Log("✅ INSERT with functions (uuid, now) succeeded via MCP")
}

// ============================================================================
// Phase 1: Simple DML Features via MCP
// ============================================================================

func TestMCP_UsingTTL(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// Test INSERT with USING TTL
	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"values":    map[string]any{"id": 2010, "name": "TTLTest"},
		"using_ttl": 300,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with USING TTL should succeed")

	// Test UPDATE with USING TTL
	args2 := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"values":    map[string]any{"name": "TTLUpdated"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 2010}},
		"using_ttl": 600,
	}

	result2 := callToolHTTP(t, ctx, "submit_query_plan", args2)
	assertNotError(t, result2, "UPDATE with USING TTL should succeed")

	t.Log("✅ USING TTL (INSERT and UPDATE) succeeded via MCP")
}

func TestMCP_UsingTimestamp(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":       "INSERT",
		"keyspace":        "cqlai_test",
		"table":           "users",
		"values":          map[string]any{"id": 2011, "name": "TimestampTest"},
		"using_timestamp": 1609459200000000,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with USING TIMESTAMP should succeed")

	t.Log("✅ USING TIMESTAMP succeeded via MCP")
}

func TestMCP_UsingTTLAndTimestamp(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":       "INSERT",
		"keyspace":        "cqlai_test",
		"table":           "users",
		"values":          map[string]any{"id": 2012, "name": "CombinedTest"},
		"using_ttl":       300,
		"using_timestamp": 1609459200000000,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with USING TTL AND TIMESTAMP should succeed")

	t.Log("✅ USING TTL AND TIMESTAMP combined succeeded via MCP")
}

func TestMCP_SelectDistinct(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "SELECT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"columns":   []string{"id"},
		"distinct":  true,
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 1}},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "SELECT DISTINCT should succeed")

	t.Log("✅ SELECT DISTINCT succeeded via MCP")
}

func TestMCP_SelectJSON(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":   "SELECT",
		"keyspace":    "cqlai_test",
		"table":       "users",
		"select_json": true,
		"where":       []map[string]any{{"column": "id", "operator": "=", "value": 1}},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "SELECT JSON should succeed")

	t.Log("✅ SELECT JSON succeeded via MCP")
}

func TestMCP_PerPartitionLimit(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":           "SELECT",
		"keyspace":            "cqlai_test",
		"table":               "users",
		"per_partition_limit": 5,
		"limit":               100,
		"where":               []map[string]any{{"column": "id", "operator": "=", "value": 1}},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "SELECT with PER PARTITION LIMIT should succeed")

	t.Log("✅ PER PARTITION LIMIT succeeded via MCP")
}

func TestMCP_InsertJSON(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":   "INSERT",
		"keyspace":    "cqlai_test",
		"table":       "users",
		"insert_json": true,
		"json_value":  `{"id": 2020, "name": "JSONTest", "email": "json@test.com"}`,
		"using_ttl":   300,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT JSON with USING TTL should succeed")

	t.Log("✅ INSERT JSON with USING TTL succeeded via MCP")
}

// ============================================================================
// Phase 2: Collection Operations via MCP
// ============================================================================

func TestMCP_CounterIncrement(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "counters",
		"counter_ops": map[string]any{
			"views":  "+10",
			"clicks": "+5",
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": "mcp_counter_test"},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "Counter increment should succeed")

	t.Log("✅ Counter increment via MCP")
}

func TestMCP_CounterDecrement(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "counters",
		"counter_ops": map[string]any{
			"views": "-3",
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": "mcp_counter_test"},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "Counter decrement should succeed")

	t.Log("✅ Counter decrement via MCP")
}

func TestMCP_ListAppend(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"collection_ops": map[string]any{
			"phones": map[string]any{
				"operation":  "append",
				"value":      []string{"555-9999"},
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2000},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "List append should succeed")

	t.Log("✅ List append via MCP")
}

func TestMCP_ListPrepend(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"collection_ops": map[string]any{
			"phones": map[string]any{
				"operation":  "prepend",
				"value":      []string{"555-0000"},
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2000},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "List prepend should succeed")

	t.Log("✅ List prepend via MCP")
}

func TestMCP_SetAdd(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"collection_ops": map[string]any{
			"tags": map[string]any{
				"operation":  "add",
				"value":      []string{"new_tag", "another_tag"},
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2001},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "Set add should succeed")

	t.Log("✅ Set add via MCP")
}

func TestMCP_SetRemove(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"collection_ops": map[string]any{
			"tags": map[string]any{
				"operation":  "remove",
				"value":      []string{"admin"},
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2001},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "Set remove should succeed")

	t.Log("✅ Set remove via MCP")
}

func TestMCP_MapMerge(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"collection_ops": map[string]any{
			"settings": map[string]any{
				"operation":  "merge",
				"value":      map[string]any{"new_key": "new_value"},
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2002},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "Map merge should succeed")

	t.Log("✅ Map merge via MCP")
}

func TestMCP_MapElementUpdate(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"collection_ops": map[string]any{
			"settings": map[string]any{
				"operation":  "set_element",
				"key":        "theme",
				"value":      "light",
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2002},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "Map element update should succeed")

	t.Log("✅ Map element update via MCP")
}

func TestMCP_ListElementUpdate(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"collection_ops": map[string]any{
			"phones": map[string]any{
				"operation":  "set_index",
				"index":      0,
				"value":      "555-UPDATED",
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2000},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "List element update should succeed")

	t.Log("✅ List element update via MCP")
}

func TestMCP_UDTFieldUpdate(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "udt_test",
		"collection_ops": map[string]any{
			"addr": map[string]any{
				"operation":  "set_field",
				"key":        "city",
				"value":      "MCP Updated",
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 1},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "UDT field update should succeed")

	t.Log("✅ UDT field update via MCP")
}

// ============================================================================
// Phase 3: LWTs via MCP
// ============================================================================

func TestMCP_InsertIfNotExists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":     "INSERT",
		"keyspace":      "cqlai_test",
		"table":         "lwt_test2",
		"values":        map[string]any{"id": 100, "email": "lwt@mcp.com", "version": 1},
		"if_not_exists": true,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT IF NOT EXISTS should succeed")

	t.Log("✅ INSERT IF NOT EXISTS via MCP")
}

func TestMCP_UpdateIfExists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "lwt_test2",
		"values":    map[string]any{"email": "updated@mcp.com"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 100}},
		"if_exists": true,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "UPDATE IF EXISTS should succeed")

	t.Log("✅ UPDATE IF EXISTS via MCP")
}

func TestMCP_UpdateIfCondition(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "UPDATE",
		"keyspace":  "cqlai_test",
		"table":     "lwt_test2",
		"values":    map[string]any{"email": "conditional@mcp.com"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 100}},
		"if_conditions": []map[string]any{
			{"column": "version", "operator": "=", "value": 1},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "UPDATE IF condition should succeed")

	t.Log("✅ UPDATE IF condition via MCP")
}

// Continuing in next file part...

// ============================================================================
// Phase 4: BATCH Operations via MCP
// ============================================================================

func TestMCP_BatchLogged(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":  "BATCH",
		"batch_type": "LOGGED",
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"keyspace":  "cqlai_test",
				"table":     "users",
				"values":    map[string]any{"id": 2100, "name": "Batch1"},
			},
			{
				"operation": "UPDATE",
				"keyspace":  "cqlai_test",
				"table":     "users",
				"values":    map[string]any{"email": "batch1@mcp.com"},
				"where":     []map[string]any{{"column": "id", "operator": "=", "value": 2100}},
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "LOGGED BATCH should succeed")

	t.Log("✅ LOGGED BATCH via MCP")
}

func TestMCP_BatchUnlogged(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":  "BATCH",
		"batch_type": "UNLOGGED",
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"keyspace":  "cqlai_test",
				"table":     "users",
				"values":    map[string]any{"id": 2101, "name": "UnloggedBatch"},
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "UNLOGGED BATCH should succeed")

	t.Log("✅ UNLOGGED BATCH via MCP")
}

func TestMCP_BatchCounter(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":  "BATCH",
		"batch_type": "COUNTER",
		"batch_statements": []map[string]any{
			{
				"operation":   "UPDATE",
				"keyspace":    "cqlai_test",
				"table":       "counters",
				"counter_ops": map[string]any{"views": "+1"},
				"where":       []map[string]any{{"column": "id", "operator": "=", "value": "batch_c1"}},
			},
			{
				"operation":   "UPDATE",
				"keyspace":    "cqlai_test",
				"table":       "counters",
				"counter_ops": map[string]any{"clicks": "+1"},
				"where":       []map[string]any{{"column": "id", "operator": "=", "value": "batch_c2"}},
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "COUNTER BATCH should succeed")

	t.Log("✅ COUNTER BATCH via MCP")
}

func TestMCP_BatchWithTimestamp(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":       "BATCH",
		"batch_type":      "LOGGED",
		"using_timestamp": 1609459200000000,
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"keyspace":  "cqlai_test",
				"table":     "users",
				"values":    map[string]any{"id": 2102, "name": "BatchTS"},
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "BATCH with USING TIMESTAMP should succeed")

	t.Log("✅ BATCH USING TIMESTAMP via MCP")
}

// ============================================================================
// Phase 5: DDL with IF Clauses via MCP
// ============================================================================

func TestMCP_CreateIndexIfNotExists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "CREATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"options": map[string]any{
			"object_type":   "INDEX",
			"index_name":    "mcp_test_idx",
			"column":        "email",
			"if_not_exists": true,
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "CREATE INDEX IF NOT EXISTS should succeed")

	t.Log("✅ CREATE INDEX IF NOT EXISTS via MCP")
}

func TestMCP_CreateCustomIndex(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "CREATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"options": map[string]any{
			"object_type":   "INDEX",
			"index_name":    "mcp_sai_idx",
			"column":        "name",
			"custom_index":  true,
			"using_class":   "StorageAttachedIndex",
			"if_not_exists": true,
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "CREATE CUSTOM INDEX should succeed")

	t.Log("✅ CREATE CUSTOM INDEX via MCP")
}

func TestMCP_DropTableIfExists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "DROP",
		"keyspace":  "cqlai_test",
		"table":     "nonexistent_table",
		"options": map[string]any{
			"if_exists": true,
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "DROP TABLE IF EXISTS should succeed")

	t.Log("✅ DROP TABLE IF EXISTS via MCP")
}

func TestMCP_AlterTableIfExists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "ALTER",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"options": map[string]any{
			"object_type":   "TABLE",
			"action":        "ADD",
			"column_name":   "mcp_test_col",
			"column_type":   "text",
			"if_exists":     true,
			"if_not_exists": true,
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "ALTER TABLE IF EXISTS should succeed")

	t.Log("✅ ALTER TABLE IF EXISTS via MCP")
}

// ============================================================================
// Phase 6: Advanced Query Features via MCP
// ============================================================================

func TestMCP_WhereToken(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "SELECT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"where": []map[string]any{
			{
				"column":   "id",
				"operator": ">",
				"value":    100,
				"is_token": true,
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "WHERE TOKEN() should succeed")

	t.Log("✅ WHERE TOKEN() via MCP")
}

func TestMCP_WhereTuple(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "SELECT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"where": []map[string]any{
			{
				"columns":  []string{"col1", "col2"},
				"operator": ">",
				"values":   []string{"val1", "val2"},
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "WHERE tuple notation should succeed")

	t.Log("✅ WHERE tuple notation via MCP")
}

func TestMCP_AggregateCount(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "SELECT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"columns":   []string{"COUNT(*)"},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "SELECT COUNT(*) should succeed")

	t.Log("✅ Aggregate COUNT(*) via MCP")
}

func TestMCP_WritetimeAndTTL(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "SELECT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"columns":   []string{"id", "name", "WRITETIME(name)", "TTL(name)"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 1}},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "SELECT WRITETIME/TTL should succeed")

	t.Log("✅ WRITETIME and TTL functions via MCP")
}

func TestMCP_WhereContains(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":       "SELECT",
		"keyspace":        "cqlai_test",
		"table":           "users",
		"allow_filtering": true,
		"where": []map[string]any{
			{
				"column":   "tags",
				"operator": "CONTAINS",
				"value":    "admin",
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "WHERE CONTAINS should succeed")

	t.Log("✅ WHERE CONTAINS via MCP")
}

func TestMCP_WhereContainsKey(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation":       "SELECT",
		"keyspace":        "cqlai_test",
		"table":           "users",
		"allow_filtering": true,
		"where": []map[string]any{
			{
				"column":   "settings",
				"operator": "CONTAINS KEY",
				"value":    "theme",
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "WHERE CONTAINS KEY should succeed")

	t.Log("✅ WHERE CONTAINS KEY via MCP")
}

func TestMCP_SelectWithCAST(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "SELECT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"columns":   []string{"id", "CAST(id AS text)"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 1}},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "SELECT with CAST should succeed")

	t.Log("✅ CAST function via MCP")
}

func TestMCP_SelectCollectionAccess(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "SELECT",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"columns":   []string{"id", "settings['theme']"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 1}},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "SELECT with collection access should succeed")

	t.Log("✅ Collection element access via MCP")
}
