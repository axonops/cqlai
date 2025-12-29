package ai

import (
	"fmt"
	"strings"
)

// Severity levels for query classification
const (
	SeverityCritical = "CRITICAL"
	SeverityHigh     = "HIGH"
	SeverityMedium   = "MEDIUM"
	SeverityLow      = "LOW"
	SeveritySafe     = "SAFE"
)

// ClassifyQuery analyzes a CQL query and determines if it's dangerous.
// Returns a QueryClassification with severity level and details.
func ClassifyQuery(query string) QueryClassification {
	upperQuery := strings.ToUpper(strings.TrimSpace(query))

	// Check CRITICAL operations
	criticalPatterns := []string{
		"CREATE ROLE",
		"CREATE TABLE",
		"CREATE KEYSPACE",
		"CREATE INDEX",
		"DROP TABLE",
		"DROP KEYSPACE",
		"DROP INDEX",
		"DROP ROLE",
		"TRUNCATE TABLE",
		"TRUNCATE ",
		"GRANT SUPERUSER",
		"GRANT ALL PERMISSIONS",
		"GRANT ALL",
		"REVOKE SUPERUSER",
		"REVOKE ALL PERMISSIONS",
		"REVOKE ALL",
	}

	for _, pattern := range criticalPatterns {
		if strings.Contains(upperQuery, pattern) {
			return QueryClassification{
				IsDangerous: true,
				Severity:    SeverityCritical,
				Operation:   extractOperation(query),
				Description: fmt.Sprintf("%s operation - will modify cluster structure or permissions", extractOperation(query)),
			}
		}
	}

	// Check HIGH operations
	highPatterns := []string{
		"DELETE FROM",
		"DELETE ",
		"UPDATE ",
		"ALTER TABLE",
		"ALTER KEYSPACE",
		"ALTER ROLE",
	}

	for _, pattern := range highPatterns {
		if strings.Contains(upperQuery, pattern) {
			return QueryClassification{
				IsDangerous: true,
				Severity:    SeverityHigh,
				Operation:   extractOperation(query),
				Description: fmt.Sprintf("%s operation - will modify data", extractOperation(query)),
			}
		}
	}

	// Everything else is SAFE (SELECT, LIST, DESCRIBE, etc.)
	return QueryClassification{
		IsDangerous: false,
		Severity:    SeveritySafe,
		Operation:   extractOperation(query),
		Description: "Read-only operation",
	}
}

// extractOperation extracts the operation type from a CQL query
func extractOperation(query string) string {
	upperQuery := strings.ToUpper(strings.TrimSpace(query))

	operations := []string{
		"CREATE",
		"DROP",
		"DELETE",
		"UPDATE",
		"ALTER",
		"TRUNCATE",
		"GRANT",
		"REVOKE",
		"SELECT",
		"INSERT",
		"DESCRIBE",
		"LIST",
		"SHOW",
		"USE",
	}

	for _, op := range operations {
		if strings.HasPrefix(upperQuery, op) {
			return op
		}
	}

	return "OTHER"
}

// IsDangerousQuery returns true if the query requires confirmation
func IsDangerousQuery(query string) bool {
	classification := ClassifyQuery(query)
	return classification.IsDangerous
}

// GetQuerySeverity returns the severity level of a query
func GetQuerySeverity(query string) string {
	classification := ClassifyQuery(query)
	return classification.Severity
}
