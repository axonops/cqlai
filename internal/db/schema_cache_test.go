package db

import (
	"testing"
	"time"
)

func TestSchemaCache_CountTotalTables(t *testing.T) {
	// Create a mock schema cache
	sc := &SchemaCache{
		Tables: map[string][]CachedTableInfo{
			"keyspace1": {
				{TableInfo: TableInfo{TableName: "table1"}},
				{TableInfo: TableInfo{TableName: "table2"}},
			},
			"keyspace2": {
				{TableInfo: TableInfo{TableName: "table3"}},
			},
		},
	}

	count := sc.CountTotalTables()
	if count != 3 {
		t.Errorf("Expected 3 tables, got %d", count)
	}
}

func TestSchemaCache_GetTableSchema(t *testing.T) {
	// Create a mock schema cache
	sc := &SchemaCache{
		Tables: map[string][]CachedTableInfo{
			"test_keyspace": {
				{
					TableInfo: TableInfo{
						KeyspaceName:   "test_keyspace",
						TableName:      "test_table",
						PartitionKeys:  []string{"id"},
						ClusteringKeys: []string{"timestamp"},
					},
				},
			},
		},
		Columns: map[string]map[string][]ColumnInfo{
			"test_keyspace": {
				"test_table": {
					{Name: "id", DataType: "uuid", Kind: "partition_key", Position: 0},
					{Name: "timestamp", DataType: "timestamp", Kind: "clustering", Position: 0},
					{Name: "data", DataType: "text", Kind: "regular", Position: -1},
				},
			},
		},
	}

	schema, err := sc.GetTableSchema("test_keyspace", "test_table")
	if err != nil {
		t.Fatalf("Failed to get table schema: %v", err)
	}

	if schema.Keyspace != "test_keyspace" {
		t.Errorf("Expected keyspace 'test_keyspace', got '%s'", schema.Keyspace)
	}

	if schema.TableName != "test_table" {
		t.Errorf("Expected table 'test_table', got '%s'", schema.TableName)
	}

	if len(schema.PartitionKeys) != 1 || schema.PartitionKeys[0] != "id" {
		t.Errorf("Expected partition key 'id', got %v", schema.PartitionKeys)
	}

	if len(schema.ClusteringKeys) != 1 || schema.ClusteringKeys[0] != "timestamp" {
		t.Errorf("Expected clustering key 'timestamp', got %v", schema.ClusteringKeys)
	}

	if len(schema.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(schema.Columns))
	}
}

func TestSchemaCache_RefreshIfNeeded(t *testing.T) {
	sc := &SchemaCache{
		LastRefresh: time.Now(),
		Keyspaces:   []string{"test"},
	}

	// Should not need refresh (just set)
	err := sc.RefreshIfNeeded(5 * time.Minute)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Simulate old refresh time
	sc.LastRefresh = time.Now().Add(-10 * time.Minute)

	// Without session, should fail
	err = sc.RefreshIfNeeded(5 * time.Minute)
	if err == nil {
		t.Error("Expected error for refresh without session")
	}
}