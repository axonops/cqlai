package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEveryPositionInACreateIndex.
//
// Completion stopped at the index name. Everything after it - ON, the bracket,
// USING, the options - was left to be typed from memory, and the options are a
// map of names nobody remembers.
func TestEveryPositionInACreateIndex(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	for _, this := range []struct {
		typed  string
		offers []string
	}{
		{"CREATE ", []string{"INDEX", "CUSTOM"}},
		{"CREATE CUSTOM ", []string{"INDEX"}},
		{"CREATE INDEX ", []string{"IF", Hint("index name")}},
		{"CREATE INDEX IF NOT EXISTS ", []string{Hint("index name")}},
		{"CREATE INDEX by_email ", []string{"ON"}},
		{"CREATE INDEX by_email ON ", []string{tableHint}},
		{"CREATE INDEX by_email ON users ", []string{"("}},
		{"CREATE INDEX by_email ON users (", []string{"keys(", "values(", "entries(", "full(", columnHint}},
		{"CREATE INDEX by_email ON users (keys(", []string{columnHint}},
		{"CREATE INDEX by_email ON users (email ", []string{")"}},
		{"CREATE INDEX by_email ON users (email) ", []string{"USING"}},
		{"CREATE INDEX by_email ON users (email) USING ", []string{"'sai'", "'SASIIndex'"}},
		{"CREATE INDEX by_email ON users (email) USING 'sai' ", []string{"WITH"}},
		{"CREATE INDEX by_email ON users (email) USING 'sai' WITH ", []string{"OPTIONS = "}},
		{"CREATE INDEX by_email ON users (email) USING 'sai' WITH OPTIONS = ", []string{"{'"}},
	} {
		got := ce.Complete(this.typed)
		for _, wanted := range this.offers {
			assert.Contains(t, got, wanted, "after %q", this.typed)
		}
	}
}

// TestTheIndexOptionsAreOffered, and each one's values with them.
func TestTheIndexOptionsAreOffered(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)
	const start = "CREATE INDEX by_v ON t (v) USING 'sai' WITH OPTIONS = {"

	got := ce.Complete(start)
	assert.Contains(t, got, "'similarity_function': ")
	assert.Contains(t, got, "'case_sensitive': ")

	assert.Equal(t, []string{"'COSINE'", "'DOT_PRODUCT'", "'EUCLIDEAN'"},
		ce.Complete(start+"'similarity_function': "))
	assert.Equal(t, []string{"'LATENCY'", "'RECALL'"}, ce.Complete(start+"'optimize_for': "))
	assert.Equal(t, []string{"'true'", "'false'"}, ce.Complete(start+"'normalize': "))
	assert.Equal(t, []string{Hint("number")}, ce.Complete(start+"'maximum_node_connections': "))

	// The same map rules as a table's: each key once, and a way to close it.
	got = ce.Complete(start + "'normalize': 'true'")
	assert.NotContains(t, got, ", 'normalize': ")
	assert.Contains(t, got, ", 'ascii': ")
	assert.Contains(t, got, "}")
}

// TestACustomIndexIsAnIndex: CUSTOM sits between CREATE and INDEX and moved
// every word along by one.
func TestACustomIndexIsAnIndex(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.Equal(t, []string{"ON"}, ce.Complete("CREATE CUSTOM INDEX by_email "))
	assert.Contains(t, ce.Complete("CREATE CUSTOM INDEX by_email ON users (email) USING "), "'SASIIndex'")
}

// TestAnIndexTargetIsNotAColumnDefinition.
//
// The brackets of a CREATE INDEX hold one column, or a function naming part of
// a collection - not the `name type` of a CREATE TABLE.
func TestAnIndexTargetIsNotAColumnDefinition(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	got := ce.Complete("CREATE INDEX by_email ON users (")
	assert.NotContains(t, got, "uuid")

	// Inside one of the functions, a column and nothing else.
	assert.Equal(t, []string{columnHint}, ce.Complete("CREATE INDEX m ON users (entries("))
}

// TestSomethingThatIsNotAnIndexIsLeftAlone.
func TestSomethingThatIsNotAnIndexIsLeftAlone(t *testing.T) {
	for _, typed := range []string{
		"CREATE TABLE t (id uuid PRIMARY KEY)",
		"SELECT * FROM users WHERE ",
		"DROP INDEX by_email",
	} {
		_, handled := indexStatementCompletions(typed)
		assert.False(t, handled, "%q is not a CREATE INDEX", typed)
	}
}
