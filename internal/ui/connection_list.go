package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/axonops/cqlai/internal/config"
)

// The list of saved connections, down the left of the CONNECT window.
//
// The window edited one set of settings, which was whatever the file held: a
// second cluster meant typing the host, the credentials and the SSL paths
// again, and saving it lost the first one. The connections are kept in the
// file now, and this is how one is chosen.
//
// A connection is a configuration with only the connection settings in it, so
// the settings on the right are the same fields reading and writing the same
// paths as before. What changes is which configuration they are pointed at.

// The rows above the connections themselves: what the pane is, the button that
// makes a connection, and a blank line between that and the list.
//
// Written down once, because every row of the pane is found by counting from
// it: the cursor, the connection being shown, the one in use, and whatever a
// click landed on.
const (
	connectionsHeading = "CONNECTIONS"

	// createConnection is a button rather than a blank row at the end of the
	// list: making one is what you are here for the first time.
	createConnection = "+ Create New Connection"
)

const (
	headingRow      = 0 // CONNECTIONS
	createRow       = 1 // + Create New Connection
	firstConnection = 3 // the first of them, under a blank line
)

// unnamedConnection is what a connection with neither a name nor a host is
// shown as, which is one just made.
const unnamedConnection = "(new connection)"

// What is written after a connection's name.
//
// Said rather than marked: the list already has a cursor and a mark for the
// connection being shown, and more symbols are more things the reader has to
// be told the meaning of.
const (
	connectedNow   = "connected"
	defaultOpening = "default"
)

// said is the note after a connection's name, or "" when there is nothing to
// say about it.
func said(words ...string) string {
	kept := make([]string, 0, len(words))
	for _, word := range words {
		if word != "" {
			kept = append(kept, word)
		}
	}
	if len(kept) == 0 {
		return ""
	}
	return " (" + strings.Join(kept, ", ") + ")"
}

// connectionRows is the left pane, top to bottom.
func (p preferences) connectionRows() []string {
	rows := make([]string, 0, len(p.connections)+firstConnection)
	rows = append(rows, connectionsHeading, createConnection, "")

	for i, conn := range p.connections {
		name := config.ConnectionName(conn)
		if name == "" {
			name = unnamedConnection
		}

		// The first of them is the one cqlai opens with, which is what being
		// the default is.
		opening := ""
		if i == 0 {
			opening = defaultOpening
		}
		connected := ""
		if p.isConnected(conn) {
			connected = connectedNow
		}

		rows = append(rows, name+said(opening, connected))
	}
	return rows
}

// isConnected reports whether this is the connection the shell is using.
//
// By name, which is how the list tells them apart: two connections to the same
// host under different names are two different connections.
func (p preferences) isConnected(conn config.Config) bool {
	name := config.ConnectionName(conn)
	return p.connected != "" && name != "" && strings.EqualFold(name, p.connected)
}

// connectedRow is the row of the list the shell's connection is on, or -1.
func (p preferences) connectedRow() int {
	for i, conn := range p.connections {
		if p.isConnected(conn) {
			return i + firstConnection
		}
	}
	return -1
}

// connectionListWidth is how wide the left pane is: the widest thing in it.
func (p preferences) connectionListWidth() int {
	width := 0
	for _, row := range p.connectionRows() {
		width = max(width, lipgloss.Width(row))
	}
	return width + 2 // the marker in front of the row
}

// showConnectionRow brings a row of the left pane into view.
func (p *preferences) showConnectionRow(row int) {
	p.listScroll = min(p.listScroll, row)
	p.listScroll = max(p.listScroll, row-p.rows+1)
	p.listScroll = max(p.listScroll, 0)
}

// connectionLine is one row of the left pane, with the marker that says where
// the cursor is and which connection the settings belong to.
func (p preferences) connectionLine(row int) string {
	rows := p.connectionRows()
	if row < 0 || row >= len(rows) {
		return ""
	}
	if rows[row] == "" {
		return ""
	}

	if row == headingRow {
		return rows[row] // a heading, and not something to put a marker on
	}

	marker := "  "
	switch {
	case p.onList && row == p.listCursor:
		marker = "> "
	case row-firstConnection == p.chosen && row >= firstConnection:
		marker = "* " // the one the settings are showing
	}
	return marker + rows[row]
}

// chooseConnection puts a connection's settings into the window, keeping
// whatever was typed into the one being left.
func (m *MainModel) chooseConnection(i int) {
	if i < 0 || i >= len(m.preferences.connections) {
		return
	}

	m.storeConnection()
	m.preferences.chosen = i
	m.loadConnection()
}

// newConnection adds a blank connection to the list and shows it.
func (m *MainModel) newConnection() {
	m.storeConnection()

	m.preferences.connections = append(m.preferences.connections, config.Config{})
	m.preferences.chosen = len(m.preferences.connections) - 1
	m.preferences.listCursor = m.preferences.chosen + firstConnection
	m.loadConnection()
}

// storeConnection puts what has been typed back into the connection it was
// typed for, so that moving down the list and back does not lose it.
func (m *MainModel) storeConnection() {
	p := &m.preferences
	if p.chosen < 0 || p.chosen >= len(p.connections) {
		return
	}

	conn := p.connections[p.chosen]
	for _, field := range p.fields {
		if err := setPrefValue(&conn, field.spec.path, field.value()); err != nil {
			return // what is wrong with it is said on the window already
		}
	}
	p.connections[p.chosen] = conn
}

// loadConnection puts the chosen connection into the settings.
func (m *MainModel) loadConnection() {
	p := &m.preferences
	if p.chosen < 0 || p.chosen >= len(p.connections) {
		return
	}

	conn := p.connections[p.chosen]
	for i, field := range p.fields {
		value := prefValue(&conn, field.spec.path)
		p.fields[i].input = newPrefInput(field.spec, value)
		p.fields[i].yes = value == "true"
	}
	p.focus = 0
	p.scroll = 0
	p.clearMatches()
	p.failed = ""
}

// connectionUnderCursor is the connection the settings are showing.
func (p preferences) connectionUnderCursor() config.Config {
	conn := config.Config{}
	if p.chosen >= 0 && p.chosen < len(p.connections) {
		conn = p.connections[p.chosen]
	}
	return conn
}

// moveConnectionCursor moves up and down the left pane, over the rows that can
// be chosen: not the heading, and not the blank line under the button.
func (m *MainModel) moveConnectionCursor(delta int) {
	rows := m.preferences.connectionRows()

	at := m.preferences.listCursor + delta
	for at > createRow && at < len(rows) && !choosable(at) {
		at += delta
	}

	m.preferences.listCursor = min(max(at, createRow), len(rows)-1)
	m.preferences.showConnectionRow(m.preferences.listCursor)
}

// choosable reports whether a row of the pane is one the cursor can sit on.
func choosable(row int) bool {
	return row == createRow || row >= firstConnection
}

// useConnectionRow acts on the row the cursor is on: the button makes a
// connection, and a name shows it.
func (m *MainModel) useConnectionRow() {
	if m.preferences.listCursor == createRow {
		m.newConnection()
		return
	}
	m.chooseConnection(m.preferences.listCursor - firstConnection)
}

// namedConnections is the connections as they will be written: named, and
// without the blank one left behind by a Create New Connection that was never
// filled in.
func namedConnections(connections []config.Config) []config.Config {
	kept := make([]config.Config, 0, len(connections))
	for _, conn := range connections {
		if config.ConnectionName(conn) == "" {
			continue
		}
		saved := config.ConnectionSettings(conn)
		saved.Name = config.ConnectionName(conn)
		kept = append(kept, saved)
	}
	return kept
}
