package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
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

// tabText renders the tab labels at one of three widths. The bar must never
// wrap onto a second line, so it drops the key hints and then shortens the
// names before anything is left out.
func tabText(width int) []string {
	full := make([]string, len(modeTabs))
	plain := make([]string, len(modeTabs))
	short := make([]string, len(modeTabs))
	for i, t := range modeTabs {
		full[i] = t.label + " (" + t.key + ")"
		plain[i] = t.label
		// Not label[:1]: two labels can share a first letter, and a bar of
		// indistinguishable letters is worse than no bar.
		short[i] = t.short
	}

	for _, candidate := range [][]string{full, plain, short} {
		if tabBarWidth(candidate) <= width {
			return candidate
		}
	}
	return short
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

	labels := tabText(width)

	spans := make([]tabSpan, 0, len(modeTabs))
	col := 0
	for i, t := range modeTabs {
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
	return spans
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
	}

	// Pad by hand rather than with lipgloss Width, which wraps instead of
	// truncating when the content is wider than the line.
	if pad := width - spans[len(spans)-1].end; pad > 0 {
		b.WriteString(strings.Repeat(" ", pad))
	}

	return b.String()
}
