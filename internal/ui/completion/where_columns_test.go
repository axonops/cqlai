package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The columns of the table, where a column goes.
//
// The walk over a row statement is a function of what has been typed and
// nothing else, so at a column - after WHERE, after AND, after SET - it could
// only offer the note "<column name>". The table is named in the statement and
// the engine can ask for its columns, so for every table, stored or virtual,
// WHERE <Tab> offered a note where it could have offered the columns.

// engineKnowing is an engine whose cache already holds a table's columns, so
// nothing has to be asked of a cluster.
func engineKnowing(table string, columns ...string) *CompletionEngine {
	ce := NewCompletionEngine(nil, nil)
	ce.cache.columns[table] = columns
	return ce
}

// TestWhereOffersTheColumns.
func TestWhereOffersTheColumns(t *testing.T) {
	ce := engineKnowing("system_views.clients", "address", "port", "username")

	got := ce.Complete("SELECT * FROM system_views.clients WHERE ")

	assert.Contains(t, got, "address")
	assert.Contains(t, got, "port")
	assert.NotContains(t, got, columnHint, "the note is for when there is nothing better to offer")
}

// TestWhatIsTypedNarrowsThem.
func TestWhatIsTypedNarrowsThem(t *testing.T) {
	ce := engineKnowing("system_views.clients", "address", "port", "username")

	assert.Equal(t, []string{"address"}, ce.Complete("SELECT * FROM system_views.clients WHERE ad"))
}

// TestTheNextConditionOffersThemToo, after AND.
func TestTheNextConditionOffersThemToo(t *testing.T) {
	ce := engineKnowing("system_views.clients", "address", "port")

	assert.Equal(t, []string{"port"},
		ce.Complete("SELECT * FROM system_views.clients WHERE address = '10.0.0.1' AND po"))
}

// TestADeleteOffersThemToo: the same walk, the same gap.
func TestADeleteOffersThemToo(t *testing.T) {
	ce := engineKnowing("shop.orders", "order_id", "customer_id")

	assert.Equal(t, []string{"order_id"}, ce.Complete("DELETE FROM shop.orders WHERE ord"))
}

// TestTheNoteStaysWhereThereAreNoColumns: a table nothing can be found out
// about still gets the note rather than an empty list.
func TestTheNoteStaysWhereThereAreNoColumns(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	got := ce.Complete("SELECT * FROM nowhere.nothing WHERE ")

	assert.Contains(t, got, columnHint)
}

// TestTheTableOfTheStatement is the word after FROM, UPDATE or INTO - once it
// is finished, so the table being typed is not taken for the one being asked
// about.
func TestTheTableOfTheStatement(t *testing.T) {
	for input, want := range map[string]string{
		"SELECT * FROM ks.t WHERE ":        "ks.t",
		"SELECT * FROM t WHERE a = 1 AND ": "t",
		"DELETE FROM ks.t WHERE ":          "ks.t",
		"UPDATE ks.t SET ":                 "ks.t",
		"INSERT INTO ks.t (":               "ks.t",
		"SELECT * FROM ks.t":               "", // still being typed
		"SELECT * FROM ":                   "",
		"SELECT * ":                        "",
	} {
		assert.Equal(t, want, tableOfTheStatement(input), "%q", input)
	}
}
