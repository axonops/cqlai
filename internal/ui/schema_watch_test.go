package ui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheFirstReadingIsJustAReading.
//
// It says what the schema is, not that it changed: reacting to it would throw
// away everything the browser had fetched a moment earlier, every time.
func TestTheFirstReadingIsJustAReading(t *testing.T) {
	m := schemaModel(t)
	require.NotEmpty(t, m.schema.definitions)

	m, _ = m.handleSchemaVersion(schemaVersionMsg{version: "abc"})

	assert.Equal(t, "abc", m.schemaVersion)
	assert.NotEmpty(t, m.schema.definitions, "nothing has changed yet")
}

// TestAChangedVersionDropsWhatTheBrowserKnows, which is the point: the change
// was made in another window and there is no other way to hear about it.
func TestAChangedVersionDropsWhatTheBrowserKnows(t *testing.T) {
	m := schemaModel(t)
	m.schemaVersion = "before"
	require.NotEmpty(t, m.schema.definitions)

	m, _ = m.handleSchemaVersion(schemaVersionMsg{version: "after"})

	assert.Equal(t, "after", m.schemaVersion)
	assert.Empty(t, m.schema.definitions, "it should have fetched again")
}

// TestTheSameVersionChangesNothing, which is every reading but the rare one.
func TestTheSameVersionChangesNothing(t *testing.T) {
	m := schemaModel(t)
	m.schemaVersion = "same"
	before := len(m.schema.definitions)

	m, _ = m.handleSchemaVersion(schemaVersionMsg{version: "same"})

	assert.Len(t, m.schema.definitions, before)
}

// TestAChangeNoticedWhileElsewhereWaitsUntilYouLook.
func TestAChangeNoticedWhileElsewhereWaitsUntilYouLook(t *testing.T) {
	m := schemaModel(t)
	m.viewMode = "history"
	m.schemaVersion = "before"

	m, _ = m.handleSchemaVersion(schemaVersionMsg{version: "after"})

	assert.True(t, m.schema.stale, "marked, not fetched")
	assert.NotEmpty(t, m.schema.definitions)

	m, _ = m.openSchema()
	assert.Empty(t, m.schema.definitions, "and fetched when the tab is opened")
}

// TestAFailedReadingSaysNothing.
//
// A cluster that cannot be asked right now is a connection problem, and the
// next query the user runs will say so far better than a line appearing in the
// console on its own.
func TestAFailedReadingSaysNothing(t *testing.T) {
	m := schemaModel(t)
	m.schemaVersion = "before"
	before := m.fullHistoryContent

	m, _ = m.handleSchemaVersion(schemaVersionMsg{err: errors.New("connection refused")})

	assert.Equal(t, "before", m.schemaVersion)
	assert.Equal(t, before, m.fullHistoryContent)
	assert.NotEmpty(t, m.schema.definitions)
}

// TestWatchingStopsWithoutAConnection: there is nothing to ask.
func TestWatchingStopsWithoutAConnection(t *testing.T) {
	m := schemaModel(t)
	require.Nil(t, m.session)

	assert.Nil(t, m.watchSchemaVersion())
}
