package ai

import (
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/db"
)

func TestSearchIndexManager_BuildIndex(t *testing.T) {
	// Create mock schema cache
	cache := &db.SchemaCache{
		Tables: map[string][]db.CachedTableInfo{
			"test_keyspace": {
				{TableInfo: db.TableInfo{TableName: "user_profiles"}},
				{TableInfo: db.TableInfo{TableName: "user_sessions"}},
				{TableInfo: db.TableInfo{TableName: "product_catalog"}},
			},
			"analytics": {
				{TableInfo: db.TableInfo{TableName: "events"}},
				{TableInfo: db.TableInfo{TableName: "metrics"}},
			},
		},
		Columns: map[string]map[string][]db.ColumnInfo{
			"test_keyspace": {
				"user_profiles": {
					{Name: "user_id", DataType: "uuid"},
					{Name: "name", DataType: "text"},
				},
				"user_sessions": {
					{Name: "session_id", DataType: "uuid"},
					{Name: "user_id", DataType: "uuid"},
				},
			},
		},
		Keyspaces:   []string{"test_keyspace", "analytics"},
		LastRefresh: time.Now(),
	}

	sim := NewSearchIndexManager(cache)

	// Build the index
	err := sim.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Check that index was built
	stats := sim.GetStats()
	if stats["total_entries"].(int) != 5 {
		t.Errorf("Expected 5 entries in index, got %v", stats["total_entries"])
	}
}

func TestSearchIndexManager_FindTables(t *testing.T) {
	// Create mock schema cache
	cache := &db.SchemaCache{
		Tables: map[string][]db.CachedTableInfo{
			"test_keyspace": {
				{TableInfo: db.TableInfo{
					KeyspaceName: "test_keyspace",
					TableName:    "user_profiles",
				}},
				{TableInfo: db.TableInfo{
					KeyspaceName: "test_keyspace",
					TableName:    "user_sessions",
				}},
				{TableInfo: db.TableInfo{
					KeyspaceName: "test_keyspace",
					TableName:    "product_catalog",
				}},
			},
		},
		Columns: map[string]map[string][]db.ColumnInfo{
			"test_keyspace": {
				"user_profiles": {
					{Name: "user_id", DataType: "uuid"},
					{Name: "name", DataType: "text"},
				},
			},
		},
		Keyspaces:   []string{"test_keyspace"},
		LastRefresh: time.Now(),
	}

	sim := NewSearchIndexManager(cache)

	// Test exact match
	matches := sim.FindTables("user_profiles", 10)
	if len(matches) == 0 {
		t.Errorf("Expected at least 1 match for 'user_profiles', got 0")
	}
	// Check that the exact match has the highest score
	if len(matches) > 0 {
		if matches[0].Table != "user_profiles" {
			t.Errorf("Expected 'user_profiles' to be first match, got '%s'", matches[0].Table)
		}
		if matches[0].Score != 100 {
			t.Errorf("Expected score 100 for exact match, got %f", matches[0].Score)
		}
	}

	// Test prefix match
	matches = sim.FindTables("user", 10)
	if len(matches) < 2 {
		t.Errorf("Expected at least 2 matches for 'user', got %d", len(matches))
	}

	// Test partial match
	matches = sim.FindTables("profile", 10)
	if len(matches) < 1 {
		t.Errorf("Expected at least 1 match for 'profile', got %d", len(matches))
	}

	// Test no match
	matches = sim.FindTables("nonexistent", 10)
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for 'nonexistent', got %d", len(matches))
	}
}

func TestTokenizeTableName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "underscore separated",
			input:    "user_profiles",
			expected: []string{"user_profiles", "user", "profiles"},
		},
		{
			name:     "camelCase",
			input:    "UserProfiles",
			expected: []string{"userprofiles", "user", "profiles"},
		},
		{
			name:     "hyphenated",
			input:    "user-profiles",
			expected: []string{"user-profiles", "user", "profiles"},
		},
		{
			name:     "single word",
			input:    "users",
			expected: []string{"users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizeTableName(tt.input)

			// Check that all expected tokens are present
			for _, exp := range tt.expected {
				found := false
				for _, tok := range tokens {
					if tok == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected token '%s' not found in %v", exp, tokens)
				}
			}
		})
	}
}