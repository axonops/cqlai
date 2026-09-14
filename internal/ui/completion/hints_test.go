package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWhatToTypeIsSaidWhereItCannotBeCompleted.
//
// Completion stopped wherever the next word was the user's to invent, and an
// empty list looks the same as completion having no idea.
func TestWhatToTypeIsSaidWhereItCannotBeCompleted(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	for typed, note := range map[string]string{
		"CREATE TABLE ":                    tableHint,
		"CREATE TABLE IF NOT EXISTS ":      tableHint,
		"CREATE KEYSPACE ":                 keyspaceHint,
		"CREATE MATERIALIZED VIEW ":        Hint("view name"),
		"CREATE INDEX ":                    Hint("index name"),
		"CREATE TABLE t (":                 columnHint,
		"CREATE TABLE t (id uuid, ":        columnHint,
		"ALTER TABLE t ADD ":               columnHint,
		"INSERT INTO t (":                  columnHint,
		"INSERT INTO t (a) VALUES (":       valueHint,
		"INSERT INTO t (a, b) VALUES (1, ": valueHint,
		"USE ":                             keyspaceHint,
		"SELECT * FROM ":                   tableHint,
		"UPDATE ":                          tableHint,
		"UPDATE t SET ":                    columnHint,
		"SELECT * FROM t WHERE ":           columnHint,
		"SELECT * FROM t WHERE id = ":      valueHint,
		"SELECT * FROM t WHERE a = 1 AND ": columnHint,
		"SELECT * FROM t ORDER BY ":        columnHint,
		"CREATE TABLE t (id int PRIMARY KEY) WITH gc_grace_seconds = ": Hint("seconds"),
	} {
		assert.Contains(t, ce.Complete(typed), note, "after %q", typed)
	}
}

// TestANoteIsNotACandidate: the positions that do have something to offer say
// so without one, and a note never crowds out a real answer.
func TestANoteIsNotACandidate(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	for _, typed := range []string{
		"",
		"SELECT ",
		"CREATE ",
		"CREATE TABLE t (id ",
		"CREATE TABLE t (id uuid ",
		"CREATE TABLE t (id uuid PRIMARY KEY) ",
		"CREATE TABLE t (id uuid PRIMARY KEY) WITH ",
		"SELECT * FROM t WHERE id = 1 ",
	} {
		for _, got := range ce.Complete(typed) {
			assert.False(t, IsHint(got), "after %q", typed)
		}
	}
}

// TestANameYouInventIsOfferedBesideTheKeywords.
//
// After CREATE TABLE the list offers IF, for IF NOT EXISTS. The name is the
// likelier thing to type, so the note belongs there with it rather than only
// where the list would be empty.
func TestANameYouInventIsOfferedBesideTheKeywords(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	got := ce.Complete("CREATE TABLE ")
	assert.Contains(t, got, "IF")
	assert.Contains(t, got, tableHint)
}

// TestTheNoteIsNeverPutIntoThePrompt, whichever key was pressed on it.
func TestTheNoteIsNeverPutIntoThePrompt(t *testing.T) {
	assert.True(t, IsHint(tableHint))
	assert.False(t, IsHint("users"))
	assert.False(t, IsHint("'filename'"))
}

// TestAMaterializedViewIsNotATable.
//
// It is one as far as its options go, and nothing like one before that: the
// rules that shape a CREATE TABLE offered the bracket its columns go in.
func TestAMaterializedViewIsNotATable(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.NotContains(t, ce.Complete("CREATE MATERIALIZED VIEW "), "(")
}

// TestAnOpenBracketIsNotOfferedTwice: `INSERT INTO t (` was answered with the
// bracket that had just been typed.
func TestAnOpenBracketIsNotOfferedTwice(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.NotContains(t, ce.Complete("INSERT INTO t ("), "(")
}

// TestATypeIsNotAComparison.
//
// `tuple<int, text> ` ends with a greater-than sign and is a finished type, so
// what follows it is what follows a type - not the value a comparison waits
// for.
func TestATypeIsNotAComparison(t *testing.T) {
	assert.False(t, endsWithComparison("CREATE TABLE t (m tuple<int, text> "))
	assert.False(t, endsWithComparison("CREATE TABLE t (m vector<float, 3> "))

	assert.True(t, endsWithComparison("SELECT * FROM t WHERE a > "))
	assert.True(t, endsWithComparison("SELECT * FROM t WHERE a >= "))
	assert.True(t, endsWithComparison("SELECT * FROM t WHERE a = "))
	assert.True(t, endsWithComparison("ALTER TABLE t WITH gc_grace_seconds="))
}

// TestTheBracketAfterInHoldsValues, not the columns the other brackets hold.
func TestTheBracketAfterInHoldsValues(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.Equal(t, []string{valueHint}, ce.Complete("SELECT * FROM t WHERE a IN ("))
	assert.Equal(t, []string{valueHint}, ce.Complete("SELECT * FROM t WHERE a IN (1, "))
}

// TestAPlaceThatTakesANumberSaysSo.
//
// Nothing else goes there, so the note replaces what the parser had to say:
// after LIMIT it offered WHERE and ALLOW FILTERING, which come before it.
func TestAPlaceThatTakesANumberSaysSo(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.Equal(t, []string{Hint("rows")}, ce.Complete("SELECT * FROM t LIMIT "))
	assert.Equal(t, []string{Hint("seconds")}, ce.Complete("UPDATE t USING TTL "))
	assert.Equal(t, []string{Hint("microseconds")}, ce.Complete("INSERT INTO t (a) VALUES (1) USING TIMESTAMP "))
}

// TestANameThatIsExpectedIsSaidWhenNothingIsKnown.
//
// With no cluster connected there are no names to offer, and these positions
// went silent: the note is what is left.
func TestANameThatIsExpectedIsSaidWhenNothingIsKnown(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.Equal(t, []string{tableHint}, ce.Complete("ALTER TABLE "))
	assert.Equal(t, []string{tableHint}, ce.Complete("COPY "))

	// And where something is known, the note does not crowd it out.
	assert.Equal(t, []string{"IF"}, ce.Complete("DROP TABLE "))
}
