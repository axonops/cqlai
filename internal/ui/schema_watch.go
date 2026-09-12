package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/axonops/cqlai/internal/logger"
)

// Noticing a schema change made somewhere else.
//
// A CREATE, ALTER or DROP typed here is the one change cqlai can know about
// without asking (#183). Everything else - cqlsh, another cqlai, an application
// - happened elsewhere, and the schema browser and the completion behind Tab
// both go on describing the cluster as it was.
//
// Cassandra keeps a UUID that changes whenever the schema does, whoever changed
// it. Reading it is one row, so it can be read on a timer; it is the only query
// that answers "has anything changed" without asking about everything.

// schemaWatchInterval is how often to ask. Often enough that a change made in
// another window is there by the time you look, rarely enough that it is
// nothing next to what the shell does anyway.
const schemaWatchInterval = 5 * time.Second

// schemaVersionMsg is what the cluster said its schema version is.
type schemaVersionMsg struct {
	version string
	err     error
}

// watchSchemaVersion asks again after the interval.
func (m *MainModel) watchSchemaVersion() tea.Cmd {
	if !m.connected() {
		return nil
	}

	session := m.session
	return tea.Tick(schemaWatchInterval, func(time.Time) tea.Msg {
		version, err := session.SchemaVersion()
		return schemaVersionMsg{version: version, err: err}
	})
}

// handleSchemaVersion notices a change, and asks again.
func (m *MainModel) handleSchemaVersion(msg schemaVersionMsg) (*MainModel, tea.Cmd) {
	next := m.watchSchemaVersion()

	if msg.err != nil {
		// Nothing to say about it: a cluster that cannot be asked right now is
		// a connection problem, and the next query the user runs will say so
		// far better than a line appearing in the console on its own.
		logger.DebugfToFile("Schema", "Reading the schema version: %v", msg.err)
		return m, next
	}

	switch {
	case msg.version == "":
		return m, next
	case m.schemaVersion == "":
		// The first reading says what the schema is, not that it changed.
		m.schemaVersion = msg.version
		return m, next
	case msg.version == m.schemaVersion:
		return m, next
	}

	logger.DebugfToFile("Schema", "Schema changed elsewhere: %s -> %s", m.schemaVersion, msg.version)
	m.schemaVersion = msg.version

	// What a DDL statement typed here does, for a change made anywhere else.
	m.schemaChanged()
	m.refreshSchemaCache()

	return m, next
}

// refreshSchemaCache rebuilds what tab completion knows about the cluster.
//
// The router does this for a statement typed here; a change made somewhere else
// leaves it just as wrong, offering tables that have been dropped and not
// offering the ones just made.
func (m *MainModel) refreshSchemaCache() {
	if !m.connected() {
		return
	}

	cache := m.session.GetSchemaCache()
	if cache == nil {
		return
	}
	if err := cache.Refresh(); err != nil {
		logger.DebugfToFile("Schema", "Refreshing the schema cache: %v", err)
	}
}
