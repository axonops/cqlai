package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnterOnGeneratedCQLRunsNothing: the window opened on Execute, so an
// Enter pressed a moment late ran CQL nobody had read. It opens on Edit: the
// statement goes into the prompt, and nothing runs.
func TestEnterOnGeneratedCQLRunsNothing(t *testing.T) {
	m := helpModel()
	m.windowWidth, m.windowHeight = 120, 40
	m.aiCQLModal = NewAICQLModal("DROP TABLE ks.orders;")
	before := m.fullHistoryContent

	m, _ = m.handleAICQLModal(tea.KeyPressMsg{Code: tea.KeyEnter})

	require.Nil(t, m.aiCQLModal)
	assert.Equal(t, "DROP TABLE ks.orders;", m.input.Value(), "the statement is in the prompt to read")
	assert.Equal(t, before, m.fullHistoryContent, "and nothing has run")
}
