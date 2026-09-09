package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

func TestClickingAConnectionFactOpensNothing(t *testing.T) {
	m := chooserModel()

	for _, seg := range m.statusBar.segments() {
		if seg.clickable() {
			continue
		}
		m.clickStatusSetting((seg.start + seg.end) / 2)
		assert.False(t, m.chooser.active, "%q is not a setting", seg.label)
	}
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
