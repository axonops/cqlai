package db

import (
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// QueryResult wraps query results with metadata
type QueryResult struct {
	Data            [][]string               // Formatted strings for display
	RawData         []map[string]interface{} // Raw values for JSON export (preserves types)
	Duration        time.Duration
	RowCount        int
	ColumnTypes     []string         // Data types of each column
	ColumnTypeInfos []gocql.TypeInfo // TypeInfo objects for each column (for UDT support)
	Headers         []string         // Column names without PK/C indicators
}

// StreamingQueryResult wraps query results for progressive loading
type StreamingQueryResult struct {
	Headers         []string         // Column headers (with PK/C indicators)
	ColumnNames     []string         // Original column names (for data lookup)
	ColumnTypes     []string         // Data types of each column
	ColumnTypeInfos []gocql.TypeInfo // TypeInfo objects for each column (for UDT support)
	Iterator        *gocql.Iter      // Iterator for fetching more rows
	StartTime       time.Time        // Query start time for duration calculation
	Keyspace        string           // Keyspace extracted from query or session
}

// KeyColumnInfo holds information about key columns
type KeyColumnInfo struct {
	Kind     string // "partition_key" or "clustering"
	Position int
}

// TypeInfoToString converts a gocql.TypeInfo to its string representation
func TypeInfoToString(typeInfo gocql.TypeInfo) string {
	if typeInfo == nil {
		return "unknown"
	}

	t := typeInfo.Type()

	// Handle UDT types specially to include the UDT name
	if t == gocql.TypeUDT {
		// Try to cast to a more specific type that might have UDT info
		// For now, return a generic "udt" - the actual UDT name needs to be
		// fetched from system tables based on the column metadata
		return "udt"
	}

	// For custom types, try to get more specific information
	if t == gocql.TypeCustom {
		// Check if it's a vector type by looking at the string representation
		// The TypeInfo interface doesn't expose the custom type name directly,
		// but we can infer it from context or default to "vector" for now
		return "vector"
	}

	// Handle collection types with their element types
	switch t {
	case gocql.TypeList:
		return "list"
	case gocql.TypeSet:
		return "set"
	case gocql.TypeMap:
		return "map"
	case gocql.TypeTuple:
		return "tuple"
	}

	return TypeToString(t)
}

// TypeToString converts a gocql.Type to its string representation
func TypeToString(t gocql.Type) string {
	switch t {
	case gocql.TypeCustom:
		return "custom"
	case gocql.TypeAscii:
		return "ascii"
	case gocql.TypeBigInt:
		return "bigint"
	case gocql.TypeBlob:
		return "blob"
	case gocql.TypeBoolean:
		return "boolean"
	case gocql.TypeCounter:
		return "counter"
	case gocql.TypeDecimal:
		return "decimal"
	case gocql.TypeDouble:
		return "double"
	case gocql.TypeFloat:
		return "float"
	case gocql.TypeInt:
		return "int"
	case gocql.TypeText:
		return "text"
	case gocql.TypeTimestamp:
		return "timestamp"
	case gocql.TypeUUID:
		return "uuid"
	case gocql.TypeVarchar:
		return "varchar"
	case gocql.TypeVarint:
		return "varint"
	case gocql.TypeTimeUUID:
		return "timeuuid"
	case gocql.TypeInet:
		return "inet"
	case gocql.TypeDate:
		return "date"
	case gocql.TypeTime:
		return "time"
	case gocql.TypeSmallInt:
		return "smallint"
	case gocql.TypeTinyInt:
		return "tinyint"
	case gocql.TypeDuration:
		return "duration"
	case gocql.TypeList:
		return "list"
	case gocql.TypeMap:
		return "map"
	case gocql.TypeSet:
		return "set"
	case gocql.TypeUDT:
		return "udt"
	case gocql.TypeTuple:
		return "tuple"
	default:
		return "unknown"
	}
}