package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUDTRegistry(t *testing.T) {
	// Note: Most tests require a real Cassandra connection since the simplified
	// registry relies on gocql's KeyspaceMetadata API

	t.Run("creation", func(t *testing.T) {
		// This test can run without a session
		registry := NewUDTRegistry(nil)
		assert.NotNil(t, registry)
	})

	t.Run("operations without session", func(t *testing.T) {
		registry := NewUDTRegistry(nil)

		// Operations should fail gracefully without a session
		_, err := registry.GetUDTDefinition("test_ks", "address")
		assert.Error(t, err)

		_, err = registry.GetUDTDefinitionOrLoad("test_ks", "address")
		assert.Error(t, err)

		// These are no-ops in the simplified version
		registry.Clear()
		registry.ClearKeyspace("test_ks")

		// Should return false without a session
		assert.False(t, registry.HasUDT("test_ks", "address"))

		// Should return nil without a session
		assert.Nil(t, registry.GetAllUDTs("test_ks"))
	})
}

func TestUDTDefinition(t *testing.T) {
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
		field, idx, err := udtDef.GetFieldByName("city")
		require.NoError(t, err)
		assert.Equal(t, 1, idx)
		assert.Equal(t, "city", field.Name)
		assert.Equal(t, "text", field.TypeStr)

		// Test non-existent field
		_, _, err = udtDef.GetFieldByName("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field nonexistent not found")
	})
}