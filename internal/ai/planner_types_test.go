package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Phase 0: Data Type Formatting Tests (TDD)
// ============================================================================
//
// These tests are written BEFORE implementation to drive the design.
// They should FAIL initially, then pass as we implement formatValue() refactor.
//
// Manual Cassandra testing completed 2026-01-01 - all syntax verified against
// Cassandra 5.0.6. See: claude-notes/features/phase0_data_types_research.md
// ============================================================================

// ============================================================================
// Primitive Formatting Tests
// ============================================================================

func TestFormatPrimitive_String(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"basic string", "hello", "'hello'"},
		{"string with single quote", "don't", "'don''t'"}, // Escape ' → ''
		{"empty string", "", "''"},
		{"string with multiple quotes", "it's a test's result", "'it''s a test''s result'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPrimitive(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFormatPrimitive_UUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000"},
		{"timeuuid", "c2ed71ab-5b6b-4f24-be98-f0461ffe4aa1", "c2ed71ab-5b6b-4f24-be98-f0461ffe4aa1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPrimitive(tt.input)
			assert.Equal(t, tt.expected, got, "UUIDs should not be quoted")
		})
	}
}

func TestFormatPrimitive_Numeric(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"int", 42, "42"},
		{"negative int", -100, "-100"},
		{"int64", int64(9223372036854775807), "9223372036854775807"},
		{"float64", 3.14, "3.14"},
		{"float32", float32(2.5), "2.5"},
		{"zero", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPrimitive(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFormatPrimitive_Boolean(t *testing.T) {
	assert.Equal(t, "true", formatPrimitive(true))
	assert.Equal(t, "false", formatPrimitive(false))
}

func TestFormatPrimitive_Nil(t *testing.T) {
	assert.Equal(t, "null", formatPrimitive(nil))
}

// ============================================================================
// Function Detection Tests
// ============================================================================

func TestIsFunctionCall(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid functions
		{"uuid()", true},
		{"now()", true},
		{"toTimestamp()", true},
		{"toDate(1234567890)", true},
		{"count(*)", true},

		// Not functions
		{"'uuid()'", false},   // Quoted
		{"\"now()\"", false},  // Quoted
		{"hello", false},      // No parentheses
		{"", false},           // Empty
		{"uuid", false},       // Missing ()
		{"(value)", false},    // Starts with (
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isFunctionCall(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestIsUUID(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"c2ed71ab-5b6b-4f24-be98-f0461ffe4aa1", true},
		{"not-a-uuid", false},
		{"550e8400-e29b-41d4", false},  // Too short
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isUUID(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// ============================================================================
// List Formatting Tests
// ============================================================================

func TestFormatList(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		elementType string
		expected    string
	}{
		{
			name:        "text list - basic",
			input:       []any{"a", "b", "c"},
			elementType: "text",
			expected:    "['a', 'b', 'c']",
		},
		{
			name:        "empty list",
			input:       []any{},
			elementType: "text",
			expected:    "[]",
		},
		{
			name:        "single element",
			input:       []any{"only"},
			elementType: "text",
			expected:    "['only']",
		},
		{
			name:        "numeric list",
			input:       []any{1, 2, 3},
			elementType: "int",
			expected:    "[1, 2, 3]",
		},
		{
			name:        "list with duplicates",
			input:       []any{"a", "b", "a"},
			elementType: "text",
			expected:    "['a', 'b', 'a']", // Lists preserve duplicates
		},
		{
			name:        "list from []string",
			input:       []string{"x", "y", "z"},
			elementType: "text",
			expected:    "['x', 'y', 'z']",
		},
		{
			name:        "list from []int",
			input:       []int{10, 20, 30},
			elementType: "int",
			expected:    "[10, 20, 30]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatList(tt.input, tt.elementType)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// ============================================================================
// Set Formatting Tests
// ============================================================================

func TestFormatSet(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		elementType string
		expected    string
	}{
		{
			name:        "text set - basic",
			input:       []any{"admin", "verified"},
			elementType: "text",
			expected:    "{'admin', 'verified'}",
		},
		{
			name:        "empty set",
			input:       []any{},
			elementType: "text",
			expected:    "{}",
		},
		{
			name:        "single element",
			input:       []any{"guest"},
			elementType: "text",
			expected:    "{'guest'}",
		},
		{
			name:        "numeric set",
			input:       []any{7, 13, 21},
			elementType: "int",
			expected:    "{7, 13, 21}",
		},
		{
			name:        "deduplicate set",
			input:       []any{"a", "b", "a", "c", "b"},
			elementType: "text",
			expected:    "{'a', 'b', 'c'}", // Deduped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSet(tt.input, tt.elementType)
			// Note: Set order may vary, but for simple test data we can check exact match
			assert.Equal(t, tt.expected, got)
		})
	}
}

// ============================================================================
// Map Formatting Tests
// ============================================================================

func TestFormatMap(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]any
		keyType   string
		valueType string
		wantKey   string // Key to check for (maps are unordered in Go)
		wantValue string // Value to check for
	}{
		{
			name:      "basic map",
			input:     map[string]any{"theme": "dark", "lang": "en"},
			keyType:   "text",
			valueType: "text",
			wantKey:   "'theme'",  // Keys should be quoted
			wantValue: "'dark'",
		},
		{
			name:      "single entry",
			input:     map[string]any{"key": "value"},
			keyType:   "text",
			valueType: "text",
			wantKey:   "'key'",
			wantValue: "'value'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMap(tt.input, tt.keyType, tt.valueType)
			assert.Contains(t, got, tt.wantKey, "Map should contain quoted key")
			assert.Contains(t, got, tt.wantValue, "Map should contain value")
			assert.True(t, strings.HasPrefix(got, "{") && strings.HasSuffix(got, "}"), "Map should use curly braces")
		})
	}
}

func TestFormatMap_Empty(t *testing.T) {
	got := formatMap(map[string]any{}, "text", "text")
	assert.Equal(t, "{}", got)
}

// ============================================================================
// UDT Formatting Tests
// ============================================================================

func TestFormatUDT(t *testing.T) {
	tests := []struct {
		name         string
		input        map[string]any
		udtName      string
		wantField    string // Unquoted field name to check
		wantValue    string // Quoted value to check
	}{
		{
			name:      "complete UDT",
			input:     map[string]any{"street": "123 Main", "city": "NYC", "zip": "10001"},
			udtName:   "address",
			wantField: "street:",  // Field name NOT quoted
			wantValue: "'123 Main'",
		},
		{
			name:      "partial UDT",
			input:     map[string]any{"city": "LA"},
			udtName:   "address",
			wantField: "city:",
			wantValue: "'LA'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUDT(tt.input, tt.udtName)
			assert.Contains(t, got, tt.wantField, "UDT field names should NOT be quoted")
			assert.Contains(t, got, tt.wantValue, "UDT values should be quoted")
			assert.True(t, strings.HasPrefix(got, "{") && strings.HasSuffix(got, "}"), "UDT should use curly braces")
			assert.NotContains(t, got, "'street':", "Field names should not have quotes")
		})
	}
}

func TestFormatUDT_Empty(t *testing.T) {
	got := formatUDT(map[string]any{}, "address")
	assert.Equal(t, "{}", got)
}

// ============================================================================
// Tuple Formatting Tests
// ============================================================================

func TestFormatTuple(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "basic tuple",
			input:    []any{"123 Main St", "NYC", "10001"},
			expected: "('123 Main St', 'NYC', '10001')",
		},
		{
			name:     "mixed types",
			input:    []any{42, "text", true},
			expected: "(42, 'text', true)",
		},
		{
			name:     "single element",
			input:    []any{"only"},
			expected: "('only')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTuple(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// ============================================================================
// Blob Formatting Tests
// ============================================================================

func TestFormatBlob(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "hex string with 0x",
			input:    "0xCAFEBABE",
			expected: "0xCAFEBABE",
		},
		{
			name:     "hex string without 0x",
			input:    "CAFEBABE",
			expected: "0xCAFEBABE",
		},
		{
			name:     "byte array",
			input:    []byte{0xCA, 0xFE, 0xBA, 0xBE},
			expected: "0xcafebabe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBlob(tt.input)
			assert.True(t, strings.HasPrefix(got, "0x"), "Blob must have 0x prefix")
			assert.Equal(t, strings.ToLower(tt.expected), strings.ToLower(got))
		})
	}
}

// ============================================================================
// Main formatValue() Tests with Type Hints
// ============================================================================

func TestFormatValue_WithTypeHints(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		typeHint string
		expected string
	}{
		// Primitives
		{"string", "hello", "text", "'hello'"},
		{"int", 42, "int", "42"},
		{"boolean", true, "boolean", "true"},
		{"uuid no hint", "550e8400-e29b-41d4-a716-446655440000", "", "550e8400-e29b-41d4-a716-446655440000"},

		// Collections with hints
		{"list with hint", []any{"a", "b"}, "list<text>", "['a', 'b']"},
		{"set with hint", []any{1, 2, 3}, "set<int>", "{1, 2, 3}"},
		{"map with hint", map[string]any{"k": "v"}, "map<text,text>", "{'k': 'v'}"},

		// Functions (auto-detected, no hint needed)
		{"uuid function", "uuid()", "", "uuid()"},
		{"now function", "now()", "", "now()"},
		{"toTimestamp function", "toTimestamp(1234567890)", "", "toTimestamp(1234567890)"},

		// Tuple
		{"tuple", []any{"a", "b", "c"}, "tuple<text,text,text>", "('a', 'b', 'c')"},

		// Blob
		{"blob", "CAFEBABE", "blob", "0xCAFEBABE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.value, tt.typeHint)
			assert.Equal(t, strings.ToLower(tt.expected), strings.ToLower(got))
		})
	}
}

// ============================================================================
// Type Hint Parsing Tests
// ============================================================================

func TestParseTypeHint(t *testing.T) {
	tests := []struct {
		hint        string
		wantBase    string
		wantElement string
	}{
		{"text", "text", ""},
		{"int", "int", ""},
		{"list<text>", "list", "text"},
		{"set<int>", "set", "int"},
		{"list<uuid>", "list", "uuid"},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.hint, func(t *testing.T) {
			gotBase, gotElement := parseTypeHint(tt.hint)
			assert.Equal(t, tt.wantBase, gotBase)
			assert.Equal(t, tt.wantElement, gotElement)
		})
	}
}

func TestParseMapTypes(t *testing.T) {
	tests := []struct {
		hint      string
		wantKey   string
		wantValue string
	}{
		{"map<text,text>", "text", "text"},
		{"map<text,int>", "text", "int"},
		{"map<int,list<text>>", "int", "list<text>"},  // Nested
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.hint, func(t *testing.T) {
			gotKey, gotValue := parseMapTypes(tt.hint)
			assert.Equal(t, tt.wantKey, gotKey)
			assert.Equal(t, tt.wantValue, gotValue)
		})
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestConvertToSlice(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		wantSlice []any
		wantOK    bool
	}{
		{
			name:      "[]any",
			input:     []any{1, 2, 3},
			wantSlice: []any{1, 2, 3},
			wantOK:    true,
		},
		{
			name:      "[]string",
			input:     []string{"a", "b"},
			wantSlice: []any{"a", "b"},
			wantOK:    true,
		},
		{
			name:      "[]int",
			input:     []int{10, 20},
			wantSlice: []any{10, 20},
			wantOK:    true,
		},
		{
			name:      "not a slice",
			input:     "string",
			wantSlice: nil,
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := convertToSlice(tt.input)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantSlice, got)
			}
		})
	}
}

func TestDeduplicateElements(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected int // Expected length after dedup
	}{
		{
			name:     "no duplicates",
			input:    []any{"a", "b", "c"},
			expected: 3,
		},
		{
			name:     "some duplicates",
			input:    []any{"a", "b", "a", "c", "b"},
			expected: 3, // a, b, c
		},
		{
			name:     "all same",
			input:    []any{"x", "x", "x"},
			expected: 1,
		},
		{
			name:     "empty",
			input:    []any{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deduplicateElements(tt.input)
			assert.Equal(t, tt.expected, len(got))
		})
	}
}

// ============================================================================
// Edge Cases and Integration Tests
// ============================================================================

func TestFormatValue_EdgeCases(t *testing.T) {
	// Nested list (if we support later)
	// nested := []any{[]any{1, 2}, []any{3, 4}}
	// got := formatValue(nested, "list<list<int>>")
	// assert.Equal(t, "[[1, 2], [3, 4]]", got)

	// Mixed function and literals
	values := map[string]any{
		"id":      "uuid()",
		"name":    "Alice",
		"created": "now()",
	}

	// uuid() should not be quoted
	gotID := formatValue(values["id"], "uuid")
	assert.Equal(t, "uuid()", gotID)

	// name should be quoted
	gotName := formatValue(values["name"], "text")
	assert.Equal(t, "'Alice'", gotName)

	// now() should not be quoted
	gotCreated := formatValue(values["created"], "timeuuid")
	assert.Equal(t, "now()", gotCreated)
}

func TestFormatValue_NoTypeHint_Inference(t *testing.T) {
	// Without type hints, should infer from value

	// String → quoted
	assert.Equal(t, "'hello'", formatValue("hello", ""))

	// Function → not quoted
	assert.Equal(t, "uuid()", formatValue("uuid()", ""))

	// Number → unquoted
	assert.Equal(t, "42", formatValue(42, ""))

	// UUID → not quoted
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", formatValue("550e8400-e29b-41d4-a716-446655440000", ""))
}
