package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTabBarNeverWraps(t *testing.T) {
	m := &MainModel{viewMode: "history", hasTable: true, hasTrace: true}

	// Every width from unusable to generous.
	for width := 1; width <= 200; width++ {
		bar := m.ViewTabBar(width)
		assert.NotContains(t, bar, "\n", "the tab bar wrapped at width %d", width)

		plain := stripAnsiForTest(bar)
		assert.LessOrEqual(t, len([]rune(strings.TrimRight(plain, " "))), width,
			"tab bar overflowed at width %d: %q", width, plain)
	}
}

func TestTabBarShowsKeysWhenThereIsRoom(t *testing.T) {
	m := &MainModel{viewMode: "history", hasTable: true, hasTrace: true}

	wide := stripAnsiForTest(m.ViewTabBar(120))
	for _, want := range []string{"CQL (F2)", "Table (F3)", "Trace (F4)", "AI (F5)"} {
		assert.Contains(t, wide, want)
	}

	// Narrower: names survive, key hints are dropped rather than wrapping.
	medium := stripAnsiForTest(m.ViewTabBar(30))
	assert.Contains(t, medium, "CQL")
	assert.Contains(t, medium, "Table")
	assert.NotContains(t, medium, "(F2)")
}

// TestTabBarMarksTheCurrentMode checks exactly one tab is the active one, and
// that it is the mode being viewed. Colour cannot be asserted here: lipgloss
// emits no escape codes without a terminal, so the test looks at the layout.
func TestTabBarMarksTheCurrentMode(t *testing.T) {
	for _, mode := range []string{"history", "table", "trace", "ai"} {
		t.Run(mode, func(t *testing.T) {
			m := &MainModel{viewMode: mode, hasTable: true, hasTrace: true}

			active := []string{}
			for _, span := range m.layoutTabs(120) {
				if span.active {
					active = append(active, span.mode)
				}
			}

			require.Len(t, active, 1, "exactly one tab should be active")
			assert.Equal(t, mode, active[0])
		})
	}
}

// TestTabBarUnknownModeHighlightsNothing guards the edge where viewMode is
// something the bar does not list, so it must not pick an arbitrary tab.
func TestTabBarUnknownModeHighlightsNothing(t *testing.T) {
	m := &MainModel{viewMode: "ai_info", hasTable: true, hasTrace: true}
	for _, span := range m.layoutTabs(120) {
		assert.False(t, span.active, "%q should not be marked active", span.mode)
	}
}

// TestTabAvailability covers the "modes with no data look unavailable" rule.
func TestTabAvailability(t *testing.T) {
	empty := &MainModel{viewMode: "history"}
	assert.True(t, empty.tabAvailable("history"))
	assert.True(t, empty.tabAvailable("ai"))
	assert.False(t, empty.tabAvailable("table"), "no results yet, so nothing to show")
	assert.False(t, empty.tabAvailable("trace"), "no trace yet, so nothing to show")

	loaded := &MainModel{viewMode: "history", hasTable: true, hasTrace: true}
	assert.True(t, loaded.tabAvailable("table"))
	assert.True(t, loaded.tabAvailable("trace"))
}

// TestModeAtMapsClicksToTabs is what makes the tabs clickable: a column on the
// tab line has to resolve back to the mode drawn there.
func TestModeAtMapsClicksToTabs(t *testing.T) {
	const width = 120
	m := &MainModel{viewMode: "history", hasTable: true, hasTrace: true}

	spans := m.layoutTabs(width)
	require.Len(t, spans, 4)

	for _, span := range spans {
		for _, col := range []int{span.start, (span.start + span.end) / 2, span.end - 1} {
			mode, ok := m.modeAt(width, col)
			assert.Equal(t, span.mode, mode, "column %d should belong to %q", col, span.mode)
			assert.True(t, ok, "%q is available and should accept a click", span.mode)
		}
	}

	// Past the last tab there is nothing to click.
	_, ok := m.modeAt(width, spans[len(spans)-1].end+1)
	assert.False(t, ok)

	// Spans do not overlap and are in display order.
	for i := 1; i < len(spans); i++ {
		assert.GreaterOrEqual(t, spans[i].start, spans[i-1].end,
			"tab %d overlaps the one before it", i)
	}
}

// TestUnavailableTabsIgnoreClicks pairs with the dimming: a tab that looks
// unavailable must not respond.
func TestUnavailableTabsIgnoreClicks(t *testing.T) {
	const width = 120
	m := &MainModel{viewMode: "history"} // no results, no trace

	for _, span := range m.layoutTabs(width) {
		mode, ok := m.modeAt(width, (span.start+span.end)/2)
		switch mode {
		case "table", "trace":
			assert.False(t, ok, "%q has nothing to show and must not be clickable", mode)
		default:
			assert.True(t, ok, "%q should be clickable", mode)
		}
	}
}

// TestClickingATabSwitchesMode covers the point of making the bar clickable.
func TestClickingATabSwitchesMode(t *testing.T) {
	const width = 120

	newModel := func() *MainModel {
		input := textinput.New()
		input.SetWidth(width)
		return &MainModel{
			input:        input,
			viewMode:     "history",
			hasTable:     true,
			hasTrace:     true,
			mouseEnabled: true,
			windowWidth:  width,
			styles:       DefaultStyles(),
		}
	}

	for _, want := range []string{"table", "trace"} {
		t.Run(want, func(t *testing.T) {
			m := newModel()

			var col int
			for _, span := range m.layoutTabs(width) {
				if span.mode == want {
					col = (span.start + span.end) / 2
				}
			}

			m.clickTab(col)
			assert.Equal(t, want, m.viewMode, "clicking the %s tab should switch to it", want)
		})
	}
}

// TestClickingAnUnavailableTabDoesNothing pairs with dimming: a tab with
// nothing behind it must not drop you on an empty screen.
func TestClickingAnUnavailableTabDoesNothing(t *testing.T) {
	const width = 120

	input := textinput.New()
	m := &MainModel{
		input:       input,
		viewMode:    "history",
		windowWidth: width,
		styles:      DefaultStyles(),
	} // no results, no trace

	for _, span := range m.layoutTabs(width) {
		if span.mode != "table" && span.mode != "trace" {
			continue
		}
		m.clickTab((span.start + span.end) / 2)
		assert.Equal(t, "history", m.viewMode,
			"the %s tab has nothing to show and must not be clickable", span.mode)
	}
}

// TestMouseCommandTogglesReporting covers MOUSE ON and MOUSE OFF, which is the
// escape hatch for anyone who would rather have text selection than clicking.
func TestMouseCommandTogglesReporting(t *testing.T) {
	input := textinput.New()
	m := &MainModel{
		input:              input,
		viewMode:           "history",
		mouseEnabled:       true,
		styles:             DefaultStyles(),
		historyViewport:    viewport.New(viewport.WithWidth(80), viewport.WithHeight(10)),
		fullHistoryContent: "",
	}

	_, _, handled := m.handleSpecialCommands("MOUSE OFF")
	assert.True(t, handled, "MOUSE OFF should be handled by the UI, not sent to Cassandra")
	assert.False(t, m.mouseEnabled)
	assert.Contains(t, m.fullHistoryContent, "Mouse: OFF")

	_, _, handled = m.handleSpecialCommands("MOUSE ON")
	assert.True(t, handled)
	assert.True(t, m.mouseEnabled)
	assert.Contains(t, m.fullHistoryContent, "Mouse: ON")

	// Bare MOUSE reports without changing anything.
	before := m.mouseEnabled
	_, _, handled = m.handleSpecialCommands("MOUSE")
	assert.True(t, handled)
	assert.Equal(t, before, m.mouseEnabled, "bare MOUSE should report, not toggle")

	// Anything else is a usage error rather than a silent no-op.
	_, _, handled = m.handleSpecialCommands("MOUSE SIDEWAYS")
	assert.True(t, handled)
	assert.Contains(t, m.fullHistoryContent, "Usage: MOUSE")
}

// TestMouseModeFollowsTheSetting checks the View actually carries it, since
// that is what the terminal acts on.
func TestMouseModeFollowsTheSetting(t *testing.T) {
	m := &MainModel{mouseEnabled: true}
	assert.Equal(t, tea.MouseModeCellMotion, m.newView("").MouseMode,
		"reporting must be on for the tabs to be clickable")

	m.mouseEnabled = false
	assert.Equal(t, tea.MouseModeNone, m.newView("").MouseMode,
		"MOUSE OFF must hand the buttons back to the terminal")

	// The alternate screen is asked for either way.
	assert.True(t, m.newView("").AltScreen)
}
