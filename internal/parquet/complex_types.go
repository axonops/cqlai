package parquet

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/logger"
)

// AppendListValue appends a list value to a ListBuilder
func (tm *TypeMapper) AppendListValue(builder *array.ListBuilder, value interface{}, elementType arrow.DataType) error {
	if value == nil {
		builder.AppendNull()
		return nil
	}

	// Start a new list
	builder.Append(true)

	// Get the value builder for list elements
	valueBuilder := builder.ValueBuilder()

	// Handle different list representations
	switch v := value.(type) {
	case []interface{}:
		for _, elem := range v {
			if err := tm.AppendValueToBuilder(valueBuilder, elem, elementType); err != nil {
				return fmt.Errorf("failed to append list element: %w", err)
			}
		}
	case []string:
		for _, elem := range v {
			if err := tm.AppendValueToBuilder(valueBuilder, elem, elementType); err != nil {
				return fmt.Errorf("failed to append list element: %w", err)
			}
		}
	case []int:
		for _, elem := range v {
			if err := tm.AppendValueToBuilder(valueBuilder, elem, elementType); err != nil {
				return fmt.Errorf("failed to append list element: %w", err)
			}
		}
	case []int32:
		for _, elem := range v {
			if err := tm.AppendValueToBuilder(valueBuilder, elem, elementType); err != nil {
				return fmt.Errorf("failed to append list element: %w", err)
			}
		}
	case []int64:
		for _, elem := range v {
			if err := tm.AppendValueToBuilder(valueBuilder, elem, elementType); err != nil {
				return fmt.Errorf("failed to append list element: %w", err)
			}
		}
	case []float32:
		for _, elem := range v {
			if err := tm.AppendValueToBuilder(valueBuilder, elem, elementType); err != nil {
				return fmt.Errorf("failed to append list element: %w", err)
			}
		}
	case []float64:
		for _, elem := range v {
			if err := tm.AppendValueToBuilder(valueBuilder, elem, elementType); err != nil {
				return fmt.Errorf("failed to append list element: %w", err)
			}
		}
	case []bool:
		for _, elem := range v {
			if err := tm.AppendValueToBuilder(valueBuilder, elem, elementType); err != nil {
				return fmt.Errorf("failed to append list element: %w", err)
			}
		}
	default:
		// Try to convert to []interface{}
		if converted, ok := value.([]interface{}); ok {
			for _, elem := range converted {
				if err := tm.AppendValueToBuilder(valueBuilder, elem, elementType); err != nil {
					return fmt.Errorf("failed to append list element: %w", err)
				}
			}
		} else {
			return fmt.Errorf("unsupported list type: %T", value)
		}
	}

	return nil
}

// AppendMapValue appends a map value to a MapBuilder
func (tm *TypeMapper) AppendMapValue(builder *array.MapBuilder, value interface{}, keyType, valueType arrow.DataType) error {
	if value == nil {
		builder.AppendNull()
		return nil
	}

	// Start a new map
	builder.Append(true)

	// Get the key and value builders
	keyBuilder := builder.KeyBuilder()
	valueBuilder := builder.ItemBuilder()

	// Handle different map representations
	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if err := tm.AppendValueToBuilder(keyBuilder, key, keyType); err != nil {
				return fmt.Errorf("failed to append map key: %w", err)
			}
			if err := tm.AppendValueToBuilder(valueBuilder, val, valueType); err != nil {
				return fmt.Errorf("failed to append map value: %w", err)
			}
		}
	case map[interface{}]interface{}:
		for key, val := range v {
			if err := tm.AppendValueToBuilder(keyBuilder, key, keyType); err != nil {
				return fmt.Errorf("failed to append map key: %w", err)
			}
			if err := tm.AppendValueToBuilder(valueBuilder, val, valueType); err != nil {
				return fmt.Errorf("failed to append map value: %w", err)
			}
		}
	case map[string]string:
		for key, val := range v {
			if err := tm.AppendValueToBuilder(keyBuilder, key, keyType); err != nil {
				return fmt.Errorf("failed to append map key: %w", err)
			}
			if err := tm.AppendValueToBuilder(valueBuilder, val, valueType); err != nil {
				return fmt.Errorf("failed to append map value: %w", err)
			}
		}
	case map[string]int:
		for key, val := range v {
			if err := tm.AppendValueToBuilder(keyBuilder, key, keyType); err != nil {
				return fmt.Errorf("failed to append map key: %w", err)
			}
			if err := tm.AppendValueToBuilder(valueBuilder, val, valueType); err != nil {
				return fmt.Errorf("failed to append map value: %w", err)
			}
		}
	default:
		return fmt.Errorf("unsupported map type: %T", value)
	}

	return nil
}

// AppendStructValue appends a struct value (for UDTs and Tuples)
func (tm *TypeMapper) AppendStructValue(builder *array.StructBuilder, value interface{}, fields []arrow.Field) error {
	if value == nil {
		builder.AppendNull()
		return nil
	}

	// Append to indicate a new struct row
	builder.Append(true)

	// Handle UDT as map
	logger.DebugfToFile("AppendStructValue", "Got value of type: %T, value: %+v", value, value)
	switch v := value.(type) {
	case string:
		// If we get a string for a UDT (formatted display), try to parse it
		logger.DebugfToFile("AppendStructValue", "Got string UDT value: %s", v)
		parsedMap := parseUDTDisplayString(v)
		logger.DebugfToFile("AppendStructValue", "Parsed map: %+v", parsedMap)
		if parsedMap != nil {
			// Successfully parsed, append the values
			for i, field := range fields {
				fieldBuilder := builder.FieldBuilder(i)
				fieldValue := parsedMap[field.Name]
				if err := tm.AppendValueToBuilder(fieldBuilder, fieldValue, field.Type); err != nil {
					// If error, append null
					fieldBuilder.AppendNull()
				}
			}
		} else {
			// Couldn't parse, append nulls for all fields
			for i := range fields {
				fieldBuilder := builder.FieldBuilder(i)
				fieldBuilder.AppendNull()
			}
		}
	case map[string]interface{}:
		// For each field in the struct
		for i, field := range fields {
			fieldBuilder := builder.FieldBuilder(i)
			fieldValue := v[field.Name]
			if err := tm.AppendValueToBuilder(fieldBuilder, fieldValue, field.Type); err != nil {
				return fmt.Errorf("failed to append struct field %s: %w", field.Name, err)
			}
		}
	case []interface{}:
		// For tuples - ordered values
		for i, fieldValue := range v {
			if i >= len(fields) {
				break
			}
			fieldBuilder := builder.FieldBuilder(i)
			if err := tm.AppendValueToBuilder(fieldBuilder, fieldValue, fields[i].Type); err != nil {
				return fmt.Errorf("failed to append tuple field %d: %w", i, err)
			}
		}
		// Append nulls for any remaining fields
		for i := len(v); i < len(fields); i++ {
			fieldBuilder := builder.FieldBuilder(i)
			fieldBuilder.AppendNull()
		}
	default:
		return fmt.Errorf("unsupported struct value type: %T", value)
	}

	return nil
}

// parseUDTFromTypeInfo creates an Arrow struct type from UDTTypeInfo
func (tm *TypeMapper) parseUDTFromTypeInfo(udt gocql.UDTTypeInfo) (*arrow.StructType, error) {
	fields := make([]arrow.Field, 0, len(udt.Elements))

	for _, elem := range udt.Elements {
		// Convert the element type to Arrow type
		var arrowType arrow.DataType

		// Try to get Arrow type from the element's TypeInfo
		// We'll use a simple type mapping based on the gocql Type
		switch elem.Type.Type() {
		case gocql.TypeText, gocql.TypeAscii, gocql.TypeVarchar:
			arrowType = arrow.BinaryTypes.String
		case gocql.TypeInt:
			arrowType = arrow.PrimitiveTypes.Int32
		case gocql.TypeBigInt:
			arrowType = arrow.PrimitiveTypes.Int64
		case gocql.TypeFloat:
			arrowType = arrow.PrimitiveTypes.Float32
		case gocql.TypeDouble:
			arrowType = arrow.PrimitiveTypes.Float64
		case gocql.TypeBoolean:
			arrowType = arrow.FixedWidthTypes.Boolean
		case gocql.TypeTuple:
			// Tuples as strings for now
			arrowType = arrow.BinaryTypes.String
		default:
			// Default to string for unknown types
			arrowType = arrow.BinaryTypes.String
		}

		fields = append(fields, arrow.Field{
			Name:     elem.Name,
			Type:     arrowType,
			Nullable: true,
		})
	}

	if len(fields) == 0 {
		// If no fields, return a generic struct
		return arrow.StructOf(
			arrow.Field{Name: "value", Type: arrow.BinaryTypes.String, Nullable: true},
		), nil
	}

	return arrow.StructOf(fields...), nil
}

// ParseUDTType parses a Cassandra UDT type string into an Arrow struct type
func (tm *TypeMapper) ParseUDTType(udtSpec string, udtInfo gocql.TypeInfo) (*arrow.StructType, error) {
	// If we have UDTTypeInfo, use it to get the actual fields
	if udtInfo != nil {
		if udt, ok := udtInfo.(gocql.UDTTypeInfo); ok {
			return tm.parseUDTFromTypeInfo(udt)
		}
	}

	// UDT format: "udt<field1:type1,field2:type2,...>"
	if !strings.HasPrefix(udtSpec, "udt<") || !strings.HasSuffix(udtSpec, ">") {
		// Try to handle it as a named UDT (would need schema lookup in real implementation)
		// For now, return a generic struct
		return arrow.StructOf(
			arrow.Field{Name: "value", Type: arrow.BinaryTypes.String, Nullable: true},
		), nil
	}

	inner := strings.TrimPrefix(udtSpec, "udt<")
	inner = strings.TrimSuffix(inner, ">")

	// Parse field definitions
	fieldDefs := splitTypeParams(inner)
	fields := make([]arrow.Field, 0, len(fieldDefs))

	for _, fieldDef := range fieldDefs {
		parts := strings.SplitN(fieldDef, ":", 2)
		if len(parts) != 2 {
			continue
		}

		fieldName := strings.TrimSpace(parts[0])
		fieldType := strings.TrimSpace(parts[1])

		arrowType, err := tm.CassandraToArrowType(fieldType)
		if err != nil {
			// Default to string if we can't parse the type
			arrowType = arrow.BinaryTypes.String
		}

		fields = append(fields, arrow.Field{
			Name:     fieldName,
			Type:     arrowType,
			Nullable: true,
		})
	}

	if len(fields) == 0 {
		// Fallback to a single string field
		fields = []arrow.Field{
			{Name: "value", Type: arrow.BinaryTypes.String, Nullable: true},
		}
	}

	return arrow.StructOf(fields...), nil
}

// ParseTupleType parses a Cassandra tuple type string into an Arrow struct type
func (tm *TypeMapper) ParseTupleType(tupleSpec string) (*arrow.StructType, error) {
	// Tuple format: "tuple<type1,type2,...>"
	if !strings.HasPrefix(tupleSpec, "tuple<") || !strings.HasSuffix(tupleSpec, ">") {
		return nil, fmt.Errorf("invalid tuple type: %s", tupleSpec)
	}

	inner := strings.TrimPrefix(tupleSpec, "tuple<")
	inner = strings.TrimSuffix(inner, ">")

	// Parse element types
	elementTypes := splitTypeParams(inner)
	fields := make([]arrow.Field, len(elementTypes))

	for i, elemType := range elementTypes {
		arrowType, err := tm.CassandraToArrowType(strings.TrimSpace(elemType))
		if err != nil {
			// Default to string if we can't parse the type
			arrowType = arrow.BinaryTypes.String
		}

		fields[i] = arrow.Field{
			Name:     fmt.Sprintf("field%d", i),
			Type:     arrowType,
			Nullable: true,
		}
	}

	return arrow.StructOf(fields...), nil
}

// ExtractListValue extracts a list value from an Arrow array
func ExtractListValue(col *array.List, idx int) interface{} {
	if col.IsNull(idx) {
		return nil
	}

	start, end := col.ValueOffsets(idx)
	valueData := col.ListValues()

	result := make([]interface{}, 0, end-start)
	for i := start; i < end; i++ {
		if valueData.IsNull(int(i)) {
			result = append(result, nil)
		} else {
			// Extract based on value type
			switch v := valueData.(type) {
			case *array.String:
				result = append(result, v.Value(int(i)))
			case *array.Int32:
				result = append(result, v.Value(int(i)))
			case *array.Int64:
				result = append(result, v.Value(int(i)))
			case *array.Float64:
				result = append(result, v.Value(int(i)))
			case *array.Boolean:
				result = append(result, v.Value(int(i)))
			default:
				// For complex nested types, recursively extract
				result = append(result, extractValue(valueData, int(i)))
			}
		}
	}

	return result
}

// ExtractMapValue extracts a map value from an Arrow array
func ExtractMapValue(col *array.Map, idx int) interface{} {
	if col.IsNull(idx) {
		return nil
	}

	start, end := col.ValueOffsets(idx)
	keys := col.Keys()
	items := col.Items()

	result := make(map[interface{}]interface{})
	for i := start; i < end; i++ {
		var key, value interface{}

		// Extract key
		if !keys.IsNull(int(i)) {
			key = extractValue(keys, int(i))
		}

		// Extract value
		if !items.IsNull(int(i)) {
			value = extractValue(items, int(i))
		}

		if key != nil {
			result[key] = value
		}
	}

	return result
}

// ExtractStructValue extracts a struct value from an Arrow array
func ExtractStructValue(col *array.Struct, idx int) interface{} {
	if col.IsNull(idx) {
		return nil
	}

	numFields := col.NumField()
	result := make(map[string]interface{})

	structType := col.DataType().(*arrow.StructType)
	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		fieldData := col.Field(i)

		if !fieldData.IsNull(idx) {
			result[field.Name] = extractValue(fieldData, idx)
		} else {
			result[field.Name] = nil
		}
	}

	return result
}

// CreateBuilderForType creates the appropriate Arrow array builder for a given type
func CreateBuilderForType(allocator memory.Allocator, dataType arrow.DataType) array.Builder {
	switch dt := dataType.(type) {
	case *arrow.ListType:
		return array.NewListBuilder(allocator, dt.Elem())
	case *arrow.MapType:
		return array.NewMapBuilder(allocator, dt.KeyType(), dt.ValueType(), false)
	case *arrow.StructType:
		return array.NewStructBuilder(allocator, dt)
	default:
		// For basic types, use the default builder creation
		return array.NewBuilder(allocator, dataType)
	}
}

// parseUDTDisplayString parses a Cassandra UDT display string like "{street: '123 Main St', city: 'NYC'}"
// into a map[string]interface{}
func parseUDTDisplayString(s string) map[string]interface{} {
	// Trim curly braces
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return nil
	}
	s = s[1 : len(s)-1]

	// If empty, return empty map
	if strings.TrimSpace(s) == "" {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})

	// Split by commas, but need to handle commas inside quoted strings
	pairs := splitRespectingQuotes(s, ',')

	for _, pair := range pairs {
		// Split by colon
		parts := splitRespectingQuotes(pair, ':')
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes from value if present
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			value = value[1 : len(value)-1]
			// Unescape single quotes
			value = strings.ReplaceAll(value, "''", "'")
			result[key] = value
		} else if value == "null" {
			result[key] = nil
		} else {
			// Try to parse as number
			if i, err := strconv.ParseInt(value, 10, 64); err == nil {
				result[key] = i
			} else if f, err := strconv.ParseFloat(value, 64); err == nil {
				result[key] = f
			} else if b, err := strconv.ParseBool(value); err == nil {
				result[key] = b
			} else {
				result[key] = value
			}
		}
	}

	return result
}

// splitRespectingQuotes splits a string by the given separator, but respects quoted strings
func splitRespectingQuotes(s string, sep rune) []string {
	var result []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for i, r := range s {
		if !inQuotes {
			if r == '\'' || r == '"' {
				inQuotes = true
				quoteChar = r
				current.WriteRune(r)
			} else if r == sep {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		} else {
			current.WriteRune(r)
			if r == quoteChar {
				// Check if it's an escaped quote
				if i+1 < len(s) && rune(s[i+1]) == quoteChar {
					// Skip the next quote as it's an escape
					continue
				}
				inQuotes = false
				quoteChar = 0
			}
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}