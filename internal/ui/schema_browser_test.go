package ui

import (
	"strings"
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectedSession is a session that passes the connected check without being
// able to answer anything.
//
// Enough for the tab to be available and the view to be reachable; a test that
// wants a tree gives the browser one directly, since what is being tested is
// the tree and the panes rather than the trip to Cassandra.
func connectedSession() *db.Session {
	return &db.Session{Session: &gocql.Session{}}
}

// schemaModel is the SCHEMA view on a cluster of three keyspaces, with the
// definitions already in hand: what is being tested is the tree and the panes,
// not the trip to Cassandra.
func schemaModel(t *testing.T) *MainModel {
	t.Helper()

	m := &MainModel{
		styles:          DefaultStyles(),
		viewMode:        "schema",
		windowWidth:     100,
		windowHeight:    24,
		input:           newTestInput(),
		historyViewport: viewport.New(viewport.WithWidth(100), viewport.WithHeight(8)),
	}
	m.input.Focus() // as it is for as long as cqlai is running
	m.schema = schemaBrowser{
		loaded:    true,
		keyspaces: []string{"my_keyspace", "system", "system_schema"},
		tables:    map[string][]string{"my_keyspace": {"events", "users"}},
		expanded:  map[string]bool{},
		definitions: map[string]string{
			"my_keyspace": "CREATE KEYSPACE my_keyspace WITH replication = {...};",
			// As FormatTableCreateStatement gives it with no header line: the
			// pane's heading already says which table this is.
			"my_keyspace.users": "CREATE TABLE my_keyspace.users (\n    id uuid PRIMARY KEY\n);",
		},
	}
	m.showSchemaDetail()
	return m
}

func treeRows(m *MainModel) []string {
	rows := make([]string, 0, len(m.schema.rows()))
	for _, row := range m.schema.rows() {
		name := strings.TrimSpace(m.schemaLine(row))
		name = strings.TrimPrefix(strings.TrimPrefix(name, "▾ "), "▸ ")
		rows = append(rows, name)
	}
	return rows
}

// TestTheTreeIsKeyspacesUntilOneIsOpened.
func TestTheTreeIsKeyspacesUntilOneIsOpened(t *testing.T) {
	m := schemaModel(t)
	assert.Equal(t, []string{"my_keyspace", "system", "system_schema"}, treeRows(m))

	m, _ = m.toggleSchemaRow()
	assert.Equal(t, []string{"my_keyspace", "events", "users", "system", "system_schema"}, treeRows(m))

	m, _ = m.toggleSchemaRow()
	assert.Equal(t, []string{"my_keyspace", "system", "system_schema"}, treeRows(m),
		"the keyspace should have folded shut again")
}

// TestTheMarkerSaysWhichWayAKeyspaceIsFolded.
func TestTheMarkerSaysWhichWayAKeyspaceIsFolded(t *testing.T) {
	m := schemaModel(t)
	row := m.schema.rows()[0]

	assert.Equal(t, "▸ my_keyspace", m.schemaLine(row))
	m, _ = m.toggleSchemaRow()
	assert.Equal(t, "▾ my_keyspace", m.schemaLine(row))
}

// TestSelectingARowShowsItsSchema, which is what the right-hand pane is for.
func TestSelectingARowShowsItsSchema(t *testing.T) {
	m := schemaModel(t)
	assert.Contains(t, strings.Join(m.schema.detail, "\n"), "CREATE KEYSPACE my_keyspace")

	m, _ = m.toggleSchemaRow()
	m, _ = m.selectSchemaRow(2) // my_keyspace.users
	assert.Contains(t, strings.Join(m.schema.detail, "\n"), "CREATE TABLE my_keyspace.users")
}

// TestClickingARowSelectsIt, and clicking the selected keyspace folds it.
//
// A first click says which one you mean and shows it; the second folds it.
// Folding on the first would make the schema, which is what the click was for,
// flash past.
func TestClickingARowSelectsIt(t *testing.T) {
	m := schemaModel(t)
	m, _ = m.toggleSchemaRow() // my_keyspace open: keyspace, events, users, ...

	// The row a click lands on is the row drawn there, below the heading.
	for i, row := range m.schema.rows() {
		got, hit := m.schemaRowAt(1, tabBarHeight+schemaHeaderRows+i)
		require.True(t, hit, "nothing at the row %s is drawn on", row.key())
		assert.Equal(t, i, got, "clicking %s landed on %s", row.key(), m.schema.rows()[got].key())
	}

	m, _ = m.clickSchema(1, tabBarHeight+schemaHeaderRows+2) // users
	assert.Equal(t, 2, m.schema.selected)
	assert.Contains(t, strings.Join(m.schema.detail, "\n"), "CREATE TABLE my_keyspace.users")

	// Clicking the open keyspace again folds it.
	m, _ = m.clickSchema(1, tabBarHeight+schemaHeaderRows)
	require.Equal(t, 0, m.schema.selected)
	m, _ = m.clickSchema(1, tabBarHeight+schemaHeaderRows)
	assert.False(t, m.schema.expanded["my_keyspace"])
}

// TestAPressOnTheDefinitionIsNotAClickOnTheTree: the right-hand pane is text,
// and a press there starts a selection so it can be copied out.
func TestAPressOnTheDefinitionIsNotAClickOnTheTree(t *testing.T) {
	m := schemaModel(t)
	m, _ = m.toggleSchemaRow()

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	_, hit := m.schemaRowAt(g.treeWidth+4, tabBarHeight+schemaHeaderRows)
	assert.False(t, hit)
	assert.True(t, m.inSchemaDetail(g.treeWidth+4, tabBarHeight+schemaHeaderRows))

	// And the heading itself is not a row of the tree.
	_, hit = m.schemaRowAt(1, tabBarHeight)
	assert.False(t, hit, "the heading is not a keyspace")
}

// TestTheArrowsWalkTheTree: down and up move, right opens, left closes.
func TestTheArrowsWalkTheTree(t *testing.T) {
	m := schemaModel(t)

	m, _, _ = m.schemaKey(tea.KeyPressMsg{Code: tea.KeyRight})
	assert.True(t, m.schema.expanded["my_keyspace"], "right should have opened it")

	m, _, _ = m.schemaKey(tea.KeyPressMsg{Code: tea.KeyDown})
	row, ok := m.schema.current()
	require.True(t, ok)
	assert.Equal(t, "events", row.table)

	// Left from a table goes up to the keyspace holding it; left again closes.
	m, _, _ = m.schemaKey(tea.KeyPressMsg{Code: tea.KeyLeft})
	row, _ = m.schema.current()
	assert.Equal(t, "", row.table)
	m, _, _ = m.schemaKey(tea.KeyPressMsg{Code: tea.KeyLeft})
	assert.False(t, m.schema.expanded["my_keyspace"])

	// And it stops at the ends rather than wrapping.
	m, _ = m.selectSchemaRow(0)
	m, _, _ = m.schemaKey(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, m.schema.selected)
	m, _ = m.selectSchemaRow(len(m.schema.rows()) - 1)
	m, _, _ = m.schemaKey(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, len(m.schema.rows())-1, m.schema.selected)
}

// TestTypingStillGoesToThePrompt. The view takes the keys that move around it
// and no others: a query can be typed while looking at the table it is about.
func TestTypingStillGoesToThePrompt(t *testing.T) {
	m := schemaModel(t)

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: 's', Text: "s"})
	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: 'e', Text: "e"})

	assert.Equal(t, "se", m.input.Value())
	assert.Equal(t, "schema", m.viewMode, "typing should not have left the view")
}

// TestEnterRunsWhatIsTypedRatherThanFoldingTheTree.
func TestEnterRunsWhatIsTypedRatherThanFoldingTheTree(t *testing.T) {
	m := schemaModel(t)
	m.input.SetValue("SELECT * FROM users")

	_, _, handled := m.schemaKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.False(t, handled, "Enter with a statement typed belongs to the prompt")

	m.input.SetValue("")
	_, _, handled = m.schemaKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.True(t, handled)
}

// TestTheSelectionStaysOnScreen as it moves down a tree longer than the view.
func TestTheSelectionStaysOnScreen(t *testing.T) {
	m := schemaModel(t)
	m.schema.keyspaces = nil
	for i := range 40 {
		m.schema.keyspaces = append(m.schema.keyspaces, "keyspace_"+string(rune('a'+i%26))+string(rune('0'+i/26)))
	}

	for i := range 40 {
		m, _ = m.selectSchemaRow(i)
		g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
		assert.GreaterOrEqual(t, m.schema.selected, g.treeFirst, "row %d scrolled off the top", i)
		assert.Less(t, m.schema.selected, g.treeFirst+g.height, "row %d scrolled off the bottom", i)
	}
}

// TestTheViewIsAsWideAndTallAsTheScreen, with the panes side by side.
func TestTheViewIsAsWideAndTallAsTheScreen(t *testing.T) {
	m := schemaModel(t)
	m, _ = m.toggleSchemaRow()

	for _, size := range [][2]int{{100, 16}, {80, 8}, {200, 40}, {40, 6}} {
		drawn := m.viewSchema(size[0], size[1])
		lines := strings.Split(drawn, "\n")
		assert.Len(t, lines, size[1], "%v: wrong number of rows", size)
		for _, line := range lines {
			assert.Equal(t, size[0], lipglossWidthOf(line), "%v: ragged row %q", size, stripAnsi(line))
		}
	}
}

// TestTheDefinitionScrollsOnItsOwn, since it is longer than the tree beside it.
func TestTheDefinitionScrollsOnItsOwn(t *testing.T) {
	m := schemaModel(t)
	m.schema.detail = strings.Split(strings.Repeat("line\n", 40), "\n")
	height := m.schemaHeight()

	m, _ = m.scrollSchemaDetail(5, height)
	assert.Equal(t, 5, m.schema.detailScroll)
	assert.Equal(t, 0, m.schema.selected, "the tree should not have moved")

	for range 100 {
		m, _ = m.scrollSchemaDetail(5, height)
	}
	assert.Equal(t, len(m.schema.detail)-height, m.schema.detailScroll, "it should stop at the end")

	for range 100 {
		m, _ = m.scrollSchemaDetail(-5, height)
	}
	assert.Zero(t, m.schema.detailScroll)
}

// TestTheWheelMovesThePaneItIsOver.
func TestTheWheelMovesThePaneItIsOver(t *testing.T) {
	m := schemaModel(t)
	m, _ = m.toggleSchemaRow() // five rows, so there is somewhere to scroll to
	m, _ = m.selectSchemaRow(0)
	m.schema.detail = strings.Split(strings.Repeat("line\n", 40), "\n")
	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())

	m, _ = m.scrollSchema(g.treeWidth+4, wheelLines)
	assert.Equal(t, wheelLines, m.schema.detailScroll)
	assert.Equal(t, 0, m.schema.selected)

	m, _ = m.scrollSchema(1, wheelLines)
	assert.Equal(t, wheelLines, m.schema.selected, "over the tree it moves the selection")
}

// TestAKeyspaceOpensOnTheOneInUse, which is the one the session is about.
func TestNothingConnectedSaysSo(t *testing.T) {
	m := schemaModel(t)
	m.schema = schemaBrowser{}
	m.loadSchema()

	assert.Equal(t, "Not connected.", m.schema.message)
	assert.Empty(t, m.schema.rows())

	// And the view still draws, rather than leaving an empty screen.
	assert.Contains(t, stripAnsi(m.viewSchema(m.windowWidth, 8)), "Not connected.")
}

// TestEachPaneSaysWhatItIs: a column of names with nothing over it, beside a
// block of text with nothing over that, leaves both to be worked out.
func TestEachPaneSaysWhatItIs(t *testing.T) {
	m := schemaModel(t)

	drawn := strings.Split(stripAnsi(m.viewSchema(m.windowWidth, 8)), "\n")
	require.GreaterOrEqual(t, len(drawn), 2)
	assert.Contains(t, drawn[0], schemaTreeHeading)
	assert.Contains(t, drawn[1], "───", "a rule under the headings")

	// The right-hand heading says what is being shown, not just that something
	// is: which keyspace, or which table.
	assert.Contains(t, drawn[0], "KEYSPACE  my_keyspace")

	m, _ = m.toggleSchemaRow()
	m, _ = m.selectSchemaRow(2)
	drawn = strings.Split(stripAnsi(m.viewSchema(m.windowWidth, 8)), "\n")
	assert.Contains(t, drawn[0], "TABLE  my_keyspace.users")

	// And the definition does not repeat it: the pane says which table this is,
	// so the CREATE statement starts on the first row rather than the third.
	assert.Contains(t, drawn[2], "CREATE TABLE")
}

// TestTheHeadingsStayWhileTheTreeScrolls: they are the two rows that say what
// each side is, so they cannot be the first thing to scroll away.
func TestTheHeadingsStayWhileTheTreeScrolls(t *testing.T) {
	m := schemaModel(t)
	m.schema.keyspaces = nil
	for i := range 40 {
		m.schema.keyspaces = append(m.schema.keyspaces, "keyspace_"+string(rune('a'+i%26)))
	}
	m, _ = m.selectSchemaRow(39)

	drawn := strings.Split(stripAnsi(m.viewSchema(m.windowWidth, 8)), "\n")
	assert.Contains(t, drawn[0], schemaTreeHeading)
	assert.Contains(t, drawn[1], "───")
}

// TestSchemaSitsNextToTheConsole: what is in the database comes before anything
// a query has made of it, and the three that follow are all views of a result.
func TestSchemaSitsNextToTheConsole(t *testing.T) {
	order := make([]string, 0, len(modeTabs))
	for _, tab := range modeTabs {
		order = append(order, tab.mode)
	}
	assert.Equal(t, []string{"history", "schema", "table", "trace", "ai"}, order)

	// And the keys read along the line with them: a bar whose keys are out of
	// order is one you have to read rather than count along.
	keys := make([]string, 0, len(modeTabs))
	for _, tab := range modeTabs {
		keys = append(keys, tab.key)
	}
	assert.Equal(t, []string{"F2", "F3", "F4", "F5", "F6"}, keys)
}

// TestTheCompletionListKeepsTheArrows.
//
// The tree is the background of this view, and the completion list is drawn
// over it. Taking the arrows unconditionally moved the tree behind while the
// list you were reading stood still.
func TestTheCompletionListKeepsTheArrows(t *testing.T) {
	m := schemaModel(t)
	m.input.SetValue("SELECT * FROM ")
	m.showCompletions = true
	m.completions = []string{"events", "users"}
	m.completionIndex = 0

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyDown})

	assert.Equal(t, 1, m.completionIndex, "down should have moved the list")
	assert.Equal(t, 0, m.schema.selected, "the tree behind it should not have moved")

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, m.completionIndex)
	assert.Equal(t, 0, m.schema.selected)

	// With the list gone, the arrows are the tree's again.
	m.showCompletions = false
	m.completions = nil
	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 1, m.schema.selected)
}

// TestAQuestionOnScreenKeepsTheArrows, for the same reason: left and right
// choose the answer, and Enter gives it.
func TestAQuestionOnScreenKeepsTheArrows(t *testing.T) {
	m := schemaModel(t)
	m.modal = NewConfirmationModal("DROP TABLE users")

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyRight})

	assert.Equal(t, 1, m.modal.Selected, "right should have moved to the other answer")
	assert.Equal(t, 0, m.schema.selected)
	assert.False(t, m.schema.expanded["my_keyspace"], "the tree should not have opened")
}

// TestTheHistoryListKeepsTheArrows.
func TestTheHistoryListKeepsTheArrows(t *testing.T) {
	m := schemaModel(t)
	m.commandHistory = []string{"SELECT * FROM users", "DROP TABLE events"}
	m, _ = m.handleCtrlR()
	require.True(t, m.historySearchMode)

	before := m.historySearchIndex
	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyUp})

	assert.NotEqual(t, before, m.historySearchIndex, "up should have moved the list")
	assert.Equal(t, 0, m.schema.selected)
}

// TestTheDefinitionCanBeCopiedOut.
//
// The point of having the schema on screen is to take things out of it - a
// column name, a whole CREATE TABLE - into the prompt. Selection was anchored to
// a viewport, and this view has none: a drag over the definition recorded a span
// against the console's content, so what reached the clipboard was whatever
// happened to be on those rows of a view you were not looking at.
func TestTheDefinitionCanBeCopiedOut(t *testing.T) {
	m := schemaModel(t)
	m, _ = m.toggleSchemaRow()
	m, _ = m.selectSchemaRow(2) // my_keyspace.users

	// What is on screen is what a selection is over, so it has to be drawn
	// before it can be dragged across.
	m.historyViewport.SetContent(strings.Repeat("console line\n", 40))
	drawn := strings.Split(m.viewSchema(m.windowWidth, m.schemaHeight()), "\n")

	// The row holding the CREATE TABLE line, and the columns it covers.
	row, text := -1, ""
	for i, line := range drawn {
		if strings.Contains(stripAnsi(line), "CREATE TABLE") {
			row, text = i, stripAnsi(line)
			break
		}
	}
	require.NotEqual(t, -1, row, "the definition should be on screen")

	// In columns, not bytes: the line has a fold marker and a divider in it, and
	// both are three bytes wide and one column wide.
	from := columnOf(text, "CREATE TABLE")
	to := from + len([]rune("CREATE TABLE my_keyspace.users ("))

	cmd := drag(m, from, tabBarHeight+row, to, tabBarHeight+row)
	assert.Equal(t, "CREATE TABLE my_keyspace.users (", clipboardText(t, cmd))
	assert.NotContains(t, clipboardText(t, cmd), "console line",
		"the selection should be over this view, not the one behind it")
}

// TestATreeNameCanBeCopiedOut too: the left-hand pane is text on screen like
// any other.
func TestATreeNameCanBeCopiedOut(t *testing.T) {
	m := schemaModel(t)
	m, _ = m.toggleSchemaRow()
	drawn := strings.Split(m.viewSchema(m.windowWidth, m.schemaHeight()), "\n")

	// Only the tree column: the definition beside it names the keyspace too.
	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	row := -1
	for i, line := range drawn {
		pane := string([]rune(stripAnsi(line))[:g.treeWidth])
		if strings.Contains(pane, "my_keyspace") {
			row = i
			break
		}
	}
	require.NotEqual(t, -1, row)

	from := columnOf(stripAnsi(drawn[row]), "my_keyspace")
	cmd := drag(m, from, tabBarHeight+row, from+len("my_keyspace"), tabBarHeight+row)
	assert.Equal(t, "my_keyspace", clipboardText(t, cmd))
}

// TestADragInTheTreeIsNotAClick: a press on the tree selects a row, and one on
// the definition starts a selection, so both have to be tried.
func TestADragInTheTreeIsNotAClick(t *testing.T) {
	m := schemaModel(t)
	m, _ = m.toggleSchemaRow()
	m.viewSchema(m.windowWidth, m.schemaHeight())

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	updated, _ := m.handleMousePress(tea.Mouse{
		X: g.treeWidth + 6, Y: tabBarHeight + schemaHeaderRows, Button: tea.MouseLeft,
	})
	assert.True(t, updated.selection.dragging, "a press on the definition starts a selection")
	assert.Equal(t, 0, updated.schema.selected, "and does not move the tree")
}

// columnOf is where a piece of text starts, counted in columns on screen rather
// than in bytes: a fold marker and a divider are three bytes each and one column
// each, and a selection is made of columns.
func columnOf(line, text string) int {
	i := strings.Index(line, text)
	if i < 0 {
		return -1
	}
	return len([]rune(line[:i]))
}

// TestADDLStatementDropsWhatTheBrowserKnows.
//
// What it keeps describes the cluster as it was: after a CREATE TABLE the tree
// and the definition beside it were still the schema from before the statement,
// and only F3 would say otherwise.
func TestADDLStatementDropsWhatTheBrowserKnows(t *testing.T) {
	m := schemaModel(t)
	m.viewMode = "history" // looking at the console, as you are when you type
	require.NotEmpty(t, m.schema.definitions)

	m.schemaChanged()

	assert.True(t, m.schema.stale, "it should know it is out of date")
	assert.NotEmpty(t, m.schema.definitions,
		"and not have fetched anything yet: the tab is not open")

	// Opening it fetches again. There is no cluster here, so what it finds is
	// nothing - the point is that it went and looked.
	m, _ = m.openSchema()
	assert.False(t, m.schema.stale)
	assert.Empty(t, m.schema.definitions, "the definitions should have been dropped")
}

// TestAChangeWhileLookingAtTheTreeShowsAtOnce.
func TestAChangeWhileLookingAtTheTreeShowsAtOnce(t *testing.T) {
	m := schemaModel(t)
	require.Equal(t, "schema", m.viewMode)

	m.schemaChanged()

	assert.Empty(t, m.schema.definitions, "it should have fetched again there and then")
	assert.False(t, m.schema.stale)
}

// TestNothingFetchedIsNothingToDropStale.
func TestNothingFetchedIsNothingToDropStale(t *testing.T) {
	m := schemaModel(t)
	m.schema = schemaBrowser{} // never opened

	m.schemaChanged()

	assert.False(t, m.schema.stale, "there is nothing out of date yet")
}

// TestRefreshingKeepsWhereYouWere: the keyspaces left open stay open, and the
// selection stays where it was, as far as the tree still goes.
func TestRefreshingKeepsWhereYouWere(t *testing.T) {
	m := schemaModel(t)
	m, _ = m.toggleSchemaRow() // my_keyspace open
	m, _ = m.selectSchemaRow(2)
	require.True(t, m.schema.expanded["my_keyspace"])

	// A cluster to find on the way back, so the tree is not empty.
	m.schema.keyspaces = []string{"my_keyspace", "system"}
	m, _ = m.refreshSchema()

	assert.True(t, m.schema.expanded["my_keyspace"], "it should still be open")
	assert.LessOrEqual(t, m.schema.selected, max(len(m.schema.rows())-1, 0))
}
