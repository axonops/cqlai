package db

import (
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// QueryResult wraps query results with metadata
type QueryResult struct {
	Data        [][]string
	Duration    time.Duration
	RowCount    int
	ColumnTypes []string // Data types of each column
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
	
	// For custom types, try to get more specific information
	t := typeInfo.Type()
	if t == gocql.TypeCustom {
		// Check if it's a vector type by looking at the string representation
		// The TypeInfo interface doesn't expose the custom type name directly,
		// but we can infer it from context or default to "vector" for now
		return "vector"
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