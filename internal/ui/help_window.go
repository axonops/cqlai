package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/router"
)

// The help window, opened by the Help button on the tab line or by F1.
//
// HELP already prints the same rows into the console, which works but scrolls
// away whatever you were looking at to tell you what PAGING does. A window you
// open, read and close leaves the screen alone.
//
// It is not a commandList. That is a list of things to pick one of - it marks a
// selection and Enter uses it - and there is nothing here to pick. What the two
// do share is the parts worth sharing: the scrollbar, the padding, and the rule
// that drawing and hit testing come from one description of the layout.

// helpTitle says how to work the window, including the second key: on the
// terminals that take F1 for themselves, this is where you find out there is
// another one.
const helpTitle = "Help  -  F1 or Alt+H  -  ↑↓ PgUp/PgDn scrolls  -  Esc closes"

// helpMargin is how much screen is left showing around the window, so it reads
// as something over the top of cqlai rather than a new screen.
const helpMargin = 4

// helpWindow is the open help, if any. Only the scroll position is state; the
// rows come from router.HelpRows each time it is drawn.
type helpWindow struct {
	active bool
	scroll int
}

// helpGeometry is where the window sits and how much of it is showing.
type helpGeometry struct {
	x, y          int
	width, height int
	rows          int // content rows, inside the border and the title
	first, last   int // range of help rows drawn, last exclusive
	innerWidth    int
}

// helpLines lays the help rows out as one line each, with the columns lined up.
//
// The category is only printed when it changes, which is how the console
// version reads: the rows carry an empty category to mean "same as above".
func helpLines(rows [][]string) []string {
	category, command := 0, 0
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		category = max(category, lipgloss.Width(row[0]))
		command = max(command, lipgloss.Width(row[1]))
	}

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		lines = append(lines, pad(row[0], category)+"  "+pad(row[1], command)+"  "+row[2])
	}
	return lines
}

// helpGeometry works out where the window sits and which rows it shows.
//
// Drawing and scrolling both go through this, so the scrollbar cannot describe
// a different window from the one on screen.
func (m *MainModel) helpGeometry(screenWidth, screenHeight int) (helpGeometry, []string, bool) {
	if !m.help.active {
		return helpGeometry{}, nil, false
	}

	lines := helpLines(router.HelpRows())
	if len(lines) == 0 {
		return helpGeometry{}, nil, false
	}

	width := min(screenWidth-helpMargin, longestLine(lines)+4) // borders, scrollbar, padding
	height := min(screenHeight-helpMargin, len(lines)+3)       // borders and the title
	if width < 20 || height < 6 {
		// Too small to be worth drawing over what is already there.
		return helpGeometry{}, nil, false
	}

	rows := height - 3
	first := min(max(m.help.scroll, 0), max(len(lines)-rows, 0))

	return helpGeometry{
		x:          (screenWidth - width) / 2,
		y:          (screenHeight - height) / 2,
		width:      width,
		height:     height,
		rows:       rows,
		first:      first,
		last:       min(first+rows, len(lines)),
		innerWidth: width - 2,
	}, lines, true
}

// longestLine is the width of the widest line.
func longestLine(lines []string) int {
	widest := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > widest {
			widest = w
		}
	}
	return widest
}

// viewHelp renders the window and says where to put it.
func (m *MainModel) viewHelp(screenWidth, screenHeight int) (Layer, bool) {
	g, lines, ok := m.helpGeometry(screenWidth, screenHeight)
	if !ok {
		return Layer{}, false
	}

	titleStyle := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())
	thumbStyle := lipgloss.NewStyle().Foreground(m.styles.Accent)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	thumb := scrollbarColumn(g.last-g.first, g.first, len(lines))
	scrolls := len(lines) > g.rows

	var b strings.Builder
	b.WriteString(titleStyle.Render(centre(helpTitle, g.innerWidth)))
	b.WriteString("\n")

	for i := g.first; i < g.last; i++ {
		// One column short, so the scrollbar has somewhere to go.
		b.WriteString(textStyle.Render(pad(" "+lines[i], g.innerWidth-1)))

		switch {
		case !scrolls:
			b.WriteString(" ")
		case thumb[i-g.first]:
			b.WriteString(thumbStyle.Render("█"))
		default:
			b.WriteString(trackStyle.Render("░"))
		}

		if i < g.last-1 {
			b.WriteString("\n")
		}
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Accent).
		Render(b.String())

	return Layer{
		Content: box,
		X:       g.x,
		Y:       g.y,
		Width:   g.width,
		Height:  g.height,
		ZIndex:  300, // over everything; it was just asked for
	}, true
}

// toggleHelp opens the window, or closes it if it is already open.
func (m *MainModel) toggleHelp() (*MainModel, tea.Cmd) {
	if m.help.active {
		m.help = helpWindow{}
		return m, nil
	}
	m.help = helpWindow{active: true}
	return m, nil
}

// scrollHelp moves the window by n rows.
//
// The view behind it does not move: the window is what you are reading.
func (m *MainModel) scrollHelp(n int) (*MainModel, tea.Cmd) {
	g, lines, ok := m.helpGeometry(m.windowWidth, m.windowHeight)
	if !ok {
		return m, nil
	}

	m.help.scroll = min(max(g.first+n, 0), max(len(lines)-g.rows, 0))
	return m, nil
}

// helpPageRows is how far PgUp and PgDn move the help window. A little less
// than a screenful, so a line or two carries over and you can see where you
// were.
const helpPageRows = 15
