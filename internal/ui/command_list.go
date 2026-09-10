package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// The box both history views draw: a list of commands to pick one from.
//
// There were two of these, written separately, and they had drifted into
// looking like different features - different indents, different centring, one
// with a separator rule and one without, the range shown on its own line in one
// and in the title in the other, and widths measured in bytes so the arrows and
// bullets threw them off. They do the same job, so they are one thing now, and
// the Ctrl+R search is the same box with a query line at the top.
type commandList struct {
	title    string
	search   string // the Ctrl+R query; searching is what having one means
	items    []string
	selected int
	scroll   int
	rows     int // how many commands to show at once
	width    int
	hint     string
}

// searching reports whether this is the Ctrl+R box rather than the plain list.
func (c commandList) searching() bool { return c.search != "" }

// scrolls reports whether there are more commands than the box can show, and
// so whether the scrollbar has anything to say.
func (c commandList) scrolls() bool { return len(c.items) > c.rows }

// last is the index one past the final command shown.
func (c commandList) last() int {
	return min(c.scroll+c.rows, len(c.items))
}

// visibleItems is how many commands the box is showing.
func (c commandList) visibleItems() int {
	return max(c.last()-c.scroll, 0)
}

// firstItemRow is the row of the first command in the rendered box, counting
// the top border as row 0.
//
// The click handler turns a screen row into a command through this, and the
// tests check it against the text actually drawn there - counting the header by
// hand in two places is how a click applies the row above the one you meant.
func (c commandList) firstItemRow() int {
	row := 2 // the top border, then the title
	if c.searching() {
		row++ // the query line
	}
	return row
}

// render draws the box.
func (c commandList) render(styles *Styles) string {
	titleStyle := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	searchStyle := lipgloss.NewStyle().Foreground(styles.AccentText.GetForeground())
	selectedStyle := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	itemStyle := lipgloss.NewStyle().Foreground(styles.MutedText.GetForeground())
	quietStyle := lipgloss.NewStyle().Foreground(styles.MutedText.GetForeground())
	hintStyle := lipgloss.NewStyle().Foreground(styles.MutedText.GetForeground()).Italic(true)

	inner := c.width - 2 // inside the border

	// The range goes in the title in both, rather than on a line of its own in
	// one of them.
	title := c.title
	if len(c.items) > c.rows {
		title = fmt.Sprintf("%s (%d-%d of %d)", title, c.scroll+1, c.last(), len(c.items))
	}

	var lines []string
	lines = append(lines, titleStyle.Render(centre(title, inner)))

	if c.searching() {
		query := c.search
		if query == searchPrompt {
			query = ""
		}
		lines = append(lines, searchStyle.Render(pad(" Search: "+query, inner)))
	}

	// A scrollbar down the right-hand edge rather than an arrow above and
	// below. The arrows said only whether there was more that way; the bar
	// says how much of the list is showing and where in it you are, which is
	// what you want when the history runs to hundreds of commands.
	thumb := scrollbarColumn(c.last()-c.scroll, c.scroll, len(c.items))
	thumbStyle := lipgloss.NewStyle().Foreground(styles.Accent)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	switch {
	case len(c.items) == 0:
		lines = append(lines, quietStyle.Render(centre("No matching commands", inner)))
	default:
		// Two columns for the marker, so selected and unselected rows line up,
		// and one at the end for the bar.
		for i := c.scroll; i < c.last(); i++ {
			text := ansi.Truncate(c.items[i], max(inner-4, 0), "…")

			row := pad("   "+text, inner-1)
			style := itemStyle
			if i == c.selected {
				row = pad(" → "+text, inner-1)
				style = selectedStyle
			}

			bar := trackStyle.Render("░")
			switch {
			case !c.scrolls():
				bar = " "
			case thumb[i-c.scroll]:
				bar = thumbStyle.Render("█")
			}
			lines = append(lines, style.Render(row)+bar)
		}
	}

	lines = append(lines, quietStyle.Render(strings.Repeat("─", max(inner, 0))))
	lines = append(lines, hintStyle.Render(centre(c.hint, inner)))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		Render(strings.Join(lines, "\n"))
}

// centre pads a string to width with the text in the middle.
//
// Columns, not bytes: the arrows and the bullet in these boxes are multi-byte,
// and measuring them by length is what left the two boxes different widths.
func centre(s string, width int) string {
	s = ansi.Truncate(s, max(width, 0), "…")
	left := max((width-lipgloss.Width(s))/2, 0)
	return strings.Repeat(" ", left) + pad(s, width-left)
}

// pad fills a string out to width.
func pad(s string, width int) string {
	s = ansi.Truncate(s, max(width, 0), "…")
	return s + strings.Repeat(" ", max(width-lipgloss.Width(s), 0))
}

// listWidth is how wide the box should be for its contents.
func listWidth(screenWidth int, parts ...string) int {
	width := 0
	for _, part := range parts {
		if w := lipgloss.Width(part) + 6; w > width { // marker, padding, border
			width = w
		}
	}

	return min(max(width, 40), min(int(float64(screenWidth)*0.8), 100))
}
