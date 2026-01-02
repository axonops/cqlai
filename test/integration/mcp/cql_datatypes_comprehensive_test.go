//go:build integration
// +build integration

package mcp

import (
	"os"
	"testing"

	"github.com/axonops/cqlai/internal/ai"
)

func init() {
	// Generate test API key
	key, _ := ai.GenerateAPIKey()
	os.Setenv("TEST_MCP_API_KEY_TYPES", key)
}

// ============================================================================
// Comprehensive Data Type Testing via MCP
// ============================================================================
// Verify EVERY Cassandra data type works through MCP interface
// ============================================================================

// TestMCP_PrimitiveTypes_Text tests text, ascii, varchar
func TestMCP_PrimitiveTypes_Text(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "type_test",
		"table":     "all_types",
		"values": map[string]any{
			"id":          3000,
			"text_col":    "text value",
			"ascii_col":   "ascii value",
			"varchar_col": "varchar value",
		},
		"value_types": map[string]any{
			"text_col":    "text",
			"ascii_col":   "ascii",
			"varchar_col": "varchar",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with text types should succeed")

	t.Log("✅ text, ascii, varchar via MCP")
}

// TestMCP_PrimitiveTypes_Integers tests all integer types
func TestMCP_PrimitiveTypes_Integers(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "type_test",
		"table":     "all_types",
		"values": map[string]any{
			"id":           3001,
			"tinyint_col":  127,
			"smallint_col": 32767,
			"int_col":      2147483647,
			"bigint_col":   9223372036854775, // Smaller value to avoid JSON precision issues
			"varint_col":   123456789012345,
		},
		"value_types": map[string]any{
			"tinyint_col":  "tinyint",
			"smallint_col": "smallint",
			"int_col":      "int",
			"bigint_col":   "bigint",
			"varint_col":   "varint",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with integer types should succeed")

	t.Log("✅ tinyint, smallint, int, bigint, varint via MCP")
}

// TestMCP_PrimitiveTypes_Floats tests floating point types
func TestMCP_PrimitiveTypes_Floats(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "type_test",
		"table":     "all_types",
		"values": map[string]any{
			"id":          3002,
			"float_col":   3.14,
			"double_col":  2.718281828,
			"decimal_col": "99.99", // Decimal as string for precision
		},
		"value_types": map[string]any{
			"float_col":   "float",
			"double_col":  "double",
			"decimal_col": "decimal",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with float types should succeed")

	t.Log("✅ float, double, decimal via MCP")
}

// TestMCP_PrimitiveTypes_Boolean tests boolean
func TestMCP_PrimitiveTypes_Boolean(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "type_test",
		"table":     "all_types",
		"values": map[string]any{
			"id":          3003,
			"boolean_col": true,
		},
		"value_types": map[string]any{
			"boolean_col": "boolean",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with boolean should succeed")

	t.Log("✅ boolean via MCP")
}

// TestMCP_PrimitiveTypes_Blob tests blob (hex data)
func TestMCP_PrimitiveTypes_Blob(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "type_test",
		"table":     "all_types",
		"values": map[string]any{
			"id":       3004,
			"blob_col": "CAFEBABE",
		},
		"value_types": map[string]any{
			"blob_col": "blob",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with blob should succeed")

	t.Log("✅ blob via MCP")
}

// TestMCP_PrimitiveTypes_DateTime tests date, time, timestamp types
func TestMCP_PrimitiveTypes_DateTime(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "type_test",
		"table":     "all_types",
		"values": map[string]any{
			"id":            3005,
			"date_col":      "2024-01-15",
			"time_col":      "14:30:00",
			"timestamp_col": "2024-01-15T14:30:00Z",
		},
		"value_types": map[string]any{
			"date_col":      "date",
			"time_col":      "time",
			"timestamp_col": "timestamp",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with date/time types should succeed")

	t.Log("✅ date, time, timestamp via MCP")
}

// TestMCP_PrimitiveTypes_UUID tests uuid and timeuuid
func TestMCP_PrimitiveTypes_UUID(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "type_test",
		"table":     "all_types",
		"values": map[string]any{
			"id":           3006,
			"uuid_col":     "550e8400-e29b-41d4-a716-446655440000",
			"timeuuid_col": "now()", // Function call
		},
		"value_types": map[string]any{
			"uuid_col":     "uuid",
			"timeuuid_col": "timeuuid",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with UUID types should succeed")

	t.Log("✅ uuid, timeuuid via MCP")
}

// TestMCP_PrimitiveTypes_Inet tests inet (IP addresses)
func TestMCP_PrimitiveTypes_Inet(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "type_test",
		"table":     "all_types",
		"values": map[string]any{
			"id":       3007,
			"inet_col": "192.168.1.1",
		},
		"value_types": map[string]any{
			"inet_col": "inet",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with inet should succeed")

	t.Log("✅ inet via MCP")
}

// TestMCP_PrimitiveTypes_Duration tests duration
func TestMCP_PrimitiveTypes_Duration(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "type_test",
		"table":     "all_types",
		"values": map[string]any{
			"id":           3008,
			"duration_col": "12h30m",
		},
		"value_types": map[string]any{
			"duration_col": "duration",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with duration should succeed")

	t.Log("✅ duration via MCP")
}

// TestMCP_ComplexTypes_UDT tests user-defined types via MCP
func TestMCP_ComplexTypes_UDT(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "cqlai_test",
		"table":     "udt_test",
		"values": map[string]any{
			"id":   3009,
			"name": "UDTTest",
			"addr": map[string]any{
				"street": "456 Oak St",
				"city":   "SF",
				"zip":    "94102",
			},
		},
		"value_types": map[string]any{
			"addr": "udt", // Mark as UDT for formatUDT()
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with UDT should succeed")

	t.Log("✅ UDT (user-defined type) via MCP")
}

// TestMCP_ComplexTypes_Tuple tests tuple type
func TestMCP_ComplexTypes_Tuple(t *testing.T) {
	ctx := startMCPFromConfigHTTP(t, "testdata/readwrite.json")
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	// Create test table with tuple
	ctx.Session.Query("CREATE TABLE IF NOT EXISTS type_test.tuple_test (id int PRIMARY KEY, coords frozen<tuple<int, int, int>>)").Exec()

	args := map[string]any{
		"operation": "INSERT",
		"keyspace":  "type_test",
		"table":     "tuple_test",
		"values": map[string]any{
			"id":     3010,
			"coords": []int{10, 20, 30},
		},
		"value_types": map[string]any{
			"coords": "tuple<int,int,int>",
		},
	}

	result := callToolHTTP(t, ctx, "submit_query_plan", args)
	assertNotError(t, result, "INSERT with tuple should succeed")

	t.Log("✅ tuple via MCP")
}
