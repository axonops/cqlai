package parquet

import (
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCassandraToArrowType(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		name          string
		cassandraType string
		expectedType  arrow.DataType
		shouldError   bool
	}{
		// Numeric types
		{"tinyint", "tinyint", arrow.PrimitiveTypes.Int8, false},
		{"smallint", "smallint", arrow.PrimitiveTypes.Int16, false},
		{"int", "int", arrow.PrimitiveTypes.Int32, false},
		{"bigint", "bigint", arrow.PrimitiveTypes.Int64, false},
		{"counter", "counter", arrow.PrimitiveTypes.Int64, false},
		{"float", "float", arrow.PrimitiveTypes.Float32, false},
		{"double", "double", arrow.PrimitiveTypes.Float64, false},

		// String types
		{"text", "text", arrow.BinaryTypes.String, false},
		{"varchar", "varchar", arrow.BinaryTypes.String, false},
		{"ascii", "ascii", arrow.BinaryTypes.String, false},

		// Binary type
		{"blob", "blob", arrow.BinaryTypes.Binary, false},

		// Boolean
		{"boolean", "boolean", arrow.FixedWidthTypes.Boolean, false},

		// UUID types
		{"uuid", "uuid", arrow.BinaryTypes.String, false},
		{"timeuuid", "timeuuid", arrow.BinaryTypes.String, false},

		// Date/Time types
		{"date", "date", arrow.FixedWidthTypes.Date32, false},
		{"time", "time", arrow.FixedWidthTypes.Time64ns, false},
		{"timestamp", "timestamp", arrow.FixedWidthTypes.Timestamp_ms, false},

		// Network type
		{"inet", "inet", arrow.BinaryTypes.String, false},

		// Variable integer type
		{"varint", "varint", arrow.BinaryTypes.String, false},

		// With spaces and mixed case
		{"TEXT", "TEXT", arrow.BinaryTypes.String, false},
		{" int ", " int ", arrow.PrimitiveTypes.Int32, false},

		// Unknown type (defaults to string)
		{"unknown", "unknown", arrow.BinaryTypes.String, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tm.CassandraToArrowType(tt.cassandraType)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedType, result)
			}
		})
	}
}

func TestCollectionTypes(t *testing.T) {
	tm := NewTypeMapper()

	t.Run("list types", func(t *testing.T) {
		result, err := tm.CassandraToArrowType("list<int>")
		require.NoError(t, err)
		listType, ok := result.(*arrow.ListType)
		require.True(t, ok)
		assert.Equal(t, arrow.PrimitiveTypes.Int32, listType.Elem())
	})

	t.Run("set types", func(t *testing.T) {
		result, err := tm.CassandraToArrowType("set<text>")
		require.NoError(t, err)
		listType, ok := result.(*arrow.ListType) // Sets are lists in Arrow
		require.True(t, ok)
		assert.Equal(t, arrow.BinaryTypes.String, listType.Elem())
	})

	t.Run("map types", func(t *testing.T) {
		result, err := tm.CassandraToArrowType("map<text, int>")
		require.NoError(t, err)
		mapType, ok := result.(*arrow.MapType)
		require.True(t, ok)
		// Check that we have a map type with the correct key and value types
		// Maps in Arrow have a value type that is a struct with key and value fields
		itemType := mapType.ItemType()
		require.NotNil(t, itemType)
		// The item type should be a struct with key and value fields
		structType, ok := itemType.(*arrow.StructType)
		require.True(t, ok, "Map item type should be a struct")
		assert.Equal(t, 2, structType.NumFields())
		assert.Equal(t, arrow.BinaryTypes.String, structType.Field(0).Type)
		assert.Equal(t, arrow.PrimitiveTypes.Int32, structType.Field(1).Type)
	})

	t.Run("nested list", func(t *testing.T) {
		result, err := tm.CassandraToArrowType("list<list<int>>")
		require.NoError(t, err)
		outerList, ok := result.(*arrow.ListType)
		require.True(t, ok)
		innerList, ok := outerList.Elem().(*arrow.ListType)
		require.True(t, ok)
		assert.Equal(t, arrow.PrimitiveTypes.Int32, innerList.Elem())
	})

	t.Run("frozen types", func(t *testing.T) {
		result, err := tm.CassandraToArrowType("frozen<list<text>>")
		require.NoError(t, err)
		listType, ok := result.(*arrow.ListType)
		require.True(t, ok)
		assert.Equal(t, arrow.BinaryTypes.String, listType.Elem())
	})
}

func TestConvertValue(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		name       string
		value      interface{}
		arrowType  arrow.DataType
		expected   interface{}
		shouldFail bool
	}{
		// Integer conversions
		{"int8", int8(42), arrow.PrimitiveTypes.Int8, int8(42), false},
		{"int16", int16(1000), arrow.PrimitiveTypes.Int16, int16(1000), false},
		{"int32", int32(100000), arrow.PrimitiveTypes.Int32, int32(100000), false},
		{"int64", int64(1000000), arrow.PrimitiveTypes.Int64, int64(1000000), false},

		// Float conversions
		{"float32", float32(3.14), arrow.PrimitiveTypes.Float32, float32(3.14), false},
		{"float64", float64(3.14159), arrow.PrimitiveTypes.Float64, float64(3.14159), false},

		// Boolean conversion
		{"bool true", true, arrow.FixedWidthTypes.Boolean, true, false},
		{"bool false", false, arrow.FixedWidthTypes.Boolean, false, false},

		// String conversion
		{"string", "hello", arrow.BinaryTypes.String, "hello", false},

		// Binary conversion
		{"binary", []byte("data"), arrow.BinaryTypes.Binary, []byte("data"), false},

		// Nil value
		{"nil", nil, arrow.BinaryTypes.String, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tm.ConvertValue(tt.value, tt.arrowType)
			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestUUIDConversion(t *testing.T) {
	tm := NewTypeMapper()

	uuid := gocql.TimeUUID()
	result, err := tm.ConvertValue(uuid, arrow.BinaryTypes.String)
	require.NoError(t, err)

	strResult, ok := result.(string)
	require.True(t, ok)
	assert.NotEmpty(t, strResult)
	assert.Equal(t, uuid.String(), strResult)
}

func TestTimeConversion(t *testing.T) {
	tm := NewTypeMapper()

	now := time.Now()

	t.Run("timestamp", func(t *testing.T) {
		result, err := tm.ConvertValue(now, arrow.FixedWidthTypes.Timestamp_ms)
		require.NoError(t, err)

		timestamp, ok := result.(arrow.Timestamp)
		require.True(t, ok)
		assert.Equal(t, arrow.Timestamp(now.UnixMilli()), timestamp)
	})

	t.Run("date32", func(t *testing.T) {
		result, err := tm.ConvertValue(now, arrow.FixedWidthTypes.Date32)
		require.NoError(t, err)

		date, ok := result.(arrow.Date32)
		require.True(t, ok)
		expectedDays := now.Unix() / 86400
		assert.Equal(t, arrow.Date32(expectedDays), date)
	})
}

func TestCreateArrowSchema(t *testing.T) {
	tm := NewTypeMapper()

	columnNames := []string{"id", "name", "age", "active", "created_at"}
	columnTypes := []string{"uuid", "text", "int", "boolean", "timestamp"}

	schema, err := tm.CreateArrowSchema(columnNames, columnTypes)
	require.NoError(t, err)
	require.NotNil(t, schema)

	assert.Equal(t, len(columnNames), schema.NumFields())

	// Check field names
	for i, name := range columnNames {
		field := schema.Field(i)
		assert.Equal(t, name, field.Name)
		assert.True(t, field.Nullable)
	}

	// Check field types
	assert.Equal(t, arrow.BinaryTypes.String, schema.Field(0).Type)     // uuid
	assert.Equal(t, arrow.BinaryTypes.String, schema.Field(1).Type)     // text
	assert.Equal(t, arrow.PrimitiveTypes.Int32, schema.Field(2).Type)   // int
	assert.Equal(t, arrow.FixedWidthTypes.Boolean, schema.Field(3).Type) // boolean
	assert.Equal(t, arrow.FixedWidthTypes.Timestamp_ms, schema.Field(4).Type) // timestamp
}

func TestCreateArrowSchemaMismatch(t *testing.T) {
	tm := NewTypeMapper()

	columnNames := []string{"id", "name"}
	columnTypes := []string{"uuid"} // Mismatch: fewer types than names

	schema, err := tm.CreateArrowSchema(columnNames, columnTypes)
	assert.Error(t, err)
	assert.Nil(t, schema)
}

func TestSplitTypeParams(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"text, int", []string{"text", " int"}},
		{"map<text, int>, list<text>", []string{"map<text, int>", " list<text>"}},
		{"text", []string{"text"}},
		{"", []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitTypeParams(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTypeParam(t *testing.T) {
	tests := []struct {
		typeStr  string
		prefix   string
		suffix   string
		expected string
	}{
		{"list<int>", "list<", ">", "int"},
		{"set<text>", "set<", ">", "text"},
		{"map<text, int>", "map<", ">", "text, int"},
		{"frozen<list<int>>", "frozen<", ">", "list<int>"},
		{"invalid", "list<", ">", ""},
	}

	for _, tt := range tests {
		t.Run(tt.typeStr, func(t *testing.T) {
			result := extractTypeParam(tt.typeStr, tt.prefix, tt.suffix)
			assert.Equal(t, tt.expected, result)
		})
	}
}