package db

import (
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSession implements a minimal gocql.Session interface for testing
type MockSession struct{}

func (m *MockSession) Query(query string, values ...interface{}) *gocql.Query {
	// Note: In real tests, we'd need to mock gocql.Query properly
	// This is a simplified version for demonstration
	return nil
}

func TestUDTRegistry_LoadKeyspaceUDTs(t *testing.T) {
	// Note: These tests would require a proper mock of gocql.Session
	// For now, we'll test the logic that doesn't require database access

	t.Run("new registry creation", func(t *testing.T) {
		// This test can run without a session
		registry := &UDTRegistry{
			definitions: make(map[string]map[string]*UDTDefinition),
		}
		assert.NotNil(t, registry)
		assert.NotNil(t, registry.definitions)
	})

	t.Run("clear operations", func(t *testing.T) {
		registry := &UDTRegistry{
			definitions: make(map[string]map[string]*UDTDefinition),
		}

		// Add some test data
		registry.definitions["keyspace1"] = map[string]*UDTDefinition{
			"address": {
				Keyspace: "keyspace1",
				Name:     "address",
				Fields: []UDTField{
					{Name: "street", TypeStr: "text"},
					{Name: "city", TypeStr: "text"},
				},
			},
		}

		registry.definitions["keyspace2"] = map[string]*UDTDefinition{
			"person": {
				Keyspace: "keyspace2",
				Name:     "person",
				Fields: []UDTField{
					{Name: "name", TypeStr: "text"},
					{Name: "age", TypeStr: "int"},
				},
			},
		}

		// Test ClearKeyspace
		registry.ClearKeyspace("keyspace1")
		assert.Nil(t, registry.definitions["keyspace1"])
		assert.NotNil(t, registry.definitions["keyspace2"])

		// Test Clear
		registry.Clear()
		assert.Len(t, registry.definitions, 0)
	})

	t.Run("get UDT definition", func(t *testing.T) {
		registry := &UDTRegistry{
			definitions: make(map[string]map[string]*UDTDefinition),
		}

		addressUDT := &UDTDefinition{
			Keyspace: "test_ks",
			Name:     "address",
			Fields: []UDTField{
				{
					Name:    "street",
					TypeStr: "text",
					TypeInfo: &CQLTypeInfo{
						BaseType: "text",
					},
				},
				{
					Name:    "city",
					TypeStr: "text",
					TypeInfo: &CQLTypeInfo{
						BaseType: "text",
					},
				},
				{
					Name:    "zip",
					TypeStr: "int",
					TypeInfo: &CQLTypeInfo{
						BaseType: "int",
					},
				},
			},
		}

		// Add the UDT to registry
		registry.definitions["test_ks"] = map[string]*UDTDefinition{
			"address": addressUDT,
		}

		// Test successful retrieval
		udt, err := registry.GetUDTDefinition("test_ks", "address")
		require.NoError(t, err)
		assert.Equal(t, addressUDT, udt)

		// Test non-existent UDT
		_, err = registry.GetUDTDefinition("test_ks", "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test non-existent keyspace
		_, err = registry.GetUDTDefinition("nonexistent_ks", "address")
		assert.Error(t, err)
	})

	t.Run("has UDT", func(t *testing.T) {
		registry := &UDTRegistry{
			definitions: make(map[string]map[string]*UDTDefinition),
		}

		registry.definitions["test_ks"] = map[string]*UDTDefinition{
			"address": {
				Keyspace: "test_ks",
				Name:     "address",
			},
		}

		assert.True(t, registry.HasUDT("test_ks", "address"))
		assert.False(t, registry.HasUDT("test_ks", "nonexistent"))
		assert.False(t, registry.HasUDT("nonexistent_ks", "address"))
	})

	t.Run("get all UDTs", func(t *testing.T) {
		registry := &UDTRegistry{
			definitions: make(map[string]map[string]*UDTDefinition),
		}

		addressUDT := &UDTDefinition{
			Keyspace: "test_ks",
			Name:     "address",
		}
		personUDT := &UDTDefinition{
			Keyspace: "test_ks",
			Name:     "person",
		}

		registry.definitions["test_ks"] = map[string]*UDTDefinition{
			"address": addressUDT,
			"person":  personUDT,
		}

		allUDTs := registry.GetAllUDTs("test_ks")
		assert.Len(t, allUDTs, 2)
		assert.Equal(t, addressUDT, allUDTs["address"])
		assert.Equal(t, personUDT, allUDTs["person"])

		// Test non-existent keyspace
		nilUDTs := registry.GetAllUDTs("nonexistent_ks")
		assert.Nil(t, nilUDTs)
	})
}

func TestUDTDefinition_Methods(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		udtDef := &UDTDefinition{
			Keyspace: "test_ks",
			Name:     "address",
			Fields: []UDTField{
				{Name: "street", TypeStr: "text"},
				{Name: "city", TypeStr: "text"},
				{Name: "zip", TypeStr: "int"},
			},
		}

		str := udtDef.String()
		assert.Contains(t, str, "test_ks.address")
		assert.Contains(t, str, "street: text")
		assert.Contains(t, str, "city: text")
		assert.Contains(t, str, "zip: int")
	})

	t.Run("GetFieldByName", func(t *testing.T) {
		udtDef := &UDTDefinition{
			Keyspace: "test_ks",
			Name:     "address",
			Fields: []UDTField{
				{Name: "street", TypeStr: "text"},
				{Name: "city", TypeStr: "text"},
				{Name: "zip", TypeStr: "int"},
			},
		}

		// Test existing field
		field, index, err := udtDef.GetFieldByName("city")
		require.NoError(t, err)
		assert.Equal(t, "city", field.Name)
		assert.Equal(t, "text", field.TypeStr)
		assert.Equal(t, 1, index)

		// Test first field
		field, index, err = udtDef.GetFieldByName("street")
		require.NoError(t, err)
		assert.Equal(t, "street", field.Name)
		assert.Equal(t, 0, index)

		// Test last field
		field, index, err = udtDef.GetFieldByName("zip")
		require.NoError(t, err)
		assert.Equal(t, "zip", field.Name)
		assert.Equal(t, 2, index)

		// Test non-existent field
		_, _, err = udtDef.GetFieldByName("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field nonexistent not found")
	})
}

func TestUDTField_ParsedTypes(t *testing.T) {
	t.Run("complex field types", func(t *testing.T) {
		testCases := []struct {
			name      string
			fieldType string
			validate  func(*testing.T, *CQLTypeInfo)
		}{
			{
				name:      "simple text field",
				fieldType: "text",
				validate: func(t *testing.T, ti *CQLTypeInfo) {
					assert.Equal(t, "text", ti.BaseType)
					assert.False(t, ti.Frozen)
				},
			},
			{
				name:      "frozen list field",
				fieldType: "frozen<list<int>>",
				validate: func(t *testing.T, ti *CQLTypeInfo) {
					assert.Equal(t, "list", ti.BaseType)
					assert.True(t, ti.Frozen)
					assert.Len(t, ti.Parameters, 1)
					assert.Equal(t, "int", ti.Parameters[0].BaseType)
				},
			},
			{
				name:      "map field",
				fieldType: "map<text, frozen<address>>",
				validate: func(t *testing.T, ti *CQLTypeInfo) {
					assert.Equal(t, "map", ti.BaseType)
					assert.False(t, ti.Frozen)
					assert.Len(t, ti.Parameters, 2)
					assert.Equal(t, "text", ti.Parameters[0].BaseType)
					assert.Equal(t, "udt", ti.Parameters[1].BaseType)
					assert.Equal(t, "address", ti.Parameters[1].UDTName)
					assert.True(t, ti.Parameters[1].Frozen)
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				typeInfo, err := ParseCQLType(tc.fieldType)
				require.NoError(t, err)
				tc.validate(t, typeInfo)
			})
		}
	})
}