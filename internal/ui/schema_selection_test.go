package ui

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Selecting out of the schema browser.
//
// The view is two panes side by side, so a line of it is "tree │ definition".
// The rule for a flowing selection gives the first line from its start column,
// the last line up to its end column, and every line between whole - which in
// this view meant dragging down the definition copied the tree with it.
//
// A selection is bounded to the pane the drag began in. The copy and the
// highlight both ask one function which columns of a line it covers, so what
// is painted is what is copied.

// selectableSchema is the browser with a table selected, drawn, and ready to
// have text dragged out of it.
func selectableSchema(t *testing.T) *MainModel {
	t.Helper()

	m := schemaModel(t)
	m.schema.expanded["my_keyspace"] = true
	m, _ = m.selectSchemaRow(rowOfTable(t, m, "users"))

	// Drawing is what the selection is over: this view has no viewport.
	m.viewSchema(m.windowWidth, m.schemaHeight())
	require.NotEmpty(t, m.schema.drawn)
	return m
}

// rowOfTable is the tree row a table sits on.
func rowOfTable(t *testing.T, m *MainModel, table string) int {
	t.Helper()

	for i, row := range m.schema.rows() {
		if row.table == table {
			return i
		}
	}
	t.Fatalf("no row for %s", table)
	return -1
}

// definitionColumn is the first column of the definition pane: the tree, then
// the three columns of the divider.
func definitionColumn(m *MainModel) int {
	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	return g.treeWidth + lipgloss.Width(schemaDivider)
}

// TestDraggingDownTheDefinitionDoesNotTakeTheTree.
//
// The bug: selecting the definition to copy it also copied the keyspace and
// table names from the pane beside it.
func TestDraggingDownTheDefinitionDoesNotTakeTheTree(t *testing.T) {
	m := selectableSchema(t)

	// From the first line of the definition to well down it.
	col := definitionColumn(m)
	m, _ = m.beginSelection(col, m.viewTop()+schemaHeaderRows)
	m, _ = m.extendSelection(m.windowWidth-1, m.viewTop()+schemaHeaderRows+2)

	copied := m.selectedText()
	require.NotEmpty(t, copied)

	assert.True(t, strings.HasPrefix(copied, "CREATE TABLE"),
		"the definition is what was dragged over, from its first column: %q", copied)

	// Names that are only in the tree. "my_keyspace" is not one of them - the
	// definition says CREATE TABLE my_keyspace.users - which is why the test
	// has to name the rows rather than the keyspace.
	for _, fromTheTree := range []string{"system_schema", "events", "KEYSPACES"} {
		assert.NotContains(t, copied, fromTheTree,
			"%q is in the tree, not the definition", fromTheTree)
	}
	assert.NotContains(t, copied, "│", "nor the divider between them")
}

// TestDraggingDownTheTreeDoesNotTakeTheDefinition, which is the same bug from
// the other side.
func TestDraggingDownTheTreeDoesNotTakeTheDefinition(t *testing.T) {
	m := selectableSchema(t)

	m, _ = m.beginSelection(0, m.viewTop()+schemaHeaderRows)
	m, _ = m.extendSelection(m.windowWidth-1, m.viewTop()+schemaHeaderRows+2)

	copied := m.selectedText()
	require.NotEmpty(t, copied)

	assert.NotContains(t, copied, "CREATE", "the definition is in the other pane")
	assert.NotContains(t, copied, "│")
}

// TestTripleClickTakesTheLineOfOnePaneOnly.
func TestTripleClickTakesTheLineOfOnePaneOnly(t *testing.T) {
	m := selectableSchema(t)

	row := m.viewTop() + schemaHeaderRows
	col := definitionColumn(m)
	for range 3 {
		m, _ = m.beginSelection(col, row)
	}

	copied := m.selectedText()
	assert.Equal(t, "CREATE TABLE my_keyspace.users (", copied,
		"a line of the definition, not a line of the screen")
	assert.NotContains(t, copied, "│")
}

// TestTheHighlightCoversWhatIsCopied.
//
// The copy and the highlight worked the span out separately, which is how a
// selection comes to look like one thing and paste as another. One function
// answers both now.
func TestTheHighlightCoversWhatIsCopied(t *testing.T) {
	m := selectableSchema(t)

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	for _, where := range []struct {
		name        string
		col         int
		left, right int
	}{
		{"the definition", definitionColumn(m), g.treeWidth + 3, g.width},
		{"the tree", 0, 0, g.treeWidth - 1},
	} {
		m, _ = m.beginSelection(where.col, m.viewTop()+schemaHeaderRows)
		m, _ = m.extendSelection(m.windowWidth-1, m.viewTop()+schemaHeaderRows+2)

		assert.Equal(t, where.left, m.selection.left, "%s starts at", where.name)
		assert.Equal(t, where.right, m.selection.right, "%s ends at", where.name)

		// A line in the middle of the drag: whole, within the pane. The
		// definition's lines are its own, starting at column zero, so the
		// span is read back against the indent the source carries.
		source, _ := m.selectionTarget()
		from, to := m.selection.columnsOn(1, m.windowWidth-source.indent, source.indent)
		assert.Equal(t, max(where.left-source.indent, 0), from, "%s", where.name)
		assert.Equal(t, where.right-source.indent, to, "%s", where.name)
	}
}

// TestEveryOtherViewStillTakesWholeLines: only the schema browser draws two
// panes side by side, and the bound must not narrow anything else.
func TestEveryOtherViewStillTakesWholeLines(t *testing.T) {
	m := selectionModel(5, "first line", "second line", "third line")

	m, _ = m.beginSelection(6, 1)
	m, _ = m.extendSelection(5, 3)

	assert.Zero(t, m.selection.left, "no pane bound outside the schema browser")
	assert.Zero(t, m.selection.right)

	// The middle line whole, which is what a flowing selection does.
	assert.Contains(t, strings.Split(m.selectedText(), "\n"), "second line")
}

// TestDraggingOffTheBottomDoesNotCrash.
//
// The guard against a view with no viewport was on the top edge only:
//
//	case row < top:
//	    if vp == nil { ... break }
//	case row >= bottom:
//	    vp.SetYOffset(...)        // nil here
//
// So dragging off the bottom of the schema browser - the view most likely to
// hold more than fits, and the one someone is most likely to be copying out of
// - dereferenced nil and took cqlai down with it. One rule, two edges, written
// once.
func TestDraggingOffTheBottomDoesNotCrash(t *testing.T) {
	m := selectableSchema(t)

	m, _ = m.beginSelection(definitionColumn(m), m.viewTop()+schemaHeaderRows)

	assert.NotPanics(t, func() {
		m, _ = m.extendSelection(m.windowWidth-1, m.windowHeight+50)
	}, "dragging past the bottom of a view with no viewport")

	assert.NotPanics(t, func() {
		m, _ = m.extendSelection(0, -20)
	}, "and past the top, which was already guarded")
}

// TestDraggingOffEitherEdgeOfTheTraceDoesNotCrash: the trace draws as a block
// too, so it had the same nil waiting on the bottom edge.
func TestDraggingOffEitherEdgeOfTheTraceDoesNotCrash(t *testing.T) {
	m := traceModel(t)
	m.viewTrace(m.windowWidth, m.viewHeight())
	require.NotEmpty(t, m.trace.drawn)

	m, _ = m.beginSelection(2, m.viewTop()+1)

	assert.NotPanics(t, func() {
		m, _ = m.extendSelection(10, m.windowHeight+50)
	})
	assert.NotPanics(t, func() {
		m, _ = m.extendSelection(10, -20)
	})
}

// TestDraggingOffTheBottomOfTheConsoleStillScrolls: the fix must not cost the
// views that do have a viewport the scrolling they had.
func TestDraggingOffTheBottomOfTheConsoleStillScrolls(t *testing.T) {
	m := selectionModel(3, "one", "two", "three", "four", "five", "six")

	m, _ = m.beginSelection(0, 1)
	before := m.historyViewport.YOffset()

	m, _ = m.extendSelection(2, 99)

	assert.Greater(t, m.historyViewport.YOffset(), before,
		"dragging past the bottom should still scroll a viewport")
}

// TestADefinitionTallerThanThePaneCanBeSelectedWhole.
//
// What the bug report was actually about: dragging to the bottom of the pane
// and expecting it to keep going. It stopped, so you could only ever copy what
// happened to be on screen.
//
// The span is over the definition now, with the pane's scroll offset, the way
// a viewport-backed view already works - so the drag can scroll it and the
// span stays on the text it was over.
func TestADefinitionTallerThanThePaneCanBeSelectedWhole(t *testing.T) {
	m := selectableSchema(t)

	// A definition with more lines than the pane has rows.
	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	var lines []string
	for i := range g.detailRows * 3 {
		lines = append(lines, fmt.Sprintf("line_%02d int,", i))
	}
	m.schema.detail = lines
	m.schema.detailScroll = 0
	m.viewSchema(m.windowWidth, m.schemaHeight())

	top := m.viewTop() + schemaHeaderRows
	m, _ = m.beginSelection(definitionColumn(m), top)

	// Drag to the bottom and hold there, as a hand does.
	for range len(lines) {
		m, _ = m.extendSelection(m.windowWidth-1, m.windowHeight+10)
	}

	assert.Positive(t, m.schema.detailScroll, "the pane should have scrolled under the drag")

	copied := m.selectedText()
	assert.Contains(t, copied, "line_00", "the line the drag started on")
	assert.Contains(t, copied, lines[len(lines)-1], "and the last one, which was never on screen at the start")
	assert.NotContains(t, copied, "│", "still one pane only")
}

// TestTheTreeDoesNotScrollUnderADrag: it is a list of rows picked with the
// keyboard, not a document to drag through, and scrolling it would move the
// text out from under a span that is over what was drawn.
func TestTheTreeDoesNotScrollUnderADrag(t *testing.T) {
	m := selectableSchema(t)
	before := m.schema.scroll

	m, _ = m.beginSelection(0, m.viewTop()+schemaHeaderRows)
	for range 5 {
		m, _ = m.extendSelection(2, m.windowHeight+10)
	}

	assert.Equal(t, before, m.schema.scroll)
}
