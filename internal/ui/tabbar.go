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

// modeTabs lists the tabs in display order. The labels are capitals so the
// line reads as a row of controls rather than a sentence.
var modeTabs = []modeTab{
	// "Console" rather than "History": Ctrl+R searches command history, which
	// is a different thing, and this view is the running transcript of what you
	// typed and what came back.
	{mode: "history", label: "CONSOLE", short: "C", key: "F2"},
	// The cluster itself, next to the console: what is in the database comes
	// before anything a query has made of it, and the three that follow are all
	// views of a result.
	{mode: "schema", label: "SCHEMA", short: "S", key: "F3"},
	// "Results" rather than "Table": the same view shows EXPAND, ASCII and JSON
	// output, none of which is a table, and "table" already means a schema
	// object to anyone using this.
	{mode: "table", label: "RESULTS", short: "R", key: "F4"},
	{mode: "trace", label: "TRACE", short: "T", key: "F5"},
	// "Chat" rather than "AI": it is a conversation, and what it is a
	// conversation with is not the useful half of the name. Two letters
	// because the console has the C.
	{mode: "ai", label: "CHAT", short: "Ch", key: "F6"},
}

// hasResults reports whether there is anything for SAVE to write.
//
// The same thing the SAVE command checks before it refuses, so the button and
// the command agree about whether there is anything there.
func (m *MainModel) hasResults() bool {
	return len(m.lastTableData) > 0
}

// tabAvailable reports whether a tab has anything to show. An unavailable tab
// is dimmed and ignores clicks; the key still works, matching what F3 and F4
// did before, so nobody loses a shortcut they were used to.
func (m *MainModel) tabAvailable(mode string) bool {
	switch mode {
	case "schema":
		// There is no schema to browse with nothing connected, and the tab says
		// so by being dimmed rather than by opening on an apology.
		return m.connected()
	case "table":
		return m.hasTable
	case "trace":
		return m.hasTrace
	case "ai":
		// Dimmed rather than hidden. A tab that is not there says nothing about
		// why; a dimmed one says the view exists and this session cannot reach
		// it, which is the question someone looking for it is asking.
		return m.aiAvailable()
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
	return modeTabs
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

// fitTabs picks the labels to draw and whether the buttons fit beside them.
//
// The order says what is worth giving up first. Key hints go before the button,
// because the keys are also in the help the button opens. The names do not: a
// bar of single letters is harder to read than a line with no button on it, and
// F1 still opens the help either way.
func fitTabs(tabs []modeTab, width int) ([]string, bool) {
	tiers := tabLabelTiers(tabs)

	for i, labels := range tiers {
		if tabBarWidth(labels) <= width-buttonReserve {
			return labels, true
		}
		// Names survive at the button's expense; the first tier, which is
		// names plus key hints, does not.
		if i > 0 && tabBarWidth(labels) <= width {
			return labels, false
		}
	}
	return tiers[len(tiers)-1], tabBarWidth(tiers[len(tiers)-1]) <= width-buttonReserve
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
	labels, withButtons := fitTabs(tabs, width)

	// File then Help, at the left-hand end, so the tabs start after them. What
	// they may take is whatever the tabs do not need.
	spans := make([]tabSpan, 0, len(tabs)+2)
	col := 0
	// The two buttons shorten together. Placing File first and letting it take
	// the longest label that fits leaves Help a single "?" while File still
	// says "FILE (Alt+F)", which reads as one of them mattering more.
	fileLabel, helpLabel := "", ""
	if withButtons {
		fileLabel, helpLabel = fitButtons(width, tabBarWidth(labels))
	}

	if fileLabel != "" {
		spans = append(spans, tabSpan{
			mode: fileMode, label: fileLabel,
			start: 0, end: lipgloss.Width(fileLabel) + 2, available: true,
		})
		col = spans[0].end
	}

	for i, t := range tabs {
		next := col
		if i > 0 || col > 0 {
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

	if helpLabel != "" {
		start := col + len(tabSeparator)
		spans = append(spans, tabSpan{
			mode: helpMode, label: helpLabel,
			start: start, end: start + lipgloss.Width(helpLabel) + 2, available: true,
		})
	}
	return spans
}

// fitButtons picks the pair of labels, longest first, that leaves the tabs
// their room. Empty strings mean there is no space for the buttons at all, and
// the tabs get the line to themselves: they are the navigation.
func fitButtons(width, tabsWidth int) (file, help string) {
	sep := len(tabSeparator)
	for i := range fileLabels {
		f, h := fileLabels[i], helpLabels[i]
		used := lipgloss.Width(f) + 2 + sep + tabsWidth + sep + lipgloss.Width(h) + 2
		if used <= width {
			return f, h
		}
	}
	return "", ""
}

// The two buttons on the tab line, one at each end.
//
// Help is on the left because it is the thing you reach for when you do not
// know where anything is, and the left is where reading starts. Save is on the
// right, out of the way of the tabs, because it acts on what is already on
// screen rather than moving you between views.
//
// Both keep their labels, longest first, and shorten before they disappear.
var (
	// Both keys, because F1 does not reach the application on Terminator,
	// Konsole and others - they take it for their own help - so a button
	// naming only F1 names a key that does nothing there.
	helpLabels = []string{"HELP (F1/Alt+H)", "HELP", "?"}
	fileLabels = []string{"FILE (Alt+F)", "FILE", "F"}
)

// buttonReserve is the room the tabs give up for the two buttons: their
// *shortest* useful labels, not their longest. Reserving for the longest would
// shorten the tab names on a terminal that was only ever going to fit "Help".
var (
	fileReserve = lipgloss.Width(" F ")
	helpReserve = lipgloss.Width(" ? ")

	buttonReserve = fileReserve + helpReserve
)

// The pseudo-modes the buttons report. Neither is a view: they open a window
// over whichever view you are in and leave it there.
const (
	helpMode = "help"
	fileMode = "file"
)

// tabSpanFor finds a span by mode, so a control can be anchored to where it is
// actually drawn rather than to a column worked out a second time.
func (m *MainModel) tabSpanFor(width int, mode string) (tabSpan, bool) {
	for _, span := range m.layoutTabs(width) {
		if span.mode == mode {
			return span, true
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
		Foreground(dimmedColour)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3a3a3a"))

	var b strings.Builder
	col := 0
	for i, span := range spans {
		if i > 0 {
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
