package cluster

import (
	"context"
)

// MetadataManager provides access to Cassandra cluster and schema metadata
//
// CRITICAL: All methods MUST delegate to gocql or system tables on EVERY call.
// NEVER cache metadata objects. The gocql driver handles all caching internally.
type MetadataManager interface {
	// ========================================================================
	// Cluster Operations
	// ========================================================================

	// GetClusterMetadata returns unified cluster and schema metadata
	GetClusterMetadata() (*ClusterMetadata, error)

	// GetClusterName returns the cluster name
	GetClusterName() (string, error)

	// GetPartitioner returns the partitioner class
	GetPartitioner() (string, error)

	// IsMultiDc checks if cluster has multiple datacenters
	IsMultiDc() (bool, error)

	// ========================================================================
	// Keyspace Operations
	// ========================================================================

	// GetKeyspace retrieves keyspace metadata
	// Returns nil if keyspace not found
	GetKeyspace(keyspace string) (*KeyspaceMetadata, error)

	// GetKeyspaceNames returns all keyspace names
	GetKeyspaceNames() ([]string, error)

	// RefreshKeyspace forces schema agreement for a keyspace
	// NOTE: gocql auto-refreshes metadata via schema events. This method
	// waits for schema agreement, then the next GetKeyspace() call will
	// have fresh metadata from gocql's internal cache.
	RefreshKeyspace(ctx context.Context, keyspace string) error

	// ========================================================================
	// Table Operations
	// ========================================================================

	// GetTable retrieves table metadata
	// Returns nil if table not found
	GetTable(keyspace, table string) (*TableMetadata, error)

	// GetTableNames returns all table names in a keyspace
	GetTableNames(keyspace string) ([]string, error)

	// TableExists checks if a table exists
	TableExists(keyspace, table string) (bool, error)

	// ========================================================================
	// Index Operations
	// ========================================================================

	// FindIndexesByColumn returns all indexes on a specific column
	FindIndexesByColumn(keyspace, table, column string) ([]*ColumnIndexInfo, error)

	// FindTableIndexes returns all indexes on a table
	FindTableIndexes(keyspace, table string) ([]*ColumnIndexInfo, error)

	// FindVectorIndexes returns vector/SAI indexes on a table
	FindVectorIndexes(keyspace, table string) ([]*ColumnIndexInfo, error)

	// FindCustomIndexes returns custom indexes on a table
	FindCustomIndexes(keyspace, table string) ([]*ColumnIndexInfo, error)

	// ========================================================================
	// Schema Object Operations
	// ========================================================================

	// GetUserTypeNames returns all UDT names in a keyspace
	GetUserTypeNames(keyspace string) ([]string, error)

	// GetUserType retrieves UDT metadata
	GetUserType(keyspace, typeName string) (*UserType, error)

	// GetFunctionNames returns all function names in a keyspace
	GetFunctionNames(keyspace string) ([]string, error)

	// GetFunction retrieves function metadata
	GetFunction(keyspace, funcName string) (*FunctionMetadata, error)

	// GetAggregateNames returns all aggregate names in a keyspace
	GetAggregateNames(keyspace string) ([]string, error)

	// GetAggregate retrieves aggregate metadata
	GetAggregate(keyspace, aggName string) (*AggregateMetadata, error)

	// GetMaterializedViewNames returns all materialized view names in a keyspace
	GetMaterializedViewNames(keyspace string) ([]string, error)

	// GetMaterializedViewsForTable returns all materialized views on a specific table
	GetMaterializedViewsForTable(keyspace, table string) ([]*MaterializedViewMetadata, error)

	// GetMaterializedView retrieves materialized view metadata
	GetMaterializedView(keyspace, viewName string) (*MaterializedViewMetadata, error)

	// ========================================================================
	// Topology Operations
	// ========================================================================

	// GetTopology returns cluster topology
	GetTopology() (*ClusterTopology, error)

	// GetHost retrieves host by ID
	GetHost(hostID string) (*HostInfo, error)

	// GetHostByRpc retrieves host by RPC address
	GetHostByRpc(rpcAddress string) (*HostInfo, error)

	// GetUpNodes returns all UP nodes
	GetUpNodes() ([]*HostInfo, error)

	// ========================================================================
	// Schema Agreement Operations
	// ========================================================================

	// WaitForSchemaAgreement waits for schema consistency
	WaitForSchemaAgreement(ctx context.Context) error

	// GetSchemaVersion returns current schema version(s)
	GetSchemaVersion() ([]string, error)

	// ========================================================================
	// Validation Operations
	// ========================================================================

	// ValidateTableSchema validates table metadata structure
	ValidateTableSchema(keyspace, table string) (*ValidationResult, error)

	// ValidateColumnType validates column type structure
	ValidateColumnType(columnType *ColumnType) *ValidationResult
}

// ManagerConfig contains configuration for the metadata manager
type ManagerConfig struct {
	// DetectVectorColumns enables special vector detection/validation
	DetectVectorColumns bool

	// DetectSAIIndexes enables special SAI index detection/validation
	DetectSAIIndexes bool
}

// DefaultManagerConfig returns default configuration
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		DetectVectorColumns: true,
		DetectSAIIndexes:    true,
	}
}
