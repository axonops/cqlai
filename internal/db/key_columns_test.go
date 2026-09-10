package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestASingleComponentKeyIsNotNumbered: a number on a key with one component
// says nothing.
func TestASingleComponentKeyIsNotNumbered(t *testing.T) {
	keys := KeyColumns{
		"id":      {Kind: "partition_key", Position: 0},
		"created": {Kind: "clustering", Position: 0},
	}

	assert.Equal(t, " (PK)", keys.Marker("id"))
	assert.Equal(t, " (C)", keys.Marker("created"))
}

// TestACompositePartitionKeyShowsItsOrder.
//
// The order of the components is part of the key, and the column order on
// screen need not match it, so it cannot be read off the display.
func TestACompositePartitionKeyShowsItsOrder(t *testing.T) {
	keys := KeyColumns{
		"tenant": {Kind: "partition_key", Position: 0},
		"region": {Kind: "partition_key", Position: 1},
	}

	assert.Equal(t, " (PK1)", keys.Marker("tenant"))
	assert.Equal(t, " (PK2)", keys.Marker("region"))
}

// TestSeveralClusteringColumnsShowTheirOrder, which is the sort order.
func TestSeveralClusteringColumnsShowTheirOrder(t *testing.T) {
	keys := KeyColumns{
		"created": {Kind: "clustering", Position: 0},
		"id":      {Kind: "clustering", Position: 1},
	}

	assert.Equal(t, " (C1)", keys.Marker("created"))
	assert.Equal(t, " (C2)", keys.Marker("id"))
}

// TestTheTwoKindsAreCountedSeparately: one partition column and two clustering
// columns means (PK) and (C1)/(C2), not (PK1).
func TestTheTwoKindsAreCountedSeparately(t *testing.T) {
	keys := KeyColumns{
		"id":      {Kind: "partition_key", Position: 0},
		"created": {Kind: "clustering", Position: 0},
		"seq":     {Kind: "clustering", Position: 1},
	}

	assert.Equal(t, " (PK)", keys.Marker("id"))
	assert.Equal(t, " (C1)", keys.Marker("created"))
	assert.Equal(t, " (C2)", keys.Marker("seq"))
}

func TestAColumnOutsideTheKeyGetsNoMarker(t *testing.T) {
	keys := KeyColumns{"id": {Kind: "partition_key", Position: 0}}

	assert.Empty(t, keys.Marker("email"))
	assert.Empty(t, KeyColumns{}.Marker("id"))
	assert.Empty(t, keys.Marker(""))
}

// TestRegularColumnsAreNotMarked: system_schema also lists "regular" and
// "static" columns, which are not part of the key.
func TestRegularColumnsAreNotMarked(t *testing.T) {
	keys := KeyColumns{
		"id":    {Kind: "partition_key", Position: 0},
		"email": {Kind: "regular", Position: -1},
		"count": {Kind: "static", Position: -1},
	}

	assert.Equal(t, " (PK)", keys.Marker("id"))
	assert.Empty(t, keys.Marker("email"))
	assert.Empty(t, keys.Marker("count"))
}

// TestEveryMarkerCanBeStrippedAgain is what keeps exports readable: a CSV
// written with markers has to be importable, so whatever Marker produces
// StripKeyMarker has to remove.
func TestEveryMarkerCanBeStrippedAgain(t *testing.T) {
	for _, keys := range []KeyColumns{
		{"id": {Kind: "partition_key", Position: 0}},
		{"a": {Kind: "partition_key", Position: 0}, "b": {Kind: "partition_key", Position: 1}},
		{"a": {Kind: "clustering", Position: 0}, "b": {Kind: "clustering", Position: 1}},
		{"a": {Kind: "partition_key", Position: 0}, "b": {Kind: "clustering", Position: 0}},
	} {
		for name := range keys {
			header := name + keys.Marker(name)
			assert.Equal(t, name, StripKeyMarker(header), "could not strip %q", header)
		}
	}
}

// TestStrippingLeavesEverythingElseAlone.
func TestStrippingLeavesEverythingElseAlone(t *testing.T) {
	for _, header := range []string{
		"id",
		"",
		"pk",                  // not a marker
		"my (PK) column",      // not at the end
		"weird(PK)",           // no space before it
		"comment (see notes)", // parentheses, but not a marker
	} {
		assert.Equal(t, header, StripKeyMarker(header))
	}
}

// TestStrippingHandlesTheOldMarkers, so a CSV exported before this change can
// still be read back.
func TestStrippingHandlesTheOldMarkers(t *testing.T) {
	assert.Equal(t, "id", StripKeyMarker("id (PK)"))
	assert.Equal(t, "created", StripKeyMarker("created (C)"))
}

func TestStrippingARowOfHeaders(t *testing.T) {
	got := StripKeyMarkers([]string{"tenant (PK1)", "region (PK2)", "created (C)", "email"})

	assert.Equal(t, []string{"tenant", "region", "created", "email"}, got)
	assert.Empty(t, StripKeyMarkers(nil))
}

// TestTheTableComesOutOfTheQuery covers what GetKeyColumns has to recognise
// before it can look anything up.
func TestTheTableComesOutOfTheQuery(t *testing.T) {
	tests := []struct{ query, keyspace, table string }{
		{"SELECT * FROM users", "", "users"},
		{"SELECT * FROM myks.users", "myks", "users"},
		{"select id from myks.users where id = 1", "myks", "users"},
		{"SELECT * FROM  myks.users ;", "myks", "users"},

		// Unquoted names are folded, the way Cassandra folds them.
		{"SELECT * FROM Users", "", "users"},
		{"SELECT * FROM MyKS.Users", "myks", "users"},

		// Quoted names keep the case they were created with.
		{`SELECT * FROM "MyTable"`, "", "MyTable"},
		{`SELECT * FROM "MyKS"."MyTable"`, "MyKS", "MyTable"},
		{`SELECT * FROM myks."MyTable"`, "myks", "MyTable"},
	}

	for _, tt := range tests {
		matches := fromClause.FindStringSubmatch(tt.query)
		if !assert.Len(t, matches, 3, "no FROM clause found in %q", tt.query) {
			continue
		}
		assert.Equal(t, tt.keyspace, cqlIdentifier(matches[1]), "keyspace of %q", tt.query)
		assert.Equal(t, tt.table, cqlIdentifier(matches[2]), "table of %q", tt.query)
	}
}

// TestQueriesWithNoTableAreIgnored rather than looking up nonsense.
func TestQueriesWithNoTableAreIgnored(t *testing.T) {
	for _, query := range []string{"SELECT 1", "", "DESCRIBE KEYSPACES"} {
		assert.Nil(t, fromClause.FindStringSubmatch(query), "%q has no FROM", query)
	}
}

// TestKeyLabelReadsBackWhatMarkerWrites. The header row shows the name and the
// row under it the label, so the label has to come back out of the header the
// marker went into. One file knows the bracket format; a second copy of it is
// how #118 happened.
func TestKeyLabelReadsBackWhatMarkerWrites(t *testing.T) {
	single := KeyColumns{
		"id":      {Kind: "partition_key", Position: 0},
		"created": {Kind: "clustering", Position: 0},
		"name":    {Kind: "regular", Position: 0},
	}
	assert.Equal(t, "PK", KeyLabel("id"+single.Marker("id")))
	assert.Equal(t, "C", KeyLabel("created"+single.Marker("created")))
	assert.Empty(t, KeyLabel("name"+single.Marker("name")))

	composite := KeyColumns{
		"a": {Kind: "partition_key", Position: 0},
		"b": {Kind: "partition_key", Position: 1},
		"c": {Kind: "clustering", Position: 0},
		"d": {Kind: "clustering", Position: 1},
	}
	assert.Equal(t, "PK1", KeyLabel("a"+composite.Marker("a")))
	assert.Equal(t, "PK2", KeyLabel("b"+composite.Marker("b")))
	assert.Equal(t, "C1", KeyLabel("c"+composite.Marker("c")))
	assert.Equal(t, "C2", KeyLabel("d"+composite.Marker("d")))
}

// TestKeyLabelIgnoresWhatIsNotAMarker. A column name can end in brackets.
func TestKeyLabelIgnoresWhatIsNotAMarker(t *testing.T) {
	for _, header := range []string{"name", "note (draft)", "count(*)", "", "(PK)"} {
		assert.Empty(t, KeyLabel(header), header)
	}
}
