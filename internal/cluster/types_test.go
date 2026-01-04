package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ColumnType Tests
// ============================================================================

func TestColumnType_IsNativeType(t *testing.T) {
	tests := []struct {
		name     string
		colType  *ColumnType
		expected bool
	}{
		{"native type", &ColumnType{IsNative: true}, true},
		{"non-native type", &ColumnType{IsNative: false}, false},
		{"nil type", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colType.IsNativeType())
		})
	}
}

func TestColumnType_IsCollectionType(t *testing.T) {
	tests := []struct {
		name     string
		colType  *ColumnType
		expected bool
	}{
		{"list type", &ColumnType{Category: TypeCategoryCollection, Name: "list<int>"}, true},
		{"set type", &ColumnType{Category: TypeCategoryCollection, Name: "set<text>"}, true},
		{"map type", &ColumnType{Category: TypeCategoryCollection, Name: "map<int, text>"}, true},
		{"simple type", &ColumnType{Category: TypeCategorySimple}, false},
		{"nil type", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colType.IsCollectionType())
		})
	}
}

func TestColumnType_IsMapType(t *testing.T) {
	tests := []struct {
		name     string
		colType  *ColumnType
		expected bool
	}{
		{"map type", &ColumnType{Category: TypeCategoryCollection, Name: "map<int, text>"}, true},
		{"Map with caps", &ColumnType{Category: TypeCategoryCollection, Name: "Map<int, text>"}, true},
		{"list type", &ColumnType{Category: TypeCategoryCollection, Name: "list<int>"}, false},
		{"simple type", &ColumnType{Category: TypeCategorySimple, Name: "int"}, false},
		{"nil type", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colType.IsMapType())
		})
	}
}

func TestColumnType_IsListType(t *testing.T) {
	tests := []struct {
		name     string
		colType  *ColumnType
		expected bool
	}{
		{"list type", &ColumnType{Category: TypeCategoryCollection, Name: "list<int>"}, true},
		{"List with caps", &ColumnType{Category: TypeCategoryCollection, Name: "List<text>"}, true},
		{"map type", &ColumnType{Category: TypeCategoryCollection, Name: "map<int, text>"}, false},
		{"set type", &ColumnType{Category: TypeCategoryCollection, Name: "set<text>"}, false},
		{"nil type", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colType.IsListType())
		})
	}
}

func TestColumnType_IsSetType(t *testing.T) {
	tests := []struct {
		name     string
		colType  *ColumnType
		expected bool
	}{
		{"set type", &ColumnType{Category: TypeCategoryCollection, Name: "set<int>"}, true},
		{"Set with caps", &ColumnType{Category: TypeCategoryCollection, Name: "Set<text>"}, true},
		{"list type", &ColumnType{Category: TypeCategoryCollection, Name: "list<int>"}, false},
		{"map type", &ColumnType{Category: TypeCategoryCollection, Name: "map<int, text>"}, false},
		{"nil type", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colType.IsSetType())
		})
	}
}

func TestColumnType_IsUDTType(t *testing.T) {
	tests := []struct {
		name     string
		colType  *ColumnType
		expected bool
	}{
		{"UDT type", &ColumnType{Category: TypeCategoryUDT}, true},
		{"simple type", &ColumnType{Category: TypeCategorySimple}, false},
		{"nil type", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colType.IsUDTType())
		})
	}
}

func TestColumnType_IsCounterType(t *testing.T) {
	tests := []struct {
		name     string
		colType  *ColumnType
		expected bool
	}{
		{"counter type", &ColumnType{Name: "counter"}, true},
		{"Counter with caps", &ColumnType{Name: "Counter"}, true},
		{"COUNTER uppercase", &ColumnType{Name: "COUNTER"}, true},
		{"int type", &ColumnType{Name: "int"}, false},
		{"nil type", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colType.IsCounterType())
		})
	}
}

func TestColumnType_IsVectorType(t *testing.T) {
	tests := []struct {
		name     string
		colType  *ColumnType
		expected bool
	}{
		{"vector type", &ColumnType{IsVector: true, VectorDimension: 384}, true},
		{"non-vector", &ColumnType{IsVector: false}, false},
		{"nil type", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colType.IsVectorType())
		})
	}
}

func TestColumnType_GetVectorDimension(t *testing.T) {
	tests := []struct {
		name     string
		colType  *ColumnType
		expected int
	}{
		{"vector 384", &ColumnType{IsVector: true, VectorDimension: 384}, 384},
		{"vector 768", &ColumnType{IsVector: true, VectorDimension: 768}, 768},
		{"non-vector", &ColumnType{IsVector: false}, 0},
		{"nil type", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colType.GetVectorDimension())
		})
	}
}

func TestColumnType_GetVectorElementType(t *testing.T) {
	tests := []struct {
		name     string
		colType  *ColumnType
		expected string
	}{
		{"vector float", &ColumnType{IsVector: true, VectorElementType: "float"}, "float"},
		{"vector float64", &ColumnType{IsVector: true, VectorElementType: "float64"}, "float64"},
		{"non-vector", &ColumnType{IsVector: false}, ""},
		{"nil type", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colType.GetVectorElementType())
		})
	}
}

// ============================================================================
// ColumnInfo Tests
// ============================================================================

func TestColumnInfo_IsPartitionKey(t *testing.T) {
	tests := []struct {
		name     string
		colInfo  *ColumnInfo
		expected bool
	}{
		{"partition key", &ColumnInfo{Kind: ColumnKindPartitionKey}, true},
		{"clustering key", &ColumnInfo{Kind: ColumnKindClusteringKey}, false},
		{"regular", &ColumnInfo{Kind: ColumnKindRegular}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colInfo.IsPartitionKey())
		})
	}
}

func TestColumnInfo_IsClusteringKey(t *testing.T) {
	tests := []struct {
		name     string
		colInfo  *ColumnInfo
		expected bool
	}{
		{"clustering key", &ColumnInfo{Kind: ColumnKindClusteringKey}, true},
		{"partition key", &ColumnInfo{Kind: ColumnKindPartitionKey}, false},
		{"regular", &ColumnInfo{Kind: ColumnKindRegular}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colInfo.IsClusteringKey())
		})
	}
}

func TestColumnInfo_IsRegular(t *testing.T) {
	tests := []struct {
		name     string
		colInfo  *ColumnInfo
		expected bool
	}{
		{"regular", &ColumnInfo{Kind: ColumnKindRegular}, true},
		{"partition key", &ColumnInfo{Kind: ColumnKindPartitionKey}, false},
		{"static", &ColumnInfo{Kind: ColumnKindStatic}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colInfo.IsRegular())
		})
	}
}

func TestColumnInfo_IsStatic(t *testing.T) {
	tests := []struct {
		name     string
		colInfo  *ColumnInfo
		expected bool
	}{
		{"static", &ColumnInfo{Kind: ColumnKindStatic}, true},
		{"regular", &ColumnInfo{Kind: ColumnKindRegular}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colInfo.IsStatic())
		})
	}
}

func TestColumnInfo_IsCounter(t *testing.T) {
	tests := []struct {
		name     string
		colInfo  *ColumnInfo
		expected bool
	}{
		{"counter column", &ColumnInfo{Type: &ColumnType{Name: "counter"}}, true},
		{"int column", &ColumnInfo{Type: &ColumnType{Name: "int"}}, false},
		{"nil type", &ColumnInfo{Type: nil}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colInfo.IsCounter())
		})
	}
}

func TestColumnInfo_HasIndex(t *testing.T) {
	tests := []struct {
		name     string
		colInfo  *ColumnInfo
		expected bool
	}{
		{"with index", &ColumnInfo{Index: &ColumnIndexInfo{Name: "idx_test"}}, true},
		{"empty index", &ColumnInfo{Index: &ColumnIndexInfo{Name: ""}}, false},
		{"nil index", &ColumnInfo{Index: nil}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colInfo.HasIndex())
		})
	}
}

func TestColumnInfo_HasSAI(t *testing.T) {
	tests := []struct {
		name     string
		colInfo  *ColumnInfo
		expected bool
	}{
		{"SAI index", &ColumnInfo{Index: &ColumnIndexInfo{Name: "idx_sai", IsSAIIndex: true}}, true},
		{"non-SAI index", &ColumnInfo{Index: &ColumnIndexInfo{Name: "idx_regular", IsSAIIndex: false}}, false},
		{"nil index", &ColumnInfo{Index: nil}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colInfo.HasSAI())
		})
	}
}

func TestColumnInfo_IsVectorColumn(t *testing.T) {
	tests := []struct {
		name     string
		colInfo  *ColumnInfo
		expected bool
	}{
		{"vector column", &ColumnInfo{Type: &ColumnType{IsVector: true, VectorDimension: 384}}, true},
		{"non-vector", &ColumnInfo{Type: &ColumnType{IsVector: false}}, false},
		{"nil type", &ColumnInfo{Type: nil}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colInfo.IsVectorColumn())
		})
	}
}

func TestColumnInfo_GetIndexesOnColumn(t *testing.T) {
	idx := &ColumnIndexInfo{Name: "test_idx"}

	tests := []struct {
		name     string
		colInfo  *ColumnInfo
		expected int
	}{
		{"with index", &ColumnInfo{Index: idx}, 1},
		{"empty name", &ColumnInfo{Index: &ColumnIndexInfo{Name: ""}}, 0},
		{"nil index", &ColumnInfo{Index: nil}, 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.colInfo.GetIndexesOnColumn()
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

// ============================================================================
// ColumnIndexInfo Tests
// ============================================================================

func TestColumnIndexInfo_IsNative(t *testing.T) {
	tests := []struct {
		name     string
		idxInfo  *ColumnIndexInfo
		expected bool
	}{
		{"native", &ColumnIndexInfo{IsNativeIndex: true}, true},
		{"custom", &ColumnIndexInfo{IsNativeIndex: false}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.idxInfo.IsNative())
		})
	}
}

func TestColumnIndexInfo_IsSAI(t *testing.T) {
	tests := []struct {
		name     string
		idxInfo  *ColumnIndexInfo
		expected bool
	}{
		{"SAI", &ColumnIndexInfo{IsSAIIndex: true}, true},
		{"non-SAI", &ColumnIndexInfo{IsSAIIndex: false}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.idxInfo.IsSAI())
		})
	}
}

func TestColumnIndexInfo_GetIndexCategoryString(t *testing.T) {
	tests := []struct {
		name     string
		idxInfo  *ColumnIndexInfo
		expected string
	}{
		{"native", &ColumnIndexInfo{IndexCategory: IndexCategoryNative}, "native"},
		{"custom", &ColumnIndexInfo{IndexCategory: IndexCategoryCustom}, "custom"},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.idxInfo.GetIndexCategoryString())
		})
	}
}

// ============================================================================
// TableMetadata Tests
// ============================================================================

func TestTableMetadata_GetPartitionKeyNames(t *testing.T) {
	table := &TableMetadata{
		PartitionKeys: []*ColumnInfo{
			{Name: "user_id"},
			{Name: "device_id"},
		},
	}

	names := table.GetPartitionKeyNames()
	assert.Equal(t, []string{"user_id", "device_id"}, names)

	// Test nil
	assert.Nil(t, (*TableMetadata)(nil).GetPartitionKeyNames())
}

func TestTableMetadata_GetClusteringKeyNames(t *testing.T) {
	table := &TableMetadata{
		ClusteringKeys: []*ColumnInfo{
			{Name: "year"},
			{Name: "month"},
			{Name: "day"},
		},
	}

	names := table.GetClusteringKeyNames()
	assert.Equal(t, []string{"year", "month", "day"}, names)
}

func TestTableMetadata_GetClusteringKeyOrders(t *testing.T) {
	table := &TableMetadata{
		ClusteringKeys: []*ColumnInfo{
			{Name: "year", ClusteringOrder: ColumnOrderDESC},
			{Name: "month", ClusteringOrder: ColumnOrderDESC},
			{Name: "day", ClusteringOrder: ColumnOrderASC},
		},
	}

	orders := table.GetClusteringKeyOrders()
	assert.Equal(t, []ColumnOrder{ColumnOrderDESC, ColumnOrderDESC, ColumnOrderASC}, orders)
}

func TestTableMetadata_GetRegularColumns(t *testing.T) {
	table := &TableMetadata{
		Columns: map[string]*ColumnInfo{
			"id":   {Name: "id", Kind: ColumnKindPartitionKey},
			"name": {Name: "name", Kind: ColumnKindRegular},
			"age":  {Name: "age", Kind: ColumnKindRegular},
			"ts":   {Name: "ts", Kind: ColumnKindClusteringKey},
		},
	}

	regularCols := table.GetRegularColumns()
	assert.Len(t, regularCols, 2)
}

func TestTableMetadata_GetCounterColumns(t *testing.T) {
	table := &TableMetadata{
		Columns: map[string]*ColumnInfo{
			"id":    {Name: "id", Type: &ColumnType{Name: "int"}},
			"views": {Name: "views", Type: &ColumnType{Name: "counter"}},
			"clicks": {Name: "clicks", Type: &ColumnType{Name: "counter"}},
		},
	}

	counterCols := table.GetCounterColumns()
	assert.Len(t, counterCols, 2)
}

func TestTableMetadata_GetUDTColumns(t *testing.T) {
	table := &TableMetadata{
		Columns: map[string]*ColumnInfo{
			"id":      {Name: "id", Type: &ColumnType{Name: "int", Category: TypeCategorySimple}},
			"address": {Name: "address", Type: &ColumnType{Name: "address_type", Category: TypeCategoryUDT}},
			"contact": {Name: "contact", Type: &ColumnType{Name: "contact_type", Category: TypeCategoryUDT}},
		},
	}

	udtCols := table.GetUDTColumns()
	assert.Len(t, udtCols, 2)
}

func TestTableMetadata_GetMapColumns(t *testing.T) {
	table := &TableMetadata{
		Columns: map[string]*ColumnInfo{
			"id":       {Name: "id", Type: &ColumnType{Name: "int"}},
			"settings": {Name: "settings", Type: &ColumnType{Name: "map<text, int>", Category: TypeCategoryCollection}},
			"tags":     {Name: "tags", Type: &ColumnType{Name: "set<text>", Category: TypeCategoryCollection}},
		},
	}

	mapCols := table.GetMapColumns()
	assert.Len(t, mapCols, 1)
	assert.Equal(t, "settings", mapCols[0].Name)
}

func TestTableMetadata_GetListColumns(t *testing.T) {
	table := &TableMetadata{
		Columns: map[string]*ColumnInfo{
			"id":     {Name: "id", Type: &ColumnType{Name: "int"}},
			"emails": {Name: "emails", Type: &ColumnType{Name: "list<text>", Category: TypeCategoryCollection}},
			"phones": {Name: "phones", Type: &ColumnType{Name: "list<text>", Category: TypeCategoryCollection}},
		},
	}

	listCols := table.GetListColumns()
	assert.Len(t, listCols, 2)
}

func TestTableMetadata_GetSetColumns(t *testing.T) {
	table := &TableMetadata{
		Columns: map[string]*ColumnInfo{
			"id":   {Name: "id", Type: &ColumnType{Name: "int"}},
			"tags": {Name: "tags", Type: &ColumnType{Name: "set<text>", Category: TypeCategoryCollection}},
		},
	}

	setCols := table.GetSetColumns()
	assert.Len(t, setCols, 1)
	assert.Equal(t, "tags", setCols[0].Name)
}

func TestTableMetadata_GetAllColumnsByKind(t *testing.T) {
	table := &TableMetadata{
		Columns: map[string]*ColumnInfo{
			"id":     {Name: "id", Kind: ColumnKindPartitionKey},
			"name":   {Name: "name", Kind: ColumnKindRegular},
			"age":    {Name: "age", Kind: ColumnKindRegular},
			"status": {Name: "status", Kind: ColumnKindStatic},
		},
	}

	regularCols := table.GetAllColumnsByKind(ColumnKindRegular)
	assert.Len(t, regularCols, 2)

	staticCols := table.GetAllColumnsByKind(ColumnKindStatic)
	assert.Len(t, staticCols, 1)
}

func TestTableMetadata_HasStaticColumns(t *testing.T) {
	tableWithStatic := &TableMetadata{
		StaticColumns: []*ColumnInfo{{Name: "status"}},
	}
	assert.True(t, tableWithStatic.HasStaticColumns())

	tableNoStatic := &TableMetadata{
		StaticColumns: []*ColumnInfo{},
	}
	assert.False(t, tableNoStatic.HasStaticColumns())

	assert.False(t, (*TableMetadata)(nil).HasStaticColumns())
}

func TestTableMetadata_FullName(t *testing.T) {
	table := &TableMetadata{
		Keyspace: "test_ks",
		Name:     "users",
	}

	assert.Equal(t, "test_ks.users", table.FullName())
	assert.Equal(t, "", (*TableMetadata)(nil).FullName())
}

func TestTableMetadata_IsSystemTable(t *testing.T) {
	tests := []struct {
		name     string
		keyspace string
		expected bool
	}{
		{"system", "system", true},
		{"system_auth", "system_auth", true},
		{"system_distributed", "system_distributed", true},
		{"system_schema", "system_schema", true},
		{"system_traces", "system_traces", true},
		{"user keyspace", "my_keyspace", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := &TableMetadata{Keyspace: tt.keyspace, Name: "test"}
			assert.Equal(t, tt.expected, table.IsSystemTable())
		})
	}
}

func TestTableMetadata_IsSystemVirtualTable(t *testing.T) {
	tests := []struct {
		name     string
		keyspace string
		expected bool
	}{
		{"system_views", "system_views", true},
		{"system_virtual_schema", "system_virtual_schema", true},
		{"system", "system", false},
		{"user keyspace", "my_keyspace", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := &TableMetadata{Keyspace: tt.keyspace, Name: "test"}
			assert.Equal(t, tt.expected, table.IsSystemVirtualTable())
		})
	}
}

func TestTableMetadata_GetAllColumns(t *testing.T) {
	table := &TableMetadata{
		PartitionKeys: []*ColumnInfo{
			{Name: "user_id", Kind: ColumnKindPartitionKey},
		},
		ClusteringKeys: []*ColumnInfo{
			{Name: "timestamp", Kind: ColumnKindClusteringKey},
		},
		Columns: map[string]*ColumnInfo{
			"user_id":   {Name: "user_id", Kind: ColumnKindPartitionKey},
			"timestamp": {Name: "timestamp", Kind: ColumnKindClusteringKey},
			"name":      {Name: "name", Kind: ColumnKindRegular},
			"email":     {Name: "email", Kind: ColumnKindRegular},
		},
	}

	allCols := table.GetAllColumns()
	assert.Len(t, allCols, 4)

	// Verify order: partition keys, then clustering keys, then regular
	assert.Equal(t, "user_id", allCols[0].Name)
	assert.Equal(t, "timestamp", allCols[1].Name)
	// Regular columns in positions 2-3
}

// ============================================================================
// KeyspaceMetadata Tests
// ============================================================================

func TestKeyspaceMetadata_GetTable(t *testing.T) {
	ks := &KeyspaceMetadata{
		Tables: map[string]*TableMetadata{
			"users": {Name: "users"},
			"posts": {Name: "posts"},
		},
	}

	assert.NotNil(t, ks.GetTable("users"))
	assert.Equal(t, "users", ks.GetTable("users").Name)
	assert.Nil(t, ks.GetTable("nonexistent"))
	assert.Nil(t, (*KeyspaceMetadata)(nil).GetTable("users"))
}

func TestKeyspaceMetadata_GetReplicationFactor(t *testing.T) {
	ks := &KeyspaceMetadata{
		Replication: &ReplicationStrategy{
			ReplicationFactor: 3,
		},
	}

	assert.Equal(t, 3, ks.GetReplicationFactor())
	assert.Equal(t, 0, (*KeyspaceMetadata)(nil).GetReplicationFactor())
}

func TestKeyspaceMetadata_IsSystemKeyspace(t *testing.T) {
	tests := []struct {
		name     string
		ksName   string
		expected bool
	}{
		{"system", "system", true},
		{"system_auth", "system_auth", true},
		{"SYSTEM uppercase", "SYSTEM", true},
		{"user keyspace", "my_keyspace", false},
		{"system_views", "system_views", false}, // Virtual, not regular system
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &KeyspaceMetadata{Name: tt.ksName}
			assert.Equal(t, tt.expected, ks.IsSystemKeyspace())
		})
	}
}

func TestKeyspaceMetadata_IsSystemVirtualKeyspace(t *testing.T) {
	tests := []struct {
		name     string
		ksName   string
		expected bool
	}{
		{"system_views", "system_views", true},
		{"system_virtual_schema", "system_virtual_schema", true},
		{"SYSTEM_VIEWS uppercase", "SYSTEM_VIEWS", true},
		{"system", "system", false},
		{"user keyspace", "my_keyspace", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &KeyspaceMetadata{Name: tt.ksName}
			assert.Equal(t, tt.expected, ks.IsSystemVirtualKeyspace())
		})
	}
}

func TestKeyspaceMetadata_GetTableCount(t *testing.T) {
	ks := &KeyspaceMetadata{
		Tables: map[string]*TableMetadata{
			"table1": {},
			"table2": {},
			"table3": {},
		},
	}

	assert.Equal(t, 3, ks.GetTableCount())
	assert.Equal(t, 0, (*KeyspaceMetadata)(nil).GetTableCount())
}

func TestKeyspaceMetadata_GetUserTypeCount(t *testing.T) {
	ks := &KeyspaceMetadata{
		UserTypes: map[string]*UserType{
			"address": {},
			"contact": {},
		},
	}

	assert.Equal(t, 2, ks.GetUserTypeCount())
}

func TestKeyspaceMetadata_GetFunctionCount(t *testing.T) {
	ks := &KeyspaceMetadata{
		Functions: map[string]*FunctionMetadata{
			"func1": {},
		},
	}

	assert.Equal(t, 1, ks.GetFunctionCount())
}

func TestKeyspaceMetadata_GetAggregateCount(t *testing.T) {
	ks := &KeyspaceMetadata{
		Aggregates: map[string]*AggregateMetadata{
			"agg1": {},
			"agg2": {},
		},
	}

	assert.Equal(t, 2, ks.GetAggregateCount())
}

func TestKeyspaceMetadata_GetMaterializedViewCount(t *testing.T) {
	ks := &KeyspaceMetadata{
		MaterializedViews: map[string]*MaterializedViewMetadata{
			"view1": {},
			"view2": {},
			"view3": {},
		},
	}

	assert.Equal(t, 3, ks.GetMaterializedViewCount())
}

// ============================================================================
// ClusterTopology Tests
// ============================================================================

func TestClusterTopology_GetHost(t *testing.T) {
	topology := &ClusterTopology{
		Hosts: map[string]*HostInfo{
			"host-1": {HostID: "host-1", DataCenter: "dc1"},
			"host-2": {HostID: "host-2", DataCenter: "dc2"},
		},
	}

	assert.NotNil(t, topology.GetHost("host-1"))
	assert.Equal(t, "dc1", topology.GetHost("host-1").DataCenter)
	assert.Nil(t, topology.GetHost("nonexistent"))
	assert.Nil(t, (*ClusterTopology)(nil).GetHost("host-1"))
}

func TestClusterTopology_GetHostsByDatacenter(t *testing.T) {
	topology := &ClusterTopology{
		DataCenters: map[string][]*HostInfo{
			"dc1": {
				{HostID: "h1", DataCenter: "dc1"},
				{HostID: "h2", DataCenter: "dc1"},
			},
			"dc2": {
				{HostID: "h3", DataCenter: "dc2"},
			},
		},
	}

	dc1Hosts := topology.GetHostsByDatacenter("dc1")
	assert.Len(t, dc1Hosts, 2)

	dc2Hosts := topology.GetHostsByDatacenter("dc2")
	assert.Len(t, dc2Hosts, 1)

	assert.Nil(t, topology.GetHostsByDatacenter("nonexistent"))
}

func TestClusterTopology_GetHostsByRack(t *testing.T) {
	topology := &ClusterTopology{
		DataCenters: map[string][]*HostInfo{
			"dc1": {
				{HostID: "h1", DataCenter: "dc1", Rack: "rack1"},
				{HostID: "h2", DataCenter: "dc1", Rack: "rack2"},
				{HostID: "h3", DataCenter: "dc1", Rack: "rack1"},
			},
		},
	}

	rack1Hosts := topology.GetHostsByRack("dc1", "rack1")
	assert.Len(t, rack1Hosts, 2)

	rack2Hosts := topology.GetHostsByRack("dc1", "rack2")
	assert.Len(t, rack2Hosts, 1)

	assert.Nil(t, topology.GetHostsByRack("dc1", "nonexistent"))
}

func TestClusterTopology_GetUpNodes(t *testing.T) {
	topology := &ClusterTopology{
		Hosts: map[string]*HostInfo{
			"h1": {State: HostStateUp},
			"h2": {State: HostStateDown},
			"h3": {State: HostStateUp},
			"h4": {State: HostStateUnknown},
		},
	}

	upNodes := topology.GetUpNodes()
	assert.Len(t, upNodes, 2)
}

func TestClusterTopology_GetDownNodes(t *testing.T) {
	topology := &ClusterTopology{
		Hosts: map[string]*HostInfo{
			"h1": {State: HostStateUp},
			"h2": {State: HostStateDown},
			"h3": {State: HostStateDown},
		},
	}

	downNodes := topology.GetDownNodes()
	assert.Len(t, downNodes, 2)
}

func TestClusterTopology_GetNodeCount(t *testing.T) {
	topology := &ClusterTopology{
		Hosts: map[string]*HostInfo{
			"h1": {},
			"h2": {},
			"h3": {},
		},
	}

	assert.Equal(t, 3, topology.GetNodeCount())
	assert.Equal(t, 0, (*ClusterTopology)(nil).GetNodeCount())
}

func TestClusterTopology_GetDatacenterCount(t *testing.T) {
	topology := &ClusterTopology{
		DataCenters: map[string][]*HostInfo{
			"dc1": {},
			"dc2": {},
			"dc3": {},
		},
	}

	assert.Equal(t, 3, topology.GetDatacenterCount())
}

func TestClusterTopology_IsMultiDC(t *testing.T) {
	multiDC := &ClusterTopology{
		DataCenters: map[string][]*HostInfo{
			"dc1": {},
			"dc2": {},
		},
	}
	assert.True(t, multiDC.IsMultiDC())

	singleDC := &ClusterTopology{
		DataCenters: map[string][]*HostInfo{
			"dc1": {},
		},
	}
	assert.False(t, singleDC.IsMultiDC())

	assert.False(t, (*ClusterTopology)(nil).IsMultiDC())
}

// ============================================================================
// ReplicationStrategy Tests
// ============================================================================

func TestReplicationStrategy_GetReplicationFactor(t *testing.T) {
	rs := &ReplicationStrategy{
		ReplicationFactor: 3,
	}

	assert.Equal(t, 3, rs.GetReplicationFactor())
	assert.Equal(t, 0, (*ReplicationStrategy)(nil).GetReplicationFactor())
}

func TestReplicationStrategy_GetDatacenterReplication(t *testing.T) {
	rs := &ReplicationStrategy{
		DataCenters: map[string]int{
			"dc1": 3,
			"dc2": 2,
		},
	}

	assert.Equal(t, 3, rs.GetDatacenterReplication("dc1"))
	assert.Equal(t, 2, rs.GetDatacenterReplication("dc2"))
	assert.Equal(t, 0, rs.GetDatacenterReplication("nonexistent"))
	assert.Equal(t, 0, (*ReplicationStrategy)(nil).GetDatacenterReplication("dc1"))
}
