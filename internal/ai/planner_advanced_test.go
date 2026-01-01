package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Phase 6: Advanced Query Features Tests
// ============================================================================

// TestRenderSelect_AggregateCount tests COUNT(*) aggregate function
func TestRenderSelect_AggregateCount(t *testing.T) {
	plan := &AIResult{
		Operation: "SELECT",
		Table:     "users",
		Columns:   []string{"COUNT(*)"},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "SELECT COUNT(*) FROM users;")
}

// TestRenderSelect_WriteTimeAndTTL tests WRITETIME() and TTL() functions
func TestRenderSelect_WriteTimeAndTTL(t *testing.T) {
	plan := &AIResult{
		Operation: "SELECT",
		Table:     "users",
		Columns:   []string{"id", "name", "WRITETIME(name)", "TTL(name)"},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "WRITETIME(name)")
	assert.Contains(t, got, "TTL(name)")
}

// TestRenderSelect_MultipleAggregates tests multiple aggregate functions
func TestRenderSelect_MultipleAggregates(t *testing.T) {
	plan := &AIResult{
		Operation: "SELECT",
		Table:     "stats",
		Columns:   []string{"MIN(value)", "MAX(value)", "AVG(value)"},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "MIN(value)")
	assert.Contains(t, got, "MAX(value)")
	assert.Contains(t, got, "AVG(value)")
}
