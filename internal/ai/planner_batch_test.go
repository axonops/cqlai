package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Phase 4: BATCH Statement Tests
// ============================================================================

// TestRenderBatch_Logged tests basic LOGGED BATCH
func TestRenderBatch_Logged(t *testing.T) {
	plan := &AIResult{
		Operation: "BATCH",
		BatchType: "LOGGED",
		BatchStatements: []AIResult{
			{
				Operation: "INSERT",
				Table:     "users",
				Values:    map[string]any{"id": 1, "name": "Alice"},
			},
			{
				Operation: "UPDATE",
				Table:     "users",
				Values:    map[string]any{"email": "alice@example.com"},
				Where:     []WhereClause{{Column: "id", Operator: "=", Value: 1}},
			},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "BEGIN BATCH")
	assert.Contains(t, got, "APPLY BATCH")
	assert.Contains(t, got, "INSERT INTO users")
	assert.Contains(t, got, "UPDATE users")
	// Statements should NOT have semicolons inside batch
	lines := strings.Split(got, "\n")
	for i, line := range lines {
		if i > 0 && i < len(lines)-1 { // Skip BEGIN and APPLY lines
			if strings.Contains(line, "INSERT") || strings.Contains(line, "UPDATE") {
				assert.NotContains(t, line, ";", "Batch statements should not have semicolons")
			}
		}
	}
}

// TestRenderBatch_Unlogged tests UNLOGGED BATCH
func TestRenderBatch_Unlogged(t *testing.T) {
	plan := &AIResult{
		Operation: "BATCH",
		BatchType: "UNLOGGED",
		BatchStatements: []AIResult{
			{
				Operation: "INSERT",
				Table:     "users",
				Values:    map[string]any{"id": 2, "name": "Bob"},
			},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "BEGIN UNLOGGED BATCH")
	assert.Contains(t, got, "APPLY BATCH")
}

// TestRenderBatch_Counter tests COUNTER BATCH
func TestRenderBatch_Counter(t *testing.T) {
	plan := &AIResult{
		Operation: "BATCH",
		BatchType: "COUNTER",
		BatchStatements: []AIResult{
			{
				Operation:  "UPDATE",
				Table:      "counters",
				CounterOps: map[string]string{"views": "+1"},
				Where:      []WhereClause{{Column: "id", Operator: "=", Value: "c1"}},
			},
			{
				Operation:  "UPDATE",
				Table:      "counters",
				CounterOps: map[string]string{"views": "+2"},
				Where:      []WhereClause{{Column: "id", Operator: "=", Value: "c2"}},
			},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "BEGIN COUNTER BATCH")
	assert.Contains(t, got, "views = views + 1")
	assert.Contains(t, got, "views = views + 2")
}

// TestRenderBatch_Empty tests empty batch (should error)
func TestRenderBatch_Empty(t *testing.T) {
	plan := &AIResult{
		Operation:       "BATCH",
		BatchType:       "LOGGED",
		BatchStatements: []AIResult{},
	}

	_, err := RenderCQL(plan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one statement")
}

// TestRenderBatch_WithTimestamp tests BATCH USING TIMESTAMP
func TestRenderBatch_WithTimestamp(t *testing.T) {
	plan := &AIResult{
		Operation:      "BATCH",
		BatchType:      "LOGGED",
		UsingTimestamp: 1609459200000000,
		BatchStatements: []AIResult{
			{
				Operation: "INSERT",
				Table:     "users",
				Values:    map[string]any{"id": 3, "name": "Charlie"},
			},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "BEGIN BATCH")
	assert.Contains(t, got, "USING TIMESTAMP 1609459200000000")
	assert.Contains(t, got, "APPLY BATCH")
}
