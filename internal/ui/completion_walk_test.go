package ui

import (
	"slices"
	"testing"

	"github.com/axonops/cqlai/internal/ui/completion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Typing a statement by pressing Tab, the way a user does.
//
// The engine's answers were tested on their own and the apply was tested on its
// own, and the two still did not add up: `compaction = {'` was taken for a
// finished value, so Tab put a space after it and the key picked from the list
// landed past a stray quote - `{' 'class': `. The statement only comes out
// right if every step is taken through the keys that a user presses, which is
// what these do.

// tabbing is a prompt with the completion engine behind it.
func tabbing(t *testing.T, typed string) *MainModel {
	t.Helper()

	m := helpModel()
	m.input.Focus()
	m.completionEngine = completion.NewCompletionEngine(nil, nil)
	m.input.SetValue(typed)
	m.input.CursorEnd()
	return m
}

// tab presses Tab and takes the completion named: the one the list has under
// that name, or the single one Tab applies by itself.
func tab(t *testing.T, m *MainModel, wanted string) *MainModel {
	t.Helper()

	before := m.input.Value()
	m, _ = m.handleTabKey()

	if !m.showCompletions {
		require.Equal(t, applyCompletion(before, wanted), m.input.Value(),
			"Tab after %q applied something other than %q", before, wanted)
		return m
	}

	chosen := slices.Index(m.completions, wanted)
	require.GreaterOrEqual(t, chosen, 0,
		"after %q, %q was not offered: %v", before, wanted, m.completions)

	m.completionIndex = chosen
	m, _ = m.handleCompletionSelection()
	return m
}

// typing is the part a user types themselves: a name completion cannot know.
func typing(m *MainModel, text string) *MainModel {
	m.input.SetValue(m.input.Value() + text)
	m.input.CursorEnd()
	return m
}

// TestACompactionMapIsTypedWithTabAlone.
func TestACompactionMapIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "ALTER TABLE users WITH ")

	m = tab(t, m, "compaction = ")
	m = tab(t, m, "{'")
	m = tab(t, m, "'class': ")
	m = tab(t, m, "'LeveledCompactionStrategy'")
	m = tab(t, m, ", 'sstable_size_in_mb': ")

	assert.Equal(t,
		"ALTER TABLE users WITH compaction = {'class': 'LeveledCompactionStrategy', 'sstable_size_in_mb': ",
		m.input.Value())
}

// TestAMapIsClosedAndTheClauseCarriesOn.
func TestAMapIsClosedAndTheClauseCarriesOn(t *testing.T) {
	m := tabbing(t, "ALTER TABLE users WITH caching = {'keys': 'ALL'")

	m = tab(t, m, "}")
	assert.Equal(t, "ALTER TABLE users WITH caching = {'keys': 'ALL'} ", m.input.Value())

	m = tab(t, m, "AND")
	m = tab(t, m, "gc_grace_seconds = ")
	assert.Equal(t, "ALTER TABLE users WITH caching = {'keys': 'ALL'} AND gc_grace_seconds = ", m.input.Value())
}

// TestAnOpenQuoteIsNotAFinishedValue.
//
// Tab put the space that follows a finished value after the quote that opens
// one, which left the map with a quote in it that belonged to nothing.
func TestAnOpenQuoteIsNotAFinishedValue(t *testing.T) {
	m := tabbing(t, "ALTER TABLE users WITH compaction = {'")

	m, _ = m.handleTabKey()

	assert.Equal(t, "ALTER TABLE users WITH compaction = {'", m.input.Value(),
		"the prompt should be left alone")
	assert.Contains(t, m.completions, "'class': ")
}

// TestAKeyIsNotOfferedTwiceWhileTyping: the map being read back has to be the
// map as it was typed, quotes and all.
func TestAKeyIsNotOfferedTwiceWhileTyping(t *testing.T) {
	m := tabbing(t, "ALTER TABLE users WITH ")

	m = tab(t, m, "compaction = ")
	m = tab(t, m, "{'")
	m = tab(t, m, "'class': ")
	m = tab(t, m, "'SizeTieredCompactionStrategy'")

	m, _ = m.handleTabKey()
	require.True(t, m.showCompletions)
	assert.NotContains(t, m.completions, ", 'class': ", "class has been given")
	assert.Contains(t, m.completions, ", 'bucket_low': ", "the strategy's own keys")
}

// TestAnOptionTypedByHandIsGivenItsEqualsSign.
//
// Taking the option from the list brings the equals sign with it. Typing the
// name left the clause with nothing to offer at all.
func TestAnOptionTypedByHandIsGivenItsEqualsSign(t *testing.T) {
	m := tabbing(t, "ALTER TABLE users WITH compaction ")

	m = tab(t, m, "= ")
	assert.Equal(t, "ALTER TABLE users WITH compaction = ", m.input.Value())

	m = tab(t, m, "{'")
	assert.Equal(t, "ALTER TABLE users WITH compaction = {'", m.input.Value())
}

// TestAScalarOptionIsTypedWithTabAlone.
func TestAScalarOptionIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "ALTER TABLE users WITH ")

	m = tab(t, m, "read_repair = ")
	m = tab(t, m, "'BLOCKING'")
	m = tab(t, m, "AND")
	m = tab(t, m, "cdc = ")
	m = tab(t, m, "true")

	assert.Equal(t, "ALTER TABLE users WITH read_repair = 'BLOCKING' AND cdc = true ", m.input.Value())
}

// TestAWholeCreateTableIsTypedWithTabAlone, brackets, types and options.
func TestAWholeCreateTableIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "CREATE TABLE test.users ")

	m = tab(t, m, "(")
	m = typing(m, "id ")

	m = tab(t, m, "uuid")
	m = tab(t, m, "PRIMARY KEY")
	m = tab(t, m, ")")
	m = tab(t, m, "WITH")
	m = tab(t, m, "gc_grace_seconds = ")

	assert.Equal(t, "CREATE TABLE test.users (id uuid PRIMARY KEY) WITH gc_grace_seconds = ", m.input.Value())
}

// TestAClusteringOrderIsTypedWithTabAlone, brackets and direction included.
func TestAClusteringOrderIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "CREATE TABLE t (a int, b int, PRIMARY KEY (a, b)) WITH ")

	m = tab(t, m, "CLUSTERING ORDER BY (")
	m = typing(m, "b ")
	m = tab(t, m, "DESC")
	m = tab(t, m, ")")
	m = tab(t, m, "AND")
	m = tab(t, m, "gc_grace_seconds = ")

	assert.Equal(t,
		"CREATE TABLE t (a int, b int, PRIMARY KEY (a, b)) WITH CLUSTERING ORDER BY (b DESC) AND gc_grace_seconds = ",
		m.input.Value())
}

// TestAPrimaryKeyClauseIsTypedWithTabAlone.
func TestAPrimaryKeyClauseIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "CREATE TABLE t (")

	m = typing(m, "a ")
	m = tab(t, m, "int")
	m = typing(m, ", b ")
	m = tab(t, m, "text")
	m = typing(m, ", PRIMARY KEY ")
	m = tab(t, m, "(")
	m = typing(m, "a, b)")
	m = tab(t, m, ")")
	m = tab(t, m, "WITH")

	// The space is the one a completed type brings with it, ready for the word
	// that usually follows a type; here the user typed a comma instead.
	assert.Equal(t, "CREATE TABLE t (a int , b text , PRIMARY KEY (a, b)) WITH ", m.input.Value())
}

// TestAnIndexIsTypedWithTabAlone, target functions and options included.
func TestAnIndexIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "CREATE ")

	m = tab(t, m, "INDEX")
	m = typing(m, "by_email ")
	m = tab(t, m, "ON")
	m = typing(m, "users ")
	m = tab(t, m, "(")
	m = typing(m, "email ")
	m = tab(t, m, ")")
	m = tab(t, m, "USING")
	m = tab(t, m, "'sai'")

	assert.Equal(t, "CREATE INDEX by_email ON users (email) USING 'sai' ", m.input.Value())

	m = tab(t, m, "WITH")
	m = tab(t, m, "OPTIONS = ")
	m = tab(t, m, "{'")
	m = tab(t, m, "'case_sensitive': ")
	m = tab(t, m, "'false'")
	m = tab(t, m, "}")

	assert.Equal(t,
		"CREATE INDEX by_email ON users (email) USING 'sai' WITH OPTIONS = {'case_sensitive': 'false'} ",
		m.input.Value())
}

// TestAnIndexOnPartOfACollectionIsTypedWithTabAlone.
func TestAnIndexOnPartOfACollectionIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "CREATE INDEX by_tag ON users (")

	m = tab(t, m, "keys(")
	m = typing(m, "tags)")
	m = tab(t, m, ")")

	assert.Equal(t, "CREATE INDEX by_tag ON users (keys(tags)) ", m.input.Value())
}

// TestAFunctionIsTypedWithTabAlone: eight keywords in a row, none of which are
// guessable.
func TestAFunctionIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "CREATE ")

	m = tab(t, m, "FUNCTION")
	m = typing(m, "maxof ")
	m = tab(t, m, "(")
	m = typing(m, "a ")
	m = tab(t, m, "int")
	m = tab(t, m, ")")
	m = tab(t, m, "CALLED ON NULL INPUT")
	m = tab(t, m, "RETURNS")
	m = tab(t, m, "int")
	m = tab(t, m, "LANGUAGE")
	m = tab(t, m, "java")
	m = tab(t, m, "AS")

	assert.Equal(t,
		"CREATE FUNCTION maxof (a int) CALLED ON NULL INPUT RETURNS int LANGUAGE java AS ",
		m.input.Value())
}

// TestAKeyspaceIsTypedWithTabAlone.
func TestAKeyspaceIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "CREATE KEYSPACE shop ")

	m = tab(t, m, "WITH")
	m = tab(t, m, "replication = ")
	m = tab(t, m, "{'class': 'SimpleStrategy', 'replication_factor': 1}")
	m = tab(t, m, "AND")
	m = tab(t, m, "durable_writes = ")
	m = tab(t, m, "true")

	assert.Equal(t,
		"CREATE KEYSPACE shop WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1} AND durable_writes = true ",
		m.input.Value())
}

// TestAMaterializedViewIsTypedWithTabAlone.
func TestAMaterializedViewIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "CREATE MATERIALIZED VIEW by_email ")

	m = tab(t, m, "AS")
	m = tab(t, m, "SELECT")
	m = tab(t, m, "*")
	m = tab(t, m, "FROM")
	m = typing(m, "users ")
	m = tab(t, m, "WHERE")
	m = typing(m, "email ")
	m = tab(t, m, "IS NOT NULL")
	m = tab(t, m, "PRIMARY KEY")
	m = tab(t, m, "(")
	m = typing(m, "email, id)")
	m = tab(t, m, "WITH")
	m = tab(t, m, "comment = ")

	assert.Equal(t,
		"CREATE MATERIALIZED VIEW by_email AS SELECT * FROM users WHERE email IS NOT NULL PRIMARY KEY (email, id) WITH comment = ",
		m.input.Value())
}

// TestAQueryIsTypedWithTabAlone, through the clauses that narrow it down.
func TestAQueryIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "SELECT * ")

	m = tab(t, m, "FROM")
	m = typing(m, "users ")
	m = tab(t, m, "WHERE")
	m = typing(m, "age ")
	m = tab(t, m, ">=")
	m = typing(m, "18 ")
	m = tab(t, m, "LIMIT")
	m = typing(m, "10 ")
	m = tab(t, m, "ALLOW FILTERING")

	assert.Equal(t, "SELECT * FROM users WHERE age >= 18 LIMIT 10 ALLOW FILTERING ", m.input.Value())
}

// TestAWriteIsTypedWithTabAlone.
func TestAWriteIsTypedWithTabAlone(t *testing.T) {
	m := tabbing(t, "UPDATE users ")

	m = tab(t, m, "USING")
	m = tab(t, m, "TTL")
	m = typing(m, "60 ")
	m = tab(t, m, "SET")
	m = typing(m, "name ")
	m = tab(t, m, "= ")
	m = typing(m, "'ada' ")
	m = tab(t, m, "WHERE")
	m = typing(m, "id = 1 ")
	m = tab(t, m, "IF")
	m = tab(t, m, "EXISTS")

	assert.Equal(t, "UPDATE users USING TTL 60 SET name = 'ada' WHERE id = 1 IF EXISTS ", m.input.Value())
}
