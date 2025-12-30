package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClassifyOperation_DQL tests DQL operation classification
func TestClassifyOperation_DQL(t *testing.T) {
	tests := []struct {
		command  string
		category OperationCategory
		risk     string
	}{
		// SELECT
		{"SELECT * FROM users", CategoryDQL, "SAFE"},
		{"select email from users where id=?", CategoryDQL, "SAFE"},

		// LIST operations
		{"LIST ROLES", CategoryDQL, "SAFE"},
		{"list roles", CategoryDQL, "SAFE"},
		{"LIST USERS", CategoryDQL, "SAFE"},
		{"LIST PERMISSIONS", CategoryDQL, "SAFE"},
		{"LIST ALL PERMISSIONS OF role", CategoryDQL, "SAFE"},

		// DESCRIBE operations
		{"DESCRIBE KEYSPACES", CategoryDQL, "SAFE"},
		{"DESC KEYSPACES", CategoryDQL, "SAFE"},
		{"DESCRIBE TABLES", CategoryDQL, "SAFE"},
		{"DESCRIBE TABLE users", CategoryDQL, "SAFE"},
		{"DESC TABLE users", CategoryDQL, "SAFE"},
		{"DESCRIBE TYPE address", CategoryDQL, "SAFE"},
		{"DESCRIBE TYPES", CategoryDQL, "SAFE"},
		{"DESCRIBE CLUSTER", CategoryDQL, "SAFE"},

		// SHOW operations
		{"SHOW VERSION", CategoryDQL, "SAFE"},
		{"SHOW HOST", CategoryDQL, "SAFE"},
		{"SHOW SESSION", CategoryDQL, "SAFE"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := ClassifyOperation(tt.command)
			assert.Equal(t, tt.category, result.Category,
				"Category mismatch for: %s", tt.command)
			assert.Equal(t, tt.risk, result.RiskLevel,
				"Risk level mismatch for: %s", tt.command)
		})
	}
}

// TestClassifyOperation_SESSION tests SESSION operation classification
func TestClassifyOperation_SESSION(t *testing.T) {
	tests := []struct {
		command string
	}{
		{"CONSISTENCY QUORUM"},
		{"CONSISTENCY LOCAL_ONE"},
		{"PAGING 100"},
		{"PAGING OFF"},
		{"TRACING ON"},
		{"TRACING OFF"},
		{"AUTOFETCH ON"},
		{"AUTOFETCH OFF"},
		{"EXPAND ON"},
		{"EXPAND OFF"},
		{"OUTPUT JSON"},
		{"OUTPUT TABLE"},
		{"CAPTURE ON"},
		{"CAPTURE OFF"},
		{"CAPTURE PARQUET 'file.parquet'"},
		{"CAPTURE JSON 'file.json'"},
		{"SAVE 'results.csv'"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := ClassifyOperation(tt.command)
			assert.Equal(t, CategorySESSION, result.Category,
				"Should be SESSION category for: %s", tt.command)
			assert.Equal(t, "SAFE", result.RiskLevel,
				"SESSION operations should be SAFE: %s", tt.command)
		})
	}
}

// TestClassifyOperation_DML tests DML operation classification
func TestClassifyOperation_DML(t *testing.T) {
	tests := []struct {
		command  string
		risk     string
	}{
		{"INSERT INTO users VALUES (...)", "LOW"},
		{"insert into users (id, name) values (?, ?)", "LOW"},

		{"UPDATE users SET name=? WHERE id=?", "MEDIUM"},
		{"update users set count = count + 1", "MEDIUM"},

		{"DELETE FROM users WHERE id=?", "HIGH"},
		{"delete from events where timestamp < ?", "HIGH"},

		{"BEGIN BATCH", "MEDIUM"},
		{"APPLY BATCH", "MEDIUM"},
		{"BEGIN UNLOGGED BATCH", "MEDIUM"},
		{"BEGIN COUNTER BATCH", "MEDIUM"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := ClassifyOperation(tt.command)
			assert.Equal(t, CategoryDML, result.Category,
				"Should be DML category for: %s", tt.command)
			assert.Equal(t, tt.risk, result.RiskLevel,
				"Risk level mismatch for: %s", tt.command)
		})
	}
}

// TestClassifyOperation_DDL tests DDL operation classification
func TestClassifyOperation_DDL(t *testing.T) {
	tests := []struct {
		command  string
		risk     string
	}{
		// CREATE operations - MEDIUM risk
		{"CREATE KEYSPACE ks WITH REPLICATION = ...", "MEDIUM"},
		{"CREATE TABLE users (...)", "MEDIUM"},
		{"CREATE COLUMNFAMILY users (...)", "MEDIUM"},
		{"CREATE INDEX ON users(email)", "MEDIUM"},
		{"CREATE CUSTOM INDEX ON users(name) USING 'SAI'", "MEDIUM"},
		{"CREATE MATERIALIZED VIEW mv AS SELECT ...", "MEDIUM"},
		{"CREATE TYPE address (...)", "MEDIUM"},
		{"CREATE FUNCTION func(...)", "MEDIUM"},
		{"CREATE OR REPLACE FUNCTION func(...)", "MEDIUM"},
		{"CREATE AGGREGATE avg(...)", "MEDIUM"},
		{"CREATE TRIGGER trig ON table", "MEDIUM"},

		// ALTER operations - HIGH risk
		{"ALTER KEYSPACE ks WITH REPLICATION = ...", "HIGH"},
		{"ALTER TABLE users ADD column text", "HIGH"},
		{"ALTER COLUMNFAMILY users ADD column", "HIGH"},
		{"ALTER MATERIALIZED VIEW mv WITH ...", "HIGH"},
		{"ALTER TYPE address ADD field text", "HIGH"},

		// DROP operations - HIGH or CRITICAL risk
		{"DROP KEYSPACE ks", "CRITICAL"},
		{"DROP TABLE users", "CRITICAL"},
		{"DROP COLUMNFAMILY users", "CRITICAL"},
		{"DROP INDEX idx_name", "HIGH"},
		{"DROP MATERIALIZED VIEW mv", "HIGH"},
		{"DROP TYPE address", "HIGH"},
		{"DROP FUNCTION func", "HIGH"},
		{"DROP AGGREGATE avg", "HIGH"},
		{"DROP TRIGGER trig", "HIGH"},

		// TRUNCATE - CRITICAL risk
		{"TRUNCATE users", "CRITICAL"},
		{"TRUNCATE TABLE users", "CRITICAL"},

		// USE - SAFE
		{"USE keyspace_name", "SAFE"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := ClassifyOperation(tt.command)
			assert.Equal(t, CategoryDDL, result.Category,
				"Should be DDL category for: %s", tt.command)
			assert.Equal(t, tt.risk, result.RiskLevel,
				"Risk level mismatch for: %s", tt.command)
		})
	}
}

// TestClassifyOperation_DCL tests DCL operation classification
func TestClassifyOperation_DCL(t *testing.T) {
	tests := []struct {
		command  string
		risk     string
	}{
		// Role operations
		{"CREATE ROLE app WITH PASSWORD='...'", "HIGH"},
		{"ALTER ROLE app WITH PASSWORD='...'", "HIGH"},
		{"DROP ROLE app", "CRITICAL"},
		{"GRANT ROLE admin TO user", "HIGH"},
		{"REVOKE ROLE admin FROM user", "HIGH"},

		// User operations (legacy)
		{"CREATE USER admin WITH PASSWORD='...'", "HIGH"},
		{"ALTER USER admin WITH PASSWORD='...'", "HIGH"},
		{"DROP USER admin", "CRITICAL"},

		// Identity operations
		{"ADD IDENTITY 'user@REALM' TO role", "HIGH"},
		{"DROP IDENTITY 'user@REALM' FROM role", "HIGH"},

		// Permission operations
		{"GRANT SELECT ON ks.table TO role", "HIGH"},
		{"REVOKE ALL ON ks.table FROM role", "HIGH"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := ClassifyOperation(tt.command)
			assert.Equal(t, CategoryDCL, result.Category,
				"Should be DCL category for: %s", tt.command)
			assert.Equal(t, tt.risk, result.RiskLevel,
				"Risk level mismatch for: %s", tt.command)
		})
	}
}

// TestClassifyOperation_FILE tests FILE operation classification
func TestClassifyOperation_FILE(t *testing.T) {
	tests := []struct {
		command string
		risk    string
	}{
		{"COPY users TO 'data.csv'", "LOW"},
		{"COPY users FROM 'data.csv'", "MEDIUM"},
		{"SOURCE 'script.cql'", "VARIABLE"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := ClassifyOperation(tt.command)
			assert.Equal(t, CategoryFILE, result.Category,
				"Should be FILE category for: %s", tt.command)
			assert.Equal(t, tt.risk, result.RiskLevel,
				"Risk level mismatch for: %s", tt.command)
		})
	}
}

// TestClassifyOperation_Unknown tests unknown operation handling
func TestClassifyOperation_Unknown(t *testing.T) {
	tests := []string{
		"FOOBAR",
		"INVALID COMMAND",
		"RANDOM STUFF",
		"",
	}

	for _, cmd := range tests {
		t.Run(cmd, func(t *testing.T) {
			result := ClassifyOperation(cmd)
			assert.Equal(t, CategoryUNKNOWN, result.Category,
				"Unknown command should be CategoryUNKNOWN: %s", cmd)
			assert.Equal(t, "CRITICAL", result.RiskLevel,
				"Unknown operations should be CRITICAL risk: %s", cmd)
		})
	}
}

// TestGetCategoryDescription tests category descriptions
func TestGetCategoryDescription(t *testing.T) {
	tests := []struct {
		category OperationCategory
		contains string
	}{
		{CategoryDQL, "read-only"},
		{CategorySESSION, "always allowed"},
		{CategoryDML, "modifies data"},
		{CategoryDDL, "modifies schema"},
		{CategoryDCL, "security"},
		{CategoryFILE, "import/export"},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			desc := GetCategoryDescription(tt.category)
			assert.Contains(t, desc, tt.contains,
				"Description for %s should contain '%s'", tt.category, tt.contains)
		})
	}
}

// TestGetCategoryOperationCount tests operation counts
func TestGetCategoryOperationCount(t *testing.T) {
	tests := []struct {
		category OperationCategory
		expected int
	}{
		{CategoryDQL, 14},
		{CategorySESSION, 8},
		{CategoryDML, 8},
		{CategoryDDL, 28},
		{CategoryDCL, 13},
		{CategoryFILE, 3},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			count := GetCategoryOperationCount(tt.category)
			assert.Equal(t, tt.expected, count,
				"Operation count mismatch for %s", tt.category)
		})
	}
}
