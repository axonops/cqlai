package router

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/session"
)

// CommandParser handles parsing and routing of commands without ANTLR
type CommandParser struct {
	session        *db.Session
	metaHandler    *MetaCommandHandler
	sessionManager *session.Manager
}

// NewCommandParser creates a new command parser
func NewCommandParser(sess *db.Session, handler *MetaCommandHandler, mgr *session.Manager) *CommandParser {
	return &CommandParser{
		session:        sess,
		metaHandler:    handler,
		sessionManager: mgr,
	}
}

// ParseCommand parses and routes a command to the appropriate handler
func (p *CommandParser) ParseCommand(command string) interface{} {
	// Trim and normalize
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}

	// Remove trailing semicolon for meta-commands
	trimmedCommand := strings.TrimSuffix(command, ";")
	upperCommand := strings.ToUpper(trimmedCommand)
	parts := strings.Fields(upperCommand)

	if len(parts) == 0 {
		return ""
	}

	// Route based on first word
	switch parts[0] {
	case "DESCRIBE", "DESC":
		return p.parseDescribe(trimmedCommand)
	case "LIST":
		return p.parseList(trimmedCommand)
	case "GRANT":
		return p.parseGrant(trimmedCommand)
	case "REVOKE":
		return p.parseRevoke(trimmedCommand)
	case "SHOW":
		// Some SHOW commands are meta-commands, others are CQL
		if p.isMetaShowCommand(upperCommand) {
			return p.metaHandler.HandleMetaCommand(command)
		}
		// Pass through to Cassandra
		return p.session.ExecuteCQLQuery(command)
	case "CONSISTENCY", "TRACING", "PAGING", "AUTOFETCH", "EXPAND", "SOURCE", "CAPTURE", "COPY", "HELP":
		// These are handled by the meta handler
		return p.metaHandler.HandleMetaCommand(command)
	case "CREATE", "ALTER", "DROP", "TRUNCATE", "USE":
		// DDL commands - pass through to Cassandra
		return p.session.ExecuteCQLQuery(command)
	case "SELECT", "INSERT", "UPDATE", "DELETE":
		// DML commands - pass through to Cassandra
		return p.session.ExecuteCQLQuery(command)
	case "BEGIN", "APPLY":
		// Batch commands - pass through to Cassandra
		return p.session.ExecuteCQLQuery(command)
	default:
		// Unknown command - let Cassandra handle it
		return p.session.ExecuteCQLQuery(command)
	}
}

// isMetaShowCommand determines if a SHOW command is a meta-command
func (p *CommandParser) isMetaShowCommand(upperCommand string) bool {
	// These SHOW commands are meta-commands
	return strings.Contains(upperCommand, "VERSION") ||
		strings.Contains(upperCommand, "HOST") ||
		strings.Contains(upperCommand, "SESSION")
}

// parseDescribe handles DESCRIBE commands
func (p *CommandParser) parseDescribe(command string) interface{} {
	// Remove any trailing semicolon first
	command = strings.TrimSuffix(strings.TrimSpace(command), ";")

	parts := strings.Fields(command)
	if len(parts) < 2 {
		return "Syntax error: DESCRIBE requires an object type or name"
	}

	// Get the object type (second word)
	upperParts := make([]string, len(parts))
	for i, part := range parts {
		upperParts[i] = strings.ToUpper(strings.TrimSuffix(part, ";"))
	}

	switch upperParts[1] {
	case "KEYSPACES":
		return p.describeKeyspaces()
	case "TABLES":
		return p.describeTables()
	case "CLUSTER":
		return p.describeCluster()
	case "TYPES":
		return p.describeTypes()
	case "FUNCTIONS":
		return p.describeFunctions()
	case "AGGREGATES":
		return p.describeAggregates()
	case "SCHEMA":
		return p.describeSchema()
	case "KEYSPACE":
		if len(parts) < 3 {
			return "Syntax error: DESCRIBE KEYSPACE requires a keyspace name"
		}
		return p.describeKeyspace(parts[2])
	case "TABLE":
		if len(parts) < 3 {
			return "Syntax error: DESCRIBE TABLE requires a table name"
		}
		return p.describeTable(parts[2])
	case "TYPE":
		if len(parts) < 3 {
			return "Syntax error: DESCRIBE TYPE requires a type name"
		}
		return p.describeType(parts[2])
	case "FUNCTION":
		if len(parts) < 3 {
			return "Syntax error: DESCRIBE FUNCTION requires a function name"
		}
		return p.describeFunction(parts[2])
	case "AGGREGATE":
		if len(parts) < 3 {
			return "Syntax error: DESCRIBE AGGREGATE requires an aggregate name"
		}
		return p.describeAggregate(parts[2])
	case "INDEX":
		if len(parts) < 3 {
			return "Syntax error: DESCRIBE INDEX requires an index name"
		}
		return p.describeIndex(parts[2])
	case "MATERIALIZED":
		if len(upperParts) >= 3 && upperParts[2] == "VIEW" {
			if len(parts) < 4 {
				return "Syntax error: DESCRIBE MATERIALIZED VIEW requires a view name"
			}
			return p.describeMaterializedView(parts[3])
		} else if len(upperParts) >= 3 && upperParts[2] == "VIEWS" {
			return p.describeMaterializedViews()
		}
		return "Syntax error: Expected DESCRIBE MATERIALIZED VIEW or DESCRIBE MATERIALIZED VIEWS"
	default:
		// Could be DESCRIBE <keyspace_or_table_name>
		// Try to determine what it is
		identifier := parts[1]
		return p.describeIdentifier(identifier)
	}
}

// parseList handles LIST commands
func (p *CommandParser) parseList(command string) interface{} {
	upperCommand := strings.ToUpper(command)

	switch {
	case strings.Contains(upperCommand, "ROLES"):
		// LIST ROLES
		return p.session.ExecuteCQLQuery("LIST ROLES")
	case strings.Contains(upperCommand, "USERS"):
		// LIST USERS
		return p.session.ExecuteCQLQuery("LIST USERS")
	case strings.Contains(upperCommand, "PERMISSIONS"):
		// LIST [ALL] PERMISSIONS [ON resource] [OF role]
		// This is complex enough to just pass through
		return p.session.ExecuteCQLQuery(command)
	default:
		return fmt.Sprintf("Unknown LIST command: %s", command)
	}
}

// parseGrant handles GRANT commands
func (p *CommandParser) parseGrant(command string) interface{} {
	// GRANT is a role/permission command - pass directly to Cassandra
	if err := p.session.ExecuteRoleCommand(command); err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return "GRANT successful"
}

// parseRevoke handles REVOKE commands
func (p *CommandParser) parseRevoke(command string) interface{} {
	// REVOKE is a role/permission command - pass directly to Cassandra
	if err := p.session.ExecuteRoleCommand(command); err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return "REVOKE successful"
}