package ai

import (
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
