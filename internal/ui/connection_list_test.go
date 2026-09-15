package ui

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
)

// The list of connections down the left of the CONNECT window.
//
// The window edited one set of settings, which was whatever the file held: a
// second cluster meant typing the host, the credentials and the SSL paths
// again, and saving it lost the first one.

// TestTheConnectionsInTheFileAreListed, the first of them marked as the one
// cqlai opens with.
func TestTheConnectionsInTheFileAreListed(t *testing.T) {
	m := connectModel(t, &config.Config{
		Host: "10.0.0.1", Port: 9042,
		Connections: []config.Config{
			{Name: "production", Host: "10.0.0.1"},
			{Name: "staging", Host: "10.0.1.1"},
		},
	})

	assert.Equal(t,
		[]string{connectionsHeading, createConnection, "", "production (default)", "staging"},
		m.preferences.connectionRows())
}

// TestAConnectionAtTheTopLevelBecomesTheFirstOfThem.
//
// A file written before there was a list has its connection at the top level.
// That is the one cqlai opens with, which is what the list calls the default.
func TestAConnectionAtTheTopLevelBecomesTheFirstOfThem(t *testing.T) {
	m := connectModel(t, &config.Config{Host: "10.0.0.1", Port: 9042, Keyspace: "shop"})

	assert.Equal(t,
		[]string{connectionsHeading, createConnection, "", "10.0.0.1 (default)"},
		m.preferences.connectionRows())
	assert.Equal(t, "shop", prefFieldValue(t, m, "Keyspace"))
}

// TestTheWindowOpensOnAConnection, so that Connect works without choosing one.
func TestTheWindowOpensOnAConnection(t *testing.T) {
	m := connectModel(t, &config.Config{Host: "127.0.0.1", Port: 9042})

	assert.Equal(t, 0, m.preferences.chosen)
	assert.Equal(t, "127.0.0.1", prefFieldValue(t, m, "Host"))
}

// TestWithNothingInTheFileThereIsOneConnectionToFillIn.
func TestWithNothingInTheFileThereIsOneConnectionToFillIn(t *testing.T) {
	m := connectModel(t, &config.Config{})

	assert.Equal(t, []string{connectionsHeading, createConnection, "", unnamedConnection + " (default)"}, m.preferences.connectionRows())
	assert.Equal(t, 0, m.preferences.chosen)
}

// TestChoosingAConnectionShowsItsSettings.
func TestChoosingAConnectionShowsItsSettings(t *testing.T) {
	m := connectModel(t, &config.Config{
		Connections: []config.Config{
			{Name: "production", Host: "10.0.0.1", Keyspace: "shop"},
			{Name: "staging", Host: "10.0.1.1", Keyspace: "trial"},
		},
	})

	m.chooseConnection(1)

	assert.Equal(t, "staging", prefFieldValue(t, m, "Name"))
	assert.Equal(t, "10.0.1.1", prefFieldValue(t, m, "Host"))
	assert.Equal(t, "trial", prefFieldValue(t, m, "Keyspace"))
}

// TestWhatWasTypedIsKeptWhenAnotherConnectionIsChosen.
//
// The settings on the right belong to the connection on the left, and moving
// down the list and back used to be the same as never having typed it.
func TestWhatWasTypedIsKeptWhenAnotherConnectionIsChosen(t *testing.T) {
	m := connectModel(t, &config.Config{
		Connections: []config.Config{
			{Name: "production", Host: "10.0.0.1"},
			{Name: "staging", Host: "10.0.1.1"},
		},
	})

	setPrefField(t, m, "Keyspace", "shop")
	m.chooseConnection(1)
	m.chooseConnection(0)

	assert.Equal(t, "shop", prefFieldValue(t, m, "Keyspace"))
}

// TestCreateNewConnectionAddsOneToFillIn.
func TestCreateNewConnectionAddsOneToFillIn(t *testing.T) {
	m := connectModel(t, &config.Config{
		Connections: []config.Config{{Name: "production", Host: "10.0.0.1"}},
	})

	m.preferences.onList = true
	m.preferences.listCursor = createRow
	m.useConnectionRow()

	assert.Equal(t, 1, m.preferences.chosen)
	assert.Empty(t, prefFieldValue(t, m, "Host"), "a new connection starts empty")
	assert.Equal(t,
		[]string{connectionsHeading, createConnection, "", "production (default)", unnamedConnection},
		m.preferences.connectionRows())
}

// TestTheListAndTheSettingsAreReachedFromEachOther.
func TestTheListAndTheSettingsAreReachedFromEachOther(t *testing.T) {
	m := connectModel(t, &config.Config{
		Connections: []config.Config{
			{Name: "production", Host: "10.0.0.1"},
			{Name: "staging", Host: "10.0.1.1"},
		},
	})

	m, _ = m.handlePreferencesKey(tea.KeyPressMsg{Code: tea.KeyLeft})
	require.True(t, m.preferences.onList, "left goes to the connections")
	assert.Equal(t, firstConnection, m.preferences.listCursor, "on the connection being shown")

	m, _ = m.handlePreferencesKey(tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = m.handlePreferencesKey(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.False(t, m.preferences.onList, "enter goes to its settings")
	assert.Equal(t, 1, m.preferences.chosen)
	assert.Equal(t, "staging", prefFieldValue(t, m, "Name"))
}

// TestTheCursorStepsOverTheBlankRowUnderTheButton.
func TestTheCursorStepsOverTheBlankRowUnderTheButton(t *testing.T) {
	m := connectModel(t, &config.Config{Connections: []config.Config{{Name: "production", Host: "10.0.0.1"}}})

	m.preferences.onList = true
	m.preferences.listCursor = firstConnection

	m.moveConnectionCursor(-1)
	assert.Equal(t, createRow, m.preferences.listCursor, "the button, not the blank row or the heading")

	m.moveConnectionCursor(-1)
	assert.Equal(t, createRow, m.preferences.listCursor, "and not the heading above it")

	m.moveConnectionCursor(1)
	assert.Equal(t, firstConnection, m.preferences.listCursor)
}

// TestSaveAndConnectWritesTheConnections.
//
// The connection cannot be made - there is no cluster - and the file is still
// written, because that is what the button does first.
func TestSaveAndConnectWritesTheConnections(t *testing.T) {
	m := connectModel(t, &config.Config{
		Connections: []config.Config{{Name: "production", Host: "10.0.0.1", Port: 9042}},
	})

	m.preferences.onList = true
	m.preferences.listCursor = createRow
	m.useConnectionRow()

	setPrefField(t, m, "Host", "10.0.2.2")
	setPrefField(t, m, "Port", "9042")
	setPrefField(t, m, "ConnectTimeout", "1")

	m, _ = m.pressPrefButton(prefSaveAndConnect)

	written := readConfigFile(t, m)
	require.Len(t, written.Connections, 2)
	assert.Equal(t, "10.0.2.2", written.Connections[0].Name, "named after its host, and now the default")
	assert.Equal(t, "production", written.Connections[1].Name)
	assert.Equal(t, "10.0.2.2", written.Host, "and it is what cqlai starts with")
}

// TestAConnectionThatWasNeverFilledInIsNotWritten.
func TestAConnectionThatWasNeverFilledInIsNotWritten(t *testing.T) {
	connections := []config.Config{
		{Name: "production", Host: "10.0.0.1"},
		{}, // Create New Connection, then Cancel
	}

	assert.Equal(t, []string{"production"}, namesOf(namedConnections(connections)))
}

// namesOf is the names of a list of connections.
func namesOf(connections []config.Config) []string {
	names := make([]string, 0, len(connections))
	for _, conn := range connections {
		names = append(names, config.ConnectionName(conn))
	}
	return names
}

// prefFieldValue is what a setting says.
func prefFieldValue(t *testing.T, m *MainModel, path string) string {
	t.Helper()
	return m.preferences.fields[prefIndex(t, m, path)].value()
}

// setPrefField types a value into a setting.
func setPrefField(t *testing.T, m *MainModel, path, value string) {
	t.Helper()
	m.setPrefField(prefIndex(t, m, path), value)
}

// readConfigFile is the file the window writes to, read back.
func readConfigFile(t *testing.T, m *MainModel) *config.Config {
	t.Helper()

	data, err := os.ReadFile(m.preferences.path) //nolint:gosec // the test wrote it
	if os.IsNotExist(err) {
		t.Fatalf("%s was not written", m.preferences.path)
	}
	require.NoError(t, err)

	cfg := &config.Config{}
	require.NoError(t, json.Unmarshal(data, cfg))
	return cfg
}

// TestTheListSaysWhichConnectionIsInUse.
//
// The window is opened to change connection, which is a question about the one
// in use: a list of names says nothing about which of them answered.
func TestTheListSaysWhichConnectionIsInUse(t *testing.T) {
	m := connectModel(t, &config.Config{
		Connections: []config.Config{
			{Name: "production", Host: "10.0.0.1"},
			{Name: "staging", Host: "10.0.1.1"},
		},
	})
	m.preferences.connected = "staging"

	assert.Equal(t,
		[]string{connectionsHeading, createConnection, "", "production (default)", "staging (connected)"},
		m.preferences.connectionRows())
	assert.Equal(t, 4, m.preferences.connectedRow())
}

// TestNothingIsMarkedWhenThereIsNoCluster.
func TestNothingIsMarkedWhenThereIsNoCluster(t *testing.T) {
	m := connectModel(t, &config.Config{
		Connections: []config.Config{{Name: "production", Host: "10.0.0.1"}},
	})

	assert.Equal(t, -1, m.preferences.connectedRow())
	for _, row := range m.preferences.connectionRows() {
		assert.NotContains(t, row, connectedNow)
	}
}

// TestTheWindowOpensOnTheConnectionInUse, which is the one a question about
// connections is usually about.
func TestTheWindowOpensOnTheConnectionInUse(t *testing.T) {
	m := prefModel(t, &config.Config{
		Connections: []config.Config{
			{Name: "production", Host: "10.0.0.1"},
			{Name: "staging", Host: "10.0.1.1", Keyspace: "trial"},
		},
	})
	m.closePreferences()

	// A session in hand, on the second of them.
	m.session = connectedSession()
	m.config = &config.Config{Name: "staging", Host: "10.0.1.1", SourcePath: m.config.SourcePath}

	m, _ = m.openConnect()
	require.True(t, m.preferences.active)

	assert.Equal(t, "staging", m.preferences.connected)
	assert.Equal(t, 1, m.preferences.chosen, "its settings are the ones showing")
	assert.Equal(t, "trial", prefFieldValue(t, m, "Keyspace"))
}

// TestTheDefaultIsTheOneCqlaiOpensWith, and saving a connection makes it that.
func TestTheDefaultIsTheOneCqlaiOpensWith(t *testing.T) {
	m := connectModel(t, &config.Config{
		Connections: []config.Config{
			{Name: "production", Host: "10.0.0.1", Port: 9042},
			{Name: "staging", Host: "10.0.1.1", Port: 9042},
		},
	})

	m.chooseConnection(1) // staging
	setPrefField(t, m, "ConnectTimeout", "1")
	m, _ = m.pressPrefButton(prefSaveAndConnect)

	written := readConfigFile(t, m)
	require.Len(t, written.Connections, 2)
	assert.Equal(t, "staging", written.Connections[0].Name, "connected to, so opened with")
	assert.Equal(t, "10.0.1.1", written.Host)

	standard, kept := written.DefaultConnection()
	require.True(t, kept)
	assert.Equal(t, "staging", config.ConnectionName(standard))
}

// TestBothPanesSayWhatTheyAre.
//
// Two columns of names and boxes with nothing above them leave the reader to
// work out which is which.
func TestBothPanesSayWhatTheyAre(t *testing.T) {
	m := connectModel(t, &config.Config{Connections: []config.Config{{Name: "production", Host: "10.0.0.1"}}})

	layer, ok := m.viewPreferences(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	drawn := stripAnsi(layer.Content)
	assert.Contains(t, drawn, connectionsHeading)
	assert.Contains(t, drawn, "CONNECTION DETAILS")

	// On the same row: one is the heading of each pane.
	for _, line := range strings.Split(drawn, "\n") {
		if strings.Contains(line, connectionsHeading) {
			assert.Contains(t, line, "CONNECTION DETAILS", "the two headings are the top of each pane")
			return
		}
	}
	t.Fatal("neither pane is titled")
}
