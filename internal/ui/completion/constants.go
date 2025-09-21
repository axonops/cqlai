package completion

// CQL Keywords and Constants used across completion files
// This file centralizes all keyword lists to avoid duplication

// Permissions for GRANT/REVOKE commands
var CQLPermissions = []string{
	"ALL",
	"ALTER",
	"AUTHORIZE",
	"CREATE",
	"DESCRIBE",
	"DROP",
	"EXECUTE",
	"MODIFY",
	"SELECT",
	"INSERT",
	"UPDATE",
	"DELETE",
	"TRUNCATE",
}

// Comparison operators for WHERE clauses
var ComparisonOperators = []string{
	"=", "!=", "<", ">", "<=", ">=", "IN", "CONTAINS", "CONTAINS KEY",
}

// Logical operators
var LogicalOperators = []string{"AND", "OR"}

// Sort orders for ORDER BY
var SortOrders = []string{"ASC", "DESC"}

// Consistency levels
var ConsistencyLevels = []string{
	"ALL",
	"EACH_QUORUM",
	"QUORUM",
	"LOCAL_QUORUM",
	"ONE",
	"TWO",
	"THREE",
	"LOCAL_ONE",
	"ANY",
	"SERIAL",
	"LOCAL_SERIAL",
}

// Data types for CREATE TABLE
var CQLDataTypes = []string{
	"ascii", "bigint", "blob", "boolean", "counter",
	"date", "decimal", "double", "float", "frozen",
	"inet", "int", "list", "map", "set",
	"smallint", "text", "time", "timestamp", "timeuuid",
	"tinyint", "tuple", "uuid", "varchar", "varint",
}

// Aggregate functions
var AggregateFunctions = []string{
	"COUNT", "MAX", "MIN", "AVG", "SUM",
}

// Time/UUID functions
var TimeFunctions = []string{
	"now", "currentTimeUUID", "currentTimestamp", "currentDate",
	"minTimeuuid", "maxTimeuuid", "toDate", "toTimestamp",
	"toUnixTimestamp", "dateOf", "unixTimestampOf",
}

// Other system functions
var SystemFunctions = []string{
	"token", "uuid", "blobAsText", "textAsBlob",
	"blobAsBigint", "bigintAsBlob", "TTL", "WRITETIME",
}

// DDL object types
var DDLObjectTypes = []string{
	"TABLE", "KEYSPACE", "INDEX", "TYPE",
	"FUNCTION", "AGGREGATE", "MATERIALIZED",
	"ROLE", "USER", "TRIGGER",
}

// Common table options for WITH clause
var TableOptions = []string{
	"bloom_filter_fp_chance",
	"caching",
	"comment",
	"compaction",
	"compression",
	"crc_check_chance",
	"dclocal_read_repair_chance",
	"default_time_to_live",
	"gc_grace_seconds",
	"max_index_interval",
	"memtable_flush_period_in_ms",
	"min_index_interval",
	"read_repair_chance",
	"speculative_retry",
}

// Replication strategies
var ReplicationStrategies = []string{
	"SimpleStrategy",
	"NetworkTopologyStrategy",
}

// TopLevelCommands are the main CQL and meta commands
var TopLevelCommands = []string{
	"SELECT", "INSERT", "UPDATE", "DELETE",
	"CREATE", "DROP", "ALTER", "TRUNCATE",
	"GRANT", "REVOKE",
	"USE",
	"DESCRIBE", "DESC",
	"BEGIN", "APPLY",
	"LIST",
	"CONSISTENCY",
	"OUTPUT",
	"TRACING",
	"PAGING",
	"AUTOFETCH",
	"SHOW",
	"HELP",
	"SOURCE",
	"CAPTURE",
	"EXPAND",
	"COPY",
}

// DescribeObjects are the objects that can be described
var DescribeObjects = []string{
	"KEYSPACE", "KEYSPACES",
	"TABLE", "TABLES",
	"TYPE", "TYPES",
	"FUNCTION", "FUNCTIONS",
	"AGGREGATE", "AGGREGATES",
	"MATERIALIZED",
	"INDEX",
	"SCHEMA",
	"CLUSTER",
}

// ResourceTypes for GRANT/REVOKE ON clause
var ResourceTypes = []string{
	"ALL", "KEYSPACE", "TABLE", "ROLE",
	"FUNCTION", "AGGREGATE", "INDEX", "MATERIALIZED",
}

// ShowCommands for SHOW command completions
var ShowCommands = []string{
	"VERSION", "HOST", "SESSION",
}

// OutputFormats for OUTPUT command
var OutputFormats = []string{
	"ASCII", "TABLE", "EXPAND", "JSON",
}

// CopyDirections for COPY command
var CopyDirections = []string{
	"TO", "FROM",
}

// CopyFileSuggestions for COPY command file paths
var CopyFileSuggestions = []string{
	"'/tmp/export.csv'",
	"'/tmp/export.parquet'",
	"'/tmp/export.json'",
	"'/home/user/data.csv'",
	"'/home/user/data.parquet'",
	"'/data/partitioned/export/'", // For partitioned datasets
	"'./export.csv'",
	"'./export.parquet'",
	"'s3://bucket/path/data.parquet'", // Cloud storage example
}

// CopyOptions for COPY command WITH clause
var CopyOptions = []string{
	// CSV options
	"HEADER", "DELIMITER", "NULLVAL", "PAGESIZE", "ENCODING", "QUOTE",
	"MAXROWS", "SKIPROWS", "MAXPARSEERRORS", "MAXINSERTERRORS",
	"MAXBATCHSIZE", "MINBATCHSIZE", "CHUNKSIZE",
	// Parquet options
	"FORMAT", "COMPRESSION", "PARTITION", "PARTITION_FILTER", "MAX_FILE_SIZE",
}

// CopyFormats for COPY command FORMAT option
var CopyFormats = []string{
	"PARQUET", "CSV", "JSON",
}

// ParquetCompressionTypes for COPY command COMPRESSION option
var ParquetCompressionTypes = []string{
	"SNAPPY", "GZIP", "ZSTD", "LZ4", "NONE",
}

// BatchTypes for BEGIN command
var BatchTypes = []string{
	"BATCH", "UNLOGGED", "COUNTER",
}

// ListTargets for LIST command
var ListTargets = []string{
	"USERS", "ROLES", "PERMISSIONS",
}

// IfClause keywords
var IfClauseKeywords = []string{
	"NOT", "EXISTS",
}

// UsingOptions for INSERT/UPDATE/DELETE
var UsingOptions = []string{
	"TTL", "TIMESTAMP",
}

// SelectKeywords for SELECT clause
var SelectKeywords = []string{
	"*", "DISTINCT", "JSON",
}

// AlterTableOperations for ALTER TABLE
var AlterTableOperations = []string{
	"ADD", "DROP", "ALTER", "RENAME", "WITH",
}

// AlterTypeOperations for ALTER TYPE
var AlterTypeOperations = []string{
	"ADD", "RENAME",
}

// AllResourceTarget for GRANT/REVOKE ON ALL
var AllResourceTargets = []string{
	"KEYSPACES", "FUNCTIONS", "ROLES",
}

// MaterializedKeyword
var MaterializedKeyword = []string{
	"VIEW",
}

// MaterializedViews plural
var MaterializedViews = []string{
	"VIEW", "VIEWS",
}

// Individual keyword constants for safer access
var WithKeyword = []string{"WITH"}

// CommonDataManipulationKeywords
var FilteringKeyword = []string{
	"FILTERING",
}

// LimitKeywords
var LimitKeywords = []string{
	"LIMIT", "PARTITION",
}

// KeyspaceOptions for WITH clause
var KeyspaceOptions = []string{
	"REPLICATION", "DURABLE_WRITES", "CLUSTERING", "COMPACT", "COMPRESSION",
}

// CreateObjectTypes for CREATE command
var CreateObjectTypes = []string{
	"AGGREGATE", "FUNCTION", "INDEX", "KEYSPACE", "MATERIALIZED",
	"ROLE", "TABLE", "TRIGGER", "TYPE", "USER",
}

// AlterObjectTypes for ALTER command
var AlterObjectTypes = []string{
	"KEYSPACE", "MATERIALIZED", "ROLE", "TABLE", "TYPE", "USER",
}

// CreateDropObjectTypes for CREATE/DROP commands (without MATERIALIZED which needs VIEW after it)
var CreateDropObjectTypes = []string{
	"KEYSPACE", "TABLE", "INDEX", "TYPE", "FUNCTION", "AGGREGATE", "MATERIALIZED", "ROLE", "USER",
}

// CreateDropObjectTypesNoMaterialized for checking in switch statements
var CreateDropObjectTypesNoMaterialized = []string{
	"KEYSPACE", "TABLE", "INDEX", "TYPE", "FUNCTION", "AGGREGATE", "ROLE", "USER",
}

// TableKeyword for TRUNCATE TABLE
var TableKeyword = []string{"TABLE"}

// SingleKeywords for various single keyword returns
var ByKeyword = []string{"BY"}
var FromKeyword = []string{"FROM"}
var IntoKeyword = []string{"INTO"}
var SetKeyword = []string{"SET"}
var WhereKeyword = []string{"WHERE"}
var ValuesKeyword = []string{"VALUES"}
var ExistsKeyword = []string{"EXISTS"}
var NotKeyword = []string{"NOT"}
var KeyKeyword = []string{"KEY"}
var OnKeyword = []string{"ON"}
var LimitKeyword = []string{"LIMIT"}
var PartitionKeyword = []string{"PARTITION"}
var TimestampKeyword = []string{"TIMESTAMP"}
var IfKeyword = []string{"IF"}
var ToKeyword = []string{"TO"}
var AsKeyword = []string{"AS"}
var OrKeyword = []string{"OR"}

// SelectCompletionKeywords
var FromCommaAs = []string{"FROM", ",", "AS"}
var AscDescComma = []string{"ASC", "DESC", ","}

// UpdateCompletionKeywords
var SetUsing = []string{"SET", "USING"}
var WhereIf = []string{"WHERE", "IF"}
var IfAnd = []string{"IF", "AND"}
var AndKeyword = []string{"AND"}

// DeleteCompletionKeywords
var WhereUsingIf = []string{"WHERE", "USING", "IF"}

// RBACPermissions for more granular permissions
var RBACPermissions = []string{"ALL", "SELECT", "MODIFY", "CREATE", "ALTER", "DROP", "AUTHORIZE"}

// RBACResourceTypes
var RBACResourceTypes = []string{"KEYSPACE", "TABLE", "ROLE", "ALL"}

// AlterSpecificTypes for ALTER command (subset)
var AlterSpecificTypes = []string{"TABLE", "KEYSPACE", "TYPE", "ROLE", "USER"}

// DescribeObjectsBasic for parser-based describe (subset)
var DescribeObjectsBasic = []string{
	"KEYSPACE",
	"KEYSPACES",
	"TABLE",
	"TABLES",
	"TYPE",
	"TYPES",
	"FUNCTION",
	"FUNCTIONS",
	"AGGREGATE",
	"AGGREGATES",
	"MATERIALIZED",
}

// TopLevelKeywords for parser-based completion (alphabetical)
var TopLevelKeywords = []string{
	"ALTER",
	"APPLY",
	"BEGIN",
	"CAPTURE",
	"CONSISTENCY",
	"COPY",
	"CREATE",
	"DELETE",
	"DESCRIBE",
	"DESC",
	"DROP",
	"EXPAND",
	"GRANT",
	"HELP",
	"INSERT",
	"LIST",
	"OUTPUT",
	"PAGING",
	"AUTOFETCH",
	"REVOKE",
	"SELECT",
	"SHOW",
	"SOURCE",
	"TRACING",
	"TRUNCATE",
	"UPDATE",
	"USE",
}
