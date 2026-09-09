package ui

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func chooserModel() *MainModel {
	return &MainModel{
		styles:       DefaultStyles(),
		mouseEnabled: true,
		windowWidth:  120,
		windowHeight: 30,
		statusBar:    testStatusBar(),
	}
}

// TestClickingASettingOpensItsChoices is the point of the feature.
func TestClickingASettingOpensItsChoices(t *testing.T) {
	m := chooserModel()

	var col int
	for _, seg := range m.statusBar.segments() {
		if seg.setting == settingConsistency {
			col = (seg.start + seg.end) / 2
		}
	}

	m.clickStatusSetting(col)

	require.True(t, m.chooser.active, "clicking CL should open its list")
	assert.Equal(t, settingConsistency, m.chooser.setting)
	assert.Contains(t, m.chooser.choices, "LOCAL_QUORUM")
	assert.Equal(t, "LOCAL_ONE", m.chooser.choices[m.chooser.selected],
		"the list should open on the value already set")
}

func TestClickingPastTheEndOfTheLineOpensNothing(t *testing.T) {
	m := chooserModel()

	last := m.statusBar.segments()[len(m.statusBar.segments())-1]
	m.clickStatusSetting(last.end + 4)

	assert.False(t, m.chooser.active, "there is no field out there")
}

// TestChooserDismissedByAnyKey: there is no keyboard navigation, so a keypress
// means you have gone back to typing and the list should get out of the way.
func TestChooserDismissedByAnyKey(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{
		{Code: tea.KeyEscape},
		{Code: 'x'},
		{Code: tea.KeyDown},
	} {
		m := chooserModel()
		m.openSettingChooser(settingConsistency, 10,
			[]string{"ANY", "ONE", "QUORUM"}, "ONE")

		m.handleSettingChooserKey(key)

		assert.False(t, m.chooser.active, "%v should dismiss the list", key.String())
		assert.Empty(t, m.fullHistoryContent, "dismissing must not change a setting")
	}
}

// TestClickingAChoiceAppliesIt is how a value gets picked, now that there is no
// keyboard navigation.
func TestClickingAChoiceAppliesIt(t *testing.T) {
	const w, h = 120, 30

	m := chooserModel()
	m.openSettingChooser(settingConsistency, 40,
		[]string{"ANY", "ONE", "QUORUM", "ALL"}, "ONE")

	g, ok := m.chooserGeometry(w, h)
	require.True(t, ok)

	// The third row inside the border is "QUORUM".
	choice, inside := m.choiceAt(w, h, g.x+2, g.y+1+2)
	require.True(t, inside, "a click inside the list should land on a choice")
	assert.Equal(t, "QUORUM", choice)
}

// TestClicksOutsideTheListMissIt covers the border and the surrounding screen,
// so clicking away closes rather than picking something by accident.
func TestClicksOutsideTheListMissIt(t *testing.T) {
	const w, h = 120, 30

	m := chooserModel()
	m.openSettingChooser(settingConsistency, 40,
		[]string{"ANY", "ONE", "QUORUM", "ALL"}, "ONE")

	g, ok := m.chooserGeometry(w, h)
	require.True(t, ok)

	outside := []struct {
		name     string
		col, row int
	}{
		{"top border", g.x + 2, g.y},
		{"bottom border", g.x + 2, g.y + g.height - 1},
		{"left border", g.x, g.y + 2},
		{"right border", g.x + g.width - 1, g.y + 2},
		{"left of the box", g.x - 1, g.y + 2},
		{"above the box", g.x + 2, g.y - 1},
		{"the status line itself", g.x + 2, h - 1},
	}

	for _, tt := range outside {
		_, inside := m.choiceAt(w, h, tt.col, tt.row)
		assert.False(t, inside, "a click on the %s is not a choice", tt.name)
	}
}

// TestEveryDrawnRowIsClickable walks the whole list, so no row is unreachable.
func TestEveryDrawnRowIsClickable(t *testing.T) {
	const w, h = 120, 30

	choices := []string{"ANY", "ONE", "TWO", "THREE", "QUORUM", "ALL"}
	m := chooserModel()
	m.openSettingChooser(settingConsistency, 20, choices, "ONE")

	g, ok := m.chooserGeometry(w, h)
	require.True(t, ok)
	require.Equal(t, len(choices), g.last-g.first, "the whole list should fit here")

	layer, ok := m.viewSettingChooser(w, h)
	require.True(t, ok)
	rows := strings.Split(layer.Content, "\n")

	for i := g.first; i < g.last; i++ {
		got, inside := m.choiceAt(w, h, g.x+1, g.y+1+(i-g.first))
		require.True(t, inside, "row %d should be clickable", i)
		assert.Equal(t, choices[i], got, "row %d should be %q", i, choices[i])

		// The row that answers the click has to be the row showing that
		// choice, or the list is off by one and nobody notices until they pick
		// the wrong consistency level.
		assert.Contains(t, stripAnsiForTest(rows[1+(i-g.first)]), choices[i])
	}
}

// TestChoosingTheCurrentValueDoesNothing: picking what is already set should be
// a no-op, not a logged change.
func TestChoosingTheCurrentValueDoesNothing(t *testing.T) {
	m := chooserModel()
	m.openSettingChooser(settingTracing, 10, []string{"ON", "OFF"}, "OFF")

	m.applySettingChoice("OFF")

	assert.False(t, m.chooser.active)
	assert.Empty(t, m.fullHistoryContent, "re-picking the current value should change nothing")
}

// TestSettingCommandMatchesTheTypedForm: applying must go through the same
// command a user would type, or the two routes drift apart.
func TestSettingCommandMatchesTheTypedForm(t *testing.T) {
	tests := []struct {
		setting, choice, want string
	}{
		{settingConsistency, "QUORUM", "CONSISTENCY QUORUM"},
		{settingPaging, "500", "PAGING 500"},
		{settingTracing, "ON", "TRACING ON"},
		{settingAutoFetch, "OFF", "AUTOFETCH OFF"},
		// Quoted, so a mixed-case keyspace name survives. The USE handler
		// strips the quotes again before looking it up.
		{settingKeyspace, "system", `USE "system"`},
		{settingKeyspace, "MyKeyspace", `USE "MyKeyspace"`},
		{"nonsense", "x", ""},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, settingCommand(tt.setting, tt.choice))
	}
}

// TestChooserSitsAboveTheStatusLine: the field is on the last row, so the list
// has to grow upward or it would be off screen.
func TestChooserSitsAboveTheStatusLine(t *testing.T) {
	const screenHeight = 30

	m := chooserModel()
	m.openSettingChooser(settingConsistency, 40,
		[]string{"ANY", "ONE", "QUORUM", "ALL"}, "ONE")

	layer, ok := m.viewSettingChooser(120, screenHeight)
	require.True(t, ok)

	assert.Equal(t, screenHeight-1, layer.Y+layer.Height,
		"the list should finish directly above the status line")
	assert.Equal(t, 40, layer.X, "and be anchored to the field it belongs to")
	assert.Greater(t, layer.ZIndex, 100, "it is the thing just clicked, so it goes on top")
}

// TestChooserNudgedLeftAtTheEdge: a field near the right edge must not push the
// list off screen.
func TestChooserNudgedLeftAtTheEdge(t *testing.T) {
	m := chooserModel()
	m.openSettingChooser(settingConsistency, 118,
		[]string{"LOCAL_QUORUM", "EACH_QUORUM"}, "LOCAL_QUORUM")

	layer, ok := m.viewSettingChooser(120, 30)
	require.True(t, ok)

	assert.GreaterOrEqual(t, layer.X, 0)
	assert.LessOrEqual(t, layer.X+layer.Width, 120,
		"the list must stay on screen")
}

// TestChooserScrollsWhenTooTall keeps a long list usable on a short screen.
func TestChooserScrollsWhenTooTall(t *testing.T) {
	long := make([]string, 40)
	for i := range long {
		long[i] = string(rune('a' + i%26))
	}

	c := settingChooser{choices: long, selected: 30}
	start, end := c.window(10)

	assert.Equal(t, 10, end-start, "the window should be the height asked for")
	assert.LessOrEqual(t, start, 30)
	assert.Greater(t, end, 30, "the selection has to be inside the window")

	// And it must not run past either end.
	c.selected = 0
	start, _ = c.window(10)
	assert.Equal(t, 0, start, "at the top the window should start at the top")

	c.selected = len(long) - 1
	_, end = c.window(10)
	assert.Equal(t, len(long), end, "at the bottom it should reach the last choice")
}

func TestChooserHiddenWhenClosed(t *testing.T) {
	m := chooserModel()
	_, ok := m.viewSettingChooser(120, 30)
	assert.False(t, ok, "nothing to draw when nothing is open")
}

// TestLongChoicesStayOnScreen: keyspace names are the one setting whose values
// are not short and known in advance, so the box has to be capped and the names
// truncated rather than pushed off the side.
func TestLongChoicesStayOnScreen(t *testing.T) {
	const screenWidth = 40

	m := chooserModel()
	m.windowWidth = screenWidth
	m.openSettingChooser(settingKeyspace, 4, []string{
		"short",
		"a_keyspace_name_far_longer_than_this_terminal_is_wide",
	}, "short")

	layer, ok := m.viewSettingChooser(screenWidth, 30)
	require.True(t, ok)

	assert.LessOrEqual(t, layer.X+layer.Width, screenWidth, "the box ran off the screen")
	for _, line := range strings.Split(layer.Content, "\n") {
		assert.LessOrEqual(t, lipgloss.Width(line), screenWidth,
			"a row of the box is wider than the screen: %q", line)
	}
}

// TestChooserGeometryMatchesWhatItDraws is the invariant behind every click on
// the list: choiceAt tests a position against chooserGeometry, so if the box
// renders even a column wider than the geometry says, the last column of it
// stops responding.
func TestChooserGeometryMatchesWhatItDraws(t *testing.T) {
	cases := [][]string{
		{"ON", "OFF"},
		{"ANY", "ONE", "TWO", "THREE", "QUORUM", "LOCAL_QUORUM", "EACH_QUORUM"},
		{"a", "a_much_longer_keyspace_name_here"},
	}

	for _, choices := range cases {
		m := chooserModel()
		m.openSettingChooser(settingKeyspace, 4, choices, choices[0])

		g, ok := m.chooserGeometry(m.windowWidth, m.windowHeight)
		require.True(t, ok)

		layer, ok := m.viewSettingChooser(m.windowWidth, m.windowHeight)
		require.True(t, ok)

		rows := strings.Split(layer.Content, "\n")
		assert.Len(t, rows, g.height, "%v: drew a different number of rows than the geometry claims", choices)
		for _, row := range rows {
			assert.Equal(t, g.width, lipgloss.Width(row),
				"%v: row %q is not the width the geometry claims", choices, row)
		}
	}
}

// TestKeyspacesAreNotAConstantList: they come from the cluster, so
// settingChoices must not pretend to know them.
func TestKeyspacesAreNotAConstantList(t *testing.T) {
	choices, _ := settingChoices(settingKeyspace, "system")
	assert.Empty(t, choices)
}

// TestClickingKSWithNoConnectionOpensNothing: an empty list would be a box
// with nothing in it.
func TestClickingKSWithNoConnectionOpensNothing(t *testing.T) {
	m := chooserModel()

	for _, seg := range m.statusBar.segments() {
		if seg.setting == settingKeyspace {
			m.clickStatusSetting((seg.start + seg.end) / 2)
		}
	}

	assert.False(t, m.chooser.active)
}

// TestConsistencyChoicesComeFromTheSessionsOwnList: the chooser must not carry
// its own set of levels, or it drifts from the one CONSISTENCY applies. That
// drift is what made picking SERIAL print an error rather than change
// anything.
func TestConsistencyChoicesComeFromTheSessionsOwnList(t *testing.T) {
	choices, _ := settingChoices(settingConsistency, "ONE")
	assert.Equal(t, db.ConsistencyLevels(), choices)
}

// chooserWithChoices opens a list of n numbered choices on a screen of the
// given height, which is how a long keyspace list is reproduced.
func chooserWithChoices(n, screenHeight int) (*MainModel, []string) {
	choices := make([]string, n)
	for i := range choices {
		choices[i] = fmt.Sprintf("keyspace_%02d", i)
	}

	m := chooserModel()
	m.windowHeight = screenHeight
	m.openSettingChooser(settingKeyspace, 4, choices, choices[0])
	return m, choices
}

// TestEveryChoiceIsReachableByScrolling is the bug: a cluster with more
// keyspaces than the box had rows showed the first few and no way to the rest.
// There is no keyboard navigation here by design, so the wheel is the only way
// the window can move.
func TestEveryChoiceIsReachableByScrolling(t *testing.T) {
	const screenHeight = 20
	m, choices := chooserWithChoices(40, screenHeight)

	seen := map[string]bool{}
	record := func() {
		g, ok := m.chooserGeometry(m.windowWidth, screenHeight)
		require.True(t, ok)
		for i := g.first; i < g.last; i++ {
			seen[choices[i]] = true
		}
	}

	record()
	require.Less(t, len(seen), len(choices), "the list should not fit, or this proves nothing")

	// Wheel to the bottom, then back to the top.
	for range len(choices) {
		m.scrollChooser(chooserScrollStep)
		record()
	}
	for range len(choices) {
		m.scrollChooser(-chooserScrollStep)
		record()
	}

	for _, choice := range choices {
		assert.True(t, seen[choice], "%s could never be scrolled to", choice)
	}
}

// TestScrollingStopsAtTheEnds: the wheel must not run the window off either
// end and leave the box empty or short.
func TestScrollingStopsAtTheEnds(t *testing.T) {
	const screenHeight = 20
	m, choices := chooserWithChoices(40, screenHeight)

	for range 200 {
		m.scrollChooser(chooserScrollStep)
	}
	g, ok := m.chooserGeometry(m.windowWidth, screenHeight)
	require.True(t, ok)
	assert.Equal(t, len(choices), g.last, "scrolling down should stop with the last choice showing")
	assert.Equal(t, g.height-2, g.last-g.first, "the box should still be full")

	for range 200 {
		m.scrollChooser(-chooserScrollStep)
	}
	g, _ = m.chooserGeometry(m.windowWidth, screenHeight)
	assert.Equal(t, 0, g.first, "scrolling up should stop at the first choice")
	assert.Equal(t, g.height-2, g.last-g.first)
}

// TestTheListUsesTheRoomAvailable: it used to be capped at ten rows however
// tall the terminal was.
func TestTheListUsesTheRoomAvailable(t *testing.T) {
	short, _ := chooserWithChoices(40, 14)
	tall, _ := chooserWithChoices(40, 40)

	sg, ok := short.chooserGeometry(short.windowWidth, 14)
	require.True(t, ok)
	tg, ok := tall.chooserGeometry(tall.windowWidth, 40)
	require.True(t, ok)

	assert.Greater(t, tg.height, sg.height, "a taller terminal should show more choices")
	assert.Greater(t, tg.last-tg.first, 10, "the ten row cap is gone")
}

// TestShortListsStayShort: only the long ones grow. A box covering the screen
// to offer ON and OFF would be absurd.
func TestShortListsStayShort(t *testing.T) {
	m := chooserModel()
	m.windowHeight = 40
	m.openSettingChooser(settingTracing, 10, []string{"ON", "OFF"}, "OFF")

	g, ok := m.chooserGeometry(m.windowWidth, 40)
	require.True(t, ok)
	assert.Equal(t, 2+2, g.height, "two choices and two borders")
}

// TestTheScrollbarSaysWhereYouAre: without it, a box holding ten of your
// thirty keyspaces looks exactly like one holding all of them, and there is no
// way to tell how far down the list you have got.
func TestTheScrollbarSaysWhereYouAre(t *testing.T) {
	const screenHeight = 24
	m, _ := chooserWithChoices(40, screenHeight)

	// The scrollbar is the last column inside the box.
	bar := func() string {
		layer, ok := m.viewSettingChooser(m.windowWidth, screenHeight)
		require.True(t, ok)

		var col []rune
		for _, row := range strings.Split(stripAnsiForTest(layer.Content), "\n")[1:] {
			runes := []rune(row)
			if len(runes) < 2 {
				continue
			}
			col = append(col, runes[len(runes)-2])
		}
		return string(col[:len(col)-1]) // drop the bottom border
	}

	atTop := bar()
	assert.Equal(t, '\u2588', []rune(atTop)[0], "the thumb should start at the top: %q", atTop)
	assert.Contains(t, atTop, "\u2591", "and there should be track below it")

	for range 200 {
		m.scrollChooser(chooserScrollStep)
	}
	atBottom := bar()
	assert.Equal(t, '\u2588', []rune(atBottom)[len([]rune(atBottom))-1],
		"scrolled to the end, the thumb should reach the bottom: %q", atBottom)
	assert.Equal(t, '\u2591', []rune(atBottom)[0], "and there should be track above it")
}

// TestNoScrollbarWhenTheListFits: a bar that is always full says nothing.
func TestNoScrollbarWhenTheListFits(t *testing.T) {
	m := chooserModel()
	m.windowHeight = 24
	m.openSettingChooser(settingTracing, 10, []string{"ON", "OFF"}, "OFF")

	layer, ok := m.viewSettingChooser(m.windowWidth, 24)
	require.True(t, ok)

	plain := stripAnsiForTest(layer.Content)
	assert.NotContains(t, plain, "\u2588")
	assert.NotContains(t, plain, "\u2591")
}

// TestTheThumbGrowsWithTheProportionShowing.
func TestTheThumbGrowsWithTheProportionShowing(t *testing.T) {
	count := func(thumb []bool) int {
		n := 0
		for _, t := range thumb {
			if t {
				n++
			}
		}
		return n
	}

	assert.Equal(t, 0, count(scrollbarColumn(10, 0, 10)), "a list that fits has no thumb")
	assert.Greater(t, count(scrollbarColumn(10, 0, 20)), count(scrollbarColumn(10, 0, 200)),
		"showing half the list should give a bigger thumb than showing a twentieth")
	assert.GreaterOrEqual(t, count(scrollbarColumn(10, 0, 1000)), 1,
		"a huge list still needs a thumb you can see")
}

// TestTheWheelScrollsTheListNotTheViewBehindIt.
func TestTheWheelScrollsTheListNotTheViewBehindIt(t *testing.T) {
	const screenHeight = 20
	m, _ := chooserWithChoices(40, screenHeight)
	m.historyViewport = viewport.New(viewport.WithWidth(80), viewport.WithHeight(10))
	m.historyViewport.SetContent(strings.Repeat("line\n", 100))
	m.viewMode = "history"

	before := m.historyViewport.YOffset()
	m.handleMouseInput(tea.MouseWheelMsg{Button: tea.MouseWheelDown, X: 5, Y: 5})

	g, ok := m.chooserGeometry(m.windowWidth, screenHeight)
	require.True(t, ok)
	assert.Equal(t, chooserScrollStep, g.first, "the wheel should have moved the list")
	assert.Equal(t, before, m.historyViewport.YOffset(), "the view behind must not have moved")
}

// settingColumn is the middle of a setting's field on the status line.
func settingColumn(m *MainModel, setting string) int {
	for _, seg := range m.statusBar.segments() {
		if seg.setting == setting {
			return (seg.start + seg.end) / 2
		}
	}
	return -1
}

// pressAt delivers a left button press, the way the terminal does.
func pressAt(m *MainModel, col, row int) {
	m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseLeft, X: col, Y: row})
}

// TestClickingTheSameSettingClosesItsList is the bug: the press closed the list
// and reopened it in the same breath, so it never appeared to go away and you
// had to click somewhere else to be rid of it.
func TestClickingTheSameSettingClosesItsList(t *testing.T) {
	m := chooserModel()
	col := settingColumn(m, settingConsistency)
	require.Positive(t, col)

	pressAt(m, col, m.windowHeight-1)
	require.True(t, m.chooser.active, "the first click should open the list")

	pressAt(m, col, m.windowHeight-1)
	assert.False(t, m.chooser.active, "clicking the same setting again should close it")
}

// TestClickingAnotherSettingSwitchesLists.
func TestClickingAnotherSettingSwitchesLists(t *testing.T) {
	m := chooserModel()

	pressAt(m, settingColumn(m, settingConsistency), m.windowHeight-1)
	require.Equal(t, settingConsistency, m.chooser.setting)

	pressAt(m, settingColumn(m, settingTracing), m.windowHeight-1)
	assert.True(t, m.chooser.active, "clicking a different setting should open that one")
	assert.Equal(t, settingTracing, m.chooser.setting)
}

// TestClickingEmptyStatusLineClosesTheList: a click on the bar but not on a
// field dismisses whatever was open.
func TestClickingEmptyStatusLineClosesTheList(t *testing.T) {
	m := chooserModel()
	pressAt(m, settingColumn(m, settingConsistency), m.windowHeight-1)
	require.True(t, m.chooser.active)

	segs := m.statusBar.segments()
	pressAt(m, segs[len(segs)-1].end+4, m.windowHeight-1)

	assert.False(t, m.chooser.active)
}

// TestClickingTheTabsClosesTheList.
func TestClickingTheTabsClosesTheList(t *testing.T) {
	m := chooserModel()
	m.viewMode = "history"
	pressAt(m, settingColumn(m, settingConsistency), m.windowHeight-1)
	require.True(t, m.chooser.active)

	pressAt(m, 2, 0)
	assert.False(t, m.chooser.active)
}
