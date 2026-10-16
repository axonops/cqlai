package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/session"
)

// Names that need double quotes (issue 257). CQL folds an unquoted name to
// lower case, and completion put MyKeyspace.TableA in as it was: it ran as
// mykeyspace.tablea, which does not exist.

// mixedCaseEngine knows a keyspace and a table with capitals in them, and a
// plain one beside them.
func mixedCaseEngine(keyspace string) *CompletionEngine {
	var mgr *session.Manager
	if keyspace != "" {
		mgr = session.NewManager(&config.Config{Keyspace: keyspace})
	}
	ce := NewCompletionEngine(nil, mgr)
	ce.cache.keyspaces = []string{"MyKeyspace", "plain"}
	ce.cache.tables["MyKeyspace"] = []string{"TableA", "events"}
	ce.cache.tables["plain"] = []string{"users"}
	ce.cache.columns["MyKeyspace.TableA"] = []string{"UserId", "name"}
	return ce
}

// TestANameWithCapitalsIsOfferedInQuotes, and one without them is not.
func TestANameWithCapitalsIsOfferedInQuotes(t *testing.T) {
	ce := mixedCaseEngine("")

	got := ce.Complete("SELECT * FROM My")
	assert.Contains(t, got, `"MyKeyspace".`)

	got = ce.Complete("SELECT * FROM pla")
	assert.Contains(t, got, "plain.")
}

// TestTheTablesOfAQuotedKeyspace: the keyspace typed in quotes, or picked
// from the list, is the one whose tables are offered.
func TestTheTablesOfAQuotedKeyspace(t *testing.T) {
	ce := mixedCaseEngine("")

	got := ce.Complete(`SELECT * FROM "MyKeyspace".`)
	assert.ElementsMatch(t, []string{`"TableA"`, "events"}, got)

	assert.Equal(t, []string{`"TableA"`}, ce.Complete(`SELECT * FROM "MyKeyspace"."Ta`))
	assert.Equal(t, []string{`"TableA"`}, ce.Complete(`SELECT * FROM "MyKeyspace".ta`))
}

// TestWhatIsTypedInAnOpenQuoteNarrowsTheNames.
func TestWhatIsTypedInAnOpenQuoteNarrowsTheNames(t *testing.T) {
	ce := mixedCaseEngine("")

	assert.Contains(t, ce.Complete(`SELECT * FROM "My`), `"MyKeyspace".`)
}

// TestTheColumnsOfAQuotedTable: they were looked up in lower case, and a
// table with capitals had none.
func TestTheColumnsOfAQuotedTable(t *testing.T) {
	ce := mixedCaseEngine("")

	got := ce.Complete(`SELECT * FROM "MyKeyspace"."TableA" WHERE `)
	assert.Contains(t, got, `"UserId"`)
	assert.Contains(t, got, "name")

	ce = mixedCaseEngine("MyKeyspace")
	assert.Contains(t, ce.Complete(`SELECT * FROM "TableA" WHERE `), `"UserId"`, "in the keyspace in use")
}

// TestAnInsertColumnListIsQuoted.
func TestAnInsertColumnListIsQuoted(t *testing.T) {
	ce := mixedCaseEngine("")

	assert.Contains(t, ce.Complete(`INSERT INTO "MyKeyspace"."TableA" `), `("UserId", name)`)
}

// TestAKeywordIsNotQuoted: only the names the cache holds are.
func TestAKeywordIsNotQuoted(t *testing.T) {
	ce := mixedCaseEngine("")

	for _, s := range ce.Complete(`SELECT * FROM "MyKeyspace"."TableA" `) {
		assert.NotContains(t, s, `"`, "%s", s)
	}
}

// TestUnquoteIdentifiers leaves strings alone.
func TestUnquoteIdentifiers(t *testing.T) {
	for in, want := range map[string]string{
		`SELECT * FROM "MyKeyspace"."TableA"`: `SELECT * FROM MyKeyspace.TableA`,
		`SELECT * FROM "My`:                   `SELECT * FROM My`,
		`SELECT * FROM "a""b"`:                `SELECT * FROM a"b`,
		`INSERT INTO t JSON '{"a": 1}'`:       `INSERT INTO t JSON '{"a": 1}'`,
		`SELECT * FROM plain`:                 `SELECT * FROM plain`,
	} {
		assert.Equal(t, want, unquoteIdentifiers(in), in)
	}
}

// TestOpenNameAt is where a name still being typed in quotes starts.
func TestOpenNameAt(t *testing.T) {
	assert.Equal(t, 14, openNameAt(`SELECT * FROM "My`))
	assert.Equal(t, 27, openNameAt(`SELECT * FROM "MyKeyspace"."Ta`))
	assert.Equal(t, -1, openNameAt(`SELECT * FROM "MyKeyspace"`))
	assert.Equal(t, -1, openNameAt(`INSERT INTO t JSON '{"a`))
	assert.Equal(t, 14, openNameAt(`SELECT * FROM "a""b`))
}

// TestAReservedWordIsQuotedAsAName: a column called order is "order", and the
// keyword ORDER beside it is not touched.
func TestAReservedWordIsQuotedAsAName(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)
	ce.cache.keyspaces = []string{"k"}
	ce.cache.tables["k"] = []string{"t", "_staging"}
	ce.cache.columns["k.t"] = []string{"id", "order"}

	assert.Contains(t, ce.Complete("SELECT * FROM k.t WHERE or"), `"order"`)
	assert.Contains(t, ce.Complete("SELECT * FROM k._s"), `"_staging"`)
	assert.Equal(t, []string{"ORDER", "BY"}, ce.quoteNames([]string{"ORDER", "BY"}))
}
