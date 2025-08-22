package validation

import "strings"

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

	// Check if the command starts with any dangerous prefix
	for _, dangerous := range dangerousCommands {
		if strings.HasPrefix(upperCommand, dangerous) {
			return true
		}
	}

	// Special case: ALTER without TABLE/KEYSPACE/MATERIALIZED VIEW is still dangerous
	if strings.HasPrefix(upperCommand, "ALTER ") {
		return true
	}

	// Special case: DROP without specific type is still dangerous
	if strings.HasPrefix(upperCommand, "DROP ") {
		return true
	}

	// Special case: REVOKE without specific type is still dangerous
	if strings.HasPrefix(upperCommand, "REVOKE ") {
		return true
	}

	return false
}
