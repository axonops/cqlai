package ui

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