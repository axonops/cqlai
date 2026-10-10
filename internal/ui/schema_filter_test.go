package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Filtering the schema tree.
//
// A cluster with a few dozen keyspaces, and forty-odd virtual tables since
// 4.0, is a long tree to walk with the arrows to find one table. The tree's
// heading row is a filter: "/" or a click on it gives it the keys, and the tree
// narrows to the keyspaces and tables whose names hold what is typed.

// filterModel is the SCHEMA view with every keyspace's tables already known, so
// starting the filter has nothing to fetch.
func filterModel(t *testing.T) *MainModel {
	t.Helper()
	m := schemaModel(t)
	m.schema.tables = map[string][]string{
		"my_keyspace": {"AuditLog", "events", "users"}, // a quoted, mixed-case name
		// Wider than the heading, so the pane's width is set by a table.
		"system":        {"local", "peers", "size_estimates_by_range_and_host"},
		"system_schema": {"keyspaces", "tables"},
	}
	return m
}

// typeKeys sends each rune as a key press, through the same dispatcher the
// terminal's keys go through.
func typeKeys(m *MainModel, text string) *MainModel {
	for _, r := range text {
		m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func pressKey(m *MainModel, code rune) *MainModel {
	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: code})
	return m
}

// TestATableIsFoundByItsName, under its keyspace, without opening it first.
func TestATableIsFoundByItsName(t *testing.T) {
	m := filterModel(t)

	m = typeKeys(m, "/use")

	assert.Equal(t, []string{"my_keyspace", "Tables (3)", "users"}, treeRows(m))
	assert.Equal(t, "", m.input.Value(), "what is typed goes to the filter, not the prompt")

	row, ok := m.schema.current()
	require.True(t, ok)
	assert.Equal(t, "users", row.table, "the selection goes to what was found")
	assert.Contains(t, strings.Join(m.schema.detail, "\n"), "CREATE TABLE my_keyspace.users")
}

// TestAKeyspaceIsFoundByItsName, and the tables found in others alongside it.
func TestAKeyspaceIsFoundByItsName(t *testing.T) {
	m := filterModel(t)

	m = typeKeys(m, "/SYSTEM_S") // and case does not matter

	assert.Equal(t, []string{"system_schema"}, treeRows(m))

	m = pressKey(m, tea.KeyBackspace)
	m = pressKey(m, tea.KeyBackspace)
	assert.Equal(t, []string{"system", "system_schema"}, treeRows(m))
}

// TestCaseDoesNotMatterInTheName either: a quoted table keeps its capitals.
func TestCaseDoesNotMatterInTheName(t *testing.T) {
	m := filterModel(t)

	m = typeKeys(m, "/auditlog")

	assert.Equal(t, []string{"my_keyspace", "Tables (3)", "AuditLog"}, treeRows(m))
}

// TestTheMatchesInsideAKeyspaceAreTheOnlyOnesShown: a table that does not
// match is not shown just because its neighbour does.
func TestTheMatchesInsideAKeyspaceAreTheOnlyOnesShown(t *testing.T) {
	m := filterModel(t)

	m = typeKeys(m, "/s")

	// my_keyspace matches by its name ("s"), and is closed; system matches
	// too. Every name here holds an s.
	assert.Contains(t, treeRows(m), "my_keyspace")
	assert.NotContains(t, treeRows(m), "local", "system is closed, and matched by its own name")

	m = pressKey(m, tea.KeyBackspace)
	m = typeKeys(m, "ers")
	assert.Equal(t, []string{"my_keyspace", "Tables (3)", "users", "system", "Tables (3)", "peers"}, treeRows(m))
}

// TestNothingMatchingSaysSo.
func TestNothingMatchingSaysSo(t *testing.T) {
	m := filterModel(t)

	m = typeKeys(m, "/zzz")

	assert.Empty(t, m.schema.rows())
	assert.Contains(t, m.viewSchema(m.windowWidth, m.schemaHeight()), schemaNoMatch)
}

// TestUpFromTheTopRowIsTheFilter: it is the row above the tree, and that is
// where up goes.
func TestUpFromTheTopRowIsTheFilter(t *testing.T) {
	m := filterModel(t)
	m = pressKey(m, tea.KeyDown)

	m = pressKey(m, tea.KeyUp)
	require.False(t, m.schema.filtering(), "up from the second row is the first row")
	assert.Equal(t, 0, m.schema.selected)

	m = pressKey(m, tea.KeyUp)
	require.True(t, m.schema.filtering(), "up from the top row is the filter")

	m = pressKey(m, tea.KeyUp)
	assert.True(t, m.schema.filtering(), "and nothing is above it")

	m = typeKeys(m, "peer")
	assert.Equal(t, []string{"system", "Tables (3)", "peers"}, treeRows(m))

	// With the match below the top row, up still stays in the filter
	// rather than walking the tree.
	m = pressKey(m, tea.KeyUp)
	assert.True(t, m.schema.filtering())
	row, _ := m.schema.current()
	assert.Equal(t, "peers", row.table)
}

// TestDownFromTheFilterIsTheTree, at the match already selected, with the tree
// left filtered and the prompt given the keys again.
func TestDownFromTheFilterIsTheTree(t *testing.T) {
	m := filterModel(t)
	m = typeKeys(m, "/ers")

	m = pressKey(m, tea.KeyDown)
	require.False(t, m.schema.filtering())
	row, _ := m.schema.current()
	assert.Equal(t, "users", row.table, "the first match, not the row after it")
	assert.Equal(t, []string{"my_keyspace", "Tables (3)", "users", "system", "Tables (3)", "peers"}, treeRows(m))

	m = pressKey(m, tea.KeyDown)
	row, _ = m.schema.current()
	assert.Equal(t, "system", row.keyspace, "and down walks the tree from there")
}

// TestEnterKeepsTheFilterAndGivesBackThePrompt.
func TestEnterKeepsTheFilterAndGivesBackThePrompt(t *testing.T) {
	m := filterModel(t)
	m = typeKeys(m, "/use")

	m = pressKey(m, tea.KeyEnter)
	assert.False(t, m.schema.filtering())
	assert.Equal(t, []string{"my_keyspace", "Tables (3)", "users"}, treeRows(m), "the tree stays narrowed")

	m = typeKeys(m, "se")
	assert.Equal(t, "se", m.input.Value(), "typing goes to the prompt again")
	assert.Equal(t, []string{"my_keyspace", "Tables (3)", "users"}, treeRows(m))
}

// TestEscapePutsTheWholeTreeBack, with what was found still selected.
func TestEscapePutsTheWholeTreeBack(t *testing.T) {
	m := filterModel(t)
	m = typeKeys(m, "/peer")

	m = pressKey(m, tea.KeyEscape)

	assert.False(t, m.schema.filtering())
	assert.Equal(t, "", m.schema.filter)
	assert.Contains(t, treeRows(m), "my_keyspace")
	row, ok := m.schema.current()
	require.True(t, ok)
	assert.Equal(t, "system.peers", row.key(), "found, and left selected")
	assert.True(t, m.schema.expanded["system"], "its keyspace is opened to show it")
}

// TestASlashInAStatementIsTyped: "/" only starts the filter with nothing at
// the prompt, the rule Enter follows in this view too.
func TestASlashInAStatementIsTyped(t *testing.T) {
	m := filterModel(t)
	m.input.SetValue("SELECT 4")

	m = typeKeys(m, "/")

	assert.False(t, m.schema.filtering())
	assert.Equal(t, "SELECT 4/", m.input.Value())
}

// TestTheShellKeysStillWorkWhileFiltering: F2 is not text, and leaves.
func TestTheShellKeysStillWorkWhileFiltering(t *testing.T) {
	m := filterModel(t)
	m = typeKeys(m, "/u")

	_, _, handled := m.schemaFilterKey(tea.KeyPressMsg{Code: tea.KeyF2})
	assert.False(t, handled, "F2 belongs to the shell")
	_, _, handled = m.schemaFilterKey(tea.KeyPressMsg{Code: 'f', Text: "f", Mod: tea.ModAlt})
	assert.False(t, handled, "Alt+F belongs to the shell")
}

// TestTheHeadingIsTheFilter: it says there is one, shows what is typed, and a
// click on it gives it the keys.
func TestTheHeadingIsTheFilter(t *testing.T) {
	m := filterModel(t)
	assert.Contains(t, m.viewSchema(m.windowWidth, m.schemaHeight()), "/ filter", "the heading says there is a filter")

	require.True(t, m.schemaTreeHeadingAt(1, tabBarHeight))
	assert.False(t, m.schemaTreeHeadingAt(1, tabBarHeight+schemaHeaderRows), "a row is not the heading")

	m, _ = m.handleMousePress(tea.Mouse{X: 1, Y: tabBarHeight, Button: tea.MouseLeft})
	require.True(t, m.schema.filtering())

	m = typeKeys(m, "even")
	assert.Contains(t, ansi.Strip(m.viewSchema(m.windowWidth, m.schemaHeight())), "/ even")
}

// TestThePaneKeepsItsWidthWhileTyping: the tree would otherwise jump about
// with each letter.
func TestThePaneKeepsItsWidthWhileTyping(t *testing.T) {
	m := filterModel(t)
	m = typeKeys(m, "/size") // the widest table in the tree, on show
	width := m.schemaTreeWidth()

	m = typeKeys(m, "x") // and gone
	require.Empty(t, m.schema.rows())
	assert.Equal(t, width, m.schemaTreeWidth())
}

// TestTheFilterHasTheOnlyCursor: it blinks like the prompt's, and the prompt's
// goes while the filter has the keys, so there is one place to type into.
func TestTheFilterHasTheOnlyCursor(t *testing.T) {
	m := filterModel(t)

	m, cmd := m.startSchemaFilter()
	assert.NotNil(t, cmd, "focusing the field starts its cursor blinking")
	assert.False(t, m.input.Focused(), "the prompt's cursor goes")

	m = pressKey(m, tea.KeyEnter)
	assert.True(t, m.input.Focused(), "and comes back when the filter lets go")
}

// TestLeavingTheViewGivesThePromptTheKeys, however it is left: a filter with
// the keys in a view that is not on screen would leave typing going nowhere.
func TestLeavingTheViewGivesThePromptTheKeys(t *testing.T) {
	m := filterModel(t)
	m, _ = m.startSchemaFilter()

	m.viewMode = "history" // a tab key, a click on a tab, a command
	_, _ = m.Update(struct{}{})

	assert.False(t, m.schema.filtering())
	assert.True(t, m.input.Focused())
}
