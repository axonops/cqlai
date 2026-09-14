package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/axonops/cqlai/internal/ui/completion"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// completingModel is the prompt with a list of completions showing, the first
// one picked.
func completingModel(t *testing.T, typed string, offered ...string) *MainModel {
	t.Helper()

	m := helpModel()
	m.input.Focus()
	m.input.SetValue(typed)
	m.input.CursorEnd()
	m.showCompletions = true
	m.completions = offered
	m.completionIndex = 0
	return m
}

// TestAWordTakenFromTheListIsFinished.
//
// Without the space you reach for one every time, which is the opposite of what
// pressing Tab was for. The Space key already did this when it completed; Enter
// and Tab did not.
func TestAWordTakenFromTheListIsFinished(t *testing.T) {
	m := completingModel(t, "SELECT * FROM users WHERE ", "id")

	m, _ = m.handleCompletionSelection()

	assert.Equal(t, "SELECT * FROM users WHERE id ", m.input.Value())
	assert.Equal(t, len(m.input.Value()), m.input.Position())
}

// TestAKeyspaceIsNotFinished, since the table name follows the dot.
func TestAKeyspaceIsNotFinished(t *testing.T) {
	m := completingModel(t, "SELECT * FROM ", "my_keyspace.")

	m, _ = m.handleCompletionSelection()

	assert.Equal(t, "SELECT * FROM my_keyspace.", m.input.Value())
}

// TestAFunctionIsNotFinished, since its argument follows the bracket.
func TestAFunctionIsNotFinished(t *testing.T) {
	m := completingModel(t, "SELECT ", "COUNT(")

	m, _ = m.handleCompletionSelection()

	assert.Equal(t, "SELECT COUNT(", m.input.Value())
}

// TestAnOpenQuoteIsNotFinished, because what goes inside it comes next.
func TestAnOpenQuoteIsNotFinished(t *testing.T) {
	m := completingModel(t, "ALTER TABLE users WITH compaction = ", "{'")

	m, _ = m.handleCompletionSelection()

	assert.Equal(t, "ALTER TABLE users WITH compaction = {'", m.input.Value())
}

// TestAClosedQuoteIsFinished: a value in quotes is a value, and the next thing
// typed is the next thing.
func TestAClosedQuoteIsFinished(t *testing.T) {
	m := completingModel(t, "COPY users TO ", "'/tmp/export.csv'")

	m, _ = m.handleCompletionSelection()

	assert.Equal(t, "COPY users TO '/tmp/export.csv' ", m.input.Value())
}

// TestSomethingAlreadyEndingInASpaceGetsNoSecondOne.
//
// The table options are offered with the equals sign they need - "compaction = "
// - and a space on the end of that would be two.
func TestSomethingAlreadyEndingInASpaceGetsNoSecondOne(t *testing.T) {
	m := completingModel(t, "CREATE TABLE t (id uuid PRIMARY KEY) WITH ", "gc_grace_seconds = ")

	m, _ = m.handleCompletionSelection()

	assert.Equal(t, "CREATE TABLE t (id uuid PRIMARY KEY) WITH gc_grace_seconds = ", m.input.Value())
}

// TestBothWaysOfChoosingAgree: Space completes too, and has always added the
// space. Enter and Tab now do the same thing.
func TestBothWaysOfChoosingAgree(t *testing.T) {
	byEnter := completingModel(t, "SELECT * FROM ", "users")
	byEnter, _ = byEnter.handleCompletionSelection()

	bySpace := completingModel(t, "SELECT * FROM ", "users")
	bySpace, _ = bySpace.handleSpaceKey(tea.KeyPressMsg{Code: ' ', Text: " "})

	require.Equal(t, byEnter.input.Value(), bySpace.input.Value())
}

// TestEveryWayOfChoosingAgrees.
//
// There were three copies of this - Enter, the Tab that finds a single match,
// and Space - and they differed. Only Space added the space that finishes a
// word, so completing FROM with Tab left the cursor against it; only Enter knew
// what to do after an equals sign.
func TestEveryWayOfChoosingAgrees(t *testing.T) {
	for _, this := range []struct{ typed, chosen, want string }{
		{"SELECT * ", "FROM", "SELECT * FROM "},
		{"SELECT * FR", "FROM", "SELECT * FROM "},
		{"SELECT * FROM ", "users", "SELECT * FROM users "},
		{"SELECT * FROM my_keyspace.", "users", "SELECT * FROM my_keyspace.users "},
		{"SELECT ", "COUNT(", "SELECT COUNT("},
	} {
		byEnter := completingModel(t, this.typed, this.chosen)
		byEnter, _ = byEnter.handleCompletionSelection()
		assert.Equal(t, this.want, byEnter.input.Value(), "Enter on %q", this.typed)

		bySpace := completingModel(t, this.typed, this.chosen)
		bySpace, _ = bySpace.handleSpaceKey(tea.KeyPressMsg{Code: ' ', Text: " "})
		assert.Equal(t, this.want, bySpace.input.Value(), "Space on %q", this.typed)

		// Tab applies a single match without a list ever being shown, which is
		// the route that was missed.
		assert.Equal(t, this.want, applyCompletion(this.typed, this.chosen),
			"Tab on %q", this.typed)
	}
}

// TestTabCompletesAKeywordWithItsSpace, end to end through the key handler.
func TestTabCompletesAKeywordWithItsSpace(t *testing.T) {
	m := helpModel()
	m.input.Focus()
	m.completionEngine = completion.NewCompletionEngine(nil, nil)
	m.input.SetValue("SELECT * FR")
	m.input.CursorEnd()

	m, _ = m.handleTabKey()

	assert.Equal(t, "SELECT * FROM ", m.input.Value())
	assert.False(t, m.showCompletions, "a single match is applied rather than listed")
}

// TestAFinishedFilenameGetsItsSpaceToo, and an unfinished one does not.
//
// A path is not a word, but a finished one ends the same way: the next thing
// typed is the next thing. A directory has more path to come, and a quote still
// open has the rest of the name to come.
func TestAFinishedFilenameGetsItsSpaceToo(t *testing.T) {
	for _, this := range []struct{ typed, chosen, want string }{
		{"SOURCE ''", "schema.cql", "SOURCE 'schema.cql' "},
		{"SOURCE ''", "exports/", "SOURCE 'exports/'"},
		{"SOURCE '/tmp/", "schema.cql", "SOURCE '/tmp/schema.cql"},
		{"SOURCE '/tmp/", "exports/", "SOURCE '/tmp/exports/"},
	} {
		m := helpModel()
		m.input.SetValue(this.typed)
		m.input.CursorEnd()

		m.applyPathCompletion(this.chosen)

		assert.Equal(t, this.want, m.input.Value(), "%q + %q", this.typed, this.chosen)
	}
}

// TestANoteAboutWhatToTypeIsNeverPutIntoThePrompt.
//
// `<column name>` says what goes there. Pressing Enter, Space or Tab on it used
// to be the only way to find out it was not a word.
func TestANoteAboutWhatToTypeIsNeverPutIntoThePrompt(t *testing.T) {
	typed := "CREATE TABLE test.users ("

	assert.Equal(t, typed, applyCompletion(typed, completion.Hint("column name")))

	m := completingModel(t, typed, completion.Hint("column name"))
	m, _ = m.handleCompletionSelection()
	assert.Equal(t, typed, m.input.Value())
}

// TestSpaceOnANoteTypesASpace, since it is the first character of the name
// being typed rather than the end of a word.
func TestSpaceOnANoteTypesASpace(t *testing.T) {
	m := completingModel(t, "CREATE TABLE ", completion.Hint("table name"))

	m, _ = m.handleSpaceKey(tea.KeyPressMsg{Code: ' ', Text: " "})

	assert.Equal(t, "CREATE TABLE  ", m.input.Value())
	assert.False(t, m.showCompletions)
}

// TestTabShowsANoteRatherThanApplyingIt.
//
// A single completion is applied as soon as Tab finds it. A single note has
// nothing to apply, and applying it would take the list straight back down.
func TestTabShowsANoteRatherThanApplyingIt(t *testing.T) {
	m := helpModel()
	m.input.Focus()
	m.input.SetValue("CREATE TABLE test.users (")
	m.input.CursorEnd()
	m.completionEngine = completion.NewCompletionEngine(nil, nil)

	m, _ = m.handleTabKey()

	require.True(t, m.showCompletions)
	assert.Equal(t, []string{completion.Hint("column name")}, m.completions)
	assert.Equal(t, "CREATE TABLE test.users (", m.input.Value())
}

// TestTheModalCallsThemNotesWhenThatIsAllTheyAre.
func TestTheModalCallsThemNotesWhenThatIsAllTheyAre(t *testing.T) {
	styles := DefaultStyles()

	note := NewCompletionModal([]string{completion.Hint("table name")}, 0).RenderContent(styles)
	assert.Contains(t, note, "What to type")
	assert.NotContains(t, note, "Accept")

	words := NewCompletionModal([]string{"IF", completion.Hint("table name")}, 0).RenderContent(styles)
	assert.Contains(t, words, "Completions")
	assert.Contains(t, words, "Accept")
}

// TestTypingTheNameTakesTheNoteDown.
//
// The note answers Tab. Left up while the name it asked for is typed, it sits
// over the screen saying something the user is already doing.
func TestTypingTheNameTakesTheNoteDown(t *testing.T) {
	m := helpModel()
	m.input.Focus()
	m.input.SetValue("CREATE TABLE ")
	m.input.CursorEnd()
	m.completionEngine = completion.NewCompletionEngine(nil, nil)

	m, _ = m.handleTabKey()
	require.True(t, m.showCompletions)
	require.Contains(t, m.completions, completion.Hint("table name"))

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: 'u', Text: "u"})

	assert.Equal(t, "CREATE TABLE u", m.input.Value())
	assert.False(t, m.showCompletions)
}
