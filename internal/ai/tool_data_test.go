package ai

import (
	"testing"

	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createMockCache creates a SchemaCache with test data
func createMockCache() *db.SchemaCache {
	cache := &db.SchemaCache{
		Keyspaces: []string{"system", "test_ks", "production"},
		Tables: map[string][]db.CachedTableInfo{
			"system": {
				{TableInfo: db.TableInfo{TableName: "local"}},
				{TableInfo: db.TableInfo{TableName: "peers"}},
			},
			"test_ks": {
				{TableInfo: db.TableInfo{TableName: "users"}},
				{TableInfo: db.TableInfo{TableName: "orders"}},
			},
			"production": {
				{TableInfo: db.TableInfo{TableName: "events"}},
			},
		},
	}
	return cache
}

// createMockResolver creates a Resolver with test data
func createMockResolver() *Resolver {
	cache := createMockCache()
	return NewResolver(cache)
}

// TestGetToolData_FuzzySearch tests fuzzy search data retrieval
func TestGetToolData_FuzzySearch(t *testing.T) {
	resolver := createMockResolver()
	cache := createMockCache()

	tests := []struct {
		name      string
		query     string
		wantError bool
		checkFunc func(t *testing.T, data any)
	}{
		{
			name:      "search for users",
			query:     "users",
			wantError: false,
			checkFunc: func(t *testing.T, data any) {
				candidates, ok := data.([]TableCandidate)
				require.True(t, ok, "Expected []TableCandidate type")
				assert.NotNil(t, candidates, "Candidates should not be nil")
				// Should return slice (may be empty if no match, but that's ok)
			},
		},
		{
			name:      "search with no matches",
			query:     "nonexistent_xyz",
			wantError: false,
			checkFunc: func(t *testing.T, data any) {
				candidates, ok := data.([]TableCandidate)
				require.True(t, ok, "Expected []TableCandidate type")
				assert.NotNil(t, candidates, "Should return empty slice, not nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := getToolData(resolver, cache, ToolFuzzySearch, tt.query)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, data)
				}
			}
		})
	}
}

// TestGetToolData_FuzzySearch_NilResolver tests error handling for nil resolver
func TestGetToolData_FuzzySearch_NilResolver(t *testing.T) {
	cache := createMockCache()

	data, err := getToolData(nil, cache, ToolFuzzySearch, "users")

	assert.Error(t, err, "Expected error for nil resolver")
	assert.Nil(t, data, "Expected nil data on error")
	assert.Contains(t, err.Error(), "resolver not available")
}

// TestGetToolData_GetSchema tests schema retrieval
func TestGetToolData_GetSchema(t *testing.T) {
	cache := createMockCache()

	tests := []struct {
		name      string
		arg       string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "invalid format - missing dot",
			arg:       "users",
			wantError: true,
			errorMsg:  "invalid table reference",
		},
		{
			name:      "invalid format - too many parts",
			arg:       "ks.table.extra",
			wantError: true,
			errorMsg:  "invalid table reference",
		},
		{
			name:      "valid format - schema not in cache",
			arg:       "test_ks.users",
			wantError: true,
			errorMsg:  "schema not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := getToolData(nil, cache, ToolGetSchema, tt.arg)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				schema, ok := data.(*db.TableSchema)
				require.True(t, ok, "Expected *db.TableSchema type")
				assert.NotNil(t, schema)
			}
		})
	}
}

// TestGetToolData_GetSchema_NilCache tests error handling for nil cache
func TestGetToolData_GetSchema_NilCache(t *testing.T) {
	data, err := getToolData(nil, nil, ToolGetSchema, "ks.table")

	assert.Error(t, err, "Expected error for nil cache")
	assert.Nil(t, data, "Expected nil data on error")
	assert.Contains(t, err.Error(), "schema cache not available")
}

// TestGetToolData_ListKeyspaces tests keyspace listing
func TestGetToolData_ListKeyspaces(t *testing.T) {
	cache := createMockCache()

	data, err := getToolData(nil, cache, ToolListKeyspaces, "")

	assert.NoError(t, err, "Expected no error for list keyspaces")
	keyspaces, ok := data.([]string)
	require.True(t, ok, "Expected []string type")
	assert.Equal(t, []string{"system", "test_ks", "production"}, keyspaces)
}

// TestGetToolData_ListKeyspaces_NilCache tests error handling for nil cache
func TestGetToolData_ListKeyspaces_NilCache(t *testing.T) {
	data, err := getToolData(nil, nil, ToolListKeyspaces, "")

	assert.Error(t, err, "Expected error for nil cache")
	assert.Nil(t, data, "Expected nil data on error")
	assert.Contains(t, err.Error(), "schema cache not available")
}

// TestGetToolData_ListTables tests table listing
func TestGetToolData_ListTables(t *testing.T) {
	cache := createMockCache()

	tests := []struct {
		name          string
		keyspace      string
		wantError     bool
		expectedCount int
	}{
		{
			name:          "list system tables",
			keyspace:      "system",
			wantError:     false,
			expectedCount: 2,
		},
		{
			name:          "list test_ks tables",
			keyspace:      "test_ks",
			wantError:     false,
			expectedCount: 2,
		},
		{
			name:          "list production tables",
			keyspace:      "production",
			wantError:     false,
			expectedCount: 1,
		},
		{
			name:          "list nonexistent keyspace",
			keyspace:      "nonexistent",
			wantError:     false,
			expectedCount: 0, // Returns empty slice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := getToolData(nil, cache, ToolListTables, tt.keyspace)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, "Expected no error for list tables")
				tables, ok := data.([]db.CachedTableInfo)
				require.True(t, ok, "Expected []db.CachedTableInfo type")
				assert.Equal(t, tt.expectedCount, len(tables),
					"Expected %d tables for keyspace %s", tt.expectedCount, tt.keyspace)
			}
		})
	}
}

// TestGetToolData_ListTables_NilCache tests error handling for nil cache
func TestGetToolData_ListTables_NilCache(t *testing.T) {
	data, err := getToolData(nil, nil, ToolListTables, "system")

	assert.Error(t, err, "Expected error for nil cache")
	assert.Nil(t, data, "Expected nil data on error")
	assert.Contains(t, err.Error(), "schema cache not available")
}

// TestGetToolData_UITools tests UI/control flow tools
func TestGetToolData_UITools(t *testing.T) {
	tools := []ToolName{
		ToolUserSelection,
		ToolNotEnoughInfo,
		ToolNotRelevant,
	}

	for _, tool := range tools {
		t.Run(string(tool), func(t *testing.T) {
			arg := "test argument"
			data, err := getToolData(nil, nil, tool, arg)

			assert.NoError(t, err, "UI tools should not error")
			assert.Equal(t, arg, data, "UI tools should return argument as-is")
		})
	}
}

// TestGetToolData_InvalidTool tests unknown tool handling
func TestGetToolData_InvalidTool(t *testing.T) {
	cache := createMockCache()
	resolver := createMockResolver()

	data, err := getToolData(resolver, cache, ToolName("INVALID_TOOL"), "test")

	assert.Error(t, err, "Expected error for invalid tool")
	assert.Nil(t, data, "Expected nil data on error")
	assert.Contains(t, err.Error(), "unknown tool")
}

// TestGetToolData_SubmitQueryPlan tests that query plan tool returns error
func TestGetToolData_SubmitQueryPlan(t *testing.T) {
	cache := createMockCache()

	data, err := getToolData(nil, cache, ToolSubmitQueryPlan, "")

	assert.Error(t, err, "ToolSubmitQueryPlan should return error")
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "should be handled by ExecuteToolCallTyped")
}

// TestGetToolData_Info tests that info tool returns error
func TestGetToolData_Info(t *testing.T) {
	cache := createMockCache()

	data, err := getToolData(nil, cache, ToolInfo, "")

	assert.Error(t, err, "ToolInfo should return error")
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "should be handled by ExecuteToolCallTyped")
}

// TestGetToolData_TypeSafety verifies all return types are correct
func TestGetToolData_TypeSafety(t *testing.T) {
	cache := createMockCache()
	resolver := createMockResolver()

	tests := []struct {
		name         string
		tool         ToolName
		arg          string
		expectedType string
	}{
		{
			name:         "FuzzySearch returns []TableCandidate",
			tool:         ToolFuzzySearch,
			arg:          "test",
			expectedType: "[]ai.TableCandidate",
		},
		{
			name:         "ListKeyspaces returns []string",
			tool:         ToolListKeyspaces,
			arg:          "",
			expectedType: "[]string",
		},
		{
			name:         "ListTables returns []db.CachedTableInfo",
			tool:         ToolListTables,
			arg:          "system",
			expectedType: "[]db.CachedTableInfo",
		},
		{
			name:         "UserSelection returns string",
			tool:         ToolUserSelection,
			arg:          "test",
			expectedType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := getToolData(resolver, cache, tt.tool, tt.arg)
			require.NoError(t, err, "Should not error for valid tool")
			assert.NotNil(t, data, "Data should not be nil")

			// Verify type matches expected
			switch tt.tool {
			case ToolFuzzySearch:
				_, ok := data.([]TableCandidate)
				assert.True(t, ok, "Type mismatch for %s", tt.tool)
			case ToolListKeyspaces:
				_, ok := data.([]string)
				assert.True(t, ok, "Type mismatch for %s", tt.tool)
			case ToolListTables:
				_, ok := data.([]db.CachedTableInfo)
				assert.True(t, ok, "Type mismatch for %s", tt.tool)
			case ToolUserSelection, ToolNotEnoughInfo, ToolNotRelevant:
				_, ok := data.(string)
				assert.True(t, ok, "Type mismatch for %s", tt.tool)
			}
		})
	}
}
