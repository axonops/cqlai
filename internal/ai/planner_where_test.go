package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// WHERE Clause Additional Tests
// ============================================================================

// TestRenderSelect_WhereContains tests WHERE CONTAINS for collections
func TestRenderSelect_WhereContains(t *testing.T) {
	plan := &AIResult{
		Operation: "SELECT",
		Table:     "users",
		Where: []WhereClause{
			{Column: "tags", Operator: "CONTAINS", Value: "admin"},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "WHERE tags CONTAINS 'admin'")
}

// TestRenderSelect_WhereContainsKey tests WHERE CONTAINS KEY for maps
func TestRenderSelect_WhereContainsKey(t *testing.T) {
	plan := &AIResult{
		Operation: "SELECT",
		Table:     "users",
		Where: []WhereClause{
			{Column: "settings", Operator: "CONTAINS KEY", Value: "theme"},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "WHERE settings CONTAINS KEY 'theme'")
}
