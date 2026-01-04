package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Deterministic Rendering Tests (Added 2026-01-04)
// ============================================================================
//
// These tests verify that CQL rendering is deterministic - same logical
// query always produces identical CQL string, regardless of Go map iteration order.
//
// CRITICAL: Without deterministic rendering, exact CQL assertions are impossible
// because Go maps iterate in random order.
// ============================================================================

// TestRenderInsert_DeterministicColumnOrder verifies INSERT columns are sorted alphabetically
func TestRenderInsert_DeterministicColumnOrder(t *testing.T) {
	plan := &AIResult{
		Operation: "INSERT",
		Table:     "users",
		Values: map[string]any{
			"z_last":   "last",
			"a_first":  "first",
			"m_middle": "middle",
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)

	// Columns should be alphabetically sorted: a_first, m_middle, z_last
	assert.Contains(t, got, "(a_first, m_middle, z_last)")
	assert.Contains(t, got, "VALUES ('first', 'middle', 'last')")
}

// TestRenderInsert_DeterministicMultipleRuns verifies same plan produces identical CQL
func TestRenderInsert_DeterministicMultipleRuns(t *testing.T) {
	plan := &AIResult{
		Operation: "INSERT",
		Table:     "users",
		Values: map[string]any{
			"col3": "c",
			"col1": "a",
			"col2": "b",
		},
	}

	// Run 10 times - should get identical output
	var results []string
	for i := 0; i < 10; i++ {
		got, err := RenderCQL(plan)
		assert.NoError(t, err)
		results = append(results, got)
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0], results[i], "Run %d should match run 1", i)
	}
}

// TestRenderUpdate_DeterministicSetClauseOrder verifies UPDATE SET columns are sorted
func TestRenderUpdate_DeterministicSetClauseOrder(t *testing.T) {
	plan := &AIResult{
		Operation: "UPDATE",
		Table:     "users",
		Values: map[string]any{
			"z_last":   "last",
			"a_first":  "first",
			"m_middle": "middle",
		},
		Where: []WhereClause{
			{Column: "id", Operator: "=", Value: 1},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)

	// SET clauses should be alphabetically sorted: a_first, m_middle, z_last
	assert.Contains(t, got, "SET a_first = 'first', m_middle = 'middle', z_last = 'last'")
}

// TestRenderUpdate_DeterministicCounterOps verifies counter operations are sorted
func TestRenderUpdate_DeterministicCounterOps(t *testing.T) {
	plan := &AIResult{
		Operation: "UPDATE",
		Table:     "counters",
		CounterOps: map[string]string{
			"views":   "+10",
			"clicks":  "+5",
			"actions": "+3",
		},
		Where: []WhereClause{
			{Column: "id", Operator: "=", Value: "page1"},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)

	// Counter operations should be alphabetically sorted: actions, clicks, views
	assert.Contains(t, got, "SET actions = actions + 3, clicks = clicks + 5, views = views + 10")
}

// TestFormatMap_DeterministicKeyOrder verifies map keys are sorted alphabetically
func TestFormatMap_DeterministicKeyOrder(t *testing.T) {
	plan := &AIResult{
		Operation: "INSERT",
		Table:     "users",
		Values: map[string]any{
			"id": 1,
			"settings": map[string]any{
				"z_key": "z",
				"a_key": "a",
				"m_key": "m",
			},
		},
		ValueTypes: map[string]string{
			"settings": "map<text,text>",
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)

	// Map keys should be alphabetically sorted: a_key, m_key, z_key
	assert.Contains(t, got, "{'a_key': 'a', 'm_key': 'm', 'z_key': 'z'}")
}

// TestFormatUDT_DeterministicFieldOrder verifies UDT fields are sorted alphabetically
func TestFormatUDT_DeterministicFieldOrder(t *testing.T) {
	plan := &AIResult{
		Operation: "INSERT",
		Table:     "users",
		Values: map[string]any{
			"id": 1,
			"addr": map[string]any{
				"zip":    "10001",
				"city":   "NYC",
				"street": "123 Main St",
			},
		},
		ValueTypes: map[string]string{
			"addr": "frozen<address>",
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)

	// UDT fields should be alphabetically sorted: city, street, zip
	assert.Contains(t, got, "{city: 'NYC', street: '123 Main St', zip: '10001'}")
}

// TestFormatSet_DeterministicElementOrder verifies set elements are sorted alphabetically
func TestFormatSet_DeterministicElementOrder(t *testing.T) {
	plan := &AIResult{
		Operation: "INSERT",
		Table:     "users",
		Values: map[string]any{
			"id":   1,
			"tags": []string{"zebra", "apple", "middle"},
		},
		ValueTypes: map[string]string{
			"tags": "set<text>",
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)

	// Set elements should be alphabetically sorted: apple, middle, zebra
	assert.Contains(t, got, "{'apple', 'middle', 'zebra'}")
}
