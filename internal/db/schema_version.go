package db

import (
	"fmt"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// What Cassandra says about the state of its schema.
//
// Every node keeps a UUID that changes whenever the schema does, whoever
// changed it - this shell, another one, an application. It is what nodetool
// prints under "Schema versions", and reading it is one row: cheap enough to
// ask for on a timer, which is the only way to notice a change made somewhere
// else.

// SchemaVersion is the cluster's schema version as this node sees it.
func (s *Session) SchemaVersion() (string, error) {
	if s == nil || s.Session == nil {
		return "", fmt.Errorf("not connected")
	}

	var version gocql.UUID
	if err := s.Query("SELECT schema_version FROM system.local").Scan(&version); err != nil {
		return "", err
	}
	return version.String(), nil
}
