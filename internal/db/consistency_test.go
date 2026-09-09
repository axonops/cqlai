package db

import (
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEveryOfferedLevelCanBeSet is the guard the hand-written lists did not
// have. Three of them offered SERIAL and LOCAL_SERIAL while SetConsistency
// rejected them, so picking one printed an error instead of changing anything.
func TestEveryOfferedLevelCanBeSet(t *testing.T) {
	levels := ConsistencyLevels()
	require.NotEmpty(t, levels)

	s := &Session{}
	for _, level := range levels {
		require.NoError(t, s.SetConsistency(level), "%s is offered but cannot be set", level)
		assert.Equal(t, level, s.Consistency(), "%s did not read back as itself", level)
	}
}

// TestSerialLevelsAreAccepted is the bug from #110. They are consistency levels
// like the rest, and cqlsh takes them for CONSISTENCY, so cqlai must too.
func TestSerialLevelsAreAccepted(t *testing.T) {
	for level, want := range map[string]gocql.Consistency{
		"SERIAL":       gocql.Serial,
		"LOCAL_SERIAL": gocql.LocalSerial,
	} {
		assert.Contains(t, ConsistencyLevels(), level)

		s := &Session{}
		require.NoError(t, s.SetConsistency(level))
		assert.Equal(t, want, s.consistency)
		assert.Equal(t, level, s.Consistency())
	}
}

// TestEveryDriverLevelIsCovered: the driver's own list is the authority on what
// a consistency level is, so anything it has and this table does not is a level
// nobody can select.
func TestEveryDriverLevelIsCovered(t *testing.T) {
	all := []gocql.Consistency{
		gocql.Any, gocql.One, gocql.Two, gocql.Three, gocql.Quorum, gocql.All,
		gocql.LocalQuorum, gocql.EachQuorum, gocql.Serial, gocql.LocalSerial, gocql.LocalOne,
	}

	require.Len(t, consistencyLevels, len(all))
	for _, level := range all {
		s := &Session{consistency: level}
		assert.NotEqual(t, "UNKNOWN", s.Consistency(), "%v has no name in the table", level)
	}
}

func TestUnknownLevelIsRejected(t *testing.T) {
	err := (&Session{}).SetConsistency("SIDEWAYS")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid consistency level: SIDEWAYS")
}

// TestChangingTheListDoesNotChangeTheTable: ConsistencyLevels builds a new
// slice each time, so a caller that sorts or trims what it gets back cannot
// reach into the table behind it.
func TestChangingTheListDoesNotChangeTheTable(t *testing.T) {
	first := ConsistencyLevels()
	first[0] = "TAMPERED"

	assert.Equal(t, "ANY", ConsistencyLevels()[0])
}
