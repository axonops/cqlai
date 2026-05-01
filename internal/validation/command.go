package validation

import (
	"fmt"
	"strings"
)

// ValidateCommandSyntax validates that a command starts with a known CQL or meta-command keyword
func ValidateCommandSyntax(command string) error {
	upperCommand := strings.ToUpper(strings.TrimSpace(command))

	// Remove trailing semicolon for validation
	upperCommand = strings.TrimSuffix(upperCommand, ";")
	upperCommand = strings.TrimSpace(upperCommand)

	if upperCommand == "" {
		return nil
	}

	// List of valid CQL command keywords
	validCQLCommands := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE",
		"CREATE", "ALTER", "DROP", "TRUNCATE",
		"USE", "GRANT", "REVOKE",
		"BEGIN", "APPLY",
		"LIST",
	}

	// List of valid meta-commands
	validMetaCommands := []string{
		"DESCRIBE", "DESC", "CONSISTENCY", "OUTPUT",
		"PAGING", "AUTOFETCH", "TRACING", "SOURCE",
		"COPY", "SHOW", "EXPAND", "CAPTURE",
		"HELP", "SAVE",
	}

	// Check if command starts with any valid keyword
	for _, cmd := range validCQLCommands {
		if strings.HasPrefix(upperCommand, cmd+" ") || upperCommand == cmd {
			return nil
		}
	}

	for _, cmd := range validMetaCommands {
		// Check for word boundary: exact match OR followed by space
		if upperCommand == cmd || strings.HasPrefix(upperCommand, cmd+" ") {
			return nil
		}
	}

	// If we get here, it's not a recognized command
	firstWord := strings.Fields(upperCommand)[0]
	return fmt.Errorf("invalid command: '%s' is not a recognized CQL or meta-command", firstWord)
}

// IsDangerousCommand checks if a command requires confirmation
func IsDangerousCommand(command string) bool {
	upperCommand := strings.ToUpper(strings.TrimSpace(command))

	// List of dangerous command prefixes
	dangerousCommands := []string{
		"ALTER",
		"DROP",
		"DELETE",
		"REVOKE",
		"TRUNCATE",
	}

	// Check if the command starts with any dangerous keyword (with word boundary)
	for _, dangerous := range dangerousCommands {
		// Check for word boundary: exact match OR followed by space
		if upperCommand == dangerous || strings.HasPrefix(upperCommand, dangerous+" ") {
			return true
		}
	}

	return false
}
