package db

import (
	"fmt"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// The consistency levels CONSISTENCY accepts.
//
// One table, because there were five hand-written copies of this list and they
// disagreed. Three offered SERIAL and LOCAL_SERIAL and SetConsistency did not
// accept them, so picking one from the status line, or completing one with Tab,
// produced "invalid consistency level" rather than a change.
//
// SERIAL and LOCAL_SERIAL belong here. They are consistency levels like the
// rest - the driver has them as Consistency values, 0x08 and 0x09, and cqlsh
// takes them for CONSISTENCY too. A read at SERIAL goes through Paxos and sees
// any in-flight lightweight transaction, which is the point of them.
//
// In protocol order, which is also the order cqlsh lists them in.
var consistencyLevels = []struct {
	name  string
	level gocql.Consistency
}{
	{"ANY", gocql.Any},
	{"ONE", gocql.One},
	{"TWO", gocql.Two},
	{"THREE", gocql.Three},
	{"QUORUM", gocql.Quorum},
	{"ALL", gocql.All},
	{"LOCAL_QUORUM", gocql.LocalQuorum},
	{"EACH_QUORUM", gocql.EachQuorum},
	{"SERIAL", gocql.Serial},
	{"LOCAL_SERIAL", gocql.LocalSerial},
	{"LOCAL_ONE", gocql.LocalOne},
}

// ConsistencyLevels names every level SetConsistency accepts.
//
// Everything that offers the user a choice of level reads this - the status
// line chooser, tab completion, the CONSISTENCY usage message - so nothing can
// offer a level that then fails to apply.
func ConsistencyLevels() []string {
	names := make([]string, 0, len(consistencyLevels))
	for _, c := range consistencyLevels {
		names = append(names, c.name)
	}
	return names
}

// Consistency returns the current consistency level.
func (s *Session) Consistency() string {
	for _, c := range consistencyLevels {
		if c.level == s.consistency {
			return c.name
		}
	}
	return "UNKNOWN"
}

// SetConsistency sets the consistency level.
func (s *Session) SetConsistency(level string) error {
	for _, c := range consistencyLevels {
		if c.name == level {
			s.consistency = c.level
			return nil
		}
	}
	return fmt.Errorf("invalid consistency level: %s", level)
}
