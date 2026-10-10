package router

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axonops/cqlai/internal/db"
)

// TestTheShellCommandsSayWhenThereIsNoConnection rather than stopping cqlai:
// SHOW VERSION typed before connecting reached into a session that was not
// there.
func TestTheShellCommandsSayWhenThereIsNoConnection(t *testing.T) {
	for _, session := range []*db.Session{nil, {}} {
		h := &MetaCommandHandler{session: session}
		for _, command := range []string{"SHOW VERSION", "SHOW HOST", "SHOW SESSION", "CONSISTENCY ONE",
			"CONSISTENCY", "TRACING ON", "PAGING 50", "AUTOFETCH ON", "SOURCE 'x.cql'", "COPY ks.t TO 'x.csv'"} {
			assert.NotPanics(t, func() {
				assert.Equal(t, notConnected, h.HandleMetaCommand(command), command)
			}, command)
		}
	}
}

// TestTheHandlerFollowsTheShellsSession: FILE > CONNECT gives the shell a new
// session, and the commands have to act on that one.
func TestTheHandlerFollowsTheShellsSession(t *testing.T) {
	old := metaHandler
	t.Cleanup(func() { metaHandler = old })

	metaHandler = nil
	first, second := &db.Session{}, &db.Session{}
	ProcessCommand("HELP", first, nil)
	assert.Same(t, first, metaHandler.session)
	ProcessCommand("HELP", second, nil)
	assert.Same(t, second, metaHandler.session, "the session connected to last")
}
