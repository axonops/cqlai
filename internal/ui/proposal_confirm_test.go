package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// proposedIn is the shell with the MCP client's change put in its prompt.
func proposedIn(t *testing.T, statement string) *MainModel {
	t.Helper()
	m := multiLineModel(t)
	m.showProposal(mcpProposalMsg{Statement: statement, Implications: []string{"it changes rows"}})
	return m
}

// TestAProposalOverSeveralLinesIsStillConfirmed: the prompt turns a newline
// into a space, so the statement in it no longer matched the one proposed,
// and Enter ran it with no dialog.
func TestAProposalOverSeveralLinesIsStillConfirmed(t *testing.T) {
	m := proposedIn(t, "UPDATE shop.orders\n  SET total = 0\n  WHERE customer = 'ann' AND id = 1;")
	require.NotEmpty(t, m.input.Value(), "it is in the prompt")

	m, _ = m.handleEnterKey()

	assert.Equal(t, ModalConfirmDangerous, m.modal.Type, "Enter asks first")
	assert.Empty(t, m.commandHistory, "and runs nothing yet")
}

// TestAProposalsCommentDoesNotTakeTheRestOfIt: on one line, a comment would
// run to the end of the statement.
func TestAProposalsCommentDoesNotTakeTheRestOfIt(t *testing.T) {
	m := proposedIn(t, "DELETE FROM shop.orders -- the one asked about\n  WHERE customer = 'ann' AND id = 1;")
	assert.NotContains(t, m.input.Value(), "--")
	assert.Contains(t, m.input.Value(), "WHERE customer = 'ann' AND id = 1;")
}

// TestAProposalKeepsTheSpacingInItsStrings: only the prompt's own changes.
func TestAProposalKeepsTheSpacingInItsStrings(t *testing.T) {
	m := proposedIn(t, "INSERT INTO shop.notes (id, body) VALUES (1, 'two  spaces');")
	assert.Contains(t, m.input.Value(), "'two  spaces'")
}

// TestAProposalTooLongForThePromptIsNotPutThere: the prompt would cut it
// short, and a statement cut short is not one to leave a press of Enter away.
func TestAProposalTooLongForThePromptIsNotPutThere(t *testing.T) {
	m := multiLineModel(t)
	m.input.CharLimit = 40
	m.showProposal(mcpProposalMsg{Statement: "UPDATE shop.orders SET note = 'a long note indeed' WHERE id = 1;"})
	assert.Empty(t, m.input.Value())
	assert.Empty(t, m.proposed)
}
