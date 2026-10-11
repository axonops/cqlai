package ui

import (
	"testing"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/session"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// TestACommandFromTheMenuSaysWhenItFails: SOURCE from the File menu dropped an
// error, so a file whose statements failed looked as if nothing had run.
func TestACommandFromTheMenuSaysWhenItFails(t *testing.T) {
	m := helpModel()

	m.runCommand("SOURCE '/no/such/file.cql'")

	assert.Contains(t, ansi.Strip(m.fullHistoryContent), "Error: ")
}

// TestTheStatusLineFollowsTheSessionsKeyspace: a USE in a SOURCE file moves
// the session, and the status line has to say so without having seen the USE.
func TestTheStatusLineFollowsTheSessionsKeyspace(t *testing.T) {
	m := helpModel()
	m.sessionManager = session.NewManager(&config.Config{})
	_ = m.sessionManager.SetKeyspace("coc")
	m.statusBar.Keyspace = "old"

	m.processCommand("SHOW VERSION")

	assert.Equal(t, "coc", m.statusBar.Keyspace)
}
