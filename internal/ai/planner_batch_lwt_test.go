package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// BATCH with LWT Tests
// ============================================================================
// Tests for single-partition LWT constraint in BATCH statements
// CQL Rule: "Batch with conditions cannot span multiple partitions"
// ============================================================================

// TestRenderBatch_LWT_SinglePartition tests valid single-partition conditional batch
func TestRenderBatch_LWT_SinglePartition(t *testing.T) {
	// All statements target same partition key (id = 100)
	plan := &AIResult{
		Operation: "BATCH",
		BatchType: "LOGGED",
		BatchStatements: []AIResult{
			{
				Operation:   "INSERT",
				Table:       "users",
				Values:      map[string]any{"id": 100, "name": "Alice"},
				IfNotExists: true, // Conditional
			},
			{
				Operation: "UPDATE",
				Table:     "users",
				Values:    map[string]any{"email": "alice@example.com"},
				Where:     []WhereClause{{Column: "id", Operator: "=", Value: 100}},
				// No IF clause, but same partition key
			},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err, "Single-partition LWT batch should succeed")
	assert.Contains(t, got, "IF NOT EXISTS")
	assert.Contains(t, got, "BEGIN BATCH")
}

// TestRenderBatch_LWT_MultiPartition_ShouldError tests invalid multi-partition conditional batch
func TestRenderBatch_LWT_MultiPartition_ShouldError(t *testing.T) {
	// Statements target DIFFERENT partition keys (100 vs 200)
	plan := &AIResult{
		Operation: "BATCH",
		BatchType: "LOGGED",
		BatchStatements: []AIResult{
			{
				Operation:   "INSERT",
				Table:       "users",
				Values:      map[string]any{"id": 100, "name": "Alice"},
				IfNotExists: true, // Conditional
			},
			{
				Operation: "INSERT",
				Table:     "users",
				Values:    map[string]any{"id": 200, "name": "Bob"},
				// Different partition key!
			},
		},
	}

	_, err := RenderCQL(plan)
	assert.Error(t, err, "Multi-partition LWT batch should be rejected")
	assert.Contains(t, err.Error(), "cannot span multiple partitions")
}

// TestRenderBatch_LWT_SamePartitionKey tests batch with matching partition keys
func TestRenderBatch_LWT_SamePartitionKey(t *testing.T) {
	// Both UPDATE statements target id = 50
	plan := &AIResult{
		Operation: "BATCH",
		BatchType: "LOGGED",
		BatchStatements: []AIResult{
			{
				Operation: "UPDATE",
				Table:     "users",
				Values:    map[string]any{"email": "updated@example.com"},
				Where:     []WhereClause{{Column: "id", Operator: "=", Value: 50}},
				IfExists:  true, // Conditional
			},
			{
				Operation: "UPDATE",
				Table:     "users",
				Values:    map[string]any{"name": "Updated Name"},
				Where:     []WhereClause{{Column: "id", Operator: "=", Value: 50}},
				// Same partition key
			},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err, "Same partition key should succeed")
	assert.Contains(t, got, "IF EXISTS")
}

// TestRenderBatch_LWT_DifferentPartitionKeys tests validation catches different keys
func TestRenderBatch_LWT_DifferentPartitionKeys(t *testing.T) {
	// Different partition keys in WHERE clauses
	plan := &AIResult{
		Operation: "BATCH",
		BatchType: "LOGGED",
		BatchStatements: []AIResult{
			{
				Operation: "UPDATE",
				Table:     "users",
				Values:    map[string]any{"email": "a@example.com"},
				Where:     []WhereClause{{Column: "id", Operator: "=", Value: 100}},
				IfExists:  true,
			},
			{
				Operation: "UPDATE",
				Table:     "users",
				Values:    map[string]any{"email": "b@example.com"},
				Where:     []WhereClause{{Column: "id", Operator: "=", Value: 200}},
				// Different partition key!
			},
		},
	}

	_, err := RenderCQL(plan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot span multiple partitions")
}

// TestRenderBatch_NoLWT_NoValidation tests non-conditional batch doesn't validate
func TestRenderBatch_NoLWT_NoValidation(t *testing.T) {
	// No IF clauses, so different partition keys are OK
	plan := &AIResult{
		Operation: "BATCH",
		BatchType: "LOGGED",
		BatchStatements: []AIResult{
			{
				Operation: "INSERT",
				Table:     "users",
				Values:    map[string]any{"id": 1, "name": "Alice"},
				// No IF clause
			},
			{
				Operation: "INSERT",
				Table:     "users",
				Values:    map[string]any{"id": 2, "name": "Bob"},
				// Different partition, but no IF clause - OK for regular batch
			},
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err, "Non-conditional batch can span partitions")
	assert.Contains(t, got, "BEGIN BATCH")
}
