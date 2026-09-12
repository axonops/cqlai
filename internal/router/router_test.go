package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWhatCountsAsASchemaChange, which the schema cache and the SCHEMA view
// both ask about: one list rather than a copy in each.
func TestWhatCountsAsASchemaChange(t *testing.T) {
	for _, command := range []string{
		"CREATE TABLE users (id uuid PRIMARY KEY)",
		"create table users (id uuid PRIMARY KEY)",
		"  ALTER TABLE users ADD name text",
		"DROP TABLE users",
		"CREATE KEYSPACE ks WITH replication = {'class': 'SimpleStrategy'}",
		"DROP MATERIALIZED VIEW users_by_name",
		"CREATE INDEX ON users (name)",
	} {
		assert.True(t, ChangesSchema(command), "%q changes the schema", command)
	}

	for _, command := range []string{
		"SELECT * FROM users",
		"INSERT INTO users (id) VALUES (uuid())",
		"UPDATE users SET name = 'a' WHERE id = 1",
		"DELETE FROM users WHERE id = 1",
		"DESCRIBE TABLES",
		"OUTPUT JSON",
		"",
	} {
		assert.False(t, ChangesSchema(command), "%q does not", command)
	}
}
