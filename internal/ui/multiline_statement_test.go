package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ran is the statement the shell last ran, as history keeps it.
func ran(t *testing.T, m *MainModel) string {
	t.Helper()
	require.NotEmpty(t, m.commandHistory, "nothing ran")
	return m.commandHistory[len(m.commandHistory)-1]
}

// TestACommentDoesNotTakeTheLinesAfterIt: the DELETE on two lines with a
// comment on the first deleted the whole partition, because the comment took
// the clustering condition with it.
func TestACommentDoesNotTakeTheLinesAfterIt(t *testing.T) {
	m := multiLineModel(t)
	m = typeLine(m, "DELETE FROM shop.orders WHERE customer = 'ann' -- just the one")
	require.True(t, m.multiLineMode)
	m = typeLine(m, "  AND id = 2;")

	assert.False(t, m.multiLineMode)
	assert.Contains(t, ran(t, m), "AND id = 2;")
	assert.NotContains(t, ran(t, m), "just the one", "history leaves the comment out")
}

// TestABatchTypedLineByLineIsOneStatement: it used to end at the first line
// ending with a semicolon, and ran its statements one by one, not as a batch.
func TestABatchTypedLineByLineIsOneStatement(t *testing.T) {
	m := multiLineModel(t)
	for _, line := range []string{
		"BEGIN BATCH",
		"  INSERT INTO shop.orders (customer, id) VALUES ('ann', 1);",
		"  UPDATE shop.customers SET last_order = 1 WHERE name = 'ann';",
	} {
		m = typeLine(m, line)
		require.True(t, m.multiLineMode, "still in the batch after %q", line)
		assert.Empty(t, m.commandHistory, "nothing has run yet")
	}
	m = typeLine(m, "APPLY BATCH;")

	assert.False(t, m.multiLineMode)
	require.Len(t, m.commandHistory, 1, "one statement")
	assert.True(t, strings.HasPrefix(ran(t, m), "BEGIN BATCH"), ran(t, m))
	assert.True(t, strings.HasSuffix(ran(t, m), "APPLY BATCH;"), ran(t, m))
}

// TestALineTheSameAsTheOneBeforeItIsKept: it was taken for a second press of
// Enter and dropped, writing different data from what was typed.
func TestALineTheSameAsTheOneBeforeItIsKept(t *testing.T) {
	m := multiLineModel(t)
	for _, line := range []string{
		"UPDATE shop.orders SET tags = tags + [",
		"'a',",
		"'a',",
		"'b'] WHERE customer = 'ann' AND id = 1;",
	} {
		m = typeLine(m, line)
	}
	assert.Contains(t, ran(t, m), "'a', 'a', 'b']")
}

// TestASemicolonInsideAStringEndsNothing.
func TestASemicolonInsideAStringEndsNothing(t *testing.T) {
	m := multiLineModel(t)
	m = typeLine(m, "INSERT INTO shop.notes (id, body) VALUES (1, 'first;")
	require.True(t, m.multiLineMode, "the string is still open")
	m = typeLine(m, "second');")
	assert.False(t, m.multiLineMode)
	assert.Contains(t, ran(t, m), "'first; second')")
}

// TestAConfirmedStatementIsRememberedOnOneLine: a dangerous statement typed
// over several lines is run from the confirmation, which kept it as typed -
// a comment, and a newline that splits it in two in the history file.
func TestAConfirmedStatementIsRememberedOnOneLine(t *testing.T) {
	m := multiLineModel(t)
	m.sessionManager = nil
	m = typeLine(m, "DELETE FROM shop.orders WHERE customer = 'ann' -- just the one")
	m = typeLine(m, "  AND id = 2;")
	m.modal = NewConfirmationModal("DELETE FROM shop.orders WHERE customer = 'ann' -- just the one\n  AND id = 2;")
	m, _ = m.answerModal(1) // Execute

	require.NotEmpty(t, m.commandHistory)
	last := m.commandHistory[len(m.commandHistory)-1]
	assert.NotContains(t, last, "\n")
	assert.NotContains(t, last, "just the one")
	assert.Contains(t, last, "AND id = 2;")
}

// TestTheHistoryFileHoldsOneStatementALine, whatever it is given.
func TestTheHistoryFileHoldsOneStatementALine(t *testing.T) {
	hm, err := NewHistoryManagerWithPath(t.TempDir() + "/history")
	require.NoError(t, err)
	require.NoError(t, hm.SaveCommand("SELECT *\nFROM t;"))
	again, err := NewHistoryManagerWithPath(hm.historyPath)
	require.NoError(t, err)
	assert.Equal(t, []string{"SELECT * FROM t;"}, again.GetHistory())
}
