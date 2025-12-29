package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClassifyQuery tests query classification for all severity levels
func TestClassifyQuery(t *testing.T) {
	tests := []struct {
		name            string
		query           string
		wantDangerous   bool
		wantSeverity    string
		wantOperation   string
	}{
		// CRITICAL operations
		{
			name:          "CREATE ROLE with password",
			query:         "CREATE ROLE app_service WITH PASSWORD = 'secret123'",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "CREATE",
		},
		{
			name:          "CREATE TABLE",
			query:         "CREATE TABLE users (id uuid PRIMARY KEY, name text)",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "CREATE",
		},
		{
			name:          "CREATE KEYSPACE",
			query:         "CREATE KEYSPACE production WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 3}",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "CREATE",
		},
		{
			name:          "CREATE INDEX",
			query:         "CREATE INDEX ON users (email)",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "CREATE",
		},
		{
			name:          "DROP TABLE",
			query:         "DROP TABLE users",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "DROP",
		},
		{
			name:          "DROP KEYSPACE",
			query:         "DROP KEYSPACE test_ks",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "DROP",
		},
		{
			name:          "DROP INDEX",
			query:         "DROP INDEX users_email_idx",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "DROP",
		},
		{
			name:          "TRUNCATE TABLE",
			query:         "TRUNCATE TABLE events",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "TRUNCATE",
		},
		{
			name:          "TRUNCATE without TABLE keyword",
			query:         "TRUNCATE events",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "TRUNCATE",
		},
		{
			name:          "GRANT SUPERUSER",
			query:         "GRANT SUPERUSER TO admin",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "GRANT",
		},
		{
			name:          "GRANT ALL PERMISSIONS",
			query:         "GRANT ALL PERMISSIONS ON ALL KEYSPACES TO app_service",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "GRANT",
		},
		{
			name:          "GRANT ALL",
			query:         "GRANT ALL ON KEYSPACE production TO app_user",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "GRANT",
		},
		{
			name:          "REVOKE SUPERUSER",
			query:         "REVOKE SUPERUSER FROM admin",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "REVOKE",
		},

		// HIGH operations
		{
			name:          "DELETE FROM",
			query:         "DELETE FROM events WHERE timestamp < '2025-01-01'",
			wantDangerous: true,
			wantSeverity:  SeverityHigh,
			wantOperation: "DELETE",
		},
		{
			name:          "DELETE without FROM",
			query:         "DELETE users WHERE id = 123",
			wantDangerous: true,
			wantSeverity:  SeverityHigh,
			wantOperation: "DELETE",
		},
		{
			name:          "UPDATE",
			query:         "UPDATE users SET email = 'new@example.com' WHERE id = 123",
			wantDangerous: true,
			wantSeverity:  SeverityHigh,
			wantOperation: "UPDATE",
		},
		{
			name:          "ALTER TABLE",
			query:         "ALTER TABLE users ADD email_verified boolean",
			wantDangerous: true,
			wantSeverity:  SeverityHigh,
			wantOperation: "ALTER",
		},
		{
			name:          "ALTER KEYSPACE",
			query:         "ALTER KEYSPACE production WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': 3}",
			wantDangerous: true,
			wantSeverity:  SeverityHigh,
			wantOperation: "ALTER",
		},
		{
			name:          "ALTER ROLE",
			query:         "ALTER ROLE app_service WITH PASSWORD = 'newpass'",
			wantDangerous: true,
			wantSeverity:  SeverityHigh,
			wantOperation: "ALTER",
		},

		// SAFE operations
		{
			name:          "SELECT query",
			query:         "SELECT * FROM users LIMIT 10",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
			wantOperation: "SELECT",
		},
		{
			name:          "SELECT with WHERE",
			query:         "SELECT email FROM users WHERE id = 123",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
			wantOperation: "SELECT",
		},
		{
			name:          "DESCRIBE TABLE",
			query:         "DESCRIBE TABLE users",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
			wantOperation: "DESCRIBE",
		},
		{
			name:          "LIST ROLES",
			query:         "LIST ROLES",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
			wantOperation: "LIST",
		},
		{
			name:          "INSERT single row",
			query:         "INSERT INTO users (id, name) VALUES (uuid(), 'John')",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
			wantOperation: "INSERT",
		},
		{
			name:          "GRANT specific permission (safe)",
			query:         "GRANT SELECT ON KEYSPACE production TO readonly_user",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
			wantOperation: "GRANT",
		},
		{
			name:          "USE keyspace",
			query:         "USE production",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
			wantOperation: "USE",
		},
		{
			name:          "SHOW VERSION",
			query:         "SHOW VERSION",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
			wantOperation: "SHOW",
		},

		// Case insensitive tests
		{
			name:          "lowercase delete",
			query:         "delete from events where id = 1",
			wantDangerous: true,
			wantSeverity:  SeverityHigh,
			wantOperation: "DELETE",
		},
		{
			name:          "mixed case create table",
			query:         "Create Table users (id uuid primary key)",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
			wantOperation: "CREATE",
		},

		// Whitespace handling
		{
			name:          "query with leading whitespace",
			query:         "   DELETE FROM events",
			wantDangerous: true,
			wantSeverity:  SeverityHigh,
			wantOperation: "DELETE",
		},
		{
			name:          "query with trailing whitespace",
			query:         "SELECT * FROM users   ",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
			wantOperation: "SELECT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyQuery(tt.query)

			assert.Equal(t, tt.wantDangerous, got.IsDangerous,
				"IsDangerous mismatch for query: %s", tt.query)
			assert.Equal(t, tt.wantSeverity, got.Severity,
				"Severity mismatch for query: %s", tt.query)
			assert.Equal(t, tt.wantOperation, got.Operation,
				"Operation mismatch for query: %s", tt.query)
		})
	}
}

// TestIsDangerousQuery tests the simple boolean check
func TestIsDangerousQuery(t *testing.T) {
	tests := []struct {
		query string
		want  bool
	}{
		{"SELECT * FROM users", false},
		{"DELETE FROM users WHERE id = 1", true},
		{"CREATE TABLE test (id uuid)", true},
		{"INSERT INTO users VALUES (uuid(), 'test')", false},
		{"TRUNCATE events", true},
		{"DESCRIBE TABLE users", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := IsDangerousQuery(tt.query)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGetQuerySeverity tests severity extraction
func TestGetQuerySeverity(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"SELECT * FROM users", SeveritySafe},
		{"DELETE FROM users", SeverityHigh},
		{"CREATE TABLE test (id uuid)", SeverityCritical},
		{"UPDATE users SET name = 'x'", SeverityHigh},
		{"DROP TABLE users", SeverityCritical},
		{"GRANT SELECT ON KEYSPACE prod TO user", SeveritySafe},
		{"GRANT ALL ON KEYSPACE prod TO user", SeverityCritical},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := GetQuerySeverity(tt.query)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestExtractOperation tests operation extraction from queries
func TestExtractOperation(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"SELECT * FROM users", "SELECT"},
		{"select * from users", "SELECT"},
		{"INSERT INTO users VALUES (uuid(), 'test')", "INSERT"},
		{"UPDATE users SET name = 'x'", "UPDATE"},
		{"DELETE FROM users WHERE id = 1", "DELETE"},
		{"CREATE TABLE users (id uuid)", "CREATE"},
		{"DROP TABLE users", "DROP"},
		{"ALTER TABLE users ADD email text", "ALTER"},
		{"TRUNCATE TABLE events", "TRUNCATE"},
		{"GRANT SELECT ON KEYSPACE prod TO user", "GRANT"},
		{"REVOKE SELECT ON KEYSPACE prod FROM user", "REVOKE"},
		{"DESCRIBE TABLE users", "DESCRIBE"},
		{"LIST ROLES", "LIST"},
		{"SHOW VERSION", "SHOW"},
		{"USE production", "USE"},
		{"   SELECT * FROM users", "SELECT"}, // leading whitespace
		{"INVALID QUERY", "OTHER"},
		{"", "OTHER"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := extractOperation(tt.query)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestClassifyQuery_EdgeCases tests edge cases and tricky patterns
func TestClassifyQuery_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		wantDangerous bool
		wantSeverity  string
	}{
		{
			name:          "SELECT with DELETE in column name (safe)",
			query:         "SELECT delete_flag FROM users",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
		},
		{
			name:          "Comment with dangerous keyword (still detected)",
			query:         "-- This will delete\nDELETE FROM users",
			wantDangerous: true,
			wantSeverity:  SeverityHigh,
		},
		{
			name:          "GRANT specific permission (safe)",
			query:         "GRANT SELECT ON TABLE users TO readonly",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
		},
		{
			name:          "GRANT ALL (dangerous)",
			query:         "GRANT ALL ON TABLE users TO admin",
			wantDangerous: true,
			wantSeverity:  SeverityCritical,
		},
		{
			name:          "Empty query",
			query:         "",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
		},
		{
			name:          "Whitespace only",
			query:         "   \n\t   ",
			wantDangerous: false,
			wantSeverity:  SeveritySafe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyQuery(tt.query)

			assert.Equal(t, tt.wantDangerous, got.IsDangerous,
				"IsDangerous mismatch for: %s", tt.query)
			assert.Equal(t, tt.wantSeverity, got.Severity,
				"Severity mismatch for: %s", tt.query)
		})
	}
}

// TestClassifyQuery_AllCriticalPatterns ensures all CRITICAL patterns are detected
func TestClassifyQuery_AllCriticalPatterns(t *testing.T) {
	criticalQueries := []string{
		"CREATE ROLE test",
		"CREATE TABLE test (id uuid)",
		"CREATE KEYSPACE test",
		"CREATE INDEX test_idx",
		"DROP TABLE users",
		"DROP KEYSPACE test",
		"DROP INDEX test_idx",
		"DROP ROLE test_role",
		"TRUNCATE TABLE events",
		"TRUNCATE events",
		"GRANT SUPERUSER TO admin",
		"GRANT ALL PERMISSIONS ON ALL KEYSPACES TO admin",
		"GRANT ALL ON KEYSPACE prod TO user",
		"REVOKE SUPERUSER FROM admin",
		"REVOKE ALL PERMISSIONS FROM user",
		"REVOKE ALL FROM user",
	}

	for _, query := range criticalQueries {
		t.Run(query, func(t *testing.T) {
			classification := ClassifyQuery(query)
			assert.True(t, classification.IsDangerous,
				"Query should be dangerous: %s", query)
			assert.Equal(t, SeverityCritical, classification.Severity,
				"Query should be CRITICAL: %s", query)
		})
	}
}

// TestClassifyQuery_AllHighPatterns ensures all HIGH patterns are detected
func TestClassifyQuery_AllHighPatterns(t *testing.T) {
	highQueries := []string{
		"DELETE FROM users WHERE id = 1",
		"DELETE users WHERE id = 1",
		"UPDATE users SET name = 'x' WHERE id = 1",
		"UPDATE users SET name = 'x'",
		"ALTER TABLE users ADD email text",
		"ALTER KEYSPACE prod WITH replication = {'class': 'SimpleStrategy'}",
		"ALTER ROLE admin WITH PASSWORD = 'new'",
	}

	for _, query := range highQueries {
		t.Run(query, func(t *testing.T) {
			classification := ClassifyQuery(query)
			assert.True(t, classification.IsDangerous,
				"Query should be dangerous: %s", query)
			assert.Equal(t, SeverityHigh, classification.Severity,
				"Query should be HIGH: %s", query)
		})
	}
}

// TestClassifyQuery_AllSafePatterns ensures safe queries are not flagged
func TestClassifyQuery_AllSafePatterns(t *testing.T) {
	safeQueries := []string{
		"SELECT * FROM users",
		"SELECT email FROM users WHERE id = 123",
		"SELECT COUNT(*) FROM events",
		"INSERT INTO users (id, name) VALUES (uuid(), 'test')",
		"INSERT INTO events VALUES (now(), 'event')",
		"DESCRIBE TABLE users",
		"DESCRIBE KEYSPACE production",
		"LIST ROLES",
		"LIST PERMISSIONS OF app_user",
		"LIST ALL PERMISSIONS",
		"SHOW VERSION",
		"USE production",
		"GRANT SELECT ON KEYSPACE production TO readonly",
		"GRANT SELECT ON TABLE users TO readonly",
		"GRANT INSERT ON KEYSPACE production TO app_user",
		"REVOKE SELECT ON KEYSPACE prod FROM old_user",
	}

	for _, query := range safeQueries {
		t.Run(query, func(t *testing.T) {
			classification := ClassifyQuery(query)
			assert.False(t, classification.IsDangerous,
				"Query should be safe: %s", query)
			assert.Equal(t, SeveritySafe, classification.Severity,
				"Query should be SAFE: %s", query)
		})
	}
}
