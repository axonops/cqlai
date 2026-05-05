package validation

import (
	"testing"
)

func TestValidateCommandSyntax(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		// Valid CQL commands
		{"SELECT statement", "SELECT * FROM users", false},
		{"SELECT without table", "SELECT", false},
		{"INSERT statement", "INSERT INTO users (id) VALUES (1)", false},
		{"UPDATE statement", "UPDATE users SET name = 'foo'", false},
		{"DELETE statement", "DELETE FROM users WHERE id = 1", false},
		{"CREATE TABLE", "CREATE TABLE foo (id int PRIMARY KEY)", false},
		{"ALTER TABLE", "ALTER TABLE foo ADD col text", false},
		{"DROP TABLE", "DROP TABLE foo", false},
		{"TRUNCATE", "TRUNCATE users", false},
		{"USE keyspace", "USE mykeyspace", false},
		{"GRANT", "GRANT SELECT ON foo TO user1", false},
		{"REVOKE", "REVOKE SELECT ON foo FROM user1", false},
		{"BEGIN BATCH", "BEGIN BATCH", false},
		{"LIST USERS", "LIST USERS", false},

		// Valid meta-commands
		{"DESCRIBE TABLE", "DESCRIBE TABLE users", false},
		{"DESCRIBE", "DESCRIBE", false},
		{"DESC", "DESC users", false},
		{"CONSISTENCY", "CONSISTENCY LOCAL_ONE", false},
		{"OUTPUT", "OUTPUT JSON", false},
		{"PAGING", "PAGING 100", false},
		{"AUTOFETCH", "AUTOFETCH ON", false},
		{"TRACING", "TRACING ON", false},
		{"SOURCE file", "SOURCE /path/to/file.cql", false},
		{"COPY TO", "COPY users TO '/tmp/file.csv'", false},
		{"SHOW VERSION", "SHOW VERSION", false},
		{"EXPAND", "EXPAND ON", false},
		{"CAPTURE", "CAPTURE '/tmp/output.txt'", false},
		{"HELP", "HELP", false},
		{"SAVE", "SAVE /path/to/query.cql", false},

		// With trailing semicolon
		{"SELECT with semicolon", "SELECT * FROM users;", false},
		{"DESCRIBE with semicolon", "DESCRIBE TABLE users;", false},

		// Invalid commands - should not match due to word boundary check
		{"DESCRIBING - invalid prefix", "DESCRIBING something", true},
		{"DESCRIPTOR - invalid prefix", "DESCRIPTOR foo", true},
		{"SELECTS - invalid prefix", "SELECTS something", true},
		{"INSERTION - invalid prefix", "INSERTION data", true},

		// Invalid command
		{"Random word", "FOOBAR something", true},
		{"Empty after trim", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommandSyntax(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommandSyntax(%q) error = %v, wantErr %v", tt.command, err, tt.wantErr)
			}
		})
	}
}

func TestIsDangerousCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		// Dangerous commands
		{"ALTER TABLE", "ALTER TABLE foo ADD col text", true},
		{"ALTER alone", "ALTER", true},
		{"DROP TABLE", "DROP TABLE foo", true},
		{"DROP KEYSPACE", "DROP KEYSPACE mykeyspace", true},
		{"DROP alone", "DROP", true},
		{"DELETE FROM", "DELETE FROM users WHERE id = 1", true},
		{"DELETE alone", "DELETE", true},
		{"REVOKE SELECT", "REVOKE SELECT ON foo FROM user1", true},
		{"REVOKE alone", "REVOKE", true},
		{"TRUNCATE TABLE", "TRUNCATE users", true},
		{"TRUNCATE alone", "TRUNCATE", true},

		// Safe commands
		{"SELECT", "SELECT * FROM users", false},
		{"INSERT", "INSERT INTO users (id) VALUES (1)", false},
		{"UPDATE", "UPDATE users SET name = 'foo'", false},
		{"CREATE TABLE", "CREATE TABLE foo (id int PRIMARY KEY)", false},
		{"DESCRIBE", "DESCRIBE TABLE users", false},
		{"USE", "USE mykeyspace", false},
		{"GRANT", "GRANT SELECT ON foo TO user1", false},

		// Word boundary tests - these should NOT be dangerous
		{"ALTERING - not a command", "ALTERING something", false},
		{"DROPPING - not a command", "DROPPING balls", false},
		{"DELETING - not a command", "DELETING files", false},
		{"TRUNCATING - not a command", "TRUNCATING data", false},
		{"REVOKING - not a command", "REVOKING access", false},

		// Case insensitivity
		{"lowercase alter", "alter table foo", true},
		{"mixed case DROP", "Drop Table foo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDangerousCommand(tt.command)
			if got != tt.want {
				t.Errorf("IsDangerousCommand(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}
