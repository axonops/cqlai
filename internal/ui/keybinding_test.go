package ui

import (
	"testing"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The v2 migration rewrote roughly seventy key bindings. These drive
// handleKeyboardInput the way the program does and assert on what the model
// actually did, so a binding that stops reaching its handler fails here rather
// than in someone's terminal.

// TestKeyStringsMatchTheDispatch pins the strings the dispatch switches on.
//
// v1 switched on msg.Type, which ignored modifiers. v2's String() spells them
// out, so "alt+left" does not match a "left" case and Alt bindings fall through
// to default in silence. That is how Alt+Left/Right stopped scrolling.
func TestKeyStringsMatchTheDispatch(t *testing.T) {
	tests := []struct {
		msg  tea.KeyPressMsg
		want string
	}{
		{tea.KeyPressMsg{Code: tea.KeyLeft}, "left"},
		{tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModAlt}, "alt+left"},
		{tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModAlt}, "alt+right"},
		{tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModAlt}, "alt+up"},
		{tea.KeyPressMsg{Code: tea.KeyDown, Mod: tea.ModAlt}, "alt+down"},
		{tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}, "ctrl+c"},
		{tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl}, "ctrl+r"},
		{tea.KeyPressMsg{Code: tea.KeySpace}, "space"},
		{tea.KeyPressMsg{Code: tea.KeyEscape}, "esc"},
		{tea.KeyPressMsg{Code: tea.KeyPgDown}, "pgdown"},
		{tea.KeyPressMsg{Code: tea.KeyF3}, "f3"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.msg.String())
	}
}

// tableModel is a model showing a wide result, which is what Alt+Left/Right act on.
func tableModel(t *testing.T) *MainModel {
	t.Helper()

	input := textinput.New()
	input.SetWidth(40)

	m := &MainModel{
		input:         input,
		viewMode:      "table",
		hasTable:      true,
		ready:         true,
		styles:        DefaultStyles(),
		tableViewport: viewport.New(viewport.WithWidth(40), viewport.WithHeight(10)),
		tableWidth:    400, // far wider than the viewport, so there is room to scroll
		lastTableData: [][]string{
			{"a", "b", "c"},
			{"1", "2", "3"},
		},
	}
	m.tableViewport.SetContent("some table content")
	return m
}

// TestAltArrowsScrollTheTableSideways is the reported regression: Alt+Left and
// Alt+Right stopped scrolling wide tables after the migration.
func TestAltArrowsScrollTheTableSideways(t *testing.T) {
	m := tableModel(t)
	require.Equal(t, 0, m.horizontalOffset)

	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModAlt})
	assert.Greater(t, m.horizontalOffset, 0,
		"Alt+Right must scroll a wide table to the right")

	scrolled := m.horizontalOffset
	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModAlt})
	assert.Less(t, m.horizontalOffset, scrolled,
		"Alt+Left must scroll back to the left")
}

// TestAltArrowsScrollTheViewport covers the vertical half of the same binding.
func TestAltArrowsScrollTheViewport(t *testing.T) {
	m := tableModel(t)
	m.tableViewport.SetContent("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14\n15\n16\n17\n18\n19\n20")

	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyDown, Mod: tea.ModAlt})
	assert.Greater(t, m.tableViewport.YOffset(), 0, "Alt+Down must scroll the viewport down")

	down := m.tableViewport.YOffset()
	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModAlt})
	assert.Less(t, m.tableViewport.YOffset(), down, "Alt+Up must scroll back up")
}

// TestFunctionKeysSwitchMode covers the four mode keys, which are the most
// visible bindings and were all rewritten.
func TestFunctionKeysSwitchMode(t *testing.T) {
	tests := []struct {
		code rune
		want string
	}{
		{tea.KeyF2, "history"},
		{tea.KeyF3, "table"},
		{tea.KeyF4, "trace"},
	}

	for _, tt := range tests {
		m := tableModel(t)
		m.hasTrace = true
		m.handleKeyboardInput(tea.KeyPressMsg{Code: tt.code})
		assert.Equal(t, tt.want, m.viewMode, "%v should switch to %q", tt.code, tt.want)
	}
}

// TestEscapeTogglesNavigationMode covers the binding the README documents as
// ESC, which CLAUDE.md still describes as Ctrl+N.
func TestEscapeTogglesNavigationMode(t *testing.T) {
	m := tableModel(t)
	require.False(t, m.navigationMode)

	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.True(t, m.navigationMode, "ESC should enter navigation mode")

	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.False(t, m.navigationMode, "ESC again should leave it")
}
