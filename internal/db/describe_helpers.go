package db

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/session"
)

// tryServerSideDescribe attempts to use server-side DESCRIBE (Cassandra 4.0+)
// Returns the result if successful, nil otherwise
func (s *Session) tryServerSideDescribe(command string) interface{} {
	if s.IsVersion4OrHigher() {
		return s.ExecuteCQLQuery(command)
	}
	return nil
}

// parseQualifiedName parses a potentially qualified name (keyspace.object or just object)
// Returns the keyspace (if present), object name, and whether it was qualified
func parseQualifiedName(name string, sessionMgr *session.Manager) (keyspace string, objectName string, isQualified bool) {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		if len(parts) == 2 {
			return parts[0], parts[1], true
		}
	}

	// Not qualified, use current keyspace if available
	if sessionMgr != nil {
		keyspace = sessionMgr.CurrentKeyspace()
	}
	return keyspace, name, false
}

// buildDescribeCommand builds a DESCRIBE command for an object
// If the name is qualified (keyspace.object), uses it as-is
// Otherwise, uses currentKeyspace if available
func buildDescribeCommand(objectType string, name string, sessionMgr *session.Manager) (string, error) {
	keyspace, objectName, isQualified := parseQualifiedName(name, sessionMgr)

	if isQualified {
		return fmt.Sprintf("DESCRIBE %s %s", objectType, name), nil
	}

	if keyspace != "" {
		return fmt.Sprintf("DESCRIBE %s %s.%s", objectType, keyspace, objectName), nil
	}

	// No keyspace available
	return fmt.Sprintf("DESCRIBE %s %s", objectType, objectName), nil
}