package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEveryPositionInACreateTable.
//
// Nothing was offered after the table name, nor after a column name, nor after
// PRIMARY; types were offered where the primary key goes, and after the closing
// bracket where the options go. The statement defeats counting words: IF NOT
// EXISTS moves everything along by three, a keyspace qualifier is one word or
// two, and a column definition is any number of them.
func TestEveryPositionInACreateTable(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	for _, this := range []struct {
		typed  string
		offers []string
		not    []string
	}{
		{typed: "CREATE TABLE test.users ", offers: []string{"("}},
		{typed: "CREATE TABLE IF NOT EXISTS test.users ", offers: []string{"("}},
		{typed: "CREATE TABLE test.users (id ", offers: []string{"uuid", "text", "vector"}},
		{typed: "CREATE TABLE test.users (id uuid ", offers: []string{"PRIMARY KEY", "STATIC", ")"}, not: []string{"uuid"}},
		{typed: "CREATE TABLE test.users (id uuid PRIMARY ", offers: []string{"KEY"}},
		{typed: "CREATE TABLE test.users (id uuid PRIMARY KEY, name ", offers: []string{"text"}},
		{typed: "CREATE TABLE test.users (id uuid PRIMARY KEY) ", offers: []string{"WITH"}, not: []string{"uuid"}},
		{typed: "CREATE TABLE test.users (id uuid PRIMARY KEY) WITH ", offers: []string{"compaction = "}},
		{typed: "CREATE TABLE test.users (id uuid PRIMARY KEY) WITH compaction = {", offers: []string{"'class': "}},
	} {
		got := ce.Complete(this.typed)

		for _, wanted := range this.offers {
			assert.Contains(t, got, wanted, "after %q", this.typed)
		}
		for _, unwanted := range this.not {
			assert.NotContains(t, got, unwanted, "after %q", this.typed)
		}
	}
}

// TestAColumnNameIsYoursToInvent, so the note saying so is all that is offered
// where one goes - rather than the list of types the parser filled the silence
// with.
func TestAColumnNameIsYoursToInvent(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.Equal(t, []string{columnHint}, ce.Complete("CREATE TABLE test.users ("))
	assert.Equal(t, []string{columnHint}, ce.Complete("CREATE TABLE test.users (id uuid PRIMARY KEY, "))
}

// TestAPartialTypeIsFilteredToIt.
func TestAPartialTypeIsFilteredToIt(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	got := ce.Complete("CREATE TABLE test.users (id te")
	assert.Equal(t, []string{"text"}, got)

	got = ce.Complete("CREATE TABLE test.users (id uuid PRIMARY KEY) WI")
	assert.Equal(t, []string{"WITH"}, got)
}

// TestPunctuationIsNotAPartialWord.
//
// After `compaction = {` there is no word being typed, and filtering the keys
// by "{" would have left none of them.
func TestPunctuationIsNotAPartialWord(t *testing.T) {
	assert.Empty(t, partialWord("CREATE TABLE t (id uuid PRIMARY KEY) WITH compaction = {"))
	assert.Equal(t, "te", partialWord("CREATE TABLE t (id te"))
	assert.Equal(t, "gc_grace", partialWord("ALTER TABLE t WITH gc_grace"))
	assert.Empty(t, partialWord("SELECT * FROM users "))
}

// TestThePrimaryKeyClauseHasItsOwnBrackets.
//
// Taking the last open bracket for the column list put the types inside them:
// `PRIMARY KEY (` is where a column is named, not defined.
func TestThePrimaryKeyClauseHasItsOwnBrackets(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.Equal(t, []string{"("}, ce.Complete("CREATE TABLE t (a int, b text, PRIMARY KEY "))
	assert.Equal(t, []string{columnHint}, ce.Complete("CREATE TABLE t (a int, b text, PRIMARY KEY ("))
	assert.Equal(t, []string{columnHint}, ce.Complete("CREATE TABLE t (a int, b text, PRIMARY KEY (a, "))
	assert.Equal(t, []string{columnHint}, ce.Complete("CREATE TABLE t (a int, b text, PRIMARY KEY ((a), "))
}

// TestAColumnThatIsFinishedClosesTheList.
func TestAColumnThatIsFinishedClosesTheList(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.Equal(t, []string{")"}, ce.Complete("CREATE TABLE t (id uuid PRIMARY KEY "))
}

// TestATypeWithParametersIsTheOneWordItIs.
//
// `map<text, text>` splits into words at the space and at the comma, and the
// definition was read as four of them: the types were offered where the
// primary key goes, and a column name where the type's own parameter goes.
func TestATypeWithParametersIsTheOneWordItIs(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	// Inside the angle brackets, another type.
	for _, typed := range []string{
		"CREATE TABLE t (m map<",
		"CREATE TABLE t (m map<text, ",
		"CREATE TABLE t (m frozen<",
		"CREATE TABLE t (m frozen<map<text, ",
	} {
		assert.Contains(t, ce.Complete(typed), "text", "after %q", typed)
	}

	// And once it closes, the type is finished and the column with it.
	for _, typed := range []string{
		"CREATE TABLE t (m map<text, text> ",
		"CREATE TABLE t (m tuple<int, text> ",
		"CREATE TABLE t (m vector<float, 3> ",
		"CREATE TABLE t (m list<text> ",
	} {
		assert.Equal(t, []string{"PRIMARY KEY", "STATIC", ")"}, ce.Complete(typed), "after %q", typed)
	}
}

// TestTheColumnListEndsAtItsOwnBracket, not at the end of the statement: a
// CREATE TABLE has more brackets after it.
func TestTheColumnListEndsAtItsOwnBracket(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	got := ce.Complete("CREATE TABLE t (a int PRIMARY KEY) WITH CLUSTERING ORDER BY (b ")

	assert.NotContains(t, got, "uuid", "that is not a column definition")
	assert.Equal(t, []string{"ASC", "DESC"}, got)
}

// TestWhatAnAlterTableChangesTakesAType.
func TestWhatAnAlterTableChangesTakesAType(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.Contains(t, ce.Complete("ALTER TABLE t ADD joined "), "timestamp")
	assert.Contains(t, ce.Complete("ALTER TABLE t ALTER joined TYPE "), "timestamp")

	// The column being named is the user's to type, and is not a type.
	assert.Equal(t, []string{columnHint}, ce.Complete("ALTER TABLE t DROP "))
	assert.Equal(t, []string{columnHint}, ce.Complete("ALTER TABLE t RENAME "))
	assert.Equal(t, []string{columnHint}, ce.Complete("ALTER TABLE t RENAME a TO "))
}
