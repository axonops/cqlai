package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Phase 5: DDL Enhancements Tests
// ============================================================================

// TestRenderCreateIndex_IfNotExists tests CREATE INDEX IF NOT EXISTS
func TestRenderCreateIndex_IfNotExists(t *testing.T) {
	plan := &AIResult{
		Operation: "CREATE",
		Keyspace:  "test_ks",
		Table:     "users",
		Options: map[string]any{
			"object_type":    "INDEX",
			"index_name":     "users_email_idx",
			"column":         "email",
			"if_not_exists":  true,
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "CREATE INDEX IF NOT EXISTS users_email_idx")
}

// TestRenderCreateType_IfNotExists tests CREATE TYPE IF NOT EXISTS
func TestRenderCreateType_IfNotExists(t *testing.T) {
	plan := &AIResult{
		Operation: "CREATE",
		Keyspace:  "test_ks",
		Schema:    map[string]string{"street": "text", "city": "text"},
		Options: map[string]any{
			"object_type":   "TYPE",
			"type_name":     "address",
			"if_not_exists": true,
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "CREATE TYPE IF NOT EXISTS")
}

// TestRenderCreateRole_IfNotExists tests CREATE ROLE IF NOT EXISTS
func TestRenderCreateRole_IfNotExists(t *testing.T) {
	plan := &AIResult{
		Operation: "CREATE",
		Options: map[string]any{
			"object_type":   "ROLE",
			"role_name":     "test_role",
			"if_not_exists": true,
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "CREATE ROLE IF NOT EXISTS test_role")
}

// TestRenderDropTable_IfExists tests DROP TABLE IF EXISTS
func TestRenderDropTable_IfExists(t *testing.T) {
	plan := &AIResult{
		Operation: "DROP",
		Keyspace:  "test_ks",
		Table:     "users",
		Options: map[string]any{
			"if_exists": true,
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "DROP TABLE IF EXISTS")
}

// TestRenderDropIndex_IfExists tests DROP INDEX IF EXISTS
func TestRenderDropIndex_IfExists(t *testing.T) {
	plan := &AIResult{
		Operation: "DROP",
		Keyspace:  "test_ks",
		Options: map[string]any{
			"object_type": "INDEX",
			"index_name":  "users_email_idx",
			"if_exists":   true,
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "DROP INDEX IF EXISTS")
}

// TestRenderCreateCustomIndex tests CREATE CUSTOM INDEX (SAI)
func TestRenderCreateCustomIndex(t *testing.T) {
	plan := &AIResult{
		Operation: "CREATE",
		Keyspace:  "test_ks",
		Table:     "users",
		Options: map[string]any{
			"object_type":  "INDEX",
			"index_name":   "users_sai_idx",
			"column":       "email",
			"custom_index": true,
			"using_class":  "StorageAttachedIndex",
		},
	}

	got, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, got, "CREATE CUSTOM INDEX")
	assert.Contains(t, got, "USING 'StorageAttachedIndex'")
}
