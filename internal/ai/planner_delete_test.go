package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// DELETE Statement Additional Tests
// ============================================================================

// TestRenderDelete_SpecificColumns tests DELETE with column list
func TestRenderDelete_SpecificColumns(t *testing.T) {
	plan := &AIResult{
		Operation: "DELETE",
		Table:     "users",
		Columns:   []string{"email", "name"},
		Where:     []WhereClause{{Column: "id", Operator: "=", Value: 1}},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "DELETE email, name FROM users WHERE id = 1;")
}

// TestRenderDelete_MapElement tests DELETE map[key] (via Columns)
func TestRenderDelete_MapElement(t *testing.T) {
	plan := &AIResult{
		Operation: "DELETE",
		Table:     "users",
		Columns:   []string{"settings['theme']"},
		Where:     []WhereClause{{Column: "id", Operator: "=", Value: 1}},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "DELETE settings['theme'] FROM users")
}

// TestRenderDelete_UDTField tests DELETE udt.field (via Columns)
func TestRenderDelete_UDTField(t *testing.T) {
	plan := &AIResult{
		Operation: "DELETE",
		Table:     "udt_test",
		Columns:   []string{"addr.city"},
		Where:     []WhereClause{{Column: "id", Operator: "=", Value: 1}},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "DELETE addr.city FROM udt_test")
}
