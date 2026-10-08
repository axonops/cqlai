package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/ui/completion"
)

// listedModel is the console with the completion list open over it.
func listedModel(t *testing.T, typed string, completions ...string) *MainModel {
	t.Helper()
	m := multiLineModel(t)
	m.input.SetValue(typed)
	m.input.CursorEnd()
	m.completions = completions
	m.completionIndex = 0
	m.showCompletions = true

	// The window is what is drawn: the click is tested against its height.
	m.windowHeight = len(strings.Split(m.View().Content, "\n"))
	return m
}

// onScreen is where a completion is drawn: the column and row of its name.
func drawnAt(t *testing.T, m *MainModel, name string) (col, row int) {
	t.Helper()
	for y, line := range strings.Split(stripAnsiForTest(m.View().Content), "\n") {
		if x := strings.Index(line, "  "+name+" "); x >= 0 {
			return len([]rune(line[:x])) + 2, y
		}
		if x := strings.Index(line, "→ "+name+" "); x >= 0 {
			return len([]rune(line[:x])) + 2, y
		}
	}
	t.Fatalf("%q is not on the screen", name)
	return 0, 0
}

func clickList(m *MainModel, col, row int) *MainModel {
	m, _ = m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseLeft, X: col, Y: row})
	return m
}

// TestClickingACompletionUsesIt, the same as Enter on it.
func TestClickingACompletionUsesIt(t *testing.T) {
	m := listedModel(t, "SELECT * FROM shop.", "customers", "orders", "products")
	col, row := drawnAt(t, m, "orders")

	m = clickList(m, col, row)

	byKey := listedModel(t, "SELECT * FROM shop.", "customers", "orders", "products")
	byKey, _ = byKey.handleDownArrow(tea.KeyPressMsg{Code: tea.KeyDown})
	byKey, _ = byKey.handleEnterKey()

	assert.Equal(t, "SELECT * FROM shop.orders ", m.input.Value())
	assert.Equal(t, byKey.input.Value(), m.input.Value(), "the same as Enter on it")
	assert.False(t, m.showCompletions, "the list closes, as it does after Enter")
}

// TestClickingACompletionInAScrolledList: the arrow above the list moves the
// rows down one, and the click has to follow.
func TestClickingACompletionInAScrolledList(t *testing.T) {
	var names []string
	for i := range 25 {
		names = append(names, fmt.Sprintf("table_%02d", i))
	}
	m := listedModel(t, "SELECT * FROM shop.", names...)
	m.completionIndex = 15
	m.completionScrollOffset = 8

	col, row := drawnAt(t, m, "table_12")
	m = clickList(m, col, row)

	assert.Equal(t, "SELECT * FROM shop.table_12 ", m.input.Value())
}

// TestClickingTheListTitleDoesNothing: a press that misses a row by one is not
// walking away from the list.
func TestClickingTheListTitleDoesNothing(t *testing.T) {
	m := listedModel(t, "SELECT * FROM shop.", "customers", "orders")
	col, row := drawnAt(t, m, "customers")

	m = clickList(m, col, row-1)

	assert.True(t, m.showCompletions)
	assert.Equal(t, "SELECT * FROM shop.", m.input.Value())
}

// TestClickingANoteDoesNothing: a note about what to type has nothing to put
// in the prompt.
func TestClickingANoteDoesNothing(t *testing.T) {
	note := completion.Hint("table name")
	m := listedModel(t, "SELECT * FROM shop.", note, "orders")
	require.True(t, completion.IsHint(note))
	col, row := drawnAt(t, m, note)

	m = clickList(m, col, row)

	assert.True(t, m.showCompletions)
	assert.Equal(t, "SELECT * FROM shop.", m.input.Value())
}

// TestClickingOutsideTheListClosesIt and leaves the prompt alone.
func TestClickingOutsideTheListClosesIt(t *testing.T) {
	m := listedModel(t, "SELECT * FROM shop.", "customers", "orders")

	m = clickList(m, 90, 3)

	assert.False(t, m.showCompletions)
	assert.Equal(t, "SELECT * FROM shop.", m.input.Value())
}

// TestTheWheelMovesThroughTheCompletions rather than scrolling the console
// behind them.
func TestTheWheelMovesThroughTheCompletions(t *testing.T) {
	m := listedModel(t, "SELECT * FROM shop.", "customers", "orders", "products")

	m, _ = m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	m, _ = m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	assert.Equal(t, 2, m.completionIndex)

	m, _ = m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	assert.Equal(t, 1, m.completionIndex)
}

// TestTheListLeavesThePromptInSight: it narrows as you type, so what is typed
// has to stay on the screen.
func TestTheListLeavesThePromptInSight(t *testing.T) {
	m := listedModel(t, "SELECT * FROM shop.", "customers", "orders")
	assert.Contains(t, stripAnsiForTest(m.View().Content), "> SELECT * FROM shop.")
}

// TestTheKeyHelpFitsOnOneLine: the box was two columns narrower inside than
// its rows, so the help and the rule under the list wrapped.
func TestTheKeyHelpFitsOnOneLine(t *testing.T) {
	m := listedModel(t, "SELECT * FROM shop.", "customers", "orders")
	view := stripAnsiForTest(m.View().Content)
	assert.Contains(t, view, "↑↓/Tab: Navigate • Enter/Click: Accept • Esc: Close")
	assert.NotContains(t, view, "│──  ")
}
