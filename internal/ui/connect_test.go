package ui

import (
	"os"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
)

// connectModel is the CONNECT window open, over a configuration file.
func connectModel(t *testing.T, cfg *config.Config) *MainModel {
	t.Helper()

	m := prefModel(t, cfg) // writes the file and opens PREFERENCES
	m.closePreferences()
	m.windowHeight = 34

	m, _ = m.openConnect()
	require.True(t, m.preferences.active)
	require.Equal(t, connecting, m.preferences.purpose)
	return m
}

// TestConnectAsksForAConnectionAndNothingElse.
//
// The same window PREFERENCES uses, with the settings a connection takes: a
// second window would be a second copy of the fields, the checking and the
// layout, and would drift from this one.
func TestConnectAsksForAConnectionAndNothingElse(t *testing.T) {
	m := connectModel(t, &config.Config{Host: "localhost", Port: 9042})

	sections := map[string]bool{}
	paths := map[string]bool{}
	for _, field := range m.preferences.fields {
		if field.spec.section != "" {
			sections[field.spec.section] = true
		}
		paths[field.spec.path] = true
	}

	assert.Equal(t, map[string]bool{"CONNECTION DETAILS": true, "SSL": true}, sections)
	for _, wanted := range []string{"Host", "Port", "Username", "Password", "SSL.Enabled", "SSL.CAPath"} {
		assert.True(t, paths[wanted], "%s should be asked for", wanted)
	}
	for _, unwanted := range []string{"PageSize", "HistoryFile", "AI.Provider"} {
		assert.False(t, paths[unwanted], "%s is not part of connecting", unwanted)
	}
}

// TestConnectOffersThreeButtons, and PREFERENCES still offers two.
func TestConnectOffersThreeButtons(t *testing.T) {
	m := connectModel(t, &config.Config{})
	assert.Equal(t,
		[]string{prefConnect, prefSaveAndConnect, prefCancel},
		m.preferences.prefButtonLabels())

	m.closePreferences()
	m, _ = m.openPreferences()
	assert.Equal(t, []string{prefSave, prefCancel}, m.preferences.prefButtonLabels())
}

// TestEachButtonIsWhereItIsDrawn.
func TestEachButtonIsWhereItIsDrawn(t *testing.T) {
	m := connectModel(t, &config.Config{})

	g, ok := m.prefGeometry(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	row := g.y + 1 + g.buttonRow
	for i, start := range m.preferences.prefButtonStarts() {
		label := m.preferences.prefButtonLabels()[i]

		got, hit := m.prefButtonAt(m.windowWidth, m.windowHeight, g.x+2+start+2, row)
		require.True(t, hit, "nothing on %s", label)
		assert.Equal(t, label, got)
	}
}

// TestCancelChangesNothing.
func TestCancelChangesNothing(t *testing.T) {
	m := connectModel(t, &config.Config{Host: "localhost"})
	m.setPrefField(prefIndex(t, m, "Host"), "elsewhere")

	m, _ = m.pressPrefButton(prefCancel)

	assert.False(t, m.preferences.active)
	assert.Nil(t, m.session, "it should not have tried to connect")
}

// TestAConnectionThatFailsKeepsTheWindow.
//
// What you want next is to change a field and try again, not to type the whole
// thing out a second time.
func TestAConnectionThatFailsKeepsTheWindow(t *testing.T) {
	m := connectModel(t, &config.Config{Host: "192.0.2.1", Port: 9042, ConnectTimeout: 1})

	m, _ = m.pressPrefButton(prefConnect)

	require.True(t, m.preferences.active, "the window should still be open")
	assert.NotEmpty(t, m.preferences.failed, "and say what went wrong")
	assert.Contains(t, stripAnsi(m.prefHint()), m.preferences.failed)

	// Typing is the answer to it, so the message goes when you start.
	m, _ = m.handlePreferencesKey(tea.KeyPressMsg{Code: 'x', Text: "x"})
	assert.Empty(t, m.preferences.failed)
}

// TestSaveAndConnectWritesBeforeItTries.
//
// The connection is to an address with nothing on it, so what is under test is
// that the file was written on the way there rather than after arriving.
func TestSaveAndConnectWritesBeforeItTries(t *testing.T) {
	m := connectModel(t, &config.Config{Host: "192.0.2.1", Port: 9042, ConnectTimeout: 1})
	m.setPrefField(prefIndex(t, m, "Host"), "192.0.2.2")

	m, _ = m.pressPrefButton(prefSaveAndConnect)

	require.True(t, m.preferences.active, "the connection should have failed")
	written := readFile(t, m.preferences.path)
	assert.Contains(t, written, "192.0.2.2", "the settings should have been saved anyway")
}

// TestConnectIsTheFirstThingInTheMenu, on its own above the line: it is not a
// file operation, it is how you get a cluster at all.
func TestConnectIsTheFirstThingInTheMenu(t *testing.T) {
	items := fileMenuItems()
	require.NotEmpty(t, items)

	assert.Equal(t, "CONNECT", items[0].label)
	assert.True(t, items[1].rule, "a line under it")

	m := menuModel(t)
	m.fileMenu.selected = 0
	m, _ = m.chooseFileMenuItem()

	assert.True(t, m.preferences.active)
	assert.Equal(t, connecting, m.preferences.purpose)
}

// TestStartingWithNoClusterSaysSoAndSaysWhat.
func TestStartingWithNoClusterSaysSoAndSaysWhat(t *testing.T) {
	m := helpModel()
	m.connectError = "dial tcp 127.0.0.1:9042: connect: connection refused"

	welcome := stripAnsiForTest(m.getWelcomeMessage())

	assert.Contains(t, welcome, "Not connected")
	assert.Contains(t, welcome, "connection refused", "it should say why")
	assert.Contains(t, welcome, "CONNECT", "and where to go about it")
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	text, err := os.ReadFile(path) //nolint:gosec // a file the test wrote
	require.NoError(t, err)
	return string(text)
}
