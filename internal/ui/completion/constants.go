package completion

import "github.com/axonops/cqlai/internal/db"

// CQL Keywords and Constants used across completion files
// This file centralizes all keyword lists to avoid duplication

// Permissions for GRANT/REVOKE commands
// CQLPermissions are what GRANT and REVOKE take.
//
// From Cassandra's Permission enum. INSERT, UPDATE, DELETE and TRUNCATE were
// offered here and are not permissions - MODIFY is the one that covers writing -
// so GRANT completed to four statements the server refuses.
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
	"SELECT_MASKED", // reading a table that has masked columns restricted
	"UNMASK",        // seeing what is behind a mask
}

// ComparisonOperators are what a WHERE clause puts between a column and a value.
//
// BETWEEN arrived in Cassandra 5.0, and the negated forms of IN and CONTAINS
// with it.
var ComparisonOperators = []string{
	"=", "!=", "<", ">", "<=", ">=",
	"IN", "NOT IN",
	"CONTAINS", "CONTAINS KEY",
	"NOT CONTAINS", "NOT CONTAINS KEY",
	"BETWEEN",
}

// Logical operators
var LogicalOperators = []string{"AND", "OR"}

// Sort orders for ORDER BY
var SortOrders = []string{"ASC", "DESC"}

// Consistency levels
// ConsistencyLevels are the levels CONSISTENCY accepts. Completing a level that
// cannot then be set is worse than not completing it at all, so this comes from
// the same table that applies them.
var ConsistencyLevels = db.ConsistencyLevels()

// Data types for CREATE TABLE
// CQLDataTypes are the types a column can be given.
//
// The native ones are Cassandra's CQL3Type.Native, less `empty`, which exists
// and is not a thing to declare a column as. The rest - frozen, list, map, set,
// tuple, vector - take parameters and are completed as the bare word.
var CQLDataTypes = []string{
	"ascii", "bigint", "blob", "boolean", "counter",
	"date", "decimal", "double", "duration", "float",
	"frozen", "inet", "int", "list", "map",
	"set", "smallint", "text", "time", "timestamp",
	"timeuuid", "tinyint", "tuple", "uuid", "varchar",
	"varint", "vector",
}

// The functions a SELECT can use, in the families Cassandra groups them in.

// AggregateFunctions work over the rows a query matched.
var AggregateFunctions = []string{
	"COUNT", "MAX", "MIN", "AVG", "SUM",
}

// ScalarMathFunctions work over one number.
var ScalarMathFunctions = []string{
	"ABS", "EXP", "LOG", "LOG10", "ROUND",
}

// CollectionFunctions work over one list, set or map.
var CollectionFunctions = []string{
	"MAP_KEYS", "MAP_VALUES",
	"COLLECTION_AVG", "COLLECTION_COUNT", "COLLECTION_MIN",
	"COLLECTION_MAX", "COLLECTION_SUM",
}

// MaskFunctions hide what a column holds, for a role without UNMASK.
var MaskFunctions = []string{
	"MASK_DEFAULT", "MASK_HASH", "MASK_INNER",
	"MASK_NULL", "MASK_OUTER", "MASK_REPLACE",
}

// TimeFunctions are the clock, and the conversions between the time types.
//
// The lower-case names are the older spellings, which still work; the
// upper-case ones are what Cassandra's own documentation uses now.
var TimeFunctions = []string{
	"CURRENT_DATE", "CURRENT_TIME", "CURRENT_TIMESTAMP", "CURRENT_TIMEUUID",
	"TO_DATE", "TO_TIMESTAMP", "TO_UNIX_TIMESTAMP",
	"MIN_TIMEUUID", "MAX_TIMEUUID",
	"now", "currentTimeUUID", "currentTimestamp", "currentDate",
	"minTimeuuid", "maxTimeuuid", "toDate", "toTimestamp",
	"toUnixTimestamp", "dateOf", "unixTimestampOf",
}

// SystemFunctions are the rest: tokens, blobs, and what a cell carries besides
// its value.
var SystemFunctions = []string{
	"token", "uuid", "blobAsText", "textAsBlob",
	"blobAsBigint", "bigintAsBlob",
	"TTL", "WRITETIME", "MAX_WRITETIME", "MIN_WRITETIME",
	"CAST",
}

// DDL object types
var DDLObjectTypes = []string{
	"TABLE", "KEYSPACE", "INDEX", "TYPE",
	"FUNCTION", "AGGREGATE", "MATERIALIZED",
	"ROLE", "USER", "TRIGGER",
}

// TableOptions are what CREATE TABLE and ALTER TABLE take after WITH.
//
// Cassandra's TableParams.Option, as of 5.0. read_repair_chance and
// dclocal_read_repair_chance were offered here and were removed from Cassandra
// in 4.0. The options newer than 5.0 - fast_path, transactional_mode,
// transactional_migration_from, pending_drop, auto_repair - are left out until
// completion knows which server it is talking to: offering one to a cluster
// that refuses it is the defect this list already had.
var TableOptions = []string{
	"additional_write_policy",
	"allow_auto_snapshot",
	"bloom_filter_fp_chance",
	"caching",
	"cdc",
	"comment",
	"compaction",
	"compression",
	"crc_check_chance",
	"default_time_to_live",
	"extensions",
	"gc_grace_seconds",
	"incremental_backups",
	"max_index_interval",
	"memtable",
	"memtable_flush_period_in_ms",
	"min_index_interval",
	"read_repair",
	"speculative_retry",
}

// The three table options that take a map, and what goes in each. Without
// these, completion stops at the opening brace of the option most often edited.

// CompactionOptions are the keys any compaction strategy takes.
var CompactionOptions = []string{
	"class",
	"enabled",
	"max_threshold",
	"min_threshold",
	"only_purge_repaired_tombstones",
	"provide_overlapping_tombstones",
	"tombstone_compaction_interval",
	"tombstone_threshold",
	"unchecked_tombstone_compaction",
}

// CompactionStrategies are the classes that option names.
var CompactionStrategies = []string{
	"SizeTieredCompactionStrategy",
	"LeveledCompactionStrategy",
	"TimeWindowCompactionStrategy",
	"UnifiedCompactionStrategy",
}

// StrategyOptions are the keys each strategy takes of its own, on top of the
// ones every strategy takes.
var StrategyOptions = map[string][]string{
	"SizeTieredCompactionStrategy": {
		"min_sstable_size", "bucket_high", "bucket_low",
	},
	"LeveledCompactionStrategy": {
		"sstable_size_in_mb", "fanout_size", "single_sstable_uplevel",
	},
	"TimeWindowCompactionStrategy": {
		"compaction_window_unit", "compaction_window_size", "timestamp_resolution",
		"expired_sstable_check_frequency_seconds", "unsafe_aggressive_sstable_expiration",
	},
	"UnifiedCompactionStrategy": {
		"scaling_parameters", "target_sstable_size", "sstable_growth",
		"max_sstables_to_compact", "base_shard_count", "min_sstable_size",
		"expired_sstable_check_frequency_seconds",
	},
}

// The values Cassandra fixes the list of. Where it does not fix one - a number
// of seconds, a fraction, a name - completion says what to type instead.

// BooleanValues are what an option that is on or off takes.
var BooleanValues = []string{"true", "false"}

// SpeculativeRetryValues are the named policies speculative_retry and
// additional_write_policy take. The other forms are a percentile (99p) or a
// delay (50ms), and MIN or MAX of the two.
var SpeculativeRetryValues = []string{"ALWAYS", "NEVER", "NONE"}

// ReadRepairValues are the read repair strategies.
var ReadRepairValues = []string{"BLOCKING", "NONE"}

// CachingValues are what the keys of the caching map take, besides a number of
// rows for rows_per_partition.
var CachingValues = []string{"ALL", "NONE"}

// TombstoneValues are what provide_overlapping_tombstones takes.
var TombstoneValues = []string{"NONE", "ROW", "CELL"}

// CompactionWindowUnits are the units a time window is measured in.
var CompactionWindowUnits = []string{"MINUTES", "HOURS", "DAYS"}

// TimestampResolutions are the resolutions a timestamp is read at.
var TimestampResolutions = []string{"SECONDS", "MILLISECONDS", "MICROSECONDS", "NANOSECONDS"}

// CompressionOptions are the keys the compression map takes.
//
// max_compressed_length is offered by cqlsh and is in no version of the server:
// CompressionParams has class, chunk_length_in_kb, enabled and
// min_compress_ratio, in 4.1, 5.0 and trunk alike.
var CompressionOptions = []string{
	"class",
	"chunk_length_in_kb",
	"enabled",
	"min_compress_ratio",
}

// Compressors are the classes the compression class takes.
var Compressors = []string{
	"LZ4Compressor",
	"SnappyCompressor",
	"DeflateCompressor",
	"ZstdCompressor",
	"NoopCompressor",
}

// CachingOptions are the keys the caching map takes.
var CachingOptions = []string{
	"keys",
	"rows_per_partition",
}

// What an index is built on and built with.

// IndexTargets are the parts of a collection an index can be built on, from
// IndexTarget.Type. A column on its own is the fifth kind and is written as it
// stands.
var IndexTargets = []string{"keys(", "values(", "entries(", "full("}

// IndexClasses are what USING takes: the implementation the index uses. Each
// is a name the server knows - `sai` and the class's own name are the same
// thing - and a class of anyone's own can be named in full.
var IndexClasses = []string{"sai", "StorageAttachedIndex", "SASIIndex", "legacy_local_table"}

// IndexOptions are the options a storage attached index takes, from
// StorageAttachedIndex.VALID_OPTIONS less the two the server fills in itself.
var IndexOptions = []string{
	"ascii",
	"case_sensitive",
	"construction_beam_width",
	"maximum_node_connections",
	"normalize",
	"optimize_for",
	"similarity_function",
}

// SimilarityFunctions are how a vector index measures distance.
var SimilarityFunctions = []string{"COSINE", "DOT_PRODUCT", "EUCLIDEAN"}

// OptimizeForValues are what a vector index is tuned for.
var OptimizeForValues = []string{"LATENCY", "RECALL"}

// UDFLanguages are what a user defined function can be written in. Java is the
// only one left: the scripted languages were deprecated in 4.1 and are gone.
var UDFLanguages = []string{"java"}

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
	"AUTOSAVE",
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

// KeyspaceOptions are what CREATE KEYSPACE and ALTER KEYSPACE take after WITH.
//
// KeyspaceParams.Option, less fast_path, which is newer than 5.0. CLUSTERING,
// COMPACT and COMPRESSION were offered here and belong to a table.
var KeyspaceOptions = []string{
	"replication", "durable_writes",
}

// ReplicationTemplates are the replication map written out, which is the one
// thing every CREATE KEYSPACE needs and nobody writes from memory.
var ReplicationTemplates = []string{
	"{'class': 'SimpleStrategy', 'replication_factor': 1}",
	"{'class': 'NetworkTopologyStrategy', 'datacenter1': 3}",
}

// RoleOptions are what CREATE ROLE and ALTER ROLE take after WITH, from the
// grammar's roleOption. The ones with an equals sign take a value; the rest
// stand on their own.
var RoleOptions = []string{
	"PASSWORD = ",
	"HASHED PASSWORD = ",
	"GENERATED PASSWORD",
	"LOGIN = ",
	"SUPERUSER = ",
	"OPTIONS = ",
	"ACCESS TO ALL DATACENTERS",
	"ACCESS TO DATACENTERS {",
	"ACCESS FROM ALL CIDRS",
	"ACCESS FROM CIDRS {",
}

// UserPasswords are what CREATE USER and ALTER USER take after WITH. A user is
// a role written the old way, and its password has no equals sign.
var UserPasswords = []string{
	"PASSWORD ",
	"HASHED PASSWORD ",
	"GENERATED PASSWORD",
}

// CreateObjectTypes are what CREATE makes.
//
// CUSTOM is not one of them: it is the word before INDEX, and it is offered
// here because this is where it is typed. The list of what CREATE makes was
// written out four times over, and this is the one the others now read.
var CreateObjectTypes = []string{
	"AGGREGATE", "CUSTOM", "FUNCTION", "INDEX", "KEYSPACE", "MATERIALIZED",
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
	"AUTOSAVE",
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
