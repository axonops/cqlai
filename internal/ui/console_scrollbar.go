package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// The scrollbar down the right-hand edge of the Console.
//
// The scroll position was already worked out - [TOP], [42%], [BOTTOM] - and
// appended to the status line, where it is pushed past the right-hand edge and
// has never once been visible. So a long transcript gave no sign of where you
// were in it, or that there was anything above.
//
// The settings lists, the command history and the help window all draw one
// through scrollbarColumn; this is the same bar beside the view you spend the
// most time in.

// scrollbarWidth is the column the bar takes. The Console viewport is that much
// narrower than the window, and its content is wrapped to the narrower width,
// so no line loses a character to it.
const scrollbarWidth = 1

// consoleWidth is the width the Console viewport is given.
func consoleWidth(windowWidth int) int {
	return max(windowWidth-scrollbarWidth, 0)
}

// withScrollbar draws a scrollbar down the right of a rendered viewport.
//
// It is drawn whether or not the content scrolls. The column is reserved
// either way - the viewport is a column narrower than the window - so leaving
// it blank wastes the space and leaves you unsure whether there is nothing
// above or no scrollbar at all. A full thumb says the whole transcript is on
// screen, which is worth knowing.
func withScrollbar(content string, total, height, offset int, styles *Styles) string {
	rows := strings.Split(content, "\n")
	if height <= 0 {
		return content
	}

	thumb := scrollbarColumn(min(height, len(rows)), offset, total)
	if total <= height {
		// Everything fits: the thumb fills the bar.
		for i := range thumb {
			thumb[i] = true
		}
	}
	thumbStyle := lipgloss.NewStyle().Foreground(styles.Accent)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	for i := range rows {
		if i >= len(thumb) {
			break
		}
		if thumb[i] {
			rows[i] += thumbStyle.Render("█")
		} else {
			rows[i] += trackStyle.Render("░")
		}
	}
	return strings.Join(rows, "\n")
}

// consoleScrollbar draws the bar beside the Console.
func (m *MainModel) consoleScrollbar(content string) string {
	return withScrollbar(content,
		m.historyViewport.TotalLineCount(),
		m.historyViewport.Height(),
		m.historyViewport.YOffset(),
		m.styles)
}
