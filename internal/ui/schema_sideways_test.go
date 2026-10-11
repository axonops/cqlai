package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Scrolling the definition sideways. A line wider than the pane was cut off at
// its edge, and nothing showed the rest of it.

// wideLine is a definition line far wider than the pane, with its end marked.
var wideLine = "    id uuid PRIMARY KEY" + strings.Repeat(" -", 80) + " THE_END"

// wideSchema is the browser on a table whose definition has a wide line.
func wideSchema(t *testing.T) *MainModel {
	t.Helper()
	m := schemaModel(t)
	m.schema.definitions["my_keyspace.users"] = "CREATE TABLE my_keyspace.users (\n" + wideLine + "\n);"
	m.schema.expanded["my_keyspace"] = true
	m, _ = m.selectSchemaRow(rowOfTable(t, m, "users"))
	require.Greater(t, m.schema.detailWidth, m.schemaGeometry(m.windowWidth, m.schemaHeight()).detailWidth)
	return m
}

// drawnDefinition is the definition pane as drawn, a line per row.
func drawnDefinition(m *MainModel) string {
	from := definitionColumn(m)
	var lines []string
	for _, line := range strings.Split(ansi.Strip(m.viewSchema(m.windowWidth, m.schemaHeight())), "\n") {
		lines = append(lines, ansi.Cut(line, from, m.windowWidth))
	}
	return strings.Join(lines, "\n")
}

// TestAltRightShowsTheRestOfAWideLine, and Alt+Left goes back.
func TestAltRightShowsTheRestOfAWideLine(t *testing.T) {
	m := wideSchema(t)
	require.NotContains(t, drawnDefinition(m), "THE_END")

	for range 30 {
		m.handleKeyboardInput(altKey(tea.KeyRight))
	}
	assert.Contains(t, drawnDefinition(m), "THE_END")
	assert.NotContains(t, drawnDefinition(m), "PRIMARY KEY", "the start has scrolled off")

	for range 30 {
		m.handleKeyboardInput(altKey(tea.KeyLeft))
	}
	assert.Equal(t, 0, m.schema.detailLeft)
	assert.Contains(t, drawnDefinition(m), "PRIMARY KEY")
}

// TestAPlainArrowStillWorksTheTree: Left and Right open and close keyspaces,
// and scroll nothing.
func TestAPlainArrowStillWorksTheTree(t *testing.T) {
	m := wideSchema(t)

	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyRight})

	assert.Equal(t, 0, m.schema.detailLeft)
}

// TestTheWheelScrollsTheDefinitionSideways: a trackpad's own sideways
// scroll, or the wheel with Shift held.
func TestTheWheelScrollsTheDefinitionSideways(t *testing.T) {
	col, row := definitionColumn(wideSchema(t))+2, tabBarHeight+schemaHeaderRows+1

	m := wideSchema(t)
	m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelRight, X: col, Y: row})
	assert.Equal(t, horizontalStep, m.schema.detailLeft, "sideways wheel")
	m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelLeft, X: col, Y: row})
	assert.Equal(t, 0, m.schema.detailLeft)

	m = wideSchema(t)
	selected := m.schema.selected
	m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelDown, Mod: tea.ModShift, X: col, Y: row})
	assert.Equal(t, horizontalStep, m.schema.detailLeft, "Shift and the wheel")
	assert.Equal(t, selected, m.schema.selected, "and the tree did not move")
	m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelUp, Mod: tea.ModShift, X: col, Y: row})
	assert.Equal(t, 0, m.schema.detailLeft)
}

// TestItStopsAtTheEndOfTheWidestLine, rather than scrolling into nothing.
func TestItStopsAtTheEndOfTheWidestLine(t *testing.T) {
	m := wideSchema(t)
	for range 100 {
		m.scrollHorizontally(horizontalStep)
	}
	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	assert.Equal(t, clampOffset(1<<20, m.schema.detailWidth, g.detailWidth-1), m.schema.detailLeft)
	assert.Contains(t, drawnDefinition(m), "THE_END")
}

// TestAnotherObjectStartsAtTheLeft: its lines have nothing to do with how far
// the last one was scrolled.
func TestAnotherObjectStartsAtTheLeft(t *testing.T) {
	m := wideSchema(t)
	m.scrollHorizontally(horizontalStep)
	require.Positive(t, m.schema.detailLeft)

	m, _ = m.moveSchemaSelection(-1)

	assert.Equal(t, 0, m.schema.detailLeft)
}

// TestASelectionWhileScrolledCopiesWhatIsUnderThePointer: the columns on
// screen are that far along the line.
func TestASelectionWhileScrolledCopiesWhatIsUnderThePointer(t *testing.T) {
	m := wideSchema(t)
	for range 30 {
		m.scrollHorizontally(horizontalStep)
	}
	drawn := strings.Split(drawnDefinition(m), "\n")
	at := -1
	for i, line := range drawn {
		if strings.Contains(line, "THE_END") {
			at = i
		}
	}
	require.GreaterOrEqual(t, at, 0)
	col := definitionColumn(m) + strings.Index(drawn[at], "THE_END")
	row := tabBarHeight + at

	cmd := drag(m, col, row, col+len("THE_END"), row)

	assert.Equal(t, "THE_END", clipboardText(t, cmd))
	painted := m.highlightSelection(m.viewSchema(m.windowWidth, m.schemaHeight()))
	assert.Contains(t, painted, selectionStyle.Render("THE_END"), "and the highlight is on it")
}
