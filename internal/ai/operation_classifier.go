package ai

import (
	"strings"
)

// OperationCategory represents a category of CQL/CQLAI operations
type OperationCategory string

const (
	CategoryDQL     OperationCategory = "dql"     // Data Query Language (read-only)
	CategorySESSION OperationCategory = "session" // Session configuration (always safe)
	CategoryDML     OperationCategory = "dml"     // Data Manipulation Language
	CategoryDDL     OperationCategory = "ddl"     // Data Definition Language
	CategoryDCL     OperationCategory = "dcl"     // Data Control Language
	CategoryFILE    OperationCategory = "file"    // File operations (import/export)
	CategoryUNKNOWN OperationCategory = "unknown" // Unknown operation - requires confirmation or error
)

// OperationInfo contains details about a classified operation
type OperationInfo struct {
	Category    OperationCategory
	Operation   string
	Description string
	RiskLevel   string // SAFE, LOW, MEDIUM, HIGH, CRITICAL
}

// ClassifyOperation determines the category of a CQL or CQLAI operation
func ClassifyOperation(command string) OperationInfo {
	// Normalize command
	cmd := strings.ToUpper(strings.TrimSpace(command))
	parts := strings.Fields(cmd)

	if len(parts) == 0 {
		return OperationInfo{Category: CategoryUNKNOWN, Operation: "", RiskLevel: "CRITICAL", Description: "Empty command"}
	}

	operation := parts[0]

	// Special handling for COPY - need to check TO/FROM
	if operation == "COPY" && len(parts) >= 3 {
		// COPY table TO/FROM file
		direction := parts[2] // Might be TO or FROM
		if direction == "TO" {
			return OperationInfo{
				Category:    CategoryFILE,
				Operation:   "COPY TO",
				Description: "Export data to file",
				RiskLevel:   "LOW",
			}
		} else if direction == "FROM" {
			return OperationInfo{
				Category:    CategoryFILE,
				Operation:   "COPY FROM",
				Description: "Import data from file",
				RiskLevel:   "MEDIUM",
			}
		}
	}

	// Handle multi-word operations
	if len(parts) >= 2 {
		twoWord := parts[0] + " " + parts[1]
		if category, info := classifyTwoWordOperation(twoWord); category != CategoryUNKNOWN {
			return OperationInfo{
				Category:    category,
				Operation:   twoWord,
				Description: info.Description,
				RiskLevel:   info.RiskLevel,
			}
		}
	}

	// Classify single-word operations
	return classifySingleWordOperation(operation)
}

// classifyTwoWordOperation handles multi-word operations
func classifyTwoWordOperation(twoWord string) (OperationCategory, OperationInfo) {
	switch twoWord {
	// DQL - Read-only list operations
	case "LIST ROLES":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "LIST ROLES", Description: "List all roles", RiskLevel: "SAFE"}
	case "LIST USERS":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "LIST USERS", Description: "List all users", RiskLevel: "SAFE"}
	case "LIST PERMISSIONS":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "LIST PERMISSIONS", Description: "List permissions", RiskLevel: "SAFE"}

	// DQL - DESCRIBE variants
	case "DESCRIBE KEYSPACES", "DESC KEYSPACES":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "DESCRIBE KEYSPACES", Description: "List keyspaces", RiskLevel: "SAFE"}
	case "DESCRIBE TABLES", "DESC TABLES":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "DESCRIBE TABLES", Description: "List tables", RiskLevel: "SAFE"}
	case "DESCRIBE TABLE", "DESC TABLE":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "DESCRIBE TABLE", Description: "Show table schema", RiskLevel: "SAFE"}
	case "DESCRIBE TYPE", "DESC TYPE":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "DESCRIBE TYPE", Description: "Show UDT schema", RiskLevel: "SAFE"}
	case "DESCRIBE TYPES", "DESC TYPES":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "DESCRIBE TYPES", Description: "List UDTs", RiskLevel: "SAFE"}
	case "DESCRIBE CLUSTER", "DESC CLUSTER":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "DESCRIBE CLUSTER", Description: "Show cluster info", RiskLevel: "SAFE"}

	// DQL - SHOW variants
	case "SHOW VERSION":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "SHOW VERSION", Description: "Show Cassandra version", RiskLevel: "SAFE"}
	case "SHOW HOST":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "SHOW HOST", Description: "Show connection details", RiskLevel: "SAFE"}
	case "SHOW SESSION":
		return CategoryDQL, OperationInfo{Category: CategoryDQL, Operation: "SHOW SESSION", Description: "Show session settings", RiskLevel: "SAFE"}

	// DML - Batch operations
	case "BEGIN BATCH":
		return CategoryDML, OperationInfo{Category: CategoryDML, Operation: "BEGIN BATCH", Description: "Start logged batch", RiskLevel: "MEDIUM"}
	case "BEGIN UNLOGGED":
		return CategoryDML, OperationInfo{Category: CategoryDML, Operation: "BEGIN UNLOGGED BATCH", Description: "Start unlogged batch", RiskLevel: "MEDIUM"}
	case "BEGIN COUNTER":
		return CategoryDML, OperationInfo{Category: CategoryDML, Operation: "BEGIN COUNTER BATCH", Description: "Start counter batch", RiskLevel: "MEDIUM"}
	case "APPLY BATCH":
		return CategoryDML, OperationInfo{Category: CategoryDML, Operation: "APPLY BATCH", Description: "Execute batch", RiskLevel: "MEDIUM"}

	// DDL - CREATE operations
	case "CREATE KEYSPACE":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "CREATE KEYSPACE", Description: "Create keyspace", RiskLevel: "MEDIUM"}
	case "CREATE TABLE", "CREATE COLUMNFAMILY":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "CREATE TABLE", Description: "Create table", RiskLevel: "MEDIUM"}
	case "CREATE INDEX":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "CREATE INDEX", Description: "Create index", RiskLevel: "MEDIUM"}
	case "CREATE CUSTOM":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "CREATE CUSTOM INDEX", Description: "Create custom index (SAI)", RiskLevel: "MEDIUM"}
	case "CREATE MATERIALIZED":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "CREATE MATERIALIZED VIEW", Description: "Create materialized view", RiskLevel: "MEDIUM"}
	case "CREATE TYPE":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "CREATE TYPE", Description: "Create UDT", RiskLevel: "MEDIUM"}
	case "CREATE FUNCTION", "CREATE OR":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "CREATE FUNCTION", Description: "Create function/aggregate", RiskLevel: "MEDIUM"}
	case "CREATE AGGREGATE":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "CREATE AGGREGATE", Description: "Create aggregate", RiskLevel: "MEDIUM"}
	case "CREATE TRIGGER":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "CREATE TRIGGER", Description: "Create trigger", RiskLevel: "MEDIUM"}

	// DDL - ALTER operations
	case "ALTER KEYSPACE":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "ALTER KEYSPACE", Description: "Modify keyspace", RiskLevel: "HIGH"}
	case "ALTER TABLE", "ALTER COLUMNFAMILY":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "ALTER TABLE", Description: "Modify table", RiskLevel: "HIGH"}
	case "ALTER MATERIALIZED":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "ALTER MATERIALIZED VIEW", Description: "Modify materialized view", RiskLevel: "HIGH"}
	case "ALTER TYPE":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "ALTER TYPE", Description: "Modify UDT", RiskLevel: "HIGH"}

	// DDL - DROP operations
	case "DROP KEYSPACE":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "DROP KEYSPACE", Description: "Delete keyspace", RiskLevel: "CRITICAL"}
	case "DROP TABLE", "DROP COLUMNFAMILY":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "DROP TABLE", Description: "Delete table", RiskLevel: "CRITICAL"}
	case "DROP INDEX":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "DROP INDEX", Description: "Delete index", RiskLevel: "HIGH"}
	case "DROP MATERIALIZED":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "DROP MATERIALIZED VIEW", Description: "Delete materialized view", RiskLevel: "HIGH"}
	case "DROP TYPE":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "DROP TYPE", Description: "Delete UDT", RiskLevel: "HIGH"}
	case "DROP FUNCTION":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "DROP FUNCTION", Description: "Delete function", RiskLevel: "HIGH"}
	case "DROP AGGREGATE":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "DROP AGGREGATE", Description: "Delete aggregate", RiskLevel: "HIGH"}
	case "DROP TRIGGER":
		return CategoryDDL, OperationInfo{Category: CategoryDDL, Operation: "DROP TRIGGER", Description: "Delete trigger", RiskLevel: "HIGH"}

	// DCL - Role operations
	case "CREATE ROLE":
		return CategoryDCL, OperationInfo{Category: CategoryDCL, Operation: "CREATE ROLE", Description: "Create role", RiskLevel: "HIGH"}
	case "ALTER ROLE":
		return CategoryDCL, OperationInfo{Category: CategoryDCL, Operation: "ALTER ROLE", Description: "Modify role", RiskLevel: "HIGH"}
	case "DROP ROLE":
		return CategoryDCL, OperationInfo{Category: CategoryDCL, Operation: "DROP ROLE", Description: "Delete role", RiskLevel: "CRITICAL"}
	case "GRANT ROLE":
		return CategoryDCL, OperationInfo{Category: CategoryDCL, Operation: "GRANT ROLE", Description: "Assign role", RiskLevel: "HIGH"}
	case "REVOKE ROLE":
		return CategoryDCL, OperationInfo{Category: CategoryDCL, Operation: "REVOKE ROLE", Description: "Remove role", RiskLevel: "HIGH"}

	// DCL - User operations (legacy)
	case "CREATE USER":
		return CategoryDCL, OperationInfo{Category: CategoryDCL, Operation: "CREATE USER", Description: "Create user", RiskLevel: "HIGH"}
	case "ALTER USER":
		return CategoryDCL, OperationInfo{Category: CategoryDCL, Operation: "ALTER USER", Description: "Modify user", RiskLevel: "HIGH"}
	case "DROP USER":
		return CategoryDCL, OperationInfo{Category: CategoryDCL, Operation: "DROP USER", Description: "Delete user", RiskLevel: "CRITICAL"}

	// DCL - Identity operations
	case "ADD IDENTITY":
		return CategoryDCL, OperationInfo{Category: CategoryDCL, Operation: "ADD IDENTITY", Description: "Add identity to role", RiskLevel: "HIGH"}
	case "DROP IDENTITY":
		return CategoryDCL, OperationInfo{Category: CategoryDCL, Operation: "DROP IDENTITY", Description: "Remove identity", RiskLevel: "HIGH"}

	// FILE - Export operations (safe in readonly mode)
	case "COPY TO":
		return CategoryFILE, OperationInfo{Category: CategoryFILE, Operation: "COPY TO", Description: "Export data to file", RiskLevel: "LOW"}

	// FILE - Import operations (only in readwrite+ mode)
	case "COPY FROM":
		return CategoryFILE, OperationInfo{Category: CategoryFILE, Operation: "COPY FROM", Description: "Import data from file", RiskLevel: "MEDIUM"}

	// SESSION - Display/output control (moved from FILE)
	case "CAPTURE ON", "CAPTURE OFF":
		return CategorySESSION, OperationInfo{Category: CategorySESSION, Operation: "CAPTURE", Description: "Control output capture", RiskLevel: "SAFE"}
	case "CAPTURE PARQUET", "CAPTURE JSON", "CAPTURE CSV":
		return CategorySESSION, OperationInfo{Category: CategorySESSION, Operation: "CAPTURE", Description: "Capture output to file", RiskLevel: "SAFE"}

	default:
		return CategoryUNKNOWN, OperationInfo{Category: CategoryUNKNOWN, Operation: twoWord, RiskLevel: "CRITICAL", Description: "Unknown/unclassified operation - requires confirmation"}
	}
}

// classifySingleWordOperation classifies single-word operations
func classifySingleWordOperation(operation string) OperationInfo {
	switch operation {
	// DQL - Read-only query
	case "SELECT":
		return OperationInfo{Category: CategoryDQL, Operation: "SELECT", Description: "Query data", RiskLevel: "SAFE"}
	case "DESCRIBE", "DESC":
		return OperationInfo{Category: CategoryDQL, Operation: "DESCRIBE", Description: "Show schema", RiskLevel: "SAFE"}
	case "SHOW":
		return OperationInfo{Category: CategoryDQL, Operation: "SHOW", Description: "Show information (cqlsh command)", RiskLevel: "SAFE"}
	case "LIST":
		return OperationInfo{Category: CategoryDQL, Operation: "LIST", Description: "List resources", RiskLevel: "SAFE"}

	// DML - Data modification
	case "INSERT":
		return OperationInfo{Category: CategoryDML, Operation: "INSERT", Description: "Insert data", RiskLevel: "LOW"}
	case "UPDATE":
		return OperationInfo{Category: CategoryDML, Operation: "UPDATE", Description: "Update data", RiskLevel: "MEDIUM"}
	case "DELETE":
		return OperationInfo{Category: CategoryDML, Operation: "DELETE", Description: "Delete data", RiskLevel: "HIGH"}
	case "BATCH":
		return OperationInfo{Category: CategoryDML, Operation: "BATCH", Description: "Batch operations", RiskLevel: "MEDIUM"}

	// DDL - Schema operations
	case "CREATE":
		return OperationInfo{Category: CategoryDDL, Operation: "CREATE", Description: "Create schema object", RiskLevel: "MEDIUM"}
	case "ALTER":
		return OperationInfo{Category: CategoryDDL, Operation: "ALTER", Description: "Modify schema object", RiskLevel: "HIGH"}
	case "DROP":
		return OperationInfo{Category: CategoryDDL, Operation: "DROP", Description: "Delete schema object", RiskLevel: "CRITICAL"}
	case "TRUNCATE":
		return OperationInfo{Category: CategoryDDL, Operation: "TRUNCATE", Description: "Delete all table data", RiskLevel: "CRITICAL"}
	case "USE":
		return OperationInfo{Category: CategoryDDL, Operation: "USE", Description: "Change keyspace", RiskLevel: "SAFE"}

	// DCL - Security operations
	case "GRANT":
		return OperationInfo{Category: CategoryDCL, Operation: "GRANT", Description: "Grant permission", RiskLevel: "HIGH"}
	case "REVOKE":
		return OperationInfo{Category: CategoryDCL, Operation: "REVOKE", Description: "Revoke permission", RiskLevel: "HIGH"}

	// SESSION - Configuration (CQLAI meta-commands)
	case "CONSISTENCY":
		return OperationInfo{Category: CategorySESSION, Operation: "CONSISTENCY", Description: "Set consistency level", RiskLevel: "SAFE"}
	case "PAGING":
		return OperationInfo{Category: CategorySESSION, Operation: "PAGING", Description: "Set page size", RiskLevel: "SAFE"}
	case "TRACING":
		return OperationInfo{Category: CategorySESSION, Operation: "TRACING", Description: "Enable/disable tracing", RiskLevel: "SAFE"}
	case "AUTOFETCH":
		return OperationInfo{Category: CategorySESSION, Operation: "AUTOFETCH", Description: "Auto-fetch all pages", RiskLevel: "SAFE"}
	case "EXPAND":
		return OperationInfo{Category: CategorySESSION, Operation: "EXPAND", Description: "Vertical output mode", RiskLevel: "SAFE"}
	case "OUTPUT":
		return OperationInfo{Category: CategorySESSION, Operation: "OUTPUT", Description: "Set output format", RiskLevel: "SAFE"}
	case "CAPTURE":
		return OperationInfo{Category: CategorySESSION, Operation: "CAPTURE", Description: "Capture output to file", RiskLevel: "SAFE"}
	case "SAVE":
		return OperationInfo{Category: CategorySESSION, Operation: "SAVE", Description: "Save results to file", RiskLevel: "SAFE"}

	// FILE - File operations
	case "SOURCE":
		return OperationInfo{Category: CategoryFILE, Operation: "SOURCE", Description: "Execute CQL from file", RiskLevel: "VARIABLE"}
	// COPY handled specially above (need TO/FROM direction)
	case "COPY":
		return OperationInfo{Category: CategoryFILE, Operation: "COPY", Description: "Import/export data (direction unknown)", RiskLevel: "MEDIUM"}

	default:
		return OperationInfo{Category: CategoryUNKNOWN, Operation: operation, Description: "Unknown/unclassified operation - requires confirmation", RiskLevel: "CRITICAL"}
	}
}

// GetCategoryDescription returns a human-readable description of a category
func GetCategoryDescription(category OperationCategory) string {
	switch category {
	case CategoryDQL:
		return "Data queries (SELECT, LIST, DESCRIBE, SHOW) - read-only"
	case CategorySESSION:
		return "Session settings (CONSISTENCY, PAGING, TRACING, CAPTURE, SAVE, etc) - always allowed"
	case CategoryDML:
		return "Data manipulation (INSERT, UPDATE, DELETE, BATCH) - modifies data"
	case CategoryDDL:
		return "Schema definition (CREATE, ALTER, DROP, TRUNCATE) - modifies schema"
	case CategoryDCL:
		return "Access control (roles, users, permissions, identities) - security changes"
	case CategoryFILE:
		return "File operations (COPY TO/FROM, SOURCE) - import/export"
	default:
		return "Unknown category"
	}
}

// GetCategoryOperationCount returns the approximate number of operations in a category
func GetCategoryOperationCount(category OperationCategory) int {
	switch category {
	case CategoryDQL:
		return 14
	case CategorySESSION:
		return 8
	case CategoryDML:
		return 8
	case CategoryDDL:
		return 28
	case CategoryDCL:
		return 13
	case CategoryFILE:
		return 3
	default:
		return 0
	}
}
