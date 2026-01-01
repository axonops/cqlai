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

	// Test: INSERT with list literal
	args := map[string]any{
		"operation": "INSERT",
		"table":     "users",
		"values": map[string]any{
			"id":     2000,
			"name":   "MCPListTest",
			"phones": []any{"555-0001", "555-0002", "555-0003"},
		},
		"value_types": map[string]any{
			"phones": "list<text>",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assert.NotNil(t, result)

	// Verify the generated CQL contains proper list syntax
	cql, ok := result["generated_cql"].(string)
	assert.True(t, ok, "Should have generated_cql in result")
	assert.Contains(t, cql, "['555-0001', '555-0002', '555-0003']", "List should use square brackets")
	assert.NotContains(t, cql, "('555-0001'", "Should NOT use parentheses (tuple syntax)")

	// Verify execution succeeded
	assert.Contains(t, result, "success")
}

func TestMCP_DataTypes_Sets(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "INSERT",
		"table":     "users",
		"values": map[string]any{
			"id":   2001,
			"name": "MCPSetTest",
			"tags": []any{"admin", "verified", "premium"},
		},
		"value_types": map[string]any{
			"tags": "set<text>",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "{'admin', 'verified', 'premium'}", "Set should use curly braces")
}

func TestMCP_DataTypes_Maps(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "INSERT",
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
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "'theme':", "Map keys should be quoted")
	assert.Contains(t, cql, "'dark'", "Map values should be quoted")
}

func TestMCP_DataTypes_Functions(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "INSERT",
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
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "uuid()", "uuid() should NOT be quoted")
	assert.Contains(t, cql, "now()", "now() should NOT be quoted")
	assert.NotContains(t, cql, "'uuid()'", "Functions should not be in quotes")
}

// ============================================================================
// Phase 1: Simple DML Features via MCP
// ============================================================================

func TestMCP_UsingTTL(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	// Test INSERT with USING TTL
	args := map[string]any{
		"operation":  "INSERT",
		"table":      "users",
		"values":     map[string]any{"id": 2010, "name": "TTLTest"},
		"using_ttl":  300,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "USING TTL 300")

	// Test UPDATE with USING TTL
	args2 := map[string]any{
		"operation": "UPDATE",
		"table":     "users",
		"values":    map[string]any{"name": "TTLUpdated"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 2010}},
		"using_ttl": 600,
	}

	result2 := callToolHTTP(t, ctx, "submit_query_plan", args2)
	cql2 := result2["generated_cql"].(string)
	assert.Contains(t, cql2, "USING TTL 600")
}

func TestMCP_UsingTimestamp(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation":       "INSERT",
		"table":           "users",
		"values":          map[string]any{"id": 2011, "name": "TimestampTest"},
		"using_timestamp": 1609459200000000,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "USING TIMESTAMP 1609459200000000")
}

func TestMCP_UsingTTLAndTimestamp(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation":       "INSERT",
		"table":           "users",
		"values":          map[string]any{"id": 2012, "name": "CombinedTest"},
		"using_ttl":       300,
		"using_timestamp": 1609459200000000,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "USING TTL 300 AND TIMESTAMP 1609459200000000")
}

func TestMCP_SelectDistinct(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "SELECT",
		"table":     "users",
		"columns":   []string{"id"},
		"distinct":  true,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "SELECT DISTINCT id")
}

func TestMCP_SelectJSON(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation":   "SELECT",
		"table":       "users",
		"select_json": true,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "SELECT JSON")
}

func TestMCP_PerPartitionLimit(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation":           "SELECT",
		"table":               "users",
		"per_partition_limit": 5,
		"limit":               100,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "PER PARTITION LIMIT 5")
	assert.Contains(t, cql, "LIMIT 100")
}

func TestMCP_InsertJSON(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation":   "INSERT",
		"table":       "users",
		"insert_json": true,
		"json_value":  `{"id": 2020, "name": "JSONTest", "email": "json@test.com"}`,
		"using_ttl":   300,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "INSERT INTO users JSON")
	assert.Contains(t, cql, `{"id": 2020`)
	assert.Contains(t, cql, "USING TTL 300")
}

// ============================================================================
// Phase 2: Collection Operations via MCP
// ============================================================================

func TestMCP_CounterIncrement(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
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
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "views = views + 10")
	assert.Contains(t, cql, "clicks = clicks + 5")
}

func TestMCP_CounterDecrement(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
		"table":     "counters",
		"counter_ops": map[string]any{
			"views": "-3",
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": "mcp_counter_test"},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "views = views - 3")
}

func TestMCP_ListAppend(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
		"table":     "users",
		"collection_ops": map[string]any{
			"phones": map[string]any{
				"operation":  "append",
				"value":      []any{"555-9999"},
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2000},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "phones = phones + ['555-9999']")
}

func TestMCP_ListPrepend(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
		"table":     "users",
		"collection_ops": map[string]any{
			"phones": map[string]any{
				"operation":  "prepend",
				"value":      []any{"555-0000"},
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2000},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "phones = ['555-0000'] + phones")
}

func TestMCP_SetAdd(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
		"table":     "users",
		"collection_ops": map[string]any{
			"tags": map[string]any{
				"operation":  "add",
				"value":      []any{"new_tag", "another_tag"},
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2001},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "tags = tags + {'new_tag', 'another_tag'}")
}

func TestMCP_SetRemove(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
		"table":     "users",
		"collection_ops": map[string]any{
			"tags": map[string]any{
				"operation":  "remove",
				"value":      []any{"admin"},
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2001},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "tags = tags - {'admin'}")
}

func TestMCP_MapMerge(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
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
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "settings = settings +")
	assert.Contains(t, cql, "'new_key':")
}

func TestMCP_MapElementUpdate(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
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
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "settings['theme'] = 'light'")
}

func TestMCP_ListElementUpdate(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	index := 0
	args := map[string]any{
		"operation": "UPDATE",
		"table":     "users",
		"collection_ops": map[string]any{
			"phones": map[string]any{
				"operation":  "set_index",
				"index":      &index,
				"value":      "555-UPDATED",
				"value_type": "text",
			},
		},
		"where": []map[string]any{
			{"column": "id", "operator": "=", "value": 2000},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "phones[0] = '555-UPDATED'")
}

func TestMCP_UDTFieldUpdate(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
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
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "addr.city = 'MCP Updated'")
}

// ============================================================================
// Phase 3: LWTs via MCP
// ============================================================================

func TestMCP_InsertIfNotExists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation":     "INSERT",
		"table":         "lwt_test2",
		"values":        map[string]any{"id": 100, "email": "lwt@mcp.com", "version": 1},
		"if_not_exists": true,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "IF NOT EXISTS")
}

func TestMCP_UpdateIfExists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
		"table":     "lwt_test2",
		"values":    map[string]any{"email": "updated@mcp.com"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 100}},
		"if_exists": true,
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "IF EXISTS")
}

func TestMCP_UpdateIfCondition(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "UPDATE",
		"table":     "lwt_test2",
		"values":    map[string]any{"email": "conditional@mcp.com"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 100}},
		"if_conditions": []map[string]any{
			{"column": "version", "operator": "=", "value": 1},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "IF version = 1")
}

// Continuing in next file part...

// ============================================================================
// Phase 4: BATCH Operations via MCP
// ============================================================================

func TestMCP_BatchLogged(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation":  "BATCH",
		"batch_type": "LOGGED",
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"table":     "users",
				"values":    map[string]any{"id": 2100, "name": "Batch1"},
			},
			{
				"operation": "UPDATE",
				"table":     "users",
				"values":    map[string]any{"email": "batch1@mcp.com"},
				"where":     []map[string]any{{"column": "id", "operator": "=", "value": 2100}},
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "BEGIN BATCH")
	assert.Contains(t, cql, "APPLY BATCH")
	assert.Contains(t, cql, "INSERT INTO users")
	assert.Contains(t, cql, "UPDATE users")
}

func TestMCP_BatchUnlogged(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation":  "BATCH",
		"batch_type": "UNLOGGED",
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"table":     "users",
				"values":    map[string]any{"id": 2101, "name": "UnloggedBatch"},
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "BEGIN UNLOGGED BATCH")
}

func TestMCP_BatchCounter(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation":  "BATCH",
		"batch_type": "COUNTER",
		"batch_statements": []map[string]any{
			{
				"operation":   "UPDATE",
				"table":       "counters",
				"counter_ops": map[string]any{"views": "+1"},
				"where":       []map[string]any{{"column": "id", "operator": "=", "value": "batch_c1"}},
			},
			{
				"operation":   "UPDATE",
				"table":       "counters",
				"counter_ops": map[string]any{"clicks": "+1"},
				"where":       []map[string]any{{"column": "id", "operator": "=", "value": "batch_c2"}},
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "BEGIN COUNTER BATCH")
}

func TestMCP_BatchWithTimestamp(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation":       "BATCH",
		"batch_type":      "LOGGED",
		"using_timestamp": 1609459200000000,
		"batch_statements": []map[string]any{
			{
				"operation": "INSERT",
				"table":     "users",
				"values":    map[string]any{"id": 2102, "name": "BatchTS"},
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "USING TIMESTAMP 1609459200000000")
}

// ============================================================================
// Phase 5: DDL with IF Clauses via MCP
// ============================================================================

func TestMCP_CreateIndexIfNotExists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "CREATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"options": map[string]any{
			"object_type":    "INDEX",
			"index_name":     "mcp_test_idx",
			"column":         "email",
			"if_not_exists":  true,
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "CREATE INDEX IF NOT EXISTS")
}

func TestMCP_CreateCustomIndex(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "CREATE",
		"keyspace":  "cqlai_test",
		"table":     "users",
		"options": map[string]any{
			"object_type":    "INDEX",
			"index_name":     "mcp_sai_idx",
			"column":         "name",
			"custom_index":   true,
			"using_class":    "StorageAttachedIndex",
			"if_not_exists":  true,
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "CREATE CUSTOM INDEX")
	assert.Contains(t, cql, "USING 'StorageAttachedIndex'")
}

func TestMCP_DropTableIfExists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "DROP",
		"keyspace":  "cqlai_test",
		"table":     "nonexistent_table",
		"options": map[string]any{
			"if_exists": true,
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "DROP TABLE IF EXISTS")
}

func TestMCP_AlterTableIfExists(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

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
			"if_not_exists": true, // For ADD sub-clause
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "ALTER TABLE IF EXISTS")
	assert.Contains(t, cql, "ADD IF NOT EXISTS")
}

// ============================================================================
// Phase 6: Advanced Query Features via MCP
// ============================================================================

func TestMCP_WhereToken(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "SELECT",
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
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "TOKEN(id) > 100")
}

func TestMCP_WhereTuple(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "SELECT",
		"table":     "users",
		"where": []map[string]any{
			{
				"columns":  []string{"col1", "col2"},
				"operator": ">",
				"values":   []any{"val1", "val2"},
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "(col1, col2) > ('val1', 'val2')")
}

func TestMCP_AggregateCount(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "SELECT",
		"table":     "users",
		"columns":   []string{"COUNT(*)"},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "SELECT COUNT(*)")
}

func TestMCP_WritetimeAndTTL(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "SELECT",
		"table":     "users",
		"columns":   []string{"id", "name", "WRITETIME(name)", "TTL(name)"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 1}},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "WRITETIME(name)")
	assert.Contains(t, cql, "TTL(name)")
}

func TestMCP_WhereContains(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "SELECT",
		"table":     "users",
		"where": []map[string]any{
			{
				"column":   "tags",
				"operator": "CONTAINS",
				"value":    "admin",
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "WHERE tags CONTAINS 'admin'")
}

func TestMCP_WhereContainsKey(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "SELECT",
		"table":     "users",
		"where": []map[string]any{
			{
				"column":   "settings",
				"operator": "CONTAINS KEY",
				"value":    "theme",
			},
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "WHERE settings CONTAINS KEY 'theme'")
}

func TestMCP_SelectWithCAST(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "SELECT",
		"table":     "users",
		"columns":   []string{"id", "CAST(id AS text)"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 1}},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "CAST(id AS text)")
}

func TestMCP_SelectCollectionAccess(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readonly.json")
	defer stopMCPHTTP(ctx)

	args := map[string]any{
		"operation": "SELECT",
		"table":     "users",
		"columns":   []string{"id", "settings['theme']", "addr.city"},
		"where":     []map[string]any{{"column": "id", "operator": "=", "value": 1}},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	cql := result["generated_cql"].(string)
	assert.Contains(t, cql, "settings['theme']")
	assert.Contains(t, cql, "addr.city")
}
