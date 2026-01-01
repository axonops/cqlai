package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRenderAlterTable tests ALTER TABLE operations
func TestRenderAlterTable(t *testing.T) {
	tests := []struct {
		name    string
		plan    *AIResult
		want    string
		wantErr bool
	}{
		{
			name: "ALTER TABLE ADD column",
			plan: &AIResult{
				Operation: "ALTER",
				Keyspace:  "test_ks",
				Table:     "users",
				Options: map[string]any{
					"object_type": "TABLE",
					"action":      "ADD",
					"column_name": "age",
					"column_type": "int",
				},
			},
			want:    "ALTER TABLE test_ks.users ADD age int",
			wantErr: false,
		},
		{
			name: "ALTER TABLE DROP column",
			plan: &AIResult{
				Operation: "ALTER",
				Keyspace:  "test_ks",
				Table:     "users",
				Options: map[string]any{
					"object_type": "TABLE",
					"action":      "DROP",
					"column_name": "age",
				},
			},
			want:    "ALTER TABLE test_ks.users DROP age",
			wantErr: false,
		},
		{
			name: "ALTER TABLE RENAME column",
			plan: &AIResult{
				Operation: "ALTER",
				Keyspace:  "test_ks",
				Table:     "users",
				Options: map[string]any{
					"object_type":     "TABLE",
					"action":          "RENAME",
					"old_column_name": "email",
					"new_column_name": "email_address",
				},
			},
			want:    "ALTER TABLE test_ks.users RENAME email TO email_address",
			wantErr: false,
		},
		{
			name: "missing action",
			plan: &AIResult{
				Operation: "ALTER",
				Keyspace:  "test_ks",
				Table:     "users",
				Options: map[string]any{
					"object_type": "TABLE",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCQL(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestRenderList tests LIST operations
func TestRenderList(t *testing.T) {
	tests := []struct {
		name    string
		plan    *AIResult
		want    string
		wantErr bool
	}{
		{
			name: "LIST ROLES",
			plan: &AIResult{
				Operation: "LIST",
				Options: map[string]any{
					"object_type": "ROLES",
				},
			},
			want:    "LIST ROLES",
			wantErr: false,
		},
		{
			name: "LIST USERS",
			plan: &AIResult{
				Operation: "LIST",
				Options: map[string]any{
					"object_type": "USERS",
				},
			},
			want:    "LIST USERS",
			wantErr: false,
		},
		{
			name: "LIST PERMISSIONS OF role",
			plan: &AIResult{
				Operation: "LIST",
				Options: map[string]any{
					"object_type": "PERMISSIONS",
					"role":        "app_admin",
				},
			},
			want:    "LIST PERMISSIONS OF app_admin",
			wantErr: false,
		},
		{
			name: "missing object_type",
			plan: &AIResult{
				Operation: "LIST",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCQL(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestRenderShow tests SHOW operations
func TestRenderShow(t *testing.T) {
	tests := []struct {
		name    string
		plan    *AIResult
		want    string
		wantErr bool
	}{
		{
			name: "SHOW VERSION",
			plan: &AIResult{
				Operation: "SHOW",
				Options: map[string]any{
					"show_type": "VERSION",
				},
			},
			want:    "SHOW VERSION",
			wantErr: false,
		},
		{
			name: "SHOW HOST",
			plan: &AIResult{
				Operation: "SHOW",
				Options: map[string]any{
					"show_type": "HOST",
				},
			},
			want:    "SHOW HOST",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCQL(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestRenderCreateIndex tests CREATE INDEX
func TestRenderCreateIndex(t *testing.T) {
	plan := &AIResult{
		Operation: "CREATE",
		Keyspace:  "test_ks",
		Table:     "users",
		Options: map[string]any{
			"object_type": "INDEX",
			"index_name":  "users_email_idx",
			"column":      "email",
		},
	}

	cql, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Equal(t, "CREATE INDEX users_email_idx ON test_ks.users (email);", cql)
}

// TestRenderCreateRole tests CREATE ROLE
func TestRenderCreateRole(t *testing.T) {
	plan := &AIResult{
		Operation: "CREATE",
		Options: map[string]any{
			"object_type": "ROLE",
			"role_name":   "app_viewer",
			"password":    "secret123",
			"login":       true,
			"superuser":   false,
		},
	}

	cql, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, cql, "CREATE ROLE app_viewer")
	assert.Contains(t, cql, "PASSWORD = 'secret123'")
	assert.Contains(t, cql, "LOGIN = true")
}

// TestRenderGrantRole tests GRANT ROLE (vs GRANT permission)
func TestRenderGrantRole(t *testing.T) {
	plan := &AIResult{
		Operation: "GRANT",
		Options: map[string]any{
			"grant_type": "ROLE",
			"role":       "developer",
			"to_role":    "app_admin",
		},
	}

	cql, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Equal(t, "GRANT developer TO app_admin;", cql)
}

// TestRenderUse tests USE operation
func TestRenderUse(t *testing.T) {
	plan := &AIResult{
		Operation: "USE",
		Keyspace:  "production",
	}

	cql, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Equal(t, "USE production", cql)
}

// TestRenderCreateType tests CREATE TYPE (UDT)
func TestRenderCreateType(t *testing.T) {
	plan := &AIResult{
		Operation: "CREATE",
		Keyspace:  "test_ks",
		Options: map[string]any{
			"object_type": "TYPE",
			"type_name":   "address",
		},
		Schema: map[string]string{
			"street": "text",
			"city":   "text",
			"zip":    "int",
		},
	}

	cql, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Contains(t, cql, "CREATE TYPE test_ks.address")
	assert.Contains(t, cql, "street text")
	assert.Contains(t, cql, "city text")
	assert.Contains(t, cql, "zip int")
}

// TestRenderTruncate tests TRUNCATE operation
func TestRenderTruncate(t *testing.T) {
	plan := &AIResult{
		Operation: "TRUNCATE",
		Keyspace:  "test_ks",
		Table:     "events",
	}

	cql, err := RenderCQL(plan)
	assert.NoError(t, err)
	assert.Equal(t, "TRUNCATE test_ks.events", cql)
}

// TestRenderGrantPermissionGranularity tests GRANT at different resource scopes
func TestRenderGrantPermissionGranularity(t *testing.T) {
	tests := []struct {
		name string
		plan *AIResult
		want string
	}{
		{
			name: "GRANT on ALL KEYSPACES",
			plan: &AIResult{
				Operation: "GRANT",
				Options: map[string]any{
					"permission":     "SELECT",
					"role":           "app_readonly",
					"resource_scope": "ALL KEYSPACES",
				},
			},
			want: "GRANT SELECT ON ALL KEYSPACES TO app_readonly;",
		},
		{
			name: "GRANT on specific KEYSPACE",
			plan: &AIResult{
				Operation: "GRANT",
				Keyspace:  "test_ks",
				Options: map[string]any{
					"permission":     "MODIFY",
					"role":           "app_readwrite",
					"resource_scope": "KEYSPACE",
				},
			},
			want: "GRANT MODIFY ON KEYSPACE test_ks TO app_readwrite;",
		},
		{
			name: "GRANT on specific TABLE",
			plan: &AIResult{
				Operation: "GRANT",
				Keyspace:  "test_ks",
				Table:     "users",
				Options: map[string]any{
					"permission":     "SELECT",
					"role":           "app_viewer",
					"resource_scope": "TABLE",
				},
			},
			want: "GRANT SELECT ON TABLE test_ks.users TO app_viewer;",
		},
		{
			name: "GRANT on ALL ROLES",
			plan: &AIResult{
				Operation: "GRANT",
				Options: map[string]any{
					"permission":     "DESCRIBE",
					"role":           "app_admin",
					"resource_scope": "ALL ROLES",
				},
			},
			want: "GRANT DESCRIBE ON ALL ROLES TO app_admin;",
		},
		{
			name: "default to KEYSPACE scope",
			plan: &AIResult{
				Operation: "GRANT",
				Keyspace:  "prod_ks",
				Options: map[string]any{
					"permission": "ALL",
					"role":       "dba_role",
				},
			},
			want: "GRANT ALL ON KEYSPACE prod_ks TO dba_role;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCQL(tt.plan)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRenderRevokePermissionGranularity tests REVOKE at different resource scopes
func TestRenderRevokePermissionGranularity(t *testing.T) {
	tests := []struct {
		name string
		plan *AIResult
		want string
	}{
		{
			name: "REVOKE from TABLE",
			plan: &AIResult{
				Operation: "REVOKE",
				Keyspace:  "test_ks",
				Table:     "sensitive_data",
				Options: map[string]any{
					"permission":     "SELECT",
					"role":           "app_viewer",
					"resource_scope": "TABLE",
				},
			},
			want: "REVOKE SELECT ON TABLE test_ks.sensitive_data FROM app_viewer;",
		},
		{
			name: "REVOKE from ALL KEYSPACES",
			plan: &AIResult{
				Operation: "REVOKE",
				Options: map[string]any{
					"permission":     "MODIFY",
					"role":           "temp_user",
					"resource_scope": "ALL KEYSPACES",
				},
			},
			want: "REVOKE MODIFY ON ALL KEYSPACES FROM temp_user;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCQL(tt.plan)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRenderGrantInvalidPermission tests that invalid permissions are rejected
func TestRenderGrantInvalidPermission(t *testing.T) {
	tests := []struct {
		name       string
		permission string
	}{
		{"invalid permission READ", "READ"},
		{"invalid permission WRITE", "WRITE"},
		{"invalid permission ADMIN", "ADMIN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &AIResult{
				Operation: "GRANT",
				Keyspace:  "test_ks",
				Options: map[string]any{
					"permission": tt.permission,
					"role":       "test_role",
				},
			}

			_, err := RenderCQL(plan)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid permission")
		})
	}
}

// TestRenderGrantAllPermissionTypes tests all valid Cassandra permissions
func TestRenderGrantAllPermissionTypes(t *testing.T) {
	permissions := []string{"CREATE", "ALTER", "DROP", "SELECT", "MODIFY", "AUTHORIZE", "DESCRIBE", "EXECUTE", "UNMASK", "SELECT_MASKED", "ALL"}

	for _, perm := range permissions {
		t.Run("GRANT_"+perm, func(t *testing.T) {
			plan := &AIResult{
				Operation: "GRANT",
				Keyspace:  "test_ks",
				Options: map[string]any{
					"permission": perm,
					"role":       "test_role",
				},
			}

			cql, err := RenderCQL(plan)
			assert.NoError(t, err)
			assert.Contains(t, cql, "GRANT")
			assert.Contains(t, cql, strings.ToUpper(perm))
		})
	}
}

// ============================================================================
// Phase 1: Simple DML Features Tests
// ============================================================================

// TestRenderInsert_WithTTL tests INSERT with USING TTL clause
func TestRenderInsert_WithTTL(t *testing.T) {
	tests := []struct {
		name    string
		plan    *AIResult
		want    string
		wantErr bool
	}{
		{
			name: "INSERT with TTL",
			plan: &AIResult{
				Operation: "INSERT",
				Table:     "users",
				Values: map[string]any{
					"id":   1,
					"name": "Alice",
				},
				UsingTTL: 300,
			},
			want:    "INSERT INTO users (id, name) VALUES (1, 'Alice') USING TTL 300;",
			wantErr: false,
		},
		{
			name: "INSERT with TTL and keyspace",
			plan: &AIResult{
				Operation: "INSERT",
				Keyspace:  "test_ks",
				Table:     "users",
				Values: map[string]any{
					"id":    100,
					"email": "test@example.com",
				},
				UsingTTL: 600,
			},
			want:    "INSERT INTO test_ks.users (id, email) VALUES (100, 'test@example.com') USING TTL 600;",
			wantErr: false,
		},
		{
			name: "INSERT without TTL (backward compatible)",
			plan: &AIResult{
				Operation: "INSERT",
				Table:     "users",
				Values: map[string]any{
					"id":   2,
					"name": "Bob",
				},
				UsingTTL: 0,
			},
			want:    "INSERT INTO users (id, name) VALUES (2, 'Bob');",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCQL(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, got, "INSERT INTO")
				if tt.plan.UsingTTL > 0 {
					assert.Contains(t, got, "USING TTL")
				} else {
					assert.NotContains(t, got, "USING TTL")
				}
			}
		})
	}
}

// TestRenderUpdate_WithTTL tests UPDATE with USING TTL clause
func TestRenderUpdate_WithTTL(t *testing.T) {
	tests := []struct {
		name    string
		plan    *AIResult
		want    string
		wantErr bool
	}{
		{
			name: "UPDATE with TTL",
			plan: &AIResult{
				Operation: "UPDATE",
				Table:     "users",
				Values: map[string]any{
					"name": "Updated",
				},
				Where: []WhereClause{
					{Column: "id", Operator: "=", Value: 1},
				},
				UsingTTL: 600,
			},
			want:    "UPDATE users USING TTL 600 SET name = 'Updated' WHERE id = 1;",
			wantErr: false,
		},
		{
			name: "UPDATE with TTL and multiple columns",
			plan: &AIResult{
				Operation: "UPDATE",
				Table:     "users",
				Values: map[string]any{
					"name":  "New Name",
					"email": "new@example.com",
				},
				Where: []WhereClause{
					{Column: "id", Operator: "=", Value: 5},
				},
				UsingTTL: 1200,
			},
			want:    "UPDATE users USING TTL 1200 SET",
			wantErr: false,
		},
		{
			name: "UPDATE without TTL",
			plan: &AIResult{
				Operation: "UPDATE",
				Table:     "users",
				Values: map[string]any{
					"name": "Regular",
				},
				Where: []WhereClause{
					{Column: "id", Operator: "=", Value: 2},
				},
			},
			want:    "UPDATE users SET name = 'Regular' WHERE id = 2;",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCQL(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, got, tt.want)
				if tt.plan.UsingTTL > 0 {
					assert.Contains(t, got, "USING TTL")
				} else {
					assert.NotContains(t, got, "USING")
				}
			}
		})
	}
}

// TestRenderInsert_WithTimestamp tests INSERT with USING TIMESTAMP clause
func TestRenderInsert_WithTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		plan    *AIResult
		want    string
		wantErr bool
	}{
		{
			name: "INSERT with TIMESTAMP",
			plan: &AIResult{
				Operation: "INSERT",
				Table:     "users",
				Values: map[string]any{
					"id":   10,
					"name": "TimestampTest",
				},
				UsingTimestamp: 1609459200000000,
			},
			want:    "INSERT INTO users (id, name) VALUES (10, 'TimestampTest') USING TIMESTAMP 1609459200000000;",
			wantErr: false,
		},
		{
			name: "INSERT with TTL and TIMESTAMP combined",
			plan: &AIResult{
				Operation: "INSERT",
				Table:     "users",
				Values: map[string]any{
					"id":   11,
					"name": "Combined",
				},
				UsingTTL:       300,
				UsingTimestamp: 1609459200000000,
			},
			want:    "USING TTL 300 AND TIMESTAMP 1609459200000000",
			wantErr: false,
		},
		{
			name: "INSERT without TIMESTAMP",
			plan: &AIResult{
				Operation: "INSERT",
				Table:     "users",
				Values: map[string]any{
					"id":   12,
					"name": "NoTimestamp",
				},
			},
			want:    "INSERT INTO users",  // Just check it's INSERT, not exact column order
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCQL(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, got, tt.want)
			}
		})
	}
}

// TestRenderUpdate_WithTimestamp tests UPDATE with USING TIMESTAMP clause
func TestRenderUpdate_WithTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		plan    *AIResult
		want    string
		wantErr bool
	}{
		{
			name: "UPDATE with TIMESTAMP",
			plan: &AIResult{
				Operation: "UPDATE",
				Table:     "users",
				Values: map[string]any{
					"name": "Updated",
				},
				Where: []WhereClause{
					{Column: "id", Operator: "=", Value: 1},
				},
				UsingTimestamp: 1609459200000000,
			},
			want:    "UPDATE users USING TIMESTAMP 1609459200000000 SET name = 'Updated' WHERE id = 1;",
			wantErr: false,
		},
		{
			name: "UPDATE with TTL and TIMESTAMP",
			plan: &AIResult{
				Operation: "UPDATE",
				Table:     "users",
				Values: map[string]any{
					"email": "new@example.com",
				},
				Where: []WhereClause{
					{Column: "id", Operator: "=", Value: 2},
				},
				UsingTTL:       600,
				UsingTimestamp: 1609459200000000,
			},
			want:    "USING TTL 600 AND TIMESTAMP 1609459200000000",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCQL(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, got, tt.want)
			}
		})
	}
}

// TestRenderDelete_WithTimestamp tests DELETE with USING TIMESTAMP clause
func TestRenderDelete_WithTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		plan    *AIResult
		want    string
		wantErr bool
	}{
		{
			name: "DELETE with TIMESTAMP",
			plan: &AIResult{
				Operation: "DELETE",
				Table:     "users",
				Where: []WhereClause{
					{Column: "id", Operator: "=", Value: 1},
				},
				UsingTimestamp: 1609459200000000,
			},
			want:    "DELETE FROM users USING TIMESTAMP 1609459200000000 WHERE id = 1;",
			wantErr: false,
		},
		{
			name: "DELETE specific columns with TIMESTAMP",
			plan: &AIResult{
				Operation: "DELETE",
				Table:     "users",
				Columns:   []string{"email", "name"},
				Where: []WhereClause{
					{Column: "id", Operator: "=", Value: 2},
				},
				UsingTimestamp: 1609459200000000,
			},
			want:    "DELETE email, name FROM users USING TIMESTAMP 1609459200000000 WHERE id = 2;",
			wantErr: false,
		},
		{
			name: "DELETE without TIMESTAMP",
			plan: &AIResult{
				Operation: "DELETE",
				Table:     "users",
				Where: []WhereClause{
					{Column: "id", Operator: "=", Value: 3},
				},
			},
			want:    "DELETE FROM users WHERE id = 3;",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCQL(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, got, tt.want)
			}
		})
	}
}
