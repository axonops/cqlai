package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRenderUpdate_UDTFieldUpdate tests UDT field update (udt.field = value)
func TestRenderUpdate_UDTFieldUpdate(t *testing.T) {
	plan := &AIResult{
		Operation: "UPDATE",
		Table:     "users",
		CollectionOps: map[string]CollectionOp{
			"address": {
				Operation: "set_field",
				Key:       "city", // Field name
				Value:     "LA",
				ValueType: "text",
			},
		},
		Where: []WhereClause{{Column: "id", Operator: "=", Value: 1}},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "address.city = 'LA'")
}
