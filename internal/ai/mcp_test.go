package ai

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractToolArg tests parameter extraction for all tools
func TestExtractToolArg(t *testing.T) {
	tests := []struct {
		name      string
		toolName  ToolName
		args      map[string]any
		want      string
		wantError bool
		errorMsg  string
	}{
		// ToolFuzzySearch tests
		{
			name:     "fuzzy search with valid query",
			toolName: ToolFuzzySearch,
			args:     map[string]any{"query": "users"},
			want:     "users",
		},
		{
			name:      "fuzzy search missing query",
			toolName:  ToolFuzzySearch,
			args:      map[string]any{},
			wantError: true,
			errorMsg:  "missing or invalid 'query' parameter",
		},
		{
			name:      "fuzzy search wrong type",
			toolName:  ToolFuzzySearch,
			args:      map[string]any{"query": 123},
			wantError: true,
			errorMsg:  "missing or invalid 'query' parameter",
		},

		// ToolGetSchema tests
		{
			name:     "get schema with valid params",
			toolName: ToolGetSchema,
			args:     map[string]any{"keyspace": "myapp", "table": "users"},
			want:     "myapp.users",
		},
		{
			name:      "get schema missing keyspace",
			toolName:  ToolGetSchema,
			args:      map[string]any{"table": "users"},
			wantError: true,
			errorMsg:  "missing or invalid 'keyspace' or 'table' parameters",
		},
		{
			name:      "get schema missing table",
			toolName:  ToolGetSchema,
			args:      map[string]any{"keyspace": "myapp"},
			wantError: true,
			errorMsg:  "missing or invalid 'keyspace' or 'table' parameters",
		},
		{
			name:      "get schema wrong type",
			toolName:  ToolGetSchema,
			args:      map[string]any{"keyspace": 123, "table": "users"},
			wantError: true,
			errorMsg:  "missing or invalid 'keyspace' or 'table' parameters",
		},

		// ToolListKeyspaces tests
		{
			name:     "list keyspaces no params",
			toolName: ToolListKeyspaces,
			args:     map[string]any{},
			want:     "",
		},
		{
			name:     "list keyspaces with extra params (ignored)",
			toolName: ToolListKeyspaces,
			args:     map[string]any{"extra": "ignored"},
			want:     "",
		},

		// ToolListTables tests
		{
			name:     "list tables with valid keyspace",
			toolName: ToolListTables,
			args:     map[string]any{"keyspace": "system"},
			want:     "system",
		},
		{
			name:      "list tables missing keyspace",
			toolName:  ToolListTables,
			args:      map[string]any{},
			wantError: true,
			errorMsg:  "missing or invalid 'keyspace' parameter",
		},
		{
			name:      "list tables wrong type",
			toolName:  ToolListTables,
			args:      map[string]any{"keyspace": 123},
			wantError: true,
			errorMsg:  "missing or invalid 'keyspace' parameter",
		},

		// ToolUserSelection tests
		{
			name:     "user selection with valid params",
			toolName: ToolUserSelection,
			args: map[string]any{
				"type":    "table",
				"options": []any{"users", "orders", "events"},
			},
			want: "table:users,orders,events",
		},
		{
			name:      "user selection missing type",
			toolName:  ToolUserSelection,
			args:      map[string]any{"options": []any{"users"}},
			wantError: true,
			errorMsg:  "missing or invalid 'type' or 'options' parameters",
		},
		{
			name:      "user selection missing options",
			toolName:  ToolUserSelection,
			args:      map[string]any{"type": "table"},
			wantError: true,
			errorMsg:  "missing or invalid 'type' or 'options' parameters",
		},

		// ToolNotEnoughInfo tests
		{
			name:     "not enough info with message",
			toolName: ToolNotEnoughInfo,
			args:     map[string]any{"message": "Please specify keyspace"},
			want:     "Please specify keyspace",
		},
		{
			name:      "not enough info missing message",
			toolName:  ToolNotEnoughInfo,
			args:      map[string]any{},
			wantError: true,
			errorMsg:  "missing or invalid 'message' parameter",
		},

		// ToolNotRelevant tests
		{
			name:     "not relevant with message",
			toolName: ToolNotRelevant,
			args:     map[string]any{"message": "Not about Cassandra"},
			want:     "Not about Cassandra",
		},

		// Unknown tool
		{
			name:      "unknown tool",
			toolName:  ToolName("INVALID"),
			args:      map[string]any{},
			wantError: true,
			errorMsg:  "unknown tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractToolArg(tt.toolName, tt.args)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestConvertToolDefinitionToMCPTool tests tool definition conversion
func TestConvertToolDefinitionToMCPTool(t *testing.T) {
	toolDef := ToolDefinition{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]any{
			"param1": map[string]any{
				"type":        "string",
				"description": "First parameter",
			},
		},
		Required: []string{"param1"},
	}

	tool, err := convertToolDefinitionToMCPTool(toolDef)

	require.NoError(t, err, "Should not error on valid tool definition")
	assert.Equal(t, "test_tool", tool.Name)
	assert.Equal(t, "A test tool", tool.Description)
	assert.NotNil(t, tool.RawInputSchema, "Should have raw input schema")

	// Verify schema is valid JSON
	var schema map[string]any
	err = json.Unmarshal(tool.RawInputSchema, &schema)
	require.NoError(t, err, "Schema should be valid JSON")

	// Verify schema structure
	assert.Equal(t, "object", schema["type"])
	assert.NotNil(t, schema["properties"])
}

// TestMetricsCollector tests metrics collection
func TestMetricsCollector(t *testing.T) {
	m := NewMetricsCollector()

	// Initial state
	snapshot := m.GetSnapshot()
	assert.Equal(t, int64(0), snapshot.TotalRequests)
	assert.Equal(t, int64(0), snapshot.SuccessfulRequests)
	assert.Equal(t, int64(0), snapshot.FailedRequests)
	assert.Equal(t, 0.0, snapshot.SuccessRate)

	// Record some successful calls
	m.RecordToolCall("fuzzy_search", true, 100*time.Millisecond)
	m.RecordToolCall("get_schema", true, 50*time.Millisecond)
	m.RecordToolCall("list_keyspaces", true, 25*time.Millisecond)

	snapshot = m.GetSnapshot()
	assert.Equal(t, int64(3), snapshot.TotalRequests)
	assert.Equal(t, int64(3), snapshot.SuccessfulRequests)
	assert.Equal(t, int64(0), snapshot.FailedRequests)
	assert.Equal(t, 100.0, snapshot.SuccessRate)

	// Verify tool-specific counts
	assert.Equal(t, int64(1), snapshot.ToolCalls["fuzzy_search"])
	assert.Equal(t, int64(1), snapshot.ToolCalls["get_schema"])
	assert.Equal(t, int64(1), snapshot.ToolCalls["list_keyspaces"])

	// Record a failure
	m.RecordToolCall("fuzzy_search", false, 10*time.Millisecond)

	snapshot = m.GetSnapshot()
	assert.Equal(t, int64(4), snapshot.TotalRequests)
	assert.Equal(t, int64(3), snapshot.SuccessfulRequests)
	assert.Equal(t, int64(1), snapshot.FailedRequests)
	assert.Equal(t, 75.0, snapshot.SuccessRate)
	assert.Equal(t, int64(2), snapshot.ToolCalls["fuzzy_search"])
}

// TestJoinHelper tests the join helper function
func TestJoinHelper(t *testing.T) {
	tests := []struct {
		name string
		strs []string
		sep  string
		want string
	}{
		{
			name: "empty slice",
			strs: []string{},
			sep:  ",",
			want: "",
		},
		{
			name: "single element",
			strs: []string{"one"},
			sep:  ",",
			want: "one",
		},
		{
			name: "multiple elements with comma",
			strs: []string{"one", "two", "three"},
			sep:  ",",
			want: "one,two,three",
		},
		{
			name: "multiple elements with colon",
			strs: []string{"users", "orders"},
			sep:  ":",
			want: "users:orders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := join(tt.strs, tt.sep)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestEncodeJSONSchema tests JSON schema encoding
func TestEncodeJSONSchema(t *testing.T) {
	schema := map[string]any{
		"query": map[string]any{
			"type":        "string",
			"description": "Search query",
		},
	}

	schemaJSON, err := encodeJSONSchema(schema)
	require.NoError(t, err, "Should encode valid schema")

	// Verify it's valid JSON
	var decoded map[string]any
	err = json.Unmarshal(schemaJSON, &decoded)
	require.NoError(t, err, "Should produce valid JSON")

	// Verify structure
	assert.Equal(t, "object", decoded["type"])
	assert.NotNil(t, decoded["properties"])
}
