package cluster

import (
	"context"
	"fmt"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// GocqlMetadataManager implements MetadataManager by delegating to gocql
//
// CRITICAL ARCHITECTURE:
// - This manager holds ONLY the gocql session and cluster config references
// - NEVER caches any metadata objects
// - ALWAYS calls session.KeyspaceMetadata() or queries system tables FRESH on every API call
// - gocql driver handles ALL caching and automatic schema change propagation (~1 second)
type GocqlMetadataManager struct {
	session *gocql.Session       // gocql session reference (ONLY state allowed)
	cluster *gocql.ClusterConfig // cluster config reference (static, OK to store)
	config  ManagerConfig        // our configuration (static, OK to store)
}

// Verify interface implementation at compile time
var _ MetadataManager = (*GocqlMetadataManager)(nil)

// NewGocqlMetadataManager creates a new metadata manager
//
// IMPORTANT: Pass in the SAME session and cluster used by the application
// The manager does not create its own session - it delegates to the provided one
func NewGocqlMetadataManager(session *gocql.Session, cluster *gocql.ClusterConfig, config ManagerConfig) *GocqlMetadataManager {
	return &GocqlMetadataManager{
		session: session,
		cluster: cluster,
		config:  config,
	}
}

// NewGocqlMetadataManagerWithDefaults creates manager with default config
func NewGocqlMetadataManagerWithDefaults(session *gocql.Session, cluster *gocql.ClusterConfig) *GocqlMetadataManager {
	return NewGocqlMetadataManager(session, cluster, DefaultManagerConfig())
}

// ============================================================================
// Cluster Operations
// ============================================================================

// GetClusterMetadata returns unified cluster and schema metadata
func (m *GocqlMetadataManager) GetClusterMetadata() (*ClusterMetadata, error) {
	// Get topology first
	topology, err := m.GetTopology()
	if err != nil {
		return nil, fmt.Errorf("failed to get topology: %w", err)
	}

	// Get all keyspaces
	keyspaceNames, err := m.GetKeyspaceNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace names: %w", err)
	}

	keyspaces := make(map[string]*KeyspaceMetadata)
	for _, ksName := range keyspaceNames {
		ksMeta, err := m.GetKeyspace(ksName)
		if err != nil {
			// Skip keyspaces we can't read
			continue
		}
		if ksMeta != nil {
			keyspaces[ksName] = ksMeta
		}
	}

	return &ClusterMetadata{
		Keyspaces: keyspaces,
		Topology:  topology,
	}, nil
}

// GetClusterName returns the cluster name
func (m *GocqlMetadataManager) GetClusterName() (string, error) {
	// ✅ Query system.local FRESH on every call
	var clusterName string
	query := "SELECT cluster_name FROM system.local"
	err := m.session.Query(query).Scan(&clusterName)
	if err != nil {
		return "", fmt.Errorf("failed to get cluster name: %w", err)
	}
	return clusterName, nil
}

// GetPartitioner returns the partitioner class
func (m *GocqlMetadataManager) GetPartitioner() (string, error) {
	// ✅ Query system.local FRESH on every call
	var partitioner string
	query := "SELECT partitioner FROM system.local"
	err := m.session.Query(query).Scan(&partitioner)
	if err != nil {
		return "", fmt.Errorf("failed to get partitioner: %w", err)
	}
	return partitioner, nil
}

// IsMultiDc checks if cluster has multiple datacenters
func (m *GocqlMetadataManager) IsMultiDc() (bool, error) {
	topology, err := m.GetTopology()
	if err != nil {
		return false, err
	}
	return topology.IsMultiDC(), nil
}

// ============================================================================
// Keyspace Operations
// ============================================================================

// GetKeyspace retrieves keyspace metadata
func (m *GocqlMetadataManager) GetKeyspace(keyspace string) (*KeyspaceMetadata, error) {
	// ✅ ALWAYS call session.KeyspaceMetadata() FRESH
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil // Keyspace not found
	}

	// Translate gocql metadata to our wrapper types
	return m.translateKeyspace(gocqlKs), nil
}

// GetKeyspaceNames returns all keyspace names
func (m *GocqlMetadataManager) GetKeyspaceNames() ([]string, error) {
	// ✅ Query system_schema.keyspaces FRESH on every call
	query := "SELECT keyspace_name FROM system_schema.keyspaces"
	iter := m.session.Query(query).Iter()

	var keyspaces []string
	var ksName string

	for iter.Scan(&ksName) {
		keyspaces = append(keyspaces, ksName)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get keyspace names: %w", err)
	}

	return keyspaces, nil
}

// RefreshKeyspace forces schema agreement and triggers metadata fetch
//
// This method:
// 1. Calls AwaitSchemaAgreement to ensure cluster schema is consistent
// 2. Calls session.KeyspaceMetadata() to trigger gocql to fetch/update metadata
//
// While gocql automatically keeps metadata updated via schema events (~1 second),
// this method is useful after DDL operations to immediately ensure metadata is current.
func (m *GocqlMetadataManager) RefreshKeyspace(ctx context.Context, keyspace string) error {
	// Wait for schema agreement across cluster
	if err := m.session.AwaitSchemaAgreement(ctx); err != nil {
		return fmt.Errorf("failed to await schema agreement: %w", err)
	}

	// Call KeyspaceMetadata to trigger gocql to fetch/update
	// We discard the result since subsequent GetKeyspace() calls will get it
	_, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return fmt.Errorf("failed to refresh keyspace metadata: %w", err)
	}

	return nil
}

// ============================================================================
// Table Operations
// ============================================================================

// GetTable retrieves table metadata
func (m *GocqlMetadataManager) GetTable(keyspace, table string) (*TableMetadata, error) {
	// ✅ Call KeyspaceMetadata() FRESH
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil // Keyspace not found
	}

	gocqlTable, ok := gocqlKs.Tables[table]
	if !ok {
		return nil, nil // Table not found
	}

	// Translate to our wrapper type
	return m.translateTable(gocqlTable), nil
}

// GetTableNames returns all table names in a keyspace
func (m *GocqlMetadataManager) GetTableNames(keyspace string) ([]string, error) {
	// ✅ Call KeyspaceMetadata() FRESH
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil // Keyspace not found
	}

	var tables []string
	for tableName := range gocqlKs.Tables {
		tables = append(tables, tableName)
	}

	return tables, nil
}

// TableExists checks if a table exists
func (m *GocqlMetadataManager) TableExists(keyspace, table string) (bool, error) {
	// ✅ Call KeyspaceMetadata() FRESH
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return false, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return false, nil
	}

	_, exists := gocqlKs.Tables[table]
	return exists, nil
}

// ============================================================================
// Index Operations
// ============================================================================

// FindIndexesByColumn returns all indexes on a specific column
func (m *GocqlMetadataManager) FindIndexesByColumn(keyspace, table, column string) ([]*ColumnIndexInfo, error) {
	tableMeta, err := m.GetTable(keyspace, table)
	if err != nil {
		return nil, err
	}

	if tableMeta == nil {
		return nil, fmt.Errorf("table %s.%s not found", keyspace, table)
	}

	colInfo, ok := tableMeta.Columns[column]
	if !ok {
		return nil, fmt.Errorf("column %s not found in table %s.%s", column, keyspace, table)
	}

	return colInfo.GetIndexesOnColumn(), nil
}

// FindTableIndexes returns all indexes on a table
func (m *GocqlMetadataManager) FindTableIndexes(keyspace, table string) ([]*ColumnIndexInfo, error) {
	tableMeta, err := m.GetTable(keyspace, table)
	if err != nil {
		return nil, err
	}

	if tableMeta == nil {
		return nil, fmt.Errorf("table %s.%s not found", keyspace, table)
	}

	var indexes []*ColumnIndexInfo
	for _, col := range tableMeta.Columns {
		if col.HasIndex() {
			indexes = append(indexes, col.Index)
		}
	}

	return indexes, nil
}

// FindVectorIndexes returns vector/SAI indexes on a table
func (m *GocqlMetadataManager) FindVectorIndexes(keyspace, table string) ([]*ColumnIndexInfo, error) {
	allIndexes, err := m.FindTableIndexes(keyspace, table)
	if err != nil {
		return nil, err
	}

	var vectorIndexes []*ColumnIndexInfo
	for _, idx := range allIndexes {
		if idx.IsVectorIndex {
			vectorIndexes = append(vectorIndexes, idx)
		}
	}

	return vectorIndexes, nil
}

// FindCustomIndexes returns custom indexes on a table
func (m *GocqlMetadataManager) FindCustomIndexes(keyspace, table string) ([]*ColumnIndexInfo, error) {
	allIndexes, err := m.FindTableIndexes(keyspace, table)
	if err != nil {
		return nil, err
	}

	var customIndexes []*ColumnIndexInfo
	for _, idx := range allIndexes {
		if idx.IndexCategory == IndexCategoryCustom {
			customIndexes = append(customIndexes, idx)
		}
	}

	return customIndexes, nil
}

// ============================================================================
// Schema Object Operations
// ============================================================================

// GetUserTypeNames returns all UDT names in a keyspace
func (m *GocqlMetadataManager) GetUserTypeNames(keyspace string) ([]string, error) {
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil
	}

	var types []string
	for typeName := range gocqlKs.UserTypes {
		types = append(types, typeName)
	}

	return types, nil
}

// GetUserType retrieves UDT metadata
func (m *GocqlMetadataManager) GetUserType(keyspace, typeName string) (*UserType, error) {
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil
	}

	gocqlType, ok := gocqlKs.UserTypes[typeName]
	if !ok {
		return nil, nil
	}

	return m.translateUserType(gocqlType), nil
}

// GetFunctionNames returns all function names in a keyspace
func (m *GocqlMetadataManager) GetFunctionNames(keyspace string) ([]string, error) {
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil
	}

	var funcs []string
	for funcName := range gocqlKs.Functions {
		funcs = append(funcs, funcName)
	}

	return funcs, nil
}

// GetFunction retrieves function metadata
func (m *GocqlMetadataManager) GetFunction(keyspace, funcName string) (*FunctionMetadata, error) {
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil
	}

	gocqlFunc, ok := gocqlKs.Functions[funcName]
	if !ok {
		return nil, nil
	}

	return m.translateFunction(gocqlFunc), nil
}

// GetAggregateNames returns all aggregate names in a keyspace
func (m *GocqlMetadataManager) GetAggregateNames(keyspace string) ([]string, error) {
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil
	}

	var aggs []string
	for aggName := range gocqlKs.Aggregates {
		aggs = append(aggs, aggName)
	}

	return aggs, nil
}

// GetAggregate retrieves aggregate metadata
func (m *GocqlMetadataManager) GetAggregate(keyspace, aggName string) (*AggregateMetadata, error) {
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil
	}

	gocqlAgg, ok := gocqlKs.Aggregates[aggName]
	if !ok {
		return nil, nil
	}

	return m.translateAggregate(gocqlAgg), nil
}

// GetMaterializedViewNames returns all materialized view names in a keyspace
func (m *GocqlMetadataManager) GetMaterializedViewNames(keyspace string) ([]string, error) {
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil
	}

	var views []string
	for viewName := range gocqlKs.MaterializedViews {
		views = append(views, viewName)
	}

	return views, nil
}

// GetMaterializedViewsForTable returns all materialized views on a specific table
func (m *GocqlMetadataManager) GetMaterializedViewsForTable(keyspace, table string) ([]*MaterializedViewMetadata, error) {
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil
	}

	var views []*MaterializedViewMetadata
	for _, gocqlView := range gocqlKs.MaterializedViews {
		// Check if view is based on the specified table
		if gocqlView.BaseTable != nil && gocqlView.BaseTable.Name == table {
			views = append(views, m.translateMaterializedView(gocqlView))
		}
	}

	return views, nil
}

// GetMaterializedView retrieves materialized view metadata
func (m *GocqlMetadataManager) GetMaterializedView(keyspace, viewName string) (*MaterializedViewMetadata, error) {
	gocqlKs, err := m.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	if gocqlKs == nil {
		return nil, nil
	}

	gocqlView, ok := gocqlKs.MaterializedViews[viewName]
	if !ok {
		return nil, nil
	}

	return m.translateMaterializedView(gocqlView), nil
}

// ============================================================================
// Topology Operations
// ============================================================================

// GetTopology returns cluster topology
func (m *GocqlMetadataManager) GetTopology() (*ClusterTopology, error) {
	// Get cluster name
	clusterName, err := m.GetClusterName()
	if err != nil {
		return nil, err
	}

	// Get partitioner
	partitioner, err := m.GetPartitioner()
	if err != nil {
		return nil, err
	}

	// Get local node
	var localHostID gocql.UUID
	var localDC, localRack, localVersion string

	query := "SELECT host_id, data_center, rack, release_version FROM system.local"
	err = m.session.Query(query).Scan(&localHostID, &localDC, &localRack, &localVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get local node info: %w", err)
	}

	// Get schema version
	var schemaVersion gocql.UUID
	err = m.session.Query("SELECT schema_version FROM system.local").Scan(&schemaVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema version: %w", err)
	}

	// Build topology
	topology := &ClusterTopology{
		ClusterName:    clusterName,
		Partitioner:    partitioner,
		Hosts:          make(map[string]*HostInfo),
		DataCenters:    make(map[string][]*HostInfo),
		SchemaVersions: []string{schemaVersion.String()},
	}

	// Add local node
	localHost := &HostInfo{
		HostID:         localHostID.String(),
		DataCenter:     localDC,
		Rack:           localRack,
		ReleaseVersion: localVersion,
		State:          HostStateUp, // Assume UP since we're connected
	}

	topology.Hosts[localHost.HostID] = localHost
	topology.DataCenters[localDC] = append(topology.DataCenters[localDC], localHost)

	// Get peer nodes
	peerQuery := "SELECT peer, data_center, rack, host_id, release_version FROM system.peers_v2"
	iter := m.session.Query(peerQuery).Iter()

	var peer, peerDC, peerRack, peerVersion string
	var peerHostID gocql.UUID

	for iter.Scan(&peer, &peerDC, &peerRack, &peerHostID, &peerVersion) {
		peerHost := &HostInfo{
			HostID:         peerHostID.String(),
			RpcAddress:     peer,
			DataCenter:     peerDC,
			Rack:           peerRack,
			ReleaseVersion: peerVersion,
			State:          HostStateUp, // Assume UP if in peers table
		}

		topology.Hosts[peerHost.HostID] = peerHost
		topology.DataCenters[peerDC] = append(topology.DataCenters[peerDC], peerHost)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get peer info: %w", err)
	}

	return topology, nil
}

// GetHost retrieves host by ID
func (m *GocqlMetadataManager) GetHost(hostID string) (*HostInfo, error) {
	topology, err := m.GetTopology()
	if err != nil {
		return nil, err
	}

	host := topology.GetHost(hostID)
	if host == nil {
		return nil, fmt.Errorf("host %s not found", hostID)
	}

	return host, nil
}

// GetHostByRpc retrieves host by RPC address
func (m *GocqlMetadataManager) GetHostByRpc(rpcAddress string) (*HostInfo, error) {
	topology, err := m.GetTopology()
	if err != nil {
		return nil, err
	}

	for _, host := range topology.Hosts {
		if host.RpcAddress == rpcAddress {
			return host, nil
		}
	}

	return nil, fmt.Errorf("host with RPC address %s not found", rpcAddress)
}

// GetUpNodes returns all UP nodes
func (m *GocqlMetadataManager) GetUpNodes() ([]*HostInfo, error) {
	topology, err := m.GetTopology()
	if err != nil {
		return nil, err
	}

	return topology.GetUpNodes(), nil
}

// ============================================================================
// Schema Agreement Operations
// ============================================================================

// WaitForSchemaAgreement waits for schema consistency
func (m *GocqlMetadataManager) WaitForSchemaAgreement(ctx context.Context) error {
	return m.session.AwaitSchemaAgreement(ctx)
}

// GetSchemaVersion returns current schema version(s)
func (m *GocqlMetadataManager) GetSchemaVersion() ([]string, error) {
	var schemaVersion gocql.UUID
	err := m.session.Query("SELECT schema_version FROM system.local").Scan(&schemaVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema version: %w", err)
	}

	// In normal operation, there's only one schema version
	// During schema changes, multiple versions might exist briefly
	return []string{schemaVersion.String()}, nil
}

// ============================================================================
// Validation Operations
// ============================================================================

// ValidateTableSchema validates table metadata structure
func (m *GocqlMetadataManager) ValidateTableSchema(keyspace, table string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Get table metadata
	tableMeta, err := m.GetTable(keyspace, table)
	if err != nil {
		return nil, err
	}

	if tableMeta == nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("table %s.%s not found", keyspace, table))
		return result, nil
	}

	// Check table name not empty
	if tableMeta.Name == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "table name is empty")
	}

	// Check at least one partition key exists
	if len(tableMeta.PartitionKeys) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "no partition keys defined")
	}

	// Check all partition key columns exist in columns map
	for i, pk := range tableMeta.PartitionKeys {
		if _, exists := tableMeta.Columns[pk.Name]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("partition key %s not found in columns", pk.Name))
		}

		// Check position matches array position
		if pk.ComponentIndex != i {
			result.Warnings = append(result.Warnings, fmt.Sprintf("partition key %s has ComponentIndex %d but is at position %d", pk.Name, pk.ComponentIndex, i))
		}

		// Check type is not nil
		if pk.Type == nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("partition key %s has nil type", pk.Name))
		}
	}

	// Check all clustering key columns exist in columns map
	for i, ck := range tableMeta.ClusteringKeys {
		if _, exists := tableMeta.Columns[ck.Name]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("clustering key %s not found in columns", ck.Name))
		}

		// Check position matches array position
		if ck.ComponentIndex != i {
			result.Warnings = append(result.Warnings, fmt.Sprintf("clustering key %s has ComponentIndex %d but is at position %d", ck.Name, ck.ComponentIndex, i))
		}

		// Check type is not nil
		if ck.Type == nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("clustering key %s has nil type", ck.Name))
		}
	}

	// Validate all column types
	for name, col := range tableMeta.Columns {
		if col.Type == nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("column %s has nil type", name))
			continue
		}

		// Validate vector columns
		if col.Type.IsVector {
			if col.Type.VectorDimension <= 0 {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("vector column %s has invalid dimension %d", name, col.Type.VectorDimension))
			}

			// TODO: Add Cassandra version check when version detection is implemented
			// For now, just warn
			if col.Type.VectorDimension > 0 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("vector column %s requires Cassandra 5.0+", name))
			}
		}

		// Validate SAI indexes
		if col.HasSAI() {
			result.Warnings = append(result.Warnings, fmt.Sprintf("SAI index on column %s requires Cassandra 5.0+", name))
		}
	}

	return result, nil
}

// ValidateColumnType validates column type structure
func (m *GocqlMetadataManager) ValidateColumnType(columnType *ColumnType) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	if columnType == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "column type is nil")
		return result
	}

	// Check type name not empty
	if columnType.Name == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "type name is empty")
	}

	// Validate collection types
	if columnType.Category == TypeCategoryCollection {
		if columnType.IsListType() || columnType.IsSetType() {
			if columnType.ElementType == nil {
				result.Valid = false
				result.Errors = append(result.Errors, "collection type missing element type")
			}
		}

		if columnType.IsMapType() {
			if columnType.KeyType == nil {
				result.Valid = false
				result.Errors = append(result.Errors, "map type missing key type")
			}
			if columnType.ValueType == nil {
				result.Valid = false
				result.Errors = append(result.Errors, "map type missing value type")
			}
		}
	}

	// Validate UDT types
	if columnType.Category == TypeCategoryUDT {
		if columnType.UDTName == "" {
			result.Valid = false
			result.Errors = append(result.Errors, "UDT type missing UDT name")
		}
	}

	// Validate tuple types
	if columnType.Category == TypeCategoryTuple {
		if len(columnType.TupleTypes) == 0 {
			result.Valid = false
			result.Errors = append(result.Errors, "tuple type has no element types")
		}
	}

	// Validate vector types
	if columnType.IsVector {
		if columnType.VectorDimension <= 0 {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("vector dimension %d is invalid (must be > 0)", columnType.VectorDimension))
		}

		if columnType.VectorDimension > 65535 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("vector dimension %d is very large", columnType.VectorDimension))
		}

		if columnType.VectorElementType != "float" && columnType.VectorElementType != "float64" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("vector element type %s is uncommon (expected float or float64)", columnType.VectorElementType))
		}
	}

	return result
}
