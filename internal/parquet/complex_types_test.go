package parquet

import (
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendListValue(t *testing.T) {
	tm := NewTypeMapper()
	allocator := memory.DefaultAllocator

	tests := []struct {
		name        string
		value       interface{}
		elementType arrow.DataType
		expected    []interface{}
		expectError bool
	}{
		{
			name:        "nil value",
			value:       nil,
			elementType: arrow.PrimitiveTypes.Int32,
			expected:    nil,
		},
		{
			name:        "[]interface{} with strings",
			value:       []interface{}{"a", "b", "c"},
			elementType: arrow.BinaryTypes.String,
			expected:    []interface{}{"a", "b", "c"},
		},
		{
			name:        "[]string",
			value:       []string{"x", "y", "z"},
			elementType: arrow.BinaryTypes.String,
			expected:    []interface{}{"x", "y", "z"},
		},
		{
			name:        "[]int",
			value:       []int{1, 2, 3},
			elementType: arrow.PrimitiveTypes.Int32,
			expected:    []interface{}{1, 2, 3},
		},
		{
			name:        "[]float64",
			value:       []float64{1.1, 2.2, 3.3},
			elementType: arrow.PrimitiveTypes.Float64,
			expected:    []interface{}{1.1, 2.2, 3.3},
		},
		{
			name:        "[]bool",
			value:       []bool{true, false, true},
			elementType: arrow.FixedWidthTypes.Boolean,
			expected:    []interface{}{true, false, true},
		},
		{
			name:        "mixed types in []interface{}",
			value:       []interface{}{1, 2, 3},
			elementType: arrow.PrimitiveTypes.Int32,
			expected:    []interface{}{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listType := arrow.ListOf(tt.elementType)
			builder := array.NewListBuilder(allocator, tt.elementType)
			defer builder.Release()

			err := tm.AppendListValue(builder, tt.value, tt.elementType)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Build the array
			arr := builder.NewArray().(*array.List)
			defer arr.Release()

			if tt.value == nil {
				assert.True(t, arr.IsNull(0))
			} else {
				assert.False(t, arr.IsNull(0))
				// Verify the list was created
				assert.Equal(t, 1, arr.Len())
			}
			_ = listType // Use listType to avoid unused variable warning
		})
	}
}

func TestAppendMapValue(t *testing.T) {
	tm := NewTypeMapper()
	allocator := memory.DefaultAllocator

	tests := []struct {
		name        string
		value       interface{}
		keyType     arrow.DataType
		valueType   arrow.DataType
		expectedLen int
		expectError bool
	}{
		{
			name:        "nil value",
			value:       nil,
			keyType:     arrow.BinaryTypes.String,
			valueType:   arrow.BinaryTypes.String,
			expectedLen: 0,
		},
		{
			name: "map[string]interface{}",
			value: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			keyType:     arrow.BinaryTypes.String,
			valueType:   arrow.BinaryTypes.String,
			expectedLen: 2,
		},
		{
			name: "map[string]string",
			value: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
			},
			keyType:     arrow.BinaryTypes.String,
			valueType:   arrow.BinaryTypes.String,
			expectedLen: 3,
		},
		{
			name: "map[string]int",
			value: map[string]int{
				"x": 10,
				"y": 20,
			},
			keyType:     arrow.BinaryTypes.String,
			valueType:   arrow.PrimitiveTypes.Int32,
			expectedLen: 2,
		},
		{
			name: "map[interface{}]interface{}",
			value: map[interface{}]interface{}{
				"foo": "bar",
				"baz": "qux",
			},
			keyType:     arrow.BinaryTypes.String,
			valueType:   arrow.BinaryTypes.String,
			expectedLen: 2,
		},
		{
			name:        "unsupported map type",
			value:       map[int]string{1: "one"},
			keyType:     arrow.PrimitiveTypes.Int32,
			valueType:   arrow.BinaryTypes.String,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := array.NewMapBuilder(allocator, tt.keyType, tt.valueType, false)
			defer builder.Release()

			err := tm.AppendMapValue(builder, tt.value, tt.keyType, tt.valueType)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Build the array
			arr := builder.NewArray().(*array.Map)
			defer arr.Release()

			if tt.value == nil {
				assert.True(t, arr.IsNull(0))
			} else {
				assert.False(t, arr.IsNull(0))
				// Verify the map was created
				assert.Equal(t, 1, arr.Len())
			}
		})
	}
}

func TestAppendStructValue(t *testing.T) {
	tm := NewTypeMapper()
	allocator := memory.DefaultAllocator

	fields := []arrow.Field{
		{Name: "field1", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "field2", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
		{Name: "field3", Type: arrow.FixedWidthTypes.Boolean, Nullable: true},
	}

	tests := []struct {
		name        string
		value       interface{}
		expectError bool
		checkValues bool
	}{
		{
			name:  "nil value",
			value: nil,
		},
		{
			name: "map[string]interface{} (UDT)",
			value: map[string]interface{}{
				"field1": "test",
				"field2": int32(42),
				"field3": true,
			},
			checkValues: true,
		},
		{
			name: "[]interface{} (Tuple)",
			value: []interface{}{
				"hello",
				int32(100),
				false,
			},
			checkValues: true,
		},
		{
			name: "Tuple with fewer values than fields",
			value: []interface{}{
				"partial",
				int32(50),
			},
			checkValues: true,
		},
		{
			name:        "unsupported struct type",
			value:       "not a struct",
			expectError: false, // Now handles gracefully by appending nulls
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structType := arrow.StructOf(fields...)
			builder := array.NewStructBuilder(allocator, structType)
			defer builder.Release()

			err := tm.AppendStructValue(builder, tt.value, fields)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Build the array
			arr := builder.NewArray().(*array.Struct)
			defer arr.Release()

			if tt.value == nil {
				assert.True(t, arr.IsNull(0))
			} else {
				assert.False(t, arr.IsNull(0))
				assert.Equal(t, 1, arr.Len())
			}
		})
	}
}

func TestParseUDTType(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		name          string
		udtSpec       string
		expectedField int
		fieldChecks   []struct {
			name string
			typ  arrow.DataType
		}
	}{
		{
			name:          "simple UDT",
			udtSpec:       "udt<name:text,age:int,active:boolean>",
			expectedField: 3,
			fieldChecks: []struct {
				name string
				typ  arrow.DataType
			}{
				{"name", arrow.BinaryTypes.String},
				{"age", arrow.PrimitiveTypes.Int32},
				{"active", arrow.FixedWidthTypes.Boolean},
			},
		},
		{
			name:          "nested UDT",
			udtSpec:       "udt<id:uuid,items:list<text>>",
			expectedField: 2,
			fieldChecks: []struct {
				name string
				typ  arrow.DataType
			}{
				{"id", &arrow.FixedSizeBinaryType{ByteWidth: 16}},
				{"items", arrow.ListOf(arrow.BinaryTypes.String)},
			},
		},
		{
			name:          "named UDT (fallback)",
			udtSpec:       "myudt",
			expectedField: 1,
			fieldChecks: []struct {
				name string
				typ  arrow.DataType
			}{
				{"value", arrow.BinaryTypes.String},
			},
		},
		{
			name:          "empty UDT",
			udtSpec:       "udt<>",
			expectedField: 1,
			fieldChecks: []struct {
				name string
				typ  arrow.DataType
			}{
				{"value", arrow.BinaryTypes.String},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structType, err := tm.ParseUDTType(tt.udtSpec, nil)
			require.NoError(t, err)
			assert.NotNil(t, structType)
			assert.Equal(t, tt.expectedField, structType.NumFields())

			for i, check := range tt.fieldChecks {
				field := structType.Field(i)
				assert.Equal(t, check.name, field.Name)
				assert.Equal(t, check.typ.String(), field.Type.String())
			}
		})
	}
}

func TestParseTupleType(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		name          string
		tupleSpec     string
		expectedField int
		expectError   bool
		fieldTypes    []arrow.DataType
	}{
		{
			name:          "simple tuple",
			tupleSpec:     "tuple<text,int,boolean>",
			expectedField: 3,
			fieldTypes: []arrow.DataType{
				arrow.BinaryTypes.String,
				arrow.PrimitiveTypes.Int32,
				arrow.FixedWidthTypes.Boolean,
			},
		},
		{
			name:          "nested tuple",
			tupleSpec:     "tuple<uuid,list<text>,map<text,int>>",
			expectedField: 3,
			fieldTypes: []arrow.DataType{
				&arrow.FixedSizeBinaryType{ByteWidth: 16},
				arrow.ListOf(arrow.BinaryTypes.String),
				arrow.MapOf(arrow.BinaryTypes.String, arrow.PrimitiveTypes.Int32),
			},
		},
		{
			name:          "single element tuple",
			tupleSpec:     "tuple<timestamp>",
			expectedField: 1,
			fieldTypes: []arrow.DataType{
				arrow.FixedWidthTypes.Timestamp_ms,
			},
		},
		{
			name:        "invalid format",
			tupleSpec:   "not_a_tuple",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structType, err := tm.ParseTupleType(tt.tupleSpec)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, structType)
			assert.Equal(t, tt.expectedField, structType.NumFields())

			for i, expectedType := range tt.fieldTypes {
				field := structType.Field(i)
				assert.Equal(t, expectedType.String(), field.Type.String())
				assert.Equal(t, true, field.Nullable)
			}
		})
	}
}

func TestExtractListValue(t *testing.T) {
	allocator := memory.DefaultAllocator

	// Create a test list
	builder := array.NewListBuilder(allocator, arrow.PrimitiveTypes.Int32)
	valueBuilder := builder.ValueBuilder().(*array.Int32Builder)

	// Add a list [1, 2, 3]
	builder.Append(true)
	valueBuilder.Append(1)
	valueBuilder.Append(2)
	valueBuilder.Append(3)

	// Add a null list
	builder.AppendNull()

	// Add an empty list
	builder.Append(true)

	arr := builder.NewArray().(*array.List)
	defer arr.Release()

	// Test extraction
	val1 := ExtractListValue(arr, 0)
	assert.Equal(t, []interface{}{int32(1), int32(2), int32(3)}, val1)

	val2 := ExtractListValue(arr, 1)
	assert.Nil(t, val2)

	val3 := ExtractListValue(arr, 2)
	assert.Equal(t, []interface{}{}, val3)
}

func TestExtractMapValue(t *testing.T) {
	allocator := memory.DefaultAllocator

	// Create a test map
	builder := array.NewMapBuilder(allocator, arrow.BinaryTypes.String, arrow.PrimitiveTypes.Int32, false)
	keyBuilder := builder.KeyBuilder().(*array.StringBuilder)
	itemBuilder := builder.ItemBuilder().(*array.Int32Builder)

	// Add a map {"a": 1, "b": 2}
	builder.Append(true)
	keyBuilder.Append("a")
	itemBuilder.Append(1)
	keyBuilder.Append("b")
	itemBuilder.Append(2)

	// Add a null map
	builder.AppendNull()

	arr := builder.NewArray().(*array.Map)
	defer arr.Release()

	// Test extraction
	val1 := ExtractMapValue(arr, 0)
	mapVal := val1.(map[interface{}]interface{})
	assert.Equal(t, 2, len(mapVal))
	assert.Equal(t, int32(1), mapVal["a"])
	assert.Equal(t, int32(2), mapVal["b"])

	val2 := ExtractMapValue(arr, 1)
	assert.Nil(t, val2)
}

func TestExtractStructValue(t *testing.T) {
	allocator := memory.DefaultAllocator

	// Create a test struct
	fields := []arrow.Field{
		{Name: "name", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "age", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
	}

	structType := arrow.StructOf(fields...)
	builder := array.NewStructBuilder(allocator, structType)

	nameBuilder := builder.FieldBuilder(0).(*array.StringBuilder)
	ageBuilder := builder.FieldBuilder(1).(*array.Int32Builder)

	// Add a struct {"name": "Alice", "age": 30}
	builder.Append(true)
	nameBuilder.Append("Alice")
	ageBuilder.Append(30)

	// Add a null struct
	builder.AppendNull()

	arr := builder.NewArray().(*array.Struct)
	defer arr.Release()

	// Test extraction
	val1 := ExtractStructValue(arr, 0)
	structVal := val1.(map[string]interface{})
	assert.Equal(t, "Alice", structVal["name"])
	assert.Equal(t, int32(30), structVal["age"])

	val2 := ExtractStructValue(arr, 1)
	assert.Nil(t, val2)
}

func TestComplexTypeIntegration(t *testing.T) {
	// This test verifies that complex types work end-to-end
	// with type mapping, appending, and extraction

	tm := NewTypeMapper()
	allocator := memory.DefaultAllocator

	// Test List<Map<String, Int>>
	t.Run("List of Maps", func(t *testing.T) {
		mapType := arrow.MapOf(arrow.BinaryTypes.String, arrow.PrimitiveTypes.Int32)
		listType := arrow.ListOf(mapType)

		builder := array.NewListBuilder(allocator, mapType)
		defer builder.Release()

		// Create test data: [{a:1, b:2}, {c:3}]
		testData := []interface{}{
			map[string]interface{}{"a": 1, "b": 2},
			map[string]interface{}{"c": 3},
		}

		err := tm.AppendListValue(builder, testData, mapType)
		require.NoError(t, err)

		arr := builder.NewArray().(*array.List)
		defer arr.Release()

		// Extract and verify
		extracted := ExtractListValue(arr, 0)
		require.NotNil(t, extracted)

		// Verify the structure
		extractedList := extracted.([]interface{})
		assert.Equal(t, 2, len(extractedList))

		_ = listType // Use listType to avoid unused variable warning
	})

	// Test Map<String, List<String>>
	t.Run("Map of Lists", func(t *testing.T) {
		listType := arrow.ListOf(arrow.BinaryTypes.String)
		mapType := arrow.MapOf(arrow.BinaryTypes.String, listType)

		builder := array.NewMapBuilder(allocator, arrow.BinaryTypes.String, listType, false)
		defer builder.Release()

		// Create test data: {fruits: ["apple", "banana"], colors: ["red", "blue"]}
		testData := map[string]interface{}{
			"fruits": []string{"apple", "banana"},
			"colors": []string{"red", "blue"},
		}

		err := tm.AppendMapValue(builder, testData, arrow.BinaryTypes.String, listType)
		require.NoError(t, err)

		arr := builder.NewArray().(*array.Map)
		defer arr.Release()

		// Extract and verify
		extracted := ExtractMapValue(arr, 0)
		require.NotNil(t, extracted)

		extractedMap := extracted.(map[interface{}]interface{})
		assert.Equal(t, 2, len(extractedMap))

		_ = mapType // Use mapType to avoid unused variable warning
	})
}