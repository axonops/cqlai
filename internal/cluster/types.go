package cluster

import (
	"fmt"
	"strings"
)

// ============================================================================
// Enumerations
// ============================================================================

// ColumnKind represents the kind/role of a column in a table
type ColumnKind string

const (
	ColumnKindPartitionKey ColumnKind = "partition_key"
	ColumnKindClusteringKey ColumnKind = "clustering_key"
	ColumnKindRegular      ColumnKind = "regular"
	ColumnKindStatic       ColumnKind = "static"
	ColumnKindCompact      ColumnKind = "compact"
)

// ColumnOrder represents the ordering direction for clustering columns
type ColumnOrder string

const (
	ColumnOrderASC  ColumnOrder = "ASC"
	ColumnOrderDESC ColumnOrder = "DESC"
)

// TypeCategory represents the category of a column type
type TypeCategory string

const (
	TypeCategorySimple     TypeCategory = "simple"
	TypeCategoryCollection TypeCategory = "collection"
	TypeCategoryUDT        TypeCategory = "udt"
	TypeCategoryTuple      TypeCategory = "tuple"
	TypeCategoryCustom     TypeCategory = "custom"
	TypeCategoryVector     TypeCategory = "vector"
)

// ReplicationStrategyClass represents the replication strategy
type ReplicationStrategyClass string

const (
	ReplicationStrategySimple            ReplicationStrategyClass = "SimpleStrategy"
	ReplicationStrategyNetworkTopology   ReplicationStrategyClass = "NetworkTopologyStrategy"
	ReplicationStrategyLocal             ReplicationStrategyClass = "LocalStrategy"
	ReplicationStrategyEverywhere        ReplicationStrategyClass = "EverywhereStrategy"
)

// CompactionStrategyClass represents the compaction strategy
type CompactionStrategyClass string

const (
	CompactionStrategySizeTiered          CompactionStrategyClass = "SizeTieredCompactionStrategy"
	CompactionStrategyLeveled             CompactionStrategyClass = "LeveledCompactionStrategy"
	CompactionStrategyTimeWindow          CompactionStrategyClass = "TimeWindowCompactionStrategy"
	CompactionStrategyUnifiedCompaction   CompactionStrategyClass = "UnifiedCompactionStrategy"
)

// CompressionClass represents the compression algorithm
type CompressionClass string

const (
	CompressionClassLZ4      CompressionClass = "LZ4Compressor"
	CompressionClassSnappy   CompressionClass = "SnappyCompressor"
	CompressionClassDeflate  CompressionClass = "DeflateCompressor"
	CompressionClassZstd     CompressionClass = "ZstdCompressor"
)

// IndexKind represents the type of secondary index
type IndexKind string

const (
	IndexKindComposites IndexKind = "COMPOSITES"
	IndexKindKeys       IndexKind = "KEYS"
	IndexKindFull       IndexKind = "FULL"
	IndexKindCustom     IndexKind = "CUSTOM"
	IndexKindValues     IndexKind = "VALUES"
	IndexKindEntries    IndexKind = "ENTRIES"
	IndexKindSAI        IndexKind = "SAI"
	IndexKindVector     IndexKind = "VECTOR"
)

// IndexCategory represents whether index is native or custom
type IndexCategory string

const (
	IndexCategoryNative IndexCategory = "native"
	IndexCategoryCustom IndexCategory = "custom"
)

// HostState represents the state of a Cassandra node
type HostState string

const (
	HostStateUp      HostState = "UP"
	HostStateDown    HostState = "DOWN"
	HostStateUnknown HostState = "UNKNOWN"
)

// ValidationResult represents validation result with errors and warnings
type ValidationResult struct {
	Valid    bool     // True if validation passed
	Errors   []string // Critical errors
	Warnings []string // Non-critical warnings
}

// ============================================================================
// Core Metadata Types
// ============================================================================

// ColumnType represents detailed type information for a column
type ColumnType struct {
	Name                    string        // Type name (e.g., "int", "text", "list<int>")
	Category                TypeCategory  // Simple, Collection, UDT, Tuple, Custom, Vector
	IsNative                bool          // True if Cassandra native type
	IsFrozen                bool          // True if frozen<...>

	// Collection-specific
	ElementType             *ColumnType   // For list/set
	KeyType                 *ColumnType   // For map
	ValueType               *ColumnType   // For map

	// UDT-specific
	UDTKeyspace             string        // Keyspace containing the UDT
	UDTName                 string        // UDT type name
	UDTFields               map[string]*ColumnType  // UDT field types

	// Tuple-specific
	TupleTypes              []*ColumnType // Tuple element types

	// Vector-specific (Cassandra 5.0+)
	IsVector                bool          // True if vector type
	VectorDimension         int           // Vector dimension (e.g., 384, 768, 1536)
	VectorElementType       string        // Element type (float, float64)
	VectorSimilarityFunction string       // Similarity metric (euclidean, cosine, dot_product)
}

// ColumnIndexInfo represents index information for a column
type ColumnIndexInfo struct {
	Name            string        // Index name
	Kind            IndexKind     // Index type
	Target          string        // What's indexed
	Options         map[string]string  // Index options
	IndexClassName  string        // Full class name
	IsNativeIndex   bool          // True if built-in Cassandra index
	IsSAIIndex      bool          // True if Storage-Attached Index
	IsVectorIndex   bool          // True if vector ANN index
	IndexCategory   IndexCategory // Native or Custom
}

// ColumnInfo represents complete information about a column
type ColumnInfo struct {
	Name            string              // Column name
	Type            *ColumnType         // Column type details
	Kind            ColumnKind          // partition_key, clustering_key, regular, static
	ComponentIndex  int                 // Position in partition/clustering key (-1 for regular)
	ClusteringOrder ColumnOrder         // ASC or DESC (for clustering columns)
	Index           *ColumnIndexInfo    // Index information (nil if not indexed)
}

// TableOptions represents table-level options
type TableOptions struct {
	Comment             string
	ReadRepairChance    float64
	DcLocalReadRepairChance float64
	GcGraceSeconds      int
	BloomFilterFpChance float64
	Caching             map[string]string
	Compression         map[string]string
	Compaction          map[string]string
	DefaultTTL          int
	MinIndexInterval    int
	MaxIndexInterval    int
	MemtableFlushPeriod int
	SpeculativeRetry    string
}

// TableMetadata represents complete table metadata
type TableMetadata struct {
	Keyspace       string                  // Parent keyspace name
	Name           string                  // Table name
	Columns        map[string]*ColumnInfo  // All columns by name
	PartitionKeys  []*ColumnInfo           // Partition key columns (ordered)
	ClusteringKeys []*ColumnInfo           // Clustering key columns (ordered)
	StaticColumns  []*ColumnInfo           // Static columns
	Options        *TableOptions           // Table options
}

// ReplicationStrategy represents keyspace replication configuration
type ReplicationStrategy struct {
	Class           ReplicationStrategyClass
	ReplicationFactor int                  // For SimpleStrategy
	DataCenters     map[string]int         // For NetworkTopologyStrategy (DC name -> RF)
}

// UserType represents a user-defined type
type UserType struct {
	Keyspace    string
	Name        string
	Fields      map[string]*ColumnType  // Field name -> type
	FieldNames  []string                // Field names in definition order
}

// FunctionMetadata represents a user-defined function
type FunctionMetadata struct {
	Keyspace      string
	Name          string
	ArgumentNames []string
	ArgumentTypes []*ColumnType
	ReturnType    *ColumnType
	Language      string
	Body          string
	CalledOnNullInput bool
}

// AggregateMetadata represents a user-defined aggregate
type AggregateMetadata struct {
	Keyspace      string
	Name          string
	ArgumentTypes []*ColumnType
	StateFunction string
	StateType     *ColumnType
	FinalFunction string
	InitialCondition string
}

// MaterializedViewMetadata represents a materialized view
type MaterializedViewMetadata struct {
	Keyspace       string
	Name           string
	BaseTable      string
	Columns        map[string]*ColumnInfo
	PartitionKeys  []*ColumnInfo
	ClusteringKeys []*ColumnInfo
	WhereClause    string
	IncludeAllColumns bool
}

// KeyspaceMetadata represents complete keyspace metadata
type KeyspaceMetadata struct {
	Name              string
	DurableWrites     bool
	Replication       *ReplicationStrategy
	Tables            map[string]*TableMetadata
	UserTypes         map[string]*UserType
	Functions         map[string]*FunctionMetadata
	Aggregates        map[string]*AggregateMetadata
	MaterializedViews map[string]*MaterializedViewMetadata
}

// HostInfo represents information about a Cassandra node
type HostInfo struct {
	HostID         string       // UUID
	RpcAddress     string       // IP address
	BroadcastAddress string     // Broadcast IP
	DataCenter     string       // Datacenter name
	Rack           string       // Rack name
	ReleaseVersion string       // Cassandra version
	State          HostState    // UP, DOWN, UNKNOWN
	Tokens         []string     // Token ranges
}

// ClusterTopology represents cluster-wide topology information
type ClusterTopology struct {
	ClusterName    string
	Partitioner    string
	Hosts          map[string]*HostInfo  // Host ID -> HostInfo
	DataCenters    map[string][]*HostInfo  // DC -> hosts in DC
	SchemaVersions []string                // Active schema versions (normally 1)
}

// ClusterMetadata represents unified cluster and schema metadata
type ClusterMetadata struct {
	Keyspaces map[string]*KeyspaceMetadata
	Topology  *ClusterTopology
}

// ============================================================================
// Helper Methods on ColumnType
// ============================================================================

// IsNativeType checks if this is a Cassandra native type
func (ct *ColumnType) IsNativeType() bool {
	return ct != nil && ct.IsNative
}

// IsCollectionType checks if this is a collection type
func (ct *ColumnType) IsCollectionType() bool {
	return ct != nil && ct.Category == TypeCategoryCollection
}

// IsMapType checks if this is a map type
func (ct *ColumnType) IsMapType() bool {
	return ct != nil && ct.Category == TypeCategoryCollection &&
		strings.HasPrefix(strings.ToLower(ct.Name), "map<")
}

// IsListType checks if this is a list type
func (ct *ColumnType) IsListType() bool {
	return ct != nil && ct.Category == TypeCategoryCollection &&
		strings.HasPrefix(strings.ToLower(ct.Name), "list<")
}

// IsSetType checks if this is a set type
func (ct *ColumnType) IsSetType() bool {
	return ct != nil && ct.Category == TypeCategoryCollection &&
		strings.HasPrefix(strings.ToLower(ct.Name), "set<")
}

// IsUDTType checks if this is a UDT
func (ct *ColumnType) IsUDTType() bool {
	return ct != nil && ct.Category == TypeCategoryUDT
}

// IsCounterType checks if this is a counter type
func (ct *ColumnType) IsCounterType() bool {
	return ct != nil && strings.ToLower(ct.Name) == "counter"
}

// IsVectorType checks if this is a vector type
func (ct *ColumnType) IsVectorType() bool {
	return ct != nil && ct.IsVector
}

// GetVectorDimension returns vector dimension (0 if not a vector)
func (ct *ColumnType) GetVectorDimension() int {
	if ct != nil && ct.IsVector {
		return ct.VectorDimension
	}
	return 0
}

// GetVectorElementType returns vector element type (nil if not a vector)
func (ct *ColumnType) GetVectorElementType() string {
	if ct != nil && ct.IsVector {
		return ct.VectorElementType
	}
	return ""
}

// ============================================================================
// Helper Methods on ColumnInfo
// ============================================================================

// IsPartitionKey checks if this column is part of the partition key
func (ci *ColumnInfo) IsPartitionKey() bool {
	return ci != nil && ci.Kind == ColumnKindPartitionKey
}

// IsClusteringKey checks if this column is part of the clustering key
func (ci *ColumnInfo) IsClusteringKey() bool {
	return ci != nil && ci.Kind == ColumnKindClusteringKey
}

// IsRegular checks if this is a regular column
func (ci *ColumnInfo) IsRegular() bool {
	return ci != nil && ci.Kind == ColumnKindRegular
}

// IsStatic checks if this is a static column
func (ci *ColumnInfo) IsStatic() bool {
	return ci != nil && ci.Kind == ColumnKindStatic
}

// IsCounter checks if this is a counter column
func (ci *ColumnInfo) IsCounter() bool {
	return ci != nil && ci.Type != nil && ci.Type.IsCounterType()
}

// HasIndex checks if this column has any index
func (ci *ColumnInfo) HasIndex() bool {
	return ci != nil && ci.Index != nil && ci.Index.Name != ""
}

// HasVectorIndex checks if this column has a vector/SAI index
func (ci *ColumnInfo) HasVectorIndex() bool {
	return ci != nil && ci.Index != nil && ci.Index.IsVectorIndex
}

// HasSAI checks if this column has a SAI index
func (ci *ColumnInfo) HasSAI() bool {
	return ci != nil && ci.Index != nil && ci.Index.IsSAI()
}

// HasCustomIndex checks if this column has a custom index
func (ci *ColumnInfo) HasCustomIndex() bool {
	return ci != nil && ci.Index != nil && ci.Index.IndexCategory == IndexCategoryCustom
}

// IsVectorColumn checks if this is a vector column
func (ci *ColumnInfo) IsVectorColumn() bool {
	return ci != nil && ci.Type != nil && ci.Type.IsVectorType()
}

// GetIndexesOnColumn returns all indexes on this column (usually 0 or 1)
func (ci *ColumnInfo) GetIndexesOnColumn() []*ColumnIndexInfo {
	if ci == nil || ci.Index == nil || ci.Index.Name == "" {
		return nil
	}
	return []*ColumnIndexInfo{ci.Index}
}

// ============================================================================
// Helper Methods on ColumnIndexInfo
// ============================================================================

// IsNative checks if this is a native Cassandra index
func (cii *ColumnIndexInfo) IsNative() bool {
	return cii != nil && cii.IsNativeIndex
}

// IsSAI checks if this is a Storage-Attached Index
func (cii *ColumnIndexInfo) IsSAI() bool {
	return cii != nil && cii.IsSAIIndex
}

// IsVector checks if this is a vector index
func (cii *ColumnIndexInfo) IsVector() bool {
	return cii != nil && cii.IsVectorIndex
}

// IsCustom checks if this is a custom index
func (cii *ColumnIndexInfo) IsCustom() bool {
	return cii != nil && cii.IndexCategory == IndexCategoryCustom
}

// GetIndexCategoryString returns category as string
func (cii *ColumnIndexInfo) GetIndexCategoryString() string {
	if cii == nil {
		return ""
	}
	return string(cii.IndexCategory)
}

// ============================================================================
// Helper Methods on TableMetadata
// ============================================================================

// GetPrimaryKeyNames returns all primary key column names (partition + clustering)
func (tm *TableMetadata) GetPrimaryKeyNames() []string {
	if tm == nil {
		return nil
	}
	var names []string
	for _, pk := range tm.PartitionKeys {
		names = append(names, pk.Name)
	}
	for _, ck := range tm.ClusteringKeys {
		names = append(names, ck.Name)
	}
	return names
}

// GetPartitionKeyNames returns partition key column names in order
func (tm *TableMetadata) GetPartitionKeyNames() []string {
	if tm == nil {
		return nil
	}
	names := make([]string, len(tm.PartitionKeys))
	for i, pk := range tm.PartitionKeys {
		names[i] = pk.Name
	}
	return names
}

// GetClusteringKeyNames returns clustering key column names in order
func (tm *TableMetadata) GetClusteringKeyNames() []string {
	if tm == nil {
		return nil
	}
	names := make([]string, len(tm.ClusteringKeys))
	for i, ck := range tm.ClusteringKeys {
		names[i] = ck.Name
	}
	return names
}

// GetPartitionKeyTypes returns partition key types in order
func (tm *TableMetadata) GetPartitionKeyTypes() []*ColumnType {
	if tm == nil {
		return nil
	}
	types := make([]*ColumnType, len(tm.PartitionKeys))
	for i, pk := range tm.PartitionKeys {
		types[i] = pk.Type
	}
	return types
}

// GetClusteringKeyTypes returns clustering key types in order
func (tm *TableMetadata) GetClusteringKeyTypes() []*ColumnType {
	if tm == nil {
		return nil
	}
	types := make([]*ColumnType, len(tm.ClusteringKeys))
	for i, ck := range tm.ClusteringKeys {
		types[i] = ck.Type
	}
	return types
}

// GetClusteringKeyOrders returns ordering for clustering columns
func (tm *TableMetadata) GetClusteringKeyOrders() []ColumnOrder {
	if tm == nil {
		return nil
	}
	orders := make([]ColumnOrder, len(tm.ClusteringKeys))
	for i, ck := range tm.ClusteringKeys {
		orders[i] = ck.ClusteringOrder
	}
	return orders
}

// GetRegularColumns filters to regular columns only
func (tm *TableMetadata) GetRegularColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.IsRegular() {
			cols = append(cols, col)
		}
	}
	return cols
}

// GetCounterColumns filters to counter columns only
func (tm *TableMetadata) GetCounterColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.IsCounter() {
			cols = append(cols, col)
		}
	}
	return cols
}

// GetUDTColumns filters to UDT columns only
func (tm *TableMetadata) GetUDTColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.Type != nil && col.Type.IsUDTType() {
			cols = append(cols, col)
		}
	}
	return cols
}

// GetMapColumns filters to map columns only
func (tm *TableMetadata) GetMapColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.Type != nil && col.Type.IsMapType() {
			cols = append(cols, col)
		}
	}
	return cols
}

// GetListColumns filters to list columns only
func (tm *TableMetadata) GetListColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.Type != nil && col.Type.IsListType() {
			cols = append(cols, col)
		}
	}
	return cols
}

// GetSetColumns filters to set columns only
func (tm *TableMetadata) GetSetColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.Type != nil && col.Type.IsSetType() {
			cols = append(cols, col)
		}
	}
	return cols
}

// GetStaticColumns returns static columns
func (tm *TableMetadata) GetStaticColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	return tm.StaticColumns
}

// GetVectorColumns filters to vector columns only
func (tm *TableMetadata) GetVectorColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.IsVectorColumn() {
			cols = append(cols, col)
		}
	}
	return cols
}

// GetIndexedColumns filters to columns with any index
func (tm *TableMetadata) GetIndexedColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.HasIndex() {
			cols = append(cols, col)
		}
	}
	return cols
}

// GetSAIIndexedColumns filters to columns with SAI indexes
func (tm *TableMetadata) GetSAIIndexedColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.HasSAI() {
			cols = append(cols, col)
		}
	}
	return cols
}

// GetCustomIndexedColumns filters to columns with custom indexes
func (tm *TableMetadata) GetCustomIndexedColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.HasCustomIndex() {
			cols = append(cols, col)
		}
	}
	return cols
}

// HasStaticColumns checks if table has any static columns
func (tm *TableMetadata) HasStaticColumns() bool {
	return tm != nil && len(tm.StaticColumns) > 0
}

// HasCounterColumns checks if table has any counter columns
func (tm *TableMetadata) HasCounterColumns() bool {
	return tm != nil && len(tm.GetCounterColumns()) > 0
}

// HasVectorColumns checks if table has any vector columns
func (tm *TableMetadata) HasVectorColumns() bool {
	return tm != nil && len(tm.GetVectorColumns()) > 0
}

// HasSAIIndexes checks if table has any SAI indexes
func (tm *TableMetadata) HasSAIIndexes() bool {
	return tm != nil && len(tm.GetSAIIndexedColumns()) > 0
}

// HasCustomIndexes checks if table has any custom indexes
func (tm *TableMetadata) HasCustomIndexes() bool {
	return tm != nil && len(tm.GetCustomIndexedColumns()) > 0
}

// GetAllColumnsByKind filters columns by kind
func (tm *TableMetadata) GetAllColumnsByKind(kind ColumnKind) []*ColumnInfo {
	if tm == nil {
		return nil
	}
	var cols []*ColumnInfo
	for _, col := range tm.Columns {
		if col.Kind == kind {
			cols = append(cols, col)
		}
	}
	return cols
}

// HasCompactStorage checks if table uses compact storage
func (tm *TableMetadata) HasCompactStorage() bool {
	// Compact storage is a legacy feature
	// Check if any column has ColumnKindCompact
	if tm == nil {
		return false
	}
	for _, col := range tm.Columns {
		if col.Kind == ColumnKindCompact {
			return true
		}
	}
	return false
}

// GetCompactionClass returns compaction strategy class
func (tm *TableMetadata) GetCompactionClass() string {
	if tm == nil || tm.Options == nil || tm.Options.Compaction == nil {
		return ""
	}
	return tm.Options.Compaction["class"]
}

// GetCompressionClass returns compression algorithm class
func (tm *TableMetadata) GetCompressionClass() string {
	if tm == nil || tm.Options == nil || tm.Options.Compression == nil {
		return ""
	}
	return tm.Options.Compression["class"]
}

// GetTableOptions returns table options
func (tm *TableMetadata) GetTableOptions() *TableOptions {
	if tm == nil {
		return nil
	}
	return tm.Options
}

// GetAllColumns returns all columns in definition order (partition keys, clustering keys, then regular/static)
func (tm *TableMetadata) GetAllColumns() []*ColumnInfo {
	if tm == nil {
		return nil
	}

	var cols []*ColumnInfo

	// Add partition keys first (in order)
	cols = append(cols, tm.PartitionKeys...)

	// Add clustering keys (in order)
	cols = append(cols, tm.ClusteringKeys...)

	// Add regular and static columns
	for _, col := range tm.Columns {
		if !col.IsPartitionKey() && !col.IsClusteringKey() {
			cols = append(cols, col)
		}
	}

	return cols
}

// GetReplicationFactor returns replication factor from parent keyspace
// Note: This requires keyspace metadata to be available
// Returns 0 if not available (caller must fetch keyspace metadata separately)
func (tm *TableMetadata) GetReplicationFactor() int {
	// This method can't access parent keyspace without circular reference
	// Caller should use: manager.GetKeyspace(tm.Keyspace).GetReplicationFactor()
	return 0
}

// FullName returns "keyspace.table"
func (tm *TableMetadata) FullName() string {
	if tm == nil {
		return ""
	}
	return fmt.Sprintf("%s.%s", tm.Keyspace, tm.Name)
}

// IsSystemTable checks if this is a system table
func (tm *TableMetadata) IsSystemTable() bool {
	if tm == nil {
		return false
	}
	ks := strings.ToLower(tm.Keyspace)
	return ks == "system" || ks == "system_auth" ||
		ks == "system_distributed" || ks == "system_schema" || ks == "system_traces"
}

// IsSystemVirtualTable checks if this is a system virtual table
func (tm *TableMetadata) IsSystemVirtualTable() bool {
	if tm == nil {
		return false
	}
	ks := strings.ToLower(tm.Keyspace)
	return ks == "system_views" || ks == "system_virtual_schema"
}

// ============================================================================
// Helper Methods on KeyspaceMetadata
// ============================================================================

// GetTable returns table by name or nil
func (km *KeyspaceMetadata) GetTable(name string) *TableMetadata {
	if km == nil || km.Tables == nil {
		return nil
	}
	return km.Tables[name]
}

// GetUserType returns UDT by name or nil
func (km *KeyspaceMetadata) GetUserType(name string) *UserType {
	if km == nil || km.UserTypes == nil {
		return nil
	}
	return km.UserTypes[name]
}

// GetFunction returns function by name or nil
func (km *KeyspaceMetadata) GetFunction(name string) *FunctionMetadata {
	if km == nil || km.Functions == nil {
		return nil
	}
	return km.Functions[name]
}

// GetAggregate returns aggregate by name or nil
func (km *KeyspaceMetadata) GetAggregate(name string) *AggregateMetadata {
	if km == nil || km.Aggregates == nil {
		return nil
	}
	return km.Aggregates[name]
}

// GetMaterializedView returns materialized view by name or nil
func (km *KeyspaceMetadata) GetMaterializedView(name string) *MaterializedViewMetadata {
	if km == nil || km.MaterializedViews == nil {
		return nil
	}
	return km.MaterializedViews[name]
}

// GetReplicationFactor returns replication factor
func (km *KeyspaceMetadata) GetReplicationFactor() int {
	if km == nil || km.Replication == nil {
		return 0
	}
	return km.Replication.ReplicationFactor
}

// IsSystemKeyspace checks if this is a system keyspace
func (km *KeyspaceMetadata) IsSystemKeyspace() bool {
	if km == nil {
		return false
	}
	ks := strings.ToLower(km.Name)
	return ks == "system" || ks == "system_auth" ||
		ks == "system_distributed" || ks == "system_schema" || ks == "system_traces"
}

// IsSystemVirtualKeyspace checks if this is a system virtual keyspace
func (km *KeyspaceMetadata) IsSystemVirtualKeyspace() bool {
	if km == nil {
		return false
	}
	ks := strings.ToLower(km.Name)
	return ks == "system_views" || ks == "system_virtual_schema"
}

// GetTableCount returns number of tables
func (km *KeyspaceMetadata) GetTableCount() int {
	if km == nil {
		return 0
	}
	return len(km.Tables)
}

// GetUserTypeCount returns number of user types
func (km *KeyspaceMetadata) GetUserTypeCount() int {
	if km == nil {
		return 0
	}
	return len(km.UserTypes)
}

// GetFunctionCount returns number of functions
func (km *KeyspaceMetadata) GetFunctionCount() int {
	if km == nil {
		return 0
	}
	return len(km.Functions)
}

// GetAggregateCount returns number of aggregates
func (km *KeyspaceMetadata) GetAggregateCount() int {
	if km == nil {
		return 0
	}
	return len(km.Aggregates)
}

// GetMaterializedViewCount returns number of materialized views
func (km *KeyspaceMetadata) GetMaterializedViewCount() int {
	if km == nil {
		return 0
	}
	return len(km.MaterializedViews)
}

// ============================================================================
// Helper Methods on ClusterTopology
// ============================================================================

// GetHost returns host by ID or nil
func (ct *ClusterTopology) GetHost(hostID string) *HostInfo {
	if ct == nil || ct.Hosts == nil {
		return nil
	}
	return ct.Hosts[hostID]
}

// GetHostByRpc returns host by RPC address or nil
func (ct *ClusterTopology) GetHostByRpc(rpcAddress string) *HostInfo {
	if ct == nil {
		return nil
	}
	for _, host := range ct.Hosts {
		if host.RpcAddress == rpcAddress {
			return host
		}
	}
	return nil
}

// GetHostsByDatacenter returns all hosts in a specific datacenter
func (ct *ClusterTopology) GetHostsByDatacenter(dc string) []*HostInfo {
	if ct == nil || ct.DataCenters == nil {
		return nil
	}
	return ct.DataCenters[dc]
}

// GetHostsByRack returns all hosts in a specific datacenter and rack
func (ct *ClusterTopology) GetHostsByRack(dc, rack string) []*HostInfo {
	if ct == nil {
		return nil
	}
	dcHosts := ct.GetHostsByDatacenter(dc)
	var rackHosts []*HostInfo
	for _, host := range dcHosts {
		if host.Rack == rack {
			rackHosts = append(rackHosts, host)
		}
	}
	return rackHosts
}

// GetReplicasByToken finds replica set for a given token
// Note: This requires token ring calculation which is complex
// For now, returns nil (to be implemented when token awareness is needed)
func (ct *ClusterTopology) GetReplicasByToken(token string) []*HostInfo {
	// TODO: Implement token ring calculation
	// This requires understanding partitioner and token ranges
	return nil
}

// GetReplicasByTokenRange finds replicas for a token range
// Note: This requires token ring calculation which is complex
// For now, returns nil (to be implemented when token awareness is needed)
func (ct *ClusterTopology) GetReplicasByTokenRange(start, end string) []*HostInfo {
	// TODO: Implement token range calculation
	return nil
}

// GetTokenRange returns token ranges assigned to a specific host
// Note: Returns the Tokens field from HostInfo
func (ct *ClusterTopology) GetTokenRange(hostID string) []string {
	host := ct.GetHost(hostID)
	if host == nil {
		return nil
	}
	return host.Tokens
}

// GetUpNodes filters to UP nodes only
func (ct *ClusterTopology) GetUpNodes() []*HostInfo {
	if ct == nil {
		return nil
	}
	var hosts []*HostInfo
	for _, host := range ct.Hosts {
		if host.State == HostStateUp {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

// GetDownNodes filters to DOWN nodes only
func (ct *ClusterTopology) GetDownNodes() []*HostInfo {
	if ct == nil {
		return nil
	}
	var hosts []*HostInfo
	for _, host := range ct.Hosts {
		if host.State == HostStateDown {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

// GetNodeCount returns total node count
func (ct *ClusterTopology) GetNodeCount() int {
	if ct == nil {
		return 0
	}
	return len(ct.Hosts)
}

// GetUpNodeCount returns count of UP nodes
func (ct *ClusterTopology) GetUpNodeCount() int {
	return len(ct.GetUpNodes())
}

// GetDatacenterCount returns number of datacenters
func (ct *ClusterTopology) GetDatacenterCount() int {
	if ct == nil {
		return 0
	}
	return len(ct.DataCenters)
}

// IsMultiDC checks if cluster has multiple datacenters
func (ct *ClusterTopology) IsMultiDC() bool {
	return ct != nil && ct.GetDatacenterCount() > 1
}

// GetSchemaVersion returns schema version(s)
func (ct *ClusterTopology) GetSchemaVersion() []string {
	if ct == nil {
		return nil
	}
	return ct.SchemaVersions
}

// ============================================================================
// Helper Methods on ReplicationStrategy
// ============================================================================

// GetReplicationFactor returns replication factor for SimpleStrategy
func (rs *ReplicationStrategy) GetReplicationFactor() int {
	if rs == nil {
		return 0
	}
	return rs.ReplicationFactor
}

// GetDatacenterReplication returns replication factor for a specific datacenter
// Used with NetworkTopologyStrategy
func (rs *ReplicationStrategy) GetDatacenterReplication(dc string) int {
	if rs == nil || rs.DataCenters == nil {
		return 0
	}
	return rs.DataCenters[dc]
}
