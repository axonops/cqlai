package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/ai"
)

// The mode tabs across the top.
//
// The four modes used to be reachable only by F-key, with nothing on screen
// saying so. The tabs put the choices and their keys in front of you, and mark
// the ones that have nothing to show yet.

// modeTab is one entry in the tab bar.
type modeTab struct {
	mode  string // matches MainModel.viewMode
	label string
	short string // one or two letters for a terminal too narrow for the label
	key   string
}

// modeTabs lists the tabs in display order. F6 is deliberately absent: it sits
// next to these keys but toggles data types rather than switching mode.
var modeTabs = []modeTab{
	// "Console" rather than "History": Ctrl+R searches command history, which
	// is a different thing, and this view is the running transcript of what you
	// typed and what came back.
	{mode: "history", label: "Console", short: "C", key: "F2"},
	// "Results" rather than "Table": the same view shows EXPAND, ASCII and JSON
	// output, none of which is a table, and "table" already means a schema
	// object to anyone using this.
	{mode: "table", label: "Results", short: "R", key: "F3"},
	{mode: "trace", label: "Trace", short: "T", key: "F4"},
	{mode: "ai", label: "AI", short: "A", key: "F5"},
}

// tabAvailable reports whether a tab has anything to show. An unavailable tab
// is dimmed and ignores clicks; the key still works, matching what F3 and F4
// did before, so nobody loses a shortcut they were used to.
func (m *MainModel) tabAvailable(mode string) bool {
	switch mode {
	case "table":
		return m.hasTable
	case "trace":
		return m.hasTrace
	default:
		return true
	}
}

// visibleTabs is the tabs to draw.
//
// AI is left out rather than dimmed when there is no provider configured.
// Results and Trace are dimmed because they fill in as you work; whether AI is
// set up cannot change while cqlai is running, so a permanently dimmed tab
// would only be taking room from the others.
func (m *MainModel) visibleTabs() []modeTab {
	if m.aiAvailable() {
		return modeTabs
	}

	tabs := make([]modeTab, 0, len(modeTabs))
	for _, t := range modeTabs {
		if t.mode != "ai" {
			tabs = append(tabs, t)
		}
	}
	return tabs
}

// aiAvailable reports whether the AI view has a provider behind it.
func (m *MainModel) aiAvailable() bool {
	return ai.IsConfigured(m.aiConfig)
}

// tabLabelTiers renders the tab labels at three widths, widest first.
//
// The bar must never wrap onto a second line, so it drops the key hints and
// then shortens the names before anything is left out.
func tabLabelTiers(tabs []modeTab) [][]string {
	full := make([]string, len(tabs))
	plain := make([]string, len(tabs))
	short := make([]string, len(tabs))
	for i, t := range tabs {
		full[i] = t.label + " (" + t.key + ")"
		plain[i] = t.label
		// Not label[:1]: two labels can share a first letter, and a bar of
		// indistinguishable letters is worse than no bar.
		short[i] = t.short
	}
	return [][]string{full, plain, short}
}

// fitTabs picks the labels to draw and whether the Help button fits beside
// them.
//
// The order says what is worth giving up first. Key hints go before the button,
// because the keys are also in the help the button opens. The names do not: a
// bar of single letters is harder to read than a line with no button on it, and
// F1 still opens the help either way.
func fitTabs(tabs []modeTab, width int) ([]string, bool) {
	tiers := tabLabelTiers(tabs)

	for i, labels := range tiers {
		if tabBarWidth(labels) <= width-helpReserve {
			return labels, true
		}
		// Names survive at the button's expense; the first tier, which is
		// names plus key hints, does not.
		if i > 0 && tabBarWidth(labels) <= width {
			return labels, false
		}
	}
	return tiers[len(tiers)-1], tabBarWidth(tiers[len(tiers)-1]) <= width-helpReserve
}

// tabSeparator sits between tabs; the padding is part of each tab's click area.
const tabSeparator = "  "

func tabBarWidth(labels []string) int {
	total := 0
	for _, l := range labels {
		total += len(l) + 2 // one space of padding each side
	}
	return total + len(tabSeparator)*(len(labels)-1)
}

// tabSpan is where a tab sits on the line, used to turn a click into a mode.
type tabSpan struct {
	mode       string
	label      string
	start, end int // column range, end exclusive
	available  bool
	active     bool
}

// layoutTabs decides what fits on the line and where each tab sits.
//
// Rendering and click handling both go through this, so a click always lands on
// the tab that was actually drawn there. On a terminal too narrow even for
// single letters, tabs are dropped from the right rather than wrapped: a second
// line would push the viewport down and break the height arithmetic.
func (m *MainModel) layoutTabs(width int) []tabSpan {
	if width <= 0 {
		return nil
	}

	tabs := m.visibleTabs()
	labels, withHelp := fitTabs(tabs, width)

	spans := make([]tabSpan, 0, len(tabs))
	col := 0
	for i, t := range tabs {
		next := col
		if i > 0 {
			next += len(tabSeparator)
		}
		end := next + len(labels[i]) + 2
		if end > width {
			break // no room left; leave the rest off rather than wrap
		}

		spans = append(spans, tabSpan{
			mode:      t.mode,
			label:     labels[i],
			start:     next,
			end:       end,
			available: m.tabAvailable(t.mode),
			active:    m.viewMode == t.mode,
		})
		col = end
	}

	if withHelp {
		if help, ok := helpSpan(width, col); ok {
			spans = append(spans, help)
		}
	}
	return spans
}

// helpLabels are the Help button's labels, longest first.
//
// Both keys, because F1 does not reach the application on Terminator, Konsole
// and others - they take it for their own help - so a button naming only F1
// names a key that does nothing on those terminals.
var helpLabels = []string{"Help (F1/Alt+H)", "Help", "?"}

// helpReserve is the room the tabs give up for the button: its *shortest*
// useful label, not its longest. Reserving for the longest would shorten the
// tab names on a terminal that was only ever going to fit "Help".
var helpReserve = lipgloss.Width(" Help ")

// helpMode is the pseudo-mode the Help button reports. It is not a view: it
// opens a window over whichever view you are in and leaves it there.
const helpMode = "help"

// helpSpan is where the Help button sits, at the right-hand end of the line.
//
// It shortens before it disappears, and it disappears before it would sit on
// top of a tab: losing the button is better than a line where a click lands on
// whatever happens to be underneath it.
func helpSpan(width, tabsEnd int) (tabSpan, bool) {
	for _, label := range helpLabels {
		start := width - lipgloss.Width(label) - 2
		if start > tabsEnd {
			return tabSpan{
				mode:      helpMode,
				label:     label,
				start:     start,
				end:       width,
				available: true,
			}, true
		}
	}
	return tabSpan{}, false
}

// modeAt returns the mode whose tab covers a column, and whether it can be
// switched to. An unavailable tab reports false so a click on it does nothing.
func (m *MainModel) modeAt(width, col int) (string, bool) {
	for _, span := range m.layoutTabs(width) {
		if col >= span.start && col < span.end {
			return span.mode, span.available
		}
	}
	return "", false
}

// ViewTabBar renders the tab line.
func (m *MainModel) ViewTabBar(width int) string {
	spans := m.layoutTabs(width)
	if len(spans) == 0 {
		return strings.Repeat(" ", max(width, 0))
	}

	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1c1c1c")).
		Background(lipgloss.Color("#87D7FF")).
		Bold(true)

	availableStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87D7FF"))

	unavailableStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#585858"))

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3a3a3a"))

	var b strings.Builder
	col := 0
	for i, span := range spans {
		// The Help button is placed against the right edge rather than after
		// the tab before it, so pad out to wherever it starts.
		switch {
		case span.mode == helpMode:
			b.WriteString(strings.Repeat(" ", max(span.start-col, 0)))
		case i > 0:
			b.WriteString(separatorStyle.Render(tabSeparator))
		}

		text := " " + span.label + " "
		switch {
		case span.active:
			b.WriteString(activeStyle.Render(text))
		case span.available:
			b.WriteString(availableStyle.Render(text))
		default:
			b.WriteString(unavailableStyle.Render(text))
		}
		col = span.end
	}

	// Pad by hand rather than with lipgloss Width, which wraps instead of
	// truncating when the content is wider than the line.
	if pad := width - col; pad > 0 {
		b.WriteString(strings.Repeat(" ", pad))
	}

	return b.String()
}
