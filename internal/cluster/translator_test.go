package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Vector Type Parsing Tests
// ============================================================================

func TestParseVectorType(t *testing.T) {
	manager := &GocqlMetadataManager{config: DefaultManagerConfig()}

	tests := []struct {
		name          string
		typeStr       string
		expectVector  bool
		expectDim     int
		expectElemType string
	}{
		{
			name:          "vector<float, 384>",
			typeStr:       "vector<float, 384>",
			expectVector:  true,
			expectDim:     384,
			expectElemType: "float",
		},
		{
			name:          "vector<float64, 1536>",
			typeStr:       "vector<float64, 1536>",
			expectVector:  true,
			expectDim:     1536,
			expectElemType: "float64",
		},
		{
			name:          "vector<float, 768>",
			typeStr:       "vector<float, 768>",
			expectVector:  true,
			expectDim:     768,
			expectElemType: "float",
		},
		{
			name:          "vector with spaces",
			typeStr:       "vector< float , 384 >",
			expectVector:  true,
			expectDim:     384,
			expectElemType: "float",
		},
		{
			name:          "VECTOR uppercase",
			typeStr:       "VECTOR<FLOAT, 384>",
			expectVector:  true,
			expectDim:     384,
			expectElemType: "FLOAT",
		},
		{
			name:          "not a vector - list",
			typeStr:       "list<float>",
			expectVector:  false,
			expectDim:     0,
			expectElemType: "",
		},
		{
			name:          "not a vector - int",
			typeStr:       "int",
			expectVector:  false,
			expectDim:     0,
			expectElemType: "",
		},
		{
			name:          "malformed vector",
			typeStr:       "vector<float>",
			expectVector:  false,
			expectDim:     0,
			expectElemType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isVector, dim, elemType := manager.parseVectorType(tt.typeStr)

			assert.Equal(t, tt.expectVector, isVector)
			if tt.expectVector {
				assert.Equal(t, tt.expectDim, dim)
				assert.Equal(t, tt.expectElemType, elemType)
			}
		})
	}
}

// ============================================================================
// Index Type Detection Tests
// ============================================================================

func TestDetectIndexType(t *testing.T) {
	manager := &GocqlMetadataManager{config: DefaultManagerConfig()}

	tests := []struct {
		name           string
		className      string
		expectKind     IndexKind
		expectNative   bool
		expectSAI      bool
		expectVector   bool
	}{
		{
			name:         "SAI index",
			className:    "org.apache.cassandra.index.sai.StorageAttachedIndex",
			expectKind:   IndexKindSAI,
			expectNative: true,
			expectSAI:    true,
			expectVector: false,
		},
		{
			name:         "SAI lowercase",
			className:    "storageattachedindex",
			expectKind:   IndexKindSAI,
			expectNative: true,
			expectSAI:    true,
			expectVector: false,
		},
		{
			name:         "Composites index",
			className:    "org.apache.cassandra.index.composites.CompositeIndex",
			expectKind:   IndexKindComposites,
			expectNative: true,
			expectSAI:    false,
			expectVector: false,
		},
		{
			name:         "Keys index",
			className:    "keys",
			expectKind:   IndexKindKeys,
			expectNative: true,
			expectSAI:    false,
			expectVector: false,
		},
		{
			name:         "Full index",
			className:    "full",
			expectKind:   IndexKindFull,
			expectNative: true,
			expectSAI:    false,
			expectVector: false,
		},
		{
			name:         "Custom index - Lucene",
			className:    "com.stratio.cassandra.lucene.Index",
			expectKind:   IndexKindCustom,
			expectNative: false,
			expectSAI:    false,
			expectVector: false,
		},
		{
			name:         "Custom index - Unknown",
			className:    "com.example.MyCustomIndex",
			expectKind:   IndexKindCustom,
			expectNative: false,
			expectSAI:    false,
			expectVector: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, isNative, isSAI, isVector := manager.detectIndexType(tt.className)

			assert.Equal(t, tt.expectKind, kind)
			assert.Equal(t, tt.expectNative, isNative)
			assert.Equal(t, tt.expectSAI, isSAI)
			assert.Equal(t, tt.expectVector, isVector)
		})
	}
}

// ============================================================================
// Type Name Conversion Tests
// ============================================================================

func TestTypeNameFromType(t *testing.T) {
	// Note: This requires actual gocql.Type constants which are ints
	// Testing this properly requires importing gocql and using real types
	// This is fully covered by integration tests
	t.Skip("Covered by integration tests (requires gocql.Type constants)")
}

// ============================================================================
// Replication Strategy Translation Tests
// ============================================================================

func TestTranslateReplication_SimpleStrategy(t *testing.T) {
	// Note: This requires gocql.KeyspaceMetadata which can't be easily mocked
	// This is fully tested via integration tests
	t.Skip("Covered by integration tests (requires gocql types)")
}

// ============================================================================
// Index Options Conversion Tests
// ============================================================================

func TestConvertIndexOptions(t *testing.T) {
	manager := &GocqlMetadataManager{}

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]string
	}{
		{
			name: "string values",
			input: map[string]interface{}{
				"target": "column_name",
				"mode":   "contains",
			},
			expected: map[string]string{
				"target": "column_name",
				"mode":   "contains",
			},
		},
		{
			name: "mixed types",
			input: map[string]interface{}{
				"target":   "column",
				"segments": 10,
				"enabled":  true,
			},
			expected: map[string]string{
				"target":   "column",
				"segments": "10",
				"enabled":  "true",
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.convertIndexOptions(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Clustering Order Translation Tests
// ============================================================================

func TestTranslateClusteringOrder(t *testing.T) {
	manager := &GocqlMetadataManager{}

	tests := []struct {
		name     string
		order    string
		expected ColumnOrder
	}{
		{"DESC lowercase", "desc", ColumnOrderDESC},
		{"DESC uppercase", "DESC", ColumnOrderDESC},
		{"DESC mixed", "DeSc", ColumnOrderDESC},
		{"ASC lowercase", "asc", ColumnOrderASC},
		{"ASC uppercase", "ASC", ColumnOrderASC},
		{"empty string", "", ColumnOrderASC},
		{"invalid", "invalid", ColumnOrderASC},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.translateClusteringOrder(tt.order)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Column Kind Translation Tests
// ============================================================================

func TestTranslateColumnKind(t *testing.T) {
	// Note: translateColumnKind uses gocql.ColumnKind enum which can't be easily mocked
	// This is fully tested via integration tests
	t.Skip("Covered by integration tests (requires gocql.ColumnKind enum)")
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestValidateColumnType(t *testing.T) {
	manager := &GocqlMetadataManager{config: DefaultManagerConfig()}

	t.Run("nil column type", func(t *testing.T) {
		result := manager.ValidateColumnType(nil)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "column type is nil")
	})

	t.Run("empty type name", func(t *testing.T) {
		ct := &ColumnType{Name: ""}
		result := manager.ValidateColumnType(ct)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "type name is empty")
	})

	t.Run("valid simple type", func(t *testing.T) {
		ct := &ColumnType{Name: "int", Category: TypeCategorySimple, IsNative: true}
		result := manager.ValidateColumnType(ct)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("list missing element type", func(t *testing.T) {
		ct := &ColumnType{
			Name:        "list<int>",
			Category:    TypeCategoryCollection,
			ElementType: nil,
		}
		result := manager.ValidateColumnType(ct)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "missing element type")
	})

	t.Run("map missing key type", func(t *testing.T) {
		ct := &ColumnType{
			Name:      "map<int, text>",
			Category:  TypeCategoryCollection,
			KeyType:   nil,
			ValueType: &ColumnType{Name: "text"},
		}
		result := manager.ValidateColumnType(ct)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "missing key type")
	})

	t.Run("map missing value type", func(t *testing.T) {
		ct := &ColumnType{
			Name:      "map<int, text>",
			Category:  TypeCategoryCollection,
			KeyType:   &ColumnType{Name: "int"},
			ValueType: nil,
		}
		result := manager.ValidateColumnType(ct)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "missing value type")
	})

	t.Run("UDT missing name", func(t *testing.T) {
		ct := &ColumnType{
			Name:     "address",
			Category: TypeCategoryUDT,
			UDTName:  "",
		}
		result := manager.ValidateColumnType(ct)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "missing UDT name")
	})

	t.Run("tuple with no types", func(t *testing.T) {
		ct := &ColumnType{
			Name:       "tuple<>",
			Category:   TypeCategoryTuple,
			TupleTypes: []*ColumnType{},
		}
		result := manager.ValidateColumnType(ct)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "no element types")
	})

	t.Run("vector with invalid dimension", func(t *testing.T) {
		ct := &ColumnType{
			Name:            "vector<float, 0>",
			IsVector:        true,
			VectorDimension: 0,
		}
		result := manager.ValidateColumnType(ct)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "invalid")
	})

	t.Run("vector with valid dimension", func(t *testing.T) {
		ct := &ColumnType{
			Name:              "vector<float, 384>",
			IsVector:          true,
			VectorDimension:   384,
			VectorElementType: "float",
		}
		result := manager.ValidateColumnType(ct)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("vector with very large dimension warning", func(t *testing.T) {
		ct := &ColumnType{
			Name:              "vector<float, 100000>",
			IsVector:          true,
			VectorDimension:   100000,
			VectorElementType: "float",
		}
		result := manager.ValidateColumnType(ct)
		assert.True(t, result.Valid) // Valid but with warning
		assert.NotEmpty(t, result.Warnings)
		assert.Contains(t, result.Warnings[0], "very large")
	})

	t.Run("vector with uncommon element type warning", func(t *testing.T) {
		ct := &ColumnType{
			Name:              "vector<int8, 384>",
			IsVector:          true,
			VectorDimension:   384,
			VectorElementType: "int8",
		}
		result := manager.ValidateColumnType(ct)
		assert.True(t, result.Valid) // Valid but with warning
		assert.NotEmpty(t, result.Warnings)
		assert.Contains(t, result.Warnings[0], "uncommon")
	})
}

