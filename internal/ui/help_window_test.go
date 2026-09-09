package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func helpModel() *MainModel {
	return &MainModel{
		viewMode:        "history",
		hasTable:        true,
		hasTrace:        true,
		styles:          DefaultStyles(),
		aiConfig:        configuredAI(),
		mouseEnabled:    true,
		windowWidth:     100,
		windowHeight:    30,
		input:           newTestInput(),
		historyViewport: viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}
}

// TestTheHelpButtonIsOnTheTabLine, at the right-hand end.
func TestTheHelpButtonIsOnTheTabLine(t *testing.T) {
	m := helpModel()

	assert.Contains(t, stripAnsiForTest(m.ViewTabBar(100)), "Help")

	spans := m.layoutTabs(100)
	help := spans[len(spans)-1]
	assert.Equal(t, helpMode, help.mode, "the button should be last")
	assert.Equal(t, 100, help.end, "and against the right edge")
}

// TestClickingHelpOpensAndClosesIt.
func TestClickingHelpOpensAndClosesIt(t *testing.T) {
	m := helpModel()

	spans := m.layoutTabs(m.windowWidth)
	help := spans[len(spans)-1]
	col := (help.start + help.end) / 2

	pressAt(m, col, 0)
	require.True(t, m.help.active, "clicking Help should open it")

	pressAt(m, col, 0)
	assert.False(t, m.help.active, "clicking it again should close it")
}

// TestF1OpensAndClosesIt too, since that is the conventional key and it was
// unbound.
func TestF1OpensAndClosesIt(t *testing.T) {
	m := helpModel()

	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyF1})
	require.True(t, m.help.active)

	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyF1})
	assert.False(t, m.help.active)
}

// TestEscapeClosesIt.
func TestEscapeClosesIt(t *testing.T) {
	m := helpModel()
	m.toggleHelp()

	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.False(t, m.help.active)
}

// TestClickingElsewhereClosesIt: the window covers the screen, so you cannot
// aim at what is underneath anyway.
func TestClickingElsewhereClosesIt(t *testing.T) {
	m := helpModel()
	m.toggleHelp()

	pressAt(m, 40, 10)

	assert.False(t, m.help.active)
}

// TestTheContentComesFromTheHelpCommand, so typing HELP and clicking Help
// cannot drift apart.
func TestTheContentComesFromTheHelpCommand(t *testing.T) {
	m := helpModel()
	m.toggleHelp()

	layer, ok := m.viewHelp(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	shown := stripAnsiForTest(layer.Content)
	for _, row := range router.HelpRows()[:12] {
		if row[1] == "" {
			continue
		}
		assert.Contains(t, shown, row[1], "the window should show the same rows as HELP")
	}
}

// TestTheWindowScrolls, and stops at both ends.
func TestTheWindowScrolls(t *testing.T) {
	m := helpModel()
	m.toggleHelp()

	first := func() string {
		layer, ok := m.viewHelp(m.windowWidth, m.windowHeight)
		require.True(t, ok)
		return strings.Split(stripAnsiForTest(layer.Content), "\n")[2]
	}

	top := first()
	m.scrollHelp(helpPageRows)
	assert.NotEqual(t, top, first(), "paging down should have moved it")

	for range 100 {
		m.scrollHelp(helpPageRows)
	}
	bottom := first()

	m.scrollHelp(helpPageRows)
	assert.Equal(t, bottom, first(), "it should stop at the end")

	for range 100 {
		m.scrollHelp(-helpPageRows)
	}
	assert.Equal(t, top, first(), "and come back to the start")
	assert.Zero(t, m.help.scroll)
}

// TestTheWheelScrollsTheWindowNotTheViewBehindIt.
func TestTheWheelScrollsTheWindowNotTheViewBehindIt(t *testing.T) {
	m := helpModel()
	m.historyViewport.SetContent(strings.Repeat("line\n", 200))
	m.toggleHelp()

	behind := m.historyViewport.YOffset()
	m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelDown, X: 50, Y: 15})

	assert.Equal(t, wheelLines, m.help.scroll)
	assert.Equal(t, behind, m.historyViewport.YOffset(), "the view behind must not have moved")
}

// TestPageKeysScrollTheWindowNotTheViewBehindIt.
func TestPageKeysScrollTheWindowNotTheViewBehindIt(t *testing.T) {
	m := helpModel()
	m.historyViewport.SetContent(strings.Repeat("line\n", 200))
	m.toggleHelp()

	behind := m.historyViewport.YOffset()
	m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyPgDown})

	assert.Equal(t, helpPageRows, m.help.scroll)
	assert.Equal(t, behind, m.historyViewport.YOffset())
}

// TestTypingDoesNotReachThePromptWhileHelpIsOpen.
func TestTypingDoesNotReachThePromptWhileHelpIsOpen(t *testing.T) {
	m := helpModel()
	m.toggleHelp()

	m.handleKeyboardInput(tea.KeyPressMsg{Code: 'x', Text: "x"})

	assert.Empty(t, m.input.Value(), "the window is what the keys are aimed at")
}

// TestTheHelpScrollbarSaysWhereYouAre.
func TestTheHelpScrollbarSaysWhereYouAre(t *testing.T) {
	m := helpModel()
	m.toggleHelp()

	bar := func() string {
		layer, ok := m.viewHelp(m.windowWidth, m.windowHeight)
		require.True(t, ok)

		var col []rune
		rows := strings.Split(stripAnsiForTest(layer.Content), "\n")
		for _, row := range rows[2 : len(rows)-1] {
			runes := []rune(row)
			col = append(col, runes[len(runes)-2])
		}
		return string(col)
	}

	atTop := bar()
	assert.Equal(t, '█', []rune(atTop)[0], "the thumb starts at the top: %q", atTop)

	for range 100 {
		m.scrollHelp(helpPageRows)
	}
	atBottom := bar()
	assert.Equal(t, '█', []rune(atBottom)[len([]rune(atBottom))-1],
		"and reaches the bottom: %q", atBottom)
}

// TestTheWindowIsSquareAndFitsTheScreen.
func TestTheWindowIsSquareAndFitsTheScreen(t *testing.T) {
	for _, size := range [][2]int{{100, 30}, {80, 24}, {200, 60}, {40, 12}} {
		m := helpModel()
		m.windowWidth, m.windowHeight = size[0], size[1]
		m.toggleHelp()

		layer, ok := m.viewHelp(size[0], size[1])
		if !ok {
			continue
		}

		assert.LessOrEqual(t, layer.X+layer.Width, size[0], "%v: too wide", size)
		assert.LessOrEqual(t, layer.Y+layer.Height, size[1], "%v: too tall", size)

		rows := strings.Split(layer.Content, "\n")
		assert.Len(t, rows, layer.Height, "%v: wrong number of rows", size)
		for _, row := range rows {
			assert.Equal(t, layer.Width, lipgloss.Width(row), "%v: ragged row %q", size, row)
		}
	}
}

// TestATerminalTooSmallGetsNoWindow rather than a box with nothing in it.
func TestATerminalTooSmallGetsNoWindow(t *testing.T) {
	m := helpModel()
	m.toggleHelp()

	_, ok := m.viewHelp(18, 30)
	assert.False(t, ok, "too narrow")

	_, ok = m.viewHelp(100, 5)
	assert.False(t, ok, "too short")
}

// TestTheButtonGivesWayToTheTabNames.
//
// Key hints go first, because the keys are in the help the button opens. The
// names do not: single letters are harder to read than a line with no button,
// and F1 still opens it.
func TestTheButtonGivesWayToTheTabNames(t *testing.T) {
	m := helpModel()

	wide := stripAnsiForTest(m.ViewTabBar(100))
	assert.Contains(t, wide, "Console (F2)")
	assert.Contains(t, wide, "Help")

	// Room for the names and the button, but not the key hints.
	medium := stripAnsiForTest(m.ViewTabBar(60))
	assert.Contains(t, medium, "Console")
	assert.NotContains(t, medium, "(F2)")
	assert.Contains(t, medium, "Help")

	// Room for the names only. The button goes rather than the names.
	tight := stripAnsiForTest(m.ViewTabBar(44))
	assert.Contains(t, tight, "Console")
	assert.NotContains(t, tight, "Help")
}

// TestTheTabLineNeverWrapsWithTheButtonOnIt, at any width.
func TestTheTabLineNeverWrapsWithTheButtonOnIt(t *testing.T) {
	m := helpModel()

	for width := 1; width <= 200; width++ {
		bar := m.ViewTabBar(width)
		assert.NotContains(t, bar, "\n", "wrapped at width %d", width)

		plain := stripAnsiForTest(bar)
		assert.LessOrEqual(t, len([]rune(strings.TrimRight(plain, " "))), width,
			"overflowed at width %d: %q", width, plain)
	}
}

// TestTheButtonNeverSitsOnATab: a click has to land on what was drawn.
func TestTheButtonNeverSitsOnATab(t *testing.T) {
	m := helpModel()

	for width := 1; width <= 200; width++ {
		spans := m.layoutTabs(width)
		for i := 1; i < len(spans); i++ {
			assert.GreaterOrEqual(t, spans[i].start, spans[i-1].end,
				"width %d: %q overlaps %q", width, spans[i].mode, spans[i-1].mode)
		}
	}
}

// TestAltHOpensItToo. Terminator, Konsole and others take F1 for their own
// help before the application sees it, so F1 cannot be the only key.
func TestAltHOpensItToo(t *testing.T) {
	m := helpModel()

	m.handleKeyboardInput(tea.KeyPressMsg{Code: 'h', Mod: tea.ModAlt})
	require.True(t, m.help.active, "Alt+H should open the help")

	m.handleKeyboardInput(tea.KeyPressMsg{Code: 'h', Mod: tea.ModAlt})
	assert.False(t, m.help.active, "and close it again")
}

// TestTheWindowNamesBothKeys, because on a terminal that takes F1 this is
// where you find out there is another one.
func TestTheWindowNamesBothKeys(t *testing.T) {
	m := helpModel()
	m.toggleHelp()

	layer, ok := m.viewHelp(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	shown := stripAnsiForTest(layer.Content)
	assert.Contains(t, shown, "F1")
	assert.Contains(t, shown, "Alt+H")
}

// TestTheHelpRowsNameBothKeys, so HELP says it as well.
func TestTheHelpRowsNameBothKeys(t *testing.T) {
	var found bool
	for _, row := range router.HelpRows() {
		if strings.Contains(row[1], "Alt+H") {
			found = true
		}
	}
	assert.True(t, found, "the help text should say how to open the help")
}
