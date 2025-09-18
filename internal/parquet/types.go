package parquet

import (
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/decimal128"
	"github.com/apache/arrow-go/v18/arrow/memory"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// TypeMapper handles conversion between Cassandra and Arrow types
type TypeMapper struct {
	allocator memory.Allocator
}

// NewTypeMapper creates a new type mapper
func NewTypeMapper() *TypeMapper {
	return &TypeMapper{
		allocator: memory.DefaultAllocator,
	}
}

// CassandraToArrowType maps Cassandra type strings to Arrow DataTypes
func (tm *TypeMapper) CassandraToArrowType(cassandraType string) (arrow.DataType, error) {
	// Normalize the type string
	cassandraType = strings.ToLower(strings.TrimSpace(cassandraType))

	// Handle parameterized types (e.g., "decimal(10,2)")
	if idx := strings.Index(cassandraType, "("); idx > 0 {
		cassandraType = cassandraType[:idx]
	}

	switch cassandraType {
	// Numeric types
	case "tinyint":
		return arrow.PrimitiveTypes.Int8, nil
	case "smallint":
		return arrow.PrimitiveTypes.Int16, nil
	case "int":
		return arrow.PrimitiveTypes.Int32, nil
	case "bigint", "counter":
		return arrow.PrimitiveTypes.Int64, nil
	case "float":
		return arrow.PrimitiveTypes.Float32, nil
	case "double":
		return arrow.PrimitiveTypes.Float64, nil
	case "decimal":
		// Using Decimal128 for decimal types
		return &arrow.Decimal128Type{Precision: 38, Scale: 10}, nil
	case "varint":
		// Variable-precision integers stored as string for compatibility
		return arrow.BinaryTypes.String, nil

	// String types
	case "ascii", "text", "varchar":
		return arrow.BinaryTypes.String, nil

	// Binary type
	case "blob":
		return arrow.BinaryTypes.Binary, nil

	// Boolean type
	case "boolean":
		return arrow.FixedWidthTypes.Boolean, nil

	// UUID types
	case "uuid", "timeuuid":
		// UUIDs are stored as strings for compatibility
		return arrow.BinaryTypes.String, nil

	// Date/Time types
	case "date":
		return arrow.FixedWidthTypes.Date32, nil
	case "time":
		return arrow.FixedWidthTypes.Time64ns, nil
	case "timestamp":
		return arrow.FixedWidthTypes.Timestamp_ms, nil
	case "duration":
		// Duration stored as int64 nanoseconds
		return arrow.PrimitiveTypes.Int64, nil

	// Network type
	case "inet":
		// IP addresses stored as strings
		return arrow.BinaryTypes.String, nil

	default:
		// Check for collection types
		if strings.HasPrefix(cassandraType, "list<") {
			return tm.parseListType(cassandraType)
		} else if strings.HasPrefix(cassandraType, "set<") {
			return tm.parseSetType(cassandraType)
		} else if strings.HasPrefix(cassandraType, "map<") {
			return tm.parseMapType(cassandraType)
		} else if strings.HasPrefix(cassandraType, "frozen<") {
			// Handle frozen types by parsing the inner type
			innerType := strings.TrimPrefix(cassandraType, "frozen<")
			innerType = strings.TrimSuffix(innerType, ">")
			return tm.CassandraToArrowType(innerType)
		}

		// Default to string for unknown types
		return arrow.BinaryTypes.String, nil
	}
}

// parseListType parses Cassandra list type to Arrow list type
func (tm *TypeMapper) parseListType(cassandraType string) (arrow.DataType, error) {
	// Extract element type from "list<type>"
	elementType := extractTypeParam(cassandraType, "list<", ">")
	if elementType == "" {
		return nil, fmt.Errorf("invalid list type: %s", cassandraType)
	}

	elemArrowType, err := tm.CassandraToArrowType(elementType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse list element type: %w", err)
	}

	return arrow.ListOf(elemArrowType), nil
}

// parseSetType parses Cassandra set type to Arrow list type (sets are lists in Arrow)
func (tm *TypeMapper) parseSetType(cassandraType string) (arrow.DataType, error) {
	// Extract element type from "set<type>"
	elementType := extractTypeParam(cassandraType, "set<", ">")
	if elementType == "" {
		return nil, fmt.Errorf("invalid set type: %s", cassandraType)
	}

	elemArrowType, err := tm.CassandraToArrowType(elementType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse set element type: %w", err)
	}

	// Sets are represented as lists in Arrow
	return arrow.ListOf(elemArrowType), nil
}

// parseMapType parses Cassandra map type to Arrow map type
func (tm *TypeMapper) parseMapType(cassandraType string) (arrow.DataType, error) {
	// Extract key and value types from "map<key,value>"
	inner := extractTypeParam(cassandraType, "map<", ">")
	if inner == "" {
		return nil, fmt.Errorf("invalid map type: %s", cassandraType)
	}

	// Split by comma (careful with nested types)
	parts := splitTypeParams(inner)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid map type parameters: %s", cassandraType)
	}

	keyType, err := tm.CassandraToArrowType(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse map key type: %w", err)
	}

	valueType, err := tm.CassandraToArrowType(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse map value type: %w", err)
	}

	return arrow.MapOf(keyType, valueType), nil
}

// ConvertValue converts a Cassandra value to Arrow-compatible format
func (tm *TypeMapper) ConvertValue(value interface{}, arrowType arrow.DataType) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch arrowType.ID() {
	case arrow.INT8:
		return tm.toInt8(value)
	case arrow.INT16:
		return tm.toInt16(value)
	case arrow.INT32:
		return tm.toInt32(value)
	case arrow.INT64:
		return tm.toInt64(value)
	case arrow.FLOAT32:
		return tm.toFloat32(value)
	case arrow.FLOAT64:
		return tm.toFloat64(value)
	case arrow.BOOL:
		return tm.toBool(value)
	case arrow.STRING:
		return tm.toString(value)
	case arrow.BINARY:
		return tm.toBinary(value)
	case arrow.DATE32:
		return tm.toDate32(value)
	case arrow.TIME64:
		return tm.toTime64(value)
	case arrow.TIMESTAMP:
		return tm.toTimestamp(value)
	case arrow.DECIMAL128:
		return tm.toDecimal128(value)
	case arrow.LIST:
		return tm.toList(value, arrowType.(*arrow.ListType))
	case arrow.MAP:
		return tm.toMap(value, arrowType.(*arrow.MapType))
	default:
		// Default conversion to string
		return fmt.Sprintf("%v", value), nil
	}
}

// Type conversion helper functions

func (tm *TypeMapper) toInt8(value interface{}) (int8, error) {
	switch v := value.(type) {
	case int8:
		return v, nil
	case int16:
		return int8(v), nil
	case int32:
		return int8(v), nil
	case int64:
		return int8(v), nil
	case int:
		return int8(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int8", value)
	}
}

func (tm *TypeMapper) toInt16(value interface{}) (int16, error) {
	switch v := value.(type) {
	case int8:
		return int16(v), nil
	case int16:
		return v, nil
	case int32:
		return int16(v), nil
	case int64:
		return int16(v), nil
	case int:
		return int16(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int16", value)
	}
}

func (tm *TypeMapper) toInt32(value interface{}) (int32, error) {
	switch v := value.(type) {
	case int8:
		return int32(v), nil
	case int16:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case int:
		return int32(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int32", value)
	}
}

func (tm *TypeMapper) toInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case time.Duration:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

func (tm *TypeMapper) toFloat32(value interface{}) (float32, error) {
	switch v := value.(type) {
	case float32:
		return v, nil
	case float64:
		return float32(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float32", value)
	}
}

func (tm *TypeMapper) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func (tm *TypeMapper) toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

func (tm *TypeMapper) toString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case gocql.UUID:
		return v.String(), nil
	case net.IP:
		return v.String(), nil
	case *big.Int:
		return v.String(), nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}

func (tm *TypeMapper) toBinary(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to binary", value)
	}
}

func (tm *TypeMapper) toDate32(value interface{}) (arrow.Date32, error) {
	switch v := value.(type) {
	case time.Time:
		// Date32 is days since Unix epoch
		days := v.Unix() / 86400
		return arrow.Date32(days), nil
	case string:
		// Try to parse date string
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return 0, fmt.Errorf("cannot parse date: %w", err)
		}
		days := t.Unix() / 86400
		return arrow.Date32(days), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to Date32", value)
	}
}

func (tm *TypeMapper) toTime64(value interface{}) (arrow.Time64, error) {
	switch v := value.(type) {
	case time.Duration:
		// Time64 in nanoseconds
		return arrow.Time64(v.Nanoseconds()), nil
	case int64:
		return arrow.Time64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to Time64", value)
	}
}

func (tm *TypeMapper) toTimestamp(value interface{}) (arrow.Timestamp, error) {
	switch v := value.(type) {
	case time.Time:
		// Timestamp in milliseconds
		return arrow.Timestamp(v.UnixMilli()), nil
	case int64:
		return arrow.Timestamp(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to Timestamp", value)
	}
}

func (tm *TypeMapper) toDecimal128(value interface{}) (decimal128.Num, error) {
	switch v := value.(type) {
	case string:
		// Parse decimal string
		// For now, return a placeholder - full implementation would parse the string
		return decimal128.New(0, 0), nil
	case float64:
		// Convert float to decimal (with potential precision loss)
		// This is a simplified conversion
		return decimal128.New(0, uint64(v*1000000000)), nil
	default:
		return decimal128.Num{}, fmt.Errorf("cannot convert %T to Decimal128", value)
	}
}

func (tm *TypeMapper) toList(value interface{}, listType *arrow.ListType) ([]interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case []string:
		result := make([]interface{}, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result, nil
	case []int:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to list", value)
	}
}

func (tm *TypeMapper) toMap(value interface{}, mapType *arrow.MapType) (map[interface{}]interface{}, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		result := make(map[interface{}]interface{})
		for k, val := range v {
			result[k] = val
		}
		return result, nil
	case map[interface{}]interface{}:
		return v, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to map", value)
	}
}

// Helper functions

func extractTypeParam(typeStr, prefix, suffix string) string {
	if !strings.HasPrefix(typeStr, prefix) || !strings.HasSuffix(typeStr, suffix) {
		return ""
	}
	return typeStr[len(prefix) : len(typeStr)-len(suffix)]
}

func splitTypeParams(params string) []string {
	if params == "" {
		return []string{""}
	}

	var result []string
	var current strings.Builder
	depth := 0

	for _, ch := range params {
		switch ch {
		case '<':
			depth++
			current.WriteRune(ch)
		case '>':
			depth--
			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 || len(result) == 0 {
		result = append(result, current.String())
	}

	return result
}

// CreateArrowSchema creates an Arrow schema from Cassandra column information
func (tm *TypeMapper) CreateArrowSchema(columnNames []string, columnTypes []string) (*arrow.Schema, error) {
	if len(columnNames) != len(columnTypes) {
		return nil, fmt.Errorf("column names and types length mismatch")
	}

	fields := make([]arrow.Field, len(columnNames))
	for i, name := range columnNames {
		arrowType, err := tm.CassandraToArrowType(columnTypes[i])
		if err != nil {
			// Default to string type if conversion fails
			arrowType = arrow.BinaryTypes.String
		}
		fields[i] = arrow.Field{
			Name:     name,
			Type:     arrowType,
			Nullable: true, // All Cassandra columns can be null
		}
	}

	return arrow.NewSchema(fields, nil), nil
}

// AppendValueToBuilder appends a value to the appropriate Arrow array builder
func (tm *TypeMapper) AppendValueToBuilder(builder array.Builder, value interface{}, arrowType arrow.DataType) error {
	if value == nil {
		builder.AppendNull()
		return nil
	}

	convertedValue, err := tm.ConvertValue(value, arrowType)
	if err != nil {
		// On conversion error, append null
		builder.AppendNull()
		return err
	}

	// Type-specific append based on builder type
	switch b := builder.(type) {
	case *array.Int8Builder:
		if v, ok := convertedValue.(int8); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.Int16Builder:
		if v, ok := convertedValue.(int16); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.Int32Builder:
		if v, ok := convertedValue.(int32); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.Int64Builder:
		if v, ok := convertedValue.(int64); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.Float32Builder:
		if v, ok := convertedValue.(float32); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.Float64Builder:
		if v, ok := convertedValue.(float64); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.BooleanBuilder:
		if v, ok := convertedValue.(bool); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.StringBuilder:
		if v, ok := convertedValue.(string); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.BinaryBuilder:
		if v, ok := convertedValue.([]byte); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.Date32Builder:
		if v, ok := convertedValue.(arrow.Date32); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.Time64Builder:
		if v, ok := convertedValue.(arrow.Time64); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.TimestampBuilder:
		if v, ok := convertedValue.(arrow.Timestamp); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	default:
		// For complex types or unknown builders, try to append as string
		if sb, ok := builder.(*array.StringBuilder); ok {
			sb.Append(fmt.Sprintf("%v", value))
		} else {
			builder.AppendNull()
		}
	}

	return nil
}
// ArrowToCassandraType converts Arrow data types back to Cassandra types
func (tm *TypeMapper) ArrowToCassandraType(arrowType arrow.DataType) string {
	switch arrowType.ID() {
	case arrow.BOOL:
		return "boolean"
	case arrow.INT8, arrow.UINT8:
		return "tinyint"
	case arrow.INT16, arrow.UINT16:
		return "smallint"
	case arrow.INT32, arrow.UINT32:
		return "int"
	case arrow.INT64, arrow.UINT64:
		return "bigint"
	case arrow.FLOAT32:
		return "float"
	case arrow.FLOAT64:
		return "double"
	case arrow.STRING, arrow.LARGE_STRING:
		return "text"
	case arrow.BINARY, arrow.LARGE_BINARY, arrow.FIXED_SIZE_BINARY:
		return "blob"
	case arrow.DATE32, arrow.DATE64:
		return "date"
	case arrow.TIMESTAMP:
		return "timestamp"
	case arrow.TIME32, arrow.TIME64:
		return "time"
	case arrow.DECIMAL128, arrow.DECIMAL256:
		return "decimal"
	default:
		return "text" // Default fallback
	}
}
