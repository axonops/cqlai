package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
)

// The SCHEMA view: the cluster's keyspaces and tables as a tree, with the
// definition of whatever is selected beside it.
//
// Reading a table's definition meant typing DESCRIBE TABLE and knowing the name
// already. What you usually want first is to look - which keyspaces are there,
// what is in them, and what does that table look like.
//
// The definitions are the ones DESCRIBE produces, through the same calls, so
// the pane and the command cannot come to disagree about what a table is.

// schemaRow is one line of the tree: a keyspace, or a table inside one.
type schemaRow struct {
	keyspace string
	table    string // empty on a keyspace row
}

// key names a row, for the definitions kept against it.
func (r schemaRow) key() string {
	if r.table == "" {
		return r.keyspace
	}
	return r.keyspace + "." + r.table
}

// schemaBrowser is the state of the view.
type schemaBrowser struct {
	loaded    bool
	keyspaces []string
	tables    map[string][]string
	expanded  map[string]bool

	// definitions are kept as they are fetched. A definition is a round trip to
	// the cluster, and walking back up a tree should not repeat it.
	definitions map[string]string

	selected int // index into rows()
	scroll   int // first row of the tree showing

	detail       []string // the definition of the selected row, by line
	detailScroll int
	detailWidth  int // the widest line, for scrolling sideways later

	// drawn is the view as it was last rendered, both panes together. Dragging
	// the mouse selects out of it: there is no viewport behind this view to
	// anchor a selection to, and what is on screen is what a selection is over.
	drawn []string

	message string // what to say when there is no tree to draw
}

// rows is the tree as it stands: every keyspace, and the tables of the ones
// that are open.
//
// Drawing, clicking and moving the selection all come from this, so a click
// cannot land on a different row from the one drawn there.
func (s schemaBrowser) rows() []schemaRow {
	rows := make([]schemaRow, 0, len(s.keyspaces)*2)
	for _, keyspace := range s.keyspaces {
		rows = append(rows, schemaRow{keyspace: keyspace})
		if !s.expanded[keyspace] {
			continue
		}
		for _, table := range s.tables[keyspace] {
			rows = append(rows, schemaRow{keyspace: keyspace, table: table})
		}
	}
	return rows
}

// current is the selected row.
func (s schemaBrowser) current() (schemaRow, bool) {
	rows := s.rows()
	if s.selected < 0 || s.selected >= len(rows) {
		return schemaRow{}, false
	}
	return rows[s.selected], true
}

// openSchema switches to the view, fetching the keyspaces the first time.
//
// Nothing is fetched until then: a cluster can hold hundreds of keyspaces, and
// this should cost nothing to anyone not looking at it.
func (m *MainModel) openSchema() (*MainModel, tea.Cmd) {
	if m.viewMode == "schema" {
		// Already here: the tab is the way to ask for it again, for a cluster
		// that has changed underneath you.
		return m.refreshSchema()
	}

	m.viewMode = "schema"
	m.leaveAIConversation()
	m.input.Focus()

	if !m.schema.loaded {
		m.loadSchema()
	}
	return m, nil
}

// refreshSchema drops everything and asks the cluster again.
func (m *MainModel) refreshSchema() (*MainModel, tea.Cmd) {
	open := m.schema.expanded
	m.schema = schemaBrowser{}
	m.loadSchema()

	// The keyspaces that were open stay open, so asking for the schema again
	// does not close the tree you were reading.
	for keyspace, wasOpen := range open {
		if wasOpen {
			m.schema.expanded[keyspace] = true
			m.schema.tables[keyspace] = m.tablesOf(keyspace)
		}
	}
	m.showSchemaDetail()
	return m, nil
}

// loadSchema fetches the keyspaces.
func (m *MainModel) loadSchema() {
	m.schema.loaded = true
	m.schema.tables = map[string][]string{}
	m.schema.expanded = map[string]bool{}
	m.schema.definitions = map[string]string{}
	m.schema.message = ""

	if !m.connected() {
		m.schema.message = "Not connected."
		return
	}

	m.schema.keyspaces = m.keyspaceChoices()
	if len(m.schema.keyspaces) == 0 {
		m.schema.message = "No keyspaces."
		return
	}

	// Open on the keyspace in use, if there is one: it is the one the session
	// is already about.
	if current := m.currentKeyspace(); current != "" {
		for i, keyspace := range m.schema.keyspaces {
			if keyspace == current {
				m.schema.selected = i
				m.schema.expanded[keyspace] = true
				m.schema.tables[keyspace] = m.tablesOf(keyspace)
				break
			}
		}
	}
	m.showSchemaDetail()
}

// tablesOf is the tables of a keyspace, fetched the first time they are wanted.
func (m *MainModel) tablesOf(keyspace string) []string {
	if names, ok := m.schema.tables[keyspace]; ok {
		return names
	}

	names := m.tableChoices(keyspace)
	m.schema.tables[keyspace] = names
	return names
}

// showSchemaDetail puts the definition of the selected row in the right-hand
// pane, fetching it the first time it is asked for.
func (m *MainModel) showSchemaDetail() {
	m.schema.detail = nil
	m.schema.detailScroll = 0
	m.schema.detailWidth = 0

	row, ok := m.schema.current()
	if !ok {
		return
	}

	text, known := m.schema.definitions[row.key()]
	if !known {
		text = m.describeSchemaRow(row)
		m.schema.definitions[row.key()] = text
	}

	m.schema.detail = strings.Split(text, "\n")
	for _, line := range m.schema.detail {
		m.schema.detailWidth = max(m.schema.detailWidth, lipglossWidthOf(line))
	}
}

// describeSchemaRow asks the cluster what something is.
//
// The same calls DESCRIBE makes: a keyspace is its whole schema, the way
// DESCRIBE KEYSPACE gives it, and a table is its CREATE TABLE.
func (m *MainModel) describeSchemaRow(row schemaRow) string {
	if !m.connected() {
		return "Not connected."
	}

	if row.table == "" {
		schema, err := m.session.DBDescribeFullSchema(m.sessionManager, row.keyspace)
		if err != nil {
			logger.DebugfToFile("Schema", "Describing keyspace %s: %v", row.keyspace, err)
			return fmt.Sprintf("Cannot describe %s: %v", row.keyspace, err)
		}
		if text, ok := schema.(string); ok {
			return text
		}
		return fmt.Sprint(schema)
	}

	info, err := m.session.DescribeTableQuery(row.keyspace, row.table)
	if err != nil {
		logger.DebugfToFile("Schema", "Describing table %s: %v", row.key(), err)
		return fmt.Sprintf("Cannot describe %s: %v", row.key(), err)
	}
	if info == nil {
		return fmt.Sprintf("%s not found.", row.key())
	}
	// Without the "Table: ks.name" line it puts at the top: the heading over
	// the pane says which table this is, and saying it twice wastes the first
	// two rows of every definition.
	return db.FormatTableCreateStatement(info, false)
}

// selectSchemaRow moves the selection and shows what it lands on.
func (m *MainModel) selectSchemaRow(i int) (*MainModel, tea.Cmd) {
	rows := m.schema.rows()
	if len(rows) == 0 {
		return m, nil
	}

	m.schema.selected = min(max(i, 0), len(rows)-1)
	m.showSchemaDetail()
	return m, nil
}

// moveSchemaSelection moves it by n rows, stopping at the ends.
func (m *MainModel) moveSchemaSelection(n int) (*MainModel, tea.Cmd) {
	return m.selectSchemaRow(m.schema.selected + n)
}

// toggleSchemaRow opens or closes a keyspace.
//
// A table has nothing to open, so pressing Enter on one shows it again - which
// is what it does anyway, and is better than doing nothing.
func (m *MainModel) toggleSchemaRow() (*MainModel, tea.Cmd) {
	row, ok := m.schema.current()
	if !ok || row.table != "" {
		return m, nil
	}

	if m.schema.expanded[row.keyspace] {
		m.schema.expanded[row.keyspace] = false
		return m, nil
	}

	m.schema.expanded[row.keyspace] = true
	m.tablesOf(row.keyspace)
	return m, nil
}

// expandSchemaRow opens a keyspace, or steps into it when it is already open.
func (m *MainModel) expandSchemaRow() (*MainModel, tea.Cmd) {
	row, ok := m.schema.current()
	if !ok || row.table != "" {
		return m, nil
	}

	if !m.schema.expanded[row.keyspace] {
		return m.toggleSchemaRow()
	}
	if len(m.schema.tables[row.keyspace]) > 0 {
		return m.moveSchemaSelection(1)
	}
	return m, nil
}

// collapseSchemaRow closes a keyspace, or goes up to the one holding a table.
func (m *MainModel) collapseSchemaRow() (*MainModel, tea.Cmd) {
	row, ok := m.schema.current()
	if !ok {
		return m, nil
	}

	if row.table != "" {
		// Up to the keyspace it is in, which is where closing happens.
		for i, r := range m.schema.rows() {
			if r.keyspace == row.keyspace && r.table == "" {
				return m.selectSchemaRow(i)
			}
		}
		return m, nil
	}

	m.schema.expanded[row.keyspace] = false
	return m, nil
}

// scrollSchemaDetail moves the right-hand pane.
func (m *MainModel) scrollSchemaDetail(n, height int) (*MainModel, tea.Cmd) {
	if len(m.schema.detail) <= height {
		m.schema.detailScroll = 0
		return m, nil
	}

	m.schema.detailScroll = min(max(m.schema.detailScroll+n, 0), len(m.schema.detail)-height)
	return m, nil
}

// leaveAIConversation puts the chat view away, which every other view has to do
// when it takes over.
func (m *MainModel) leaveAIConversation() {
	if !m.aiConversationActive {
		return
	}

	m.aiConversationActive = false
	m.aiConversationInput.SetValue("")
	m.aiProcessing = false
	m.input.Placeholder = "Enter CQL command..."
	m.input.SetValue("")
}

// lipglossWidthOf is the drawn width of a line, ignoring any colour in it.
func lipglossWidthOf(line string) int {
	return len([]rune(stripAnsi(line)))
}

// schemaOwnsKeys reports whether the tree should take the keys that move around
// it.
//
// Only when nothing in front of it wants them. The tree is the background of
// this view, and the things drawn over it - the completion list under the
// prompt, the command history, a confirmation waiting for an answer - are all
// worked with the same arrows. Taking them unconditionally moved the tree while
// the list you were looking at stood still.
func (m *MainModel) schemaOwnsKeys() bool {
	switch {
	case m.viewMode != "schema":
		return false
	case m.showCompletions && len(m.completions) > 0:
		return false
	case m.historySearchMode:
		return false
	case m.modal.Type != ModalNone:
		return false
	}
	return true
}

// schemaKey handles the keys that work the tree, and reports whether it took
// the press.
//
// Only the keys that move around the view: everything else goes to the prompt,
// which still works from here. Typing a query while looking at a table's
// definition is the whole point of having both on one screen.
func (m *MainModel) schemaKey(msg tea.KeyPressMsg) (*MainModel, tea.Cmd, bool) {
	height := m.schemaHeight()

	switch msg.String() {
	case "up":
		updated, cmd := m.moveSchemaSelection(-1)
		return updated, cmd, true
	case "down":
		updated, cmd := m.moveSchemaSelection(1)
		return updated, cmd, true
	case "left":
		updated, cmd := m.collapseSchemaRow()
		return updated, cmd, true
	case "right":
		updated, cmd := m.expandSchemaRow()
		return updated, cmd, true
	case "enter":
		// Only on a keyspace, and only when the prompt is empty: a typed
		// statement is what Enter is for everywhere in cqlai, and this view
		// does not take that away.
		if strings.TrimSpace(m.input.Value()) != "" {
			return m, nil, false
		}
		updated, cmd := m.toggleSchemaRow()
		return updated, cmd, true
	case "pgup":
		updated, cmd := m.moveSchemaSelection(-height)
		return updated, cmd, true
	case "pgdown":
		updated, cmd := m.moveSchemaSelection(height)
		return updated, cmd, true
	case "home":
		updated, cmd := m.selectSchemaRow(0)
		return updated, cmd, true
	case "end":
		updated, cmd := m.selectSchemaRow(len(m.schema.rows()) - 1)
		return updated, cmd, true
	case "alt+up":
		updated, cmd := m.scrollSchemaDetail(-1, height)
		return updated, cmd, true
	case "alt+down":
		updated, cmd := m.scrollSchemaDetail(1, height)
		return updated, cmd, true
	}
	return m, nil, false
}

// clickSchema selects the row under the pointer, and opens or closes a keyspace
// when the row was already selected.
//
// A first click says which one you mean and shows it; the second folds it.
// Clicking a keyspace shut the moment you asked to see it would make its
// schema, which is what the click was for, flash past.
func (m *MainModel) clickSchema(col, row int) (*MainModel, tea.Cmd) {
	i, ok := m.schemaRowAt(col, row)
	if !ok {
		return m, nil
	}

	if i == m.schema.selected {
		return m.toggleSchemaRow()
	}
	return m.selectSchemaRow(i)
}

// scrollSchema moves whichever pane the pointer is over.
//
// Which side of the divider decides it, and nothing else: the wheel arrives
// over any row of the view, the headings included, and all of them belong to
// the pane they are drawn in.
func (m *MainModel) scrollSchema(col, delta int) (*MainModel, tea.Cmd) {
	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	if col >= g.treeWidth {
		return m.scrollSchemaDetail(delta, g.height)
	}

	// The tree scrolls by moving the selection, so what is showing and what is
	// selected cannot drift apart - and the definition follows, which is what
	// scrolling a tree of schema objects is for.
	return m.moveSchemaSelection(delta)
}

// connected reports whether there is a cluster to ask.
//
// The session is built before it is connected, so having one is not the same as
// being able to query through it: a half-built session takes a query straight
// into a nil pointer inside the driver.
func (m *MainModel) connected() bool {
	return m.session != nil && m.session.Session != nil
}
