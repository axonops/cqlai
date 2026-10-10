package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
)

// The SCHEMA view: the cluster's keyspaces as a tree, each holding its tables,
// materialized views, indexes, types, functions, aggregates and triggers, one
// group of each, with the definition of whatever is selected beside it.
//
// Reading a table's definition meant typing DESCRIBE TABLE and knowing the name
// already. What you usually want first is to look - which keyspaces are there,
// what is in them, and what does that table look like.
//
// The definitions are the ones DESCRIBE produces, through the same calls, so
// the pane and the command cannot come to disagree about what a table is.

// schemaRow is one line of the tree: a keyspace, a group of one kind of object
// inside it, or one object.
type schemaRow struct {
	keyspace string
	kind     string // "" on a keyspace row; db.KindTables and the others below it
	table    string // the object's name; empty on a keyspace or a group row
}

func (r schemaRow) isKeyspace() bool { return r.kind == "" }
func (r schemaRow) isGroup() bool    { return r.kind != "" && r.table == "" }

// key names a row, for the definitions kept against it.
func (r schemaRow) key() string {
	switch {
	case r.isKeyspace():
		return r.keyspace
	case r.isGroup():
		return r.keyspace + " " + r.kind
	case r.kind == db.KindTables:
		return r.keyspace + "." + r.table
	}
	return r.kind + " " + r.keyspace + "." + r.table
}

// schemaKindLabels is what a group is called in the tree, and what one of it
// is called over its definition.
var schemaKindLabels = map[string][2]string{
	db.KindTables:     {"Tables", "TABLE"},
	db.KindViews:      {"Materialized views", "MATERIALIZED VIEW"},
	db.KindIndexes:    {"Indexes", "INDEX"},
	db.KindTypes:      {"Types", "TYPE"},
	db.KindFunctions:  {"Functions", "FUNCTION"},
	db.KindAggregates: {"Aggregates", "AGGREGATE"},
	db.KindTriggers:   {"Triggers", "TRIGGER"},
}

// schemaBrowser is the state of the view.
type schemaBrowser struct {
	loaded    bool
	keyspaces []string
	tables    map[string][]string
	expanded  map[string]bool

	// objects is what a keyspace holds besides its tables, by kind, fetched
	// with its tables. groups says which groups are open, by group row key;
	// one not in it is open if it is the tables, and closed otherwise.
	objects map[string]map[string][]string
	groups  map[string]bool

	// definitions are kept as they are fetched. A definition is a round trip to
	// the cluster, and walking back up a tree should not repeat it.
	definitions map[string]string

	selected int // index into rows()
	scroll   int // first row of the tree showing

	detail       []string // the definition of the selected row, by line
	detailScroll int
	detailWidth  int // the widest line, for scrolling sideways later

	// review is what the model made of the definition, drawn under it.
	review schemaReview

	// drawn is the view as it was last rendered, both panes together. Dragging
	// the mouse selects out of it: there is no viewport behind this view to
	// anchor a selection to, and what is on screen is what a selection is over.
	drawn []string

	message string // what to say when there is no tree to draw

	// stale says the cluster has changed under what is kept here, so it is
	// fetched again before it is next looked at.
	stale bool

	// filter narrows the tree to the keyspaces and tables whose names contain
	// it, ignoring case. Empty shows everything.
	filter string

	// filterInput is the field the filter is typed into, while it has the
	// keys. A text input like the prompt, so its cursor blinks and moves the
	// way the prompt's does.
	filterInput textinput.Model
}

// filtering says the filter has the keys: what is typed goes into it rather
// than into the prompt.
func (s schemaBrowser) filtering() bool {
	return s.filterInput.Focused()
}

// rows is the tree as it stands: every keyspace, the groups of the ones that
// are open, and what is in the groups that are open.
//
// Drawing, clicking and moving the selection all come from this, so a click
// cannot land on a different row from the one drawn there.
func (s schemaBrowser) rows() []schemaRow {
	if s.filter != "" {
		return s.filteredRows()
	}

	rows := make([]schemaRow, 0, len(s.keyspaces)*2)
	for _, keyspace := range s.keyspaces {
		rows = append(rows, s.keyspaceRows(keyspace)...)
	}
	return rows
}

// keyspaceRows is a keyspace's row, and below it, when it is open, its groups
// and what is in the open ones.
func (s schemaBrowser) keyspaceRows(keyspace string) []schemaRow {
	rows := []schemaRow{{keyspace: keyspace}}
	if !s.expanded[keyspace] {
		return rows
	}
	for _, kind := range s.kindsOf(keyspace) {
		rows = append(rows, schemaRow{keyspace: keyspace, kind: kind})
		if !s.groupOpen(keyspace, kind) {
			continue
		}
		for _, name := range s.names(keyspace, kind) {
			rows = append(rows, schemaRow{keyspace: keyspace, kind: kind, table: name})
		}
	}
	return rows
}

// kindsOf is the kinds a keyspace has any of, tables first: a group with
// nothing in it is not drawn.
func (s schemaBrowser) kindsOf(keyspace string) []string {
	var kinds []string
	for _, kind := range append([]string{db.KindTables}, db.SchemaObjectKinds...) {
		if len(s.names(keyspace, kind)) > 0 {
			kinds = append(kinds, kind)
		}
	}
	return kinds
}

// names is what a keyspace has of one kind.
func (s schemaBrowser) names(keyspace, kind string) []string {
	if kind == db.KindTables {
		return s.tables[keyspace]
	}
	return s.objects[keyspace][kind]
}

// groupOpen reports whether a group shows what is in it. The tables are open
// until closed, so opening a keyspace shows them as it always has; the rest
// are closed until opened.
func (s schemaBrowser) groupOpen(keyspace, kind string) bool {
	if open, set := s.groups[schemaRow{keyspace: keyspace, kind: kind}.key()]; set {
		return open
	}
	return kind == db.KindTables
}

// filteredRows is the tree narrowed to what the filter names.
//
// A keyspace whose name matches is listed as it would be anyway. A keyspace
// that does not match is listed only for the objects of it that do, under
// their groups, and those are shown whether the keyspace and the group are
// open or not: they are what was searched for, and making someone open them
// to see it would hide the answer behind a click.
func (s schemaBrowser) filteredRows() []schemaRow {
	want := strings.ToLower(s.filter)
	matches := func(name string) bool { return strings.Contains(strings.ToLower(name), want) }

	var rows []schemaRow
	for _, keyspace := range s.keyspaces {
		if matches(keyspace) {
			rows = append(rows, s.keyspaceRows(keyspace)...)
			continue
		}

		var found []schemaRow
		for _, kind := range s.kindsOf(keyspace) {
			var hits []schemaRow
			for _, name := range s.names(keyspace, kind) {
				if matches(name) {
					hits = append(hits, schemaRow{keyspace: keyspace, kind: kind, table: name})
				}
			}
			if len(hits) > 0 {
				found = append(found, schemaRow{keyspace: keyspace, kind: kind})
				found = append(found, hits...)
			}
		}
		if len(found) > 0 {
			rows = append(rows, schemaRow{keyspace: keyspace})
			rows = append(rows, found...)
		}
	}
	return rows
}

// firstMatch is the first row whose own name the filter is in - the object or
// keyspace being looked for, rather than the keyspace or group listed above
// it.
func (s schemaBrowser) firstMatch() int {
	want := strings.ToLower(s.filter)
	for i, row := range s.rows() {
		if row.isGroup() {
			continue
		}
		name := row.keyspace
		if row.table != "" {
			name = row.table
		}
		if strings.Contains(strings.ToLower(name), want) {
			return i
		}
	}
	return 0
}

// current is the selected row.
func (s schemaBrowser) current() (schemaRow, bool) {
	rows := s.rows()
	if s.selected < 0 || s.selected >= len(rows) {
		return schemaRow{}, false
	}
	return rows[s.selected], true
}

// schemaChanged is called when a statement changes the schema.
//
// What the browser keeps describes the cluster as it was. Fetching it again
// here would be a round trip per statement, and a SOURCE file of fifty of them
// would make fifty: it is marked instead, and fetched again when it is next
// looked at. Unless that is now - a change made while looking at the tree
// should appear in it.
func (m *MainModel) schemaChanged() {
	if !m.schema.loaded {
		return // nothing has been fetched, so nothing is out of date
	}

	m.schema.stale = true
	if m.viewMode == "schema" {
		m.refreshSchema()
	}
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

	switch {
	case !m.schema.loaded:
		m.loadSchema()
	case m.schema.stale:
		// Changed since it was last looked at, by a statement typed here.
		return m.refreshSchema()
	}
	return m, nil
}

// refreshSchema drops everything and asks the cluster again.
func (m *MainModel) refreshSchema() (*MainModel, tea.Cmd) {
	open := m.schema.expanded
	groups := m.schema.groups
	selected := m.schema.selected
	m.schema = schemaBrowser{}
	m.loadSchema()

	// The keyspaces and groups that were open stay open, so asking for the
	// schema again does not close the tree you were reading.
	for keyspace, wasOpen := range open {
		if wasOpen {
			m.openKeyspace(keyspace)
		}
	}
	for key, isOpen := range groups {
		m.schema.groups[key] = isOpen
	}

	// And the selection stays where it was, as far as the tree still goes.
	if rows := len(m.schema.rows()); rows > 0 {
		m.schema.selected = min(selected, rows-1)
	}
	m.showSchemaDetail()
	return m, nil
}

// loadSchema fetches the keyspaces.
func (m *MainModel) loadSchema() {
	m.schema.loaded = true
	m.schema.tables = map[string][]string{}
	m.schema.expanded = map[string]bool{}
	m.schema.objects = map[string]map[string][]string{}
	m.schema.groups = map[string]bool{}
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
				m.openKeyspace(keyspace)
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

// objectsOf is what a keyspace holds besides its tables, fetched the first
// time it is wanted.
func (m *MainModel) objectsOf(keyspace string) map[string][]string {
	if m.schema.objects == nil {
		m.schema.objects = map[string]map[string][]string{}
	}
	if objects, ok := m.schema.objects[keyspace]; ok {
		return objects
	}
	objects := map[string][]string{}
	if m.connected() {
		objects = m.session.KeyspaceObjects(keyspace)
	}
	m.schema.objects[keyspace] = objects
	return objects
}

// loadAllSchemaNames fetches the names in every keyspace not yet opened, for
// the filter to search: seven queries in all, however many keyspaces there
// are, rather than seven for each and one more for each of their tables.
func (m *MainModel) loadAllSchemaNames() {
	missing := false
	for _, keyspace := range m.schema.keyspaces {
		_, haveTables := m.schema.tables[keyspace]
		_, haveObjects := m.schema.objects[keyspace]
		missing = missing || !haveTables || !haveObjects
	}
	if !missing || !m.connected() {
		for _, keyspace := range m.schema.keyspaces {
			m.tablesOf(keyspace)
			m.objectsOf(keyspace)
		}
		return
	}

	tables, objects := m.session.AllSchemaNames()
	if m.schema.tables == nil {
		m.schema.tables = map[string][]string{}
	}
	if m.schema.objects == nil {
		m.schema.objects = map[string]map[string][]string{}
	}
	for _, keyspace := range m.schema.keyspaces {
		if _, ok := m.schema.tables[keyspace]; !ok {
			m.schema.tables[keyspace] = tables[keyspace]
		}
		if _, ok := m.schema.objects[keyspace]; !ok {
			found := objects[keyspace]
			if found == nil || m.session.IsVirtualKeyspace(keyspace) {
				found = map[string][]string{} // as objectsOf has it
			}
			m.schema.objects[keyspace] = found
		}
	}
}

// openKeyspace opens a keyspace, fetching what it holds the first time.
func (m *MainModel) openKeyspace(keyspace string) {
	m.schema.expanded[keyspace] = true
	m.tablesOf(keyspace)
	m.objectsOf(keyspace)
}

// showSchemaDetail puts the definition of the selected row in the right-hand
// pane, fetching it the first time it is asked for.
func (m *MainModel) showSchemaDetail() {
	m.schema.detail = nil
	m.schema.detailScroll = 0
	m.schema.detailWidth = 0

	row, ok := m.schema.current()
	if !ok {
		m.closeSchemaReview()
		return
	}

	// A review is of the definition it was asked about. Moving to another one
	// puts it away rather than leaving an answer about one table under
	// another's definition.
	if !m.reviewIsAbout(row.key()) {
		m.closeSchemaReview()
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
// DESCRIBE KEYSPACE gives it, a table is its CREATE TABLE, and so on. A group
// is the names in it, fetched already: walking past one with the arrows
// should not fetch the definition of everything in it.
func (m *MainModel) describeSchemaRow(row schemaRow) string {
	if row.isGroup() {
		return m.describeSchemaGroup(row)
	}
	if !m.connected() {
		return "Not connected."
	}

	if row.kind != "" && row.kind != db.KindTables {
		text, err := m.session.DescribeSchemaObject(row.kind, row.keyspace, row.table)
		if err != nil {
			logger.DebugfToFile("Schema", "Describing %s: %v", row.key(), err)
			return fmt.Sprintf("Cannot describe %s.%s: %v", row.keyspace, row.table, err)
		}
		return text
	}

	if row.isKeyspace() {
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

	// A virtual table is not in system_schema, which is where the definition
	// below is built from - only the server can describe it.
	if m.session.IsVirtualKeyspace(row.keyspace) {
		text, err := m.session.DescribeVirtualTable(row.keyspace, row.table)
		if err != nil {
			logger.DebugfToFile("Schema", "Describing virtual table %s: %v", row.key(), err)
			return fmt.Sprintf("Cannot describe %s: %v", row.key(), err)
		}
		return text
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

// describeSchemaGroup is what a group holds, by name.
func (m *MainModel) describeSchemaGroup(row schemaRow) string {
	names := m.schema.names(row.keyspace, row.kind)
	label := strings.ToLower(schemaKindLabels[row.kind][0])
	lines := []string{fmt.Sprintf("%d %s in %s:", len(names), label, row.keyspace), ""}
	for _, name := range names {
		lines = append(lines, "    "+name)
	}
	return strings.Join(append(lines, "", "Select one to see its definition."), "\n")
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

// toggleSchemaRow opens or closes a keyspace or a group.
//
// An object has nothing to open, so pressing Enter on one shows it again -
// which is what it does anyway, and is better than doing nothing.
func (m *MainModel) toggleSchemaRow() (*MainModel, tea.Cmd) {
	row, ok := m.schema.current()
	if !ok {
		return m, nil
	}

	switch {
	case row.isGroup():
		m.schema.setGroupOpen(row, !m.schema.groupOpen(row.keyspace, row.kind))
	case !row.isKeyspace():
	case m.schema.expanded[row.keyspace]:
		m.schema.expanded[row.keyspace] = false
	default:
		m.openKeyspace(row.keyspace)
	}
	return m, nil
}

func (s *schemaBrowser) setGroupOpen(row schemaRow, open bool) {
	if s.groups == nil {
		s.groups = map[string]bool{}
	}
	s.groups[row.key()] = open
}

// expandSchemaRow opens a keyspace or a group, or steps into it when it is
// already open.
func (m *MainModel) expandSchemaRow() (*MainModel, tea.Cmd) {
	row, ok := m.schema.current()
	if !ok {
		return m, nil
	}

	switch {
	case row.isKeyspace() && !m.schema.expanded[row.keyspace]:
		return m.toggleSchemaRow()
	case row.isKeyspace() && len(m.schema.kindsOf(row.keyspace)) > 0:
		return m.moveSchemaSelection(1)
	case row.isGroup() && !m.schema.groupOpen(row.keyspace, row.kind):
		return m.toggleSchemaRow()
	case row.isGroup():
		return m.moveSchemaSelection(1)
	}
	return m, nil
}

// collapseSchemaRow closes a keyspace or a group, or goes up a level: from an
// object to its group, and from a closed group to its keyspace.
func (m *MainModel) collapseSchemaRow() (*MainModel, tea.Cmd) {
	row, ok := m.schema.current()
	if !ok {
		return m, nil
	}

	up := func(want schemaRow) (*MainModel, tea.Cmd) {
		for i, r := range m.schema.rows() {
			if r == want {
				return m.selectSchemaRow(i)
			}
		}
		return m, nil
	}

	switch {
	case row.isKeyspace():
		m.schema.expanded[row.keyspace] = false
		return m, nil
	case !row.isGroup():
		return up(schemaRow{keyspace: row.keyspace, kind: row.kind})
	case m.schema.groupOpen(row.keyspace, row.kind) && m.schema.filter == "":
		m.schema.setGroupOpen(row, false)
		return m, nil
	}
	return up(schemaRow{keyspace: row.keyspace})
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
	return cellWidth(line)
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
	// The filter, while it has the keys, takes what is typed. Everything it
	// does not take - the arrows, the page keys, the shell's own keys - goes on
	// to the tree and the shell as it would without it.
	if m.schema.filtering() {
		if updated, cmd, handled := m.schemaFilterKey(msg); handled {
			return updated, cmd, true
		}
	} else if msg.String() == "/" && strings.TrimSpace(m.input.Value()) == "" {
		// Only with an empty prompt, the rule Enter follows here too: a
		// statement being typed keeps its "/".
		updated, cmd := m.startSchemaFilter()
		return updated, cmd, true
	}

	height := m.schemaHeight()

	switch msg.String() {
	case "up":
		// The filter is the row above the tree, and up from the top row is
		// where it is looked for.
		if m.schema.selected == 0 {
			updated, cmd := m.startSchemaFilter()
			return updated, cmd, true
		}
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
		return m.scrollSchemaDetail(delta, g.detailRows)
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

// Filtering the tree.
//
// A cluster with a few dozen keyspaces, and forty-odd virtual tables on top
// since 4.0, is a long thing to walk with the arrows to find one table. The
// filter is the tree's heading row; "/", up from the top row, or a click there
// gives it the keys.

// startSchemaFilter gives the filter the keys, and the cursor: the prompt's
// stops blinking, so there is one place to type into on the screen.
//
// The tree fetches a keyspace's tables when it is first opened, which leaves
// most of them unknown - and a table that has not been fetched cannot be found
// by its name. So they are all fetched now, once; tablesOf keeps them.
func (m *MainModel) startSchemaFilter() (*MainModel, tea.Cmd) {
	m.loadAllSchemaNames()

	in := textinput.New()
	in.Prompt = schemaFilterPrefix
	in.CharLimit = 256
	heading := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	styles := in.Styles()
	styles.Focused.Prompt = heading
	styles.Focused.Text = heading
	in.SetStyles(styles)
	in.SetValue(m.schema.filter)
	in.CursorEnd()

	m.input.Blur()
	cmd := in.Focus()
	m.schema.filterInput = in
	return m, cmd
}

// stopSchemaFilter takes the keys from the filter, keeping what it found, and
// gives the prompt its cursor back.
func (m *MainModel) stopSchemaFilter() tea.Cmd {
	if !m.schema.filtering() {
		return nil
	}
	m.schema.filterInput.Blur()
	return m.input.Focus()
}

// settleSchemaFilter gives the prompt the keys back when the view the filter
// is in has gone, however it went - a tab key, a click on a tab, a command. A
// filter with the keys in a view that is not on screen would leave typing
// going nowhere that can be seen.
func (m *MainModel) settleSchemaFilter() tea.Cmd {
	if m.viewMode == "schema" {
		return nil
	}
	return m.stopSchemaFilter()
}

// schemaFilterEditKeys are the keys the filter's field takes besides what is
// typed: the ones that edit a line, as they do at the prompt.
var schemaFilterEditKeys = map[string]bool{
	"backspace": true, "delete": true, "left": true, "right": true,
	"home": true, "end": true, "ctrl+a": true, "ctrl+e": true,
	"ctrl+u": true, "ctrl+k": true, "ctrl+w": true, "ctrl+h": true,
	"alt+backspace": true, "ctrl+left": true, "ctrl+right": true,
}

// schemaFilterKey is what the filter does with a key, and whether it took it.
func (m *MainModel) schemaFilterKey(msg tea.KeyPressMsg) (*MainModel, tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		cmd := m.stopSchemaFilter()
		updated, selectCmd := m.clearSchemaFilter()
		return updated, tea.Batch(cmd, selectCmd), true
	case "enter", "down":
		// Out of the field, keeping what it found: down into the tree, at the
		// match already selected, and the prompt has the keys again.
		return m, m.stopSchemaFilter(), true
	case "up":
		// Nothing is above the filter.
		return m, nil, true
	}

	// What is typed and the keys that edit it, and nothing else. A shell key -
	// F2, Alt+F, Ctrl+Q - is not text, and has to reach the shell from here as
	// it does from anywhere.
	typed := msg.Text != "" && !isShellKey(msg.String()) && msg.Mod&(tea.ModCtrl|tea.ModAlt) == 0
	if !typed && !schemaFilterEditKeys[msg.String()] {
		return m, nil, false
	}

	var cmd tea.Cmd
	m.schema.filterInput, cmd = m.schema.filterInput.Update(msg)
	if m.schema.filterInput.Value() == m.schema.filter {
		return m, cmd, true // the cursor moved, and nothing else
	}
	updated, selectCmd := m.setSchemaFilter(m.schema.filterInput.Value())
	return updated, tea.Batch(cmd, selectCmd), true
}

// setSchemaFilter narrows the tree, and moves the selection to the first row
// that is what was searched for - so the definition beside it is the answer.
func (m *MainModel) setSchemaFilter(filter string) (*MainModel, tea.Cmd) {
	m.schema.filter = filter
	m.schema.scroll = 0
	if filter == "" {
		return m.selectSchemaRow(0)
	}
	return m.selectSchemaRow(m.schema.firstMatch())
}

// clearSchemaFilter puts the whole tree back, with what was found still
// selected: a table found by its name is opened in the full tree rather than
// lost inside a keyspace that is closed.
func (m *MainModel) clearSchemaFilter() (*MainModel, tea.Cmd) {
	found, ok := m.schema.current()

	m.schema.filter = ""
	m.schema.scroll = 0

	if !ok {
		return m.selectSchemaRow(0)
	}
	if !found.isKeyspace() {
		m.schema.expanded[found.keyspace] = true
	}
	if found.table != "" {
		m.schema.setGroupOpen(schemaRow{keyspace: found.keyspace, kind: found.kind}, true)
	}
	for i, row := range m.schema.rows() {
		if row == found {
			return m.selectSchemaRow(i)
		}
	}
	return m.selectSchemaRow(0)
}
